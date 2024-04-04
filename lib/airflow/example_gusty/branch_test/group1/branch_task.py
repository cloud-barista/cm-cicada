# ---
# operator: airflow.operators.python.BranchPythonOperator
# python_callable: choose_branch
# dependencies:
#   - start_task
# ---

import random

def choose_branch(**kwargs):
    branches = ['b1', 'b2', 'b3']
    chosen = random.choice(branches)
    print(f'chosen: {chosen}')
    return chosen
