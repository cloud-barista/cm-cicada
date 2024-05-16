package model

import "time"

type TaskTemplate struct {
	ID        string    `json:"id" mapstructure:"id" validate:"required"`
	Name      string    `json:"name" mapstructure:"name" validate:"required"`
	Data      Data      `json:"data" mapstructure:"data" validate:"required"`
	CreatedAt time.Time `json:"created_at" mapstructure:"created_at"`
	UpdatedAt time.Time `json:"updated_at" mapstructure:"updated_at"`
}
