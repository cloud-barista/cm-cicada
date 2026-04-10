package controller

import (
	"net/http"

	"github.com/cloud-barista/cm-cicada/dao"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/labstack/echo/v4"
)

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

	wfId, err := requireParam(c, "wfId", "wfId")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
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
	wfId, err := requireParam(c, "wfId", "wfId")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
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
