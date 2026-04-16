package bootstrap

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/google/uuid"

	"github.com/cloud-barista/cm-cicada/dao"
	"github.com/cloud-barista/cm-cicada/db"
	"github.com/cloud-barista/cm-cicada/lib/config"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
)

// WorkflowTemplateInit loads workflow template JSON descriptors from the
// configured templates directory and upserts them into the DB. Each file is
// decoded directly into model.WorkflowTemplate — no Swagger fetch is involved.
func WorkflowTemplateInit() error {
	jsonDir := config.CMCicadaConfig.CMCicada.WorkflowTemplate.TemplatesDirectory

	files, err := filepath.Glob(jsonDir + "*.json")
	if err != nil {
		return err
	}

	for _, file := range files {
		workflowTemplate, err := decodeWorkflowTemplateFile(file)
		if err != nil {
			return err
		}

		previous := dao.WorkflowTemplateGetByName(workflowTemplate.Name)
		if previous != nil {
			workflowTemplate.ID = previous.ID
		} else {
			workflowTemplate.ID = uuid.New().String()
		}

		if err := db.DB.Save(workflowTemplate).Error; err != nil {
			return err
		}
	}

	return nil
}

func decodeWorkflowTemplateFile(path string) (*model.WorkflowTemplate, error) {
	jsonFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = jsonFile.Close()
	}()

	workflowTemplate := &model.WorkflowTemplate{}
	decoder := json.NewDecoder(jsonFile)
	if err := decoder.Decode(workflowTemplate); err != nil {
		return nil, err
	}
	return workflowTemplate, nil
}
