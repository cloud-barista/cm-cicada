package bootstrap

import (
	"testing"

	"gopkg.in/yaml.v3"
)

// sampleSwagger is a minimal Swagger 2.0 doc: one endpoint whose 200 response is
// a $ref, plus definitions the response and request bodies point at.
const sampleSwagger = `
basePath: /beetle/api
paths:
  /recommendation/infra:
    post:
      operationId: recommendInfra
      parameters:
        - in: body
          name: body
          schema:
            $ref: '#/definitions/RecommendReq'
      responses:
        "200":
          description: OK
          schema:
            $ref: '#/definitions/RecommendRes'
        "400":
          description: bad request
          schema:
            $ref: '#/definitions/ErrRes'
  /list:
    get:
      operationId: listThings
      responses:
        "200":
          description: OK
          schema:
            type: array
            items:
              $ref: '#/definitions/Thing'
  /noschema:
    get:
      operationId: noSchema
      responses:
        "204":
          description: no content
definitions:
  RecommendReq:
    type: object
    required: [region]
    properties:
      region: { type: string }
  RecommendRes:
    type: object
    properties:
      targetInfra: { type: string }
      cloudInfraModel: { $ref: '#/definitions/Model' }
  Model:
    type: object
    properties:
      id: { type: string }
  Thing:
    type: object
    properties:
      name: { type: string }
  ErrRes:
    type: object
    properties:
      message: { type: string }
`

func parseSample(t *testing.T) *SwaggerSpec {
	t.Helper()
	var spec SwaggerSpec
	if err := yaml.Unmarshal([]byte(sampleSwagger), &spec); err != nil {
		t.Fatalf("unmarshal sample swagger: %v", err)
	}
	return &spec
}

func TestProcessEndpointAddsResponseSchema(t *testing.T) {
	tc, err := processEndpoint("beetle_api", parseSample(t), "/recommendation/infra", "POST")
	if err != nil {
		t.Fatalf("processEndpoint: %v", err)
	}

	raw, ok := tc.Spec["response_schema"]
	if !ok {
		t.Fatal("spec has no response_schema")
	}
	rs, ok := raw.(map[string]any)
	if !ok {
		t.Fatalf("response_schema is %T, want map", raw)
	}
	if rs["type"] != "object" {
		t.Errorf("response_schema.type = %v, want object", rs["type"])
	}
	props, ok := rs["properties"].(map[string]any)
	if !ok {
		t.Fatalf("response_schema.properties missing/wrong type: %T", rs["properties"])
	}
	if _, ok := props["targetInfra"]; !ok {
		t.Errorf("response_schema missing targetInfra; got keys %v", keysOf(props))
	}
	// The nested $ref (cloudInfraModel -> Model) must be resolved, not left as a ref.
	nested, ok := props["cloudInfraModel"].(map[string]any)
	if !ok {
		t.Fatalf("cloudInfraModel not a resolved object: %T", props["cloudInfraModel"])
	}
	if _, ok := nested["properties"]; !ok {
		t.Errorf("cloudInfraModel $ref was not resolved: %v", nested)
	}

	// The success response is the 200, not the 400.
	if _, hasMessage := props["message"]; hasMessage {
		t.Error("response_schema picked the 400 error body instead of the 200")
	}
}

func TestProcessEndpointResponseSchemaArray(t *testing.T) {
	tc, err := processEndpoint("x", parseSample(t), "/list", "GET")
	if err != nil {
		t.Fatalf("processEndpoint: %v", err)
	}
	rs, ok := tc.Spec["response_schema"].(map[string]any)
	if !ok {
		t.Fatalf("response_schema missing/wrong type: %T", tc.Spec["response_schema"])
	}
	if rs["type"] != "array" {
		t.Errorf("response_schema.type = %v, want array", rs["type"])
	}
	items, ok := rs["items"].(map[string]any)
	if !ok {
		t.Fatalf("array response_schema has no items object: %T", rs["items"])
	}
	if _, ok := items["properties"]; !ok {
		t.Errorf("array item $ref not resolved: %v", items)
	}
}

func TestProcessEndpointNoResponseSchema(t *testing.T) {
	tc, err := processEndpoint("x", parseSample(t), "/noschema", "GET")
	if err != nil {
		t.Fatalf("processEndpoint: %v", err)
	}
	if _, ok := tc.Spec["response_schema"]; ok {
		t.Error("response_schema should be absent when the response has no body schema")
	}
}

func keysOf(m map[string]any) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	return ks
}
