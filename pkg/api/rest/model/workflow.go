package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
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
	IsDeletedTask bool                   `json:"is_deleted_task"`
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

type CreateTaskReq struct {
	Name          string                 `json:"name" mapstructure:"name" validate:"required"`
	TaskComponent string                 `json:"task_component" mapstructure:"task_component" validate:"required"`
	RequestBody   string                 `json:"request_body" mapstructure:"request_body" validate:"required"`
	PathParams    map[string]string      `json:"path_params" mapstructure:"path_params"`
	QueryParams   map[string]string      `json:"query_params" mapstructure:"query_params"`
	Extra         map[string]interface{} `json:"extra,omitempty" mapstructure:"extra"`
	Dependencies  []string               `json:"dependencies" mapstructure:"dependencies"`
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
	ID           string     `gorm:"primaryKey" json:"id" mapstructure:"id" validate:"required"`
	Name         string     `json:"name" mapstructure:"name" validate:"required"`
	WorkflowID   string     `gorm:"column:workflow_id;index" json:"workflow_id" mapstructure:"workflow_id" validate:"required"`
	WorkflowKey  string     `gorm:"column:workflow_key;index" json:"-" mapstructure:"-"`
	TaskGroupKey string     `gorm:"column:task_group_key;index" json:"-" mapstructure:"-"`
	IsDeleted    bool       `gorm:"column:is_deleted;index;default:false" json:"-" mapstructure:"-"`
	DeletedAt    *time.Time `gorm:"column:deleted_at" json:"-" mapstructure:"-"`
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

type CreateWorkflowReq struct {
	SpecVersion string        `json:"spec_version" mapstructure:"spec_version"`
	Name        string        `json:"name" mapstructure:"name" validate:"required"`
	Data        CreateDataReq `json:"data" mapstructure:"data" validate:"required"`
}

type WorkflowRun struct {
	WorkflowRunID          string                 `json:"workflow_run_id,omitempty"`
	WorkflowID             *string                `json:"workflow_id,omitempty"`
	DagID                  *string                `json:"dag_id,omitempty"`
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
	WorkflowRunID                string    `json:"workflow_run_id,omitempty"`
	WorkflowID                   *string   `json:"workflow_id,omitempty"`
	DagID                        *string   `json:"dag_id,omitempty"`
	IsDeletedTask                bool      `json:"is_deleted_task"`
	TaskID                       string    `json:"task_id,omitempty"`
	TaskName                     string    `json:"task_name,omitempty"`
	State                        string    `json:"state,omitempty"`
	StartDate                    time.Time `json:"start_date,omitempty"`
	EndDate                      time.Time `json:"end_date,omitempty"`
	DurationDate                 float64   `json:"duration_date"`
	ExecutionDate                time.Time `json:"execution_date,omitempty"`
	TryNumber                    int       `json:"try_number"`
	IsSoftwareMigrationTask      bool      `json:"is_software_migration_task"`
	SoftwareMigrationExecutionID string    `json:"software_migration_execution_id,omitempty"`
}

type TaskInstanceReference struct {
	// The task ID.
	TaskId   *string `json:"task_id,omitempty"`
	TaskName string  `json:"task_name,omitempty"`
	// DB workflow ID.
	WorkflowID *string `json:"workflow_id,omitempty"`
	// The DAG ID.
	DagID *string `json:"dag_id,omitempty"`
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
	IsDeletedTask bool      `json:"is_deleted_task"`
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

type WorkflowStatus struct {
	State string `json:"state"`
	Count int    `json:"count"`
}
