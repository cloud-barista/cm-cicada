package controller

import (
	"net/http"
	"time"

	"github.com/cloud-barista/cm-cicada/dao"
	"github.com/cloud-barista/cm-cicada/lib/airflow"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/mapper"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/google/uuid"
	"github.com/jollaman999/utils/logger"
	"github.com/labstack/echo/v4"
	"github.com/mitchellh/mapstructure"
)

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

	workflowData, err := mapper.CreateDataReqToData(specVersion, createWorkflowReq.Data)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	if err := airflow.ValidateWorkflow(&model.Workflow{
		Name:        createWorkflowReq.Name,
		SpecVersion: specVersion,
		Data:        workflowData,
	}); err != nil {
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

	sourceType, sourceTemplateID, err := mapper.ResolveCreateSourceType(specVersion, createWorkflowReq.Data)
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
	wfName, err := requireParam(c, "wfName", "wfName")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	includeDeleted, err := queryBool(c, "include_deleted")
	if err != nil {
		return common.ReturnErrorMsg(c, "Invalid include_deleted value.")
	}

	var workflowByName *model.Workflow
	if includeDeleted {
		workflowByName, err = dao.WorkflowGetByNameIncludeDeleted(wfName)
	} else {
		workflowByName, err = dao.WorkflowGetByName(wfName)
	}
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	var workflow *model.Workflow
	if includeDeleted {
		workflow, err = mapper.GetWorkflowFromDBIncludeDeleted(workflowByName.ID)
	} else {
		workflow, err = mapper.GetWorkflowFromDB(workflowByName.ID)
	}
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

	includeDeleted, err := queryBool(c, "include_deleted")
	if err != nil {
		return common.ReturnErrorMsg(c, "Invalid include_deleted value.")
	}

	workflow := &model.Workflow{
		Name: c.QueryParam("name"),
	}

	var workflows *[]model.Workflow
	if includeDeleted {
		workflows, err = dao.WorkflowGetListIncludeDeleted(workflow, page, row)
	} else {
		workflows, err = dao.WorkflowGetList(workflow, page, row)
	}
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, workflows, " ")
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

	wfId, err := requireParam(c, "wfId", "wfId")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

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

	workflowData, err := mapper.CreateDataReqToData(specVersion, updateWorkflowReq.Data)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	validateTarget := *oldWorkflow
	validateTarget.SpecVersion = specVersion
	validateTarget.Data = workflowData
	if err := airflow.ValidateWorkflow(&validateTarget); err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	diff, err := mapper.BuildWorkflowGraphDiff(oldWorkflow, workflowData)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	for _, tg := range diff.TaskGroupsToUpsert {
		taskGroup := tg
		if err := dao.TaskGroupSave(&taskGroup); err != nil {
			return common.ReturnErrorMsg(c, err.Error())
		}
	}
	for _, t := range diff.TasksToUpsert {
		task := t
		if err := dao.TaskSave(&task); err != nil {
			return common.ReturnErrorMsg(c, err.Error())
		}
	}
	for _, t := range diff.TasksToSoftDrop {
		task := t
		if err := dao.TaskDelete(&task); err != nil {
			return common.ReturnErrorMsg(c, err.Error())
		}
	}
	for _, tg := range diff.TaskGroupsToSoftDrop {
		taskGroup := tg
		if err := dao.TaskGroupDelete(&taskGroup); err != nil {
			return common.ReturnErrorMsg(c, err.Error())
		}
	}

	oldWorkflow.SpecVersion = specVersion
	oldWorkflow.Data = diff.WorkflowData

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
	wfId, err := requireParam(c, "wfId", "wfId")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
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
