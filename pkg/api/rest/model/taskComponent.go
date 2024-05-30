package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type TaskComponent struct {
	ID        string    `gorm:"primaryKey" json:"id" mapstructure:"id" validate:"required"`
	Data      TaskData  `gorm:"column:data" json:"data" mapstructure:"data" validate:"required"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at" mapstructure:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at" mapstructure:"updated_at"`
}

type TaskData struct {
	TaskName        string          `json:"task_name" mapstructure:"task_name" validate:"required"`
	OperatorOptions OperatorOptions `json:"operator_options" mapstructure:"operator_options" validate:"required"`
	ParmaOption     ParmaOption     `json:"param_option" mapstructure:"param_option" validate:"required"`
}
type ParmaOption struct {
	Params Params `json:"params" mapstructure:"params" validate:"required"`
}

type Params struct {
	Required   []string    `json:"required" mapstructure:"required" validate:"required"`
	Properties interface{} `json:"properties" mapstructure:"properties" validate:"required"`
}

func (d TaskData) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *TaskData) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("Invalid type for TaskData")
	}
	return json.Unmarshal(bytes, d)
}
