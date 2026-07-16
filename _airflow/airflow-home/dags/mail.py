from airflow.models import DAG
from airflow.configuration import conf
from airflow.providers.standard.operators.python import PythonOperator
from airflow.providers.smtp.operators.smtp import EmailOperator
from airflow.sdk import BaseHook
from urllib.parse import quote
import pendulum
import requests
import os

# Airflow REST API 접속 정보를 담은 커넥션. docker-compose에서
# AIRFLOW_CONN_AIRFLOW_API 환경변수로 정의한다.
AIRFLOW_API_CONN_ID = os.environ.get("AIRFLOW_API_CONN_ID", "airflow_api")

# Airflow API가 돌려주는 태스크 상태 문자열 기준.
FAILED_STATES = ("failed", "upstream_failed")


def load_email_template():
    current_dir = os.path.dirname(__file__)
    file_path = os.path.join(current_dir, 'templates', 'email_template.html')
    with open(file_path, 'r', encoding='utf-8') as f:
        return f.read()


email_template_content = load_email_template()


class _AirflowAPI:
    """Airflow REST API(/api/v2) 클라이언트.

    Airflow 3부터 태스크 코드에서 메타DB에 ORM으로 직접 접근할 수 없다
    ("Direct database access via the ORM is not allowed in Airflow 3.0").
    그래서 다른 워크플로우의 DAG run 정보는 REST API로 조회한다.
    """

    def __init__(self):
        conn = BaseHook.get_connection(AIRFLOW_API_CONN_ID)
        scheme = conn.schema or "http"
        port = f":{conn.port}" if conn.port else ""
        self.base_url = f"{scheme}://{conn.host}{port}"

        # Airflow 3은 Basic 인증을 받지 않는다. /auth/token 으로 JWT를 발급받아야 한다.
        resp = requests.post(
            f"{self.base_url}/auth/token",
            json={"username": conn.login, "password": conn.password},
            timeout=30,
        )
        resp.raise_for_status()
        self._headers = {"Authorization": f"Bearer {resp.json()['access_token']}"}

    def get(self, path):
        resp = requests.get(f"{self.base_url}{path}", headers=self._headers, timeout=30)
        resp.raise_for_status()
        return resp.json()

    def patch(self, path, body):
        resp = requests.patch(
            f"{self.base_url}{path}", headers=self._headers, json=body, timeout=30
        )
        resp.raise_for_status()
        return resp.json()


def collect_failed_tasks(**context):
    conf_data = context['dag_run'].conf or {}
    source_workflow_id = conf_data.get('source_workflow_id')
    source_workflow_run_id = conf_data.get('source_workflow_run_id')
    if not source_workflow_id or not source_workflow_run_id:
        raise ValueError("source_workflow_id, source_workflow_run_id 전달되지 않았습니다.")

    api = _AirflowAPI()

    # workflow_id / run_id를 경로 세그먼트 하나로 안전하게 넣는다.
    # ('/' 같은 문자가 들어와도 경로를 벗어나지 않게 한다)
    run_path = (
        f"/api/v2/dags/{quote(source_workflow_id, safe='')}"
        f"/dagRuns/{quote(source_workflow_run_id, safe='')}"
    )

    try:
        source_workflow_run = api.get(run_path)
    except requests.HTTPError as e:
        if e.response is not None and e.response.status_code == 404:
            raise ValueError("해당하는 DAG Run을 찾을 수 없습니다.")
        raise

    workflow_info = source_workflow_run.get("conf") or {}
    # Airflow 3에서 [webserver] 섹션이 없어지고 base_url이 [api]로 옮겨졌다.
    airflow_base_url = conf.get("api", "base_url")

    failed_tasks = []
    task_infos = []

    for ti in api.get(f"{run_path}/taskInstances").get("task_instances", []):
        state = ti.get("state")
        task_id = ti["task_id"]

        task_infos.append({
            "task_id": task_id,
            "task_name": ti.get("task_display_name") or task_id,
            "state": state,
            "try_number": ti.get("try_number"),
            # Airflow 3 API 응답에는 log_url이 없어서 UI 경로로 직접 조립한다.
            "log_url": f"{airflow_base_url}/dags/{source_workflow_id}"
                       f"/runs/{source_workflow_run_id}/tasks/{task_id}",
        })

        if state in FAILED_STATES:
            failed_tasks.append(task_id)

    # 기존의 session.merge()/session.commit() 대체.
    # PATCH가 받는 state는 queued/success/failed 뿐이다.
    dag_state = "failed" if failed_tasks else "success"
    api.patch(run_path, {"state": dag_state})

    dag_run_url = f"{airflow_base_url}/dags/{source_workflow_id}/runs/{source_workflow_run_id}"

    result = {
        "dag_id": workflow_info.get("workflow_key"),
        "dag_name": workflow_info.get("workflow_name"),
        "workflow_id": workflow_info.get("workflow_id"),
        "dag_run_id": source_workflow_run_id,
        "dag_state": dag_state,
        "failed_tasks": failed_tasks,
        "tasks": task_infos,
        "airflow_base_url": airflow_base_url,
        "dag_run_url": dag_run_url,
    }

    print("XCOM RESULT =", result)
    return result


with DAG(
    dag_id="monitor_dag",
    # Airflow 3에서 days_ago()가 삭제되었다. 외부 트리거 전용(schedule=None)이라
    # start_date는 고정값으로 충분하다.
    default_args={"start_date": pendulum.datetime(2024, 1, 1, tz="UTC")},
    schedule=None,
    catchup=False,
) as dag:

    collect_task = PythonOperator(
        task_id="collect_failed_tasks",
        python_callable=collect_failed_tasks,
    )

    email_task = EmailOperator(
        task_id="send_email",
        to="{{ dag_run.conf['to_email'] }}",
        subject="Workflow 상태 보고서 - {{ (ti.xcom_pull(task_ids='collect_failed_tasks') or {}).get('dag_name', '') }}",
        html_content=email_template_content,
    )

    collect_task >> email_task
