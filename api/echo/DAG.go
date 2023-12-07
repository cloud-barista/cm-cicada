package echo

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

func CreateDAG(c echo.Context) error {
	// TODO

	return c.JSONPretty(http.StatusOK, "TODO", " ")
}

func DAG() {
	e.GET("/dag/create", CreateDAG)
}
