package workflow

import (
	"net/http"

	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/util"
	"github.com/cloud-barista/cm-cicada/pkg/service"
	"github.com/labstack/echo/v4"
	"github.com/mitchellh/mapstructure"
)

var workflowService = service.NewWorkflowService()

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
			util.ToTimeHookFunc()),
		Result: &createWorkflowReq,
	})
	if err != nil {
		return err
	}

	err = decoder.Decode(data)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	workflow, err := workflowService.CreateWorkflow(createWorkflowReq)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

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
//	@Param		wfId path string true "ID of the workflow."
//	@Success	200	{object}	model.Workflow		"Successfully get the workflow."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get the workflow."
//	@Router		/workflow/{wfId} [get]
func GetWorkflow(c echo.Context) error {
	wfId := c.Param("wfId")

	workflow, err := workflowService.GetWorkflow(wfId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
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

	workflow, err := workflowService.GetWorkflowByName(wfName)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
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

	workflows, err := workflowService.ListWorkflow(workflow, page, row)
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
//	@Param		wfId path string true "ID of the workflow."
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
			util.ToTimeHookFunc()),
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
	workflow, err := workflowService.UpdateWorkflow(wfId, updateWorkflowReq)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, workflow, " ")
}

// DeleteWorkflow godoc
//
//	@ID		delete-workflow
//	@Summary	Delete Workflow
//	@Description	Delete the workflow.
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce	json
//	@Param		wfId path string true "ID of the workflow."
//	@Success	200	{object}	model.SimpleMsg		"Successfully delete the workflow"
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to delete the workflow"
//	@Router		/workflow/{wfId} [delete]
func DeleteWorkflow(c echo.Context) error {
	wfId := c.Param("wfId")

	err := workflowService.DeleteWorkflow(wfId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, model.SimpleMsg{Message: "success"}, " ")
}
