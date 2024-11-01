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
from airflow.operators.trigger_dagrun import TriggerDagRunOperator
with DAG(
    dag_id="example_task_group2",
    schedule=None,
    start_date=days_ago(1),
    catchup=False,
    tags=["example"],
) as dag:

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

        trigger = TriggerDagRunOperator(
          trigger_dag_id='monitor_dag',
          task_id='trigger',
          execution_date='{{ execution_date }}',
          wait_for_completion=False,
          conf= {
            "source_dag_id": "{{ dag.dag_id }}",      # 현재 DAG의 dag_id
            "source_dag_run_id": "{{ dag_run.run_id }}"  # 현재 DAG의 dag_run_id
          },
          poke_interval=30,
          reset_dag_run=True,
          # trigger_rule="all_done"
          trigger_rule="one_done"
        )
        start >> section_1 >> section_2 >> section_3 >> end >> trigger


