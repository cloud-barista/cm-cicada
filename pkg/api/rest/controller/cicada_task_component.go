package controller

import (
	"net/http"

	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/service"
	"github.com/labstack/echo/v4"
)

// RunScript godoc
//
//	@ID				run-script
//	@Summary		Run Script on Target
//	@Description	Run script on target with NS ID, MCI ID and VM ID.
//	@Tags			[Cicada Built-in API]
//	@Accept			json
//	@Produce		json
//	@Param			request body 	model.RunScriptReq true "Workflow content"
//	@Success		200	{object}	model.ScriptResult		"Result of the script running"
//	@Failure		400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure		500	{object}	common.ErrorResponse	"Failed to run script"
//	@Router			/run_script [post]
func RunScript(c echo.Context) error {
	req := new(model.RunScriptReq)
	if err := c.Bind(req); err != nil {
		return err
	}

	svc := service.NewCicadaBuiltinService()
	result, err := svc.RunScript(*req)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, result, " ")
}

// SleepTime godoc
//
//	@ID				sleep-time
//	@Summary		Run Sleep Command
//	@Description	Runs sleep command on cicada and waits for configured time. Wait for 10 seconds if time value is not provided.
//	@Tags			[Cicada Built-in API]
//	@Accept			json
//	@Produce		json
//	@Param			request body 	model.SleepTimeReq true "SleepTime request"
//	@Success		200	{object}	model.ScriptResult		"Result of sleep"
//	@Failure		400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure		500	{object}	common.ErrorResponse	"Failed to run script"
//	@Router			/sleep_time [post]
func SleepTime(c echo.Context) error {
	req := new(model.SleepTimeReq)
	if err := c.Bind(req); err != nil {
		return err
	}

	svc := service.NewCicadaBuiltinService()
	result, err := svc.SleepTime(*req)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, result, " ")
}
