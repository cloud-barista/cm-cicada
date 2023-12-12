package echo

import (
	"github.com/cloud-barista/cm-cicada/lib/airflow"
	"github.com/cloud-barista/cm-cicada/model"
	"github.com/labstack/echo/v4"
	"github.com/mitchellh/mapstructure"
	"net/http"
)

func CreateDAG(c echo.Context) error {
	var DAG model.DAG

	data, err := getJSONRawBody(c)
	if err != nil {
		return returnErrorMsg(c, err.Error())
	}

	err = mapstructure.Decode(data, &DAG)
	if err != nil {
		return returnErrorMsg(c, err.Error())
	}

	err = airflow.Conn.CreateDAG(&DAG)
	if err != nil {
		return returnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, DAG, " ")
}

func GetDAGs(c echo.Context) error {
	dags, err := airflow.Conn.GetDAGs()
	if err != nil {
		return returnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, dags, " ")
}

func RunDAG(c echo.Context) error {
	dagID := c.QueryParam("dag_id")
	if dagID == "" {
		return returnErrorMsg(c, "Please provide the dag_id parameter.")
	}

	dagRun, err := airflow.Conn.RunDAG(dagID)
	if err != nil {
		return returnErrorMsg(c, err.Error())
	}

	return c.JSONPretty(http.StatusOK, dagRun, " ")
}

func DAG() {
	e.POST("/dag/create", CreateDAG)
	e.GET("/dag/dags", GetDAGs)
	e.POST("/dag/run", RunDAG)
}
