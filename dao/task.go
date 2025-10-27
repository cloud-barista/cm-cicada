package dao

import (
	"errors"

	"github.com/cloud-barista/cm-cicada/db"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"gorm.io/gorm"
)

func TaskCreate(task *model.TaskDBModel) (*model.TaskDBModel, error) {
	result := db.DB.Create(task)
	err := result.Error
	if err != nil {
		return nil, err
	}

	return task, nil
}

func TaskGet(id string) (*model.TaskDBModel, error) {
	task := &model.TaskDBModel{}

	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	result := db.DB.Where("id = ?", id).First(task)
	err := result.Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("task not found with the provided id")
		}
		return nil, err
	}

	return task, nil
}

func TaskGetByWorkflowIDAndName(workflowID string, name string) (*model.TaskDBModel, error) {
	task := &model.TaskDBModel{}

	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	result := db.DB.Where("workflow_id = ? and name = ?", workflowID, name).First(task)
	err := result.Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("task not found with the provided name")
		}
		return nil, err
	}

	return task, nil
}

func TaskDelete(task *model.TaskDBModel) error {
	result := db.DB.Delete(task)
	err := result.Error
	if err != nil {
		return err
	}

	return nil
}
