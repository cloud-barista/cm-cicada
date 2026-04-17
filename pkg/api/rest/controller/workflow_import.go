package controller

import (
	"net/http"

	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/service"
	"github.com/labstack/echo/v4"
)

// GetImportErrors godoc
//
//	@ID			get-import-errors
//	@Summary	List Airflow Import Errors
//	@Description	List DAG import errors reported by Airflow.
//	@Tags	[Admin]
//	@Accept	json
//	@Produce	json
//	@Success	200	{object}	airflow.ImportErrorCollection		"Successfully get the importErrors."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get the importErrors."
//	@Router	 /importErrors [get]
func GetImportErrors(c echo.Context) error {
	svc := service.NewWorkflowRuntimeService()
	result, err := svc.GetImportErrors()
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, result, " ")
}
