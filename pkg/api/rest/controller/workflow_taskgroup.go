package controller

import (
	"net/http"

	"github.com/cloud-barista/cm-cicada/dao"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/mapper"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/labstack/echo/v4"
)

// ListTaskGroup godoc
//
//	@ID		list-task-group
//	@Summary	List TaskGroup
//	@Description	Get a task group list of the workflow.
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce	json
//	@Param		wfId path string true "DB workflow ID."
//	@Success	200	{object}	[]model.TaskGroup	"Successfully get a task group list."
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

	var workflow *model.Workflow
	if includeDeleted {
		workflow, err = mapper.GetWorkflowFromDBIncludeDeleted(wfId)
	} else {
		workflow, err = mapper.GetWorkflowFromDB(wfId)
	}
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	taskGroups := make([]model.TaskGroup, 0, len(workflow.Data.TaskGroups))
	taskGroups = append(taskGroups, workflow.Data.TaskGroups...)

	if includeDeleted {
		taskGroupDBs, err := dao.TaskGroupGetListByWorkflowID(wfId, true)
		if err != nil {
			return common.ReturnErrorMsg(c, err.Error())
		}

		taskDBs, err := dao.TaskGetListByWorkflowID(wfId, true)
		if err != nil {
			return common.ReturnErrorMsg(c, err.Error())
		}

		taskByGroupID := make(map[string][]model.Task)
		for _, taskDB := range taskDBs {
			taskByGroupID[taskDB.TaskGroupID] = append(taskByGroupID[taskDB.TaskGroupID], model.Task{
				ID:   taskDB.ID,
				Name: taskDB.Name,
			})
		}

		groupByID := make(map[string]model.TaskGroup)
		for _, tg := range taskGroups {
			groupByID[tg.ID] = tg
		}

		for _, tgDB := range taskGroupDBs {
			if _, exists := groupByID[tgDB.ID]; exists {
				continue
			}
			taskGroups = append(taskGroups, model.TaskGroup{
				ID:          tgDB.ID,
				Name:        tgDB.Name,
				Description: "",
				Tasks:       taskByGroupID[tgDB.ID],
			})
		}
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

	var workflow *model.Workflow
	if includeDeleted {
		workflow, err = mapper.GetWorkflowFromDBIncludeDeleted(wfId)
	} else {
		workflow, err = mapper.GetWorkflowFromDB(wfId)
	}
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	for _, tg := range workflow.Data.TaskGroups {
		if tg.ID == tgId {
			return c.JSONPretty(http.StatusOK, tg, " ")
		}
	}

	if includeDeleted {
		tgDB, err := dao.TaskGroupGetIncludeDeleted(tgId)
		if err != nil {
			return common.ReturnErrorMsg(c, "Task group not found.")
		}
		return c.JSONPretty(http.StatusOK, model.TaskGroup{
			ID:          tgDB.ID,
			Name:        tgDB.Name,
			Description: "",
			Tasks:       []model.Task{},
		}, " ")
	}

	return common.ReturnErrorMsg(c, "Task group not found.")
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
	tgId, err := requireParam(c, "tgId", "tgId")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	includeDeleted, err := queryBool(c, "include_deleted")
	if err != nil {
		return common.ReturnErrorMsg(c, "Invalid include_deleted value.")
	}

	var tgDB *model.TaskGroupDBModel
	if includeDeleted {
		tgDB, err = dao.TaskGroupGetIncludeDeleted(tgId)
	} else {
		tgDB, err = dao.TaskGroupGet(tgId)
	}
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	var workflow *model.Workflow
	if includeDeleted {
		workflow, err = mapper.GetWorkflowFromDBIncludeDeleted(tgDB.WorkflowID)
	} else {
		workflow, err = mapper.GetWorkflowFromDB(tgDB.WorkflowID)
	}
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	for _, tg := range workflow.Data.TaskGroups {
		if tg.ID == tgId {
			return c.JSONPretty(http.StatusOK, model.TaskGroupDirectly{
				ID:          tg.ID,
				WorkflowID:  tgDB.WorkflowID,
				Name:        tg.Name,
				Description: tg.Description,
				Tasks:       tg.Tasks,
			}, " ")
		}
	}

	if includeDeleted {
		return c.JSONPretty(http.StatusOK, model.TaskGroupDirectly{
			ID:          tgDB.ID,
			WorkflowID:  tgDB.WorkflowID,
			Name:        tgDB.Name,
			Description: "",
			Tasks:       []model.Task{},
		}, " ")
	}

	return common.ReturnErrorMsg(c, "task group not found.")
}
