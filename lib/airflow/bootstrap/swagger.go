package bootstrap

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
)

// ConfigFile describes one task component example JSON entry under
// lib/airflow/example/task_component/.
//
// Two formats are supported in the same directory:
//   - V1 (legacy, kept under legacy/): Swagger introspection — set
//     api_connection_id / swagger_yaml_endpoint / endpoint / method,
//     or extra for native Airflow operators.
//   - V2 (catalog-based): set type + spec directly. Skips Swagger fetch.
//
// Files are distinguished at runtime: when `type` is non-empty the descriptor
// is treated as V2.
type ConfigFile struct {
	Name        string `json:"name"`
	Description string `json:"description"`

	// V1 fields (Swagger fetch path).
	APIConnectionID     string                 `json:"api_connection_id,omitempty"`
	SwaggerYAMLEndpoint string                 `json:"swagger_yaml_endpoint,omitempty"`
	Endpoint            string                 `json:"endpoint,omitempty"`
	Method              string                 `json:"method,omitempty"`
	Extra               map[string]interface{} `json:"extra,omitempty"`

	// V2 fields (catalog-based, no Swagger fetch).
	Type string         `json:"type,omitempty"`
	Spec map[string]any `json:"spec,omitempty"`
}

// SwaggerSpec is a minimal Swagger 2.0 document model tailored for
// TaskComponent introspection — only fields we actually consume are declared.
type SwaggerSpec struct {
	Swagger     string                 `yaml:"swagger"`
	BasePath    string                 `yaml:"basePath"`
	Paths       map[string]PathItem    `yaml:"paths"`
	Definitions map[string]SchemaModel `yaml:"definitions"`
}

type SchemaModel struct {
	Type        string                 `yaml:"type"`
	Required    []string               `yaml:"required"`
	Properties  map[string]SchemaModel `yaml:"properties"`
	Items       *SchemaModel           `yaml:"items,omitempty"`
	Ref         string                 `yaml:"$ref,omitempty"`
	Description string                 `yaml:"description,omitempty"`
	Default     interface{}            `yaml:"default,omitempty"`
	Enum        []string               `yaml:"enum,omitempty"`
	Example     interface{}            `yaml:"example,omitempty"`
}

type PathItem map[string]Operation

type Operation struct {
	OperationID string      `yaml:"operationId"`
	Description string      `yaml:"description"`
	Parameters  []Parameter `yaml:"parameters"`
}

type ParameterSchema struct {
	Ref   string       `yaml:"$ref"`
	Items *SchemaModel `yaml:"items,omitempty"`
	Type  string       `yaml:"type,omitempty"`
}

type Parameter struct {
	Name        string           `yaml:"name"`
	In          string           `yaml:"in"`
	Required    bool             `yaml:"required"`
	Type        string           `yaml:"type"`
	Schema      *ParameterSchema `yaml:"schema,omitempty"`
	Description string           `yaml:"description,omitempty"`
	Default     interface{}      `yaml:"default,omitempty"`
	Enum        []string         `yaml:"enum,omitempty"`
	Example     interface{}      `yaml:"example,omitempty"`
}

func normalizeURL(url string) string {
	re := regexp.MustCompile(`/{2,}`)
	return re.ReplaceAllString(url, "/")
}

func normalizeURLWithProtocol(url string) string {
	parts := strings.SplitN(url, "://", 2)

	if len(parts) == 2 {
		protocol := parts[0]
		path := parts[1]

		normalizedPath := normalizeURL(path)

		return protocol + "://" + normalizedPath
	}

	return normalizeURL(url)
}

