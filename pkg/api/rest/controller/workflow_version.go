package controller

import (
	"net/http"
	"strconv"

	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/service"
	"github.com/labstack/echo/v4"
)

var _ model.WorkflowVersion // swag type reference

// ListWorkflowVersion godoc
//
//	@ID		list-workflowVersion
//	@Summary	List Workflow Versions
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

	svc := service.NewWorkflowService()
	versions, err := svc.ListWorkflowVersions(wfId, page, row)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, versions, " ")
}

// GetWorkflowVersion godoc
//
//	@ID		get-WorkflowVersion
//	@Summary	Get Workflow Version
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

	svc := service.NewWorkflowService()
	version, err := svc.GetWorkflowVersion(wfId, verId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, version, " ")
}

// RollbackWorkflow godoc
//
//	@ID		rollback-workflow
//	@Summary	Rollback Workflow to a Past Version
//	@Description	Restore the workflow's task graph and metadata from the snapshot identified by versionNo. Existing active tasks are soft-deleted and re-issued with fresh UUIDs from the target version's raw_data. The rollback itself is recorded as a new WorkflowVersion with action="rollback" and source_template_id=<source version id>. workflow_schedules rows are left untouched.
//	@Tags		[Workflow]
//	@Produce	json
//	@Param		wfId path 	string true "Workflow ID."
//	@Param		versionNo path 	int true "Target version_no within the workflow (1-based, positive integer)."
//	@Success	200	{object}	model.Workflow		"Workflow after rollback."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	404	{object}	common.ErrorResponse	"Workflow or version not found."
//	@Failure	409	{object}	common.ErrorResponse	"Cannot rollback to this version (e.g. delete action)."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to rollback the workflow."
//	@Router		/workflow/{wfId}/version/{versionNo}/rollback [post]
func RollbackWorkflow(c echo.Context) error {
	wfId, err := requireParam(c, "wfId", "wfId")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	versionNoStr := c.Param("versionNo")
	versionNo, err := strconv.Atoi(versionNoStr)
	if err != nil || versionNo <= 0 {
		return common.ReturnErrorMsg(c, "version_no must be a positive integer")
	}

	svc := service.NewWorkflowService()
	workflow, err := svc.RollbackWorkflow(wfId, versionNo)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, workflow, " ")
}
