from __future__ import annotations
import pendulum
from airflow.models import DAG
from airflow.operators.empty import EmptyOperator
from airflow.operators.bash_operator import BashOperator
from airflow.operators.python import PythonOperator
from airflow.operators.email import EmailOperator
from airflow.utils.state import State
from airflow.utils.task_group import TaskGroup
from airflow.utils.dates import days_ago

def collect_failed_tasks(**context):
    dag_run = context.get('dag_run')
    ti = context.get('ti')
    
    failed_tasks = []
    for task_instance in dag_run.get_task_instances():
        if task_instance.state == State.FAILED:
            failed_tasks.append(task_instance.task_id)
    
    # Push DAG state and failed task list to XCom
    ti.xcom_push(key='dag_state', value=dag_run.get_state())
    ti.xcom_push(key='failed_tasks', value=failed_tasks)

with DAG(
    dag_id="example_task_group",
    schedule=None,
    start_date=days_ago(1),
    catchup=False,
    tags=["example"],
) as dag:

    with TaskGroup("main_flow", tooltip="Main workflow sequence") as main_flow:
        start = EmptyOperator(task_id="start")

        with TaskGroup("section_1", tooltip="Tasks for section_1") as section_1:
            task_1 = EmptyOperator(task_id="task_1")
            task_2 = BashOperator(task_id="task_2", bash_command="echo 1")
            task_3 = EmptyOperator(task_id="task_3")
            task_1 >> [task_2, task_3]

        with TaskGroup("section_2", tooltip="Tasks for section_2 (designed to fail)") as section_2:
            fail_task_1 = BashOperator(task_id="fail_task_1", bash_command="exit 1")
            fail_task_2 = BashOperator(task_id="fail_task_2", bash_command="exit 1")
            fail_task_1 >> fail_task_2

        with TaskGroup("section_3", tooltip="Tasks for section_3") as section_3:
            task_1 = EmptyOperator(task_id="task_1")

            with TaskGroup("inner_section_3", tooltip="Tasks for inner_section3") as inner_section_3:
                task_2 = BashOperator(task_id="task_2", bash_command="echo 1")
                task_3 = EmptyOperator(task_id="task_3")
                task_4 = EmptyOperator(task_id="task_4")
                [task_2, task_3] >> task_4

        # 모든 작업이 완료되면 end task로 연결되도록 설정
        end = EmptyOperator(task_id="end")
        start >> section_1 >> section_2 >> section_3 >> end

    with TaskGroup("notification_flow", tooltip="Notification and Failure Collection") as notification_flow:
        collect_failed = PythonOperator(
            task_id="collect_failed_tasks",
            python_callable=collect_failed_tasks,
            provide_context=True,
            trigger_rule="all_done"
        )

        notify_email = EmailOperator(
            task_id="send_email",
            to="yby987654@gmail.com",
            subject="Workflow Execution Report : {{ dag.dag_id }} ",
            html_content="""<h3>Workflow Execution Complete</h3>
            <p><strong>Workflow ID:</strong> {{ dag.dag_id }}</p>
            <p><strong>Workflow Run ID:</strong> {{ run_id }}</p>
            <p><strong>Workflow 상태:</strong> {{ ti.xcom_pull(task_ids='collect_failed_tasks', key='dag_state') }}</p>
            
            {% set failed_tasks = ti.xcom_pull(task_ids='collect_failed_tasks', key='failed_tasks', default=[]) %}
            
            {% if failed_tasks %}
            <p><strong>실패한 Tasks:</strong> {{ failed_tasks }}</p>
            <p><strong>Task Try Number:</strong> {{ ti.try_number }}</p>
            {% else %}
            <p>모든 Tasks가 성공적으로 완료되었습니다.</p>
            {% endif %}""",
            trigger_rule="all_done"
        )

        collect_failed >> notify_email

    # TaskGroup 순서 설정
    main_flow >> notification_flow
