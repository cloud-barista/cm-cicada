package db

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
	"strings"

	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
)

func TaskComponentInit() error {
	// JSON 파일이 위치한 디렉토리
	jsonDir := "lib/airflow/example/task_template/"

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
		defer jsonFile.Close()

		// JSON 파일 파싱하여 데이터베이스에 삽입
		var data model.TaskData
		decoder := json.NewDecoder(jsonFile)
		err = decoder.Decode(&data)
		if err != nil {
			return err
		}

		// CreatedAt 필드를 현재 시간으로 설정
		createdAt := time.Now()

		// 파일명에서 확장자를 제외한 파일명만 추출
		baseName := filepath.Base(file)
		baseNameWithoutExt := strings.TrimSuffix(baseName, filepath.Ext(baseName))

		// WorkflowTemplate 생성
		taskComponent := model.TaskComponent{
			ID:      baseNameWithoutExt, // 파일명으로 설정
			Data:      data,
			CreatedAt: createdAt,
		}

		// 삽입
		err = DB.Create(&taskComponent).Error
		if err != nil {
			return err
		}
	}

	return nil
}
