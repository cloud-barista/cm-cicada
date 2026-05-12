package service

import (
	"errors"

	"github.com/cloud-barista/cm-cicada/lib/airflow/catalog"
)

type TaskTypeService struct{}

func NewTaskTypeService() *TaskTypeService {
	return &TaskTypeService{}
}

// List returns all task type definitions in catalog file order.
func (s *TaskTypeService) List() []catalog.TaskTypeDef {
	return catalog.List()
}

// Get returns the catalog entry for the given id, or an error if not found.
func (s *TaskTypeService) Get(id string) (*catalog.TaskTypeDef, error) {
	if id == "" {
		return nil, errors.New("task type id is empty")
	}
	def, ok := catalog.Get(id)
	if !ok {
		return nil, errors.New("task type not found: " + id)
	}
	return &def, nil
}
