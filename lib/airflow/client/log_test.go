package client

import (
	"encoding/json"
	"strings"
	"testing"
)

// Airflow 3 returns task logs as an array of structured records, or of plain
// strings, where the Airflow 2 API returned one blob. Both must render to text.
func TestRenderLogContent(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
	}{
		{
			name:    "structured records with timestamps",
			content: `[{"timestamp":"2026-07-15T08:26:48.114353Z","event":"Pre Execute","level":"info"},{"event":"::endgroup::"}]`,
			want:    []string{"2026-07-15T08:26:48.114Z", "Pre Execute", "::endgroup::"},
		},
		{
			name:    "record without a timestamp renders bare",
			content: `[{"event":"only text"}]`,
			want:    []string{"only text"},
		},
		{
			name:    "plain string array",
			content: `["line one","line two"]`,
			want:    []string{"line one", "line two"},
		},
		{
			name:    "empty array",
			content: `[]`,
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderLogContent(json.RawMessage(tt.content))
			for _, want := range tt.want {
				if !strings.Contains(got, want) {
					t.Errorf("rendered log missing %q\ngot: %q", want, got)
				}
			}
		})
	}
}

func TestRenderLogContentEmpty(t *testing.T) {
	if got := renderLogContent(nil); got != "" {
		t.Errorf("renderLogContent(nil) = %q, want empty", got)
	}
}

// An unrecognised shape must not silently swallow the log.
func TestRenderLogContentUnknownShape(t *testing.T) {
	got := renderLogContent(json.RawMessage(`{"unexpected":"object"}`))
	if !strings.Contains(got, "unexpected") {
		t.Errorf("unknown log shape was dropped; got %q", got)
	}
}

// jwtExpiry must read the exp claim so the client renews on the server's
// schedule rather than a guess.
func TestJWTExpiry(t *testing.T) {
	// {"exp":1893456000} -> 2030-01-01T00:00:00Z
	token := "header." + base64URL(`{"exp":1893456000}`) + ".sig"
	got := jwtExpiry(token)
	if got.Unix() != 1893456000 {
		t.Errorf("jwtExpiry = %v (unix %d), want unix 1893456000", got, got.Unix())
	}
}

func TestJWTExpiryMalformedFallsBack(t *testing.T) {
	for _, token := range []string{"", "not-a-jwt", "a.b", "a.!!!notbase64!!!.c"} {
		got := jwtExpiry(token)
		if got.IsZero() {
			t.Errorf("jwtExpiry(%q) returned zero time; want a fallback in the future", token)
		}
	}
}

func base64URL(s string) string {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	var out strings.Builder
	data := []byte(s)
	for i := 0; i < len(data); i += 3 {
		var b [3]byte
		n := copy(b[:], data[i:])
		out.WriteByte(alphabet[b[0]>>2])
		out.WriteByte(alphabet[(b[0]&0x03)<<4|b[1]>>4])
		if n > 1 {
			out.WriteByte(alphabet[(b[1]&0x0f)<<2|b[2]>>6])
		}
		if n > 2 {
			out.WriteByte(alphabet[b[2]&0x3f])
		}
	}
	return out.String()
}
