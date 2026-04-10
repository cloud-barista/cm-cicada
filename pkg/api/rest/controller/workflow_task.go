package controller

import (
	"net/http"

	"github.com/cloud-barista/cm-cicada/dao"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/mapper"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/labstack/echo/v4"
)

// ListTaskFromTaskGroup godoc
//
//	@ID		list-task-from-task-group
//	@Summary	List Task from Task Group
//	@Description	Get a task list from the task group.
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce	json
//	@Param		wfId path string true "DB workflow ID."
//	@Param		tgId path string true "ID of the task group."
//	@Success	200	{object}	[]model.Task		"Successfully get a task list from the task group."
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

	var workflow *model.Workflow
	if includeDeleted {
		workflow, err = mapper.GetWorkflowFromDBIncludeDeleted(wfId)
	} else {
		workflow, err = mapper.GetWorkflowFromDB(wfId)
	}
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	var tasks []model.Task
	for _, tg := range workflow.Data.TaskGroups {
		if tg.ID == tgId {
			tasks = append(tasks, tg.Tasks...)
			break
		}
	}

	if includeDeleted {
		taskDBs, err := dao.TaskGetListByWorkflowID(wfId, true)
		if err != nil {
			return common.ReturnErrorMsg(c, err.Error())
		}
		existing := make(map[string]bool)
		for _, task := range tasks {
			existing[task.ID] = true
		}
		for _, taskDB := range taskDBs {
			if taskDB.TaskGroupID != tgId {
				continue
			}
			if existing[taskDB.ID] {
				continue
			}
			tasks = append(tasks, model.Task{
				ID:   taskDB.ID,
				Name: taskDB.Name,
			})
		}
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
			for _, task := range tg.Tasks {
				if task.ID == taskId {
					return c.JSONPretty(http.StatusOK, task, " ")
				}
			}

			break
		}
	}

	if includeDeleted {
		taskDB, err := dao.TaskGetIncludeDeleted(taskId)
		if err != nil {
			return common.ReturnErrorMsg(c, "Task not found.")
		}
		return c.JSONPretty(http.StatusOK, model.Task{
			ID:   taskDB.ID,
			Name: taskDB.Name,
		}, " ")
	}

	return common.ReturnErrorMsg(c, "Task not found.")
}

// ListTask godoc
//
//	@ID		list-task
//	@Summary	List Task
//	@Description	Get a task list of the workflow.
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce	json
//	@Param		wfId path string true "DB workflow ID."
//	@Success	200	{object}	[]model.Task		"Successfully get a task list."
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

	var workflow *model.Workflow
	if includeDeleted {
		workflow, err = mapper.GetWorkflowFromDBIncludeDeleted(wfId)
	} else {
		workflow, err = mapper.GetWorkflowFromDB(wfId)
	}
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	var tasks []model.Task
	for _, tg := range workflow.Data.TaskGroups {
		tasks = append(tasks, tg.Tasks...)
	}

	if includeDeleted {
		taskDBs, err := dao.TaskGetListByWorkflowID(wfId, true)
		if err != nil {
			return common.ReturnErrorMsg(c, err.Error())
		}
		existing := make(map[string]bool)
		for _, task := range tasks {
			existing[task.ID] = true
		}
		for _, taskDB := range taskDBs {
			if existing[taskDB.ID] {
				continue
			}
			tasks = append(tasks, model.Task{
				ID:   taskDB.ID,
				Name: taskDB.Name,
			})
		}
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
		for _, task := range tg.Tasks {
			if task.ID == taskId {
				return c.JSONPretty(http.StatusOK, task, " ")
			}
		}
	}

	if includeDeleted {
		taskDB, err := dao.TaskGetIncludeDeleted(taskId)
		if err != nil {
			return common.ReturnErrorMsg(c, "Task not found.")
		}
		return c.JSONPretty(http.StatusOK, model.Task{
			ID:   taskDB.ID,
			Name: taskDB.Name,
		}, " ")
	}

	return common.ReturnErrorMsg(c, "Task not found.")
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
	taskId, err := requireParam(c, "taskId", "taskId")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	includeDeleted, err := queryBool(c, "include_deleted")
	if err != nil {
		return common.ReturnErrorMsg(c, "Invalid include_deleted value.")
	}

	var tDB *model.TaskDBModel
	if includeDeleted {
		tDB, err = dao.TaskGetIncludeDeleted(taskId)
	} else {
		tDB, err = dao.TaskGet(taskId)
	}
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	var tgDB *model.TaskGroupDBModel
	if includeDeleted {
		tgDB, err = dao.TaskGroupGetIncludeDeleted(tDB.TaskGroupID)
	} else {
		tgDB, err = dao.TaskGroupGet(tDB.TaskGroupID)
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
		if tg.ID == tgDB.ID {
			for _, task := range tg.Tasks {
				if task.ID == taskId {
					return c.JSONPretty(http.StatusOK, model.TaskDirectly{
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
					}, " ")
				}
			}
		}
	}

	if includeDeleted {
		return c.JSONPretty(http.StatusOK, model.TaskDirectly{
			ID:            tDB.ID,
			WorkflowID:    tDB.WorkflowID,
			TaskGroupID:   tDB.TaskGroupID,
			Name:          tDB.Name,
			TaskComponent: "",
			RequestBody:   "",
			PathParams:    nil,
			QueryParams:   nil,
			Extra:         nil,
			Dependencies:  nil,
		}, " ")
	}

	return common.ReturnErrorMsg(c, "task not found.")
}
