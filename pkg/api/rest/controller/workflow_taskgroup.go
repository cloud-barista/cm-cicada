package controller

import (
	"net/http"

	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/service"
	"github.com/labstack/echo/v4"
)

var _ model.TaskGroup

// ListTaskGroup godoc
//
//	@ID		list-task-group
//	@Summary	List TaskGroup
//	@Description	Get a task group list of the workflow.
//	@Tags		[Task Group]
//	@Accept		json
//	@Produce	json
//	@Param		wfId path string true "DB workflow ID."
//	@Success	200	{array}		model.TaskGroup	"Successfully get a task group list."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get a task group list."
//	@Router		/workflow/{wfId}/task_group [get]
func ListTaskGroup(c echo.Context) error {
	wfId, err := requireParam(c, "wfId", "wfId")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	includeDeleted, err := queryBool(c, "include_deleted")
	if err != nil {
		return common.ReturnErrorMsg(c, "Invalid include_deleted value.")
	}

	svc := service.NewWorkflowTaskService()
	taskGroups, err := svc.ListTaskGroup(wfId, includeDeleted)
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
//	@Tags	[Task Group]
//	@Accept	json
//	@Produce	json
//	@Param	wfId path string true "DB workflow ID."
//	@Param	tgId path string true "ID of the task group."
//	@Success	200	{object}	model.Task		"Successfully get the task group."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get the task group."
//	@Router	/workflow/{wfId}/task_group/{tgId} [get]
func GetTaskGroup(c echo.Context) error {
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
	tg, err := svc.GetTaskGroup(wfId, tgId, includeDeleted)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, tg, " ")
}

// GetTaskGroupDirectly godoc
//
//	@ID		get-task-group-directly
//	@Summary	Get TaskGroup Directly
//	@Description	Get the task group directly.
//	@Tags	[Task Group]
//	@Accept	json
//	@Produce	json
//	@Param	tgId path string true "ID of the task group."
//	@Success	200	{object}	model.Task		"Successfully get the task group."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get the task group."
//	@Router	/task_group/{tgId} [get]
func GetTaskGroupDirectly(c echo.Context) error {
	tgId, err := requireParam(c, "tgId", "tgId")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	includeDeleted, err := queryBool(c, "include_deleted")
	if err != nil {
		return common.ReturnErrorMsg(c, "Invalid include_deleted value.")
	}

	svc := service.NewWorkflowTaskService()
	tg, err := svc.GetTaskGroupDirectly(tgId, includeDeleted)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, tg, " ")
}
