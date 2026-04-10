package dao

import (
	"errors"
	"strings"
	"time"

	"github.com/cloud-barista/cm-cicada/db"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"gorm.io/gorm"
)

func TaskCreate(task *model.TaskDBModel) (*model.TaskDBModel, error) {
	if err := ensureDB(); err != nil {
		return nil, err
	}

	task.IsDeleted = false
	task.DeletedAt = nil
	task.DeletedBy = ""
	if task.TaskKey == "" {
		task.TaskKey = task.ID
	}

	result := db.DB.Create(task)
	if result.Error != nil {
		return nil, result.Error
	}

	return task, nil
}

func TaskSave(task *model.TaskDBModel) error {
	if err := ensureDB(); err != nil {
		return err
	}

	var existing model.TaskDBModel
	result := db.DB.Unscoped().Where("id = ?", task.ID).First(&existing)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			_, err := TaskCreate(task)
			return err
		}
		return result.Error
	}

	task.IsDeleted = false
	task.DeletedAt = nil
	task.DeletedBy = ""
	if task.TaskKey == "" {
		task.TaskKey = existing.TaskKey
		if task.TaskKey == "" {
			task.TaskKey = task.ID
		}
	}

	return db.DB.Model(&model.TaskDBModel{}).
		Where("id = ?", task.ID).
		Updates(task).Error
}

func TaskGet(id string) (*model.TaskDBModel, error) {
	if err := ensureDB(); err != nil {
		return nil, err
	}

	task := &model.TaskDBModel{}
	result := db.DB.Where("id = ? AND is_deleted = ?", id, false).First(task)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("task not found with the provided id")
		}
		return nil, result.Error
	}

	if task.TaskKey == "" {
		task.TaskKey = task.ID
		_ = db.DB.Model(&model.TaskDBModel{}).
			Where("id = ?", task.ID).
			Update("task_key", task.TaskKey).Error
	}

	return task, nil
}

func TaskGetIncludeDeleted(id string) (*model.TaskDBModel, error) {
	if err := ensureDB(); err != nil {
		return nil, err
	}

	task := &model.TaskDBModel{}
	result := db.DB.Where("id = ?", id).First(task)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("task not found with the provided id")
		}
		return nil, result.Error
	}

	if task.TaskKey == "" {
		task.TaskKey = task.ID
		_ = db.DB.Model(&model.TaskDBModel{}).
			Where("id = ?", task.ID).
			Update("task_key", task.TaskKey).Error
	}

	return task, nil
}

func TaskGetByWorkflowIDAndName(workflowID string, name string) (*model.TaskDBModel, error) {
	if err := ensureDB(); err != nil {
		return nil, err
	}

	task := &model.TaskDBModel{}
	result := db.DB.Where("workflow_id = ? AND name = ? AND is_deleted = ?", workflowID, name, false).First(task)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("task not found with the provided name")
		}
		return nil, result.Error
	}

	if task.TaskKey == "" {
		task.TaskKey = task.ID
		_ = db.DB.Model(&model.TaskDBModel{}).
			Where("id = ?", task.ID).
			Update("task_key", task.TaskKey).Error
	}

	return task, nil
}

func TaskGetByWorkflowIDAndNameIncludeDeleted(workflowID string, name string) (*model.TaskDBModel, error) {
	if err := ensureDB(); err != nil {
		return nil, err
	}

	task := &model.TaskDBModel{}
	result := db.DB.Where("workflow_id = ? AND name = ?", workflowID, name).First(task)
	if result.Error != nil {
		return nil, result.Error
	}

	if task.TaskKey == "" {
		task.TaskKey = task.ID
		_ = db.DB.Model(&model.TaskDBModel{}).
			Where("id = ?", task.ID).
			Update("task_key", task.TaskKey).Error
	}

	return task, nil
}

func TaskGetByWorkflowIDAndTaskKey(workflowID string, taskKey string) (*model.TaskDBModel, error) {
	if err := ensureDB(); err != nil {
		return nil, err
	}

	task := &model.TaskDBModel{}
	result := db.DB.Where("workflow_id = ? AND task_key = ? AND is_deleted = ?", workflowID, taskKey, false).First(task)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("task not found with the provided task key")
		}
		return nil, result.Error
	}

	if task.TaskKey == "" {
		task.TaskKey = task.ID
		_ = db.DB.Model(&model.TaskDBModel{}).
			Where("id = ?", task.ID).
			Update("task_key", task.TaskKey).Error
	}

	return task, nil
}

