package controller

import (
	"github.com/cloud-barista/cm-cicada/dao"
	"github.com/cloud-barista/cm-cicada/lib/airflow"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/labstack/echo/v4"
	"github.com/mitchellh/mapstructure"
	"net/http"
	"reflect"
	"time"
)

func toTimeHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if t != reflect.TypeOf(time.Time{}) {
			return data, nil
		}

		switch f.Kind() {
		case reflect.String:
			return time.Parse(time.RFC3339, data.(string))
		case reflect.Float64:
			return time.Unix(0, int64(data.(float64))*int64(time.Millisecond)), nil
		case reflect.Int64:
			return time.Unix(0, data.(int64)*int64(time.Millisecond)), nil
		default:
			return data, nil
		}
		// Convert it by parsing
	}
}

// CreateWorkflow godoc
//
// @Summary		Create Workflow
// @Description	Create a workflow.
// @Tags		[Workflow]
// @Accept		json
// @Produce		json
// @Param		request body model.Workflow true "Workflow content"
// @Success		200	{object}	model.WorkflowTemplate	"Successfully create the workflow."
// @Failure		400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse	"Failed to create DAG."
// @Router		/workflow [post]
func CreateWorkflow(c echo.Context) error {
	var workflow model.Workflow

	data, err := common.GetJSONRawBody(c)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Metadata: nil,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			toTimeHookFunc()),
		Result: &workflow,
	})
	if err != nil {
		return err
	}

	err = decoder.Decode(data)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	err = airflow.Client.CreateDAG(&workflow)
	if err != nil {
		return common.ReturnErrorMsg(c, "Failed to create the workflow.")
	}

	_, err = dao.WorkflowCreate(&workflow)
	if err != nil {
		{
			return common.ReturnErrorMsg(c, err.Error())
		}
	}

	return c.JSONPretty(http.StatusOK, workflow, " ")
}

// GetWorkflow godoc
//
// @Summary		Get Workflow
// @Description	Get the workflow.
// @Tags		[Workflow]
// @Accept		json
// @Produce		json
// @Param		id path string true "ID of the workflow."
// @Success		200	{object}	model.Workflow			"Successfully get the workflow."
// @Failure		400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse	"Failed to get the workflow."
// @Router		/workflow/{id} [get]
func GetWorkflow(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return common.ReturnErrorMsg(c, "Please provide the id.")
	}

	workflow, err := dao.WorkflowGet(id)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	_, err = airflow.Client.GetDAG(id)
	if err != nil {
		return common.ReturnErrorMsg(c, "Failed to get the workflow from the airflow server.")
	}

	return c.JSONPretty(http.StatusOK, workflow, " ")
}

// ListWorkflow godoc
//
// @Summary		List Workflow
// @Description	Get a workflow list.
// @Tags		[Workflow]
// @Accept		json
// @Produce		json
// @Param		page query string false "Page of the connection information list."
// @Param		row query string false "Row of the connection information list."
// @Success		200	{object}	[]model.Workflow		"Successfully get a workflow list."
// @Failure		400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse	"Failed to get a workflow list."
// @Router		/workflow [get]
func ListWorkflow(c echo.Context) error {
	page, row, err := common.CheckPageRow(c)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	workflows, err := dao.WorkflowGetList(page, row)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, workflows, " ")
}

// RunWorkflow godoc
//
// @Summary		Run Workflow
// @Description	Run the workflow.
// @Tags		[Workflow]
// @Accept		json
// @Produce		json
// @Param		id path string true "ID of the workflow."
// @Success		200	{object}	model.Workflow			"Successfully run the workflow."
// @Failure		400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse	"Failed to run the Workflow"
// @Router		/workflow/run/{id} [post]
func RunWorkflow(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return common.ReturnErrorMsg(c, "Please provide the id.")
	}

	workflow, err := dao.WorkflowGet(id)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	_, err = airflow.Client.RunDAG(id)
	if err != nil {
		return common.ReturnInternalError(c, err, "Failed to run the workflow.")
	}

	return c.JSONPretty(http.StatusOK, workflow, " ")
}

// UpdateWorkflow godoc
//
// @Summary		Update Workflow
// @Description	Update the workflow content.
// @Tags		[Workflow]
// @Accept		json
// @Produce		json
// @Param		id path string true "ID of the workflow."
// @Param		Workflow body model.Workflow true "Workflow to modify."
// @Success		200	{object}	model.Workflow	"Successfully update the workflow"
// @Failure		400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse	"Failed to update the workflow"
// @Router		/workflow/{id} [put]
func UpdateWorkflow(c echo.Context) error {
	Workflow := new(model.Workflow)
	err := c.Bind(Workflow)
	if err != nil {
		return err
	}

	Workflow.ID = c.Param("id")
	_, err = dao.WorkflowGet(Workflow.ID)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	err = dao.WorkflowUpdate(Workflow)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, Workflow, " ")
}

// DeleteWorkflow godoc
//
// @Summary		Delete Workflow
// @Description	Delete the workflow.
// @Tags		[Workflow]
// @Accept		json
// @Produce		json
// @Param		id path string true "ID of the workflow."
// @Success		200	{object}	model.Workflow	"Successfully delete the workflow"
// @Failure		400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse	"Failed to delete the workflow"
// @Router		/workflow/{id} [delete]
func DeleteWorkflow(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return common.ReturnErrorMsg(c, "Please provide the id.")
	}

	workflow, err := dao.WorkflowGet(id)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	err = airflow.Client.DeleteDAG(id)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	err = dao.WorkflowDelete(workflow)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, workflow, " ")
}
