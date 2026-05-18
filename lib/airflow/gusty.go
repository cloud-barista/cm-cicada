package airflow

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/cloud-barista/cm-cicada/common"
	"github.com/cloud-barista/cm-cicada/dao"
	"github.com/cloud-barista/cm-cicada/lib/airflow/catalog"
	"github.com/cloud-barista/cm-cicada/lib/config"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/jollaman999/utils/fileutil"
	"github.com/jollaman999/utils/logger"
	"gopkg.in/yaml.v3"
)

func ValidateWorkflow(workflow *model.Workflow) error {
	if workflow == nil {
		return errors.New("workflow is nil")
	}
	return checkWorkflow(workflow)
}

func checkWorkflow(workflow *model.Workflow) error {
	if err := validateTaskDependencyGraph(workflow.Data.TaskGroups); err != nil {
		return err
	}

	for _, tg := range workflow.Data.TaskGroups {
		for _, t := range tg.Tasks {
			taskComponent := dao.TaskComponentGetByName(t.TaskComponent)
			if taskComponent == nil {
				return errors.New("task component '" + t.TaskComponent + "' not found")
			}
		}
	}

	return nil
}

func validateTaskDependencyGraph(taskGroups []model.TaskGroup) error {
	taskNames := make(map[string]struct{})
	taskPathByName := make(map[string]string)
	dependenciesByTask := make(map[string][]string)

	for _, tg := range taskGroups {
		if tg.Name == "" {
			return errors.New("task group name should not be empty")
		}

		for _, t := range tg.Tasks {
			if t.Name == "" {
				return errors.New("task name should not be empty")
			}
			if _, exists := taskNames[t.Name]; exists {
				return errors.New("Duplicated task name: " + t.Name)
			}

			taskNames[t.Name] = struct{}{}
			taskPathByName[t.Name] = tg.Name + "." + t.Name
			dependenciesByTask[t.Name] = append([]string{}, t.Dependencies...)
		}
	}

	for taskName, deps := range dependenciesByTask {
		for _, dep := range deps {
			if taskName == dep {
				return errors.New("cycle dependency found in " + taskPathByName[taskName])
			}

			if _, exists := taskNames[dep]; !exists {
				return errors.New("wrong dependency found in " + taskPathByName[taskName] + " (" + dep + ")")
			}
		}
	}

	const (
		unvisited = 0
		visiting  = 1
		visited   = 2
	)

	visitState := make(map[string]int)
	var dfs func(taskName string) error
	dfs = func(taskName string) error {
		visitState[taskName] = visiting

		for _, dep := range dependenciesByTask[taskName] {
			switch visitState[dep] {
			case visiting:
				return errors.New("cycle dependency found in " + taskPathByName[taskName])
			case unvisited:
				if err := dfs(dep); err != nil {
					return err
				}
			}
		}

		visitState[taskName] = visited
		return nil
	}

	for taskName := range dependenciesByTask {
		if visitState[taskName] == unvisited {
			if err := dfs(taskName); err != nil {
				return err
			}
		}
	}

	return nil
}

func isTaskExist(workflow *model.Workflow, taskID string) bool {
	for _, tg := range workflow.Data.TaskGroups {
		for _, t := range tg.Tasks {
			if t.Name == taskID {
				return true
			}
		}
	}

	return false
}

func parseEndpoint(pathParams map[string]string, queryParams map[string]string, endpoint string) (string, error) {
	pathParamKeys := reflect.ValueOf(pathParams).MapKeys()
	for _, key := range pathParamKeys {
		if pathParams[key.String()] == "" {
			return endpoint, fmt.Errorf("path parameter %s is empty", pathParams[key.String()])
		}
		endpoint = strings.ReplaceAll(endpoint, "{"+key.String()+"}", pathParams[key.String()])
	}

	queryParamKeys := reflect.ValueOf(queryParams).MapKeys()
	if len(queryParamKeys) > 0 {
		var queryParamsString string

		for _, key := range queryParamKeys {
			if queryParams[key.String()] == "" {
				continue
			}
			queryParamsString += fmt.Sprintf("%v=%v&", key.String(), queryParams[key.String()])
		}

		if queryParamsString != "" {
			queryParamsString = strings.TrimRight(queryParamsString, "&")

			if !strings.HasSuffix(endpoint, "?") {
				endpoint += "?"
			}
			endpoint += queryParamsString
		}
	}

	return endpoint, nil
}

