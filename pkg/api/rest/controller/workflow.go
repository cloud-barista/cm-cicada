package controller

import (
	"errors"
	"github.com/cloud-barista/cm-cicada/dao"
	"github.com/cloud-barista/cm-cicada/lib/airflow"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/google/uuid"
	"github.com/jollaman999/utils/logger"
	"github.com/labstack/echo/v4"
	"github.com/mitchellh/mapstructure"
	"net/http"
	"reflect"
	"time"
)

func toTimeHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if t != reflect.TypeOf(time.Time{}) {
			return data, nil
		}

		switch f.Kind() {
		case reflect.String:
			return time.Parse(time.RFC3339, data.(string))
		case reflect.Float64:
			return time.Unix(0, int64(data.(float64))*int64(time.Millisecond)), nil
		case reflect.Int64:
			return time.Unix(0, data.(int64)*int64(time.Millisecond)), nil
		default:
			return data, nil
		}
		// Convert it by parsing
	}
}

func createDataReqToData(createDataReq model.CreateDataReq) (model.Data, error) {
	var taskGroups []model.TaskGroup
	var allTasks []model.Task

	for _, tgReq := range createDataReq.TaskGroups {
		var tasks []model.Task
		for _, tReq := range tgReq.Tasks {
			tasks = append(tasks, model.Task{
				ID:            uuid.New().String(),
				Name:          tReq.Name,
				TaskComponent: tReq.TaskComponent,
				RequestBody:   tReq.RequestBody,
				PathParams:    tReq.PathParams,
				Dependencies:  tReq.Dependencies,
			})
		}

		allTasks = append(allTasks, tasks...)
		taskGroups = append(taskGroups, model.TaskGroup{
			ID:          uuid.New().String(),
			Name:        tgReq.Name,
			Description: tgReq.Description,
			Tasks:       tasks,
		})
	}

	for i, tgReq := range createDataReq.TaskGroups {
		for j, tg := range taskGroups {
			if tgReq.Name == tg.Name {
				if i == j {
					continue
				}

				return model.Data{}, errors.New("Duplicated task group name: " + tg.Name)
			}
		}
	}

	for i, tCheck := range allTasks {
		for j, t := range allTasks {
			if tCheck.Name == t.Name {
				if i == j {
					continue
				}

				return model.Data{}, errors.New("Duplicated task name: " + t.Name)
			}
		}
	}

	return model.Data{
		Description: createDataReq.Description,
		TaskGroups:  taskGroups,
	}, nil
}

// CreateWorkflow godoc
//
//	@ID				create-workflow
//	@Summary		Create Workflow
//	@Description	Create a workflow.
//	@Tags			[Workflow]
//	@Accept			json
//	@Produce		json
//	@Param			request body 	model.CreateWorkflowReq true "Workflow content"
//	@Success		200	{object}	model.WorkflowTemplate	"Successfully create the workflow."
//	@Failure		400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure		500	{object}	common.ErrorResponse	"Failed to create workflow."
//	@Router			/workflow [post]
func CreateWorkflow(c echo.Context) error {
	var createWorkflowReq model.CreateWorkflowReq

	data, err := common.GetJSONRawBody(c)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Metadata: nil,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			toTimeHookFunc()),
		Result: &createWorkflowReq,
	})
	if err != nil {
		return err
	}

	err = decoder.Decode(data)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	if createWorkflowReq.Name == "" {
		return common.ReturnErrorMsg(c, "Please provide the name.")
	}

	workflowData, err := createDataReqToData(createWorkflowReq.Data)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	var workflow model.Workflow
	workflow.ID = uuid.New().String()
	workflow.Name = createWorkflowReq.Name
	workflow.Data = workflowData

	_, err = dao.WorkflowCreate(&workflow)
	if err != nil {
		{
			return common.ReturnErrorMsg(c, err.Error())
		}
	}

	err = airflow.Client.CreateDAG(&workflow)
	if err != nil {
		return common.ReturnErrorMsg(c, "Failed to create the workflow. (Error:"+err.Error()+")")
	}

	for _, tg := range workflow.Data.TaskGroups {
		_, err = dao.TaskGroupCreate(&model.TaskGroupDBModel{
			ID:         tg.ID,
			Name:       tg.Name,
			WorkflowID: workflow.ID,
		})
		if err != nil {
			return common.ReturnErrorMsg(c, err.Error())
		}

		for _, t := range tg.Tasks {
			_, err = dao.TaskCreate(&model.TaskDBModel{
				ID:          t.ID,
				Name:        t.Name,
				WorkflowID:  workflow.ID,
				TaskGroupID: tg.ID,
			})
			if err != nil {
				return common.ReturnErrorMsg(c, err.Error())
			}
		}
	}

	return c.JSONPretty(http.StatusOK, workflow, " ")
}

