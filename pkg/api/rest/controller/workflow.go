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
				Name:          tReq.Name,
				TaskComponent: tReq.TaskComponent,
				RequestBody:   tReq.RequestBody,
				Dependencies:  tReq.Dependencies,
			})
			allTasks = append(allTasks, tasks...)
		}

		taskGroups = append(taskGroups, model.TaskGroup{
			ID:          uuid.New().String(),
			Name:        tgReq.Name,
			Description: tgReq.Description,
			Tasks:       tasks,
		})
	}

	for i, tgReq := range createDataReq.TaskGroups {
		for j, tg := range taskGroups {
			if i == j {
				continue
			}

			if tgReq.Name == tg.Name {
				return model.Data{}, errors.New("Duplicated task group name: " + tg.Name)
			}
		}
	}

	for i, tCheck := range allTasks {
		for j, t := range allTasks {
			if i == j {
				continue
			}

			if tCheck.Name == t.Name {
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
// @Summary		Create Workflow
// @Description	Create a workflow.
// @Tags		[Workflow]
// @Accept		json
// @Produce		json
// @Param		request body 	model.CreateWorkflowReq true "Workflow content"
// @Success		200	{object}	model.WorkflowTemplate	"Successfully create the workflow."
// @Failure		400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse	"Failed to create DAG."
// @Router		/cicada/workflow [post]
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

	err = airflow.Client.CreateDAG(&workflow)
	if err != nil {
		return common.ReturnErrorMsg(c, "Failed to create the workflow.")
	}

	_, err = dao.WorkflowCreate(&workflow)
	if err != nil {
		{
			return common.ReturnErrorMsg(c, err.Error())
		}
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
// @Summary		Get Workflow
// @Description	Get the workflow.
// @Tags		[Workflow]
// @Accept		json
// @Produce		json
// @Param		wfId path string true "ID of the workflow."
// @Success		200	{object}	model.Workflow			"Successfully get the workflow."
// @Failure		400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse	"Failed to get the workflow."
// @Router		/cicada/workflow/{wfId} [get]
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
		logger.Println(logger.ERROR, true, err)

		workflow.Data.TaskGroups[i].ID = tg.ID

		for j, t := range tg.Tasks {
			_, err = dao.TaskGetByWorkflowIDAndName(wfId, tg.Name)
			logger.Println(logger.ERROR, true, err)

			workflow.Data.TaskGroups[i].Tasks[j].ID = t.ID
		}
	}

	_, err = airflow.Client.GetDAG(wfId)
	if err != nil {
		return common.ReturnErrorMsg(c, "Failed to get the workflow from the airflow server.")
	}

	return c.JSONPretty(http.StatusOK, workflow, " ")
}

// ListWorkflow godoc
//
// @Summary		List Workflow
// @Description	Get a workflow list.
// @Tags		[Workflow]
// @Accept		json
// @Produce		json
// @Param		name path string false "Name of the workflow"
// @Param		page query string false "Page of the workflow list."
// @Param		row query string false "Row of the workflow list."
// @Success		200	{object}	[]model.Workflow		"Successfully get a workflow list."
// @Failure		400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse	"Failed to get a workflow list."
// @Router		/cicada/workflow [get]
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
			logger.Println(logger.ERROR, true, err)

			(*workflows)[i].Data.TaskGroups[j].ID = tg.ID

			for k, t := range tg.Tasks {
				_, err = dao.TaskGetByWorkflowIDAndName(w.ID, tg.Name)
				logger.Println(logger.ERROR, true, err)

				(*workflows)[i].Data.TaskGroups[j].Tasks[k].ID = t.ID
			}
		}
	}

	return c.JSONPretty(http.StatusOK, workflows, " ")
}

// RunWorkflow godoc
//
// @Summary		Run Workflow
// @Description	Run the workflow.
// @Tags		[Workflow]
// @Accept		json
// @Produce		json
// @Param		wfId path string true "ID of the workflow."
// @Success		200	{object}	model.SimpleMsg			"Successfully run the workflow."
// @Failure		400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse	"Failed to run the Workflow"
// @Router		/cicada/workflow/{wfId}/run [post]
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
// @Summary		Update Workflow
// @Description	Update the workflow content.
// @Tags		[Workflow]
// @Accept		json
// @Produce		json
// @Param		wfId path string true "ID of the workflow."
// @Param		Workflow body 	model.UpdateWorkflowReq true "Workflow to modify."
// @Success		200	{object}	model.Workflow	"Successfully update the workflow"
// @Failure		400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse	"Failed to update the workflow"
// @Router		/cicada/workflow/{wfId} [put]
func UpdateWorkflow(c echo.Context) error {
	Workflow := new(model.Workflow)
	err := c.Bind(Workflow)
	if err != nil {
		return err
	}

	Workflow.ID = c.Param("wfId")
	_, err = dao.WorkflowGet(Workflow.ID)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	err = dao.WorkflowUpdate(Workflow)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, Workflow, " ")
}

