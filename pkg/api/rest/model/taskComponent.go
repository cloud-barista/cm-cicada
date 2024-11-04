package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type PropertyDef struct {
	Type        string                 `json:"type"`
	Required    []string               `json:"required,omitempty"`
	Properties  map[string]PropertyDef `json:"properties,omitempty"`
	Items       *PropertyDef           `json:"items,omitempty"`
	Description string                 `json:"description,omitempty"`
	Default     interface{}            `json:"default,omitempty"`
	Enum        []string               `json:"enum,omitempty"`
	Example     interface{}            `json:"example,omitempty"`
}

type ParameterStructure struct {
	Required   []string               `json:"required,omitempty"`
	Properties map[string]PropertyDef `json:"properties,omitempty"`
}

type TaskComponentOptions struct {
	APIConnectionID string `json:"api_connection_id"`
	Endpoint        string `json:"endpoint"`
	Method          string `json:"method"`
	RequestBody     string `json:"request_body"`
}

type TaskComponentData struct {
	Options     TaskComponentOptions `json:"options"`
	Extra 			map[string]interface{} `json:"extra,omitempty"`
	BodyParams  ParameterStructure   `json:"body_params,omitempty"`
	PathParams  ParameterStructure   `json:"path_params,omitempty"`
	QueryParams ParameterStructure   `json:"query_params,omitempty"`
}

type TaskComponent struct {
	ID          string            `gorm:"primaryKey" json:"id" mapstructure:"id" validate:"required"`
	Name        string            `gorm:"index:,column:name,unique;type:text collate nocase" json:"name" mapstructure:"name" validate:"required"`
	Description string            `gorm:"column:description" json:"description"`
	Data        TaskComponentData `gorm:"column:data" json:"data" mapstructure:"data" validate:"required"`
	CreatedAt   time.Time         `gorm:"column:created_at;autoCreateTime:false" json:"created_at" mapstructure:"created_at"`
	UpdatedAt   time.Time         `gorm:"column:updated_at;autoCreateTime:false" json:"updated_at" mapstructure:"updated_at"`
	IsExample   bool              `gorm:"column:is_example" json:"is_example" mapstructure:"is_example"`
}

type CreateTaskComponentReq struct {
	Name string            `json:"name" mapstructure:"name" validate:"required"`
	Data TaskComponentData `gorm:"column:data" json:"data" mapstructure:"data" validate:"required"`
}

type Params struct {
	Required   []string    `json:"required" mapstructure:"required" validate:"required"`
	Properties interface{} `json:"properties" mapstructure:"properties" validate:"required"`
}

func (d TaskComponentData) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *TaskComponentData) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("Invalid type for Data")
	}
	return json.Unmarshal(bytes, d)
}
