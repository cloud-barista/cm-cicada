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

type ParamOption struct {
	Params     Params `json:"params" mapstructure:"params" validate:"required"`
	PathParams Params `json:"path_params" mapstructure:"path_params"`
}

type TaskData struct {
	Options     Options     `json:"options" mapstructure:"options" `
	ParmaOption ParamOption `json:"param_option" mapstructure:"param_option"`
	Extra 			map[string]interface{} `json:"extra,omitempty" mapstructure:"extra"`
}

type TaskComponent struct {
	ID        string    `gorm:"primaryKey" json:"id" mapstructure:"id" validate:"required"`
	Name      string    `gorm:"index:,column:name,unique;type:text collate nocase" json:"name" mapstructure:"name" validate:"required"`
	Data      TaskData  `gorm:"column:data" json:"data" mapstructure:"data" validate:"required"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at" mapstructure:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at" mapstructure:"updated_at"`
}

type CreateTaskComponentReq struct {
	Name string   `json:"name" mapstructure:"name" validate:"required"`
	Data TaskData `gorm:"column:data" json:"data" mapstructure:"data" validate:"required"`
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
		return errors.New("Invalid type for Data")
	}
	return json.Unmarshal(bytes, d)
}
