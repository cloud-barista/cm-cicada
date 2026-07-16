package airflow

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/cloud-barista/cm-cicada/lib/airflow/client"
	"github.com/cloud-barista/cm-cicada/lib/config"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/jollaman999/utils/fileutil"
	"github.com/jollaman999/utils/logger"
)

// dagLock serializes requests for a single DAG. waiters counts the callers
// holding or queued on it, so the map entry only goes away once the last one
// is done.
type dagLock struct {
	mu      sync.Mutex
	waiters int
}

var dagRequests = make(map[string]*dagLock)
var dagRequestsLock = sync.Mutex{}

// callDagRequestLock serializes requests for one DAG and returns the function
// that releases it. Different DAGs never block each other.
//
// The entry is reference counted rather than deleted on unlock: dropping it
// while another caller still held or awaited it handed two callers separate
// mutexes for the same DAG, which defeated the exclusion entirely. Every access
// to the map is made under dagRequestsLock, since reading it while another
// goroutine deletes is a data race and can crash the process.
func callDagRequestLock(workflowID string) func() {
	dagRequestsLock.Lock()
	entry, exist := dagRequests[workflowID]
	if !exist {
		entry = &dagLock{}
		dagRequests[workflowID] = entry
	}
	entry.waiters++
	dagRequestsLock.Unlock()

	entry.mu.Lock()

	// The release is idempotent so a double call cannot unlock a mutex this
	// caller no longer owns, matching the previous guard.
	var once sync.Once
	return func() {
		once.Do(func() {
			entry.mu.Unlock()

			dagRequestsLock.Lock()
			entry.waiters--
			if entry.waiters == 0 {
				delete(dagRequests, workflowID)
			}
			dagRequestsLock.Unlock()
		})
	}
}

