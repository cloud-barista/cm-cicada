package db

import (
	"encoding/json"

	"github.com/cloud-barista/cm-cicada/lib/config"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/google/uuid"
	"os"
	"path/filepath"
)

func taskComponentGetByName(name string) *model.TaskComponent {
	taskComponent := &model.TaskComponent{}

	result := DB.Where("name = ?", name).First(taskComponent)
	err := result.Error
	if err != nil {
		return nil
	}

	return taskComponent
}

func TaskComponentInit() error {
	// JSON 파일이 위치한 디렉토리
	jsonDir := config.CMCicadaConfig.CMCicada.TaskComponent.ExamplesDirectory

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
		var taskComponent model.TaskComponent
		decoder := json.NewDecoder(jsonFile)
		err = decoder.Decode(&taskComponent)
		if err != nil {
			return err
		}

		previous := taskComponentGetByName(taskComponent.Name)
		if previous != nil {
			taskComponent.ID = previous.ID
		} else {
			taskComponent.ID = uuid.New().String()
		}

		// 삽입
		err = DB.Save(&taskComponent).Error
		if err != nil {
			return err
		}
	}

	return nil
}
