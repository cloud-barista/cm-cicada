from airflow.models import DAG, DagRun
from airflow.operators.python import PythonOperator
from airflow.operators.email import EmailOperator
from airflow.utils.dates import days_ago
from airflow.utils.state import State
import os



def load_email_template():
    current_dir = os.path.dirname(__file__)  # mail.py 파일이 있는 경로
    file_path = os.path.join(current_dir, 'templates', 'email_template.html')
    with open(file_path, 'r') as f:
        return f.read()

email_template_content = load_email_template()

# 실패한 태스크 수집 함수
def collect_failed_tasks(**context):
    from airflow.utils.db import provide_session

    @provide_session
    def _inner(session=None):
        # conf에서 전달받은 dag_id와 dag_run_id
        source_workflow_id = context['dag_run'].conf.get('source_workflow_id')
        source_workflow_run_id = context['dag_run'].conf.get('source_workflow_run_id')

        if not source_workflow_id or not source_workflow_run_id:
            raise ValueError("source_workflow_id source_workflow_run_id 전달되지 않았습니다.")

        # source_workflow_id source_workflow_run_id 이용해 DagRun 정보 가져오기
        source_workflow_run = session.query(DagRun).filter_by(
            dag_id=source_workflow_id,
            run_id=source_workflow_run_id
        ).first()

        if not source_workflow_run:
            raise ValueError("해당하는 DAG Run을 찾을 수 없습니다.")

        # 실패한 태스크 ID 목록 수집
        failed_tasks = []
        for task_instance in source_workflow_run.get_task_instances():
            if task_instance.state != State.SUCCESS:
                failed_tasks.append(task_instance.task_id)

        # DAG 상태를 실패한 태스크가 1개 이상인 경우 "failed"로 설정
        dag_state = "failed" if failed_tasks else "success"

        # 결과 반환
        return {
            "dag_id": source_workflow_id,
            "dag_run_id": source_workflow_run_id,
            "dag_state": dag_state,
            "failed_tasks": failed_tasks
        }

    return _inner()

# DAG 정의
with DAG(
    dag_id="monitor_dag",
    default_args={'start_date': days_ago(1)},
    schedule_interval=None,
) as dag:

    # 실패한 태스크 수집 태스크
    collect_task = PythonOperator(
        task_id='collect_failed_tasks',
        python_callable=collect_failed_tasks,
        provide_context=True,
    )

    # EmailOperator 설정
    email_task = EmailOperator(
        task_id='send_email',
        to="{{ dag_run.conf['to_email'] }}",
        subject='Workflow 상태 보고서',
        html_content=email_template_content
    )

    # collect_task의 반환 값을 email_task에 전달하기 위한 PythonOperator 후크
    def set_email_params(ti, **context):
        # Pull the result from the previous task
        task_result = ti.xcom_pull(task_ids='collect_failed_tasks')
        if task_result:
            # Update the email_task.params directly
            email_task.params.update(task_result)
        print( "params : " , email_task.params)

    update_email_params = PythonOperator(
        task_id="update_email_params",
        python_callable=set_email_params,
        provide_context=True,
    )

    # 의존성 설정
    collect_task >> update_email_params >> email_task