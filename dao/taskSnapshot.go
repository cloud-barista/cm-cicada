package dao

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/cloud-barista/cm-cicada/db"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TaskSnapshotCreate(snapshot *model.TaskSnapshot) (*model.TaskSnapshot, error) {
	if err := ensureDB(); err != nil {
		return nil, err
	}

	if snapshot.ID == "" {
		snapshot.ID = uuid.New().String()
	}
	if snapshot.CapturedAt.IsZero() {
		snapshot.CapturedAt = time.Now()
	}

	if err := db.DB.Create(snapshot).Error; err != nil {
		return nil, err
	}

	return snapshot, nil
}

func TaskSnapshotCreateFromTask(taskDB *model.TaskDBModel, rawTask model.Task, snapshotType string) error {
	if err := ensureDB(); err != nil {
		return err
	}

	rawJSON, err := json.Marshal(rawTask)
	if err != nil {
		return err
	}

	_, err = TaskSnapshotCreate(&model.TaskSnapshot{
		WorkflowID:   taskDB.WorkflowID,
		WorkflowKey:  taskDB.WorkflowKey,
		TaskID:       taskDB.ID,
		TaskKey:      taskDB.TaskKey,
		TaskName:     taskDB.Name,
		TaskGroupID:  taskDB.TaskGroupID,
		SnapshotType: snapshotType,
		RawTask:      string(rawJSON),
		CapturedAt:   time.Now(),
	})
	return err
}

func TaskSnapshotGetLatest(workflowID string, taskID string) (*model.TaskSnapshot, error) {
	if err := ensureDB(); err != nil {
		return nil, err
	}

	snapshot := &model.TaskSnapshot{}
	err := db.DB.
		Where("workflow_id = ? AND task_id = ?", workflowID, taskID).
		Order("captured_at desc").
		First(snapshot).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("task snapshot not found")
		}
		return nil, err
	}

	return snapshot, nil
}

func TaskSnapshotGetLatestRawTask(workflowID string, taskID string) (*model.Task, error) {
	snapshot, err := TaskSnapshotGetLatest(workflowID, taskID)
	if err != nil {
		return nil, err
	}

	var task model.Task
	if err := json.Unmarshal([]byte(snapshot.RawTask), &task); err != nil {
		return nil, err
	}
	return &task, nil
}
