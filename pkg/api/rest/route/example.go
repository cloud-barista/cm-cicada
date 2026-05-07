package route

import (
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/cloud-barista/cm-cicada/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/controller"
)

// Example registers the built-in example endpoints used by the http_xcom
// workflow demo. They are intentionally trivial — a fixed-payload GET and a
// body-echoing POST — so the demo does not depend on any external module.
func Example(e *echo.Echo) {
	prefix := "/" + strings.ToLower(common.ShortModuleName) + "/example"
	e.GET(prefix+"/data", controller.ExampleSampleData)
	e.POST(prefix+"/echo", controller.ExampleEcho)
}
