package airflow

import (
	"errors"
	"github.com/apache/airflow-client-go/airflow"
	"github.com/cloud-barista/cm-cicada/lib/config"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/jollaman999/utils/fileutil"
	"github.com/jollaman999/utils/logger"
	"sync"
)

var dagRequests = make(map[string]*sync.Mutex)
var dagRequestsLock = sync.Mutex{}

func callDagRequestLock(workflowID string) func() {
	dagRequestsLock.Lock()
	_, exist := dagRequests[workflowID]
	if !exist {
		dagRequests[workflowID] = new(sync.Mutex)
	}
	dagRequestsLock.Unlock()

	dagRequests[workflowID].Lock()

	return func() {
		dagRequests[workflowID].Unlock()

		dagRequestsLock.Lock()
		delete(dagRequests, workflowID)
		dagRequestsLock.Unlock()
	}
}

func (client *client) CreateDAG(workflow *model.Workflow) error {
	deferFunc := callDagRequestLock(workflow.ID)
	defer func() {
		deferFunc()
	}()

	err := writeGustyYAMLs(workflow)
	if err != nil {
		return err
	}

	return nil
}

func (client *client) GetDAG(dagID string) (airflow.DAG, error) {
	deferFunc := callDagRequestLock(dagID)
	defer func() {
		deferFunc()
	}()

	ctx, cancel := Context()
	defer cancel()
	resp, _, err := client.api.DAGApi.GetDag(ctx, dagID).Execute()
	if err != nil {
		logger.Println(logger.ERROR, false,
			"AIRFLOW: Error occurred while getting the DAG. (Error: "+err.Error()+").")
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
	deferFunc := callDagRequestLock(dagID)
	defer func() {
		deferFunc()
	}()

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

func (client *client) DeleteDAG(dagID string) error {
	deferFunc := callDagRequestLock(dagID)
	defer func() {
		deferFunc()
	}()

	dagDir := config.CMCicadaConfig.CMCicada.DAGDirectoryHost + "/" + dagID
	err := fileutil.DeleteDir(dagDir)
	if err != nil {
		logger.Println(logger.ERROR, true,
			"AIRFLOW: Failed to delete dag directory. (Error: "+err.Error()+").")
	}

	ctx, cancel := Context()
	defer cancel()
	_, err = client.api.DAGApi.DeleteDag(ctx, dagID).Execute()
	if err != nil {
		logger.Println(logger.ERROR, false,
			"AIRFLOW: Error occurred while deleting the DAG. (Error: "+err.Error()+").")
	}

	return err
}
