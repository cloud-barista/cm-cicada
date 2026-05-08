package controller

import (
	"net/http"

	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/service"
	"github.com/labstack/echo/v4"
)

var _ model.WorkflowStatus // swag type reference

// GetWorkflowStatus godoc
//
//	@ID		get-WorkflowStatus
//	@Summary	Get Workflow Status
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

	svc := service.NewWorkflowRuntimeService()
	statusList, err := svc.GetWorkflowStatus(wfId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, statusList, " ")
}
