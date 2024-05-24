package dao

import (
	"errors"
	"github.com/cloud-barista/cm-cicada/db"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"gorm.io/gorm"
	"time"
)

func WorkflowCreate(workflow *model.Workflow) (*model.Workflow, error) {
	workflow.CreatedAt = time.Now()
	workflow.UpdatedAt = time.Now()

	result := db.DB.Create(workflow)
	err := result.Error
	if err != nil {
		return nil, err
	}

	return workflow, nil
}

func WorkflowGet(id string) (*model.Workflow, error) {
	workflow := &model.Workflow{}

	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	result := db.DB.Where("id = ?", id).First(workflow)
	err := result.Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("workflow not found with the provided id")
		}
		return nil, err
	}

	return workflow, nil
}

func WorkflowGetList(page int, row int) (*[]model.Workflow, error) {
	WorkflowList := &[]model.Workflow{}
	// Ensure db.DB is not nil to avoid runtime panics
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}

	result := db.DB.Scopes(func(d *gorm.DB) *gorm.DB {
		var filtered = d

		if page != 0 && row != 0 {
			offset := (page - 1) * row

			return filtered.Offset(offset).Limit(row)
		} else if row != 0 && page == 0 {
			filtered.Error = errors.New("row is not 0 but page is 0")
			return filtered
		} else if page != 0 && row == 0 {
			filtered.Error = errors.New("page is not 0 but row is 0")
			return filtered
		}
		return filtered
	}).Find(WorkflowList)

	err := result.Error
	if err != nil {
		return nil, err
	}

	return WorkflowList, nil
}

func WorkflowUpdate(workflow *model.Workflow) error {
	workflow.UpdatedAt = time.Now()

	result := db.DB.Model(&model.Workflow{}).Where("id = ?", workflow.ID).Updates(workflow)
	err := result.Error
	if err != nil {
		return err
	}

	return nil
}

func WorkflowDelete(workflow *model.Workflow) error {
	result := db.DB.Delete(workflow)
	err := result.Error
	if err != nil {
		return err
	}

	return nil
}
