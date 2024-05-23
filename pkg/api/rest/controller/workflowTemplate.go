package controller

import (
	// "errors"
	"github.com/cloud-barista/cm-cicada/dao"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	_ "github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	// "github.com/jollaman999/utils/logger"
	"github.com/labstack/echo/v4"
	"net/http"
)

// GetWorkflowTemplate godoc
//
// @Summary		Get WorkflowTemplate
// @Description	Get the workflow template.
// @Tags		[Workflow Template]
// @Accept		json
// @Produce		json
// @Param		id path string true "id of the WorkflowTemplate"
// @Success		200	{object}	model.WorkflowTemplate			"Successfully get the workflow template"
// @Failure		400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse	"Failed to get the workflow template"
// @Router		/workflow_template/{id} [get]
func GetWorkflowTemplate(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return common.ReturnErrorMsg(c, "id is empty")
	}
	workflowTemplate, err := dao.WorkflowTemplateGet(id)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	return c.JSONPretty(http.StatusOK, workflowTemplate, "")
}

// ListWorkflowTemplate godoc
//
// @Summary		List WorkflowTemplate
// @Description	Get a list of workflow template.
// @Tags		[Workflow Template]
// @Accept		json
// @Produce		json
// @Param		page query string false "Page of the workflow template list."
// @Param		row query string false "Row of the workflow template list."
// @Param		uuid query string false "UUID of the workflow template."
// @Param		name query string false "Migration group name."
// @Success		200	{object}	[]model.WorkflowTemplate		"Successfully get a list of workflow template."
// @Failure		400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse	"Failed to get a list of workflow template."
// @Router			/workflow_template [get]
func ListWorkflowTemplate(c echo.Context) error {
	page, row, err := common.CheckPageRow(c)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	workflowTemplateList, err := dao.WorkflowTemplateGetList(page, row)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	return c.JSONPretty(http.StatusOK, workflowTemplateList, "")
}