func TaskGetByWorkflowKeyAndTaskKey(workflowKey string, taskKey string) (*model.TaskDBModel, error) {
	if err := ensureDB(); err != nil {
		return nil, err
	}

	task := &model.TaskDBModel{}
	if isLikelyUUID(taskKey) {
		result := db.DB.Where("workflow_key = ? AND task_key = ? AND is_deleted = ?", workflowKey, taskKey, false).First(task)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				// Fallback: Airflow task_id may be task name, not task_key (UUID).
				fallback := db.DB.Where("workflow_key = ? AND name = ? AND is_deleted = ?", workflowKey, taskKey, false).First(task)
				if fallback.Error != nil {
					if errors.Is(fallback.Error, gorm.ErrRecordNotFound) {
						return nil, errors.New("task not found with the provided task key or name")
					}
					return nil, fallback.Error
				}
				return task, nil
			}
			return nil, result.Error
		}
	} else {
		// Prefer name lookup to avoid log noise when taskKey is a name.
		result := db.DB.Where("workflow_key = ? AND name = ? AND is_deleted = ?", workflowKey, taskKey, false).First(task)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				fallback := db.DB.Where("workflow_key = ? AND task_key = ? AND is_deleted = ?", workflowKey, taskKey, false).First(task)
				if fallback.Error != nil {
					if errors.Is(fallback.Error, gorm.ErrRecordNotFound) {
						return nil, errors.New("task not found with the provided task key or name")
					}
					return nil, fallback.Error
				}
				return task, nil
			}
			return nil, result.Error
		}
	}

	if task.TaskKey == "" {
		task.TaskKey = task.ID
		_ = db.DB.Model(&model.TaskDBModel{}).
			Where("id = ?", task.ID).
			Update("task_key", task.TaskKey).Error
	}

	return task, nil
}

func isLikelyUUID(value string) bool {
	if len(value) != 36 {
		return false
	}
	// Fast shape check for UUID: 8-4-4-4-12 hex with dashes.
	dashPos := []int{8, 13, 18, 23}
	for i, ch := range value {
		if containsInt(dashPos, i) {
			if ch != '-' {
				return false
			}
			continue
		}
		if !strings.ContainsRune("0123456789abcdefABCDEF", ch) {
			return false
		}
	}
	return true
}

func containsInt(list []int, v int) bool {
	for _, n := range list {
		if n == v {
			return true
		}
	}
	return false
}

func TaskGetListByWorkflowID(workflowID string, includeDeleted bool) ([]model.TaskDBModel, error) {
	if err := ensureDB(); err != nil {
		return nil, err
	}

	tasks := []model.TaskDBModel{}
	query := db.DB.Where("workflow_id = ?", workflowID)
	if !includeDeleted {
		query = query.Where("is_deleted = ?", false)
	}

	if err := query.Find(&tasks).Error; err != nil {
		return nil, err
	}

	for i := range tasks {
		if tasks[i].TaskKey == "" {
			tasks[i].TaskKey = tasks[i].ID
			_ = db.DB.Model(&model.TaskDBModel{}).
				Where("id = ?", tasks[i].ID).
				Update("task_key", tasks[i].TaskKey).Error
		}
	}

	return tasks, nil
}

func TaskDelete(task *model.TaskDBModel) error {
	if err := ensureDB(); err != nil {
		return err
	}

	now := time.Now()
	return db.DB.Model(&model.TaskDBModel{}).
		Where("id = ? AND is_deleted = ?", task.ID, false).
		Updates(map[string]interface{}{
			"is_deleted": true,
			"deleted_at": &now,
		}).Error
}

func TaskSoftDeleteByWorkflowID(workflowID string) error {
	if err := ensureDB(); err != nil {
		return err
	}

	now := time.Now()
	return db.DB.Model(&model.TaskDBModel{}).
		Where("workflow_id = ? AND is_deleted = ?", workflowID, false).
		Updates(map[string]interface{}{
			"is_deleted": true,
			"deleted_at": &now,
		}).Error
}
