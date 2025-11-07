package airflow

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"

	"github.com/apache/airflow-client-go/airflow"
	"github.com/cloud-barista/cm-cicada/lib/config"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/jollaman999/utils/fileutil"
	"github.com/jollaman999/utils/logger"
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
		_, exist := dagRequests[workflowID]
		if !exist {
			return
		}

		dagRequests[workflowID].Unlock()

		dagRequestsLock.Lock()
		delete(dagRequests, workflowID)
		dagRequestsLock.Unlock()
	}
}

func (client *Client) CreateDAG(workflow *model.Workflow) error {
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

func (client *Client) GetDAG(dagID string) (airflow.DAG, error) {
	deferFunc := callDagRequestLock(dagID)
	defer func() {
		deferFunc()
	}()

	ctx, cancel := Context()
	defer cancel()
	resp, _, err := client.DAGApi.GetDag(ctx, dagID).Execute()
	if err != nil {
		logger.Println(logger.ERROR, false,
			"AIRFLOW: Error occurred while getting the DAG. (Error: "+err.Error()+").")
	}

	return resp, err
}

func (client *Client) GetDAGs() (airflow.DAGCollection, error) {
	ctx, cancel := Context()
	defer cancel()
	resp, _, err := client.DAGApi.GetDags(ctx).Execute()
	if err != nil {
		logger.Println(logger.ERROR, false,
			"AIRFLOW: Error occurred while getting DAGs. (Error: "+err.Error()+").")
	}
	return resp, err
}

func (client *Client) RunDAG(dagID string) (airflow.DAGRun, error) {
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
	resp, _, err := client.DAGRunApi.PostDagRun(ctx, dagID).DAGRun(*airflow.NewDAGRun()).Execute()
	if err != nil {
		errMsg := "AIRFLOW: Error occurred while running the DAG. (DAG ID: " + dagID + ", Error: " + err.Error() + ")"
		logger.Println(logger.ERROR, false, errMsg)

		return airflow.DAGRun{}, errors.New(errMsg)
	}

	logger.Println(logger.INFO, false, "AIRFLOW: Running the DAG. (DAG ID: "+dagID+")")

	return resp, err
}

func (client *Client) DeleteDAG(dagID string, deleteFolderOnly bool) error {
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
		_, err = client.DAGApi.DeleteDag(ctx, dagID).Execute()
		if err != nil {
			logger.Println(logger.ERROR, false,
				"AIRFLOW: Error occurred while deleting the DAG. (Error: "+err.Error()+").")
		}
	}

	return err
}
func (client *Client) GetDAGRuns(dagID string) (airflow.DAGRunCollection, error) {
	deferFunc := callDagRequestLock(dagID)
	defer func() {
		deferFunc()
	}()
	ctx, cancel := Context()
	defer cancel()
	resp, _, err := client.DAGRunApi.GetDagRuns(ctx, dagID).Execute()
	if err != nil {
		logger.Println(logger.ERROR, false,
			"AIRFLOW: Error occurred while getting DAGRuns. (Error: "+err.Error()+").")
	}
	return resp, err
}

func (client *Client) GetTaskInstances(dagID string, dagRunId string) (airflow.TaskInstanceCollection, error) {
	deferFunc := callDagRequestLock(dagID)
	defer func() {
		deferFunc()
	}()
	ctx, cancel := Context()
	defer cancel()
	resp, _, err := client.TaskInstanceApi.GetTaskInstances(ctx, dagID, dagRunId).Execute()
	if err != nil {
		logger.Println(logger.ERROR, false,
			"AIRFLOW: Error occurred while getting TaskInstances. (Error: "+err.Error()+").")
	}
	return resp, err
}

