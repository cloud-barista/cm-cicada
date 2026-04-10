package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/cloud-barista/cm-cicada/dao"
	"github.com/cloud-barista/cm-cicada/lib/airflow"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/labstack/echo/v4"
)

// GetTaskLogs godoc
//
//	@ID			get-task-logs
//	@Summary	Get Task Logs
//	@Description	Get the task Logs.
//	@Tags	[Workflow]
//	@Accept	json
//	@Produce	json
//	@Param	wfId path string true "DB workflow ID."
//	@Param	wfRunId path string true "ID of the workflowRunId."
//	@Param	taskId path string true "ID of the task."
//	@Param	taskTryNum path string true "ID of the taskTryNum."
//	@Success	200	{object}	airflow.InlineResponse200		"Successfully get the task Logs."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get the task Logs."
//	@Router	 /workflow/{wfId}/workflowRun/{wfRunId}/task/{taskId}/taskTryNum/{taskTryNum}/logs [get]
func GetTaskLogs(c echo.Context) error {
	wfId, err := requireParam(c, "wfId", "wfId")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	wfRunId, err := requireParam(c, "wfRunId", "wfRunId")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	taskId, err := requireParam(c, "taskId", "taskId")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	taskInfo, err := dao.TaskGet(taskId)
	if err != nil {
		return common.ReturnErrorMsg(c, "Invalid get task from taskId.")
	}
	workflow, err := dao.WorkflowGet(wfId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	taskTryNum, err := requireParam(c, "taskTryNum", "taskTryNum")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	taskTryNumToInt, err := strconv.Atoi(taskTryNum)
	if err != nil {
		return common.ReturnErrorMsg(c, "Invalid taskTryNum format.")
	}
	client, err := airflow.GetClient()
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	logs, err := client.GetTaskLogs(
		workflowDagID(workflow),
		common.UrlDecode(wfRunId),
		taskAirflowID(taskInfo),
		taskTryNumToInt,
	)
	if err != nil {
		return common.ReturnErrorMsg(c, "Failed to get the workflow logs: "+err.Error())
	}

	taskLog := model.TaskLog{
		Content: *logs.Content,
	}

	return c.JSONPretty(http.StatusOK, taskLog, " ")
}

// GetTaskLogDownload godoc
//
//	@ID			get-task-logs-download
//	@Summary	Download Task Logs
//	@Description	Download the task logs as a file.
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce	text/plain
//	@Param		wfId path string true "DB workflow ID."
//	@Param		wfRunId path string true "ID of the workflowRunId."
//	@Param		taskId path string true "ID of the task."
//	@Param		taskTryNum path string true "ID of the taskTryNum."
//	@Success	200 {file} file "Log file downloaded successfully."
//	@Failure	400 {object} common.ErrorResponse "Sent bad request."
//	@Failure	500 {object} common.ErrorResponse "Failed to get the task Logs."
//	@Router		/workflow/{wfId}/workflowRun/{wfRunId}/task/{taskId}/taskTryNum/{taskTryNum}/logs/download [get]
func GetTaskLogDownload(c echo.Context) error {
	wfId, err := requireParam(c, "wfId", "wfId")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	wfRunId, err := requireParam(c, "wfRunId", "wfRunId")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	taskId, err := requireParam(c, "taskId", "taskId")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	taskInfo, err := dao.TaskGet(taskId)
	if err != nil {
		return common.ReturnErrorMsg(c, "Invalid get task from taskId.")
	}
	workflow, err := dao.WorkflowGet(wfId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	taskTryNum, err := requireParam(c, "taskTryNum", "taskTryNum")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	taskTryNumToInt, err := strconv.Atoi(taskTryNum)
	if err != nil {
		return common.ReturnErrorMsg(c, "Invalid taskTryNum format.")
	}
	client, err := airflow.GetClient()
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	logs, err := client.GetTaskLogs(
		workflowDagID(workflow),
		common.UrlDecode(wfRunId),
		taskAirflowID(taskInfo),
		taskTryNumToInt,
	)
	if err != nil {
		return common.ReturnErrorMsg(c, "Failed to get the workflow logs: "+err.Error())
	}
	filename := fmt.Sprintf("%s_%s_%s.log", wfId, wfRunId, taskInfo.Name)
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Response().Header().Set("Content-Type", "text/plain")
	return c.Blob(http.StatusOK, "text/plain", []byte(*logs.Content))
}

// GetEventLogs godoc
//
//	@ID			get-event-logs
//	@Summary		Get Eventlog
//	@Description	Get Eventlog.
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce	json
//	@Param		wfId path string true "DB workflow ID."
//	@Param		wfRunId query string false "ID of the workflow run."
//	@Param		taskId query string false "ID of the task."
//	@Success	200	{object}	[]model.EventLog			"Successfully get the workflow."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get the workflow."
//	@Router	/workflow/{wfId}/eventlogs [get]
func GetEventLogs(c echo.Context) error {
	wfId, err := requireParam(c, "wfId", "wfId")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	workflow, err := dao.WorkflowGet(wfId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	var wfRunId, taskId, airflowTaskID string

	if c.QueryParam("wfRunId") != "" {
		wfRunId = c.QueryParam("wfRunId")
	}
	if c.QueryParam("taskId") != "" {
		taskId = c.QueryParam("taskId")
		taskDBInfo, err := dao.TaskGet(taskId)
		if err != nil {
			return common.ReturnErrorMsg(c, "Failed to get the taskInstances: "+err.Error())
		}
		airflowTaskID = taskAirflowID(taskDBInfo)
	}
	var eventLogs model.EventLogs
	client, err := airflow.GetClient()
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	logs, err := client.GetEventLogs(workflowDagID(workflow), wfRunId, airflowTaskID)
	if err != nil {
		return common.ReturnErrorMsg(c, "Failed to get the taskInstances: "+err.Error())
	}
	err = json.Unmarshal(logs, &eventLogs)
	if err != nil {
		fmt.Println(err)
	}
	var logList []model.EventLog
	for _, eventlog := range eventLogs.EventLogs {
		var taskID, taskName, runID string
		if eventlog.TaskID != "" {
			taskDBInfo, err := dao.TaskGetByWorkflowIDAndTaskKey(workflow.ID, eventlog.TaskID)
			if err != nil {
				taskDBInfo, err = dao.TaskGetByWorkflowKeyAndTaskKey(workflowDagID(workflow), eventlog.TaskID)
			}
			if err != nil {
				taskDBInfo, err = dao.TaskGetByWorkflowIDAndName(wfId, eventlog.TaskID)
			}
			if err != nil {
				taskDBInfo, err = dao.TaskGetByWorkflowIDAndTaskKeyIncludeDeleted(workflow.ID, eventlog.TaskID)
			}
			if err != nil {
				taskDBInfo, err = dao.TaskGetByWorkflowKeyAndTaskKeyIncludeDeleted(workflowDagID(workflow), eventlog.TaskID)
			}
			if err != nil {
				taskDBInfo, err = dao.TaskGetByWorkflowIDAndNameIncludeDeleted(wfId, eventlog.TaskID)
			}
			if err != nil {
				return common.ReturnErrorMsg(c, "Failed to get the taskInstances: "+err.Error())
			}
			taskID = taskDBInfo.ID
			taskName = taskDBInfo.Name
		}
		eventlog.WorkflowID = wfId
		if eventlog.RunID != "" {
			runID = eventlog.RunID
		}

		log := model.EventLog{
			WorkflowID:    eventlog.WorkflowID,
			WorkflowRunID: runID,
			TaskID:        taskID,
			TaskName:      taskName,
			Extra:         eventlog.Extra,
			Event:         eventlog.Event,
			When:          eventlog.When,
		}
		logList = append(logList, log)
	}
	return c.JSONPretty(http.StatusOK, logList, " ")
}
