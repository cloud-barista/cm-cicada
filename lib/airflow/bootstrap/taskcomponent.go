package bootstrap

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jollaman999/utils/logger"
	"gorm.io/gorm"

	"github.com/cloud-barista/cm-cicada/dao"
	"github.com/cloud-barista/cm-cicada/db"
	"github.com/cloud-barista/cm-cicada/lib/airflow/catalog"
	"github.com/cloud-barista/cm-cicada/lib/config"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
)

// TaskComponentInit loads example TaskComponents at module startup.
//
// For each JSON descriptor under
// config.CMCicadaConfig.CMCicada.TaskComponent.ExamplesDirectory the function
// introspects the referenced module's Swagger spec, converts the target
// endpoint's parameter schemas into TaskComponent fields, and upserts the
// result as an example (IsExample=true) TaskComponent.
//
// When the descriptor carries a pre-populated `extra` object the Swagger round
// trip is skipped and the extras become the task component options verbatim.
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

		taskComponent, err := buildTaskComponent(configFile)
		if err != nil {
			logger.Println(logger.WARN, true, err.Error())
			continue
		}
		if taskComponent == nil {
			continue
		}

		taskComponent.Name = configFile.Name
		taskComponent.Description = configFile.Description

		now := time.Now()

		previous := dao.TaskComponentGetByName(taskComponent.Name)
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

		if err := db.DB.Session(&gorm.Session{SkipHooks: true}).Save(taskComponent).Error; err != nil {
			return fmt.Errorf("failed to save task component: %v", err)
		}
	}

	return nil
}

// buildTaskComponent resolves a ConfigFile into an in-memory TaskComponent.
// Returns (nil, nil) when the descriptor references an unknown connection —
// the caller treats that as a soft-skip to keep startup resilient.
//
// Resolution order:
//  1. V2 (`type` set): catalog-based, copy `spec` verbatim after type validation.
//  2. V1 with `extra`: match operator class against catalog, copy remaining
//     fields into Spec.
//  3. V1 Swagger fetch: introspect remote spec (legacy path).
func buildTaskComponent(configFile ConfigFile) (*model.TaskComponent, error) {
	if configFile.Type != "" {
		return buildTaskComponentV2(configFile)
	}

	if configFile.Extra != nil {
		return buildTaskComponentFromExtra(configFile.Extra)
	}

	connection, ok := findConnection(configFile.APIConnectionID)
	if !ok {
		return nil, fmt.Errorf("failed to find connection with ID %s", configFile.APIConnectionID)
	}

	spec, err := fetchAndParseYAML(connection, configFile.SwaggerYAMLEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch and parse swagger spec: %v", err)
	}

	endpoint := strings.TrimPrefix(configFile.Endpoint, spec.BasePath)
	taskComponent, err := processEndpoint(connection.ID, spec, endpoint, configFile.Method)
	if err != nil {
		return nil, fmt.Errorf("failed to process endpoint: %v", err)
	}
	return taskComponent, nil
}

// buildTaskComponentV2 produces a catalog-typed TaskComponent.
//
// When the spec carries `swagger_yaml_endpoint`, the function fetches the
// referenced Swagger document at boot, prefixes the configured `endpoint` with
// the spec's BasePath, and inlines a generated `request_body` plus parameter
// schemas (`path_params_schema`, `query_params_schema`, `body_params_schema`)
// — the same enrichment V1 descriptors received. The enrichment keys
// overwrite their counterparts in the source spec; `swagger_yaml_endpoint` is
// dropped after use because it is only metadata for the boot-time fetch.
func buildTaskComponentV2(configFile ConfigFile) (*model.TaskComponent, error) {
	if !catalog.Has(configFile.Type) {
		return nil, fmt.Errorf("unknown task type in catalog: %s", configFile.Type)
	}

	spec := model.Spec{}
	for k, v := range configFile.Spec {
		spec[k] = v
	}

	swaggerEndpoint, _ := spec["swagger_yaml_endpoint"].(string)
	if swaggerEndpoint != "" {
		apiConnID, _ := spec["api_connection_id"].(string)
		if apiConnID == "" {
			return nil, fmt.Errorf("swagger_yaml_endpoint requires api_connection_id in spec")
		}

		connection, ok := findConnection(apiConnID)
		if !ok {
			return nil, fmt.Errorf("failed to find connection with ID %s", apiConnID)
		}

		swagger, err := fetchAndParseYAML(connection, swaggerEndpoint)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch and parse swagger spec: %v", err)
		}

		endpoint, _ := spec["endpoint"].(string)
		method, _ := spec["method"].(string)
		endpoint = strings.TrimPrefix(endpoint, swagger.BasePath)

		enriched, err := processEndpoint(connection.ID, swagger, endpoint, method)
		if err != nil {
			return nil, fmt.Errorf("failed to process endpoint: %v", err)
		}

		for k, v := range enriched.Spec {
			spec[k] = v
		}
		delete(spec, "swagger_yaml_endpoint")
	}

	return &model.TaskComponent{
		Type: configFile.Type,
		Spec: spec,
	}, nil
}

func buildTaskComponentFromExtra(extra map[string]any) (*model.TaskComponent, error) {
	operatorClass, _ := extra["operator"].(string)
	if operatorClass == "" {
		return nil, fmt.Errorf("extra.operator is missing or not a string")
	}

	typeID, ok := findTypeByOperator(operatorClass)
	if !ok {
		return nil, fmt.Errorf("no catalog task type matches operator class: %s", operatorClass)
	}

	specMap := model.Spec{}
	for k, v := range extra {
		if k == "operator" {
			continue
		}
		specMap[k] = v
	}

	return &model.TaskComponent{
		Type: typeID,
		Spec: specMap,
	}, nil
}

func findTypeByOperator(operatorClass string) (string, bool) {
	for _, t := range catalog.List() {
		if t.OperatorClass == operatorClass {
			return t.ID, true
		}
	}
	return "", false
}

func findConnection(id string) (model.Connection, bool) {
	for _, connection := range config.CMCicadaConfig.CMCicada.AirflowServer.Connections {
		if connection.ID == id {
			return connection, true
		}
	}
	return model.Connection{}, false
}
