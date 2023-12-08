package airflow

import (
	"context"
	"errors"
	"github.com/apache/airflow-client-go/airflow"
	"github.com/jollaman999/utils/logger"
)

func (conn *Connection) GetDAGs() (airflow.DAGCollection, error) {
	ctx, cancel := context.WithTimeout(conn.ctx, conn.timeout)
	defer cancel()
	resp, _, err := conn.cli.DAGApi.GetDags(ctx).Execute()
	if err != nil {
		logger.Println(logger.ERROR, false,
			"AIRFLOW: Error occurred while getting DAGs. (Error: "+err.Error()+").")
	}

	return resp, err
}

func (conn *Connection) RunDAG(dagID string) (airflow.DAGRun, error) {
	dags, err := conn.GetDAGs()
	if err != nil {
		return airflow.DAGRun{}, err
	}

	var found = false

	for _, dag := range *dags.Dags {
		if *dag.DagId == dagID {
			found = true
			break
		}
	}

	if !found {
		logger.Println(logger.DEBUG, false,
			"AIRFLOW: Received the request with none existing DAG ID. (DAG ID: "+dagID+")")
		return airflow.DAGRun{}, errors.New("provided dag_id is not exist")
	}

	ctx, cancel := context.WithTimeout(conn.ctx, conn.timeout)
	defer cancel()
	resp, _, err := conn.cli.DAGRunApi.PostDagRun(ctx, dagID).DAGRun(*airflow.NewDAGRun()).Execute()
	if err != nil {
		logger.Println(logger.ERROR, false,
			"AIRFLOW: Error occurred while running the DAG. (DAG ID: "+dagID+", Error: "+err.Error()+")")
	}

	logger.Println(logger.INFO, false, "AIRFLOW: Running the DAG. (DAG ID: "+dagID+")")

	return resp, err
}
