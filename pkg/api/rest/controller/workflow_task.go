package controller

import (
	"net/http"

	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/service"
	"github.com/labstack/echo/v4"
)

var _ model.Task

// ListTaskFromTaskGroup godoc
//
//	@ID		list-task-from-task-group
//	@Summary	List Task from Task Group
//	@Description	Get a task list from the task group.
//	@Tags		[Task]
//	@Accept		json
//	@Produce	json
//	@Param		wfId path string true "DB workflow ID."
//	@Param		tgId path string true "ID of the task group."
//	@Success	200	{array}		model.Task		"Successfully get a task list from the task group."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get a task list from the task group."
//	@Router		/workflow/{wfId}/task_group/{tgId}/task [get]
func ListTaskFromTaskGroup(c echo.Context) error {
	wfId, err := requireParam(c, "wfId", "wfId")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	tgId, err := requireParam(c, "tgId", "tgId")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	includeDeleted, err := queryBool(c, "include_deleted")
	if err != nil {
		return common.ReturnErrorMsg(c, "Invalid include_deleted value.")
	}

	svc := service.NewWorkflowTaskService()
	tasks, err := svc.ListTaskFromTaskGroup(wfId, tgId, includeDeleted)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, tasks, " ")
}

// GetTaskFromTaskGroup godoc
//
//	@ID		get-task-from-task-group
//	@Summary	Get Task from Task Group
//	@Description	Get the task from the task group.
//	@Tags		[Task]
//	@Accept		json
//	@Produce	json
//	@Param		wfId path string true "DB workflow ID."
//	@Param		tgId path string true "ID of the task group."
//	@Param		taskId path string true "ID of the task."
//	@Success	200	{object}	model.Task		"Successfully get the task from the task group."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get the task from the task group."
//	@Router		/workflow/{wfId}/task_group/{tgId}/task/{taskId} [get]
func GetTaskFromTaskGroup(c echo.Context) error {
	wfId, err := requireParam(c, "wfId", "wfId")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	tgId, err := requireParam(c, "tgId", "tgId")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	taskId, err := requireParam(c, "taskId", "taskId")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	includeDeleted, err := queryBool(c, "include_deleted")
	if err != nil {
		return common.ReturnErrorMsg(c, "Invalid include_deleted value.")
	}

	svc := service.NewWorkflowTaskService()
	task, err := svc.GetTaskFromTaskGroup(wfId, tgId, taskId, includeDeleted)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, task, " ")
}

// ListTask godoc
//
//	@ID		list-task
//	@Summary	List Task
//	@Description	Get a task list of the workflow.
//	@Tags		[Task]
//	@Accept		json
//	@Produce	json
//	@Param		wfId path string true "DB workflow ID."
//	@Success	200	{array}		model.Task		"Successfully get a task list."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get a task list."
//	@Router		/workflow/{wfId}/task [get]
func ListTask(c echo.Context) error {
	wfId, err := requireParam(c, "wfId", "wfId")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	includeDeleted, err := queryBool(c, "include_deleted")
	if err != nil {
		return common.ReturnErrorMsg(c, "Invalid include_deleted value.")
	}

	svc := service.NewWorkflowTaskService()
	tasks, err := svc.ListTask(wfId, includeDeleted)
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
//	@Tags		[Task]
//	@Accept		json
//	@Produce	json
//	@Param		wfId path string true "DB workflow ID."
//	@Param		taskId path string true "ID of the task."
//	@Success	200	{object}	model.Task		"Successfully get the task."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get the task."
//	@Router		/workflow/{wfId}/task/{taskId} [get]
func GetTask(c echo.Context) error {
	wfId, err := requireParam(c, "wfId", "wfId")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	taskId, err := requireParam(c, "taskId", "taskId")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	includeDeleted, err := queryBool(c, "include_deleted")
	if err != nil {
		return common.ReturnErrorMsg(c, "Invalid include_deleted value.")
	}

	svc := service.NewWorkflowTaskService()
	task, err := svc.GetTask(wfId, taskId, includeDeleted)
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
//	@Tags		[Task]
//	@Accept		json
//	@Produce	json
//	@Param		taskId path string true "ID of the task."
//	@Success	200	{object}	model.TaskDirectly	"Successfully get the task."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get the task."
//	@Router		/task/{taskId} [get]
func GetTaskDirectly(c echo.Context) error {
	taskId, err := requireParam(c, "taskId", "taskId")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	includeDeleted, err := queryBool(c, "include_deleted")
	if err != nil {
		return common.ReturnErrorMsg(c, "Invalid include_deleted value.")
	}

	svc := service.NewWorkflowTaskService()
	task, err := svc.GetTaskDirectly(taskId, includeDeleted)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, task, " ")
}
