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
func buildTaskComponent(configFile ConfigFile) (*model.TaskComponent, error) {
	if configFile.Extra != nil {
		taskComponent := &model.TaskComponent{}
		taskComponent.Data.Options.Extra = configFile.Extra
		return taskComponent, nil
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

func findConnection(id string) (model.Connection, bool) {
	for _, connection := range config.CMCicadaConfig.CMCicada.AirflowServer.Connections {
		if connection.ID == id {
			return connection, true
		}
	}
	return model.Connection{}, false
}
