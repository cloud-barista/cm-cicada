package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

const (
	WorkflowSpecVersion_1_0 = "1.0"
)

const (
	WorkflowSpecVersion_LATEST = WorkflowSpecVersion_1_0
)

// Task is the domain/API representation of a task inside a workflow graph.
// Kept separate from TaskDBModel so API payloads and DB rows can diverge
// without contaminating each other's tag sets.
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

// TaskDirectly is a flat variant used by the "direct create" path where a
// task is created with its workflow/task-group IDs embedded in the payload.
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

type CreateTaskGroupReq struct {
	Name        string          `json:"name" mapstructure:"name" validate:"required"`
	Description string          `json:"description" mapstructure:"description"`
	Tasks       []CreateTaskReq `json:"tasks" mapstructure:"tasks" validate:"required"`
}

// Data is the graph container serialized into the `data` column of the
// workflows table (JSON-encoded via the Valuer/Scanner pair below).
type Data struct {
	Description string      `json:"description" mapstructure:"description"`
	TaskGroups  []TaskGroup `json:"task_groups" mapstructure:"task_groups" validate:"required"`
}

type CreateDataReq struct {
	Description string               `json:"description" mapstructure:"description"`
	TaskGroups  []CreateTaskGroupReq `json:"task_groups" mapstructure:"task_groups" validate:"required"`
}

type CreateWorkflowReq struct {
	SpecVersion string        `json:"spec_version" mapstructure:"spec_version"`
	Name        string        `json:"name" mapstructure:"name" validate:"required"`
	Data        CreateDataReq `json:"data" mapstructure:"data" validate:"required"`
}

// --- driver.Valuer / sql.Scanner for JSON-encoded columns ---

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
