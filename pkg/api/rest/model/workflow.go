package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type Options struct {
	APIConnectionID string `json:"api_connection_id" mapstructure:"api_connection_id" validate:"required"`
	Endpoint        string `json:"endpoint" mapstructure:"endpoint" validate:"required"`
	Method          string `json:"method" mapstructure:"method" validate:"required"`
	RequestBody     string `json:"request_body" mapstructure:"request_body" validate:"required"`
}

type Task struct {
	ID            string   `json:"id" mapstructure:"id" validate:"required"`
	TaskComponent string   `json:"task_component" mapstructure:"task_component" validate:"required"`
	Options       Options  `json:"options" mapstructure:"options" validate:"required"`
	Dependencies  []string `json:"dependencies" mapstructure:"dependencies"`
}

type TaskGroup struct {
	ID          string `json:"id" mapstructure:"id" validate:"required"`
	Description string `json:"description" mapstructure:"description"`
	Tasks       []Task `json:"tasks" mapstructure:"tasks" validate:"required"`
}

type Data struct {
	Description string      `json:"description" mapstructure:"description"`
	TaskGroups  []TaskGroup `json:"task_groups" mapstructure:"task_groups" validate:"required"`
}

type Workflow struct {
	ID        string    `gorm:"primaryKey" json:"id" mapstructure:"id" validate:"required"`
	UUID      string    `gorm:"column:uuid" json:"-" mapstructure:"uuid"`
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
