package route

import (
	"github.com/cloud-barista/cm-cicada/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/controller"
	"github.com/labstack/echo/v4"
	"strings"
)

func TaskComponent(e *echo.Echo) {
	e.POST("/"+strings.ToLower(common.ShortModuleName)+"/task_component", controller.CreateTaskComponent)
	e.GET("/"+strings.ToLower(common.ShortModuleName)+"/task_component/:tcId", controller.GetTaskComponent)
	e.GET("/"+strings.ToLower(common.ShortModuleName)+"/task_component", controller.ListTaskComponent)
	e.PUT("/"+strings.ToLower(common.ShortModuleName)+"/task_component/:tcId", controller.UpdateTaskComponent)
	e.DELETE("/"+strings.ToLower(common.ShortModuleName)+"/task_component/:tcId", controller.DeleteTaskComponent)
}
