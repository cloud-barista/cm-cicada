package mapper

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/cloud-barista/cm-cicada/dao"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/google/uuid"
	"github.com/jollaman999/utils/logger"
)

type WorkflowGraphDiff struct {
	WorkflowData         model.Data
	TaskGroupsToUpsert   []model.TaskGroupDBModel
	TasksToUpsert        []model.TaskDBModel
	TaskGroupsToSoftDrop []model.TaskGroupDBModel
	TasksToSoftDrop      []model.TaskDBModel
}

func CreateDataReqToData(specVersion string, createDataReq model.CreateDataReq) (model.Data, error) {
	specVersionSpilit := strings.Split(specVersion, ".")
	if len(specVersionSpilit) != 2 {
		return model.Data{}, errors.New("invalid workflow spec version: " + specVersion)
	}

	specVersionMajor, err := strconv.Atoi(specVersionSpilit[0])
	if err != nil {
		return model.Data{}, errors.New("invalid workflow spec version: " + specVersion)
	}

	specVersionMinor, err := strconv.Atoi(specVersionSpilit[1])
	if err != nil {
		return model.Data{}, errors.New("invalid workflow spec version: " + specVersion)
	}

	var taskGroups []model.TaskGroup
	var allTasks []model.Task

	if specVersionMajor > 0 && specVersionMajor <= 1 {
		if specVersionMinor == 0 {
			// v1.0
			for _, tgReq := range createDataReq.TaskGroups {
				var tasks []model.Task
				for _, tReq := range tgReq.Tasks {
					tasks = append(tasks, model.Task{
						ID:            uuid.New().String(),
						Name:          tReq.Name,
						TaskComponent: tReq.TaskComponent,
						RequestBody:   tReq.RequestBody,
						PathParams:    tReq.PathParams,
						QueryParams:   tReq.QueryParams,
						Extra:         tReq.Extra,
						Dependencies:  tReq.Dependencies,
					})
				}

				allTasks = append(allTasks, tasks...)
				taskGroups = append(taskGroups, model.TaskGroup{
					ID:          uuid.New().String(),
					Name:        tgReq.Name,
					Description: tgReq.Description,
					Tasks:       tasks,
				})
			}

			for i, tgReq := range createDataReq.TaskGroups {
				for j, tg := range taskGroups {
					if tgReq.Name == tg.Name {
						if i == j {
							continue
						}

						return model.Data{}, errors.New("Duplicated task group name: " + tg.Name)
					}
				}
			}

			for i, tCheck := range allTasks {
				for j, t := range allTasks {
					if tCheck.Name == t.Name {
						if i == j {
							continue
						}

						return model.Data{}, errors.New("Duplicated task name: " + t.Name)
					}
				}
			}
		} else {
			return model.Data{}, errors.New("Unsupported workflow spec version: " + specVersion)
		}
	} else {
		return model.Data{}, errors.New("Unsupported workflow spec version: " + specVersion)
	}

	return model.Data{
		Description: createDataReq.Description,
		TaskGroups:  taskGroups,
	}, nil
}

