from airflow.utils.state import State

def collect_failed_tasks(**context):
    dag_run = context['dag_run']
    failed_tasks = []

    for task_instance in dag_run.get_task_instances():
        if task_instance.state == State.FAILED:
            failed_tasks.append(task_instance.task_id)

    # DAG 상태 및 실패한 작업 목록을 XCom에 푸시
    context['ti'].xcom_push(key='dag_state', value=dag_run.get_state())
    context['ti'].xcom_push(key='failed_tasks', value=failed_tasks)