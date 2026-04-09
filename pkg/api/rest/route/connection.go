package route

import (
	"strings"

	"github.com/cloud-barista/cm-cicada/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/controller"
	"github.com/labstack/echo/v4"
)

func Connection(e *echo.Echo) {
	base := "/" + strings.ToLower(common.ShortModuleName) + "/connection"
	e.POST(base, controller.CreateConnection)
	e.GET(base, controller.ListConnection)
	e.GET(base+"/:connId", controller.GetConnection)
	e.PUT(base+"/:connId", controller.UpdateConnection)
	e.DELETE(base+"/:connId", controller.DeleteConnection)
}
