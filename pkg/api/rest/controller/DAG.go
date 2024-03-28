package controller

import (
	"github.com/cloud-barista/cm-cicada/lib/airflow"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/labstack/echo/v4"
	"github.com/mitchellh/mapstructure"
	"net/http"
)

// CreateDAG godoc
//
// @Summary		Create a DAG in Airflow
// @Description	Create a DAG.
// @Tags			[DAG] Create DAG
// @Accept			json
// @Produce		json
// @Param			request body model.DAG true "query params"
// @Success		200	{object}	model.DAG	"Successfully get DAGs."
// @Failure		400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse	"Failed to get DAGs."
// @Router			/dag/create [post]
func CreateDAG(c echo.Context) error {
	var DAG model.DAG

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
		return common.ReturnInternalError(c, err, "Failed to create DAG.")
	}

	return c.JSONPretty(http.StatusOK, DAG, " ")
}

// GetDAGs godoc
//
// @Summary		Get a list of DAGs from Airflow
// @Description	Get a list of DAGs.
// @Tags			[DAG] Get DAGs
// @Accept			json
// @Produce		json
// @Success		200	{object}	airflow.DAGCollection	"Successfully get DAGs."
// @Failure		400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse	"Failed to get DAGs."
// @Router			/dag/dags [get]
func GetDAGs(c echo.Context) error {
	dags, err := airflow.Client.GetDAGs()
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, dags, " ")
}

// RunDAG godoc
//
// @Summary		Run the DAG in Airflow
// @Description	Run the DAG.
// @Tags			[DAG] Run DAG
// @Accept			json
// @Produce		json
// @Param			dag_id query string true "DAG ID"
// @Success		200	{object}	model.DAG	"Successfully run the DAG."
// @Failure		400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure		500	{object}	common.ErrorResponse	"Failed to run DAG"
// @Router			/dag/run [post]
func RunDAG(c echo.Context) error {
	dagID := c.QueryParam("dag_id")
	if dagID == "" {
		return common.ReturnErrorMsg(c, "Please provide the dag_id parameter.")
	}

	dagRun, err := airflow.Client.RunDAG(dagID)
	if err != nil {
		return common.ReturnInternalError(c, err, "Failed to run DAG.")
	}

	return c.JSONPretty(http.StatusOK, dagRun, " ")
}
