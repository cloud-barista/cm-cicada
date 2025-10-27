package controller

import (
	"net/http"

	"github.com/cloud-barista/cm-cicada/lib/ssh"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/labstack/echo/v4"
)

// RunScript godoc
//
//	@ID				run-script
//	@Summary		Run script on target
//	@Description	Run script on target with NS ID, MCI ID and VM ID.
//	@Tags			[Cicada Task Component]
//	@Accept			json
//	@Produce		json
//	@Param			request body 	model.RunScriptReq true "Workflow content"
//	@Success		200	{object}	model.ScriptResult		"Result of the script running"
//	@Failure		400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure		500	{object}	common.ErrorResponse	"Failed to run script"
//	@Router			/run_script [post]
func RunScript(c echo.Context) error {
	runScriptReq := new(model.RunScriptReq)
	err := c.Bind(runScriptReq)
	if err != nil {
		return err
	}

	if runScriptReq.NSID == "" {
		return common.ReturnErrorMsg(c, "Please provide the ns_id.")
	}

	if runScriptReq.MCIID == "" {
		return common.ReturnErrorMsg(c, "Please provide the mci_id.")
	}

	if runScriptReq.VMID == "" {
		return common.ReturnErrorMsg(c, "Please provide the vm_id.")
	}

	if runScriptReq.Content == "" {
		return common.ReturnErrorMsg(c, "Please provide the content.")
	}

	var result model.ScriptResult

	output, err := ssh.ExecuteScript(runScriptReq.NSID, runScriptReq.MCIID, runScriptReq.VMID, runScriptReq.Content)
	if err != nil {
		result.IsSuccess = false
		result.Error = err.Error()
	} else {
		result.IsSuccess = true
	}
	result.Output = string(output)

	return c.JSONPretty(http.StatusOK, result, " ")
}
