package db

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jollaman999/utils/logger"
	"gorm.io/gorm"

	"github.com/cloud-barista/cm-cicada/lib/config"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

type ConfigFile struct {
	Name                string                 `json:"name"`
	Description         string                 `json:"description"`
	APIConnectionID     string                 `json:"api_connection_id"`
	SwaggerYAMLEndpoint string                 `json:"swagger_yaml_endpoint"`
	Endpoint            string                 `json:"endpoint"`
	Extra               map[string]interface{} `json:"extra"`
}

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

func TaskComponentGetByName(name string) *model.TaskComponent {
	taskComponent := &model.TaskComponent{}
	result := DB.Where("name = ?", name).First(taskComponent)
	if result.Error != nil {
		return nil
	}
	return taskComponent
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

func TaskComponentInit() error {
	jsonDir := config.CMCicadaConfig.CMCicada.TaskComponent.ExamplesDirectory

	files, err := filepath.Glob(jsonDir + "*.json")
	if err != nil {
		return fmt.Errorf("failed to read directory: %v", err)
	}

	for _, file := range files {
		configData, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %v", file, err)
		}

		var configFile ConfigFile
		if err := json.Unmarshal(configData, &configFile); err != nil {
			return fmt.Errorf("failed to parse config file %s: %v", file, err)
		}
		var taskComponent *model.TaskComponent
		if configFile.Extra != nil {
			taskComponent = &model.TaskComponent{}
			taskComponent.Data.Options.Extra = configFile.Extra

		} else {
			var connectionFound bool
			var connection model.Connection
			for _, connection = range config.CMCicadaConfig.CMCicada.AirflowServer.Connections {
				if connection.ID == configFile.APIConnectionID {
					connectionFound = true
					break
				}
			}
			if !connectionFound {
				logger.Println(logger.WARN, true, fmt.Sprintf("failed to find connection with ID %s", configFile.APIConnectionID))
				continue
				// return fmt.Errorf("failed to find connection with ID %s", configFile.APIConnectionID)
			}

			spec, err := fetchAndParseYAML(connection, configFile.SwaggerYAMLEndpoint)
			if err != nil {
				logger.Println(logger.WARN, true, fmt.Sprintf("failed to fetch and parse swagger spec: %v", err))
				continue
			}

			endpoint := strings.TrimPrefix(configFile.Endpoint, spec.BasePath)
			taskComponent, err = processEndpoint(connection.ID, spec, endpoint)
			if err != nil {
				logger.Println(logger.WARN, true, fmt.Sprintf("failed to process endpoint: %v", err))
				continue
				// return fmt.Errorf("failed to process endpoint: %v", err)
			}
		}
		taskComponent.Name = configFile.Name
		taskComponent.Description = configFile.Description

		now := time.Now()

		previous := TaskComponentGetByName(taskComponent.Name)
		if previous != nil {
			taskComponent.ID = previous.ID
			taskComponent.CreatedAt = previous.CreatedAt
			taskComponent.UpdatedAt = now
		} else {
			taskComponent.ID = uuid.New().String()
			taskComponent.CreatedAt = now
			taskComponent.UpdatedAt = now
		}

		taskComponent.IsExample = true

		if err := DB.Session(&gorm.Session{SkipHooks: true}).Save(taskComponent).Error; err != nil {
			return fmt.Errorf("failed to save task component: %v", err)
		}
	}

	return nil
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

func getPropertyType(schema SchemaModel) string {
	if schema.Type != "" {
		return schema.Type
	}

	if schema.Ref != "" {
		return "object"
	}

	if len(schema.Properties) > 0 {
		return "object"
	}

	if schema.Items != nil {
		return "array"
	}

	return "object"
}

func convertToPropertyDef(schema SchemaModel) model.PropertyDef {
	property := model.PropertyDef{
		Type:        getPropertyType(schema),
		Required:    schema.Required,
		Properties:  make(map[string]model.PropertyDef),
		Description: schema.Description,
		Default:     schema.Default,
		Enum:        schema.Enum,
		Example:     schema.Example,
	}

	if property.Type == "object" || len(schema.Properties) > 0 {
		for name, prop := range schema.Properties {
			property.Properties[name] = convertToPropertyDef(prop)
		}
	}

	if property.Type == "array" && schema.Items != nil {
		property.Items = &model.PropertyDef{
			Type:        getPropertyType(*schema.Items),
			Properties:  make(map[string]model.PropertyDef),
			Description: schema.Items.Description,
			Default:     schema.Items.Default,
			Enum:        schema.Items.Enum,
			Example:     schema.Items.Example,
		}

		if len(schema.Items.Properties) > 0 {
			for name, prop := range schema.Items.Properties {
				property.Items.Properties[name] = convertToPropertyDef(prop)
			}
		}
	}

	return property
}

func convertSchemaToParams(schema SchemaModel) model.ParameterStructure {
	params := model.ParameterStructure{
		Required:   schema.Required,
		Properties: make(map[string]model.PropertyDef),
	}

	for name, propSchema := range schema.Properties {
		params.Properties[name] = convertToPropertyDef(propSchema)
	}

	return params
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

func processEndpoint(connectionID string, spec *SwaggerSpec, targetEndpoint string) (*model.TaskComponent, error) {
	targetEndpoint = normalizePath(targetEndpoint)
	for path, pathItem := range spec.Paths {
		if normalizePath(path) == targetEndpoint {
			for method, operation := range pathItem {
				taskComponent := &model.TaskComponent{
					Data: model.TaskComponentData{
						Options: model.TaskComponentOptions{
							APIConnectionID: connectionID,
							Endpoint:        joinPaths(spec.BasePath, path),
							Method:          strings.ToUpper(method),
						},
					},
				}

				pathParams := model.ParameterStructure{
					Properties: make(map[string]model.PropertyDef),
				}
				queryParams := model.ParameterStructure{
					Properties: make(map[string]model.PropertyDef),
				}

				for _, param := range operation.Parameters {
					switch param.In {
					case "path":
						if param.Required {
							pathParams.Required = append(pathParams.Required, param.Name)
						}
						pathParams.Properties[param.Name] = model.PropertyDef{
							Type:        param.Type,
							Description: param.Description,
							Default:     param.Default,
							Enum:        param.Enum,
						}
					case "query":
						if param.Required {
							queryParams.Required = append(queryParams.Required, param.Name)
						}
						queryParams.Properties[param.Name] = model.PropertyDef{
							Type:        param.Type,
							Description: param.Description,
							Default:     param.Default,
							Enum:        param.Enum,
						}
					case "body":
						if param.Schema != nil && param.Schema.Ref != "" {
							schema := resolveSchemaRef(param.Schema.Ref, spec.Definitions)
							taskComponent.Data.BodyParams = convertSchemaToParams(schema)

							requestBodyExample := generateRequestBodyExample(schema)
							if requestBodyExample != "" {
								taskComponent.Data.Options.RequestBody = requestBodyExample
							}
						}
					}
				}

				if len(pathParams.Properties) > 0 {
					taskComponent.Data.PathParams = pathParams
				}
				if len(queryParams.Properties) > 0 {
					taskComponent.Data.QueryParams = queryParams
				}

				return taskComponent, nil
			}
		}
	}

	return nil, fmt.Errorf("endpoint not found: %s", targetEndpoint)
}
