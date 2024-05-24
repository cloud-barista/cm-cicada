package controller

import (
	_ "github.com/cloud-barista/cm-cicada/pkg/api/rest/common" // Need for swag
	"github.com/labstack/echo/v4"
	"net/http"
)

type SimpleMsg struct {
	Message string `json:"message"`
}

var OkMessage = SimpleMsg{}
var IsReady = false

// CheckReady func is for checking Cicada server health.
// @Summary Check Ready
// @Description Check Cicada is ready
// @Tags [Admin] System management
// @Accept		json
// @Produce		json
// @Success		200 {object}	SimpleMsg				"Successfully get ready state."
// @Failure		500	{object}	common.ErrorResponse	"Failed to check ready state."
//
// @Router /cicada/readyz [get]
func CheckReady(c echo.Context) error {
	status := http.StatusOK

	if !IsReady {
		status = http.StatusServiceUnavailable
	}

	return c.JSONPretty(status, &OkMessage, " ")
}
