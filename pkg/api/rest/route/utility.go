package route

import (
	"github.com/cloud-barista/cm-cicada/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/controller"
	"strings"

	"github.com/labstack/echo/v4"
)

func RegisterUtility(e *echo.Echo) {
	e.GET("/"+strings.ToLower(common.ShortModuleName)+"/health", controller.GetHealth)
}