func BuildWorkflowGraphDiff(workflow *model.Workflow, incoming model.Data) (*WorkflowGraphDiff, error) {
	workflowKey := workflowDagID(workflow)
	taskGroupsFromDB, err := dao.TaskGroupGetListByWorkflowID(workflow.ID, true)
	if err != nil {
		return nil, err
	}
	tasksFromDB, err := dao.TaskGetListByWorkflowID(workflow.ID, true)
	if err != nil {
		return nil, err
	}

	taskGroupByName := make(map[string]model.TaskGroupDBModel)
	activeTaskGroups := make(map[string]model.TaskGroupDBModel)
	for _, tg := range taskGroupsFromDB {
		current, exists := taskGroupByName[tg.Name]
		if !exists || (current.IsDeleted && !tg.IsDeleted) {
			taskGroupByName[tg.Name] = tg
		}
		if !tg.IsDeleted {
			activeTaskGroups[tg.ID] = tg
		}
	}

	taskByName := make(map[string]model.TaskDBModel)
	activeTasks := make(map[string]model.TaskDBModel)
	for _, t := range tasksFromDB {
		current, exists := taskByName[t.Name]
		if !exists || (current.IsDeleted && !t.IsDeleted) {
			taskByName[t.Name] = t
		}
		if !t.IsDeleted {
			activeTasks[t.ID] = t
		}
	}

	diff := &WorkflowGraphDiff{
		WorkflowData: model.Data{
			Description: incoming.Description,
			TaskGroups:  make([]model.TaskGroup, 0, len(incoming.TaskGroups)),
		},
	}
	seenTaskGroupIDs := make(map[string]bool)
	seenTaskIDs := make(map[string]bool)

	for _, incomingTG := range incoming.TaskGroups {
		resolvedTG := incomingTG
		taskGroupModel, exists := taskGroupByName[incomingTG.Name]
		if !exists {
			taskGroupModel = model.TaskGroupDBModel{
				ID:           uuid.New().String(),
				TaskGroupKey: uuid.New().String(),
			}
		}
		if taskGroupModel.TaskGroupKey == "" {
			taskGroupModel.TaskGroupKey = taskGroupModel.ID
		}
		resolvedTG.ID = taskGroupModel.ID
		resolvedTG.Tasks = make([]model.Task, 0, len(incomingTG.Tasks))

		taskGroupModel.Name = incomingTG.Name
		taskGroupModel.WorkflowID = workflow.ID
		taskGroupModel.WorkflowKey = workflowKey
		taskGroupModel.IsDeleted = false
		diff.TaskGroupsToUpsert = append(diff.TaskGroupsToUpsert, taskGroupModel)
		seenTaskGroupIDs[taskGroupModel.ID] = true

		for _, incomingTask := range incomingTG.Tasks {
			resolvedTask := incomingTask
			taskModel, exists := taskByName[incomingTask.Name]
			if !exists {
				taskModel = model.TaskDBModel{
					ID:      uuid.New().String(),
					TaskKey: uuid.New().String(),
				}
			}
			if taskModel.TaskKey == "" {
				taskModel.TaskKey = taskModel.ID
			}

			resolvedTask.ID = taskModel.ID
			resolvedTG.Tasks = append(resolvedTG.Tasks, resolvedTask)

			taskModel.Name = incomingTask.Name
			taskModel.WorkflowID = workflow.ID
			taskModel.WorkflowKey = workflowKey
			taskModel.TaskGroupID = taskGroupModel.ID
			taskModel.TaskGroupKey = taskGroupModel.TaskGroupKey
			taskModel.IsDeleted = false
			diff.TasksToUpsert = append(diff.TasksToUpsert, taskModel)
			seenTaskIDs[taskModel.ID] = true
		}

		diff.WorkflowData.TaskGroups = append(diff.WorkflowData.TaskGroups, resolvedTG)
	}

	for _, tg := range activeTaskGroups {
		if !seenTaskGroupIDs[tg.ID] {
			diff.TaskGroupsToSoftDrop = append(diff.TaskGroupsToSoftDrop, tg)
		}
	}
	for _, t := range activeTasks {
		if !seenTaskIDs[t.ID] {
			diff.TasksToSoftDrop = append(diff.TasksToSoftDrop, t)
		}
	}

	return diff, nil
}

func ResolveCreateSourceType(specVersion string, createDataReq model.CreateDataReq) (string, string, error) {
	templates, err := dao.WorkflowTemplateGetList(&model.WorkflowTemplate{}, 0, 0)
	if err != nil {
		return "", "", err
	}

	reqJSON, err := json.Marshal(createDataReq)
	if err != nil {
		return "", "", err
	}

	for _, tmpl := range *templates {
		if tmpl.SpecVersion != specVersion {
			continue
		}

		tmplJSON, err := json.Marshal(tmpl.Data)
		if err != nil {
			return "", "", err
		}

		if string(reqJSON) == string(tmplJSON) {
			return "example", tmpl.ID, nil
		}
	}

	return "custom", "", nil
}

func GetWorkflowFromDB(workflowID string) (*model.Workflow, error) {
	workflow, err := dao.WorkflowGet(workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get the workflow from DB. Error: %s", err.Error())
	}

	for i, tg := range workflow.Data.TaskGroups {
		tgDB, err := dao.TaskGroupGetByWorkflowIDAndName(workflowID, tg.Name)
		if err != nil {
			logger.Println(logger.ERROR, true, err)
		} else {
			workflow.Data.TaskGroups[i].ID = tgDB.ID
		}

		for j, t := range tg.Tasks {
			tDB, err := dao.TaskGetByWorkflowIDAndName(workflowID, t.Name)
			if err != nil {
				logger.Println(logger.ERROR, true, err)
			} else {
				workflow.Data.TaskGroups[i].Tasks[j].ID = tDB.ID
			}
		}
	}

	return workflow, nil
}

func GetWorkflowFromDBIncludeDeleted(workflowID string) (*model.Workflow, error) {
	workflow, err := dao.WorkflowGetIncludeDeleted(workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get the workflow from DB. Error: %s", err.Error())
	}

	for i, tg := range workflow.Data.TaskGroups {
		tgDB, err := dao.TaskGroupGetByWorkflowIDAndNameIncludeDeleted(workflowID, tg.Name)
		if err != nil {
			logger.Println(logger.ERROR, true, err)
		} else {
			workflow.Data.TaskGroups[i].ID = tgDB.ID
		}

		for j, t := range tg.Tasks {
			tDB, err := dao.TaskGetByWorkflowIDAndNameIncludeDeleted(workflowID, t.Name)
			if err != nil {
				logger.Println(logger.ERROR, true, err)
			} else {
				workflow.Data.TaskGroups[i].Tasks[j].ID = tDB.ID
			}
		}
	}

	return workflow, nil
}

func workflowDagID(workflow *model.Workflow) string {
	if workflow.WorkflowKey != "" {
		return workflow.WorkflowKey
	}
	return workflow.ID
}
