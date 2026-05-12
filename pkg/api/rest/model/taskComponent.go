package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// Spec is a free-form key/value map persisted as JSON. Catalog
// (conf/task_types.yaml) defines which keys are valid per task type.
type Spec map[string]any

func (s Spec) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

func (s *Spec) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("invalid type for Spec")
	}
	return json.Unmarshal(bytes, s)
}

// TaskComponent is a reusable definition that workflow tasks can reference.
// Type points to a catalog entry (conf/task_types.yaml), Spec carries the
// component-level values for that type.
type TaskComponent struct {
	ID          string    `gorm:"primaryKey" json:"id" mapstructure:"id" validate:"required"`
	Name        string    `gorm:"index:,column:name,unique;type:text collate nocase" json:"name" mapstructure:"name" validate:"required"`
	Description string    `gorm:"column:description" json:"description"`
	Type        string    `gorm:"column:type;index" json:"type" mapstructure:"type" validate:"required"`
	Spec        Spec      `gorm:"column:spec;type:text" json:"spec"`
	IsExample   bool      `gorm:"column:is_example" json:"is_example" mapstructure:"is_example"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime:false" json:"created_at" mapstructure:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoCreateTime:false" json:"updated_at" mapstructure:"updated_at"`
}

type CreateTaskComponentReq struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	Type        string `json:"type" validate:"required"`
	Spec        Spec   `json:"spec"`
}
