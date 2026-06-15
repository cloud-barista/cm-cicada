package dao

import (
	"errors"

	"gorm.io/gorm"

	"github.com/cloud-barista/cm-cicada/db"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
)

// WorkflowScheduleCreate inserts a new workflow_schedules row.
func WorkflowScheduleCreate(s *model.WorkflowSchedule) error {
	return db.DB.Create(s).Error
}

// WorkflowScheduleGetByID returns a single schedule row by id, or nil when
// no row matches (no error in that case).
func WorkflowScheduleGetByID(id string) (*model.WorkflowSchedule, error) {
	var s model.WorkflowSchedule
	if err := db.DB.Where("id = ?", id).First(&s).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

// WorkflowScheduleListByWorkflowID returns every schedule row attached to a
// workflow, ordered by run_at ascending.
func WorkflowScheduleListByWorkflowID(workflowID string) ([]model.WorkflowSchedule, error) {
	var out []model.WorkflowSchedule
	if err := db.DB.Where("workflow_id = ?", workflowID).
		Order("run_at ASC").
		Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// WorkflowScheduleGetLatest returns the workflow's most recently created
// schedule row regardless of status. Used by the GET endpoint so callers
// can see the last lifecycle state (active / executed / canceled). Returns
// (nil, nil) when the workflow has no schedule history.
func WorkflowScheduleGetLatest(workflowID string) (*model.WorkflowSchedule, error) {
	var s model.WorkflowSchedule
	err := db.DB.Where("workflow_id = ?", workflowID).
		Order("created_at DESC").
		First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

// WorkflowScheduleGetActive returns the single active schedule for a
// workflow (the one that drives the DAG metadata), or nil when none exists.
// Multiple active rows for the same workflow are not expected — service
// layer guards against creating more than one — but if they exist this
// returns the earliest run_at.
func WorkflowScheduleGetActive(workflowID string) (*model.WorkflowSchedule, error) {
	var s model.WorkflowSchedule
	err := db.DB.Where("workflow_id = ? AND status = ?", workflowID, model.WorkflowScheduleStatusActive).
		Order("run_at ASC").
		First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

// WorkflowScheduleUpdateStatus transitions a schedule row to a new status.
func WorkflowScheduleUpdateStatus(id string, status model.WorkflowScheduleStatus) error {
	return db.DB.Model(&model.WorkflowSchedule{}).
		Where("id = ?", id).
		Update("status", status).Error
}
