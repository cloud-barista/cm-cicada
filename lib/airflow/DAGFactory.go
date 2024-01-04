package airflow

import (
	"errors"
	"github.com/cloud-barista/cm-cicada/lib/config"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/jollaman999/utils/fileutil"
	"gopkg.in/yaml.v3"
	"strings"
)

func changeModelDAGToYAMLString(DAG *model.DAG) (string, error) {
	if DAG.DAGId == "" {
		return "", errors.New("dag_id is not set")
	}

	if DAG.DefaultArgs.Owner == "" {
		return "", errors.New("owner is not set")
	}

	if DAG.DefaultArgs.StartDate == "" {
		return "", errors.New("start_date is not set")
	}

	retries := 1
	retryDelaySec := 300

	if DAG.DefaultArgs.Retries > 0 {
		retries = DAG.DefaultArgs.Retries
	} else {
		DAG.DefaultArgs.Retries = retries
	}
	if DAG.DefaultArgs.RetryDelaySec > 0 {
		retryDelaySec = DAG.DefaultArgs.RetryDelaySec
	} else {
		DAG.DefaultArgs.RetryDelaySec = retryDelaySec
	}

	defaultView := "graph"
	orientation := "LR"

	if DAG.DefaultView != "" {
		defaultView = DAG.DefaultView
	} else {
		DAG.DefaultView = defaultView
	}
	if DAG.Orientation != "" {
		orientation = DAG.Orientation
	} else {
		DAG.Orientation = orientation
	}

	var taskGroups []model.DAGFactoryDAGTaskGroup
	for _, tg := range DAG.TaskGroups {
		if tg.TaskGroupName == "" {
			return "", errors.New("task group name should not be empty")
		}

		taskGroups = append(taskGroups, model.DAGFactoryDAGTaskGroup{
			DAGFactoryDAGTaskGroupStruct: model.DAGFactoryDAGTaskGroupStruct{
				Tooltip: tg.Tooltip,
			}})
	}

	var tasks []model.DAGFactoryDAGTask
	for _, t := range DAG.Tasks {
		if t.TaskName == "" {
			return "", errors.New("task name should not be empty")
		}

		taskOptions := make(map[string]any)

		taskOptions["operator"] = t.Operator
		taskOptions["task_group_name"] = t.TaskGroupName
		taskOptions["dependencies"] = t.Dependencies

		for _, operatorOption := range t.OperatorOptions {
			taskOptions[operatorOption.Name] = operatorOption.Value
		}

		tasks = append(tasks, model.DAGFactoryDAGTask{
			DAGTaskStruct: taskOptions,
		})
	}

	template := model.DAGFactory{
		DAGStruct: model.DAGFactoryDAGStruct{
			DefaultArgs: model.DAGFactoryDAGDefaultArgs{
				Owner:         DAG.DefaultArgs.Owner,
				StartDate:     DAG.DefaultArgs.StartDate,
				Retries:       retries,
				RetryDelaySec: retryDelaySec,
			},
			DefaultView: defaultView,
			Orientation: orientation,
			Description: DAG.Description,
			TaskGroups:  taskGroups,
			Tasks:       tasks,
		}}

	bytes, err := yaml.Marshal(template)
	if err != nil {
		return "", err
	}
	parsed := string(bytes)

	parsed = strings.Replace(parsed,
		"'###dag_struct###'",
		DAG.DAGId,
		1)

	for _, tg := range DAG.TaskGroups {
		parsed = strings.Replace(parsed,
			"- '###dag_factory_dag_task_group_struct###'",
			tg.TaskGroupName,
			1)
	}

	for _, t := range DAG.Tasks {
		parsed = strings.Replace(parsed,
			"- '###dag_task_struct###'",
			t.TaskName,
			1)
	}

	return parsed, nil
}

func writeDAGFactoryPythonCode(DAGFactoryYAMLAirflowPath string, DAGFactoryPythonCodeHostPath string) error {
	pythonCode := "from airflow import DAG\n" +
		"import dagfactory\n" +
		"\n" +
		"dag_factory = dagfactory.DagFactory(\"" + DAGFactoryYAMLAirflowPath + "\")\n" +
		"\n" +
		"dag_factory.clean_dags(globals())\n" +
		"dag_factory.generate_dags(globals())"

	return fileutil.WriteFile(DAGFactoryPythonCodeHostPath, pythonCode)
}

func CreateDAGFactoryYAML(DAG *model.DAG) error {
	yamlString, err := changeModelDAGToYAMLString(DAG)
	if err != nil {
		return err
	}

	DAGFactoryYAMLHostPath := config.CMCicadaConfig.CMCicada.DAGDirectoryHost + "/" + DAG.DAGId + ".yml"
	DAGFactoryYAMLAirflowPath := config.CMCicadaConfig.CMCicada.DAGDirectoryAirflow + "/" + DAG.DAGId + ".yml"
	DAGFactoryPythonCodeHostPath := config.CMCicadaConfig.CMCicada.DAGDirectoryHost + "/" + DAG.DAGId + ".py"
	err = fileutil.WriteFile(DAGFactoryYAMLHostPath, yamlString)
	if err != nil {
		return err
	}

	err = writeDAGFactoryPythonCode(DAGFactoryYAMLAirflowPath, DAGFactoryPythonCodeHostPath)
	if err != nil {
		return err
	}

	return nil
}
