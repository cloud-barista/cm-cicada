package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// TaskDBModel is the persistent representation of a task row.
type TaskDBModel struct {
	ID           string     `gorm:"primaryKey" json:"id" mapstructure:"id" validate:"required"`
	Name         string     `json:"name" mapstructure:"name" validate:"required"`
	WorkflowID   string     `gorm:"column:workflow_id;index" json:"workflow_id" mapstructure:"workflow_id" validate:"required"`
	WorkflowKey  string     `gorm:"column:workflow_key;index" json:"-" mapstructure:"-"`
	TaskGroupID  string     `gorm:"column:task_group_id;index" json:"task_group_id" mapstructure:"task_group_id" validate:"required"`
	TaskGroupKey string     `gorm:"column:task_group_key;index" json:"-" mapstructure:"-"`
	TaskKey      string     `gorm:"column:task_key;index" json:"-" mapstructure:"-"`
	IsDeleted    bool       `gorm:"column:is_deleted;index;default:false" json:"-" mapstructure:"-"`
	DeletedAt    *time.Time `gorm:"column:deleted_at" json:"-" mapstructure:"-"`
}

// TaskGroupDBModel is the persistent representation of a task group row.
type TaskGroupDBModel struct {
	ID           string     `gorm:"primaryKey" json:"id" mapstructure:"id" validate:"required"`
	Name         string     `json:"name" mapstructure:"name" validate:"required"`
	WorkflowID   string     `gorm:"column:workflow_id;index" json:"workflow_id" mapstructure:"workflow_id" validate:"required"`
	WorkflowKey  string     `gorm:"column:workflow_key;index" json:"-" mapstructure:"-"`
	TaskGroupKey string     `gorm:"column:task_group_key;index" json:"-" mapstructure:"-"`
	IsDeleted    bool       `gorm:"column:is_deleted;index;default:false" json:"-" mapstructure:"-"`
	DeletedAt    *time.Time `gorm:"column:deleted_at" json:"-" mapstructure:"-"`
}

// TaskSnapshot preserves the raw payload of a deleted task for audit/history.
type TaskSnapshot struct {
	ID           string    `gorm:"primaryKey" json:"id"`
	WorkflowID   string    `gorm:"column:workflow_id;index:idx_task_snapshots_workflow_task,priority:1;index" json:"workflow_id"`
	WorkflowKey  string    `gorm:"column:workflow_key;index" json:"workflow_key"`
	TaskID       string    `gorm:"column:task_id;index:idx_task_snapshots_workflow_task,priority:2;index" json:"task_id"`
	TaskKey      string    `gorm:"column:task_key;index" json:"task_key"`
	TaskName     string    `gorm:"column:task_name;index" json:"task_name"`
	TaskGroupID  string    `gorm:"column:task_group_id;index" json:"task_group_id"`
	SnapshotType string    `gorm:"column:snapshot_type;index" json:"snapshot_type"`
	RawTask      string    `gorm:"column:raw_task;type:text" json:"raw_task"`
	CapturedAt   time.Time `gorm:"column:captured_at;index" json:"captured_at"`
}

// Workflow is the persistent root aggregate — the workflow record in DB.
type Workflow struct {
	ID               string     `gorm:"primaryKey" json:"id" mapstructure:"id" validate:"required"`
	WorkflowKey      string     `gorm:"column:workflow_key;index;unique" json:"-" mapstructure:"-"`
	SpecVersion      string     `gorm:"column:spec_version" json:"spec_version" mapstructure:"spec_version" validate:"required"`
	Name             string     `gorm:"column:name;type:text collate nocase" json:"name" mapstructure:"name" validate:"required"`
	Data             Data       `gorm:"column:data" json:"data" mapstructure:"data" validate:"required"`
	CurrentVersionID string     `gorm:"column:current_version_id;index" json:"-" mapstructure:"-"`
	IsDeleted        bool       `gorm:"column:is_deleted;index;default:false" json:"-" mapstructure:"-"`
	DeletedAt        *time.Time `gorm:"column:deleted_at" json:"-" mapstructure:"-"`
	CreatedAt        time.Time  `gorm:"column:created_at;autoCreateTime:false" json:"created_at" mapstructure:"created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;autoCreateTime:false" json:"updated_at" mapstructure:"updated_at"`
}

// WorkflowVersion preserves a snapshot of a Workflow at a given revision.
type WorkflowVersion struct {
	ID               string    `gorm:"primaryKey" json:"id" mapstructure:"id" validate:"required"`
	SpecVersion      string    `gorm:"column:spec_version" json:"spec_version" mapstructure:"spec_version" validate:"required"`
	WorkflowID       string    `gorm:"column:workflowId;index" json:"workflowId" mapstructure:"workflowId" validate:"required"`
	VersionNo        int       `gorm:"column:version_no;index" json:"version_no" mapstructure:"version_no"`
	RawData          Workflow  `gorm:"column:raw_data" json:"data" mapstructure:"data"`
	Action           string    `gorm:"column:action" json:"action" mapstructure:"action" validate:"required"`
	SourceType       string    `gorm:"column:source_type" json:"source_type,omitempty" mapstructure:"source_type"`
	SourceTemplateID string    `gorm:"column:source_template_id" json:"source_template_id,omitempty" mapstructure:"source_template_id"`
	CreatedAt        time.Time `gorm:"column:created_at" json:"created_at" mapstructure:"created_at"`
}

// --- driver.Valuer / sql.Scanner implementations for JSON columns ---

func (d Workflow) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *Workflow) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("invalid type for Workflow")
	}
	return json.Unmarshal(bytes, d)
}
