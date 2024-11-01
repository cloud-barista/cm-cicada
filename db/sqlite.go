package db

import (
	"strconv"

	"github.com/cloud-barista/cm-cicada/common"
	"github.com/cloud-barista/cm-cicada/lib/config"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/glebarez/sqlite"
	"github.com/jollaman999/utils/logger"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Open() error {
	var err error

	DB, err = gorm.Open(sqlite.Open(common.RootPath+"/"+common.ModuleName+".db"), &gorm.Config{})
	if err != nil {
		logger.Panicln(logger.ERROR, true, err)
	}

	err = DB.AutoMigrate(&model.TaskDBModel{})
	if err != nil {
		logger.Panicln(logger.ERROR, true, err)
	}

	err = DB.AutoMigrate(&model.TaskGroupDBModel{})
	if err != nil {
		logger.Panicln(logger.ERROR, true, err)
	}

	err = DB.AutoMigrate(&model.TaskComponent{})
	if err != nil {
		logger.Panicln(logger.ERROR, true, err)
	}

	err = DB.AutoMigrate(&model.WorkflowTemplate{})
	if err != nil {
		logger.Panicln(logger.ERROR, true, err)
	}

	err = DB.AutoMigrate(&model.WorkflowVersion{})
	if err != nil {
		logger.Panicln(logger.ERROR, true, err)
	}
	
	err = DB.AutoMigrate(&model.Workflow{})
	if err != nil {
		logger.Panicln(logger.ERROR, true, err)
	}

	logger.Println(logger.INFO, false, "Loading workflow templates...")
	err = WorkflowTemplateInit()
	if err != nil {
		logger.Println(logger.ERROR, true, err)
	}

	taskComponentLoadExamples, _ := strconv.ParseBool(config.CMCicadaConfig.CMCicada.TaskComponent.LoadExamples)
	if taskComponentLoadExamples {
		logger.Println(logger.INFO, false, "Loading task components...")
		err = TaskComponentInit()
		if err != nil {
			logger.Println(logger.ERROR, true, err)
		}
	}

	return err
}

func Close() {
	if DB != nil {
		sqlDB, _ := DB.DB()
		_ = sqlDB.Close()
	}
}
