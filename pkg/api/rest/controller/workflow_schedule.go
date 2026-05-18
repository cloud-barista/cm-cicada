package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mitchellh/mapstructure"

	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/service"
)

// ScheduleWorkflow godoc
//
//	@ID		schedule-workflow
//	@Summary	Schedule Workflow (one-shot)
//	@Description	Register a one-shot future execution. Airflow handles the actual triggering via DAG metadata (schedule="@once" + start_date=run_at + catchup=false). Only one active schedule per workflow is allowed.
//	@Tags		[Workflow]
//	@Accept		json
//	@Produce	json
//	@Param		wfId path 	string true "Workflow ID."
//	@Param		request body 	model.CreateWorkflowScheduleReq true "Schedule body (run_at)."
//	@Success	200	{object}	model.WorkflowSchedule	"Successfully scheduled the workflow."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	409	{object}	common.ErrorResponse	"Active schedule already exists."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to schedule the workflow."
//	@Router		/workflow/{wfId}/schedule [post]
func ScheduleWorkflow(c echo.Context) error {
	wfID := c.Param("wfId")
	if wfID == "" {
		return common.ReturnErrorMsg(c, "please provide the workflow id")
	}

	var req model.CreateWorkflowScheduleReq

	data, err := common.GetJSONRawBody(c)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Metadata: nil,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			toTimeHookFunc()),
		Result: &req,
	})
	if err != nil {
		return err
	}
	if err := decoder.Decode(data); err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	svc := service.NewWorkflowScheduleService()
	row, err := svc.Schedule(wfID, req)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	return c.JSONPretty(http.StatusOK, row, " ")
}

// GetWorkflowSchedule godoc
//
//	@ID		get-workflow-schedule
//	@Summary	Get Workflow Schedule
//	@Description	Return the workflow's most recently created schedule row regardless of status. Inspect .status to interpret: "active" (will fire), "executed" (already ran), "canceled" (user canceled). Returns null when the workflow has no schedule history.
//	@Tags		[Workflow]
//	@Produce	json
//	@Param		wfId path 	string true "Workflow ID."
//	@Success	200	{object}	model.WorkflowSchedule	"Latest schedule (any status), or null."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to fetch the schedule."
//	@Router		/workflow/{wfId}/schedule [get]
func GetWorkflowSchedule(c echo.Context) error {
	wfID := c.Param("wfId")
	if wfID == "" {
		return common.ReturnErrorMsg(c, "please provide the workflow id")
	}
	svc := service.NewWorkflowScheduleService()
	row, err := svc.GetLatest(wfID)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	return c.JSONPretty(http.StatusOK, row, " ")
}

// CancelWorkflowSchedule godoc
//
//	@ID		cancel-workflow-schedule
//	@Summary	Cancel Workflow Schedule
//	@Description	Cancel the workflow's active schedule. The DAG metadata is rewritten so Airflow no longer triggers the workflow on the originally requested cadence. Returns 404 when no active schedule exists.
//	@Tags		[Workflow]
//	@Produce	json
//	@Param		wfId path 	string true "Workflow ID."
//	@Success	200	{object}	model.WorkflowSchedule	"Successfully canceled."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	404	{object}	common.ErrorResponse	"No active schedule for this workflow."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to cancel the schedule."
//	@Router		/workflow/{wfId}/schedule [delete]
func CancelWorkflowSchedule(c echo.Context) error {
	wfID := c.Param("wfId")
	if wfID == "" {
		return common.ReturnErrorMsg(c, "please provide the workflow id")
	}
	svc := service.NewWorkflowScheduleService()
	row, err := svc.Cancel(wfID)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	return c.JSONPretty(http.StatusOK, row, " ")
}
