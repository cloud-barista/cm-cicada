package model

type WorkflowTemplate struct {
	ID   string        `gorm:"primaryKey" json:"id" mapstructure:"id" validate:"required"`
	Name string        `gorm:"unique,column:name" json:"name" mapstructure:"name" validate:"required"`
	Data CreateDataReq `gorm:"column:data" json:"data" mapstructure:"data" validate:"required"`
}
