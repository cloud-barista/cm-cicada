package service

import (
	"fmt"
	"time"

	"github.com/cloud-barista/cm-cicada/dao"
	"github.com/cloud-barista/cm-cicada/lib/airflow"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/jollaman999/utils/logger"
)

// TaskService interface defines the contract for task business logic
type TaskService interface {
	ListTask(wfId string) ([]model.Task, error)
	GetTask(wfId, taskId string) (*model.Task, error)
	GetTaskDirectly(taskId string) (*model.TaskDirectly, error)
	GetTaskLogs(wfId, wfRunId, taskId string, taskTryNum int) (*model.TaskLog, error)
	GetTaskLogDownload(wfId, wfRunId, taskId string, taskTryNum int) ([]byte, string, error)
	GetTaskInstances(wfId, wfRunId string) ([]model.TaskInstance, error)
	ClearTaskInstances(wfId, wfRunId string, option model.TaskClearOption) ([]model.TaskInstanceReference, error)
}

// taskService is the concrete implementation of TaskService
type taskService struct{}

// NewTaskService creates a new instance of TaskService
func NewTaskService() TaskService {
	return &taskService{}
}

// ListTask retrieves all tasks for a workflow
func (s *taskService) ListTask(wfId string) ([]model.Task, error) {
	if wfId == "" {
		return nil, fmt.Errorf("please provide the wfId")
	}

	workflow, err := dao.WorkflowGet(wfId)
	if err != nil {
		return nil, err
	}

	var tasks []model.Task
	for _, tg := range workflow.Data.TaskGroups {
		tasks = append(tasks, tg.Tasks...)
	}

	return tasks, nil
}

// GetTask retrieves a specific task from a workflow
func (s *taskService) GetTask(wfId, taskId string) (*model.Task, error) {
	if wfId == "" {
		return nil, fmt.Errorf("please provide the wfId")
	}
	if taskId == "" {
		return nil, fmt.Errorf("please provide the taskId")
	}

	workflow, err := dao.WorkflowGet(wfId)
	if err != nil {
		return nil, err
	}

	for _, tg := range workflow.Data.TaskGroups {
		for _, task := range tg.Tasks {
			if task.ID == taskId {
				return &task, nil
			}
		}
	}

	return nil, fmt.Errorf("task not found")
}

