package route

import (
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/controller"
	"github.com/labstack/echo/v4"
)

func WorkflowTemplate(e *echo.Echo) {
	e.GET("/workflow_template/:uuid", controller.GetWorkflowTemplate)
	e.GET("/workflow_template", controller.ListWorkflowTemplate)
}
