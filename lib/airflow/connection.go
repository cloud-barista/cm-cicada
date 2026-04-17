package airflow

import (
	"errors"

	"github.com/apache/airflow-client-go/airflow"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/jollaman999/utils/logger"
)

func (client *Client) RegisterConnection(connection *model.Connection) error {
	ctx, cancel := Context()
	defer cancel()

	description := airflow.NullableString{}
	description.Set(&connection.Description)

	host := airflow.NullableString{}
	host.Set(&connection.Host)

	login := airflow.NullableString{}
	login.Set(&connection.Login)

	schema := airflow.NullableString{}
	schema.Set(&connection.Schema)

	port := airflow.NullableInt32{}
	port.Set(&connection.Port)

	extra := airflow.NullableString{}
	extra.Set(&connection.Extra)

	conn := airflow.Connection{
		ConnectionId: &connection.ID,
		ConnType:     &connection.Type,
		Description:  description,
		Host:         host,
		Login:        login,
		Schema:       schema,
		Port:         port,
		Password:     &connection.Password,
		Extra:        extra,
	}

	_, _ = client.ConnectionApi.DeleteConnection(ctx, connection.ID).Execute()

	_, _, err := client.ConnectionApi.PostConnection(ctx).Connection(conn).Execute()
	if err != nil {
		errMsg := "AIRFLOW: Error occurred while registering connection. (ConnID: " + connection.ID + ", Error: " + err.Error() + ")."
		logger.Println(logger.ERROR, false, errMsg)

		return errors.New(errMsg)
	}

	return nil
}

func (client *Client) CreateConnection(connection *model.Connection) (*model.Connection, error) {
	ctx, cancel := Context()
	defer cancel()

	conn := toAirflowConnection(connection)
	created, _, err := client.ConnectionApi.PostConnection(ctx).Connection(conn).Execute()
	if err != nil {
		errMsg := "AIRFLOW: Error occurred while creating connection. (ConnID: " + connection.ID + ", Error: " + err.Error() + ")."
		logger.Println(logger.ERROR, false, errMsg)
		return nil, errors.New(errMsg)
	}

	result := toModelConnection(created)
	return &result, nil
}

func (client *Client) GetConnection(connectionID string) (*model.Connection, error) {
	ctx, cancel := Context()
	defer cancel()

	conn, _, err := client.ConnectionApi.GetConnection(ctx, connectionID).Execute()
	if err != nil {
		errMsg := "AIRFLOW: Error occurred while getting connection. (ConnID: " + connectionID + ", Error: " + err.Error() + ")."
		logger.Println(logger.ERROR, false, errMsg)
		return nil, errors.New(errMsg)
	}

	result := toModelConnection(conn)
	return &result, nil
}

func (client *Client) ListConnections(limit int32, offset int32, orderBy string) ([]model.Connection, error) {
	ctx, cancel := Context()
	defer cancel()

	req := client.ConnectionApi.GetConnections(ctx)
	if limit > 0 {
		req = req.Limit(limit)
	}
	if offset > 0 {
		req = req.Offset(offset)
	}
	if orderBy != "" {
		req = req.OrderBy(orderBy)
	}

	collection, _, err := req.Execute()
	if err != nil {
		errMsg := "AIRFLOW: Error occurred while listing connections. (Error: " + err.Error() + ")."
		logger.Println(logger.ERROR, false, errMsg)
		return nil, errors.New(errMsg)
	}

	connections := make([]model.Connection, 0)
	if collection.Connections != nil {
		for _, item := range *collection.Connections {
			connections = append(connections, toModelConnectionFromItem(item))
		}
	}

	return connections, nil
}

func (client *Client) UpdateConnection(connectionID string, connection *model.Connection) (*model.Connection, error) {
	ctx, cancel := Context()
	defer cancel()

	conn := toAirflowConnection(connection)
	updated, _, err := client.ConnectionApi.PatchConnection(ctx, connectionID).Connection(conn).Execute()
	if err != nil {
		errMsg := "AIRFLOW: Error occurred while updating connection. (ConnID: " + connectionID + ", Error: " + err.Error() + ")."
		logger.Println(logger.ERROR, false, errMsg)
		return nil, errors.New(errMsg)
	}

	result := toModelConnection(updated)
	return &result, nil
}

func (client *Client) DeleteConnection(connectionID string) error {
	ctx, cancel := Context()
	defer cancel()

	_, err := client.ConnectionApi.DeleteConnection(ctx, connectionID).Execute()
	if err != nil {
		errMsg := "AIRFLOW: Error occurred while deleting connection. (ConnID: " + connectionID + ", Error: " + err.Error() + ")."
		logger.Println(logger.ERROR, false, errMsg)
		return errors.New(errMsg)
	}

	return nil
}

func toAirflowConnection(connection *model.Connection) airflow.Connection {
	description := airflow.NullableString{}
	description.Set(&connection.Description)

	host := airflow.NullableString{}
	host.Set(&connection.Host)

	login := airflow.NullableString{}
	login.Set(&connection.Login)

	schema := airflow.NullableString{}
	schema.Set(&connection.Schema)

	port := airflow.NullableInt32{}
	port.Set(&connection.Port)

	extra := airflow.NullableString{}
	extra.Set(&connection.Extra)

	conn := airflow.Connection{
		ConnectionId: &connection.ID,
		ConnType:     &connection.Type,
		Description:  description,
		Host:         host,
		Login:        login,
		Schema:       schema,
		Port:         port,
		Password:     &connection.Password,
		Extra:        extra,
	}
	return conn
}

func toModelConnection(conn airflow.Connection) model.Connection {
	return model.Connection{
		ID:          valueOrEmpty(conn.ConnectionId),
		Type:        valueOrEmpty(conn.ConnType),
		Description: nullableStringValue(conn.Description),
		Host:        nullableStringValue(conn.Host),
		Port:        nullableInt32Value(conn.Port),
		Schema:      nullableStringValue(conn.Schema),
		Login:       nullableStringValue(conn.Login),
		Password:    valueOrEmpty(conn.Password),
		Extra:       nullableStringValue(conn.Extra),
	}
}

func toModelConnectionFromItem(item airflow.ConnectionCollectionItem) model.Connection {
	return model.Connection{
		ID:          valueOrEmpty(item.ConnectionId),
		Type:        valueOrEmpty(item.ConnType),
		Description: nullableStringValue(item.Description),
		Host:        nullableStringValue(item.Host),
		Port:        nullableInt32Value(item.Port),
		Schema:      nullableStringValue(item.Schema),
		Login:       nullableStringValue(item.Login),
	}
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func nullableStringValue(value airflow.NullableString) string {
	val := value.Get()
	if val == nil {
		return ""
	}
	return *val
}

func nullableInt32Value(value airflow.NullableInt32) int32 {
	val := value.Get()
	if val == nil {
		return 0
	}
	return *val
}
