package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	af "github.com/apache/airflow-client-go/airflow"
	"github.com/cloud-barista/cm-cicada/dao"
	"github.com/cloud-barista/cm-cicada/lib/airflow"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/mapper"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/jollaman999/utils/logger"
)

type WorkflowRuntimeService struct{}

func NewWorkflowRuntimeService() *WorkflowRuntimeService {
	return &WorkflowRuntimeService{}
}

func (s *WorkflowRuntimeService) GetWorkflowRuns(wfID string) ([]model.WorkflowRun, error) {
	workflow, err := dao.WorkflowGet(wfID)
	if err != nil {
		return nil, err
	}

	client, err := airflow.GetClient()
	if err != nil {
		return nil, err
	}

	runList, err := client.GetDAGRuns(common.WorkflowDagID(workflow))
	if err != nil {
		return nil, errors.New("failed to get the workflow runs: " + err.Error())
	}

	dbWorkflowID := workflow.ID
	var runs []model.WorkflowRun
	for _, dagRun := range *runList.DagRuns {
		runs = append(runs, model.WorkflowRun{
			WorkflowID:             &dbWorkflowID,
			DagID:                  dagRun.DagId,
			WorkflowRunID:          dagRun.GetDagRunId(),
			DataIntervalStart:      dagRun.GetDataIntervalStart(),
			DataIntervalEnd:        dagRun.GetDataIntervalEnd(),
			State:                  string(dagRun.GetState()),
			ExecutionDate:          dagRun.GetExecutionDate(),
			StartDate:              dagRun.GetStartDate(),
			EndDate:                dagRun.GetEndDate(),
			RunType:                dagRun.GetRunType(),
			LastSchedulingDecision: dagRun.GetLastSchedulingDecision(),
			DurationDate:           dagRun.GetEndDate().Sub(dagRun.GetStartDate()).Seconds(),
		})
	}

	return runs, nil
}

func (s *WorkflowRuntimeService) GetWorkflowStatus(wfID string) ([]model.WorkflowStatus, error) {
	workflow, err := dao.WorkflowGet(wfID)
	if err != nil {
		return nil, err
	}

	client, err := airflow.GetClient()
	if err != nil {
		return nil, err
	}

	enumStatus := client.GetAllowedDagStateEnumValues()
	dagID := common.WorkflowDagID(workflow)
	var statusList []model.WorkflowStatus
	for _, v := range enumStatus {
		resp, err := client.GetDagStatus(dagID, string(*v.Ptr()))
		if err != nil {
			logger.Println(logger.ERROR, false,
				"AIRFLOW: Error occurred while getting DAGRuns. (Error: "+err.Error()+").")
		}
		statusList = append(statusList, model.WorkflowStatus{
			State: string(*v.Ptr()),
			Count: len(*resp.DagRuns),
		})
	}

	return statusList, nil
}

func (s *WorkflowRuntimeService) GetImportErrors() (af.ImportErrorCollection, error) {
	client, err := airflow.GetClient()
	if err != nil {
		return af.ImportErrorCollection{}, err
	}

	result, err := client.GetImportErrors()
	if err != nil {
		return af.ImportErrorCollection{}, errors.New("failed to get import errors: " + err.Error())
	}

	return result, nil
}

func (s *WorkflowRuntimeService) GetTaskLogs(wfID, wfRunID, taskID string, taskTryNum int) (*model.TaskLog, error) {
	taskInfo, err := dao.TaskGet(taskID)
	if err != nil {
		return nil, errors.New("invalid get task from taskId")
	}
	workflow, err := dao.WorkflowGet(wfID)
	if err != nil {
		return nil, err
	}

	client, err := airflow.GetClient()
	if err != nil {
		return nil, err
	}
	logs, err := client.GetTaskLogs(
		common.WorkflowDagID(workflow),
		common.UrlDecode(wfRunID),
		taskRuntimeAirflowID(taskInfo),
		taskTryNum,
	)
	if err != nil {
		return nil, errors.New("failed to get the workflow logs: " + err.Error())
	}

	return &model.TaskLog{Content: logs.GetContent()}, nil
}

func (s *WorkflowRuntimeService) GetTaskLogDownload(wfID, wfRunID, taskID string, taskTryNum int) (string, []byte, error) {
	taskInfo, err := dao.TaskGet(taskID)
	if err != nil {
		return "", nil, errors.New("invalid get task from taskId")
	}

	taskLog, err := s.GetTaskLogs(wfID, wfRunID, taskID, taskTryNum)
	if err != nil {
		return "", nil, err
	}

	filename := fmt.Sprintf("%s_%s_%s.log", wfID, wfRunID, taskInfo.Name)
	return filename, []byte(taskLog.Content), nil
}

