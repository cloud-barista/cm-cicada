# Cloud Migration Workflow Management
This is a subsystem of the Cloud-Barista platform that provides workflow management for cloud migration.

## Overview

* Create and management workflow through Airflow.
* Create workflow based on gusty.

## Development environment
* Tested operating systems (OSs):
    * Ubuntu 24.04, Ubuntu 22.04, Ubuntu 18.04
* Language:
    * Go: 1.21.6

## How to run

* Build and run binary with Airflow server
   ```shell
   make run
   ```

* Stop binary with Airflow server
   ```shell
   make stop
   ```

## About configuration file
- Configuration file name is 'cm-cicada.yaml'
- The configuration file must be placed in one of the following directories.
    - .cm-cicada/conf directory under user's home directory
    - 'conf' directory where running the binary
    - 'conf' directory where placed in the path of 'CMCICADA_ROOT' environment variable
- Configuration options
    - task_component
        - load_examples : Load task component examples if true.
        - examples_directory : Specify directory where task component examples are located. Must be set if 'load_examples' is true.
    - workflow_template
        - templates_directory : Specify directory where workflow templates are located.
    - airflow-server
        - address : Specify Airflow server's address ({IP or Domain}:{Port})
        - use_tls : Must be true if Airflow server uses HTTPS.
        - skip_tls_verify : Skip TLS/SSL certificate verification. Must be set if 'use_tls' is true.
        - init_retry : Retry count of initializing Airflow server connection used by cm-cicada.
        - timeout : HTTP timeout value as seconds.
        - username : Airflow login username.
        - password : Airflow login password.
        - connections : Pre-define Airflow connections (Set multiple connections)
          - id : ID of connection
          - type : Type of connection
          - description : Description of connection
          - host : Host address or URL of connection
          - port : Port number for use connection
          - schema : Connection schema
          - login : Username for use connection
          - password : Password for use connection
    - dag_directory_host : Specify DAG directory of the host. (Mounted DAG directory used by Airflow container.)
    - dag_directory_container : Specify DAG directory of Airflow container. (DAG directory inside the container.)
- listen
    - port : Listen port of the API.
- Configuration file example
  ```yaml
  cm-cicada:
    task_component:
        load_examples: true
        examples_directory: "./lib/airflow/example/task_component/"
    workflow_template:
        templates_directory: "./lib/airflow/example/workflow_template/"
    airflow-server:
        address: 127.0.0.1:8080
        use_tls: false
        # skip_tls_verify: true
        init_retry: 5
        timeout: 10
        username: "airflow"
        password: "airflow_pass"
        connections:
          - id: honeybee_api
            type: http
            description: HoneyBee API
            host: 127.0.0.1
            port: 8081
            schema: http
          - id: beetle_api
            type: http
            description: Beetle API
            host: 127.0.0.1
            port: 8056
            schema: http
            login: default
            password: default
          - id: tumblebug_api
            type: http
            description: TumbleBug API
            host: 127.0.0.1
            port: 1323
            schema: http
            login: default
            password: default
    dag_directory_host: "./_airflow/airflow-home/dags"
    dag_directory_container: "/usr/local/airflow/dags" # Use dag_directory_host for dag_directory_container, if this value is empty
    listen:
        port: 8083
  ```

## Health-check

Check if CM-Cicada is running

```bash
curl http://127.0.0.1:8083/cicada/readyz

# Output if it's running successfully
# {"message":"CM-Cicada API server is ready"}
```

## Check out all APIs
* [Cicada APIs (Swagger Document)](https://cloud-barista.github.io/cb-tumblebug-api-web/?url=https://raw.githubusercontent.com/cloud-barista/cm-cicada/main/pkg/api/rest/docs/swagger.yaml)