// GetWorkflow godoc
//
//	@ID				get-workflow
//	@Summary		Get Workflow
//	@Description	Get the workflow.
//	@Tags			[Workflow]
//	@Accept			json
//	@Produce		json
//	@Param			wfId path string true "ID of the workflow."
//	@Success		200	{object}	model.Workflow			"Successfully get the workflow."
//	@Failure		400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure		500	{object}	common.ErrorResponse	"Failed to get the workflow."
//	@Router			/workflow/{wfId} [get]
func GetWorkflow(c echo.Context) error {
	wfId := c.Param("wfId")
	if wfId == "" {
		return common.ReturnErrorMsg(c, "Please provide the wfId.")
	}

	workflow, err := dao.WorkflowGet(wfId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	for i, tg := range workflow.Data.TaskGroups {
		_, err = dao.TaskGroupGetByWorkflowIDAndName(wfId, tg.Name)
		if err != nil {
			logger.Println(logger.ERROR, true, err)
		}

		workflow.Data.TaskGroups[i].ID = tg.ID

		for j, t := range tg.Tasks {
			_, err = dao.TaskGetByWorkflowIDAndName(wfId, tg.Name)
			if err != nil {
				logger.Println(logger.ERROR, true, err)
			}

			workflow.Data.TaskGroups[i].Tasks[j].ID = t.ID
		}
	}

	_, err = airflow.Client.GetDAG(wfId)
	if err != nil {
		return common.ReturnErrorMsg(c, "Failed to get the workflow from the airflow server.")
	}

	return c.JSONPretty(http.StatusOK, workflow, " ")
}

// GetWorkflowByName godoc
//
//	@ID				get-workflow-by-name
//	@Summary		Get Workflow by Name
//	@Description	Get the workflow by name.
//	@Tags			[Workflow]
//	@Accept			json
//	@Produce		json
//	@Param			wfName path string true "Name of the workflow."
//	@Success		200	{object}	model.Workflow			"Successfully get the workflow."
//	@Failure		400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure		500	{object}	common.ErrorResponse	"Failed to get the workflow."
//	@Router			/workflow/name/{wfName} [get]
func GetWorkflowByName(c echo.Context) error {
	wfName := c.Param("wfName")
	if wfName == "" {
		return common.ReturnErrorMsg(c, "Please provide the wfName.")
	}

	workflow, err := dao.WorkflowGetByName(wfName)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	for i, tg := range workflow.Data.TaskGroups {
		_, err = dao.TaskGroupGetByWorkflowIDAndName(workflow.ID, tg.Name)
		if err != nil {
			logger.Println(logger.ERROR, true, err)
		}

		workflow.Data.TaskGroups[i].ID = tg.ID

		for j, t := range tg.Tasks {
			_, err = dao.TaskGetByWorkflowIDAndName(workflow.ID, tg.Name)
			if err != nil {
				logger.Println(logger.ERROR, true, err)
			}

			workflow.Data.TaskGroups[i].Tasks[j].ID = t.ID
		}
	}

	_, err = airflow.Client.GetDAG(workflow.ID)
	if err != nil {
		return common.ReturnErrorMsg(c, "Failed to get the workflow from the airflow server.")
	}

	return c.JSONPretty(http.StatusOK, workflow, " ")
}

// ListWorkflow godoc
//
//	@ID				list-workflow
//	@Summary		List Workflow
//	@Description	Get a workflow list.
//	@Tags			[Workflow]
//	@Accept			json
//	@Produce		json
//	@Param			name query string false "Name of the workflow"
//	@Param			page query string false "Page of the workflow list."
//	@Param			row query string false "Row of the workflow list."
//	@Success		200	{object}	[]model.Workflow		"Successfully get a workflow list."
//	@Failure		400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure		500	{object}	common.ErrorResponse	"Failed to get a workflow list."
//	@Router			/workflow [get]
func ListWorkflow(c echo.Context) error {
	page, row, err := common.CheckPageRow(c)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	workflow := &model.Workflow{
		Name: c.QueryParam("name"),
	}

	workflows, err := dao.WorkflowGetList(workflow, page, row)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	for i, w := range *workflows {
		for j, tg := range workflow.Data.TaskGroups {
			_, err = dao.TaskGroupGetByWorkflowIDAndName(w.ID, tg.Name)
			if err != nil {
				logger.Println(logger.ERROR, true, err)
			}

			(*workflows)[i].Data.TaskGroups[j].ID = tg.ID

			for k, t := range tg.Tasks {
				_, err = dao.TaskGetByWorkflowIDAndName(w.ID, tg.Name)
				if err != nil {
					logger.Println(logger.ERROR, true, err)
				}

				(*workflows)[i].Data.TaskGroups[j].Tasks[k].ID = t.ID
			}
		}
	}

	return c.JSONPretty(http.StatusOK, workflows, " ")
}

// RunWorkflow godoc
//
//	@ID				run-workflow
//	@Summary		Run Workflow
//	@Description	Run the workflow.
//	@Tags			[Workflow]
//	@Accept			json
//	@Produce		json
//	@Param			wfId path string true "ID of the workflow."
//	@Success		200	{object}	model.SimpleMsg			"Successfully run the workflow."
//	@Failure		400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure		500	{object}	common.ErrorResponse	"Failed to run the Workflow"
//	@Router			/workflow/{wfId}/run [post]
func RunWorkflow(c echo.Context) error {
	wfId := c.Param("wfId")
	if wfId == "" {
		return common.ReturnErrorMsg(c, "Please provide the id.")
	}

	workflow, err := dao.WorkflowGet(wfId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	_, err = airflow.Client.RunDAG(workflow.ID)
	if err != nil {
		return common.ReturnInternalError(c, err, "Failed to run the workflow.")
	}

	return c.JSONPretty(http.StatusOK, model.SimpleMsg{Message: "success"}, " ")
}

// UpdateWorkflow godoc
//
//	@ID				update-workflow
//	@Summary		Update Workflow
//	@Description	Update the workflow content.
//	@Tags			[Workflow]
//	@Accept			json
//	@Produce		json
//	@Param			wfId path string true "ID of the workflow."
//	@Param			Workflow body 	model.CreateWorkflowReq true "Workflow to modify."
//	@Success		200	{object}	model.Workflow	"Successfully update the workflow"
//	@Failure		400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure		500	{object}	common.ErrorResponse	"Failed to update the workflow"
//	@Router			/workflow/{wfId} [put]
func UpdateWorkflow(c echo.Context) error {
	var updateWorkflowReq model.CreateWorkflowReq

	data, err := common.GetJSONRawBody(c)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Metadata: nil,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			toTimeHookFunc()),
		Result: &updateWorkflowReq,
	})
	if err != nil {
		return err
	}

	err = decoder.Decode(data)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	wfId := c.Param("wfId")
	oldWorkflow, err := dao.WorkflowGet(wfId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	if updateWorkflowReq.Name != "" {
		oldWorkflow.Name = updateWorkflowReq.Name
	}

	workflowData, err := createDataReqToData(updateWorkflowReq.Data)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	// Remove old task groups and tasks from the database
	for _, tg := range oldWorkflow.Data.TaskGroups {
		taskGroup, err := dao.TaskGroupGet(tg.ID)
		if err != nil {
			logger.Println(logger.ERROR, true, err)
		}
		err = dao.TaskGroupDelete(taskGroup)
		if err != nil {
			logger.Println(logger.ERROR, true, err)
		}

		for _, t := range tg.Tasks {
			task, err := dao.TaskGet(t.ID)
			if err != nil {
				logger.Println(logger.ERROR, true, err)
			}
			err = dao.TaskDelete(task)
			if err != nil {
				logger.Println(logger.ERROR, true, err)
			}
		}
	}

	// Create task groups and tasks to the database
	for _, tg := range workflowData.TaskGroups {
		_, err = dao.TaskGroupCreate(&model.TaskGroupDBModel{
			ID:         tg.ID,
			Name:       tg.Name,
			WorkflowID: wfId,
		})
		if err != nil {
			return common.ReturnErrorMsg(c, err.Error())
		}

		for _, t := range tg.Tasks {
			_, err = dao.TaskCreate(&model.TaskDBModel{
				ID:          t.ID,
				Name:        t.Name,
				WorkflowID:  wfId,
				TaskGroupID: tg.ID,
			})
			if err != nil {
				return common.ReturnErrorMsg(c, err.Error())
			}
		}
	}

	oldWorkflow.Data = workflowData

	err = dao.WorkflowUpdate(oldWorkflow)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	err = airflow.Client.DeleteDAG(oldWorkflow.ID, true)
	if err != nil {
		return common.ReturnErrorMsg(c, "Failed to update the workflow. (Error:"+err.Error()+")")
	}

	err = airflow.Client.CreateDAG(oldWorkflow)
	if err != nil {
		return common.ReturnErrorMsg(c, "Failed to update the workflow. (Error:"+err.Error()+")")
	}

	return c.JSONPretty(http.StatusOK, oldWorkflow, " ")
}

