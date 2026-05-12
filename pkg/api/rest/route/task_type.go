package route

import (
	"strings"

	"github.com/cloud-barista/cm-cicada/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/controller"
	"github.com/labstack/echo/v4"
)

func TaskType(e *echo.Echo) {
	base := "/" + strings.ToLower(common.ShortModuleName) + "/task_type"
	e.GET(base, controller.ListTaskType)
	e.GET(base+"/:id", controller.GetTaskType)
}
