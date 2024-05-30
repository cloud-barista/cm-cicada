package route

import (
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/controller"
	"github.com/labstack/echo/v4"
)

func Workflow(e *echo.Echo) {
	e.POST("/workflow", controller.CreateWorkflow)
	e.GET("/workflow/:uuid", controller.GetWorkflow)
	e.GET("/workflow", controller.ListWorkflow)
	e.PUT("/workflow/:uuid", controller.UpdateWorkflow)
	e.POST("/workflow/run/:uuid", controller.RunWorkflow)
	e.DELETE("/workflow/:uuid", controller.DeleteWorkflow)
}
