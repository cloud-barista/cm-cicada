package service

import (
	"errors"

	"github.com/cloud-barista/cm-cicada/lib/airflow"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
)

type ConnectionService struct{}

func NewConnectionService() *ConnectionService {
	return &ConnectionService{}
}

func (s *ConnectionService) Create(conn *model.Connection) (*model.Connection, error) {
	if conn.ID == "" {
		return nil, errors.New("please provide the id")
	}
	if conn.Type == "" {
		return nil, errors.New("please provide the type")
	}

	client, err := airflow.GetClient()
	if err != nil {
		return nil, err
	}

	return client.CreateConnection(conn)
}

func (s *ConnectionService) Get(id string) (*model.Connection, error) {
	client, err := airflow.GetClient()
	if err != nil {
		return nil, err
	}

	return client.GetConnection(id)
}

func (s *ConnectionService) List(page, row int, orderBy string) ([]model.Connection, error) {
	var limit, offset int32
	if row > 0 {
		limit = int32(row)
		offset = int32(page * row)
	}

	client, err := airflow.GetClient()
	if err != nil {
		return nil, err
	}

	return client.ListConnections(limit, offset, orderBy)
}

func (s *ConnectionService) Update(id string, conn *model.Connection) (*model.Connection, error) {
	if conn.ID != "" && conn.ID != id {
		return nil, errors.New("path connId and body id do not match")
	}
	conn.ID = id

	if conn.Type == "" {
		return nil, errors.New("please provide the type")
	}

	client, err := airflow.GetClient()
	if err != nil {
		return nil, err
	}

	return client.UpdateConnection(id, conn)
}

func (s *ConnectionService) Delete(id string) error {
	client, err := airflow.GetClient()
	if err != nil {
		return err
	}

	return client.DeleteConnection(id)
}
