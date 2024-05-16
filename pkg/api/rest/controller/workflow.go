package controller

import (
	"github.com/cloud-barista/cm-cicada/lib/airflow"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/labstack/echo/v4"
	"github.com/mitchellh/mapstructure"
	"net/http"
)

// CreateWorkflow godoc
//
// @Summary		Create Workflow
// @Description	Create a DAG in Airflow.
// @Tags		[Workflow]
// @Accept		json
// @Produce		json
// @Param		request body model.Workflow true "query params"
// @Success		200	{object}	model.Workflow			"Successfully create the DAG."
// @Failure		400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse	"Failed to create DAG."
// @Router		/workflow [post]
func CreateWorkflow(c echo.Context) error {
	var DAG model.Workflow

	data, err := common.GetJSONRawBody(c)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	err = mapstructure.Decode(data, &DAG)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	err = airflow.Client.CreateDAG(&DAG)
	if err != nil {
		return common.ReturnInternalError(c, err, "Failed to create workflow.")
	}

	return c.JSONPretty(http.StatusOK, DAG, " ")
}

// GetWorkflow godoc
//
// @Summary		List Workflow
// @Description	Get a list of DAGs from Airflow
// @Tags		[Workflow]
// @Accept		json
// @Produce		json
// @Success		200	{object}	airflow.DAGCollection	"Successfully get a workflow list."
// @Failure		400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse	"Failed to get a workflow list."
// @Router		/workflow/{id} [get]
func GetWorkflow(c echo.Context) error {
	dagID := c.Param("id")
	if dagID == "" {
		return common.ReturnErrorMsg(c, "Please provide the id.")
	}

	dag, err := airflow.Client.GetDAG(dagID)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, dag, " ")
}

// ListWorkflow godoc
//
// @Summary		List Workflow
// @Description	Get a list of DAGs from Airflow
// @Tags		[Workflow]
// @Accept		json
// @Produce		json
// @Success		200	{object}	airflow.DAGCollection	"Successfully get a workflow list."
// @Failure		400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse	"Failed to get a workflow list."
// @Router		/workflow [get]
func ListWorkflow(c echo.Context) error {
	dags, err := airflow.Client.GetDAGs()
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, dags, " ")
}

// RunWorkflow godoc
//
// @Summary		Run Workflow
// @Description	Get the DAG in Airflow
// @Tags		[Workflow]
// @Accept		json
// @Produce		json
// @Param		dag_id query string true "Workflow ID"
// @Success		200	{object}	model.Workflow	"Successfully run the Workflow."
// @Failure		400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse	"Failed to run Workflow"
// @Router		/workflow/run/{id} [post]
func RunWorkflow(c echo.Context) error {
	dagID := c.Param("id")
	if dagID == "" {
		return common.ReturnErrorMsg(c, "Please provide the id.")
	}

	dagRun, err := airflow.Client.RunDAG(dagID)
	if err != nil {
		return common.ReturnInternalError(c, err, "Failed to run workflow.")
	}

	return c.JSONPretty(http.StatusOK, dagRun, " ")
}
