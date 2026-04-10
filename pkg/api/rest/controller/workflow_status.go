package controller

import (
	"net/http"

	"github.com/cloud-barista/cm-cicada/dao"
	"github.com/cloud-barista/cm-cicada/lib/airflow"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/jollaman999/utils/logger"
	"github.com/labstack/echo/v4"
)

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
