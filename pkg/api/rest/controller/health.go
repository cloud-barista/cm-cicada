package controller

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

type SimpleMsg struct {
	Message string `json:"message" example:"Any message"`
}

// GetHealth func is for checking Cicada server health.
// RestGetHealth godoc
// @Summary Check Cicada is alive
// @Description Check Cicada is alive
// @Tags [Admin] System management
// @Accept  json
// @Produce  json
// @Success 200 {object} SimpleMsg
// @Failure 404 {object} SimpleMsg
// @Failure 500 {object} SimpleMsg
// @Router /cicada/health [get]
func GetHealth(c echo.Context) error {
	okMessage := SimpleMsg{}
	okMessage.Message = "CM-Cicada API server is running"
	return c.JSONPretty(http.StatusOK, &okMessage, " ")
}
