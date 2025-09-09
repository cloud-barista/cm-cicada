package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

const (
	WorkflowSpecVersion_1_0 = "1.0"
)

const (
	WorkflowSpecVersion_LATEST = WorkflowSpecVersion_1_0
)

type Task struct {
	ID            string                 `gorm:"primaryKey" json:"id" mapstructure:"id" validate:"required"`
	Name          string                 `json:"name" mapstructure:"name" validate:"required"`
	TaskComponent string                 `json:"task_component" mapstructure:"task_component" validate:"required"`
	RequestBody   string                 `json:"request_body" mapstructure:"request_body" validate:"required"`
	PathParams    map[string]string      `json:"path_params" mapstructure:"path_params"`
	QueryParams   map[string]string      `json:"query_params" mapstructure:"query_params"`
	Extra         map[string]interface{} `json:"extra,omitempty" mapstructure:"extra"`
	Dependencies  []string               `json:"dependencies" mapstructure:"dependencies"`
}

type TaskDirectly struct {
	ID            string                 `gorm:"primaryKey" json:"id" mapstructure:"id" validate:"required"`
	WorkflowID    string                 `json:"workflow_id" mapstructure:"workflow_id" validate:"required"`
	TaskGroupID   string                 `json:"task_group_id" mapstructure:"task_group_id" validate:"required"`
	Name          string                 `json:"name" mapstructure:"name" validate:"required"`
	TaskComponent string                 `json:"task_component" mapstructure:"task_component" validate:"required"`
	RequestBody   string                 `json:"request_body" mapstructure:"request_body" validate:"required"`
	PathParams    map[string]string      `json:"path_params" mapstructure:"path_params"`
	QueryParams   map[string]string      `json:"query_params" mapstructure:"query_params"`
	Extra         map[string]interface{} `json:"extra,omitempty" mapstructure:"extra"`
	Dependencies  []string               `json:"dependencies" mapstructure:"dependencies"`
}

type TaskDBModel struct {
	ID          string `gorm:"primaryKey" json:"id" mapstructure:"id" validate:"required"`
	Name        string `json:"name" mapstructure:"name" validate:"required"`
	WorkflowID  string `gorm:"column:workflow_id" json:"workflow_id" mapstructure:"workflow_id" validate:"required"`
	TaskGroupID string `gorm:"column:task_group_id" json:"task_group_id" mapstructure:"task_group_id" validate:"required"`
}

type CreateTaskReq struct {
	Name          string            `json:"name" mapstructure:"name" validate:"required"`
	TaskComponent string            `json:"task_component" mapstructure:"task_component" validate:"required"`
	RequestBody   string            `json:"request_body" mapstructure:"request_body" validate:"required"`
	PathParams    map[string]string `json:"path_params" mapstructure:"path_params"`
	Dependencies  []string          `json:"dependencies" mapstructure:"dependencies"`
}

type TaskGroup struct {
	ID          string `gorm:"primaryKey" json:"id" mapstructure:"id" validate:"required"`
	Name        string `json:"name" mapstructure:"name" validate:"required"`
	Description string `json:"description" mapstructure:"description"`
	Tasks       []Task `json:"tasks" mapstructure:"tasks" validate:"required"`
}

type TaskGroupDirectly struct {
	ID          string `json:"id" mapstructure:"id" validate:"required"`
	WorkflowID  string `json:"workflow_id" mapstructure:"workflow_id" validate:"required"`
	Name        string `json:"name" mapstructure:"name" validate:"required"`
	Description string `json:"description" mapstructure:"description"`
	Tasks       []Task `json:"tasks" mapstructure:"tasks" validate:"required"`
}

type TaskGroupDBModel struct {
	ID         string `gorm:"primaryKey" json:"id" mapstructure:"id" validate:"required"`
	Name       string `json:"name" mapstructure:"name" validate:"required"`
	WorkflowID string `gorm:"column:workflow_id" json:"workflow_id" mapstructure:"workflow_id" validate:"required"`
}

type CreateTaskGroupReq struct {
	Name        string          `json:"name" mapstructure:"name" validate:"required"`
	Description string          `json:"description" mapstructure:"description"`
	Tasks       []CreateTaskReq `json:"tasks" mapstructure:"tasks" validate:"required"`
}

type Data struct {
	Description string      `json:"description" mapstructure:"description"`
	TaskGroups  []TaskGroup `json:"task_groups" mapstructure:"task_groups" validate:"required"`
}

type CreateDataReq struct {
	Description string               `json:"description" mapstructure:"description"`
	TaskGroups  []CreateTaskGroupReq `json:"task_groups" mapstructure:"task_groups" validate:"required"`
}

type Workflow struct {
	ID          string    `gorm:"primaryKey" json:"id" mapstructure:"id" validate:"required"`
	SpecVersion string    `gorm:"column:spec_version" json:"spec_version" mapstructure:"spec_version" validate:"required"`
	Name        string    `gorm:"index:,column:name,unique;type:text collate nocase" json:"name" mapstructure:"name" validate:"required"`
	Data        Data      `gorm:"column:data" json:"data" mapstructure:"data" validate:"required"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime:false" json:"created_at" mapstructure:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoCreateTime:false" json:"updated_at" mapstructure:"updated_at"`
}

type WorkflowVersion struct {
	ID          string    `gorm:"primaryKey" json:"id" mapstructure:"id" validate:"required"`
	SpecVersion string    `gorm:"column:spec_version" json:"spec_version" mapstructure:"spec_version" validate:"required"`
	WorkflowID  string    `gorm:"column:workflowId" json:"workflowId" mapstructure:"workflowId" validate:"required"`
	Data        Workflow  `gorm:"column:data" json:"data" mapstructure:"data"`
	Action      string    `gorm:"column:action" json:"action" mapstructure:"action" validate:"required"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at" mapstructure:"created_at"`
}

