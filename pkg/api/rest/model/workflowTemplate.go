package model

type GetWorkflowTemplate struct {
	Name string        `gorm:"index:,column:name,unique;type:text collate nocase" json:"name" mapstructure:"name" validate:"required"`
	Data CreateDataReq `gorm:"column:data" json:"data" mapstructure:"data" validate:"required"`
}

type WorkflowTemplate struct {
	ID   string        `gorm:"primaryKey" json:"id" mapstructure:"id" validate:"required"`
	Name string        `gorm:"index:,column:name,unique;type:text collate nocase" json:"name" mapstructure:"name" validate:"required"`
	Data CreateDataReq `gorm:"column:data" json:"data" mapstructure:"data" validate:"required"`
}
