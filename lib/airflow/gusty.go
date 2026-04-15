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
	"github.com/cloud-barista/cm-cicada/db"
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
			taskComponent := db.TaskComponentGetByName(t.TaskComponent)
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
	type defaultArgs struct {
		Owner         string `yaml:"owner"`
		StartDate     string `yaml:"start_date"`
		Retries       int    `yaml:"retries"`
		RetryDelaySec int    `yaml:"retry_delay_sec"`
	}

	var dagInfo struct {
		defaultArgs    defaultArgs `yaml:"default_args"`
		Description    string      `yaml:"description"`
		DagDisplayName string      `yaml:"dag_display_name"`
	}

	dagInfo.defaultArgs = defaultArgs{
		Owner:         strings.ToLower(common.ModuleName),
		StartDate:     time.Now().Format(time.DateOnly),
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
			if db.DB != nil {
				taskModel := &model.TaskDBModel{}
				result := db.DB.Where("workflow_id = ? AND name = ? AND is_deleted = ?", workflow.ID, t.Name, false).First(taskModel)
				if result.Error == nil {
					switch {
					case taskModel.TaskKey != "":
						taskAirflowID = taskModel.TaskKey
					case taskModel.ID != "":
						taskAirflowID = taskModel.ID
					}
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
	// Task 옵션을 구성한다(기본 HTTP 또는 커스텀 operator).
	taskOptions := make(map[string]any)
	taskComponent := db.TaskComponentGetByName(t.TaskComponent)
	if taskComponent == nil {
		return nil, errors.New("task component '" + t.TaskComponent + "' not found")
	}

	logger.Println(logger.INFO, true, fmt.Sprintf("task component extra: %v", taskComponent.Data.Options.Extra))
	logger.Println(logger.INFO, true, fmt.Sprintf("task  extra: %v", t))

	if taskComponent.Data.Options.Extra != nil {
		taskOptions = copyMap(taskComponent.Data.Options.Extra)
		if t.Extra != nil {
			mergeMaps(taskOptions, t.Extra)
		}
		return taskOptions, nil
	}

	if isTaskExist(workflow, t.RequestBody) {
		taskOptions["operator"] = "local.JsonHttpRequestOperator"
		xcomTaskID, exists := taskAirflowIDByName[t.RequestBody]
		if exists {
			taskOptions["xcom_task"] = xcomTaskID
		} else {
			taskOptions["xcom_task"] = t.RequestBody
		}
	} else {
		taskOptions["operator"] = "airflow.providers.http.operators.http.SimpleHttpOperator"
		taskOptions["headers"] = map[string]any{
			"Content-Type": "application/json",
		}
		taskOptions["log_response"] = true
		taskOptions["data"] = t.RequestBody
		// Allow workflow task.extra.headers and coerce header values to string.
		applyTaskExtraHeaders(taskOptions, t.Extra)
	}

	taskOptions["http_conn_id"] = taskComponent.Data.Options.APIConnectionID
	endpoint, err := parseEndpoint(t.PathParams, t.QueryParams, taskComponent.Data.Options.Endpoint)
	if err != nil {
		return nil, err
	}
	taskOptions["endpoint"] = endpoint
	taskOptions["method"] = taskComponent.Data.Options.Method

	return taskOptions, nil
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

func copyMap(src map[string]any) map[string]any {
	// 중첩 map까지 안전하게 복사한다.
	if src == nil {
		return nil
	}
	dst := make(map[string]any, len(src))
	for k, v := range src {
		if nestedMap, ok := v.(map[string]any); ok {
			dst[k] = copyMap(nestedMap)
		} else {
			dst[k] = v
		}
	}
	return dst
}

func mergeMaps(dst map[string]any, src map[string]any) {
	// dst에 src를 병합한다(중첩 map 지원).
	if dst == nil || src == nil {
		return
	}

	for k, srcValue := range src {
		// dst에 같은 키가 없으면 그대로 추가
		if _, exists := dst[k]; !exists {
			dst[k] = srcValue
			continue
		}

		// dst에 같은 키가 있는 경우
		dstValue := dst[k]

		if dstMap, dstOk := dstValue.(map[string]any); dstOk {
			if srcMap, srcOk := srcValue.(map[string]any); srcOk {
				mergeMaps(dstMap, srcMap)
				continue
			}
		}

		dst[k] = srcValue
	}
}

func applyTaskExtraHeaders(taskOptions map[string]any, taskExtra map[string]any) {
	// task.extra.headers를 기본 헤더에 병합하고 문자열로 정규화한다.
	if taskOptions == nil || taskExtra == nil {
		return
	}

	rawHeaders, exists := taskExtra["headers"]
	if !exists {
		return
	}

	merged := map[string]any{
		"Content-Type": "application/json",
	}

	if existing, ok := toStringHeaderMap(taskOptions["headers"]); ok {
		for key, value := range existing {
			merged[key] = value
		}
	}

	userHeaders, ok := toStringHeaderMap(rawHeaders)
	if !ok {
		return
	}
	for key, value := range userHeaders {
		merged[key] = value
	}

	taskOptions["headers"] = merged
}

func toStringHeaderMap(raw any) (map[string]string, bool) {
	// 헤더 map을 string map으로 변환한다.
	if raw == nil {
		return nil, false
	}

	out := make(map[string]string)

	switch headers := raw.(type) {
	case map[string]any:
		for key, value := range headers {
			if key == "" {
				continue
			}
			strValue, ok := toHeaderString(value)
			if !ok {
				continue
			}
			out[key] = strValue
		}
		return out, true

	case map[string]string:
		for key, value := range headers {
			if key == "" {
				continue
			}
			out[key] = value
		}
		return out, true
	}

	return nil, false
}

func toHeaderString(value any) (string, bool) {
	// 다양한 타입의 값을 헤더용 문자열로 변환한다.
	switch v := value.(type) {
	case string:
		return v, true
	case bool:
		if v {
			return "true", true
		}
		return "false", true
	case int:
		return fmt.Sprintf("%d", v), true
	case int8:
		return fmt.Sprintf("%d", v), true
	case int16:
		return fmt.Sprintf("%d", v), true
	case int32:
		return fmt.Sprintf("%d", v), true
	case int64:
		return fmt.Sprintf("%d", v), true
	case uint:
		return fmt.Sprintf("%d", v), true
	case uint8:
		return fmt.Sprintf("%d", v), true
	case uint16:
		return fmt.Sprintf("%d", v), true
	case uint32:
		return fmt.Sprintf("%d", v), true
	case uint64:
		return fmt.Sprintf("%d", v), true
	case float32:
		return fmt.Sprintf("%v", v), true
	case float64:
		return fmt.Sprintf("%v", v), true
	case fmt.Stringer:
		return v.String(), true
	default:
		return "", false
	}
}
