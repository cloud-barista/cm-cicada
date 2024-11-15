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
