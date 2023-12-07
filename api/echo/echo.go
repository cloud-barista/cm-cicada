package echo

import (
	"github.com/cloud-barista/cm-cicada/lib/config"
	"github.com/jollaman999/utils/logger"
	"github.com/labstack/echo/v4"
)

var e *echo.Echo

func Init() {
	e = echo.New()

	DAG()

	err := e.Start(":" + config.CMCicadaConfig.CMCicada.Listen.Port)
	logger.Panicln(logger.ERROR, true, err)
}