// fetchAndParseYAML retrieves a Swagger YAML document from a remote module and
// decodes it into a SwaggerSpec.
func fetchAndParseYAML(connection model.Connection, swaggerYAMLEndpoint string) (*SwaggerSpec, error) {
	url := connection.Schema + "://" + connection.Host + ":" + strconv.Itoa(int(connection.Port)) +
		"/" + swaggerYAMLEndpoint
	url = normalizeURLWithProtocol(url)

	ctx := context.Background()
	client := &http.Client{}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	if connection.Login != "" && connection.Password != "" {
		req.SetBasicAuth(connection.Login, connection.Password)
	}

	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var spec SwaggerSpec
	if err := yaml.Unmarshal(responseBody, &spec); err != nil {
		return nil, err
	}

	return &spec, nil
}

func normalizePath(path string) string {
	path = strings.TrimRight(path, "/")
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return path
}

func joinPaths(basePath, endpoint string) string {
	basePath = strings.TrimRight(basePath, "/")
	endpoint = strings.TrimLeft(endpoint, "/")
	if basePath == "" {
		return "/" + endpoint
	}
	return basePath + "/" + endpoint
}

// resolveSchemaRef recursively inlines $ref references in a schema so callers
// work with a self-contained tree. Per-property overrides (description/default/
// enum/example) from the referencing site take precedence over the referenced
// definition.
func resolveSchemaRef(ref string, definitions map[string]SchemaModel) SchemaModel {
	parts := strings.Split(ref, "/")
	defName := parts[len(parts)-1]
	schema := definitions[defName]

	for name, prop := range schema.Properties {
		if prop.Ref != "" {
			resolved := resolveSchemaRef(prop.Ref, definitions)

			if prop.Description != "" {
				resolved.Description = prop.Description
			}
			if prop.Default != nil {
				resolved.Default = prop.Default
			}
			if len(prop.Enum) > 0 {
				resolved.Enum = prop.Enum
			}
			if prop.Example != nil {
				resolved.Example = prop.Example
			}
			schema.Properties[name] = resolved
		} else if prop.Items != nil && prop.Items.Ref != "" {
			resolvedItem := resolveSchemaRef(prop.Items.Ref, definitions)

			if prop.Items.Description != "" {
				resolvedItem.Description = prop.Items.Description
			}
			if prop.Items.Default != nil {
				resolvedItem.Default = prop.Items.Default
			}
			if len(prop.Items.Enum) > 0 {
				resolvedItem.Enum = prop.Items.Enum
			}
			if prop.Items.Example != nil {
				resolvedItem.Example = prop.Items.Example
			}
			prop.Items = &resolvedItem
			prop.Type = "array"
			schema.Properties[name] = prop
		}
	}

	if schema.Type == "" {
		if len(schema.Properties) > 0 {
			schema.Type = "object"
		} else if schema.Items != nil {
			schema.Type = "array"
		}
	}

	return schema
}

func generateExampleValue(schema SchemaModel) interface{} {
	if schema.Example != nil {
		return schema.Example
	}

	switch schema.Type {
	case "string":
		if len(schema.Enum) > 0 {
			return schema.Enum[0]
		}
		return "string"
	case "integer":
		return 0
	case "number":
		return 0.0
	case "boolean":
		return false
	case "array":
		if schema.Items != nil {
			return []interface{}{generateExampleValue(*schema.Items)}
		}
		return []interface{}{}
	case "object":
		if len(schema.Properties) > 0 {
			obj := make(map[string]interface{})
			for name, prop := range schema.Properties {
				obj[name] = generateExampleValue(prop)
			}
			return obj
		}
		return make(map[string]interface{})
	default:
		return nil
	}
}

func generateRequestBodyExample(schema SchemaModel) string {
	example := generateExampleValue(schema)
	if example == nil {
		return ""
	}

	jsonBytes, err := json.MarshalIndent(example, "", "    ")
	if err != nil {
		return ""
	}

	return string(jsonBytes)
}