func (s *WorkflowRuntimeService) GetEventLogs(wfID, wfRunID, taskID string) ([]model.EventLog, error) {
	workflow, err := dao.WorkflowGet(wfID)
	if err != nil {
		return nil, err
	}

	var airflowTaskID string
	if taskID != "" {
		taskDBInfo, err := dao.TaskGet(taskID)
		if err != nil {
			return nil, errors.New("failed to get the taskInstances: " + err.Error())
		}
		airflowTaskID = taskRuntimeAirflowID(taskDBInfo)
	}

	client, err := airflow.GetClient()
	if err != nil {
		return nil, err
	}
	logs, err := client.GetEventLogs(common.WorkflowDagID(workflow), wfRunID, airflowTaskID)
	if err != nil {
		return nil, errors.New("failed to get the taskInstances: " + err.Error())
	}

	var eventLogs model.EventLogs
	if err := json.Unmarshal(logs, &eventLogs); err != nil {
		return nil, err
	}

	logList := make([]model.EventLog, 0, len(eventLogs.EventLogs))
	for _, eventlog := range eventLogs.EventLogs {
		mappedTaskID := ""
		taskName := ""
		isDeletedTask := false
		if eventlog.TaskID != "" {
			taskDBInfo, mappedDeleted, err := s.findTaskByAirflowTaskID(workflow, eventlog.TaskID)
			if err != nil {
				return nil, errors.New("failed to get the taskInstances: " + err.Error())
			}
			mappedTaskID = taskDBInfo.ID
			taskName = taskDBInfo.Name
			isDeletedTask = mappedDeleted
		}

		workflowRunID := ""
		if eventlog.RunID != "" {
			workflowRunID = eventlog.RunID
		}

		logList = append(logList, model.EventLog{
			WorkflowID:    wfID,
			WorkflowRunID: workflowRunID,
			TaskID:        mappedTaskID,
			TaskName:      taskName,
			IsDeletedTask: isDeletedTask,
			Extra:         eventlog.Extra,
			Event:         eventlog.Event,
			When:          eventlog.When,
		})
	}

	return logList, nil
}

func (s *WorkflowRuntimeService) GetTaskInstances(wfID, wfRunID string) ([]model.TaskInstance, error) {
	workflow, err := mapper.GetWorkflowFromDB(wfID)
	if err != nil {
		return nil, err
	}

	client, err := airflow.GetClient()
	if err != nil {
		return nil, err
	}

	runList, err := client.GetTaskInstances(common.WorkflowDagID(workflow), common.UrlDecode(wfRunID))
	if err != nil {
		return nil, errors.New("failed to get the taskInstances: " + err.Error())
	}

	layout := time.RFC3339Nano
	taskInstances := make([]model.TaskInstance, 0)
	dbWorkflowID := workflow.ID

	for _, taskInstance := range *runList.TaskInstances {
		taskDBInfo, isDeletedTask, err := s.findTaskByAirflowTaskID(workflow, taskInstance.GetTaskId())
		if err != nil {
			return nil, errors.New("failed to get the taskInstances: " + err.Error())
		}
		taskID := taskDBInfo.ID

		executionDate, err := time.Parse(layout, taskInstance.GetExecutionDate())
		if err != nil {
			continue
		}
		startDate, err := time.Parse(layout, taskInstance.GetStartDate())
		if err != nil {
			startDate = executionDate
		}
		endDate, err := time.Parse(layout, taskInstance.GetEndDate())
		if err != nil {
			endDate = executionDate
		}

		isSoftwareMigrationTask := false
		executionID := ""
		for _, tg := range workflow.Data.TaskGroups {
			for _, task := range tg.Tasks {
				if strings.Contains(task.TaskComponent, "grasshopper") &&
					strings.Contains(task.TaskComponent, "software") &&
					strings.Contains(task.TaskComponent, "migration") &&
					task.ID == taskID {
					isSoftwareMigrationTask = true
					xcomData, err := client.GetXComValue(
						taskInstance.GetDagId(),
						taskInstance.GetDagRunId(),
						taskInstance.GetTaskId(),
						"return_value",
					)
					if err != nil {
						logger.Println(logger.WARN, false,
							"Failed to get xcom data for task: "+taskInstance.GetTaskId()+" (Error: "+err.Error()+")")
					} else if xcomData != nil {
						if execID, ok := xcomData["execution_id"].(string); ok {
							executionID = execID
						}
					}
					break
				}
			}
		}

		taskInstances = append(taskInstances, model.TaskInstance{
			WorkflowID:                   &dbWorkflowID,
			DagID:                        taskInstance.DagId,
			IsDeletedTask:                isDeletedTask,
			WorkflowRunID:                taskInstance.GetDagRunId(),
			TaskID:                       taskID,
			TaskName:                     taskDBInfo.Name,
			State:                        string(taskInstance.GetState()),
			ExecutionDate:                executionDate,
			StartDate:                    startDate,
			EndDate:                      endDate,
			DurationDate:                 float64(taskInstance.GetDuration()),
			TryNumber:                    int(taskInstance.GetTryNumber()),
			IsSoftwareMigrationTask:      isSoftwareMigrationTask,
			SoftwareMigrationExecutionID: executionID,
		})
	}

	return taskInstances, nil
}

