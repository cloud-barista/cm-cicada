package controller

import (
	"github.com/cloud-barista/cm-cicada/lib/airflow"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/labstack/echo/v4"
	"github.com/mitchellh/mapstructure"
	"net/http"
)

type GetDAGResponse struct {
	model.DAG
}

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

	err = airflow.Conn.CreateDAG(&DAG)
	if err != nil {
		return common.ReturnInternalError(c, err, "Failed to create DAG.")
	}

	return c.JSONPretty(http.StatusOK, DAG, " ")
}

// GetDAGs godoc
//
//	@Summary		Get a list of DAG in Workflow Engine
//	@Description	Get information of DAG.
//	@Tags			[Sample] Get DAG
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	GetDAGResponse	"Successfully get DAGs"
//	@Failure		404	{object}	GetDAGResponse	"Failed to get DAGs"
//	@Router			/dag/dags [get]
func GetDAGs(c echo.Context) error {
	dags, err := airflow.Conn.GetDAGs()
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, dags, " ")
}

func RunDAG(c echo.Context) error {
	dagID := c.QueryParam("dag_id")
	if dagID == "" {
		return common.ReturnErrorMsg(c, "Please provide the dag_id parameter.")
	}

	dagRun, err := airflow.Conn.RunDAG(dagID)
	if err != nil {
		return common.ReturnInternalError(c, err, "Failed to run DAG.")
	}

	return c.JSONPretty(http.StatusOK, dagRun, " ")
}
