# Task 응답을 다음 Task로 전달하기 (전체 / 특정 항목 / 조합)

cm-cicada 워크플로에서 어떤 task의 HTTP 응답을 **다음 task의 요청 본문(request body)** 으로
넘길 수 있습니다. 별도의 task 타입이 필요 없고, **`http` task의 `request_body` 값의 형태**만으로
동작이 결정됩니다.

각 task는 실행 후 자신의 응답(JSON)을 Airflow XCom의 `return_value` 키에 저장하고,
`request_body`가 다른 task를 가리키면 cm-cicada가 그 값을 가져와 본문을 구성합니다.

> 전제: 참조되는 소스 task의 응답이 **JSON** 이어야 합니다(특정 항목 추출/조합 시).

---

## `request_body` 형태별 동작

| `request_body` 형태 | 동작 |
|---|---|
| `"<task>"` | 그 task의 **응답 전체**를 본문으로 |
| `"<task>.<jsonpath>"` | 그 task 응답에서 **JSONPath로 뽑은 항목**을 본문으로 |
| `${<task>.<jsonpath>}` 를 포함한 JSON | placeholder를 upstream 결과로 치환해 **본문을 조합** |
| 그 외(리터럴 JSON 등) | **그대로** 본문으로 (일반 HTTP 호출) |

앞의 3가지는 `local.JsonHttpRequestOperator`(upstream XCom을 당김)로, 마지막은 표준
`HttpOperator`로 실행됩니다 — cm-cicada가 자동으로 판별합니다.

---

## 1) 응답을 통째로 넘기기

`request_body`에 **이전 task 이름**만 지정합니다.

```json
{
  "name": "infra_migration",
  "task_component": "beetle_task_infra_migration",
  "spec": { "request_body": "infra_recommend" },
  "dependencies": ["infra_recommend"]
}
```

- `infra_recommend` task의 **응답 전체**가 `infra_migration` 요청 본문으로 전달됩니다.

> 예외(하위호환 fallback): 참조 형태가 `"<task>"`(경로 없음)이고 엔드포인트가 `/beetle/migration`
> 으로 시작하면, 예전 동작대로 응답에서 `targetInfra` 항목만 자동 추출합니다. `"<task>.<jsonpath>"`
> 로 경로를 명시하면 이 fallback보다 우선합니다.

---

## 2) 특정 항목만 넘기기

`request_body`를 **`<task>.<jsonpath>`** 로 지정합니다.

```json
{
  "name": "infra_migration",
  "task_component": "beetle_task_infra_migration",
  "spec": { "request_body": "infra_recommend_get.cloudInfraModel" },
  "dependencies": ["infra_recommend_get"]
}
```

- task 이름 뒤의 경로는 JSONPath로 해석됩니다. `$`가 없으면 앞에 `$.`를 붙여 처리하므로
  `infra_recommend_get.cloudInfraModel` == `infra_recommend_get.$.cloudInfraModel` 입니다.

### JSONPath 예시 (task 이름 뒤 부분)

| `request_body` | 의미 | 추출 결과 예 |
|---|---|---|
| `infra_recommend.targetInfra` | 최상위 `targetInfra` 객체 | `{ "name": "infra01", ... }` |
| `infra_recommend.targetInfra.name` | 중첩 키 | `"infra01"` |
| `infra_migration.$.data.node[0].id` | 배열 인덱스 | `"migrated-...-1"` |
| `infra_recommend.$.targetInfra.nodeGroups[*].name` | 다중 매치(배열로 반환) | `["g1", "g2"]` |

추출 규칙: 매치 1건 → 그 값, 여러 건 → 값들의 JSON 배열, 0건 → 오류(task 실패).

---

## 3) 여러 결과에서 필드만 뽑아 조합하기

정적 필드와 upstream 결과의 특정 값을 섞어 본문을 만들려면, 리터럴 JSON 안에
**`${<task>.<jsonpath>}`** placeholder를 넣습니다.

