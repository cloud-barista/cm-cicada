package service

import (
	"errors"

	"github.com/cloud-barista/cm-cicada/dao"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
)

type WorkflowTemplateService struct{}

func NewWorkflowTemplateService() *WorkflowTemplateService {
	return &WorkflowTemplateService{}
}

func (s *WorkflowTemplateService) Get(id string) (*model.GetWorkflowTemplate, error) {
	return dao.WorkflowTemplateGet(id)
}

func (s *WorkflowTemplateService) GetByName(name string) (*model.GetWorkflowTemplate, error) {
	wt := dao.WorkflowTemplateGetByName(name)
	if wt == nil {
		return nil, errors.New("workflow template not found with the provided name")
	}
	return &model.GetWorkflowTemplate{
		SpecVersion: wt.SpecVersion,
		Name:        wt.Name,
		Data:        wt.Data,
	}, nil
}

func (s *WorkflowTemplateService) List(name string, page, row int) (*[]model.WorkflowTemplate, error) {
	filter := &model.WorkflowTemplate{Name: name}
	return dao.WorkflowTemplateGetList(filter, page, row)
}