func writeModelToYAMLFile(model any, filePath string) error {
	bytes, err := yaml.Marshal(model)
	if err != nil {
		return err
	}
	parsed := string(bytes)

	return fileutil.WriteFile(filePath, parsed)
}

func writeGustyYAMLs(workflow *model.Workflow) error {
	// 워크플로우를 Gusty용 DAG YAML로 변환하고, 기존 파일의 잔여물을 정리한다.
	if err := checkWorkflow(workflow); err != nil {
		return err
	}

	dagDir, err := ensureWorkflowDir(workflow)
	if err != nil {
		return err
	}

	if err := writeDAGMetadata(workflow, dagDir); err != nil {
		return err
	}

	taskAirflowIDByName := buildTaskAirflowIDByName(workflow)
	expectedTaskGroupDirs := make(map[string]struct{})
	expectedTaskFilePaths := make(map[string]struct{})
	for _, tg := range workflow.Data.TaskGroups {
		if err := writeTaskGroupYAML(
			workflow,
			dagDir,
			tg,
			taskAirflowIDByName,
			expectedTaskGroupDirs,
			expectedTaskFilePaths,
		); err != nil {
			return err
		}
	}

	// Safety guard: never wipe an existing DAG directory on an empty expected set.
	if len(expectedTaskGroupDirs) == 0 {
		return nil
	}

	if err := cleanupStaleDAGEntries(dagDir, expectedTaskGroupDirs, expectedTaskFilePaths); err != nil {
		return errors.New("failed to cleanup stale DAG files. (Error: " + err.Error() + ")")
	}

	return nil
}

func ensureWorkflowDir(workflow *model.Workflow) (string, error) {
	// 워크플로우 전용 DAG 디렉터리를 보장한다.
	dagID := workflow.ID
	if workflow.WorkflowKey != "" {
		dagID = workflow.WorkflowKey
	}
	dagDir := filepath.Clean(filepath.Join(config.CMCicadaConfig.CMCicada.DAGDirectoryHost, dagID))

	if err := fileutil.CreateDirIfNotExist(dagDir); err != nil {
		return "", errors.New("failed to create the Workflow directory (Workflow ID=" + workflow.ID +
			", Workflow Name=" + workflow.Name + ", Description: " + workflow.Data.Description)
	}
	return dagDir, nil
}

func writeDAGMetadata(workflow *model.Workflow, dagDir string) error {
	// DAG 메타데이터(METADATA.yml)를 기록한다.
	// 워크플로우에 active schedule이 있으면 Airflow가 알아서 1회 실행하도록
	// schedule="@once", start_date=run_at,  catchup=false 를 함께 기록한다.
	type defaultArgs struct {
		Owner         string `yaml:"owner"`
		StartDate     string `yaml:"start_date"`
		Retries       int    `yaml:"retries"`
		RetryDelaySec int    `yaml:"retry_delay_sec"`
	}

	var dagInfo struct {
		DefaultArgs    defaultArgs `yaml:"default_args"`
		Description    string      `yaml:"description"`
		DagDisplayName string      `yaml:"dag_display_name"`
		Schedule       string      `yaml:"schedule,omitempty"`
		Catchup        *bool       `yaml:"catchup,omitempty"`
	}

	startDate := time.Now().Format(time.DateOnly)
	if schedule, err := dao.WorkflowScheduleGetActive(workflow.ID); err == nil && schedule != nil {
		switch schedule.Type {
		case model.WorkflowScheduleTypeOnce:
			if schedule.RunAt != nil {
				startDate = schedule.RunAt.UTC().Format(time.RFC3339)
				dagInfo.Schedule = "@once"
				catchup := false
				dagInfo.Catchup = &catchup
			}
		case model.WorkflowScheduleTypeCron:
			if schedule.Cron != nil && *schedule.Cron != "" {
				dagInfo.Schedule = *schedule.Cron
				catchup := false
				dagInfo.Catchup = &catchup
			}
		}
	}

	dagInfo.DefaultArgs = defaultArgs{
		Owner:         strings.ToLower(common.ModuleName),
		StartDate:     startDate,
		Retries:       0,
		RetryDelaySec: 0,
	}
	dagInfo.Description = workflow.Data.Description
	dagInfo.DagDisplayName = workflow.Name

	filePath := filepath.Join(dagDir, "METADATA.yml")
	if err := writeModelToYAMLFile(dagInfo, filePath); err != nil {
		return errors.New("failed to write YAML file (FilePath: " + filePath + ", Error: " + err.Error() + ")")
	}
	return nil
}

