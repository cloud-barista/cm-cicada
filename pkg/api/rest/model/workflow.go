package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type DefaultArgs struct {
	Owner          string `json:"owner" mapstructure:"owner" validate:"required"`
	StartDate      string `json:"start_date" mapstructure:"start_date" validate:"required"`
	Retries        int    `json:"retries" mapstructure:"retries"`                 // default: 1
	RetryDelaySec  int    `json:"retry_delay_sec" mapstructure:"retry_delay_sec"` // default: 300
	Email          string `json:"email" mapstructure:"email"`
	EmailOnFailure bool   `json:"email_on_failure" mapstructure:"email_on_failure"`
	EmailOnRetry   bool   `json:"email_on_retry" mapstructure:"email_on_retry"`
}

type OperatorOptions []struct {
	Name  string `json:"name" mapstructure:"name" validate:"required"`
	Value any    `json:"value" mapstructure:"value" validate:"required"`
}

type Task struct {
	TaskName        string          `json:"task_name" mapstructure:"task_name" validate:"required"`
	TaskComponent   string          `json:"task_component" mapstructure:"task_component" validate:"required"`
	Operator        string          `json:"operator" mapstructure:"operator" validate:"required"`
	OperatorOptions OperatorOptions `json:"operator_options" mapstructure:"operator_options" validate:"required"`
	Dependencies    []string        `json:"dependencies" mapstructure:"dependencies" validate:"required"`
}

type TaskGroup struct {
	TaskGroupName string `json:"task_group_name" mapstructure:"task_group_name" validate:"required"`
	Description   string `json:"description" mapstructure:"description"`
	Tasks         []Task `json:"tasks" mapstructure:"tasks" validate:"required"`
}

type Data struct {
	DefaultArgs DefaultArgs `json:"default_args" mapstructure:"default_args" validate:"required"`
	Description string      `json:"description" mapstructure:"description"`
	TaskGroups  []TaskGroup `json:"task_groups" mapstructure:"task_groups" validate:"required"`
}

type Workflow struct {
	ID        string    `json:"id" mapstructure:"id" validate:"required"`
	Name      string    `json:"name" mapstructure:"name" validate:"required"`
	Data      Data      `json:"data" mapstructure:"data" validate:"required"`
	CreatedAt time.Time `json:"created_at" mapstructure:"created_at"`
	UpdatedAt time.Time `json:"updated_at" mapstructure:"updated_at"`
}
func (d DefaultArgs) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *DefaultArgs) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("Invalid type for DefaultArgs")
	}
	return json.Unmarshal(bytes, d)
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
		return errors.New("Invalid type for Data")
	}
	return json.Unmarshal(bytes, d)
}
