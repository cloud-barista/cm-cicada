from typing import Any
from airflow.sdk import BaseOperator
from airflow.providers.http.hooks.http import HttpHook
from jsonpath_ng.ext import parse as jsonpath_parse

import json
import re

# Matches "${<task>.<jsonpath>}" references embedded in a body_template.
TEMPLATE_REF_PATTERN = re.compile(r'\$\{([^}]+)\}')

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


def extract_jsonpath(data: str, xcom_path: str) -> str:
    """Extract item(s) from a JSON string using a JSONPath expression.

    Returns the matched value as a JSON string. A single match is returned as
    that value; multiple matches are returned as a JSON array of the values.
    """
    try:
        obj = json.loads(data)
    except json.JSONDecodeError as e:
        raise ValueError(f"Failed to parse data as JSON: {e}")

    matches = jsonpath_parse(xcom_path).find(obj)
    if not matches:
        raise ValueError(f"JSONPath '{xcom_path}' matched nothing in the source response")

    if len(matches) == 1:
        result = matches[0].value
    else:
        result = [m.value for m in matches]

    return json.dumps(result, indent=4)


def extract_jsonpath_value(data: str, path: str):
    """Return the raw value selected by a JSONPath from a JSON string.

    A single match returns that value; multiple matches return a list of values.
    Unlike extract_jsonpath(), the value is not re-serialized, so a scalar can be
    substituted directly into a surrounding JSON string.
    """
    try:
        obj = json.loads(data)
    except json.JSONDecodeError as e:
        raise ValueError(f"Failed to parse data as JSON: {e}")

    matches = jsonpath_parse(path).find(obj)
    if not matches:
        raise ValueError(f"JSONPath '{path}' matched nothing in the source response")

    if len(matches) == 1:
        return matches[0].value
    return [m.value for m in matches]


class JsonHttpRequestOperator(BaseOperator):
    def __init__(self, http_conn_id: str, method: str, endpoint: str, xcom_task: str = "",
                 xcom_path: str = "", data_template: str = "", xcom_task_ids: dict = None,
                 *args, **kwargs) -> None:
        self.http_conn_id = http_conn_id
        self.method = method
        self.endpoint = endpoint
        self.xcom_task = xcom_task
        self.xcom_path = xcom_path
        self.data_template = data_template
        self.xcom_task_ids = xcom_task_ids or {}
        self.args = args
        self.kwargs = kwargs
        super(JsonHttpRequestOperator, self).__init__(*args, **kwargs)

    def render_body_template(self, context) -> str:
        # Replace each "${<task>.<jsonpath>}" reference with the matching item of
        # that task's result, composing a request body from multiple upstream tasks.
        def replace(match):
            ref = match.group(1).strip()
            task_name, sep, path = ref.partition('.')
            if not sep or not path:
                raise ValueError(
                    f"Invalid template reference '${{{ref}}}': expected '${{<task>.<jsonpath>}}'")
            task_id = self.xcom_task_ids.get(task_name, task_name)
            xcom_data = context['ti'].xcom_pull(task_ids=[task_id], key='return_value')
            if not xcom_data or len(xcom_data) == 0 or xcom_data[0] is None:
                raise ValueError(
                    f"No xcom data found for task '{task_name}' (task_id='{task_id}')")
            value = extract_jsonpath_value(str(xcom_data[0]), path)
            return value if isinstance(value, str) else json.dumps(value)

        return TEMPLATE_REF_PATTERN.sub(replace, self.data_template)

    def execute(self, context) -> None:
        # Template mode: compose the body from "${<task>.<jsonpath>}" references.
        if self.data_template:
            data = self.render_body_template(context)
            print("=== Request Body (template) ===")
            print(data)
            execute_http(context, self.http_conn_id, self.method, self.endpoint, data)
            return

        xcom_data = context['ti'].xcom_pull(task_ids=[self.xcom_task], key='return_value')
        data = ""

        if xcom_data and len(xcom_data) > 0:
            data = str(xcom_data[0])
            print(f"=== xcom data (task_id='{self.xcom_task}', key='return_value') ===")
            print(data)
        else:
            raise ValueError(f"No xcom data found for task_id='{self.xcom_task}', key='return_value'")

        print(f"=== endpoint='{self.endpoint}' ===")
        if self.xcom_path:
            # Use only the item(s) selected by the configured JSONPath.
            data = extract_jsonpath(data, self.xcom_path)
            print(f"=== extracted by JSONPath '{self.xcom_path}' ===")
            print(data)
        elif self.endpoint.startswith('/beetle/migration'):
            # Backward-compatible fallback: beetle migration consumes only targetInfra.
            try:
                json_data = json.loads(data)

                if 'targetInfra' in json_data:
                    data = json.dumps(json_data['targetInfra'], indent=4)
                    print("=== targetInfra content ===")
                    print(data)
                else:
                    raise ValueError("targetInfra key not found in the JSON data")

            except json.JSONDecodeError as e:
                raise ValueError(f"Failed to parse data as JSON: {e}")

        print("=== Request Body ===")
        print(data)
        execute_http(context, self.http_conn_id, self.method, self.endpoint, data)
