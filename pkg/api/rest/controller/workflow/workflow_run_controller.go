package workflow

import (
	"net/http"

	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/cloud-barista/cm-cicada/pkg/service"
	"github.com/labstack/echo/v4"
)

var (
	workflowRunService = service.NewWorkflowService()
	eventService       = service.NewEventService()
)

// RunWorkflow godoc
//
//	@ID		run-workflow
//	@Summary	Run Workflow
//	@Description	Run the workflow.
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce	json
//	@Param		wfId path string true "ID of the workflow."
//	@Success	200	{object}	model.SimpleMsg		"Successfully run the workflow."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to run the Workflow"
//	@Router		/workflow/{wfId}/run [post]
func RunWorkflow(c echo.Context) error {
	wfId := c.Param("wfId")

	err := workflowRunService.RunWorkflow(wfId)
	if err != nil {
		return common.ReturnInternalError(c, err, "Failed to run the workflow.")
	}

	return c.JSONPretty(http.StatusOK, model.SimpleMsg{Message: "success"}, " ")
}

// GetWorkflowRuns godoc
//
//	@ID			get-workflow-runs
//	@Summary	Get workflowRuns
//	@Description	Get the task Logs.
//	@Tags	[Workflow]
//	@Accept	json
//	@Produce	json
//	@Param	wfId path string true "ID of the workflow."
//	@Success	200	{object}	[]model.WorkflowRun		"Successfully get the workflowRuns."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get the workflowRuns."
//	@Router	 /workflow/{wfId}/runs [get]
func GetWorkflowRuns(c echo.Context) error {
	wfId := c.Param("wfId")

	transformedRuns, err := workflowRunService.GetWorkflowRuns(wfId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, transformedRuns, " ")
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

	statusList, err := workflowRunService.GetWorkflowStatus(wfId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, statusList, " ")
}

// GetEventLogs godoc
//
//	@ID				get-event-logs
//	@Summary		Get Eventlog
//	@Description	Get Eventlog.
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce	json
//	@Param		wfId path string true "ID of the workflow."
//	@Param		wfRunId query string false "ID of the workflow run."
//	@Param		taskId query string false "ID of the task."
//	@Success	200	{object}	[]model.EventLog			"Successfully get the workflow."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get the workflow."
//	@Router	/workflow/{wfId}/eventlogs [get]
func GetEventLogs(c echo.Context) error {
	wfId := c.Param("wfId")
	wfRunId := c.QueryParam("wfRunId")
	taskId := c.QueryParam("taskId")

	logList, err := eventService.GetEventLogs(wfId, wfRunId, taskId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
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
	logs, err := eventService.GetImportErrors()
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, logs, " ")
}