func buildTaskAirflowIDByName(workflow *model.Workflow) map[string]string {
	// task name -> Airflow task_id(UUID) 매핑을 구성한다.
	taskAirflowIDByName := make(map[string]string)

	for _, tg := range workflow.Data.TaskGroups {
		for _, t := range tg.Tasks {
			taskAirflowID := t.Name
			if taskModel, err := dao.TaskGetByWorkflowIDAndName(workflow.ID, t.Name); err == nil {
				switch {
				case taskModel.TaskKey != "":
					taskAirflowID = taskModel.TaskKey
				case taskModel.ID != "":
					taskAirflowID = taskModel.ID
				}
			}
			taskAirflowIDByName[t.Name] = taskAirflowID
		}
	}

	return taskAirflowIDByName
}

func writeTaskGroupYAML(
	workflow *model.Workflow,
	dagDir string,
	tg model.TaskGroup,
	taskAirflowIDByName map[string]string,
	expectedTaskGroupDirs map[string]struct{},
	expectedTaskFilePaths map[string]struct{},
) error {
	// TaskGroup 메타데이터와 각 Task YAML을 생성한다.
	tgDirName := tg.ID
	if tgDirName == "" {
		tgDirName = tg.Name
	}
	tgDir := filepath.Clean(filepath.Join(dagDir, tgDirName))
	expectedTaskGroupDirs[filepath.Clean(tgDir)] = struct{}{}
	if err := fileutil.CreateDirIfNotExist(tgDir); err != nil {
		return err
	}

	var taskGroupMeta struct {
		Tooltip         string `yaml:"tooltip"`
		TaskDisplayName string `yaml:"task_display_name"`
	}
	taskGroupMeta.Tooltip = tg.Description
	taskGroupMeta.TaskDisplayName = tg.Name

	metadataPath := filepath.Join(tgDir, "METADATA.yml")
	if err := writeModelToYAMLFile(taskGroupMeta, metadataPath); err != nil {
		return errors.New("failed to write YAML file (FilePath: " + metadataPath + ", Error: " + err.Error() + ")")
	}

	for _, t := range tg.Tasks {
		if err := writeTaskYAML(workflow, tgDir, t, taskAirflowIDByName, expectedTaskFilePaths); err != nil {
			return err
		}
	}

	return nil
}

func writeTaskYAML(
	workflow *model.Workflow,
	tgDir string,
	t model.Task,
	taskAirflowIDByName map[string]string,
	expectedTaskFilePaths map[string]struct{},
) error {
	// 단일 Task YAML을 생성하고, 예상 파일 목록에 등록한다.
	taskOptions, err := buildTaskOptions(workflow, t, taskAirflowIDByName)
	if err != nil {
		return err
	}

	taskOptions["dependencies"] = resolveDependencies(t.Dependencies, taskAirflowIDByName)

	if taskID, exists := taskAirflowIDByName[t.Name]; exists {
		taskOptions["task_id"] = taskID
	} else {
		taskOptions["task_id"] = t.Name
	}
	if t.Name != "" {
		taskOptions["task_display_name"] = t.Name
	}

	taskFileName := t.ID
	if taskFileName == "" {
		taskFileName = t.Name
	}
	filePath := filepath.Clean(filepath.Join(tgDir, taskFileName+".yml"))
	expectedTaskFilePaths[filepath.Clean(filePath)] = struct{}{}

	if err := writeModelToYAMLFile(taskOptions, filePath); err != nil {
		return errors.New("failed to write YAML file (FilePath: " + filePath + ", Error: " + err.Error() + ")")
	}

	return nil
}

