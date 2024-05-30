package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type OperatorOptions []struct {
	Name  string `json:"name" mapstructure:"name" validate:"required"`
	Value any    `json:"value" mapstructure:"value" validate:"required"`
}

type Task struct {
	TaskName        string          `json:"task_name" mapstructure:"task_name" validate:"required"`
	TaskComponent   string          `json:"task_component" mapstructure:"task_component" validate:"required"`
	Operator        string          `json:"operator" mapstructure:"operator" validate:"required"`
	OperatorOptions OperatorOptions `json:"operator_options" mapstructure:"operator_options" validate:"required"`
	Dependencies    []string        `json:"dependencies" mapstructure:"dependencies"`
}

type TaskGroup struct {
	TaskGroupName string `json:"task_group_name" mapstructure:"task_group_name" validate:"required"`
	Description   string `json:"description" mapstructure:"description"`
	Tasks         []Task `json:"tasks" mapstructure:"tasks" validate:"required"`
}

type Data struct {
	Description string      `json:"description" mapstructure:"description"`
	TaskGroups  []TaskGroup `json:"task_groups" mapstructure:"task_groups" validate:"required"`
}

type Workflow struct {
	ID        string    `gorm:"primaryKey" json:"id" mapstructure:"id" validate:"required"`
	Data      Data      `gorm:"column:data" json:"data" mapstructure:"data" validate:"required"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at" mapstructure:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at" mapstructure:"updated_at"`
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
