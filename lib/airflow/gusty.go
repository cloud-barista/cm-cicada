package airflow

import (
	"errors"
	"github.com/cloud-barista/cm-cicada/lib/config"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/google/uuid"
	"github.com/jollaman999/utils/fileutil"
	"gopkg.in/yaml.v3"
)

func checkDAG(dag *model.Workflow) error {
	if dag.DefaultArgs.Owner == "" {
		return errors.New("owner is not set")
	}

	if dag.DefaultArgs.StartDate == "" {
		return errors.New("start_date is not set")
	}

	var taskNames []string

	for _, tg := range dag.TaskGroups {
		if tg.TaskGroupName == "" {
			return errors.New("task group name should not be empty")
		}

		for _, t := range tg.Tasks {
			if t.TaskName == "" {
				return errors.New("task name should not be empty")
			}

			taskNames = append(taskNames, t.TaskName)
		}
	}

	for _, tg := range dag.TaskGroups {
		for _, t := range tg.Tasks {
			for _, dep := range t.Dependencies {
				var depFound bool
				for _, tName := range taskNames {
					if tName == dep {
						depFound = true
						break
					}
				}
				if !depFound {
					return errors.New("wrong dependency found in " + tg.TaskGroupName + "." + t.TaskName + " (" + dep + ")")
				}
			}
		}
	}

	return nil
}

func writeModelToYAMLFile(model any, filePath string) error {
	bytes, err := yaml.Marshal(model)
	if err != nil {
		return err
	}
	parsed := string(bytes)

	return fileutil.WriteFile(filePath, parsed)
}

func writeGustyYAMLs(dag *model.Workflow) error {
	err := checkDAG(dag)
	if err != nil {
		return err
	}

	dag.ID = uuid.New().String()

	dagDir := config.CMCicadaConfig.CMCicada.DAGDirectoryHost + "/" + dag.ID
	err = fileutil.CreateDirIfNotExist(dagDir)
	if err != nil {
		return errors.New("failed to create the Workflow directory (Workflow ID=" + dag.ID +
			", Description: " + dag.Description)
	}

	if dag.DefaultArgs.Retries < 0 {
		dag.DefaultArgs.Retries = 1
	}
	if dag.DefaultArgs.RetryDelaySec < 0 {
		dag.DefaultArgs.RetryDelaySec = 300
	}

	var dagInfo struct {
		DefaultArgs model.DefaultArgs `yaml:"default_args"`
		Description string            `yaml:"description"`
	}

	dagInfo.DefaultArgs = dag.DefaultArgs
	dagInfo.Description = dag.Description

	filePath := dagDir + "/METADATA.yml"

	err = writeModelToYAMLFile(dagInfo, filePath)
	if err != nil {
		return errors.New("failed to write YAML file (FilePath: " + filePath + ", Error: " + err.Error() + ")")
	}

	for _, tg := range dag.TaskGroups {
		err = fileutil.CreateDirIfNotExist(dagDir + "/" + tg.TaskGroupName)
		if err != nil {
			return err
		}

		var taskGroup struct {
			Tooltip string `yaml:"tooltip"`
		}

		taskGroup.Tooltip = tg.Description

		filePath = dagDir + "/" + tg.TaskGroupName + "/METADATA.yml"

		err = writeModelToYAMLFile(taskGroup, filePath)
		if err != nil {
			return errors.New("failed to write YAML file (FilePath: " + filePath + ", Error: " + err.Error() + ")")
		}

		for _, t := range tg.Tasks {
			taskOptions := make(map[string]any)

			taskOptions["operator"] = t.Operator
			taskOptions["dependencies"] = t.Dependencies
			for _, operatorOption := range t.OperatorOptions {
				taskOptions[operatorOption.Name] = operatorOption.Value
			}

			filePath = dagDir + "/" + tg.TaskGroupName + "/" + t.TaskName + ".yml"

			err = writeModelToYAMLFile(taskOptions, filePath)
			if err != nil {
				return errors.New("failed to write YAML file (FilePath: " + filePath + ", Error: " + err.Error() + ")")
			}
		}
	}

	return nil
}