// DeleteWorkflow godoc
//
//	@ID				delete-workflow
//	@Summary		Delete Workflow
//	@Description	Delete the workflow.
//	@Tags			[Workflow]
//	@Accept			json
//	@Produce		json
//	@Param			wfId path string true "ID of the workflow."
//	@Success		200	{object}	model.SimpleMsg			"Successfully delete the workflow"
//	@Failure		400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure		500	{object}	common.ErrorResponse	"Failed to delete the workflow"
//	@Router			/workflow/{wfId} [delete]
func DeleteWorkflow(c echo.Context) error {
	wfId := c.Param("wfId")
	if wfId == "" {
		return common.ReturnErrorMsg(c, "Please provide the wfId.")
	}

	workflow, err := dao.WorkflowGet(wfId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	err = airflow.Client.DeleteDAG(workflow.ID, false)
	if err != nil {
		logger.Println(logger.ERROR, true, "AIRFLOW: "+err.Error())
	}

	for _, tg := range workflow.Data.TaskGroups {
		taskGroup, err := dao.TaskGroupGet(tg.ID)
		if err != nil {
			logger.Println(logger.ERROR, true, err)
		}
		err = dao.TaskGroupDelete(taskGroup)
		if err != nil {
			logger.Println(logger.ERROR, true, err)
		}

		for _, t := range tg.Tasks {
			task, err := dao.TaskGet(t.ID)
			if err != nil {
				logger.Println(logger.ERROR, true, err)
			}
			err = dao.TaskDelete(task)
			if err != nil {
				logger.Println(logger.ERROR, true, err)
			}
		}
	}

	err = dao.WorkflowDelete(workflow)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, model.SimpleMsg{Message: "success"}, " ")
}

