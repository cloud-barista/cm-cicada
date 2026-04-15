package service

import (
	"errors"

	"github.com/cloud-barista/cm-cicada/dao"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/mapper"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
)

type WorkflowTaskService struct{}

func NewWorkflowTaskService() *WorkflowTaskService {
	return &WorkflowTaskService{}
}

func (s *WorkflowTaskService) ListTaskGroup(wfID string, includeDeleted bool) ([]model.TaskGroup, error) {
	workflow, err := s.getWorkflowByID(wfID, includeDeleted)
	if err != nil {
		return nil, err
	}

	taskGroups := make([]model.TaskGroup, 0, len(workflow.Data.TaskGroups))
	taskGroups = append(taskGroups, workflow.Data.TaskGroups...)

	if includeDeleted {
		taskGroupDBs, err := dao.TaskGroupGetListByWorkflowID(wfID, true)
		if err != nil {
			return nil, err
		}

		taskDBs, err := dao.TaskGetListByWorkflowID(wfID, true)
		if err != nil {
			return nil, err
		}

		taskByGroupID := make(map[string][]model.Task)
		for _, taskDB := range taskDBs {
			taskByGroupID[taskDB.TaskGroupID] = append(taskByGroupID[taskDB.TaskGroupID], s.restoreTaskFromSnapshot(taskDB))
		}

		groupByID := make(map[string]model.TaskGroup)
		for _, tg := range taskGroups {
			groupByID[tg.ID] = tg
		}

		for _, tgDB := range taskGroupDBs {
			if _, exists := groupByID[tgDB.ID]; exists {
				continue
			}
			taskGroups = append(taskGroups, model.TaskGroup{
				ID:          tgDB.ID,
				Name:        tgDB.Name,
				Description: "",
				Tasks:       normalizeTasks(taskByGroupID[tgDB.ID]),
			})
		}
	}

	return taskGroups, nil
}

func (s *WorkflowTaskService) GetTaskGroup(wfID string, tgID string, includeDeleted bool) (*model.TaskGroup, error) {
	workflow, err := s.getWorkflowByID(wfID, includeDeleted)
	if err != nil {
		return nil, err
	}

	for _, tg := range workflow.Data.TaskGroups {
		if tg.ID == tgID {
			copy := tg
			return &copy, nil
		}
	}

	if includeDeleted {
		tgDB, err := dao.TaskGroupGetIncludeDeleted(tgID)
		if err != nil {
			return nil, errors.New("task group not found")
		}
		return &model.TaskGroup{
			ID:          tgDB.ID,
			Name:        tgDB.Name,
			Description: "",
			Tasks:       []model.Task{},
		}, nil
	}

	return nil, errors.New("task group not found")
}

func (s *WorkflowTaskService) GetTaskGroupDirectly(tgID string, includeDeleted bool) (*model.TaskGroupDirectly, error) {
	var (
		tgDB *model.TaskGroupDBModel
		err  error
	)
	if includeDeleted {
		tgDB, err = dao.TaskGroupGetIncludeDeleted(tgID)
	} else {
		tgDB, err = dao.TaskGroupGet(tgID)
	}
	if err != nil {
		return nil, err
	}

	workflow, err := s.getWorkflowByID(tgDB.WorkflowID, includeDeleted)
	if err != nil {
		return nil, err
	}

	for _, tg := range workflow.Data.TaskGroups {
		if tg.ID == tgID {
			return &model.TaskGroupDirectly{
				ID:          tg.ID,
				WorkflowID:  tgDB.WorkflowID,
				Name:        tg.Name,
				Description: tg.Description,
				Tasks:       tg.Tasks,
			}, nil
		}
	}

	if includeDeleted {
		return &model.TaskGroupDirectly{
			ID:          tgDB.ID,
			WorkflowID:  tgDB.WorkflowID,
			Name:        tgDB.Name,
			Description: "",
			Tasks:       []model.Task{},
		}, nil
	}

	return nil, errors.New("task group not found")
}

func (s *WorkflowTaskService) ListTaskFromTaskGroup(wfID string, tgID string, includeDeleted bool) ([]model.Task, error) {
	workflow, err := s.getWorkflowByID(wfID, includeDeleted)
	if err != nil {
		return nil, err
	}

	var tasks []model.Task
	for _, tg := range workflow.Data.TaskGroups {
		if tg.ID == tgID {
			tasks = append(tasks, tg.Tasks...)
			break
		}
	}

	if includeDeleted {
		taskDBs, err := dao.TaskGetListByWorkflowID(wfID, true)
		if err != nil {
			return nil, err
		}
		existing := make(map[string]bool)
		for _, task := range tasks {
			existing[task.ID] = true
		}
		for _, taskDB := range taskDBs {
			if taskDB.TaskGroupID != tgID {
				continue
			}
			if existing[taskDB.ID] {
				continue
			}
			tasks = append(tasks, s.restoreTaskFromSnapshot(taskDB))
		}
	}

	return tasks, nil
}

func (s *WorkflowTaskService) GetTaskFromTaskGroup(wfID string, tgID string, taskID string, includeDeleted bool) (*model.Task, error) {
	workflow, err := s.getWorkflowByID(wfID, includeDeleted)
	if err != nil {
		return nil, err
	}

	for _, tg := range workflow.Data.TaskGroups {
		if tg.ID == tgID {
			for _, task := range tg.Tasks {
				if task.ID == taskID {
					copy := task
					return &copy, nil
				}
			}
			break
		}
	}

	if includeDeleted {
		taskDB, err := dao.TaskGetIncludeDeleted(taskID)
		if err != nil {
			return nil, errors.New("task not found")
		}
		task := s.restoreTaskFromSnapshot(*taskDB)
		return &task, nil
	}

	return nil, errors.New("task not found")
}