type CreateWorkflowReq struct {
	SpecVersion string        `json:"spec_version" mapstructure:"spec_version"`
	Name        string        `json:"name" mapstructure:"name" validate:"required"`
	Data        CreateDataReq `json:"data" mapstructure:"data" validate:"required"`
}

type WorkflowRun struct {
	WorkflowRunID          string                 `json:"workflow_run_id,omitempty"`
	WorkflowID             *string                `json:"workflow_id,omitempty"`
	LogicalDate            string                 `json:"logical_date,omitempty"`
	ExecutionDate          time.Time              `json:"execution_date,omitempty"`
	StartDate              time.Time              `json:"start_date,omitempty"`
	EndDate                time.Time              `json:"end_date,omitempty"`
	DurationDate           float64                `json:"duration_date,omitempty"`
	DataIntervalStart      time.Time              `json:"data_interval_start,omitempty"`
	DataIntervalEnd        time.Time              `json:"data_interval_end,omitempty"`
	LastSchedulingDecision time.Time              `json:"last_scheduling_decision,omitempty"`
	RunType                string                 `json:"run_type,omitempty"`
	State                  string                 `json:"state,omitempty"`
	ExternalTrigger        *bool                  `json:"external_trigger,omitempty"`
	Conf                   map[string]interface{} `json:"conf,omitempty"`
	Note                   string                 `json:"note,omitempty"`
}

type TaskInstance struct {
	WorkflowRunID string    `json:"workflow_run_id,omitempty"`
	WorkflowID    *string   `json:"workflow_id,omitempty"`
	TaskID        string    `json:"task_id,omitempty"`
	TaskName      string    `json:"task_name,omitempty"`
	State         string    `json:"state,omitempty"`
	StartDate     time.Time `json:"start_date,omitempty"`
	EndDate       time.Time `json:"end_date,omitempty"`
	DurationDate  float64   `json:"duration_date"`
	ExecutionDate time.Time `json:"execution_date,omitempty"`
	TryNumber     int       `json:"try_number"`
}

type TaskInstanceReference struct {
	// The task ID.
	TaskId   *string `json:"task_id,omitempty"`
	TaskName string  `json:"task_name,omitempty"`
	// The DAG ID.
	WorkflowID *string `json:"workflow_id,omitempty"`
	// The DAG run ID.
	WorkflowRunID *string `json:"workflow_run_id,omitempty"`
	ExecutionDate *string `json:"execution_date,omitempty"`
}

type TaskLog struct {
	Content string `json:"content,omitempty"`
}
type EventLogs struct {
	EventLogs    []EventLog `json:"event_logs"`
	TotalEntries int        `json:"total_entries"`
}

type EventLog struct {
	WorkflowRunID string    `json:"workflow_run_id"`
	RunID         string    `json:"run_id,omitempty"`
	WorkflowID    string    `json:"workflow_id"`
	TaskID        string    `json:"task_id"`
	TaskName      string    `json:"task_name"`
	Event         string    `json:"event,omitempty"`
	When          time.Time `json:"when,omitempty"`
	Extra         string    `json:"extra,omitempty"`
}

type TaskClearOption struct {
	DryRun            bool     `json:"dryRun"`
	TaskIds           []string `json:"taskIds"`
	IncludeDownstream bool     `json:"includeDownstream"`
	//IncludeFuture     bool     `json:"includeFuture"`
	//IncludeParentdag  bool     `json:"includeParentdag"`
	//IncludePast       bool     `json:"includePast"`
	//IncludeSubdags    bool     `json:"includeSubdags"`
	IncludeUpstream bool `json:"includeUpstream"`
	OnlyFailed      bool `json:"onlyFailed"`
	OnlyRunning     bool `json:"onlyRunning"`
	ResetDagRuns    bool `json:"resetDagRuns"`
}

func (d Data) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *Data) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("invalid type for Data")
	}
	return json.Unmarshal(bytes, d)
}

func (d CreateDataReq) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *CreateDataReq) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("invalid type for CreateDataReq")
	}
	return json.Unmarshal(bytes, d)
}
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

// AfterCreate Hook for Workflow to add WorkflowVersion on create
func (w *Workflow) AfterCreate(tx *gorm.DB) (err error) {
	workflowVersion := WorkflowVersion{
		ID:         "create_" + time.Now().String(),
		WorkflowID: w.ID,
		// Version:    "",  // 기본 버전
		Data:      *w,
		Action:    "create",
		CreatedAt: time.Now(),
	}

	if err := tx.Create(&workflowVersion).Error; err != nil {
		return err
	}
	return nil
}

// AfterUpdate Hook for Workflow to add WorkflowVersion on update
func (w *Workflow) AfterUpdate(tx *gorm.DB) (err error) {
	workflowVersion := WorkflowVersion{
		ID: "update_" + time.Now().String(),
		// ID:         uuid.New().String(),
		WorkflowID: w.ID,
		// Version:    "new_version", // 새로운 버전 설정 로직 추가 가능
		Data:      *w,
		Action:    "update",
		CreatedAt: time.Now(),
	}

	if err := tx.Create(&workflowVersion).Error; err != nil {
		return err
	}
	return nil
}

// BeforeDelete Hook for Workflow to delete WorkflowVersion on delete
func (w *Workflow) BeforeDelete(tx *gorm.DB) (err error) {
	if err := tx.Where("workflowId = ?", w.ID).Delete(&WorkflowVersion{}).Error; err != nil {
		return err
	}
	return nil
}

type WorkflowStatus struct {
	State string `json:"state"`
	Count int    `json:"count"`
}
