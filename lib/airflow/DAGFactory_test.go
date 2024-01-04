package airflow

import (
	"fmt"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"testing"
)

func Test_Logger_Prepare(t *testing.T) {
	example := model.DAG{
		DAGId: "docker_nexus_dag_factory",
		DefaultArgs: model.DAGDefaultArgs{
			Owner:         "ish",
			StartDate:     "2023-10-31",
			Retries:       0,
			RetryDelaySec: 0,
		},
		DefaultView: "",
		Orientation: "",
		Description: "",
		TaskGroups: []model.DAGTaskGroup{
			{
				TaskGroupName: "infra_migration",
				Tooltip:       "this is a task group of infra migration",
			},
			{
				TaskGroupName: "data_migration",
				Tooltip:       "this is a task group of data migration",
			},
			{
				TaskGroupName: "software_migration",
				Tooltip:       "this is a task group of software migration",
			},
		},
		Tasks: []model.DAGTask{
			{
				TaskName: "infra_task_start",
				Operator: "airflow.operators.bash.BashOperator",
				OperatorOptions: []model.DAGOperatorOption{
					{
						Name:  "bash_command",
						Value: "echo 'Infra migration is started.'",
					},
				},
				TaskGroupName: "infra_migration",
				Dependencies:  []string{},
			},
			{
				TaskName: "infra_task_end",
				Operator: "airflow.operators.bash.BashOperator",
				OperatorOptions: []model.DAGOperatorOption{
					{
						Name:  "bash_command",
						Value: "echo 'Infra migration is ended.'",
					},
				},
				TaskGroupName: "infra_migration",
				Dependencies:  []string{"infra_task_start"},
			},
			{
				TaskName: "data_task_start",
				Operator: "airflow.operators.bash.BashOperator",
				OperatorOptions: []model.DAGOperatorOption{
					{
						Name:  "bash_command",
						Value: "echo 'Data migration is started.'",
					},
				},
				TaskGroupName: "data_migration",
				Dependencies:  []string{"infra_migration"},
			},
			{
				TaskName: "data_task_end",
				Operator: "airflow.operators.bash.BashOperator",
				OperatorOptions: []model.DAGOperatorOption{
					{
						Name:  "bash_command",
						Value: "echo 'Data migration is ended.'",
					},
				},
				TaskGroupName: "data_migration",
				Dependencies:  []string{"data_task_start"},
			},
			{
				TaskName: "software_task_start",
				Operator: "airflow.operators.bash.BashOperator",
				OperatorOptions: []model.DAGOperatorOption{
					{
						Name:  "bash_command",
						Value: "echo 'Software migration is started.'",
					},
				},
				TaskGroupName: "software_migration",
				Dependencies:  []string{"data_migration"},
			},
			{
				TaskName: "mkdir_data",
				Operator: "airflow.operators.bash.BashOperator",
				OperatorOptions: []model.DAGOperatorOption{
					{
						Name:  "bash_command",
						Value: "mkdir -p /data/volume/nexus",
					},
				},
				TaskGroupName: "software_migration",
				Dependencies:  []string{"software_task_start"},
			},
			{
				TaskName: "chown_data",
				Operator: "airflow.operators.bash.BashOperator",
				OperatorOptions: []model.DAGOperatorOption{
					{
						Name:  "bash_command",
						Value: "chmod 777 /data/volume/nexus",
					},
				},
				TaskGroupName: "software_migration",
				Dependencies:  []string{"mkdir_data"},
			},
			{
				TaskName: "docker_sample_task",
				Operator: "airflow.providers.docker.operators.docker.DockerOperator",
				OperatorOptions: []model.DAGOperatorOption{
					{
						Name:  "image",
						Value: "sonatype/nexus3:latest",
					},
					{
						Name:  "mounts",
						Value: "[{\"source\": \"/data/volume/nexus\", \"target\": \"/nexus-data\", \"type\": \"bind\"}]",
					},
					{
						Name:  "api_version",
						Value: "auto",
					},
					{
						Name:  "auto_remove",
						Value: "success",
					},
					{
						Name:  "docker_url",
						Value: "unix://var/run/docker.sock",
					},
					{
						Name:  "network_mode",
						Value: "bridge",
					},
				},
				TaskGroupName: "software_migration",
				Dependencies:  []string{"chown_data"},
			},
			{
				TaskName: "software_task_end",
				Operator: "airflow.operators.bash.BashOperator",
				OperatorOptions: []model.DAGOperatorOption{
					{
						Name:  "bash_command",
						Value: "echo 'Software migration is ended.'",
					},
				},
				TaskGroupName: "software_migration",
				Dependencies:  []string{"docker_sample_task"},
			},
		},
	}

	result, err := changeModelDAGToYAMLString(&example)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(result)
}
