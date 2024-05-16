package route

import (
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/controller"
	"github.com/labstack/echo/v4"
)

func TaskTemplate(e *echo.Echo) {
	e.POST("/task_template", controller.CreateTaskTemplate)
	e.GET("/task_template/:id", controller.GetTaskTemplate)
	e.GET("/task_template", controller.GetTaskTemplate)
	e.PUT("/task_template/:id", controller.UpdateTaskTemplate)
	e.DELETE("/task_template/:id", controller.DeleteTaskTemplate)
}
