import os
import airflow
from gusty import create_dag
from airflow.utils.state import State

#####################
## DAG Directories ##
#####################

# point to your dags directory
dag_parent_dir = os.path.join(os.environ['AIRFLOW_HOME'], "dags")

# assumes any subdirectories in the dags directory are Gusty DAGs (with METADATA.yml) (excludes subdirectories like __pycache__)
dag_directories = [os.path.join(dag_parent_dir, name) for name in os.listdir(dag_parent_dir) if os.path.isdir(os.path.join(dag_parent_dir, name)) and not name.endswith('__')]

####################
## DAG Generation ##
####################


for dag_directory in dag_directories:
    dag_id = os.path.basename(dag_directory)
    globals()[dag_id] = create_dag(dag_directory,
                                   tags = ['default', 'tags'],
                                   task_group_defaults={"tooltip": "default tooltip"},
                                   wait_for_defaults={"retries": 10, "check_existence": True},
                                   latest_only=False)

                                   
def collect_failed_tasks(**context):
    dag_run = context['dag_run']
    failed_tasks = []

    for task_instance in dag_run.get_task_instances():
        if task_instance.state == State.FAILED:
            failed_tasks.append(task_instance.task_id)

    # DAG 상태 및 실패한 작업 목록을 XCom에 푸시
    context['ti'].xcom_push(key='dag_state', value=dag_run.get_state())
    context['ti'].xcom_push(key='failed_tasks', value=failed_tasks)
