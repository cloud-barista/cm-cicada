package service

import (
	"errors"

	"github.com/cloud-barista/cm-cicada/dao"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/google/uuid"
)

type TaskComponentService struct{}

func NewTaskComponentService() *TaskComponentService {
	return &TaskComponentService{}
}

func (s *TaskComponentService) Create(req model.CreateTaskComponentReq) (*model.TaskComponent, error) {
	if req.Name == "" {
		return nil, errors.New("please provide the name")
	}

	taskComponent := model.TaskComponent{
		ID:   uuid.New().String(),
		Name: req.Name,
		Data: req.Data,
	}

	created, err := dao.TaskComponentCreate(&taskComponent)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *TaskComponentService) Get(id string) (*model.TaskComponent, error) {
	return dao.TaskComponentGet(id)
}

func (s *TaskComponentService) GetByName(name string) (*model.TaskComponent, error) {
	tc := dao.TaskComponentGetByName(name)
	if tc == nil {
		return nil, errors.New("task component not found with the provided name")
	}
	return tc, nil
}

func (s *TaskComponentService) List(page, row int) (*[]model.TaskComponent, error) {
	return dao.TaskComponentGetList(page, row)
}

func (s *TaskComponentService) Update(id string, req model.CreateTaskComponentReq) (*model.TaskComponent, error) {
	existing, err := dao.TaskComponentGet(id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	existing.Data = req.Data

	if err := dao.TaskComponentUpdate(existing); err != nil {
		return nil, err
	}

	return existing, nil
}

func (s *TaskComponentService) Delete(id string) error {
	taskComponent, err := dao.TaskComponentGet(id)
	if err != nil {
		return err
	}

	return dao.TaskComponentDelete(taskComponent)
}
