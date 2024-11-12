package airflow

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/cloud-barista/cm-cicada/common"
	"github.com/cloud-barista/cm-cicada/db"
	"github.com/cloud-barista/cm-cicada/lib/config"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/jollaman999/utils/fileutil"
	"gopkg.in/yaml.v3"
)

func checkWorkflow(workflow *model.Workflow) error {
	var taskNames []string

	for _, tg := range workflow.Data.TaskGroups {
		if tg.Name == "" {
			return errors.New("task group name should not be empty")
		}

		for _, t := range tg.Tasks {
			if t.Name == "" {
				return errors.New("task name should not be empty")
			}

			taskNames = append(taskNames, t.Name)
		}
	}

	for _, tg := range workflow.Data.TaskGroups {
		for _, t := range tg.Tasks {
			taskComponent := db.TaskComponentGetByName(t.TaskComponent)
			if taskComponent == nil {
				return errors.New("task component '" + t.TaskComponent + "' not found")
			}

			for _, dep := range t.Dependencies {
				var depFound bool
				for _, tName := range taskNames {
					if tName == dep {
						depFound = true
						break
					}
				}
				if !depFound {
					return errors.New("wrong dependency found in " + tg.Name + "." + t.Name + " (" + dep + ")")
				}
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

func parseEndpoint(pathParams map[string]string, queryParams map[string]string, endpoint string) string {
	pathParamKeys := reflect.ValueOf(pathParams).MapKeys()
	for _, key := range pathParamKeys {
		endpoint = strings.ReplaceAll(endpoint, "{"+key.String()+"}", pathParams[key.String()])
	}

	queryParamKeys := reflect.ValueOf(pathParams).MapKeys()
	if len(queryParamKeys) > 0 {
		if !strings.Contains(endpoint, "?") {
			endpoint += "?"
		}
		for _, key := range queryParamKeys {
			endpoint += fmt.Sprintf("%v=%v&", key.String(), queryParams[key.String()])
		}
		endpoint = strings.TrimRight(endpoint, "&")
	}

	return endpoint
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
	err := checkWorkflow(workflow)
	if err != nil {
		return err
	}

	dagDir := config.CMCicadaConfig.CMCicada.DAGDirectoryHost + "/" + workflow.ID
	err = fileutil.CreateDirIfNotExist(dagDir)
	if err != nil {
		return errors.New("failed to create the Workflow directory (Workflow ID=" + workflow.ID +
			", Workflow Name=" + workflow.Name + ", Description: " + workflow.Data.Description)
	}

	type defaultArgs struct {
		Owner         string `yaml:"owner"`
		StartDate     string `yaml:"start_date"`
		Retries       int    `yaml:"retries"`
		RetryDelaySec int    `yaml:"retry_delay_sec"`
	}

	var dagInfo struct {
		defaultArgs defaultArgs `yaml:"default_args"`
		Description string      `yaml:"description"`
	}

	dagInfo.defaultArgs = defaultArgs{
		Owner:         strings.ToLower(common.ModuleName),
		StartDate:     time.Now().Format(time.DateOnly),
		Retries:       0,
		RetryDelaySec: 0,
	}
	dagInfo.Description = workflow.Data.Description

	filePath := dagDir + "/METADATA.yml"

	err = writeModelToYAMLFile(dagInfo, filePath)
	if err != nil {
		return errors.New("failed to write YAML file (FilePath: " + filePath + ", Error: " + err.Error() + ")")
	}

	for _, tg := range workflow.Data.TaskGroups {
		err = fileutil.CreateDirIfNotExist(dagDir + "/" + tg.Name)
		if err != nil {
			return err
		}

		var taskGroup struct {
			Tooltip string `yaml:"tooltip"`
		}

		taskGroup.Tooltip = tg.Description

		filePath = dagDir + "/" + tg.Name + "/METADATA.yml"

		err = writeModelToYAMLFile(taskGroup, filePath)
		if err != nil {
			return errors.New("failed to write YAML file (FilePath: " + filePath + ", Error: " + err.Error() + ")")
		}

		for _, t := range tg.Tasks {
			taskOptions := make(map[string]any)
			taskComponent := db.TaskComponentGetByName(t.TaskComponent)
			if taskComponent == nil {
				return errors.New("task component '" + t.TaskComponent + "' not found")
			}
			if taskComponent.Data.Options.Extra != nil {
				taskOptions = taskComponent.Data.Options.Extra

			} else {
				if isTaskExist(workflow, t.RequestBody) {
					taskOptions["operator"] = "local.JsonHttpRequestOperator"
					taskOptions["xcom_task"] = t.RequestBody
				} else {
					taskOptions["operator"] = "airflow.providers.http.operators.http.SimpleHttpOperator"

					type headers struct {
						ContentType string `json:"Content-Type" yaml:"Content-Type"`
					}
					taskOptions["headers"] = headers{
						ContentType: "application/json",
					}

					taskOptions["log_response"] = true

					taskOptions["data"] = t.RequestBody
				}

				taskOptions["http_conn_id"] = taskComponent.Data.Options.APIConnectionID
				taskOptions["endpoint"] = parseEndpoint(t.PathParams, t.QueryParams, taskComponent.Data.Options.Endpoint)
				taskOptions["method"] = taskComponent.Data.Options.Method
			}

			taskOptions["dependencies"] = t.Dependencies

			taskOptions["task_id"] = t.Name

			filePath = dagDir + "/" + tg.Name + "/" + t.Name + ".yml"

			err = writeModelToYAMLFile(taskOptions, filePath)
			if err != nil {
				return errors.New("failed to write YAML file (FilePath: " + filePath + ", Error: " + err.Error() + ")")
			}
		}
	}

	return nil
}
