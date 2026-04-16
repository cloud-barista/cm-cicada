package controller

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/service"
	"github.com/labstack/echo/v4"
)

var _ model.EventLog

// GetTaskLogs godoc
//
//	@ID			get-task-logs
//	@Summary	Get Task Logs
//	@Description	Get the task Logs.
//	@Tags	[Workflow Execution]
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
	taskTryNum, err := requireParam(c, "taskTryNum", "taskTryNum")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	taskTryNumToInt, err := strconv.Atoi(taskTryNum)
	if err != nil {
		return common.ReturnErrorMsg(c, "Invalid taskTryNum format.")
	}

	svc := service.NewWorkflowRuntimeService()
	taskLog, err := svc.GetTaskLogs(wfId, wfRunId, taskId, taskTryNumToInt)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, taskLog, " ")
}

// GetTaskLogDownload godoc
//
//	@ID			get-task-logs-download
//	@Summary	Download Task Logs
//	@Description	Download the task logs as a file.
//	@Tags		[Workflow Execution]
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
	taskTryNum, err := requireParam(c, "taskTryNum", "taskTryNum")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	taskTryNumToInt, err := strconv.Atoi(taskTryNum)
	if err != nil {
		return common.ReturnErrorMsg(c, "Invalid taskTryNum format.")
	}

	svc := service.NewWorkflowRuntimeService()
	filename, content, err := svc.GetTaskLogDownload(wfId, wfRunId, taskId, taskTryNumToInt)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Response().Header().Set("Content-Type", "text/plain")
	return c.Blob(http.StatusOK, "text/plain", content)
}

// GetEventLogs godoc
//
//	@ID			get-event-logs
//	@Summary		Get Eventlog
//	@Description	Get Eventlog.
//	@Tags		[Workflow Execution]
//	@Accept		json
//	@Produce	json
//	@Param		wfId path string true "DB workflow ID."
//	@Param		wfRunId query string false "ID of the workflow run."
//	@Param		taskId query string false "ID of the task."
//	@Success	200	{array}		model.EventLog			"Successfully get the workflow."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get the workflow."
//	@Router	/workflow/{wfId}/eventlogs [get]
func GetEventLogs(c echo.Context) error {
	wfId, err := requireParam(c, "wfId", "wfId")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	wfRunId := c.QueryParam("wfRunId")
	taskId := c.QueryParam("taskId")

	svc := service.NewWorkflowRuntimeService()
	logs, err := svc.GetEventLogs(wfId, wfRunId, taskId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, logs, " ")
}
