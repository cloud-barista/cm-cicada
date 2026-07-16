// Package client is a minimal client for the Apache Airflow 3 REST API (/api/v2).
//
// It replaces github.com/apache/airflow-client-go, which is generated from the
// Airflow 2 spec and unmaintained since Feb 2023. Airflow 3 removed /api/v1
// outright (404) and rejects Basic auth (401), so the generated client cannot
// talk to an Airflow 3 server at all.
//
// Only the operations cm-cicada actually uses are implemented.
package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// tokenRenewMargin renews the JWT slightly before it actually expires so an
// in-flight request doesn't race the expiry.
const tokenRenewMargin = 60 * time.Second

// Config describes how to reach the Airflow API server.
type Config struct {
	// Address is host or host:port, without a scheme.
	Address       string
	Username      string
	Password      string
	UseTLS        bool
	SkipTLSVerify bool
	Timeout       time.Duration
}

// Client talks to the Airflow 3 public API. It is safe for concurrent use.
type Client struct {
	baseURL string
	cfg     Config
	http    *http.Client

	// tokenMu guards the cached JWT.
	tokenMu     sync.Mutex
	token       string
	tokenExpiry time.Time
}

// APIError is a non-2xx response from the Airflow API.
type APIError struct {
	StatusCode int
	Method     string
	Path       string
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("airflow api: %s %s: %d: %s", e.Method, e.Path, e.StatusCode, e.Body)
}

// IsNotFound reports whether err is a 404 from the Airflow API. Callers use it
// to tell "no such DAG" apart from a transport failure.
func IsNotFound(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}

// New builds a client. It does not contact the server.
func New(cfg Config) *Client {
	scheme := "http"
	transport := http.DefaultTransport
	if cfg.UseTLS {
		scheme = "https"
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.SkipTLSVerify},
		}
	}

	return &Client{
		baseURL: scheme + "://" + cfg.Address,
		cfg:     cfg,
		http: &http.Client{
			Timeout:   cfg.Timeout,
			Transport: transport,
		},
	}
}

// BaseURL returns the server root, e.g. "http://127.0.0.1:8080".
func (c *Client) BaseURL() string { return c.baseURL }

// Timeout returns the per-request timeout.
func (c *Client) Timeout() time.Duration { return c.cfg.Timeout }

// Health pings the API server. Airflow 3 serves this without authentication,
// so it works as a liveness probe before any token exists.
func (c *Client) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v2/version", nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusOK {
		return &APIError{StatusCode: resp.StatusCode, Method: http.MethodGet, Path: "/api/v2/version"}
	}
	return nil
}

// token returns a valid JWT, fetching a new one if the cached one is missing or
// close to expiry.
func (c *Client) accessToken(ctx context.Context, force bool) (string, error) {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()

	if !force && c.token != "" && time.Now().Before(c.tokenExpiry.Add(-tokenRenewMargin)) {
		return c.token, nil
	}

	body, err := json.Marshal(map[string]string{
		"username": c.cfg.Username,
		"password": c.cfg.Password,
	})
	if err != nil {
		return "", err
	}

	// Airflow 3 issues API tokens here. This endpoint is outside /api/v2.
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/auth/token", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", &APIError{
			StatusCode: resp.StatusCode,
			Method:     http.MethodPost,
			Path:       "/auth/token",
			Body:       truncate(string(respBody)),
		}
	}

	var out struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(respBody, &out); err != nil {
		return "", fmt.Errorf("airflow api: cannot decode token response: %w", err)
	}
	if out.AccessToken == "" {
		return "", fmt.Errorf("airflow api: token response had no access_token")
	}

	c.token = out.AccessToken
	c.tokenExpiry = jwtExpiry(out.AccessToken)

	return c.token, nil
}

// jwtExpiry reads the "exp" claim so we renew on the server's schedule rather
// than guessing. An unreadable claim falls back to a conservative lifetime.
func jwtExpiry(token string) time.Time {
	const fallback = 10 * time.Minute

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return time.Now().Add(fallback)
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return time.Now().Add(fallback)
	}
	var claims struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil || claims.Exp == 0 {
		return time.Now().Add(fallback)
	}
	return time.Unix(claims.Exp, 0)
}

// do performs an authenticated request against the API and decodes a JSON
// response into out (which may be nil to discard the body).
//
// A 401 is retried once with a freshly minted token: the cached JWT can expire
// or be invalidated server-side between calls.
func (c *Client) do(ctx context.Context, method, path string, query url.Values, body any, out any) error {
	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}

	send := func(force bool) (*http.Response, []byte, error) {
		token, err := c.accessToken(ctx, force)
		if err != nil {
			return nil, nil, err
		}

		u := c.baseURL + path
		if len(query) > 0 {
			u += "?" + query.Encode()
		}

		var reader io.Reader
		if payload != nil {
			reader = bytes.NewReader(payload)
		}
		req, err := http.NewRequestWithContext(ctx, method, u, reader)
		if err != nil {
			return nil, nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept", "application/json")
		if payload != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.http.Do(req)
		if err != nil {
			return nil, nil, err
		}
		defer func() { _ = resp.Body.Close() }()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, nil, err
		}
		return resp, respBody, nil
	}

	resp, respBody, err := send(false)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		resp, respBody, err = send(true)
		if err != nil {
			return err
		}
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return &APIError{
			StatusCode: resp.StatusCode,
			Method:     method,
			Path:       path,
			Body:       truncate(string(respBody)),
		}
	}

	if out == nil || len(respBody) == 0 {
		return nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("airflow api: %s %s: cannot decode response: %w", method, path, err)
	}
	return nil
}

// doRaw performs an authenticated request and returns the undecoded body.
func (c *Client) doRaw(ctx context.Context, method, path string, query url.Values) ([]byte, error) {
	var raw json.RawMessage
	if err := c.do(ctx, method, path, query, nil, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func truncate(s string) string {
	const max = 512
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

// pathEscape escapes a value for use as a single URL path segment, so an ID
// containing '/' or other reserved characters cannot break out of its segment.
//
// '+' and ':' are legal in a path segment and deliberately stay literal: unlike
// in a query string, '+' here does not mean space. DAG run IDs look like
// manual__2026-07-15T07:30:49.391273+00:00 and round-trip as-is.
func pathEscape(s string) string { return url.PathEscape(s) }