func buildTaskOptions(
	workflow *model.Workflow,
	t model.Task,
	taskAirflowIDByName map[string]string,
) (map[string]any, error) {
	taskComponent := dao.TaskComponentGetByName(t.TaskComponent)
	if taskComponent == nil {
		return nil, errors.New("task component '" + t.TaskComponent + "' not found")
	}

	typeDef, ok := catalog.Get(taskComponent.Type)
	if !ok {
		return nil, errors.New("unknown task type in catalog: " + taskComponent.Type)
	}

	logger.Println(logger.INFO, true,
		fmt.Sprintf("task component name=%s type=%s spec=%v",
			taskComponent.Name, taskComponent.Type, taskComponent.Spec))

	switch taskComponent.Type {
	case "http":
		return buildHTTPTaskOptions(typeDef, taskComponent, t)
	case "http_xcom":
		return buildHTTPXcomTaskOptions(typeDef, taskComponent, t, workflow, taskAirflowIDByName)
	case "bash":
		return buildBashTaskOptions(typeDef, taskComponent, t)
	case "ssh":
		return buildSSHTaskOptions(typeDef, taskComponent, t)
	case "trigger_workflow":
		return buildTriggerWorkflowTaskOptions(typeDef, taskComponent, t)
	default:
		// Fallback: pass-through with operator class only.
		return map[string]any{"operator": typeDef.OperatorClass}, nil
	}
}

func buildHTTPTaskOptions(typeDef catalog.TaskTypeDef, c *model.TaskComponent, t model.Task) (map[string]any, error) {
	merged := mergeSpecs(c.Spec, t.Spec)

	taskOptions := map[string]any{
		"operator":     typeDef.OperatorClass,
		"http_conn_id": specString(merged, "api_connection_id"),
		"method":       specString(merged, "method"),
		"log_response": true,
	}

	body := specString(merged, "request_body")
	if body != "" {
		taskOptions["data"] = body
	}

	headers := map[string]any{"Content-Type": "application/json"}
	if extra, ok := merged["headers"].(map[string]any); ok {
		for k, v := range extra {
			headers[k] = v
		}
	}
	taskOptions["headers"] = headers

	endpointTemplate := specString(merged, "endpoint")
	pathParams := specStringMap(merged, "path_params")
	queryParams := specStringMap(merged, "query_params")
	endpoint, err := parseEndpoint(pathParams, queryParams, endpointTemplate)
	if err != nil {
		return nil, err
	}
	taskOptions["endpoint"] = endpoint

	return taskOptions, nil
}

func buildHTTPXcomTaskOptions(
	typeDef catalog.TaskTypeDef,
	c *model.TaskComponent,
	t model.Task,
	workflow *model.Workflow,
	taskAirflowIDByName map[string]string,
) (map[string]any, error) {
	merged := mergeSpecs(c.Spec, t.Spec)

	taskOptions := map[string]any{
		"operator":     typeDef.OperatorClass,
		"http_conn_id": specString(merged, "api_connection_id"),
		"method":       specString(merged, "method"),
	}
	endpointTemplate := specString(merged, "endpoint")
	pathParams := specStringMap(merged, "path_params")
	queryParams := specStringMap(merged, "query_params")
	endpoint, err := parseEndpoint(pathParams, queryParams, endpointTemplate)
	if err != nil {
		return nil, err
	}
	taskOptions["endpoint"] = endpoint

	xcomSource := specString(merged, "request_body")
	if xcomSource == "" {
		return nil, errors.New("http_xcom task is missing spec.request_body")
	}
	if isTaskExist(workflow, xcomSource) {
		if id, ok := taskAirflowIDByName[xcomSource]; ok {
			taskOptions["xcom_task"] = id
		} else {
			taskOptions["xcom_task"] = xcomSource
		}
	} else {
		taskOptions["xcom_task"] = xcomSource
	}
	return taskOptions, nil
}

func buildBashTaskOptions(typeDef catalog.TaskTypeDef, c *model.TaskComponent, t model.Task) (map[string]any, error) {
	merged := mergeSpecs(c.Spec, t.Spec)
	cmd := specString(merged, "bash_command")
	if cmd == "" {
		return nil, errors.New("bash task is missing spec.bash_command")
	}
	return map[string]any{
		"operator":     typeDef.OperatorClass,
		"bash_command": cmd,
	}, nil
}

func buildSSHTaskOptions(typeDef catalog.TaskTypeDef, c *model.TaskComponent, t model.Task) (map[string]any, error) {
	merged := mergeSpecs(c.Spec, t.Spec)
	connID := specString(merged, "ssh_conn_id")
	if connID == "" {
		return nil, errors.New("ssh task is missing spec.ssh_conn_id")
	}
	cmd := specString(merged, "command")
	if cmd == "" {
		return nil, errors.New("ssh task is missing spec.command")
	}
	return map[string]any{
		"operator":    typeDef.OperatorClass,
		"ssh_conn_id": connID,
		"command":     cmd,
	}, nil
}

