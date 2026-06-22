# Task 응답을 다음 Task로 전달하기 (전체 / 특정 항목)

cm-cicada 워크플로에서 어떤 task의 HTTP 응답을 **다음 task의 요청 본문(request body)** 으로
넘길 수 있습니다. 이때 **응답 전체**를 그대로 넘길지, **특정 항목만** 골라 넘길지 선택할 수 있습니다.

이 기능은 `http_xcom` 계열 task(오퍼레이터 `local.JsonHttpRequestOperator`,
예: `beetle_task_infra_migration`, `honeybee_register_target_info_to_source_group`)에서 동작합니다.

---

## 동작 원리

1. 각 task는 실행 후 자신의 응답(JSON)을 Airflow XCom의 `return_value` 키로 저장합니다.
2. `http_xcom` task의 `request_body`에 **업스트림 task 이름**을 적으면, cm-cicada는 해당 task의
   XCom 값을 가져와 이번 요청의 본문으로 사용합니다.
3. `response_path`(JSONPath)를 함께 지정하면, 응답 전체가 아니라 그 **경로에 해당하는 항목만**
   추출해 본문으로 사용합니다.

> 전제: 소스 task의 응답이 **JSON**이어야 합니다(특정 항목 추출 시).

---

## 1) 응답을 통째로 넘기기

`request_body`에 **이전 task 이름**만 지정합니다. `response_path`는 생략합니다.

```json
{
  "name": "infra_migration",
  "task_component": "beetle_task_infra_migration",
  "request_body": "infra_recommend",
  "dependencies": ["infra_recommend"]
}
```

- `infra_recommend` task의 **응답 전체**가 `infra_migration` 요청 본문으로 전달됩니다.

> 예외(하위호환 fallback): `response_path`가 없고 엔드포인트가 `/beetle/migration`으로 시작하면,
> 예전 동작대로 응답에서 `targetInfra` 항목만 자동으로 추출합니다. 명시적으로 `response_path`를
> 지정하면 이 fallback보다 우선합니다.

---

## 2) 특정 항목만 넘기기

`response_path`에 **JSONPath**를 지정합니다. 응답에서 그 경로에 해당하는 값만 추출됩니다.

```json
{
  "name": "infra_migration",
  "task_component": "beetle_task_infra_migration",
  "request_body": "infra_recommend",
  "response_path": "$.targetInfra",
  "dependencies": ["infra_recommend"]
}
```

- `infra_recommend` 응답이 아래와 같다면,

  ```json
  {
    "status": "ok",
    "targetInfra": { "name": "infra01", "nodeGroups": [ ... ] },
    "targetSpecList": [ ... ]
  }
  ```

- `infra_migration` 요청 본문으로는 `targetInfra` 부분만 전달됩니다.

  ```json
  { "name": "infra01", "nodeGroups": [ ... ] }
  ```

### JSONPath 예시

| `response_path` | 의미 | 추출 결과 예 |
|---|---|---|
| `$.targetInfra` | 최상위 `targetInfra` 객체 | `{ "name": "infra01", ... }` |
| `$.targetInfra.name` | 중첩 키 | `"infra01"` |
| `$.items[0].id` | 배열 인덱스 | `1` |
| `$.targetInfra.nodeGroups[*].name` | 다중 매치(배열로 반환) | `["g1", "g2"]` |
| `$.nodeGroups[?(@.name=='g1')]` | 필터(확장 문법) | `{ "name": "g1", ... }` |

추출 규칙:
- 매치 **1건** → 그 값 자체
- 매치 **여러 건** → 값들의 **JSON 배열**
- 매치 **0건** → 오류(task 실패)

---

## 우선순위 정리

`http_xcom` task가 본문을 결정하는 순서:

1. `response_path`가 있으면 → 해당 JSONPath로 추출한 항목
2. 없고 엔드포인트가 `/beetle/migration*`이면 → 응답의 `targetInfra`(하위호환 fallback)
3. 그 외 → 응답 전체

---

## 필드 요약

| 필드 | 필수 | 설명 |
|---|---|---|
| `request_body` | 예 | 본문으로 사용할 **업스트림 task 이름**(XCom 소스) |
| `response_path` | 아니오 | 소스 응답에서 추출할 **JSONPath**(예: `$.targetInfra`). 생략 시 응답 전체 |
| `dependencies` | 권장 | 소스 task가 먼저 실행되도록 의존성에 포함 |

---

## 참고

- 카탈로그 정의: [`conf/task_types.yaml`](../conf/task_types.yaml) 의 `http_xcom` 타입
- 본문 구성 로직: [`lib/airflow/gusty.go`](../lib/airflow/gusty.go) 의 `buildHTTPXcomTaskOptions`
- 추출 구현: [`_airflow/airflow-home/operators/json_http_request_operator.py`](../_airflow/airflow-home/operators/json_http_request_operator.py)
  (JSONPath 처리는 `jsonpath-ng` 사용)
