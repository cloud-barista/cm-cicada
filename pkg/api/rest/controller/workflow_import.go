package controller

import (
	"net/http"

	"github.com/cloud-barista/cm-cicada/lib/airflow"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/labstack/echo/v4"
)

// GetImportErrors godoc
//
//	@ID			get-import-errors
//	@Summary	Get importErrors
//	@Description	Get the importErrors.
//	@Tags	[Admin]
//	@Accept	json
//	@Produce	json
//	@Success	200	{object}	airflow.ImportErrorCollection		"Successfully get the importErrors."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get the importErrors."
//	@Router	 /importErrors [get]
func GetImportErrors(c echo.Context) error {
	client, err := airflow.GetClient()
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	logs, err := client.GetImportErrors()
	if err != nil {
		return common.ReturnErrorMsg(c, "Failed to get the taskInstances: "+err.Error())
	}

	return c.JSONPretty(http.StatusOK, logs, " ")
}