func (client *Client) GetTaskLogs(dagID, dagRunID, taskID string, taskTryNumber int) (airflow.InlineResponse200, error) {
	deferFunc := callDagRequestLock(dagID)
	defer func() {
		deferFunc()
	}()
	ctx, cancel := Context()
	defer cancel()

	// TaskInstanceApi 인스턴스를 사용하여 로그 요청
	logs, _, err := client.TaskInstanceApi.GetLog(ctx, dagID, dagRunID, taskID, int32(taskTryNumber)).FullContent(true).Execute()
	logger.Println(logger.INFO, false, logs)
	if err != nil {
		logger.Println(logger.ERROR, false,
			"AIRFLOW: Error occurred while getting TaskInstance logs. (Error: "+err.Error()+").")
	}

	return logs, nil
}
func (client *Client) ClearTaskInstance(dagID string, dagRunID string, option model.TaskClearOption) (airflow.TaskInstanceReferenceCollection, error) {
	deferFunc := callDagRequestLock(dagID)
	defer func() {
		deferFunc()
	}()
	ctx, cancel := Context()
	defer cancel()
	defaultOption := false
	clearTask := airflow.ClearTaskInstances{
		DryRun:            &option.DryRun,
		TaskIds:           &option.TaskIds,
		IncludeSubdags:    &defaultOption,
		IncludeParentdag:  &defaultOption,
		IncludeUpstream:   &option.IncludeUpstream,
		IncludeDownstream: &option.IncludeDownstream,
		IncludeFuture:     &defaultOption,
		IncludePast:       &defaultOption,
		OnlyFailed:        &option.OnlyFailed,
		OnlyRunning:       &option.OnlyRunning,
		ResetDagRuns:      &option.ResetDagRuns,
		DagRunId:          *airflow.NewNullableString(&dagRunID),
	}

	logger.Println(logger.DEBUG, false,
		"ClearTaskInstances 요청 내용 : {} ", clearTask.TaskIds)

	// 요청 생성
	request := client.DAGApi.PostClearTaskInstances(ctx, dagID).ClearTaskInstances(clearTask)

	// 요청 실행
	response, _, err := client.DAGApi.PostClearTaskInstancesExecute(request)
	if err != nil {
		logger.Println(logger.ERROR, false, "AIRFLOW: Error occurred while clearing TaskInstance.")

		return airflow.TaskInstanceReferenceCollection{}, err
	}
	logger.Println(logger.WARN, false, "response : ", response.GetTaskInstances())

	if response.TaskInstances == nil || len(*response.TaskInstances) == 0 {
		logger.Println(logger.WARN, false, "AIRFLOW: 요청은 성공했지만 반환된 TaskInstances가 없습니다.")
	}

	return response, nil
}

func (client *Client) GetEventLogs(dagID string, dagRunId string, taskId string) ([]byte, error) {
	deferFunc := callDagRequestLock(dagID)
	defer func() {
		deferFunc()
	}()
	ctx, cancel := Context()
	defer cancel()

	localBasePath, err := client.GetConfig().ServerURLWithContext(ctx, "EventLogApiService.GetEventLog")
	if err != nil {
		fmt.Println("Error occurred while getting event logs:", err)
	}

	baseURL := "http://" + client.GetConfig().Host + localBasePath + "/eventLogs"
	queryParams := map[string]string{
		"offset":          "0",
		"limit":           "100",
		"dag_id":          dagID,
		"run_id":          dagRunId,
		"task_id":         taskId,
		"order_by":        "-when",
		"excluded_events": "gantt,landing_times,tries,duration,calendar,graph,grid,tree,tree_data",
	}
	query := url.Values{}
	for key, value := range queryParams {
		query.Add(key, value)
	}
	queryString := query.Encode()
	fullURL := fmt.Sprintf("%s?%s", baseURL, queryString)
	httpclient := client.GetConfig().HTTPClient

	// 요청 생성
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		fmt.Println("Error occurred while creating the request:", err)
	}
	cred := ctx.Value(airflow.ContextBasicAuth).(airflow.BasicAuth)
	addBasicAuth(req, cred.UserName, cred.Password)
	res, err := httpclient.Do(req)
	if err != nil {
		fmt.Println("Error occurred while sending the request:", err)
	}
	defer func() {
		_ = res.Body.Close()
	}()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
	}

	return body, err
}