// GetTaskDirectly retrieves task information directly from database
func (s *taskService) GetTaskDirectly(taskId string) (*model.TaskDirectly, error) {
	if taskId == "" {
		return nil, fmt.Errorf("please provide the taskId")
	}

	tDB, err := dao.TaskGet(taskId)
	if err != nil {
		return nil, err
	}

	tgDB, err := dao.TaskGroupGet(tDB.TaskGroupID)
	if err != nil {
		return nil, err
	}

	workflow, err := dao.WorkflowGet(tgDB.WorkflowID)
	if err != nil {
		return nil, err
	}

	for _, tg := range workflow.Data.TaskGroups {
		if tg.ID == tgDB.ID {
			for _, task := range tg.Tasks {
				if task.ID == taskId {
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

	return nil, fmt.Errorf("task not found")
}

// GetTaskLogs retrieves logs for a specific task execution
func (s *taskService) GetTaskLogs(wfId, wfRunId, taskId string, taskTryNum int) (*model.TaskLog, error) {
	if wfId == "" {
		return nil, fmt.Errorf("please provide the wfId")
	}
	if wfRunId == "" {
		return nil, fmt.Errorf("please provide the wfRunId")
	}
	if taskId == "" {
		return nil, fmt.Errorf("please provide the taskId")
	}

	taskInfo, err := dao.TaskGet(taskId)
	if err != nil {
		return nil, fmt.Errorf("invalid get task from taskId")
	}

	client, err := airflow.GetClient()
	if err != nil {
		return nil, err
	}

	logs, err := client.GetTaskLogs(wfId, common.UrlDecode(wfRunId), taskInfo.Name, taskTryNum)
	if err != nil {
		return nil, fmt.Errorf("failed to get the workflow logs: %w", err)
	}

	taskLog := &model.TaskLog{
		Content: *logs.Content,
	}

	return taskLog, nil
}

// GetTaskLogDownload retrieves logs for download
func (s *taskService) GetTaskLogDownload(wfId, wfRunId, taskId string, taskTryNum int) ([]byte, string, error) {
	if wfId == "" {
		return nil, "", fmt.Errorf("please provide the wfId")
	}
	if wfRunId == "" {
		return nil, "", fmt.Errorf("please provide the wfRunId")
	}
	if taskId == "" {
		return nil, "", fmt.Errorf("please provide the taskId")
	}

	taskInfo, err := dao.TaskGet(taskId)
	if err != nil {
		return nil, "", fmt.Errorf("invalid get task from taskId")
	}

	client, err := airflow.GetClient()
	if err != nil {
		return nil, "", err
	}

	logs, err := client.GetTaskLogs(wfId, common.UrlDecode(wfRunId), taskInfo.Name, taskTryNum)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get the workflow logs: %w", err)
	}

	filename := fmt.Sprintf("%s_%s_%s.log", wfId, wfRunId, taskInfo.Name)
	return []byte(*logs.Content), filename, nil
}

// GetTaskInstances retrieves task instances for a workflow run
func (s *taskService) GetTaskInstances(wfId, wfRunId string) ([]model.TaskInstance, error) {
	if wfId == "" {
		return nil, fmt.Errorf("please provide the wfId")
	}
	if wfRunId == "" {
		return nil, fmt.Errorf("please provide the wfRunId")
	}

	client, err := airflow.GetClient()
	if err != nil {
		return nil, err
	}

	runList, err := client.GetTaskInstances(common.UrlDecode(wfId), common.UrlDecode(wfRunId))
	if err != nil {
		return nil, fmt.Errorf("failed to get the taskInstances: %w", err)
	}

	var taskInstances []model.TaskInstance
	layout := time.RFC3339Nano

	for _, taskInstance := range *runList.TaskInstances {
		taskDBInfo, err := dao.TaskGetByWorkflowIDAndName(taskInstance.GetDagId(), taskInstance.GetTaskId())
		if err != nil {
			return nil, fmt.Errorf("failed to get the taskInstances: %w", err)
		}
		taskId := &taskDBInfo.ID

		executionDate, err := time.Parse(layout, taskInstance.GetExecutionDate())
		if err != nil {
			fmt.Println("Error parsing execution date:", err)
			continue
		}
		startDate, err := time.Parse(layout, taskInstance.GetExecutionDate())
		if err != nil {
			fmt.Println("Error parsing start date:", err)
			continue
		}
		endDate, err := time.Parse(layout, taskInstance.GetExecutionDate())
		if err != nil {
			fmt.Println("Error parsing end date:", err)
			continue
		}

		taskInfo := model.TaskInstance{
			WorkflowID:    taskInstance.DagId,
			WorkflowRunID: taskInstance.GetDagRunId(),
			TaskID:        *taskId,
			TaskName:      taskInstance.GetTaskId(),
			State:         string(taskInstance.GetState()),
			ExecutionDate: executionDate,
			StartDate:     startDate,
			EndDate:       endDate,
			DurationDate:  float64(taskInstance.GetDuration()),
			TryNumber:     int(taskInstance.GetTryNumber()),
		}
		taskInstances = append(taskInstances, taskInfo)
	}

	return taskInstances, nil
}

// ClearTaskInstances clears task instances based on the provided options
func (s *taskService) ClearTaskInstances(wfId, wfRunId string, option model.TaskClearOption) ([]model.TaskInstanceReference, error) {
	if wfId == "" {
		return nil, fmt.Errorf("please provide the wfId")
	}
	if wfRunId == "" {
		return nil, fmt.Errorf("please provide the wfRunId")
	}

	// Convert task IDs to task names
	var taskNameList []string
	for _, taskId := range option.TaskIds {
		taskInfo, err := dao.TaskGet(taskId)
		if err != nil {
			return nil, fmt.Errorf("failed to get task info for ID %s: %w", taskId, err)
		}
		taskNameList = append(taskNameList, taskInfo.Name)
	}
	option.TaskIds = taskNameList

	if err := common.ValidateTaskClearOptions(option); err != nil {
		return nil, err
	}

	client, err := airflow.GetClient()
	if err != nil {
		return nil, err
	}

	clearList, err := client.ClearTaskInstance(wfId, common.UrlDecode(wfRunId), option)
	if err != nil {
		return nil, fmt.Errorf("failed to get the taskInstances: %w", err)
	}

	logger.Println(logger.DEBUG, false, "clearList 요청 내용 : {} ", &clearList)

	var taskInstanceReferences []model.TaskInstanceReference

	if clearList.TaskInstances == nil || len(*clearList.TaskInstances) == 0 {
		logger.Println(logger.DEBUG, false, "TaskInstances is nil or empty")
		return taskInstanceReferences, nil
	}

	for _, taskInstance := range *clearList.TaskInstances {
		taskDBInfo, err := dao.TaskGetByWorkflowIDAndName(taskInstance.GetDagId(), taskInstance.GetTaskId())
		if err != nil {
			return nil, fmt.Errorf("failed to get the taskInstances: %w", err)
		}
		taskId := &taskDBInfo.ID

		taskInfo := model.TaskInstanceReference{
			WorkflowID:    taskInstance.DagId,
			WorkflowRunID: taskInstance.DagRunId,
			TaskId:        taskId,
			TaskName:      taskInstance.GetTaskId(),
			ExecutionDate: taskInstance.ExecutionDate,
		}
		logger.Println(logger.DEBUG, false, "TaskInstanceReferences  ", taskInstanceReferences)
		taskInstanceReferences = append(taskInstanceReferences, taskInfo)
	}

	logger.Println(logger.DEBUG, false, "TaskInstanceReferences ", taskInstanceReferences)
	return taskInstanceReferences, nil
}
