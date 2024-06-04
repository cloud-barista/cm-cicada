package db

import (
	"encoding/json"
	"github.com/cloud-barista/cm-cicada/lib/config"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/google/uuid"
	"os"
	"path/filepath"
)

func workflowTemplateGetByName(name string) *model.WorkflowTemplate {
	workflowTemplate := &model.WorkflowTemplate{}

	result := DB.Where("name = ?", name).First(workflowTemplate)
	err := result.Error
	if err != nil {
		return nil
	}

	return workflowTemplate
}

func WorkflowTemplateInit() error {
	// JSON 파일이 위치한 디렉토리
	jsonDir := config.CMCicadaConfig.CMCicada.WorkflowTemplate.TemplatesDirectory

	// JSON 파일 목록 가져오기
	files, err := filepath.Glob(jsonDir + "*.json")
	if err != nil {
		return err
	}

	// 각 JSON 파일에 대해 처리
	for _, file := range files {
		// 파일 열기
		jsonFile, err := os.Open(file)
		if err != nil {
			return err
		}
		defer func() {
			_ = jsonFile.Close()
		}()

		// JSON 파일 파싱하여 데이터베이스에 삽입
		var workflowTemplate model.WorkflowTemplate
		decoder := json.NewDecoder(jsonFile)
		err = decoder.Decode(&workflowTemplate)
		if err != nil {
			return err
		}

		previous := workflowTemplateGetByName(workflowTemplate.Name)
		if previous != nil {
			workflowTemplate.ID = previous.ID
			for i, tg := range workflowTemplate.Data.TaskGroups {
				workflowTemplate.Data.TaskGroups[i].ID = tg.ID
				for j, t := range tg.Tasks {
					workflowTemplate.Data.TaskGroups[i].Tasks[j].ID = t.ID
				}
			}
		} else {
			workflowTemplate.ID = uuid.New().String()
			for i, tg := range workflowTemplate.Data.TaskGroups {
				workflowTemplate.Data.TaskGroups[i].ID = uuid.New().String()
				for j := range tg.Tasks {
					workflowTemplate.Data.TaskGroups[i].Tasks[j].ID = uuid.New().String()
				}
			}
		}

		// 삽입
		err = DB.Save(&workflowTemplate).Error
		if err != nil {
			return err
		}
	}

	return nil
}