func (client *Client) GetImportErrors() (airflow.ImportErrorCollection, error) {
	ctx, cancel := Context()
	defer cancel()

	logs, _, err := client.ImportErrorApi.GetImportErrors(ctx).Execute()
	logger.Println(logger.INFO, false, logs)
	if err != nil {
		logger.Println(logger.ERROR, false,
			"AIRFLOW: Error occurred while getting import dag errors. (Error: "+err.Error()+").")
	}

	return logs, nil
}
func (client *Client) GetDagStatus(dagID string, status string) (airflow.DAGRunCollection, error) {
	deferFunc := callDagRequestLock(dagID)
	defer func() {
		deferFunc()
	}()
	ctx, cancel := Context()
	defer cancel()
	resp, _, err := client.DAGRunApi.GetDagRuns(ctx, dagID).State([]string{status}).Execute()
	if err != nil {
		logger.Println(logger.ERROR, false,
			"AIRFLOW: Error occurred while getting DAGRuns. (Error: "+err.Error()+").")
	}
	// enumStatus := airflow.AllowedDagStateEnumValues
	// var statusList []model.WorkflowStatus
	// for _, v := range enumStatus {

	// 	resp, _, err := client.DAGRunApi.GetDagRuns(ctx, dagID).State([]string{v}).Execute()
	// 	if err != nil {
	// 		logger.Println(logger.ERROR, false,
	// 			"AIRFLOW: Error occurred while getting DAGRuns. (Error: "+err.Error()+").")
	// 	}
	// 	statusList = append(statusList, model.WorkflowStatus{
	// 		State: string(*v.Ptr()),
	// 		Count: len(*resp.DagRuns),
	// 	})
	// }

	return resp, nil
}

func (client *Client) GetAllowedDagStateEnumValues() []airflow.DagState {
	return airflow.AllowedDagStateEnumValues
}

func (client *Client) PatchDag(dagID string, dagBody airflow.DAG) (airflow.DAG, error) {
	ctx, cancel := Context()
	defer cancel()

	// TaskInstanceApi 인스턴스를 사용하여 로그 요청
	logs, _, err := client.DAGApi.PatchDag(ctx, dagID).DAG(dagBody).Execute()
	logger.Println(logger.INFO, false, logs)
	if err != nil {
		logger.Println(logger.ERROR, false,
			"AIRFLOW: Error occurred while patch dag errors. (Error: "+err.Error()+").")
	}

	return logs, nil
}

func (client *Client) GetXComValue(dagID, dagRunID, taskID, xcomKey string) (map[string]interface{}, error) {
	deferFunc := callDagRequestLock(dagID)
	defer func() {
		deferFunc()
	}()
	ctx, cancel := Context()
	defer cancel()

	// XCom API를 사용하여 특정 task의 xcom 데이터 조회
	xcomEntry, _, err := client.XComApi.GetXcomEntry(ctx, dagID, dagRunID, taskID, xcomKey).Execute()
	if err != nil {
		logger.Println(logger.ERROR, false,
			"AIRFLOW: Error occurred while getting XCom value. (Error: "+err.Error()+").")
		return nil, err
	}

	// xcom value를 JSON에서 map으로 파싱
	if xcomEntry.Value != nil {
		var valueMap map[string]interface{}
		err := json.Unmarshal([]byte(*xcomEntry.Value), &valueMap)
		if err != nil {
			logger.Println(logger.ERROR, false,
				"AIRFLOW: Error occurred while parsing XCom value. (Error: "+err.Error()+").")
			return nil, err
		}
		return valueMap, nil
	}

	return nil, nil
}

func addBasicAuth(req *http.Request, username, password string) {
	auth := username + ":" + password
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Add("Authorization", "Basic "+encodedAuth)
}