// schemaToMap renders a resolved Swagger SchemaModel into a JSON-friendly map.
// Empty fields are omitted so the resulting spec keys stay terse.
func schemaToMap(schema SchemaModel) map[string]any {
	m := map[string]any{}

	typ := schema.Type
	if typ == "" {
		switch {
		case len(schema.Properties) > 0:
			typ = "object"
		case schema.Items != nil:
			typ = "array"
		}
	}
	if typ != "" {
		m["type"] = typ
	}
	if len(schema.Required) > 0 {
		m["required"] = schema.Required
	}
	if len(schema.Properties) > 0 {
		props := map[string]any{}
		for name, p := range schema.Properties {
			props[name] = schemaToMap(p)
		}
		m["properties"] = props
	}
	if schema.Items != nil {
		m["items"] = schemaToMap(*schema.Items)
	}
	if schema.Description != "" {
		m["description"] = schema.Description
	}
	if schema.Default != nil {
		m["default"] = schema.Default
	}
	if len(schema.Enum) > 0 {
		m["enum"] = schema.Enum
	}
	if schema.Example != nil {
		m["example"] = schema.Example
	}
	return m
}

// parameterToMap renders a Swagger path/query parameter into a JSON-friendly
// map. Body parameters are handled separately via schemaToMap.
func parameterToMap(p Parameter) map[string]any {
	m := map[string]any{}
	if p.Type != "" {
		m["type"] = p.Type
	}
	if p.Required {
		m["required"] = true
	}
	if p.Description != "" {
		m["description"] = p.Description
	}
	if p.Default != nil {
		m["default"] = p.Default
	}
	if len(p.Enum) > 0 {
		m["enum"] = p.Enum
	}
	if p.Example != nil {
		m["example"] = p.Example
	}
	return m
}

// processEndpoint walks a SwaggerSpec, locates the target endpoint+method, and
// builds a TaskComponent of type "http" with the resolved endpoint, method,
// connection id, generated request body example, and parameter schemas
// (path_params_schema, query_params_schema, body_params_schema). Schemas are
// stored as raw maps so any swagger metadata (description/default/enum/
// example/required) is preserved without bespoke types.
func processEndpoint(connectionID string, spec *SwaggerSpec, targetEndpoint, targetMethod string) (*model.TaskComponent, error) {
	targetEndpoint = normalizePath(targetEndpoint)
	for path, pathItem := range spec.Paths {
		if normalizePath(path) == targetEndpoint {
			methodFoundCount := 0

			var method string
			var operation Operation

			for method, operation = range pathItem {
				methodFoundCount++

				if targetMethod != "" && strings.EqualFold(targetMethod, method) {
					break
				}
			}

			if targetMethod == "" && methodFoundCount > 1 {
				return nil, fmt.Errorf("multiple methods found with the same endpoint: %s"+
					" (Please specify the method from the task component example JSON file.)", targetEndpoint)
			}

			specMap := model.Spec{
				"api_connection_id": connectionID,
				"method":            strings.ToUpper(method),
				"endpoint":          joinPaths(spec.BasePath, path),
			}

			pathSchema := map[string]any{}
			querySchema := map[string]any{}

			for _, param := range operation.Parameters {
				switch param.In {
				case "path":
					pathSchema[param.Name] = parameterToMap(param)
				case "query":
					querySchema[param.Name] = parameterToMap(param)
				case "body":
					if param.Schema == nil || param.Schema.Ref == "" {
						continue
					}
					bodySchema := resolveSchemaRef(param.Schema.Ref, spec.Definitions)
					specMap["body_params_schema"] = schemaToMap(bodySchema)
					if example := generateRequestBodyExample(bodySchema); example != "" {
						specMap["request_body"] = example
					}
				}
			}

			if len(pathSchema) > 0 {
				specMap["path_params_schema"] = pathSchema
			}
			if len(querySchema) > 0 {
				specMap["query_params_schema"] = querySchema
			}

			return &model.TaskComponent{
				Type: "http",
				Spec: specMap,
			}, nil
		}
	}

	return nil, fmt.Errorf("endpoint not found: %s", targetEndpoint)
}
