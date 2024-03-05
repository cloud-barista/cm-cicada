# Management of cloud migration workflow
This is a sub-system on Cloud-Barista platform provides a features of create and management workflow for cloud migration.

## Overview

* Create and management workflow through Airflow.
* Create workflow based on dag-factory.

## Execution and development environment
* Tested operating systems (OSs):
    * Ubuntu 23.10, Ubuntu 22.04, Ubuntu 18.04
* Language:
    * Go: 1.21.5

## Download the source code

Clone CM-Cicada repository

```bash
git clone https://github.com/cloud-barista/cm-cicada.git
```

## How to run

* Build and run the binary
   ```shell
  cd cm-cicada
   make start
   ```

## About configuration file
- Configuration file name is 'cm-cicada.yaml'
- The configuration file must be placed in one of the following directories.
    - .cm-cicada/conf directory under user's home directory
    - 'conf' directory where running the binary
    - 'conf' directory where placed in the path of 'CMCICADA_ROOT' environment variable
- Configuration options
    - airflow-server
        - address : Specify Airflow server's address ({IP or Domain}:{Port})
        - use_tls : Must be true if Airflow server uses HTTPS.
        - skip_tls_verify : Skip TLS/SSL certificate verification. Must be set if 'use_tls' is true.
        - timeout : HTTP timeout value as seconds.
        - username : Airflow login username.
        - password : Airflow login password.
    - dag_directory_host : Specify DAG directory of the host. (Mounted DAG directory used by Airflow containers)
    - dag_directory_airflow : Specify DAG directory of Airflow container. (DAG directory inside the container.)
- listen
    - port : Listen port of the API.
- Configuration file example
  ```yaml
  cm-cicada:
       airflow-server:
            address: 127.0.0.1:8080
            use_tls: false
            # skip_tls_verify: true
            timeout: 10
            username: "airflow"
            password: "airflow"
       dag_directory_host: "/home/ish/test/airflow/dags"
       dag_directory_airflow: "/opt/airflow/dags" # Use dag_directory_host for dag_directory_airflow, if this value is empty
       listen:
            port: 8083
  ```

## Health-check CM-Cicada

Check if CM-Cicada is running

```bash
curl http://localhost:8083/cicada/health

# Output if it's running successfully
# {"message":"CM-Cicada API server is running"}
```
