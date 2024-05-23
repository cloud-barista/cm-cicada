package model

import (
	"time"
)

type WorkflowTemplate struct {
	ID        string    `gorm:"primaryKey" json:"id" mapstructure:"id" validate:"required"`
	Data      Data      `gorm:"column:data" json:"data" mapstructure:"data" validate:"required"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at" mapstructure:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at" mapstructure:"updated_at"`
}
