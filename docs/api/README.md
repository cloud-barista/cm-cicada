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
      "default_args": {
        "owner": "ish",
        "start_date": "2024-03-29",
        "retries": 0,
        "retry_delay_sec": 0
      },
      "description": "Migrate Server",
      "task_groups": [
        {
          "task_group_name": "migrate_infra",
          "description": "Migrate Server",
          "tasks": [
            {
              "task_name": "beetle_task_migration_infra",
              "operator": "airflow.providers.http.operators.http.SimpleHttpOperator",
              "operator_options": [
                {
                  "name": "task_id",
                  "value": "migrate_infra"
                },
                {
                  "name": "http_conn_id",
                  "value": "beetle_api"
                },
                {
                  "name": "endpoint",
                  "value": "/beetle/migration/infra"
                },
                {
                  "name": "method",
                  "value": "POST"
                },
                {
                  "name": "data",
                  "value": "{\n    \"name\": \"recommended-infra01\",\n    \"installMonAgent\": \"no\",\n    \"label\": \"DynamicVM\",\n    \"systemLabel\": \"\",\n    \"description\": \"Made in CB-TB\",\n    \"vm\": [\n        {\n            \"name\": \"recommended-vm01\",\n            \"subGroupSize\": \"3\",\n            \"label\": \"DynamicVM\",\n            \"description\": \"Description\",\n            \"commonSpec\": \"azure-koreacentral-standard-b4ms\",\n            \"commonImage\": \"ubuntu22-04\",\n            \"rootDiskType\": \"default\",\n            \"rootDiskSize\": \"default\",\n            \"vmUserPassword\": \"test\",\n            \"connectionName\": \"azure-koreacentral\"\n        }\n    ]\n}"
                },
                {
                  "name": "headers",
                  "value": {
                    "Content-Type": "application/json"
                  }
                }
              ],
              "dependencies": []
            }
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
