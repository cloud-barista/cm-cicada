package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/cloud-barista/cm-cicada/dao"
	"github.com/cloud-barista/cm-cicada/lib/airflow"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/google/uuid"
	"github.com/jollaman999/utils/logger"
	"github.com/labstack/echo/v4"
	"github.com/mitchellh/mapstructure"
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

func createDataReqToData(specVersion string, createDataReq model.CreateDataReq) (model.Data, error) {
	specVersionSpilit := strings.Split(specVersion, ".")
	if len(specVersionSpilit) != 2 {
		return model.Data{}, errors.New("invalid workflow spec version: " + specVersion)
	}

	specVersionMajor, err := strconv.Atoi(specVersionSpilit[0])
	if err != nil {
		return model.Data{}, errors.New("invalid workflow spec version: " + specVersion)
	}

	specVersionMinor, err := strconv.Atoi(specVersionSpilit[1])
	if err != nil {
		return model.Data{}, errors.New("invalid workflow spec version: " + specVersion)
	}

	var taskGroups []model.TaskGroup
	var allTasks []model.Task

	if specVersionMajor > 0 && specVersionMajor <= 1 {
		if specVersionMinor == 0 {
			// v1.0
			for _, tgReq := range createDataReq.TaskGroups {
				var tasks []model.Task
				for _, tReq := range tgReq.Tasks {
					tasks = append(tasks, model.Task{
						ID:            uuid.New().String(),
						Name:          tReq.Name,
						TaskComponent: tReq.TaskComponent,
						RequestBody:   tReq.RequestBody,
						PathParams:    tReq.PathParams,
						QueryParams:   tReq.QueryParams,
						Extra:         tReq.Extra,
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
		} else {
			return model.Data{}, errors.New("Unsupported workflow spec version: " + specVersion)
		}
	} else {
		return model.Data{}, errors.New("Unsupported workflow spec version: " + specVersion)
	}

	return model.Data{
		Description: createDataReq.Description,
		TaskGroups:  taskGroups,
	}, nil
}

func workflowDagID(workflow *model.Workflow) string {
	if workflow.WorkflowKey != "" {
		return workflow.WorkflowKey
	}
	return workflow.ID
}

func taskAirflowID(task *model.TaskDBModel) string {
	if task.TaskKey != "" {
		return task.TaskKey
	}
	if task.ID != "" {
		return task.ID
	}
	return task.Name
}

type workflowGraphDiff struct {
	workflowData         model.Data
	taskGroupsToUpsert   []model.TaskGroupDBModel
	tasksToUpsert        []model.TaskDBModel
	taskGroupsToSoftDrop []model.TaskGroupDBModel
	tasksToSoftDrop      []model.TaskDBModel
}

func buildWorkflowGraphDiff(workflow *model.Workflow, incoming model.Data) (*workflowGraphDiff, error) {
	workflowKey := workflowDagID(workflow)
	taskGroupsFromDB, err := dao.TaskGroupGetListByWorkflowID(workflow.ID, true)
	if err != nil {
		return nil, err
	}
	tasksFromDB, err := dao.TaskGetListByWorkflowID(workflow.ID, true)
	if err != nil {
		return nil, err
	}

	taskGroupByName := make(map[string]model.TaskGroupDBModel)
	activeTaskGroups := make(map[string]model.TaskGroupDBModel)
	for _, tg := range taskGroupsFromDB {
		current, exists := taskGroupByName[tg.Name]
		if !exists || (current.IsDeleted && !tg.IsDeleted) {
			taskGroupByName[tg.Name] = tg
		}
		if !tg.IsDeleted {
			activeTaskGroups[tg.ID] = tg
		}
	}

	taskByName := make(map[string]model.TaskDBModel)
	activeTasks := make(map[string]model.TaskDBModel)
	for _, t := range tasksFromDB {
		current, exists := taskByName[t.Name]
		if !exists || (current.IsDeleted && !t.IsDeleted) {
			taskByName[t.Name] = t
		}
		if !t.IsDeleted {
			activeTasks[t.ID] = t
		}
	}

	diff := &workflowGraphDiff{
		workflowData: model.Data{
			Description: incoming.Description,
			TaskGroups:  make([]model.TaskGroup, 0, len(incoming.TaskGroups)),
		},
	}
	seenTaskGroupIDs := make(map[string]bool)
	seenTaskIDs := make(map[string]bool)

	for _, incomingTG := range incoming.TaskGroups {
		resolvedTG := incomingTG
		taskGroupModel, exists := taskGroupByName[incomingTG.Name]
		if !exists {
			taskGroupModel = model.TaskGroupDBModel{
				ID:           uuid.New().String(),
				TaskGroupKey: uuid.New().String(),
			}
		}
		if taskGroupModel.TaskGroupKey == "" {
			taskGroupModel.TaskGroupKey = taskGroupModel.ID
		}
		resolvedTG.ID = taskGroupModel.ID
		resolvedTG.Tasks = make([]model.Task, 0, len(incomingTG.Tasks))

		taskGroupModel.Name = incomingTG.Name
		taskGroupModel.WorkflowID = workflow.ID
		taskGroupModel.WorkflowKey = workflowKey
		taskGroupModel.IsDeleted = false
		diff.taskGroupsToUpsert = append(diff.taskGroupsToUpsert, taskGroupModel)
		seenTaskGroupIDs[taskGroupModel.ID] = true

		for _, incomingTask := range incomingTG.Tasks {
			resolvedTask := incomingTask
			taskModel, exists := taskByName[incomingTask.Name]
			if !exists {
				taskModel = model.TaskDBModel{
					ID:      uuid.New().String(),
					TaskKey: uuid.New().String(),
				}
			}
			if taskModel.TaskKey == "" {
				taskModel.TaskKey = taskModel.ID
			}

			resolvedTask.ID = taskModel.ID
			resolvedTG.Tasks = append(resolvedTG.Tasks, resolvedTask)

			taskModel.Name = incomingTask.Name
			taskModel.WorkflowID = workflow.ID
			taskModel.WorkflowKey = workflowKey
			taskModel.TaskGroupID = taskGroupModel.ID
			taskModel.TaskGroupKey = taskGroupModel.TaskGroupKey
			taskModel.IsDeleted = false
			diff.tasksToUpsert = append(diff.tasksToUpsert, taskModel)
			seenTaskIDs[taskModel.ID] = true
		}

		diff.workflowData.TaskGroups = append(diff.workflowData.TaskGroups, resolvedTG)
	}

	for _, tg := range activeTaskGroups {
		if !seenTaskGroupIDs[tg.ID] {
			diff.taskGroupsToSoftDrop = append(diff.taskGroupsToSoftDrop, tg)
		}
	}
	for _, t := range activeTasks {
		if !seenTaskIDs[t.ID] {
			diff.tasksToSoftDrop = append(diff.tasksToSoftDrop, t)
		}
	}

	return diff, nil
}

func resolveCreateSourceType(specVersion string, createDataReq model.CreateDataReq) (string, string, error) {
	templates, err := dao.WorkflowTemplateGetList(&model.WorkflowTemplate{}, 0, 0)
	if err != nil {
		return "", "", err
	}

	reqJSON, err := json.Marshal(createDataReq)
	if err != nil {
		return "", "", err
	}

	for _, tmpl := range *templates {
		if tmpl.SpecVersion != specVersion {
			continue
		}

		tmplJSON, err := json.Marshal(tmpl.Data)
		if err != nil {
			return "", "", err
		}

		if string(reqJSON) == string(tmplJSON) {
			return "example", tmpl.ID, nil
		}
	}

	return "custom", "", nil
}

// CreateWorkflow godoc
//
//	@ID		create-workflow
//	@Summary	Create Workflow
//	@Description	Create a workflow.
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce	json
//	@Param		request body 	model.CreateWorkflowReq true "Workflow content"
//	@Success	200	{object}	model.WorkflowTemplate	"Successfully create the workflow."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to create workflow."
//	@Router		/workflow [post]
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

	var specVersion = model.WorkflowSpecVersion_LATEST
	if createWorkflowReq.SpecVersion != "" {
		specVersion = createWorkflowReq.SpecVersion
	}

	workflowData, err := createDataReqToData(specVersion, createWorkflowReq.Data)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	var workflow model.Workflow
	workflow.ID = uuid.New().String()
	workflow.WorkflowKey = uuid.New().String()
	workflow.SpecVersion = specVersion
	workflow.Name = createWorkflowReq.Name
	workflow.Data = workflowData

	var success bool
	_, err = dao.WorkflowCreate(&workflow)
	if err != nil {
		{
			return common.ReturnErrorMsg(c, err.Error())
		}
	}
	defer func() {
		if !success {
			_ = dao.TaskSoftDeleteByWorkflowID(workflow.ID)
			_ = dao.TaskGroupSoftDeleteByWorkflowID(workflow.ID)
			_ = dao.WorkflowDelete(&workflow)
		}
	}()

	for _, tg := range workflow.Data.TaskGroups {
		_, err = dao.TaskGroupCreate(&model.TaskGroupDBModel{
			ID:           tg.ID,
			Name:         tg.Name,
			WorkflowID:   workflow.ID,
			WorkflowKey:  workflow.WorkflowKey,
			TaskGroupKey: tg.ID,
		})
		if err != nil {
			return common.ReturnErrorMsg(c, err.Error())
		}

		for _, t := range tg.Tasks {
			_, err = dao.TaskCreate(&model.TaskDBModel{
				ID:           t.ID,
				Name:         t.Name,
				WorkflowID:   workflow.ID,
				WorkflowKey:  workflow.WorkflowKey,
				TaskGroupID:  tg.ID,
				TaskGroupKey: tg.ID,
				TaskKey:      t.ID,
			})
			if err != nil {
				return common.ReturnErrorMsg(c, err.Error())
			}
		}
	}

	sourceType, sourceTemplateID, err := resolveCreateSourceType(specVersion, createWorkflowReq.Data)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	_, err = dao.WorkflowCreateSnapshot(&workflow, "create", sourceType, sourceTemplateID)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	client, err := airflow.GetClient()
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	err = client.CreateDAG(&workflow)
	if err != nil {
		return common.ReturnErrorMsg(c, "Failed to create the workflow. (Error:"+err.Error()+")")
	}
	success = true

	return c.JSONPretty(http.StatusOK, workflow, " ")
}

func getWorkflowFromDB(workflowID string) (*model.Workflow, error) {
	workflow, err := dao.WorkflowGet(workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get the workflow from DB. Error: %s", err.Error())
	}

	for i, tg := range workflow.Data.TaskGroups {
		tgDB, err := dao.TaskGroupGetByWorkflowIDAndName(workflowID, tg.Name)
		if err != nil {
			logger.Println(logger.ERROR, true, err)
		} else {
			workflow.Data.TaskGroups[i].ID = tgDB.ID
		}

		for j, t := range tg.Tasks {
			tDB, err := dao.TaskGetByWorkflowIDAndName(workflowID, t.Name)
			if err != nil {
				logger.Println(logger.ERROR, true, err)
			} else {
				workflow.Data.TaskGroups[i].Tasks[j].ID = tDB.ID
			}
		}
	}

	return workflow, nil
}

// GetWorkflow godoc
//
//	@ID		get-workflow
//	@Summary	Get Workflow
//	@Description	Get the workflow.
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce	json
//	@Param		wfId path string true "DB workflow ID."
//	@Success	200	{object}	model.Workflow		"Successfully get the workflow."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get the workflow."
//	@Router		/workflow/{wfId} [get]
func GetWorkflow(c echo.Context) error {
	wfId := c.Param("wfId")
	if wfId == "" {
		return common.ReturnErrorMsg(c, "Please provide the wfId.")
	}

	workflow, err := getWorkflowFromDB(wfId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	client, err := airflow.GetClient()
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	_, err = client.GetDAG(workflowDagID(workflow))
	if err != nil {
		return common.ReturnErrorMsg(c, "Failed to get the workflow from the airflow server.")
	}

	return c.JSONPretty(http.StatusOK, workflow, " ")
}

// GetWorkflowByName godoc
//
//	@ID		get-workflow-by-name
//	@Summary	Get Workflow by Name
//	@Description	Get the workflow by name.
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce	json
//	@Param		wfName path string true "Name of the workflow."
//	@Success	200	{object}	model.Workflow		"Successfully get the workflow."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get the workflow."
//	@Router		/workflow/name/{wfName} [get]
func GetWorkflowByName(c echo.Context) error {
	wfName := c.Param("wfName")
	if wfName == "" {
		return common.ReturnErrorMsg(c, "Please provide the wfName.")
	}

	workflowByName, err := dao.WorkflowGetByName(wfName)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	workflow, err := getWorkflowFromDB(workflowByName.ID)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	client, err := airflow.GetClient()
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	_, err = client.GetDAG(workflowDagID(workflow))
	if err != nil {
		return common.ReturnErrorMsg(c, "Failed to get the workflow from the airflow server.")
	}

	return c.JSONPretty(http.StatusOK, workflow, " ")
}

// ListWorkflow godoc
//
//	@ID		list-workflow
//	@Summary	List Workflow
//	@Description	Get a workflow list.
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce	json
//	@Param		name query string false "Name of the workflow"
//	@Param		page query string false "Page of the workflow list."
//	@Param		row query string false "Row of the workflow list."
//	@Success	200	{object}	[]model.Workflow	"Successfully get a workflow list."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get a workflow list."
//	@Router		/workflow [get]
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

	return c.JSONPretty(http.StatusOK, workflows, " ")
}

// RunWorkflow godoc
//
//	@ID		run-workflow
//	@Summary	Run Workflow
//	@Description	Run the workflow.
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce	json
//	@Param		wfId path string true "DB workflow ID."
//	@Success	200	{object}	model.SimpleMsg		"Successfully run the workflow."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to run the Workflow"
//	@Router		/workflow/{wfId}/run [post]
func RunWorkflow(c echo.Context) error {
	wfId := c.Param("wfId")
	if wfId == "" {
		return common.ReturnErrorMsg(c, "Please provide the id.")
	}

	workflow, err := dao.WorkflowGet(wfId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	client, err := airflow.GetClient()
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	conf := map[string]interface{}{
		"workflow_id":   workflow.ID,
		"workflow_key":  workflowDagID(workflow),
		"workflow_name": workflow.Name,
	}
	_, err = client.RunDAG(workflowDagID(workflow), conf)
	if err != nil {
		return common.ReturnInternalError(c, err, "Failed to run the workflow.")
	}

	return c.JSONPretty(http.StatusOK, model.SimpleMsg{Message: "success"}, " ")
}

// UpdateWorkflow godoc
//
//	@ID		update-workflow
//	@Summary	Update Workflow
//	@Description	Update the workflow content.
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce	json
//	@Param		wfId path string true "DB workflow ID."
//	@Param		Workflow body 	model.CreateWorkflowReq true "Workflow to modify."
//	@Success	200	{object}	model.Workflow	"Successfully update the workflow"
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to update the workflow"
//	@Router		/workflow/{wfId} [put]
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

	var specVersion = model.WorkflowSpecVersion_LATEST
	if updateWorkflowReq.SpecVersion != "" {
		specVersion = updateWorkflowReq.SpecVersion
	}

	workflowData, err := createDataReqToData(specVersion, updateWorkflowReq.Data)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	diff, err := buildWorkflowGraphDiff(oldWorkflow, workflowData)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	for _, tg := range diff.taskGroupsToUpsert {
		taskGroup := tg
		if err := dao.TaskGroupSave(&taskGroup); err != nil {
			return common.ReturnErrorMsg(c, err.Error())
		}
	}
	for _, t := range diff.tasksToUpsert {
		task := t
		if err := dao.TaskSave(&task); err != nil {
			return common.ReturnErrorMsg(c, err.Error())
		}
	}
	for _, t := range diff.tasksToSoftDrop {
		task := t
		if err := dao.TaskDelete(&task); err != nil {
			return common.ReturnErrorMsg(c, err.Error())
		}
	}
	for _, tg := range diff.taskGroupsToSoftDrop {
		taskGroup := tg
		if err := dao.TaskGroupDelete(&taskGroup); err != nil {
			return common.ReturnErrorMsg(c, err.Error())
		}
	}

	oldWorkflow.SpecVersion = specVersion
	oldWorkflow.Data = diff.workflowData

	err = dao.WorkflowUpdate(oldWorkflow)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	_, err = dao.WorkflowCreateSnapshot(oldWorkflow, "update", "modified", "")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	client, err := airflow.GetClient()
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	err = client.CreateDAG(oldWorkflow)
	if err != nil {
		return common.ReturnErrorMsg(c, "Failed to update the workflow. (Error:"+err.Error()+")")
	}

	return c.JSONPretty(http.StatusOK, oldWorkflow, " ")
}

// DeleteWorkflow godoc
//
//	@ID		delete-workflow
//	@Summary	Delete Workflow
//	@Description	Delete the workflow.
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce	json
//	@Param		wfId path string true "DB workflow ID."
//	@Success	200	{object}	model.SimpleMsg		"Successfully delete the workflow"
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to delete the workflow"
//	@Router		/workflow/{wfId} [delete]
func DeleteWorkflow(c echo.Context) error {
	wfId := c.Param("wfId")
	if wfId == "" {
		return common.ReturnErrorMsg(c, "Please provide the wfId.")
	}

	workflow, err := dao.WorkflowGet(wfId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	client, err := airflow.GetClient()
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	err = client.DeleteDAG(workflowDagID(workflow), true)
	if err != nil {
		logger.Println(logger.ERROR, true, "AIRFLOW: "+err.Error())
	}

	if err := dao.TaskSoftDeleteByWorkflowID(workflow.ID); err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	if err := dao.TaskGroupSoftDeleteByWorkflowID(workflow.ID); err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	err = dao.WorkflowDelete(workflow)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	workflow.IsDeleted = true
	now := time.Now()
	workflow.DeletedAt = &now
	_, err = dao.WorkflowCreateSnapshot(workflow, "delete", "custom", "")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, model.SimpleMsg{Message: "success"}, " ")
}

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
//	@ID		list-task-from-task-group
//	@Summary	List Task from Task Group
//	@Description	Get a task list from the task group.
//	@Tags	[Workflow]
//	@Accept	json
//	@Produce	json
//	@Param	wfId path string true "DB workflow ID."
//	@Param	tgId path string true "ID of the task group."
//	@Success	200	{object}	[]model.Task		"Successfully get a task list from the task group."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get a task list from the task group."
//	@Router	/workflow/{wfId}/task_group/{tgId}/task [get]
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
						PathParams:    task.PathParams,
						QueryParams:   task.QueryParams,
						Extra:         task.Extra,
						Dependencies:  task.Dependencies,
					}, " ")
				}
			}
		}
	}

	return common.ReturnErrorMsg(c, "task not found.")
}

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
	wfId := c.Param("wfId")
	if wfId == "" {
		return common.ReturnErrorMsg(c, "Please provide the wfId.")
	}
	wfRunId := c.Param("wfRunId")
	if wfRunId == "" {
		return common.ReturnErrorMsg(c, "Please provide the wfRunId.")
	}

	taskId := c.Param("taskId")
	if taskId == "" {
		return common.ReturnErrorMsg(c, "Please provide the taskId.")
	}
	taskInfo, err := dao.TaskGet(taskId)
	if err != nil {
		return common.ReturnErrorMsg(c, "Invalid get tasK from taskId.")
	}
	workflow, err := dao.WorkflowGet(wfId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	taskTryNum := c.Param("taskTryNum")
	if taskTryNum == "" {
		return common.ReturnErrorMsg(c, "Please provide the taskTryNum.")
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

// GetWorkflowRuns godoc
//
//	@ID			get-workflow-runs
//	@Summary	Get workflowRuns
//	@Description	Get the task Logs.
//	@Tags	[Workflow]
//	@Accept	json
//	@Produce	json
//	@Param	wfId path string true "DB workflow ID."
//	@Success	200	{object}	[]model.WorkflowRun		"Successfully get the workflowRuns."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get the workflowRuns."
//	@Router	 /workflow/{wfId}/runs [get]
func GetWorkflowRuns(c echo.Context) error {
	wfId := c.Param("wfId")
	if wfId == "" {
		return common.ReturnErrorMsg(c, "Please provide the wfId.")
	}
	workflow, err := dao.WorkflowGet(wfId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	client, err := airflow.GetClient()
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	runList, err := client.GetDAGRuns(workflowDagID(workflow))
	if err != nil {
		return common.ReturnErrorMsg(c, "Failed to get the workflow runs: "+err.Error())
	}

	var transformedRuns []model.WorkflowRun

	for _, dagRun := range *runList.DagRuns {
		transformedRun := model.WorkflowRun{
			WorkflowID:             dagRun.DagId,
			WorkflowRunID:          dagRun.GetDagRunId(),
			DataIntervalStart:      dagRun.GetDataIntervalStart(),
			DataIntervalEnd:        dagRun.GetDataIntervalEnd(),
			State:                  string(dagRun.GetState()),
			ExecutionDate:          dagRun.GetExecutionDate(),
			StartDate:              dagRun.GetStartDate(),
			EndDate:                dagRun.GetEndDate(),
			RunType:                dagRun.GetRunType(),
			LastSchedulingDecision: dagRun.GetLastSchedulingDecision(),
			DurationDate:           (dagRun.GetEndDate().Sub(dagRun.GetStartDate()).Seconds()),
		}
		transformedRuns = append(transformedRuns, transformedRun)
	}

	return c.JSONPretty(http.StatusOK, transformedRuns, " ")
}

// GetTaskInstances godoc
//
//	@ID			get-task-instances
//	@Summary	Get taskInstances
//	@Description	Get the task Logs.
//	@Tags	[Workflow]
//	@Accept	json
//	@Produce	json
//	@Param	wfId path string true "DB workflow ID."
//	@Param	wfRunId path string true "DB workflow ID."
//	@Success	200	{object}	model.TaskInstance		"Successfully get the taskInstances."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get the taskInstances."
//	@Router	 /workflow/{wfId}/workflowRun/{wfRunId}/taskInstances [get]
func GetTaskInstances(c echo.Context) error {
	wfId := c.Param("wfId")
	if wfId == "" {
		return common.ReturnErrorMsg(c, "Please provide the wfId.")
	}
	wfRunId := c.Param("wfRunId")
	if wfRunId == "" {
		return common.ReturnErrorMsg(c, "Please provide the wfRunId.")
	}
	workflow, err := getWorkflowFromDB(wfId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	client, err := airflow.GetClient()
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	runList, err := client.GetTaskInstances(workflowDagID(workflow), common.UrlDecode(wfRunId))
	if err != nil {
		return common.ReturnErrorMsg(c, "Failed to get the taskInstances: "+err.Error())
	}
	var taskInstances []model.TaskInstance
	layout := time.RFC3339Nano

	for _, taskInstance := range *runList.TaskInstances {
		taskDBInfo, err := dao.TaskGetByWorkflowKeyAndTaskKey(workflowDagID(workflow), taskInstance.GetTaskId())
		if err != nil {
			taskDBInfo, err = dao.TaskGetByWorkflowIDAndName(workflow.ID, taskInstance.GetTaskId())
		}
		if err != nil {
			return common.ReturnErrorMsg(c, "Failed to get the taskInstances: "+err.Error())
		}
		taskId := &taskDBInfo.ID
		executionDate, err := time.Parse(layout, taskInstance.GetExecutionDate())
		if err != nil {
			fmt.Println("Error parsing execution date:", err)
			continue
		}
		startDate, err := time.Parse(layout, taskInstance.GetExecutionDate())
		if err != nil {
			fmt.Println("Error parsing start date:", err)
			continue
		}
		endDate, err := time.Parse(layout, taskInstance.GetExecutionDate())
		if err != nil {
			fmt.Println("Error parsing end date:", err)
			continue
		}

		var isSoftwareMigrationTask bool
		var executionID string
		for _, tg := range workflow.Data.TaskGroups {
			for _, task := range tg.Tasks {
				if strings.Contains(task.TaskComponent, "grasshopper") &&
					strings.Contains(task.TaskComponent, "software") &&
					strings.Contains(task.TaskComponent, "migration") &&
					task.ID == *taskId {
					isSoftwareMigrationTask = true

					// software migration task인 경우 xcom에서 execution_id 조회
					xcomData, err := client.GetXComValue(
						taskInstance.GetDagId(),
						taskInstance.GetDagRunId(),
						taskInstance.GetTaskId(),
						"return_value",
					)
					if err != nil {
						logger.Println(logger.WARN, false,
							"Failed to get xcom data for task: "+taskInstance.GetTaskId()+" (Error: "+err.Error()+")")
					} else if xcomData != nil {
						if execID, ok := xcomData["execution_id"].(string); ok {
							executionID = execID
						}
					}
					break
				}
			}
		}

		taskInfo := model.TaskInstance{
			WorkflowID:                   taskInstance.DagId,
			WorkflowRunID:                taskInstance.GetDagRunId(),
			TaskID:                       *taskId,
			TaskName:                     taskDBInfo.Name,
			State:                        string(taskInstance.GetState()),
			ExecutionDate:                executionDate,
			StartDate:                    startDate,
			EndDate:                      endDate,
			DurationDate:                 float64(taskInstance.GetDuration()),
			TryNumber:                    int(taskInstance.GetTryNumber()),
			IsSoftwareMigrationTask:      isSoftwareMigrationTask,
			SoftwareMigrationExecutionID: executionID,
		}
		taskInstances = append(taskInstances, taskInfo)
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
//	@Param	wfId path string true "DB workflow ID."
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
			toTimeHookFunc()),
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
	if wfId == "" {
		return common.ReturnErrorMsg(c, "Please provide the wfId.")
	}
	wfRunId := c.Param("wfRunId")
	if wfRunId == "" {
		return common.ReturnErrorMsg(c, "Please provide the wfRunId.")
	}

	workflow, err := dao.WorkflowGet(wfId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	var taskKeyList []string
	for _, taskId := range taskClearOption.TaskIds {
		taskInfo, err := dao.TaskGet(taskId)
		if err != nil {
			return common.ReturnErrorMsg(c, fmt.Sprintf("failed to get task info for ID %s: %v", taskId, err))
		}
		taskKeyList = append(taskKeyList, taskAirflowID(taskInfo))
	}
	taskClearOption.TaskIds = taskKeyList
	if err := common.ValidateTaskClearOptions(taskClearOption); err != nil {
		fmt.Printf("옵션 검증 실패: %v\n", err)
		return common.ReturnErrorMsg(c, err.Error())
	}

	TaskInstanceReferences := make([]model.TaskInstanceReference, 0)
	client, err := airflow.GetClient()
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	clearList, err := client.ClearTaskInstance(workflowDagID(workflow), common.UrlDecode(wfRunId), taskClearOption)
	if err != nil {
		return common.ReturnErrorMsg(c, "Failed to get the taskInstances: "+err.Error())
	}
	logger.Println(logger.DEBUG, false, "clearList 요청 내용 : {} ", &clearList)
	if clearList.TaskInstances == nil || len(*clearList.TaskInstances) == 0 {
		logger.Println(logger.DEBUG, false, "TaskInstances is nil or empty")

	}
	for _, taskInstance := range *clearList.TaskInstances {
		taskDBInfo, err := dao.TaskGetByWorkflowKeyAndTaskKey(workflowDagID(workflow), taskInstance.GetTaskId())
		if err != nil {
			taskDBInfo, err = dao.TaskGetByWorkflowIDAndName(workflow.ID, taskInstance.GetTaskId())
		}
		if err != nil {
			return common.ReturnErrorMsg(c, "Failed to get the taskInstances: "+err.Error())
		}
		taskId := &taskDBInfo.ID
		taskInfo := model.TaskInstanceReference{
			WorkflowID:    taskInstance.DagId,
			WorkflowRunID: taskInstance.DagRunId,
			TaskId:        taskId,
			TaskName:      taskDBInfo.Name,
			ExecutionDate: taskInstance.ExecutionDate,
		}
		logger.Println(logger.DEBUG, false, "TaskInstanceReferences  ", TaskInstanceReferences)
		TaskInstanceReferences = append(TaskInstanceReferences, taskInfo)
	}
	logger.Println(logger.DEBUG, false, "TaskInstanceReferences ", TaskInstanceReferences)

	return c.JSONPretty(http.StatusOK, TaskInstanceReferences, " ")
}

// GetEventLogs godoc
//
//	@ID				get-event-logs
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
	wfId := c.Param("wfId")
	if wfId == "" {
		return common.ReturnErrorMsg(c, "Please provide the wfId.")
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
			taskDBInfo, err := dao.TaskGetByWorkflowKeyAndTaskKey(workflowDagID(workflow), eventlog.TaskID)
			if err != nil {
				taskDBInfo, err = dao.TaskGetByWorkflowIDAndName(wfId, eventlog.TaskID)
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

// GetImportErrors godoc
//
//	@ID			get-import-errors
//	@Summary	Get importErrors
//	@Description	Get the importErrors.
//	@Tags	[Workflow]
//	@Accept	json
//	@Produce	json
//	@Success	200	{object}	airflow.ImportErrorCollection		"Successfully get the importErrors."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get the importErrors."
//	@Router	 /importErrors [get]
func GetImportErrors(c echo.Context) error {
	client, err := airflow.GetClient()
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	logs, err := client.GetImportErrors()
	if err != nil {
		return common.ReturnErrorMsg(c, "Failed to get the taskInstances: "+err.Error())
	}

	return c.JSONPretty(http.StatusOK, logs, " ")
}

// ListWorkflowVersion godoc
//
//	@ID		list-workflowVersion
//	@Summary	List workflowVersion
//	@Description	Get a workflowVersion list.
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce	json
//	@Param		wfId path string true "wfId of the workflow"
//	@Param		page query string false "Page of the workflowVersion list."
//	@Param		row query string false "Row of the workflowVersion list."
//	@Success	200	{object}	[]model.WorkflowVersion	"Successfully get a workflowVersion list."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get a workflowVersion list."
//	@Router		/workflow/{wfId}/version [get]
func ListWorkflowVersion(c echo.Context) error {
	page, row, err := common.CheckPageRow(c)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	wfId := c.Param("wfId")
	if wfId == "" {
		return common.ReturnErrorMsg(c, "Please provide the wfId.")
	}

	workflow := &model.WorkflowVersion{
		WorkflowID: wfId,
	}

	workflows, err := dao.WorkflowVersionGetList(workflow, page, row)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, workflows, " ")
}

// GetWorkflowVersion godoc
//
//	@ID		get-WorkflowVersion
//	@Summary	Get WorkflowVersion
//	@Description	Get the WorkflowVersion.
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce	json
//	@Param		wfId path string true "wfId of the workflow"
//	@Param		verId path string true "ID of the WorkflowVersion."
//	@Success	200	{object}	model.Workflow		"Successfully get the WorkflowVersion."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get the WorkflowVersion."
//	@Router		/workflow/{wfId}/version/{verId} [get]
func GetWorkflowVersion(c echo.Context) error {
	wfId := c.Param("wfId")
	if wfId == "" {
		return common.ReturnErrorMsg(c, "Please provide the wfId.")
	}
	verId := common.UrlDecode(c.Param("verId"))
	if verId == "" {
		return common.ReturnErrorMsg(c, "Please provide the verId.")
	}

	workflow, err := dao.WorkflowVersionGet(verId, wfId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, workflow, " ")
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
	wfId := c.Param("wfId")
	if wfId == "" {
		return common.ReturnErrorMsg(c, "Please provide the wfId.")
	}
	wfRunId := c.Param("wfRunId")
	if wfRunId == "" {
		return common.ReturnErrorMsg(c, "Please provide the wfRunId.")
	}

	taskId := c.Param("taskId")
	if taskId == "" {
		return common.ReturnErrorMsg(c, "Please provide the taskId.")
	}
	taskInfo, err := dao.TaskGet(taskId)
	if err != nil {
		return common.ReturnErrorMsg(c, "Invalid get tasK from taskId.")
	}
	workflow, err := dao.WorkflowGet(wfId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	taskTryNum := c.Param("taskTryNum")
	if taskTryNum == "" {
		return common.ReturnErrorMsg(c, "Please provide the taskTryNum.")
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

// GetWorkflowStatus godoc
//
//	@ID		get-WorkflowStatus
//	@Summary	Get WorkflowStatus
//	@Description	Get the WorkflowStatus.
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce	json
//	@Param		wfId path string true "wfId of the workflow"
//	@Success	200	{object}	[]model.WorkflowStatus		"Successfully get the WorkflowVersion."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get the WorkflowVersion."
//	@Router		/workflow/{wfId}/status [get]
func GetWorkflowStatus(c echo.Context) error {
	wfId := c.Param("wfId")
	if wfId == "" {
		return common.ReturnErrorMsg(c, "Please provide the wfId.")
	}
	workflow, err := dao.WorkflowGet(wfId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	client, err := airflow.GetClient()
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	enumStatus := client.GetAllowedDagStateEnumValues()
	var statusList []model.WorkflowStatus
	for _, v := range enumStatus {

		resp, err := client.GetDagStatus(workflowDagID(workflow), string(*v.Ptr()))
		if err != nil {
			logger.Println(logger.ERROR, false,
				"AIRFLOW: Error occurred while getting DAGRuns. (Error: "+err.Error()+").")
		}
		statusList = append(statusList, model.WorkflowStatus{
			State: string(*v.Ptr()),
			Count: len(*resp.DagRuns),
		})
	}

	return c.JSONPretty(http.StatusOK, statusList, " ")
}
