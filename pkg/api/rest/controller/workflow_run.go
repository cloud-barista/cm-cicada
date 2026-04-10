package controller

import (
	"net/http"

	"github.com/cloud-barista/cm-cicada/dao"
	"github.com/cloud-barista/cm-cicada/lib/airflow"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/labstack/echo/v4"
)

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
	wfId, err := requireParam(c, "wfId", "id")
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
