package service

import (
	"fmt"

	"github.com/cloud-barista/cm-cicada/dao"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
)

// TaskGroupService interface defines the contract for task group business logic
type TaskGroupService interface {
	ListTaskGroup(wfId string) ([]model.TaskGroup, error)
	GetTaskGroup(wfId, tgId string) (*model.TaskGroup, error)
	GetTaskGroupDirectly(tgId string) (*model.TaskGroupDirectly, error)
	ListTaskFromTaskGroup(wfId, tgId string) ([]model.Task, error)
	GetTaskFromTaskGroup(wfId, tgId, taskId string) (*model.Task, error)
}

// taskGroupService is the concrete implementation of TaskGroupService
type taskGroupService struct{}

// NewTaskGroupService creates a new instance of TaskGroupService
func NewTaskGroupService() TaskGroupService {
	return &taskGroupService{}
}

// ListTaskGroup retrieves all task groups for a workflow
func (s *taskGroupService) ListTaskGroup(wfId string) ([]model.TaskGroup, error) {
	if wfId == "" {
		return nil, fmt.Errorf("please provide the wfId")
	}

	workflow, err := dao.WorkflowGet(wfId)
	if err != nil {
		return nil, err
	}

	var taskGroups []model.TaskGroup
	taskGroups = append(taskGroups, workflow.Data.TaskGroups...)

	return taskGroups, nil
}

// GetTaskGroup retrieves a specific task group from a workflow
func (s *taskGroupService) GetTaskGroup(wfId, tgId string) (*model.TaskGroup, error) {
	if wfId == "" {
		return nil, fmt.Errorf("please provide the wfId")
	}
	if tgId == "" {
		return nil, fmt.Errorf("please provide the tgId")
	}

	workflow, err := dao.WorkflowGet(wfId)
	if err != nil {
		return nil, err
	}

	for _, tg := range workflow.Data.TaskGroups {
		if tg.ID == tgId {
			return &tg, nil
		}
	}

	return nil, fmt.Errorf("task group not found")
}

// GetTaskGroupDirectly retrieves task group information directly from database
func (s *taskGroupService) GetTaskGroupDirectly(tgId string) (*model.TaskGroupDirectly, error) {
	if tgId == "" {
		return nil, fmt.Errorf("please provide the tgId")
	}

	tgDB, err := dao.TaskGroupGet(tgId)
	if err != nil {
		return nil, err
	}

	workflow, err := dao.WorkflowGet(tgDB.WorkflowID)
	if err != nil {
		return nil, err
	}

	for _, tg := range workflow.Data.TaskGroups {
		if tg.ID == tgId {
			return &model.TaskGroupDirectly{
				ID:          tg.ID,
				WorkflowID:  tgDB.WorkflowID,
				Name:        tg.Name,
				Description: tg.Description,
				Tasks:       tg.Tasks,
			}, nil
		}
	}

	return nil, fmt.Errorf("task group not found")
}

// ListTaskFromTaskGroup retrieves all tasks from a specific task group
func (s *taskGroupService) ListTaskFromTaskGroup(wfId, tgId string) ([]model.Task, error) {
	if wfId == "" {
		return nil, fmt.Errorf("please provide the wfId")
	}
	if tgId == "" {
		return nil, fmt.Errorf("please provide the tgId")
	}

	workflow, err := dao.WorkflowGet(wfId)
	if err != nil {
		return nil, err
	}

	var tasks []model.Task
	for _, tg := range workflow.Data.TaskGroups {
		if tg.ID == tgId {
			tasks = append(tasks, tg.Tasks...)
			break
		}
	}

	return tasks, nil
}

// GetTaskFromTaskGroup retrieves a specific task from a task group
func (s *taskGroupService) GetTaskFromTaskGroup(wfId, tgId, taskId string) (*model.Task, error) {
	if wfId == "" {
		return nil, fmt.Errorf("please provide the wfId")
	}
	if tgId == "" {
		return nil, fmt.Errorf("please provide the tgId")
	}
	if taskId == "" {
		return nil, fmt.Errorf("please provide the taskId")
	}

	workflow, err := dao.WorkflowGet(wfId)
	if err != nil {
		return nil, err
	}

	for _, tg := range workflow.Data.TaskGroups {
		if tg.ID == tgId {
			for _, task := range tg.Tasks {
				if task.ID == taskId {
					return &task, nil
				}
			}
			break
		}
	}

	return nil, fmt.Errorf("task not found")
}
