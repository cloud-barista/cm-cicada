# URL + Method 기반 Task Component 자동 생성

cm-cicada는 작은 **descriptor JSON**(API 연결 + 엔드포인트 URL + HTTP method)만 작성하면,
부팅 시점에 대상 모듈의 **Swagger 문서를 읽어** Task Component(=task 템플릿)를 **자동 생성**합니다.
복잡한 요청 본문/파라미터 스키마를 손으로 적지 않아도 되고, 원본 API가 바뀌면 다시 시작할 때
최신 스펙으로 갱신됩니다.

---

## 전체 흐름

부팅 시 `TaskComponentInit()`(`lib/airflow/bootstrap/taskcomponent.go`)이 다음을 수행합니다.

1. `examples_directory`(기본 `./lib/airflow/example/task_component/`)의 `*.json` descriptor를 모두 읽음.
2. descriptor의 `api_connection_id`로 연결 정보를 찾아, `swagger_yaml_endpoint`에서 **Swagger YAML을 HTTP로 가져옴**
   (연결에 `login`/`password`가 있으면 Basic Auth 사용).
3. Swagger의 `paths`에서 descriptor의 **`endpoint` + `method`** 에 해당하는 오퍼레이션을 찾음.
4. 그 오퍼레이션의 파라미터(path/query/body)를 변환해 Task Component의 `spec`을 채움
   (`processEndpoint`, `lib/airflow/bootstrap/swagger.go`).
5. 결과를 **이름(`name`) 기준으로 upsert**(`IsExample=true`)하여 DB에 저장.

> 부팅 시점에 **라이브로** Swagger를 가져오므로, 대상 모듈(tumblebug/beetle 등)이 떠 있어야 자동 생성이 됩니다.
> 가져오지 못하면 해당 descriptor는 경고 로그를 남기고 건너뜁니다(부팅은 계속됨).

---

## 1단계 — API 연결 정의

`conf/cm-cicada.yaml`의 `connections`에 대상 모듈을 등록합니다. `api_connection_id`는 여기의 `id`와 매칭됩니다.

```yaml
connections:
  - id: tumblebug_api          # descriptor의 api_connection_id 와 일치
    type: http
    description: TumbleBug API
    host: localhost
    port: 1323
    schema: http
    login: default
    password: default
```

연결 필드: `id`, `type`, `description`, `host`, `port`, `schema`, `login`, `password`.

---

## 2단계 — Descriptor 작성 (URL + Method)

`examples_directory`에 JSON 파일 하나를 둡니다. 권장 형식(V2, 카탈로그 기반):

```json
{
  "name": "tumblebug_infra_dynamic",
  "description": "Create Infra Dynamically from common spec and image.",
  "type": "http",
  "spec": {
    "api_connection_id": "tumblebug_api",
    "swagger_yaml_endpoint": "/tumblebug/api/doc.yaml",
    "method": "POST",
    "endpoint": "/ns/{nsId}/infraDynamic"
  }
}
```

| 필드 | 설명 |
|---|---|
| `name` | Task Component 식별자(파일명이 아니라 이 값이 DB 키). 같은 이름이면 갱신(upsert) |
| `description` | 설명 |
| `type` | 카탈로그 task 타입(`conf/task_types.yaml`). 보통 `http`(또는 `http_xcom`) |
| `spec.api_connection_id` | 위에서 정의한 연결 `id` |
| `spec.swagger_yaml_endpoint` | 그 모듈의 Swagger YAML 경로(예: `/tumblebug/api/doc.yaml`) |
| `spec.method` | 대상 HTTP method (`GET`/`POST`/...) |
| `spec.endpoint` | 대상 엔드포인트 URL(경로 파라미터는 `{nsId}`처럼) |

> 파일 이름은 자유이며 식별은 `name`으로 합니다.

---

## 자동 생성되는 내용

`endpoint`(BasePath 제거 후) + `method`로 Swagger 오퍼레이션을 찾아 `spec`에 다음을 채웁니다.

- `api_connection_id`, `method`(대문자), `endpoint`(BasePath 결합한 최종 경로)
- **`path_params_schema`** — `in: path` 파라미터별 `{type, required, description, default, enum, example}`
- **`query_params_schema`** — `in: query` 파라미터
- **`body_params_schema`** — `in: body` 스키마(`$ref`를 **재귀적으로 인라인** 해석; 참조 지점의
  description/default/enum/example가 우선 적용)
- **`request_body`** — 위 본문 스키마로부터 생성한 **예시 JSON 문자열**
  (각 필드는 `example` → `enum[0]` → 타입 기본값(string→"string", integer→0, boolean→false …) 순으로 채움)

예시 descriptor → 자동 생성된 Task Component의 실제 결과(`request_body`, `body_params`, `path_params`,
`query_params` 등)는 [`README.md`](../README.md)의 *"Task Component 자동 생성"* 예시를 참고하세요.

---

## 주의사항

- **method 필수 조건**: 같은 `endpoint`에 여러 method가 있는데 `method`를 비워두면 오류가 납니다.
  (어떤 오퍼레이션인지 특정할 수 없으므로 반드시 `method` 지정)
- **엔드포인트 매칭**: `endpoint`에서 Swagger의 `basePath`를 제거한 뒤 `paths`와 비교합니다.
  매칭되는 경로가 없으면 `endpoint not found` 오류.
- **라이브 의존**: Swagger를 부팅 시 가져오므로 대상 모듈이 응답해야 합니다. 못 가져오면 해당 항목만 skip.
- **갱신**: descriptor를 바꾸고 재시작하면 같은 `name`의 컴포넌트가 최신 스펙으로 덮어써집니다.

---

## 다른 형식(참고)

- **V1(legacy)**: `type`/`spec` 없이 최상위에 `api_connection_id`, `swagger_yaml_endpoint`, `endpoint`,
  `method`를 두는 옛 형식. 동일한 Swagger introspection으로 처리됩니다(`legacy/` 디렉토리).
- **`extra`(native operator)**: Swagger를 거치지 않고, `extra.operator`(예: Airflow 기본 오퍼레이터 클래스)와
  나머지 옵션을 그대로 Task Component `spec`으로 사용. 카탈로그에서 operator class로 타입을 매칭합니다.

---

## 관련 소스

- descriptor 로딩/해석: [`lib/airflow/bootstrap/taskcomponent.go`](../lib/airflow/bootstrap/taskcomponent.go)
- Swagger fetch & 엔드포인트 변환: [`lib/airflow/bootstrap/swagger.go`](../lib/airflow/bootstrap/swagger.go) (`processEndpoint`)
- 카탈로그 task 타입: [`conf/task_types.yaml`](../conf/task_types.yaml)
- 연결 설정: [`conf/cm-cicada.yaml`](../conf/cm-cicada.yaml)
- descriptor 예시 모음: [`lib/airflow/example/task_component/`](../lib/airflow/example/task_component/)
