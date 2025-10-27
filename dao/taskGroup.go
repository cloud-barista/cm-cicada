package dao

import (
	"errors"

	"github.com/cloud-barista/cm-cicada/db"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"gorm.io/gorm"
)

func TaskGroupCreate(taskGroup *model.TaskGroupDBModel) (*model.TaskGroupDBModel, error) {
	result := db.DB.Create(taskGroup)
	err := result.Error
	if err != nil {
		return nil, err
	}

	return taskGroup, nil
}

func TaskGroupGet(id string) (*model.TaskGroupDBModel, error) {
	taskGroup := &model.TaskGroupDBModel{}

	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	result := db.DB.Where("id = ?", id).First(taskGroup)
	err := result.Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("task_group not found with the provided id")
		}
		return nil, err
	}

	return taskGroup, nil
}

func TaskGroupGetByWorkflowIDAndName(workflowID string, name string) (*model.TaskGroupDBModel, error) {
	taskGroup := &model.TaskGroupDBModel{}

	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	result := db.DB.Where("workflow_id = ? and name = ?", workflowID, name).First(taskGroup)
	err := result.Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("task_group not found with the provided name")
		}
		return nil, err
	}

	return taskGroup, nil
}

func TaskGroupDelete(taskGroup *model.TaskGroupDBModel) error {
	result := db.DB.Delete(taskGroup)
	err := result.Error
	if err != nil {
		return err
	}

	return nil
}
