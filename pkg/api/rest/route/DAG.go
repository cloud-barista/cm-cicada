package route

import (
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/controller"
	"github.com/labstack/echo/v4"
)

func DAG(e *echo.Echo) {
	e.POST("/dag/create", controller.CreateDAG)
	e.GET("/dag/dags", controller.GetDAGs)
	e.POST("/dag/run", controller.RunDAG)
}
