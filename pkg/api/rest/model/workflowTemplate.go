package model

import (
	"time"
)

type WorkflowTemplate struct {
	UUID      string    `gorm:"primaryKey" json:"uuid" mapstructure:"uuid" validate:"required"`
	Name      string    `gorm:"column:name" json:"name" mapstructure:"name" validate:"required"`
	Data      Data      `gorm:"column:data" json:"data" mapstructure:"data" validate:"required"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at" mapstructure:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at" mapstructure:"updated_at"`
}
