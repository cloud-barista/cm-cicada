package airflow

import (
	"errors"

	"github.com/cloud-barista/cm-cicada/lib/airflow/client"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/jollaman999/utils/logger"
)

// RegisterConnection creates a connection, replacing any existing one with the
// same ID.
func (c *Client) RegisterConnection(connection *model.Connection) error {
	ctx, cancel := Context()
	defer cancel()

	// Delete first so re-registering an existing connection updates it.
	// A missing connection is fine, so the error is deliberately ignored.
	_ = c.api.DeleteConnection(ctx, connection.ID)

	_, err := c.api.CreateConnection(ctx, toAirflowConnection(connection))
	if err != nil {
		errMsg := "AIRFLOW: Error occurred while registering connection. (ConnID: " + connection.ID + ", Error: " + err.Error() + ")."
		logger.Println(logger.ERROR, false, errMsg)

		return errors.New(errMsg)
	}

	return nil
}

func (c *Client) CreateConnection(connection *model.Connection) (*model.Connection, error) {
	ctx, cancel := Context()
	defer cancel()

	created, err := c.api.CreateConnection(ctx, toAirflowConnection(connection))
	if err != nil {
		errMsg := "AIRFLOW: Error occurred while creating connection. (ConnID: " + connection.ID + ", Error: " + err.Error() + ")."
		logger.Println(logger.ERROR, false, errMsg)
		return nil, errors.New(errMsg)
	}

	result := toModelConnection(created)
	return &result, nil
}

func (c *Client) GetConnection(connectionID string) (*model.Connection, error) {
	ctx, cancel := Context()
	defer cancel()

	conn, err := c.api.GetConnection(ctx, connectionID)
	if err != nil {
		errMsg := "AIRFLOW: Error occurred while getting connection. (ConnID: " + connectionID + ", Error: " + err.Error() + ")."
		logger.Println(logger.ERROR, false, errMsg)
		return nil, errors.New(errMsg)
	}

	result := toModelConnection(conn)
	return &result, nil
}

func (c *Client) ListConnections(limit int32, offset int32, orderBy string) ([]model.Connection, error) {
	ctx, cancel := Context()
	defer cancel()

	collection, err := c.api.ListConnections(ctx, limit, offset, orderBy)
	if err != nil {
		errMsg := "AIRFLOW: Error occurred while listing connections. (Error: " + err.Error() + ")."
		logger.Println(logger.ERROR, false, errMsg)
		return nil, errors.New(errMsg)
	}

	connections := make([]model.Connection, 0, len(collection.Connections))
	for _, item := range collection.Connections {
		connections = append(connections, toModelConnectionSummary(item))
	}

	return connections, nil
}

func (c *Client) UpdateConnection(connectionID string, connection *model.Connection) (*model.Connection, error) {
	ctx, cancel := Context()
	defer cancel()

	updated, err := c.api.UpdateConnection(ctx, connectionID, toAirflowConnection(connection))
	if err != nil {
		errMsg := "AIRFLOW: Error occurred while updating connection. (ConnID: " + connectionID + ", Error: " + err.Error() + ")."
		logger.Println(logger.ERROR, false, errMsg)
		return nil, errors.New(errMsg)
	}

	result := toModelConnection(updated)
	return &result, nil
}

func (c *Client) DeleteConnection(connectionID string) error {
	ctx, cancel := Context()
	defer cancel()

	err := c.api.DeleteConnection(ctx, connectionID)
	if err != nil {
		errMsg := "AIRFLOW: Error occurred while deleting connection. (ConnID: " + connectionID + ", Error: " + err.Error() + ")."
		logger.Println(logger.ERROR, false, errMsg)
		return errors.New(errMsg)
	}

	return nil
}

func toAirflowConnection(connection *model.Connection) client.Connection {
	return client.Connection{
		ConnectionID: connection.ID,
		ConnType:     connection.Type,
		Description:  &connection.Description,
		Host:         &connection.Host,
		Login:        &connection.Login,
		Schema:       &connection.Schema,
		Port:         &connection.Port,
		Password:     &connection.Password,
		Extra:        &connection.Extra,
	}
}

func toModelConnection(conn client.Connection) model.Connection {
	return model.Connection{
		ID:          conn.ConnectionID,
		Type:        conn.ConnType,
		Description: stringOrEmpty(conn.Description),
		Host:        stringOrEmpty(conn.Host),
		Port:        int32OrZero(conn.Port),
		Schema:      stringOrEmpty(conn.Schema),
		Login:       stringOrEmpty(conn.Login),
		Password:    stringOrEmpty(conn.Password),
		Extra:       stringOrEmpty(conn.Extra),
	}
}

// toModelConnectionSummary maps a connection for list responses, leaving out
// password and extra.
//
// The Airflow 2 API had a separate, narrower schema for collection items that
// carried neither field. Airflow 3 returns the full connection in listings, and
// it only masks extra keys whose names it recognises as secret — a key like
// "custom_creds" comes back in plaintext. Listings stay narrow so a bulk read
// cannot hand out credentials.
func toModelConnectionSummary(conn client.Connection) model.Connection {
	summary := toModelConnection(conn)
	summary.Password = ""
	summary.Extra = ""
	return summary
}

func stringOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func int32OrZero(value *int32) int32 {
	if value == nil {
		return 0
	}
	return *value
}
