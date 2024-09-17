from typing import Any
from airflow.utils.decorators import apply_defaults
from airflow.models.baseoperator import BaseOperator
from airflow.hooks.http_hook import HttpHook

import json

def replaceRight(original, old, new, count_right):
    repeat=0
    text = original
    old_len = len(old)

    count_find = original.count(old)
    if (count_right == -1) or (count_right > count_find) :
        repeat = count_find
    else :
        repeat = count_right

    while(repeat):
      find_index = text.rfind(old)
      text = text[:find_index] + new + text[find_index+old_len:]

      repeat -= 1

    return text

def execute_http(context, http_conn_id: str, method: str, endpoint: str, data: str) -> None:
    http_hook = HttpHook(method, http_conn_id)
    response = http_hook.run(endpoint, data, headers={'Content-Type': 'application/json'})
    response = json.dumps(response.json())
    print("=== Response ===")
    print(response)
    context['ti'].xcom_push(key='return_value', value=response)


class JsonHttpRequestOperator(BaseOperator):
    @apply_defaults
    def __init__(self, http_conn_id: str, method: str, endpoint: str, xcom_task: str, *args, **kwargs) -> None:
        self.http_conn_id = http_conn_id
        self.method = method
        self.endpoint = endpoint
        self.xcom_task = xcom_task
        self.args = args
        self.kwargs = kwargs
        super(JsonHttpRequestOperator, self).__init__(*args, **kwargs)

    def execute(self, context) -> None:
        xcom_data = context['ti'].xcom_pull(task_ids=[self.xcom_task], key='return_value')
        data = ""

        if xcom_data and len(xcom_data) > 0:
            data = str(xcom_data[0])
            print(f"=== xcom data (task_id='{self.xcom_task}', key='return_value') ===")
            print(data)
        else:
            raise ValueError(f"No xcom data found for task_id='{self.xcom_task}', key='return_value'")

        print('=== xcom data (task_id=\'' + self.xcom_task + '\', key=\'return_value\') ===')
        print(data)
        data = data.replace('\\n', '')

        if '[\'{' in data:
            data = data.replace('[\'{', '{')
        else:
            data = data.replace('[{', '{', 1)

        if '}\']' in data:
            data = replaceRight(data, '}\']', '}', -1)
        else:
            data = replaceRight(data, '}]', '}', 1)

        print("=== Request Body ===")
        print(data)
        execute_http(context, self.http_conn_id, self.method, self.endpoint, data)