// DeleteWorkflow godoc
//
// @Summary		Delete Workflow
// @Description	Delete the workflow.
// @Tags		[Workflow]
// @Accept		json
// @Produce		json
// @Param		wfId path string true "ID of the workflow."
// @Success		200	{object}	model.SimpleMsg			"Successfully delete the workflow"
// @Failure		400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse	"Failed to delete the workflow"
// @Router		/cicada/workflow/{wfId} [delete]
func DeleteWorkflow(c echo.Context) error {
	wfId := c.Param("wfId")
	if wfId == "" {
		return common.ReturnErrorMsg(c, "Please provide the wfId.")
	}

	workflow, err := dao.WorkflowGet(wfId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	err = airflow.Client.DeleteDAG(workflow.ID)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	for _, tg := range workflow.Data.TaskGroups {
		taskGroup, err := dao.TaskGroupGet(tg.ID)
		logger.Println(logger.ERROR, true, err)
		err = dao.TaskGroupDelete(taskGroup)
		logger.Println(logger.ERROR, true, err)

		for _, t := range tg.Tasks {
			task, err := dao.TaskGet(t.ID)
			logger.Println(logger.ERROR, true, err)
			err = dao.TaskDelete(task)
			logger.Println(logger.ERROR, true, err)
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
// @Summary		List TaskGroup
// @Description	Get a task group list of the workflow.
// @Tags		[Workflow]
// @Accept		json
// @Produce		json
// @Param		wfId path string true "ID of the workflow."
// @Success		200	{object}	[]model.TaskGroup		"Successfully get a task group list."
// @Failure		400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse	"Failed to get a task group list."
// @Router		/cicada/workflow/{wfId}/task_group [get]
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
// @Summary		Get TaskGroup
// @Description	Get the task group.
// @Tags		[Workflow]
// @Accept		json
// @Produce		json
// @Param		wfId path string true "ID of the workflow."
// @Param		tgId path string true "ID of the task group."
// @Success		200	{object}	model.Task				"Successfully get the task group."
// @Failure		400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse	"Failed to get the task group."
// @Router		/cicada/workflow/{wfId}/task_group/{tgId} [get]
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

// ListTaskFromTaskGroup godoc
//
// @Summary		List Task from Task Group
// @Description	Get a task list from the task group.
// @Tags		[Workflow]
// @Accept		json
// @Produce		json
// @Param		wfId path string true "ID of the workflow."
// @Param		tgId path string true "ID of the task group."
// @Success		200	{object}	[]model.Task			"Successfully get a task list from the task group."
// @Failure		400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse	"Failed to get a task list from the task group."
// @Router		/cicada/workflow/{wfId}/task_group/{tgId}/task [get]
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
// @Summary		Get Task from Task Group
// @Description	Get the task from the task group.
// @Tags		[Workflow]
// @Accept		json
// @Produce		json
// @Param		wfId path string true "ID of the workflow."
// @Param		tgId path string true "ID of the task group."
// @Param		taskId path string true "ID of the task."
// @Success		200	{object}	model.Task			"Successfully get the task from the task group."
// @Failure		400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse	"Failed to get the task from the task group."
// @Router		/cicada/workflow/{wfId}/task_group/{tgId}/task/{taskId} [get]
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
// @Summary		List Task
// @Description	Get a task list of the workflow.
// @Tags		[Workflow]
// @Accept		json
// @Produce		json
// @Param		wfId path string true "ID of the workflow."
// @Success		200	{object}	[]model.Task			"Successfully get a task list."
// @Failure		400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse	"Failed to get a task list."
// @Router		/cicada/workflow/{wfId}/task [get]
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
// @Summary		Get Task
// @Description	Get the task.
// @Tags		[Workflow]
// @Accept		json
// @Produce		json
// @Param		wfId path string true "ID of the workflow."
// @Param		taskId path string true "ID of the task."
// @Success		200	{object}	model.Task				"Successfully get the task."
// @Failure		400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse	"Failed to get the task."
// @Router		/cicada/workflow/{wfId}/task/{taskId} [get]
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