func (s *WorkflowTaskService) ListTask(wfID string, includeDeleted bool) ([]model.Task, error) {
	workflow, err := s.getWorkflowByID(wfID, includeDeleted)
	if err != nil {
		return nil, err
	}

	var tasks []model.Task
	for _, tg := range workflow.Data.TaskGroups {
		tasks = append(tasks, tg.Tasks...)
	}

	if includeDeleted {
		taskDBs, err := dao.TaskGetListByWorkflowID(wfID, true)
		if err != nil {
			return nil, err
		}
		existing := make(map[string]bool)
		for _, task := range tasks {
			existing[task.ID] = true
		}
		for _, taskDB := range taskDBs {
			if existing[taskDB.ID] {
				continue
			}
			tasks = append(tasks, s.restoreTaskFromSnapshot(taskDB))
		}
	}

	return tasks, nil
}

func (s *WorkflowTaskService) GetTask(wfID string, taskID string, includeDeleted bool) (*model.Task, error) {
	workflow, err := s.getWorkflowByID(wfID, includeDeleted)
	if err != nil {
		return nil, err
	}

	for _, tg := range workflow.Data.TaskGroups {
		for _, task := range tg.Tasks {
			if task.ID == taskID {
				copy := task
				return &copy, nil
			}
		}
	}

	if includeDeleted {
		taskDB, err := dao.TaskGetIncludeDeleted(taskID)
		if err != nil {
			return nil, errors.New("task not found")
		}
		task := s.restoreTaskFromSnapshot(*taskDB)
		return &task, nil
	}

	return nil, errors.New("task not found")
}

func (s *WorkflowTaskService) GetTaskDirectly(taskID string, includeDeleted bool) (*model.TaskDirectly, error) {
	var (
		tDB *model.TaskDBModel
		err error
	)
	if includeDeleted {
		tDB, err = dao.TaskGetIncludeDeleted(taskID)
	} else {
		tDB, err = dao.TaskGet(taskID)
	}
	if err != nil {
		return nil, err
	}

	var tgDB *model.TaskGroupDBModel
	if includeDeleted {
		tgDB, err = dao.TaskGroupGetIncludeDeleted(tDB.TaskGroupID)
	} else {
		tgDB, err = dao.TaskGroupGet(tDB.TaskGroupID)
	}
	if err != nil {
		return nil, err
	}

	workflow, err := s.getWorkflowByID(tgDB.WorkflowID, includeDeleted)
	if err != nil {
		return nil, err
	}

	for _, tg := range workflow.Data.TaskGroups {
		if tg.ID == tgDB.ID {
			for _, task := range tg.Tasks {
				if task.ID == taskID {
					return &model.TaskDirectly{
						ID:            task.ID,
						WorkflowID:    tDB.WorkflowID,
						TaskGroupID:   tDB.TaskGroupID,
						Name:          task.Name,
						TaskComponent: task.TaskComponent,
						RequestBody:   task.RequestBody,
						PathParams:    task.PathParams,
						QueryParams:   task.QueryParams,
						Extra:         task.Extra,
						Dependencies:  task.Dependencies,
					}, nil
				}
			}
		}
	}

	if includeDeleted {
		task := s.restoreTaskFromSnapshot(*tDB)
		return &model.TaskDirectly{
			ID:            tDB.ID,
			WorkflowID:    tDB.WorkflowID,
			TaskGroupID:   tDB.TaskGroupID,
			Name:          task.Name,
			TaskComponent: task.TaskComponent,
			RequestBody:   task.RequestBody,
			PathParams:    task.PathParams,
			QueryParams:   task.QueryParams,
			Extra:         task.Extra,
			Dependencies:  task.Dependencies,
		}, nil
	}

	return nil, errors.New("task not found")
}

func (s *WorkflowTaskService) getWorkflowByID(wfID string, includeDeleted bool) (*model.Workflow, error) {
	if includeDeleted {
		return mapper.GetWorkflowFromDBIncludeDeleted(wfID)
	}
	return mapper.GetWorkflowFromDB(wfID)
}

func (s *WorkflowTaskService) restoreTaskFromSnapshot(taskDB model.TaskDBModel) model.Task {
	task := model.Task{
		ID:            taskDB.ID,
		Name:          taskDB.Name,
		TaskComponent: "",
		RequestBody:   "",
		PathParams:    nil,
		QueryParams:   nil,
		Extra:         nil,
		Dependencies:  []string{},
		IsDeletedTask: taskDB.IsDeleted,
	}

	snapshotTask, err := dao.TaskSnapshotGetLatestRawTask(taskDB.WorkflowID, taskDB.ID)
	if err != nil || snapshotTask == nil {
		return task
	}

	restored := *snapshotTask
	restored.ID = taskDB.ID
	if restored.Name == "" {
		restored.Name = taskDB.Name
	}
	if restored.Dependencies == nil {
		restored.Dependencies = []string{}
	}
	restored.IsDeletedTask = taskDB.IsDeleted
	return restored
}

func normalizeTasks(tasks []model.Task) []model.Task {
	if tasks == nil {
		return []model.Task{}
	}
	return tasks
}
