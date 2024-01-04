package server

import (
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/middlewares"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/route"

	"github.com/cloud-barista/cm-cicada/lib/config"
	_ "github.com/cloud-barista/cm-cicada/pkg/api/rest/docs" // Cicada Documentation
	"github.com/jollaman999/utils/logger"
	"github.com/labstack/echo/v4"
)

func Init() {
	e := echo.New()

	e.Use(middlewares.CustomLogger())

	route.DAG(e)
	route.RegisterSwagger(e)
	route.RegisterUtility(e)

	err := e.Start(":" + config.CMCicadaConfig.CMCicada.Listen.Port)
	logger.Panicln(logger.ERROR, true, err)
}
