package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cloud-barista/cm-cicada/dao"
	"github.com/cloud-barista/cm-cicada/lib/airflow"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/mapper"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/jollaman999/utils/logger"
)

// timeOrZero unwraps a nullable Airflow timestamp. Airflow 3 leaves several of
// these null (logical_date on a manual run, end_date while still running).
func timeOrZero(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

// effectiveLogicalDate returns the run's logical date, falling back to run_after.
//
// Airflow 2 always stamped execution_date, including on manual runs. Airflow 3
// leaves logical_date null for them and records the trigger time in run_after,
// which carries the same meaning. Without the fallback every manually triggered
// run reports execution_date as the zero time.
func effectiveLogicalDate(logicalDate, runAfter *time.Time) time.Time {
	if logicalDate != nil && !logicalDate.IsZero() {
		return *logicalDate
	}
	return timeOrZero(runAfter)
}

func floatOrZero(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}

func stringOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

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
	for _, dagRun := range runList.DagRuns {
		dagID := dagRun.DagID
		logicalDate := effectiveLogicalDate(dagRun.LogicalDate, dagRun.RunAfter)

		runs = append(runs, model.WorkflowRun{
			WorkflowID:        &dbWorkflowID,
			DagID:             &dagID,
			WorkflowRunID:     dagRun.DagRunID,
			DataIntervalStart: timeOrZero(dagRun.DataIntervalStart),
			DataIntervalEnd:   timeOrZero(dagRun.DataIntervalEnd),
			State:             dagRun.State,
			// Airflow 3 renamed execution_date to logical_date. ExecutionDate is
			// kept populated so existing API consumers keep working.
			LogicalDate:            logicalDate.Format(time.RFC3339Nano),
			ExecutionDate:          logicalDate,
			StartDate:              timeOrZero(dagRun.StartDate),
			EndDate:                timeOrZero(dagRun.EndDate),
			RunType:                dagRun.RunType,
			LastSchedulingDecision: timeOrZero(dagRun.LastSchedulingDecision),
			// Airflow 3 reports duration directly instead of making us subtract.
			DurationDate: floatOrZero(dagRun.Duration),
			Conf:         dagRun.Conf,
			Note:         stringOrEmpty(dagRun.Note),
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
	for _, state := range enumStatus {
		resp, err := client.GetDagStatus(dagID, state)
		if err != nil {
			logger.Println(logger.ERROR, false,
				"AIRFLOW: Error occurred while getting DAGRuns. (Error: "+err.Error()+").")
			continue
		}
		statusList = append(statusList, model.WorkflowStatus{
			State: state,
			// total_entries is the count across all pages; counting the returned
			// slice would stop at one page.
			Count: resp.TotalEntries,
		})
	}

	return statusList, nil
}

func (s *WorkflowRuntimeService) GetImportErrors() (model.ImportErrors, error) {
	client, err := airflow.GetClient()
	if err != nil {
		return model.ImportErrors{}, err
	}

	result, err := client.GetImportErrors()
	if err != nil {
		return model.ImportErrors{}, errors.New("failed to get import errors: " + err.Error())
	}

	importErrors := make([]model.ImportError, 0, len(result.ImportErrors))
	for _, e := range result.ImportErrors {
		importErrors = append(importErrors, model.ImportError{
			ImportErrorID: e.ImportErrorID,
			Timestamp:     timeOrZero(e.Timestamp),
			Filename:      e.Filename,
			BundleName:    stringOrEmpty(e.BundleName),
			StackTrace:    e.StackTrace,
		})
	}

	return model.ImportErrors{
		ImportErrors: importErrors,
		TotalEntries: result.TotalEntries,
	}, nil
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

	return &model.TaskLog{Content: logs}, nil
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
	eventLogs, err := client.GetEventLogs(common.WorkflowDagID(workflow), wfRunID, airflowTaskID)
	if err != nil {
		return nil, errors.New("failed to get the taskInstances: " + err.Error())
	}

	logList := make([]model.EventLog, 0, len(eventLogs.EventLogs))
	for _, eventlog := range eventLogs.EventLogs {
		mappedTaskID := ""
		taskName := ""
		isDeletedTask := false
		if taskID := stringOrEmpty(eventlog.TaskID); taskID != "" {
			taskDBInfo, mappedDeleted, err := s.findTaskByAirflowTaskID(workflow, taskID)
			if err != nil {
				return nil, errors.New("failed to get the taskInstances: " + err.Error())
			}
			mappedTaskID = taskDBInfo.ID
			taskName = taskDBInfo.Name
			isDeletedTask = mappedDeleted
		}

		logList = append(logList, model.EventLog{
			WorkflowID:    wfID,
			WorkflowRunID: stringOrEmpty(eventlog.RunID),
			TaskID:        mappedTaskID,
			TaskName:      taskName,
			IsDeletedTask: isDeletedTask,
			Extra:         stringOrEmpty(eventlog.Extra),
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

	taskInstances := make([]model.TaskInstance, 0)
	dbWorkflowID := workflow.ID

	for _, taskInstance := range runList.TaskInstances {
		taskDBInfo, isDeletedTask, err := s.findTaskByAirflowTaskID(workflow, taskInstance.TaskID)
		if err != nil {
			return nil, errors.New("failed to get the taskInstances: " + err.Error())
		}
		taskID := taskDBInfo.ID

		// Airflow 3 hands these back as typed timestamps, so there is no string
		// parsing to fail. logical_date is null on manually triggered runs.
		executionDate := effectiveLogicalDate(taskInstance.LogicalDate, taskInstance.RunAfter)
		startDate := executionDate
		if taskInstance.StartDate != nil {
			startDate = *taskInstance.StartDate
		}
		endDate := executionDate
		if taskInstance.EndDate != nil {
			endDate = *taskInstance.EndDate
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
						taskInstance.DagID,
						taskInstance.DagRunID,
						taskInstance.TaskID,
						"return_value",
					)
					if err != nil {
						logger.Println(logger.WARN, false,
							"Failed to get xcom data for task: "+taskInstance.TaskID+" (Error: "+err.Error()+")")
					} else if xcomData != nil {
						if execID, ok := xcomData["execution_id"].(string); ok {
							executionID = execID
						}
					}
					break
				}
			}
		}

		dagID := taskInstance.DagID
		taskInstances = append(taskInstances, model.TaskInstance{
			WorkflowID:                   &dbWorkflowID,
			DagID:                        &dagID,
			IsDeletedTask:                isDeletedTask,
			WorkflowRunID:                taskInstance.DagRunID,
			TaskID:                       taskID,
			TaskName:                     taskDBInfo.Name,
			State:                        stringOrEmpty(taskInstance.State),
			ExecutionDate:                executionDate,
			StartDate:                    startDate,
			EndDate:                      endDate,
			DurationDate:                 floatOrZero(taskInstance.Duration),
			TryNumber:                    taskInstance.TryNumber,
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

	for _, taskInstance := range clearList.TaskInstances {
		taskDBInfo, _, err := s.findTaskByAirflowTaskID(workflow, taskInstance.TaskID)
		if err != nil {
			return nil, errors.New("failed to get the taskInstances: " + err.Error())
		}
		taskID := taskDBInfo.ID
		dagID := taskInstance.DagID
		dagRunID := taskInstance.DagRunID

		var executionDate *string
		if date := effectiveLogicalDate(taskInstance.LogicalDate, taskInstance.RunAfter); !date.IsZero() {
			formatted := date.Format(time.RFC3339Nano)
			executionDate = &formatted
		}

		refs = append(refs, model.TaskInstanceReference{
			WorkflowID:    &dbWorkflowID,
			DagID:         &dagID,
			WorkflowRunID: &dagRunID,
			TaskId:        &taskID,
			TaskName:      taskDBInfo.Name,
			ExecutionDate: executionDate,
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