func (c *Client) CreateDAG(workflow *model.Workflow) error {
	dagID := workflow.ID
	if workflow.WorkflowKey != "" {
		dagID = workflow.WorkflowKey
	}
	deferFunc := callDagRequestLock(dagID)
	defer func() {
		deferFunc()
	}()

	err := writeGustyYAMLs(workflow)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetDAG(dagID string) (client.DAG, error) {
	deferFunc := callDagRequestLock(dagID)
	defer func() {
		deferFunc()
	}()

	ctx, cancel := Context()
	defer cancel()

	dag, err := c.api.GetDAG(ctx, dagID)
	if err != nil {
		logger.Println(logger.ERROR, false,
			"AIRFLOW: Error occurred while getting the DAG. (Error: "+err.Error()+").")
	}

	return dag, err
}

func (c *Client) RunDAG(dagID string, conf map[string]interface{}) (client.DAGRun, error) {
	deferFunc := callDagRequestLock(dagID)
	defer func() {
		deferFunc()
	}()

	ctx, cancel := Context()
	defer cancel()

	// Ask about this DAG directly instead of listing every DAG and scanning:
	// the list endpoint is capped by [api] fallback_page_limit, so a scan
	// silently misses DAGs once there are more than one page of them.
	exists, err := c.api.DAGExists(ctx, dagID)
	if err != nil {
		errMsg := "AIRFLOW: Error occurred while looking up the DAG. (DAG ID: " + dagID + ", Error: " + err.Error() + ")"
		logger.Println(logger.ERROR, false, errMsg)

		return client.DAGRun{}, errors.New(errMsg)
	}
	if !exists {
		logger.Println(logger.DEBUG, false,
			"AIRFLOW: Received the request with none existing DAG ID. (DAG ID: "+dagID+")")
		return client.DAGRun{}, errors.New("provided dag_id is not exist")
	}

	dagRun, err := c.api.TriggerDAGRun(ctx, dagID, conf)
	if err != nil {
		errMsg := "AIRFLOW: Error occurred while running the DAG. (DAG ID: " + dagID + ", Error: " + err.Error() + ")"
		logger.Println(logger.ERROR, false, errMsg)

		return client.DAGRun{}, errors.New(errMsg)
	}

	logger.Println(logger.INFO, false, "AIRFLOW: Running the DAG. (DAG ID: "+dagID+")")

	return dagRun, nil
}

func (c *Client) DeleteDAG(dagID string, deleteFolderOnly bool) error {
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

	if !deleteFolderOnly {
		ctx, cancel := Context()
		defer cancel()

		err = c.api.DeleteDAG(ctx, dagID)
		if err != nil {
			logger.Println(logger.ERROR, false,
				"AIRFLOW: Error occurred while deleting the DAG. (Error: "+err.Error()+").")
		}
	}

	return err
}

func (c *Client) GetDAGRuns(dagID string) (client.DAGRunCollection, error) {
	deferFunc := callDagRequestLock(dagID)
	defer func() {
		deferFunc()
	}()

	ctx, cancel := Context()
	defer cancel()

	runs, err := c.api.GetDAGRuns(ctx, dagID, nil)
	if err != nil {
		logger.Println(logger.ERROR, false,
			"AIRFLOW: Error occurred while getting DAGRuns. (Error: "+err.Error()+").")
	}

	return runs, err
}

func (c *Client) GetTaskInstances(dagID string, dagRunID string) (client.TaskInstanceCollection, error) {
	deferFunc := callDagRequestLock(dagID)
	defer func() {
		deferFunc()
	}()

	ctx, cancel := Context()
	defer cancel()

	instances, err := c.api.GetTaskInstances(ctx, dagID, dagRunID)
	if err != nil {
		logger.Println(logger.ERROR, false,
			"AIRFLOW: Error occurred while getting TaskInstances. (Error: "+err.Error()+").")
	}

	return instances, err
}

// GetTaskLogs returns the log of one task attempt as text.
func (c *Client) GetTaskLogs(dagID, dagRunID, taskID string, taskTryNumber int) (string, error) {
	deferFunc := callDagRequestLock(dagID)
	defer func() {
		deferFunc()
	}()

	ctx, cancel := Context()
	defer cancel()

	logs, err := c.api.GetTaskLog(ctx, dagID, dagRunID, taskID, taskTryNumber)
	if err != nil {
		logger.Println(logger.ERROR, false,
			"AIRFLOW: Error occurred while getting TaskInstance logs. (Error: "+err.Error()+").")
		return "", err
	}

	return logs, nil
}

func (c *Client) ClearTaskInstance(dagID string, dagRunID string, option model.TaskClearOption) (client.TaskInstanceCollection, error) {
	deferFunc := callDagRequestLock(dagID)
	defer func() {
		deferFunc()
	}()

	ctx, cancel := Context()
	defer cancel()

	// include_subdags / include_parentdag are gone in Airflow 3 along with SubDags.
	clearTask := client.ClearTaskInstancesRequest{
		DryRun:            option.DryRun,
		TaskIDs:           option.TaskIds,
		DagRunID:          &dagRunID,
		IncludeUpstream:   option.IncludeUpstream,
		IncludeDownstream: option.IncludeDownstream,
		IncludeFuture:     false,
		IncludePast:       false,
		OnlyFailed:        option.OnlyFailed,
		OnlyRunning:       option.OnlyRunning,
		ResetDagRuns:      option.ResetDagRuns,
	}

	response, err := c.api.ClearTaskInstances(ctx, dagID, clearTask)
	if err != nil {
		logger.Println(logger.ERROR, false,
			"AIRFLOW: Error occurred while clearing TaskInstance. (Error: "+err.Error()+").")

		return client.TaskInstanceCollection{}, err
	}

	if len(response.TaskInstances) == 0 {
		logger.Println(logger.WARN, false, "AIRFLOW: 요청은 성공했지만 반환된 TaskInstances가 없습니다.")
	}

	return response, nil
}

func (c *Client) GetEventLogs(dagID string, dagRunID string, taskID string) (client.EventLogCollection, error) {
	deferFunc := callDagRequestLock(dagID)
	defer func() {
		deferFunc()
	}()

	ctx, cancel := Context()
	defer cancel()

	logs, err := c.api.GetEventLogs(ctx, dagID, dagRunID, taskID, 100, 0)
	if err != nil {
		logger.Println(logger.ERROR, false,
			"AIRFLOW: Error occurred while getting event logs. (Error: "+err.Error()+").")
		return client.EventLogCollection{}, err
	}

	return logs, nil
}

func (c *Client) GetImportErrors() (client.ImportErrorCollection, error) {
	ctx, cancel := Context()
	defer cancel()

	importErrors, err := c.api.GetImportErrors(ctx)
	if err != nil {
		logger.Println(logger.ERROR, false,
			"AIRFLOW: Error occurred while getting import dag errors. (Error: "+err.Error()+").")
		return client.ImportErrorCollection{}, err
	}

	return importErrors, nil
}

func (c *Client) GetDagStatus(dagID string, status string) (client.DAGRunCollection, error) {
	deferFunc := callDagRequestLock(dagID)
	defer func() {
		deferFunc()
	}()

	ctx, cancel := Context()
	defer cancel()

	runs, err := c.api.GetDAGRuns(ctx, dagID, []string{status})
	if err != nil {
		logger.Println(logger.ERROR, false,
			"AIRFLOW: Error occurred while getting DAGRuns. (Error: "+err.Error()+").")
		return client.DAGRunCollection{}, err
	}

	return runs, nil
}

// GetAllowedDagStateEnumValues returns the states a DAG run can report.
func (c *Client) GetAllowedDagStateEnumValues() []string {
	return client.DagRunStates
}

// GetXComValue reads an XCom value and decodes it as a JSON object.
//
// Task code pushes these as JSON strings, so a string value is unmarshalled;
// Airflow 3 can also hand back an already-decoded object.
func (c *Client) GetXComValue(dagID, dagRunID, taskID, xcomKey string) (map[string]interface{}, error) {
	deferFunc := callDagRequestLock(dagID)
	defer func() {
		deferFunc()
	}()

	ctx, cancel := Context()
	defer cancel()

	entry, err := c.api.GetXComEntry(ctx, dagID, dagRunID, taskID, xcomKey)
	if err != nil {
		logger.Println(logger.ERROR, false,
			"AIRFLOW: Error occurred while getting XCom value. (Error: "+err.Error()+").")
		return nil, err
	}

	switch value := entry.Value.(type) {
	case nil:
		return nil, nil
	case string:
		var valueMap map[string]interface{}
		if err := json.Unmarshal([]byte(value), &valueMap); err != nil {
			logger.Println(logger.ERROR, false,
				"AIRFLOW: Error occurred while parsing XCom value. (Error: "+err.Error()+").")
			return nil, err
		}
		return valueMap, nil
	case map[string]interface{}:
		return value, nil
	default:
		return nil, errors.New("unexpected XCom value type")
	}
}
