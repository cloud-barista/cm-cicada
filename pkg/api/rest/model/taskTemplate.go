package model

import "time"

type TaskTemplate struct {
	ID        string    `json:"id" mapstructure:"id" validate:"required"`
	Name      string    `json:"name" mapstructure:"name" validate:"required"`
	Task      Task      `json:"task" mapstructure:"task" validate:"required"`
	CreatedAt time.Time `json:"created_at" mapstructure:"created_at"`
	UpdatedAt time.Time `json:"updated_at" mapstructure:"updated_at"`
}