```json
{
  "name": "install_docker",
  "task_component": "cicada_task_run_script",
  "spec": {
    "request_body": "{\n  \"content\": \"<base64>\",\n  \"ns_id\": \"mig01\",\n  \"infra_id\": \"${infra_migration.$.data.id}\",\n  \"node_id\": \"${infra_migration.$.data.node[0].id}\"\n}"
  },
  "dependencies": ["sleep_for_1m_30s"]
}
```

- 각 `${...}`가 해당 task 결과의 JSONPath 값으로 치환됩니다. 서로 다른 task를 여러 개
  참조할 수도 있습니다.
- 참조하는 task는 (직접이든 간접이든) 먼저 실행되도록 `dependencies`로 순서를 보장하세요.

---

## 넘길 수 있는 항목 확인하기 (`response_schema`)

`<task>.<jsonpath>` 로 무엇을 뽑을 수 있는지는 **그 task의 응답 구조**에 달려 있습니다.
Swagger에서 자동 생성된 `http` task component는 부팅 시 대상 API의 **성공 응답(2xx) 본문
스키마**를 함께 읽어 `spec.response_schema` 에 넣어 두므로, task component를 조회하면
어떤 필드를 참조할 수 있는지 바로 볼 수 있습니다.

```
GET /cicada/task_component/{id}
GET /cicada/task_component/name/{name}
```

```json
{
  "name": "beetle_task_recommend_infra",
  "type": "http",
  "spec": {
    "api_connection_id": "beetle_api",
    "method": "POST",
    "endpoint": "/beetle/api/recommendation/infra",
    "request_body": "...",
    "body_params_schema": { "...": "요청 본문 스키마" },
    "response_schema": {
      "type": "object",
      "properties": {
        "targetInfra": { "type": "object", "properties": { "name": { "type": "string" }, "...": {} } },
        "cloudInfraModel": { "type": "object", "properties": {} }
      }
    }
  }
}
```

- `response_schema.properties` 의 키가 곧 `<task>.<key>` 로 참조 가능한 항목입니다
  (위 예에서는 `<task>.targetInfra`, `<task>.targetInfra.name`, `<task>.cloudInfraModel` 등).
- 응답이 배열이면 `response_schema.type` 이 `array`, 원소 구조는 `items` 에 담깁니다.
  이때는 `<task>.$[*].field` 처럼 JSONPath로 접근합니다.
- Swagger의 `$ref` 는 (중첩 포함) 실제 구조로 풀어서 저장합니다. 응답에 본문 스키마가 없으면
  (예: `204`) `response_schema` 는 생략됩니다.

> Swagger 문서는 부팅 시점에 라이브로 읽으므로, 대상 모듈이 떠 있을 때 최신 스키마로 갱신됩니다.
> 손으로 만든 component에는 `response_schema` 가 없을 수 있습니다.

---

## 우선순위 정리

`http` task가 본문을 결정하는 순서:

1. `request_body`에 `${<task>...}` placeholder가 있으면 → 조합(템플릿) 모드
2. `request_body`가 `"<task>"` / `"<task>.<jsonpath>"` 형태면 → 전체 / 특정 항목
3. 그 외 → 리터럴 본문(표준 HTTP 호출)

---

## 참고

- 타입/스키마: [`conf/task_types.yaml`](../conf/task_types.yaml) 의 `http` 타입 (`request_body`)
- 본문 구성 로직: [`lib/airflow/gusty.go`](../lib/airflow/gusty.go) 의 `buildHTTPTaskOptions`
  (`parseTaskReference` / `hasTemplateRef` / `resolveTemplateXcomTaskIDs`)
- 추출/치환 구현: [`_airflow/airflow-home/operators/json_http_request_operator.py`](../_airflow/airflow-home/operators/json_http_request_operator.py)
  (JSONPath 처리는 `jsonpath-ng` 사용)
- `response_schema` 생성: [`lib/airflow/bootstrap/swagger.go`](../lib/airflow/bootstrap/swagger.go)
  의 `processEndpoint` (`successResponse` / `resolveResponseSchema`)
