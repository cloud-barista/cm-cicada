package route

import (
	"github.com/cloud-barista/cm-cicada/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/controller"
	"github.com/labstack/echo/v4"
	"strings"
)

func Workflow(e *echo.Echo) {
	e.POST("/"+strings.ToLower(common.ShortModuleName)+"/workflow", controller.CreateWorkflow)
	e.GET("/"+strings.ToLower(common.ShortModuleName)+"/workflow/:wfId", controller.GetWorkflow)
	e.GET("/"+strings.ToLower(common.ShortModuleName)+"/workflow", controller.ListWorkflow)
	e.PUT("/"+strings.ToLower(common.ShortModuleName)+"/workflow/:wfId", controller.UpdateWorkflow)
	e.POST("/"+strings.ToLower(common.ShortModuleName)+"/workflow/:wfId/run", controller.RunWorkflow)
	e.DELETE("/"+strings.ToLower(common.ShortModuleName)+"/workflow/:wfId", controller.DeleteWorkflow)
}
