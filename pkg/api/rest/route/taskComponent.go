package route

import (
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/controller"
	"github.com/labstack/echo/v4"
)

func TaskComponent(e *echo.Echo) {
	e.POST("/task_component", controller.CreateTaskComponent)
	e.GET("/task_component/:id", controller.GetTaskComponent)
	e.GET("/task_component", controller.ListTaskComponent)
	e.PUT("/task_component/:id", controller.UpdateTaskComponent)
	e.DELETE("/task_component/:id", controller.DeleteTaskComponent)
}
