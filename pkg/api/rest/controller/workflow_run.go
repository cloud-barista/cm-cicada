package controller

import (
	"net/http"

	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/service"
	"github.com/labstack/echo/v4"
)

// RunWorkflow godoc
//
//	@ID		run-workflow
//	@Summary	Run Workflow
//	@Description	Run the workflow.
//	@Tags		[Workflow Execution]
//	@Accept		json
//	@Produce	json
//	@Param		wfId path string true "DB workflow ID."
//	@Success	200	{object}	model.SimpleMsg		"Successfully run the workflow."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to run the Workflow"
//	@Router		/workflow/{wfId}/run [post]
func RunWorkflow(c echo.Context) error {
	wfId, err := requireParam(c, "wfId", "id")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	svc := service.NewWorkflowService()
	if err := svc.RunWorkflow(wfId); err != nil {
		if common.IsNotFoundError(err) {
			return common.ReturnErrorMsg(c, err.Error())
		}
		return common.ReturnInternalError(c, err, "Failed to run the workflow.")
	}

	return c.JSONPretty(http.StatusOK, model.SimpleMsg{Message: "success"}, " ")
}

// GetWorkflowRuns godoc
//
//	@ID			get-workflow-runs
//	@Summary	List Workflow Runs
//	@Description	Get the task Logs.
//	@Tags	[Workflow Execution]
//	@Accept	json
//	@Produce	json
//	@Param	wfId path string true "DB workflow ID."
//	@Success	200	{object}	[]model.WorkflowRun		"Successfully get the workflowRuns."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get the workflowRuns."
//	@Router	 /workflow/{wfId}/runs [get]
func GetWorkflowRuns(c echo.Context) error {
	wfId, err := requireParam(c, "wfId", "wfId")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	svc := service.NewWorkflowRuntimeService()
	runs, err := svc.GetWorkflowRuns(wfId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, runs, " ")
}
