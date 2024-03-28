import json
import os
from datetime import datetime

from airflow import DAG
from airflow.providers.http.operators.http import SimpleHttpOperator
from airflow.models.connection import Connection

DAG_ID = "beetle_migrate_infra_test_1"

dag = DAG(
    DAG_ID,
    default_args={"retries": 1},
    start_date=datetime(2024, 3, 27),
)

task_migrate_infra = SimpleHttpOperator(
    task_id="migrate_infra",
    http_conn_id="beetle_api",
    endpoint="/beetle/migration/infra",
    method="POST",
    data=json.dumps(
{
    "name": "recommended-infra01",
    "installMonAgent": "no",
    "label": "DynamicVM",
    "systemLabel": "",
    "description": "Made in CB-TB",
    "vm": [
        {
            "name": "recommended-vm01",
            "subGroupSize": "3",
            "label": "DynamicVM",
            "description": "Description",
            "commonSpec": "azure-koreacentral-standard-b4ms",
            "commonImage": "ubuntu22-04",
            "rootDiskType": "default",
            "rootDiskSize": "default",
            "vmUserPassword": "test",
            "connectionName": "azure-koreacentral"
        }
    ]
}
    ),
    headers={"Content-Type": "application/json"},
    dag=dag,
)

(
        task_migrate_infra
)
