package route

import (
	"github.com/cloud-barista/cm-cicada/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/controller"
	"github.com/labstack/echo/v4"
	"strings"
)

func WorkflowTemplate(e *echo.Echo) {
	e.GET("/"+strings.ToLower(common.ShortModuleName)+"/workflow_template/:wftId", controller.GetWorkflowTemplate)
	e.GET("/"+strings.ToLower(common.ShortModuleName)+"/workflow_template", controller.ListWorkflowTemplate)
}
