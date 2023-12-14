# API 가이드


## API 작성 규칙
[cm-beetle Useful samples to add new APIs 참고](https://github.com/cloud-barista/cm-beetle/blob/main/docs/useful-samples-to-add-new-apis.md)

## Using APIs

### Create DAG
- URL : http://127.0.0.1:8083/dag/create
- Method : POST
  - Request JSON body is needed
    <details>
    <summary>Request JSON body example</summary>
    
    ```json
    {
      "dag_id": "docker_nexus_dag_factory",
      "default_args": {
          "owner": "ish",
          "start_date": "2023-10-31"
      },
      "task_groups": [
          {
              "task_group_name": "infra_migration",
              "Tooltip": "this is a task group of infra migration"
          },
          {
              "task_group_name": "data_migration",
              "Tooltip": "this is a task group of data migration"
          },
          {
              "task_group_name": "software_migration",
              "Tooltip": "this is a task group of software migration"
          }
      ],
      "tasks": [
          {
              "task_name": "infra_task_start",
              "operator": "airflow.operators.bash.BashOperator",
              "operator_options": [
                  {
                      "name": "bash_command",
                      "value": "echo 'Infra migration is started.'"
                  }
              ],
              "task_group_name": "infra_migration",
              "dependencies": []
          },
          {
              "task_name": "infra_task_end",
              "operator": "airflow.operators.bash.BashOperator",
              "operator_options": [
                  {
                      "name": "bash_command",
                      "value": "echo 'Infra migration is ended.'"
                  }
              ],
              "task_group_name": "infra_migration",
              "dependencies": [
                  "infra_task_start"
              ]
          },
          {
              "task_name": "data_task_start",
              "operator": "airflow.operators.bash.BashOperator",
              "operator_options": [
                  {
                      "name": "bash_command",
                      "value": "echo 'Data migration is started.'"
                  }
              ],
              "task_group_name": "data_migration",
              "dependencies": [
                  "infra_migration"
              ]
          },
          {
              "task_name": "data_task_end",
              "operator": "airflow.operators.bash.BashOperator",
              "operator_options": [
                  {
                      "name": "bash_command",
                      "value": "echo 'Data migration is ended.'"
                  }
              ],
              "task_group_name": "data_migration",
              "dependencies": [
                  "data_task_start"
              ]
          },
          {
              "task_name": "software_task_start",
              "operator": "airflow.operators.bash.BashOperator",
              "operator_options": [
                  {
                      "name": "bash_command",
                      "value": "echo 'Software migration is started.'"
                  }
              ],
              "task_group_name": "software_migration",
              "dependencies": [
                  "data_migration"
              ]
          },
          {
              "task_name": "mkdir_data",
              "operator": "airflow.operators.bash.BashOperator",
              "operator_options": [
                  {
                      "name": "bash_command",
                      "value": "mkdir -p /data/volume/nexus"
                  }
              ],
              "task_group_name": "software_migration",
              "dependencies": [
                  "software_task_start"
              ]
          },
          {
              "task_name": "chown_data",
              "operator": "airflow.operators.bash.BashOperator",
              "operator_options": [
                  {
                      "name": "bash_command",
                      "value": "chmod 777 /data/volume/nexus"
                  }
              ],
              "task_group_name": "software_migration",
              "dependencies": [
                  "mkdir_data"
              ]
          },
          {
              "task_name": "docker_sample_task",
              "operator": "airflow.providers.docker.operators.docker.DockerOperator",
              "operator_options": [
                  {
                      "name": "image",
                      "value": "sonatype/nexus3:latest"
                  },
                  {
                      "name": "mounts",
                      "value": "[{\"source\": \"/data/volume/nexus\", \"target\": \"/nexus-data\", \"type\": \"bind\"}]"
                  },
                  {
                      "name": "api_version",
                      "value": "auto"
                  },
                  {
                      "name": "auto_remove",
                      "value": "success"
                  },
                  {
                      "name": "docker_url",
                      "value": "unix://var/run/docker.sock"
                  },
                  {
                      "name": "network_mode",
                      "value": "bridge"
                  }
              ],
              "task_group_name": "software_migration",
              "dependencies": [
                  "chown_data"
              ]
          },
          {
              "task_name": "software_task_end",
              "operator": "airflow.operators.bash.BashOperator",
              "operator_options": [
                  {
                      "name": "bash_command",
                      "value": "echo 'Software migration is ended.'"
                  }
              ],
              "task_group_name": "software_migration",
              "dependencies": [
                  "docker_sample_task"
              ]
          }
      ]
    }
    ```
    </details>
- Example
    ```
    http://127.0.0.1:8084/dag/create
    ```

### Get DAGs
- URL : http://127.0.0.1:8083/dag/dags
- Method : GET
- Example
    ```
    http://127.0.0.1:8083/dag/dags
    ```
    <details>
    <summary>Reply example</summary>

    ```json
    {
      "dags": [
          {
              "dag_id": "docker_nexus_dag",
              "default_view": "grid",
              "description": null,
              "file_token": "Ii9vcHQvYWlyZmxvdy9kYWdzL2RvY2tlcl9uZXh1c19kYWcucHki.aYSiGi_KzgwGikzw29o84KV2iTc",
              "fileloc": "/opt/airflow/dags/docker_nexus_dag.py",
              "has_import_errors": false,
              "has_task_concurrency_limits": false,
              "is_active": true,
              "is_paused": false,
              "is_subdag": false,
              "last_expired": null,
              "last_parsed_time": "2023-12-14T08:36:23.094077Z",
              "last_pickled": null,
              "max_active_runs": 16,
              "max_active_tasks": 16,
              "next_dagrun": null,
              "next_dagrun_create_after": null,
              "next_dagrun_data_interval_end": null,
              "next_dagrun_data_interval_start": null,
              "owners": [
                  "ish"
              ],
              "pickle_id": null,
              "root_dag_id": null,
              "schedule_interval": null,
              "scheduler_lock": null,
              "tags": [],
              "timetable_description": "Never, external triggers only"
          },
          {
              "dag_id": "docker_nexus_dag_factory",
              "default_view": "graph",
              "description": "",
              "file_token": "Ii9vcHQvYWlyZmxvdy9kYWdzL2RvY2tlcl9uZXh1c19kYWdfZmFjdG9yeS5weSI.-YO7N7Rha6m3gzMCLblkiVJrR5I",
              "fileloc": "/opt/airflow/dags/docker_nexus_dag_factory.py",
              "has_import_errors": false,
              "has_task_concurrency_limits": false,
              "is_active": true,
              "is_paused": false,
              "is_subdag": false,
              "last_expired": null,
              "last_parsed_time": "2023-12-14T08:35:22.878555Z",
              "last_pickled": null,
              "max_active_runs": 16,
              "max_active_tasks": 16,
              "next_dagrun": "2023-12-14T03:00:00Z",
              "next_dagrun_create_after": "2023-12-15T03:00:00Z",
              "next_dagrun_data_interval_end": "2023-12-15T03:00:00Z",
              "next_dagrun_data_interval_start": "2023-12-14T03:00:00Z",
              "owners": [
                  "ish"
              ],
              "pickle_id": null,
              "root_dag_id": null,
              "schedule_interval": {
                  "__type": "TimeDelta",
                  "days": 1,
                  "microseconds": 0,
                  "seconds": 0
              },
              "scheduler_lock": null,
              "tags": [],
              "timetable_description": ""
          },
          {
              "dag_id": "docker_nexus_task_group",
              "default_view": "grid",
              "description": null,
              "file_token": "Ii9vcHQvYWlyZmxvdy9kYWdzL2RvY2tlcl9uZXh1c190YXNrX2dyb3VwLnB5Ig.7oiD7S6ceY_x8VK1alMrPe2hWp0",
              "fileloc": "/opt/airflow/dags/docker_nexus_task_group.py",
              "has_import_errors": false,
              "has_task_concurrency_limits": false,
              "is_active": true,
              "is_paused": false,
              "is_subdag": false,
              "last_expired": null,
              "last_parsed_time": "2023-12-14T08:36:23.093919Z",
              "last_pickled": null,
              "max_active_runs": 16,
              "max_active_tasks": 16,
              "next_dagrun": null,
              "next_dagrun_create_after": null,
              "next_dagrun_data_interval_end": null,
              "next_dagrun_data_interval_start": null,
              "owners": [
                  "ish"
              ],
              "pickle_id": null,
              "root_dag_id": null,
              "schedule_interval": null,
              "scheduler_lock": null,
              "tags": [],
              "timetable_description": "Never, external triggers only"
          }
      ],
      "total_entries": 3
    }
    ```
    </details>

### Run DAG
- URL : http://127.0.0.1:8083/dag/run
- Method : POST
- Parameters
  - Needed
    - dag_id : DAG ID string
- Example
    ```
    http://127.0.0.1:8083/dag/run?dag_id=docker_nexus_dag_factory
    ```
    <details>
    <summary>Reply example</summary>

    ```json
    {
        "conf": {},
        "dag_id": "docker_nexus_dag_factory",
        "dag_run_id": "manual__2023-12-14T08:43:04.444301+00:00",
        "data_interval_end": "2023-12-14T08:43:04.444301Z",
        "data_interval_start": "2023-12-13T08:43:04.444301Z",
        "end_date": null,
        "execution_date": "2023-12-14T08:43:04.444301Z",
        "external_trigger": true,
        "last_scheduling_decision": null,
        "logical_date": "2023-12-14T08:43:04.444301Z",
        "note": null,
        "run_type": "manual",
        "start_date": null,
        "state": "queued"
    }
    ```
    </details>