// ListTaskGroup godoc
//
//	@ID				list-task-group
//	@Summary		List TaskGroup
//	@Description	Get a task group list of the workflow.
//	@Tags			[Workflow]
//	@Accept			json
//	@Produce		json
//	@Param			wfId path string true "ID of the workflow."
//	@Success		200	{object}	[]model.TaskGroup		"Successfully get a task group list."
//	@Failure		400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure		500	{object}	common.ErrorResponse	"Failed to get a task group list."
//	@Router			/workflow/{wfId}/task_group [get]
func ListTaskGroup(c echo.Context) error {
	wfId := c.Param("wfId")
	if wfId == "" {
		return common.ReturnErrorMsg(c, "Please provide the wfId.")
	}

	workflow, err := dao.WorkflowGet(wfId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	var taskGroups []model.TaskGroup
	taskGroups = append(taskGroups, workflow.Data.TaskGroups...)

	return c.JSONPretty(http.StatusOK, taskGroups, " ")
}

// GetTaskGroup godoc
//
//	@ID				get-task-group
//	@Summary		Get TaskGroup
//	@Description	Get the task group.
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce		json
//	@Param		wfId path string true "ID of the workflow."
//	@Param		tgId path string true "ID of the task group."
//	@Success		200	{object}	model.Task				"Successfully get the task group."
//	@Failure		400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure		500	{object}	common.ErrorResponse	"Failed to get the task group."
//	@Router		/workflow/{wfId}/task_group/{tgId} [get]
func GetTaskGroup(c echo.Context) error {
	wfId := c.Param("wfId")
	if wfId == "" {
		return common.ReturnErrorMsg(c, "Please provide the wfId.")
	}

	tgId := c.Param("tgId")
	if tgId == "" {
		return common.ReturnErrorMsg(c, "Please provide the tgId.")
	}

	workflow, err := dao.WorkflowGet(wfId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	for _, tg := range workflow.Data.TaskGroups {
		if tg.ID == tgId {
			return c.JSONPretty(http.StatusOK, tg, " ")
		}
	}

	return common.ReturnErrorMsg(c, "Task group not found.")
}

// GetTaskGroupDirectly godoc
//
//	@ID				get-task-group-directly
//	@Summary		Get TaskGroup Directly
//	@Description	Get the task group directly.
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce		json
//	@Param		tgId path string true "ID of the task group."
//	@Success		200	{object}	model.Task				"Successfully get the task group."
//	@Failure		400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure		500	{object}	common.ErrorResponse	"Failed to get the task group."
//	@Router		/task_group/{tgId} [get]
func GetTaskGroupDirectly(c echo.Context) error {
	tgId := c.Param("tgId")
	if tgId == "" {
		return common.ReturnErrorMsg(c, "Please provide the tgId.")
	}

	tgDB, err := dao.TaskGroupGet(tgId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	workflow, err := dao.WorkflowGet(tgDB.WorkflowID)
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

	return common.ReturnErrorMsg(c, "task group not found.")
}

// ListTaskFromTaskGroup godoc
//
//	@ID				list-task-from-task-group
//	@Summary		List Task from Task Group
//	@Description	Get a task list from the task group.
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce		json
//	@Param		wfId path string true "ID of the workflow."
//	@Param		tgId path string true "ID of the task group."
//	@Success		200	{object}	[]model.Task			"Successfully get a task list from the task group."
//	@Failure		400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure		500	{object}	common.ErrorResponse	"Failed to get a task list from the task group."
//	@Router		/workflow/{wfId}/task_group/{tgId}/task [get]
func ListTaskFromTaskGroup(c echo.Context) error {
	wfId := c.Param("wfId")
	if wfId == "" {
		return common.ReturnErrorMsg(c, "Please provide the wfId.")
	}

	tgId := c.Param("tgId")
	if tgId == "" {
		return common.ReturnErrorMsg(c, "Please provide the tgId.")
	}

	workflow, err := dao.WorkflowGet(wfId)
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

	return c.JSONPretty(http.StatusOK, tasks, " ")
}

// GetTaskFromTaskGroup godoc
//
//	@ID				get-task-from-task-group
//	@Summary		Get Task from Task Group
//	@Description	Get the task from the task group.
//	@Tags			[Workflow]
//	@Accept			json
//	@Produce		json
//	@Param			wfId path string true "ID of the workflow."
//	@Param			tgId path string true "ID of the task group."
//	@Param			taskId path string true "ID of the task."
//	@Success		200	{object}	model.Task			"Successfully get the task from the task group."
//	@Failure		400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure		500	{object}	common.ErrorResponse	"Failed to get the task from the task group."
//	@Router			/workflow/{wfId}/task_group/{tgId}/task/{taskId} [get]
func GetTaskFromTaskGroup(c echo.Context) error {
	wfId := c.Param("wfId")
	if wfId == "" {
		return common.ReturnErrorMsg(c, "Please provide the wfId.")
	}

	tgId := c.Param("tgId")
	if tgId == "" {
		return common.ReturnErrorMsg(c, "Please provide the tgId.")
	}

	taskId := c.Param("taskId")
	if taskId == "" {
		return common.ReturnErrorMsg(c, "Please provide the taskId.")
	}

	workflow, err := dao.WorkflowGet(wfId)
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

	return common.ReturnErrorMsg(c, "Task not found.")
}

// ListTask godoc
//
//	@ID				list-task
//	@Summary		List Task
//	@Description	Get a task list of the workflow.
//	@Tags			[Workflow]
//	@Accept			json
//	@Produce		json
//	@Param			wfId path string true "ID of the workflow."
//	@Success		200	{object}	[]model.Task			"Successfully get a task list."
//	@Failure		400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure		500	{object}	common.ErrorResponse	"Failed to get a task list."
//	@Router			/workflow/{wfId}/task [get]
func ListTask(c echo.Context) error {
	wfId := c.Param("wfId")
	if wfId == "" {
		return common.ReturnErrorMsg(c, "Please provide the wfId.")
	}

	workflow, err := dao.WorkflowGet(wfId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	var tasks []model.Task
	for _, tg := range workflow.Data.TaskGroups {
		tasks = append(tasks, tg.Tasks...)
	}

	return c.JSONPretty(http.StatusOK, tasks, " ")
}

// GetTask godoc
//
//	@ID				get-task
//	@Summary		Get Task
//	@Description	Get the task.
//	@Tags			[Workflow]
//	@Accept			json
//	@Produce		json
//	@Param			wfId path string true "ID of the workflow."
//	@Param			taskId path string true "ID of the task."
//	@Success		200	{object}	model.Task				"Successfully get the task."
//	@Failure		400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure		500	{object}	common.ErrorResponse	"Failed to get the task."
//	@Router			/workflow/{wfId}/task/{taskId} [get]
func GetTask(c echo.Context) error {
	wfId := c.Param("wfId")
	if wfId == "" {
		return common.ReturnErrorMsg(c, "Please provide the wfId.")
	}

	taskId := c.Param("taskId")
	if taskId == "" {
		return common.ReturnErrorMsg(c, "Please provide the taskId.")
	}

	workflow, err := dao.WorkflowGet(wfId)
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

	return common.ReturnErrorMsg(c, "Task not found.")
}

// GetTaskDirectly godoc
//
//	@ID				get-task-directly
//	@Summary		Get Task Directly
//	@Description	Get the task directly.
//	@Tags			[Workflow]
//	@Accept			json
//	@Produce		json
//	@Param			taskId path string true "ID of the task."
//	@Success		200	{object}	model.TaskDirectly		"Successfully get the task."
//	@Failure		400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure		500	{object}	common.ErrorResponse	"Failed to get the task."
//	@Router			/task/{taskId} [get]
func GetTaskDirectly(c echo.Context) error {
	taskId := c.Param("taskId")
	if taskId == "" {
		return common.ReturnErrorMsg(c, "Please provide the taskId.")
	}

	tDB, err := dao.TaskGet(taskId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	tgDB, err := dao.TaskGroupGet(tDB.TaskGroupID)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	workflow, err := dao.WorkflowGet(tgDB.WorkflowID)
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
						Dependencies:  task.Dependencies,
					}, " ")
				}
			}
		}
	}

	return common.ReturnErrorMsg(c, "task not found.")
}
