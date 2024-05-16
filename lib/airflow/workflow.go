package airflow

import (
	"errors"
	"github.com/apache/airflow-client-go/airflow"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/jollaman999/utils/logger"
)

func (client *client) CreateDAG(DAG *model.Workflow) error {
	err := writeGustyYAMLs(DAG)
	if err != nil {
		return err
	}

	return nil
}

func (client *client) GetDAG(dagID string) (airflow.DAG, error) {
	ctx, cancel := Context()
	defer cancel()
	resp, _, err := client.api.DAGApi.GetDag(ctx, dagID).Execute()
	if err != nil {
		logger.Println(logger.ERROR, false,
			"AIRFLOW: Error occurred while getting DAGs. (Error: "+err.Error()+").")
	}

	return resp, err
}

func (client *client) GetDAGs() (airflow.DAGCollection, error) {
	ctx, cancel := Context()
	defer cancel()
	resp, _, err := client.api.DAGApi.GetDags(ctx).Execute()
	if err != nil {
		logger.Println(logger.ERROR, false,
			"AIRFLOW: Error occurred while getting DAGs. (Error: "+err.Error()+").")
	}

	return resp, err
}

func (client *client) RunDAG(dagID string) (airflow.DAGRun, error) {
	dags, err := client.GetDAGs()
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

	ctx, cancel := Context()
	defer cancel()
	resp, _, err := client.api.DAGRunApi.PostDagRun(ctx, dagID).DAGRun(*airflow.NewDAGRun()).Execute()
	if err != nil {
		errMsg := "AIRFLOW: Error occurred while running the DAG. (DAG ID: " + dagID + ", Error: " + err.Error() + ")"
		logger.Println(logger.ERROR, false, errMsg)

		return airflow.DAGRun{}, errors.New(errMsg)
	}

	logger.Println(logger.INFO, false, "AIRFLOW: Running the DAG. (DAG ID: "+dagID+")")

	return resp, err
}