func (s *WorkflowRuntimeService) ClearTaskInstances(wfID, wfRunID string, option model.TaskClearOption) ([]model.TaskInstanceReference, error) {
	workflow, err := dao.WorkflowGet(wfID)
	if err != nil {
		return nil, err
	}

	taskKeyList := make([]string, 0, len(option.TaskIds))
	for _, taskID := range option.TaskIds {
		taskInfo, err := dao.TaskGet(taskID)
		if err != nil {
			return nil, fmt.Errorf("failed to get task info for ID %s: %w", taskID, err)
		}
		taskKeyList = append(taskKeyList, taskRuntimeAirflowID(taskInfo))
	}
	option.TaskIds = taskKeyList

	client, err := airflow.GetClient()
	if err != nil {
		return nil, err
	}
	clearList, err := client.ClearTaskInstance(common.WorkflowDagID(workflow), common.UrlDecode(wfRunID), option)
	if err != nil {
		return nil, errors.New("failed to get the taskInstances: " + err.Error())
	}

	refs := make([]model.TaskInstanceReference, 0)
	dbWorkflowID := workflow.ID
	if clearList.TaskInstances == nil || len(*clearList.TaskInstances) == 0 {
		return refs, nil
	}

	for _, taskInstance := range *clearList.TaskInstances {
		taskDBInfo, _, err := s.findTaskByAirflowTaskID(workflow, taskInstance.GetTaskId())
		if err != nil {
			return nil, errors.New("failed to get the taskInstances: " + err.Error())
		}
		taskID := taskDBInfo.ID
		refs = append(refs, model.TaskInstanceReference{
			WorkflowID:    &dbWorkflowID,
			DagID:         taskInstance.DagId,
			WorkflowRunID: taskInstance.DagRunId,
			TaskId:        &taskID,
			TaskName:      taskDBInfo.Name,
			ExecutionDate: taskInstance.ExecutionDate,
		})
	}

	return refs, nil
}

func (s *WorkflowRuntimeService) findTaskByAirflowTaskID(workflow *model.Workflow, airflowTaskID string) (*model.TaskDBModel, bool, error) {
	taskDBInfo, err := dao.TaskGetByWorkflowIDAndTaskKey(workflow.ID, airflowTaskID)
	if err != nil {
		taskDBInfo, err = dao.TaskGetByWorkflowKeyAndTaskKey(common.WorkflowDagID(workflow), airflowTaskID)
	}
	if err != nil {
		taskDBInfo, err = dao.TaskGetByWorkflowIDAndName(workflow.ID, airflowTaskID)
	}
	if err == nil {
		return taskDBInfo, taskDBInfo.IsDeleted, nil
	}

	if err != nil {
		taskDBInfo, err = dao.TaskGetByWorkflowIDAndTaskKeyIncludeDeleted(workflow.ID, airflowTaskID)
	}
	if err != nil {
		taskDBInfo, err = dao.TaskGetByWorkflowKeyAndTaskKeyIncludeDeleted(common.WorkflowDagID(workflow), airflowTaskID)
	}
	if err != nil {
		taskDBInfo, err = dao.TaskGetByWorkflowIDAndNameIncludeDeleted(workflow.ID, airflowTaskID)
	}
	if err != nil {
		return nil, false, err
	}
	return taskDBInfo, taskDBInfo.IsDeleted, nil
}

func taskRuntimeAirflowID(task *model.TaskDBModel) string {
	if task.TaskKey != "" {
		return task.TaskKey
	}
	if task.ID != "" {
		return task.ID
	}
	return task.Name
}
