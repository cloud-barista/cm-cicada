package client

import (
	"context"
	"net/url"
)

// CreateConnection registers a connection.
func (c *Client) CreateConnection(ctx context.Context, conn Connection) (Connection, error) {
	var out Connection
	err := c.do(ctx, "POST", "/api/v2/connections", nil, conn, &out)
	return out, err
}

// GetConnection fetches one connection.
func (c *Client) GetConnection(ctx context.Context, connectionID string) (Connection, error) {
	var out Connection
	err := c.do(ctx, "GET", "/api/v2/connections/"+pathEscape(connectionID), nil, nil, &out)
	return out, err
}

// ListConnections lists connections. Non-positive limit/offset and an empty
// orderBy are omitted so the server's defaults apply.
func (c *Client) ListConnections(ctx context.Context, limit, offset int32, orderBy string) (ConnectionCollection, error) {
	q := url.Values{}
	if limit > 0 {
		q.Set("limit", itoa(int(limit)))
	}
	if offset > 0 {
		q.Set("offset", itoa(int(offset)))
	}
	if orderBy != "" {
		q.Set("order_by", orderBy)
	}

	var out ConnectionCollection
	err := c.do(ctx, "GET", "/api/v2/connections", q, nil, &out)
	return out, err
}

// UpdateConnection patches a connection.
func (c *Client) UpdateConnection(ctx context.Context, connectionID string, conn Connection) (Connection, error) {
	var out Connection
	err := c.do(ctx, "PATCH", "/api/v2/connections/"+pathEscape(connectionID), nil, conn, &out)
	return out, err
}

// DeleteConnection removes a connection.
func (c *Client) DeleteConnection(ctx context.Context, connectionID string) error {
	return c.do(ctx, "DELETE", "/api/v2/connections/"+pathEscape(connectionID), nil, nil, nil)
}
