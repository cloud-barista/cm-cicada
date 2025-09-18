package workflow

import (
	"net/http"

	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	_ "github.com/cloud-barista/cm-cicada/pkg/api/rest/model" // for swagger
	"github.com/cloud-barista/cm-cicada/pkg/service"
	"github.com/labstack/echo/v4"
)

var taskGroupService = service.NewTaskGroupService()

// ListTaskGroup godoc
//
//	@ID		list-task-group
//	@Summary	List TaskGroup
//	@Description	Get a task group list of the workflow.
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce	json
//	@Param		wfId path string true "ID of the workflow."
//	@Success	200	{object}	[]model.TaskGroup	"Successfully get a task group list."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get a task group list."
//	@Router		/workflow/{wfId}/task_group [get]
func ListTaskGroup(c echo.Context) error {
	wfId := c.Param("wfId")

	taskGroups, err := taskGroupService.ListTaskGroup(wfId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, taskGroups, " ")
}

// GetTaskGroup godoc
//
//	@ID		get-task-group
//	@Summary	Get TaskGroup
//	@Description	Get the task group.
//	@Tags	[Workflow]
//	@Accept	json
//	@Produce	json
//	@Param	wfId path string true "ID of the workflow."
//	@Param	tgId path string true "ID of the task group."
//	@Success	200	{object}	model.Task		"Successfully get the task group."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get the task group."
//	@Router	/workflow/{wfId}/task_group/{tgId} [get]
func GetTaskGroup(c echo.Context) error {
	wfId := c.Param("wfId")
	tgId := c.Param("tgId")

	taskGroup, err := taskGroupService.GetTaskGroup(wfId, tgId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, taskGroup, " ")
}

// GetTaskGroupDirectly godoc
//
//	@ID		get-task-group-directly
//	@Summary	Get TaskGroup Directly
//	@Description	Get the task group directly.
//	@Tags	[Workflow]
//	@Accept	json
//	@Produce	json
//	@Param	tgId path string true "ID of the task group."
//	@Success	200	{object}	model.Task		"Successfully get the task group."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get the task group."
//	@Router	/task_group/{tgId} [get]
func GetTaskGroupDirectly(c echo.Context) error {
	tgId := c.Param("tgId")

	taskGroup, err := taskGroupService.GetTaskGroupDirectly(tgId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, taskGroup, " ")
}

// ListTaskFromTaskGroup godoc
//
//	@ID		list-task-from-task-group
//	@Summary	List Task from Task Group
//	@Description	Get a task list from the task group.
//	@Tags	[Workflow]
//	@Accept	json
//	@Produce	json
//	@Param	wfId path string true "ID of the workflow."
//	@Param	tgId path string true "ID of the task group."
//	@Success	200	{object}	[]model.Task		"Successfully get a task list from the task group."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get a task list from the task group."
//	@Router	/workflow/{wfId}/task_group/{tgId}/task [get]
func ListTaskFromTaskGroup(c echo.Context) error {
	wfId := c.Param("wfId")
	tgId := c.Param("tgId")

	tasks, err := taskGroupService.ListTaskFromTaskGroup(wfId, tgId)
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
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce	json
//	@Param		wfId path string true "ID of the workflow."
//	@Param		tgId path string true "ID of the task group."
//	@Param		taskId path string true "ID of the task."
//	@Success	200	{object}	model.Task		"Successfully get the task from the task group."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get the task from the task group."
//	@Router		/workflow/{wfId}/task_group/{tgId}/task/{taskId} [get]
func GetTaskFromTaskGroup(c echo.Context) error {
	wfId := c.Param("wfId")
	tgId := c.Param("tgId")
	taskId := c.Param("taskId")

	task, err := taskGroupService.GetTaskFromTaskGroup(wfId, tgId, taskId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, task, " ")
}
