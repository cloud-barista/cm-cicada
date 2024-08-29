package controller

import (
	"github.com/cloud-barista/cm-cicada/dao"
	"github.com/cloud-barista/cm-cicada/db"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	_ "github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
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
// @Param		wftId path string true "ID of the WorkflowTemplate"
// @Success		200	{object}	model.GetWorkflowTemplate	"Successfully get the workflow template"
// @Failure		400	{object}	common.ErrorResponse		"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse		"Failed to get the workflow template"
// @Router		/workflow_template/{wftId} [get]
func GetWorkflowTemplate(c echo.Context) error {
	wftId := c.Param("wftId")
	if wftId == "" {
		return common.ReturnErrorMsg(c, "wftId is empty")
	}
	workflowTemplate, err := dao.WorkflowTemplateGet(wftId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	return c.JSONPretty(http.StatusOK, workflowTemplate, "")
}

// GetWorkflowTemplateByName godoc
//
// @Summary		Get WorkflowTemplate by Name
// @Description	Get the workflow template by name.
// @Tags		[Workflow Template]
// @Accept		json
// @Produce		json
// @Param		wfName path string true "Name of the WorkflowTemplate"
// @Success		200	{object}	model.GetWorkflowTemplate	"Successfully get the workflow template"
// @Failure		400	{object}	common.ErrorResponse		"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse		"Failed to get the workflow template"
// @Router		/workflow_template/name/{wfName} [get]
func GetWorkflowTemplateByName(c echo.Context) error {
	wfName := c.Param("wfName")
	if wfName == "" {
		return common.ReturnErrorMsg(c, "wfName is empty")
	}
	workflowTemplate := db.WorkflowTemplateGetByName(wfName)
	if workflowTemplate == nil {
		return common.ReturnErrorMsg(c, "workflow template not found with the provided name")
	}
	return c.JSONPretty(http.StatusOK, model.GetWorkflowTemplate{
		Name: workflowTemplate.Name,
		Data: workflowTemplate.Data,
	}, "")
}

// ListWorkflowTemplate godoc
//
// @Summary		List WorkflowTemplate
// @Description	Get a list of workflow template.
// @Tags		[Workflow Template]
// @Accept		json
// @Produce		json
// @Param		name query string false "Name of the workflow template"
// @Param		page query string false "Page of the workflow template list."
// @Param		row query string false "Row of the workflow template list."
// @Success		200	{object}	[]model.WorkflowTemplate		"Successfully get a list of workflow template."
// @Failure		400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse	"Failed to get a list of workflow template."
// @Router		/workflow_template [get]
func ListWorkflowTemplate(c echo.Context) error {
	page, row, err := common.CheckPageRow(c)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	workflowTemplate := &model.WorkflowTemplate{
		Name: c.QueryParam("name"),
	}

	workflowTemplateList, err := dao.WorkflowTemplateGetList(workflowTemplate, page, row)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	return c.JSONPretty(http.StatusOK, workflowTemplateList, "")
}