func buildTriggerWorkflowTaskOptions(typeDef catalog.TaskTypeDef, c *model.TaskComponent, t model.Task) (map[string]any, error) {
	merged := mergeSpecs(c.Spec, t.Spec)
	taskOptions := map[string]any{
		"operator": typeDef.OperatorClass,
	}
	for k, v := range merged {
		taskOptions[k] = v
	}
	if _, ok := taskOptions["trigger_dag_id"]; !ok {
		return nil, errors.New("trigger_workflow task is missing trigger_dag_id")
	}
	return taskOptions, nil
}

// mergeSpecs returns a new Spec that contains all keys from base, overridden
// by keys from override (task-level wins over component-level).
func mergeSpecs(base, override model.Spec) model.Spec {
	out := model.Spec{}
	for k, v := range base {
		out[k] = v
	}
	for k, v := range override {
		out[k] = v
	}
	return out
}

// specStringMap extracts a map[string]string from a Spec value.
// Accepts both map[string]string and map[string]any (string-valued).
func specStringMap(s model.Spec, key string) map[string]string {
	if s == nil {
		return nil
	}
	v, ok := s[key]
	if !ok {
		return nil
	}
	switch m := v.(type) {
	case map[string]string:
		return m
	case map[string]any:
		out := make(map[string]string, len(m))
		for k, val := range m {
			if str, ok := val.(string); ok {
				out[k] = str
			} else {
				out[k] = fmt.Sprint(val)
			}
		}
		return out
	}
	return nil
}

func specString(s model.Spec, key string) string {
	if s == nil {
		return ""
	}
	v, ok := s[key]
	if !ok {
		return ""
	}
	if str, ok := v.(string); ok {
		return str
	}
	return fmt.Sprint(v)
}

func resolveDependencies(dependencies []string, taskAirflowIDByName map[string]string) []string {
	// 의존성 이름을 Airflow task_id로 치환한다.
	resolved := make([]string, 0, len(dependencies))
	for _, dep := range dependencies {
		if depID, exists := taskAirflowIDByName[dep]; exists {
			resolved = append(resolved, depID)
		} else {
			resolved = append(resolved, dep)
		}
	}
	return resolved
}

func cleanupStaleDAGEntries(
	dagDir string,
	expectedTaskGroupDirs map[string]struct{},
	expectedTaskFilePaths map[string]struct{},
) error {
	// 현재 워크플로우에 없는 잔여 TaskGroup/Task 파일을 정리한다.
	normalizedTaskGroupDirs := make(map[string]struct{}, len(expectedTaskGroupDirs))
	for path := range expectedTaskGroupDirs {
		normalizedTaskGroupDirs[filepath.Clean(path)] = struct{}{}
	}
	normalizedTaskFilePaths := make(map[string]struct{}, len(expectedTaskFilePaths))
	for path := range expectedTaskFilePaths {
		normalizedTaskFilePaths[filepath.Clean(path)] = struct{}{}
	}

	entries, err := os.ReadDir(dagDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		path := filepath.Clean(filepath.Join(dagDir, entry.Name()))
		if entry.IsDir() {
			if _, exists := normalizedTaskGroupDirs[path]; !exists {
				if err := os.RemoveAll(path); err != nil {
					return err
				}
				continue
			}

			if err := cleanupStaleTaskFilesInGroup(path, normalizedTaskFilePaths); err != nil {
				return err
			}
			continue
		}

		// Keep workflow metadata and remove stale top-level YAML files.
		if strings.EqualFold(entry.Name(), "METADATA.yml") {
			continue
		}
		if strings.EqualFold(filepath.Ext(entry.Name()), ".yml") {
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				return err
			}
		}
	}

	return nil
}

func cleanupStaleTaskFilesInGroup(groupDir string, expectedTaskFilePaths map[string]struct{}) error {
	// TaskGroup 디렉터리 내 잔여 Task YAML 파일을 정리한다.
	entries, err := os.ReadDir(groupDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		path := filepath.Join(groupDir, entry.Name())
		if entry.IsDir() {
			// Nested directories are not expected in a task-group folder.
			if err := os.RemoveAll(path); err != nil {
				return err
			}
			continue
		}

		if strings.EqualFold(entry.Name(), "METADATA.yml") {
			continue
		}
		if !strings.EqualFold(filepath.Ext(entry.Name()), ".yml") {
			continue
		}
		if _, exists := expectedTaskFilePaths[path]; exists {
			continue
		}

		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	return nil
}
