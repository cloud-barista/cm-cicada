package workflow

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/util"
	"github.com/cloud-barista/cm-cicada/pkg/service"
	"github.com/labstack/echo/v4"
	"github.com/mitchellh/mapstructure"
)

var taskService = service.NewTaskService()

// ListTask godoc
//
//	@ID		list-task
//	@Summary	List Task
//	@Description	Get a task list of the workflow.
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce	json
//	@Param		wfId path string true "ID of the workflow."
//	@Success	200	{object}	[]model.Task		"Successfully get a task list."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get a task list."
//	@Router		/workflow/{wfId}/task [get]
func ListTask(c echo.Context) error {
	wfId := c.Param("wfId")

	tasks, err := taskService.ListTask(wfId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, tasks, " ")
}

// GetTask godoc
//
//	@ID		get-task
//	@Summary	Get Task
//	@Description	Get the task.
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce	json
//	@Param		wfId path string true "ID of the workflow."
//	@Param		taskId path string true "ID of the task."
//	@Success	200	{object}	model.Task		"Successfully get the task."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get the task."
//	@Router		/workflow/{wfId}/task/{taskId} [get]
func GetTask(c echo.Context) error {
	wfId := c.Param("wfId")
	taskId := c.Param("taskId")

	task, err := taskService.GetTask(wfId, taskId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, task, " ")
}

// GetTaskDirectly godoc
//
//	@ID		get-task-directly
//	@Summary	Get Task Directly
//	@Description	Get the task directly.
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce	json
//	@Param		taskId path string true "ID of the task."
//	@Success	200	{object}	model.TaskDirectly	"Successfully get the task."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get the task."
//	@Router		/task/{taskId} [get]
func GetTaskDirectly(c echo.Context) error {
	taskId := c.Param("taskId")

	task, err := taskService.GetTaskDirectly(taskId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, task, " ")
}

// GetTaskLogs godoc
//
//	@ID			get-task-logs
//	@Summary	Get Task Logs
//	@Description	Get the task Logs.
//	@Tags	[Workflow]
//	@Accept	json
//	@Produce	json
//	@Param	wfId path string true "ID of the workflow."
//	@Param	wfRunId path string true "ID of the workflowRunId."
//	@Param	taskId path string true "ID of the task."
//	@Param	taskTyNum path string true "ID of the taskTryNum."
//	@Success	200	{object}	airflow.InlineResponse200		"Successfully get the task Logs."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get the task Logs."
//	@Router	 /workflow/{wfId}/workflowRun/{wfRunId}/task/{taskId}/taskTryNum/{taskTyNum}/logs [get]
func GetTaskLogs(c echo.Context) error {
	wfId := c.Param("wfId")
	wfRunId := c.Param("wfRunId")
	taskId := c.Param("taskId")
	taskTyNum := c.Param("taskTyNum")

	taskTyNumToInt, err := strconv.Atoi(taskTyNum)
	if err != nil {
		return common.ReturnErrorMsg(c, "Invalid taskTryNum format.")
	}

	taskLog, err := taskService.GetTaskLogs(wfId, wfRunId, taskId, taskTyNumToInt)
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
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce	text/plain
//	@Param		wfId path string true "ID of the workflow."
//	@Param		wfRunId path string true "ID of the workflowRunId."
//	@Param		taskId path string true "ID of the task."
//	@Param		taskTyNum path string true "ID of the taskTryNum."
//	@Success	200 {file} file "Log file downloaded successfully."
//	@Failure	400 {object} common.ErrorResponse "Sent bad request."
//	@Failure	500 {object} common.ErrorResponse "Failed to get the task Logs."
//	@Router		/workflow/{wfId}/workflowRun/{wfRunId}/task/{taskId}/taskTryNum/{taskTyNum}/logs/download [get]
func GetTaskLogDownload(c echo.Context) error {
	wfId := c.Param("wfId")
	wfRunId := c.Param("wfRunId")
	taskId := c.Param("taskId")
	taskTyNum := c.Param("taskTyNum")

	taskTyNumToInt, err := strconv.Atoi(taskTyNum)
	if err != nil {
		return common.ReturnErrorMsg(c, "Invalid taskTryNum format.")
	}

	logData, filename, err := taskService.GetTaskLogDownload(wfId, wfRunId, taskId, taskTyNumToInt)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Response().Header().Set("Content-Type", "text/plain")
	return c.Blob(http.StatusOK, "text/plain", logData)
}

// GetTaskInstances godoc
//
//	@ID			get-task-instances
//	@Summary	Get taskInstances
//	@Description	Get the task Logs.
//	@Tags	[Workflow]
//	@Accept	json
//	@Produce	json
//	@Param	wfId path string true "ID of the workflow."
//	@Param	wfRunId path string true "ID of the workflow."
//	@Success	200	{object}	model.TaskInstance		"Successfully get the taskInstances."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get the taskInstances."
//	@Router	 /workflow/{wfId}/workflowRun/{wfRunId}/taskInstances [get]
func GetTaskInstances(c echo.Context) error {
	wfId := c.Param("wfId")
	wfRunId := c.Param("wfRunId")

	taskInstances, err := taskService.GetTaskInstances(wfId, wfRunId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, taskInstances, " ")
}

// ClearTaskInstances godoc
//
//	@ID			clear-task-instances
//	@Summary	Clear taskInstances
//	@Description	Clear the task Instance.
//	@Tags	[Workflow]
//	@Accept	json
//	@Produce	json
//	@Param	wfId path string true "ID of the workflow."
//	@Param	wfRunId path string true "ID of the wfRunId."
//
// @Param		request body 	model.TaskClearOption true "Workflow content"
// @Success	200	{object}	model.TaskInstanceReference		"Successfully clear the taskInstances."
// @Failure	400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure	500	{object}	common.ErrorResponse	"Failed to clear the taskInstances."
// @Router	 /workflow/{wfId}/workflowRun/{wfRunId}/range [post]
func ClearTaskInstances(c echo.Context) error {
	var taskClearOption model.TaskClearOption

	data, err := common.GetJSONRawBody(c)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Metadata: nil,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			util.ToTimeHookFunc()),
		Result: &taskClearOption,
	})
	if err != nil {
		return err
	}

	err = decoder.Decode(data)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	wfId := c.Param("wfId")
	wfRunId := c.Param("wfRunId")

	taskInstanceReferences, err := taskService.ClearTaskInstances(wfId, wfRunId, taskClearOption)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, taskInstanceReferences, " ")
}
