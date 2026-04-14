package db

import (
	"fmt"
	"strconv"

	"github.com/cloud-barista/cm-cicada/common"
	"github.com/cloud-barista/cm-cicada/lib/config"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/jollaman999/utils/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Open() error {
	var err error
	DB, err = gorm.Open(sqlite.Open(common.RootPath+"/"+common.ModuleName+".db?_journal_mode=WAL&_busy_timeout=10000"), &gorm.Config{})
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

	err = DB.AutoMigrate(&model.TaskSnapshot{})
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

	err = ensureWorkflowActiveNameUniqueIndex()
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

func ensureWorkflowActiveNameUniqueIndex() error {
	// Legacy global unique index blocks re-create of a soft-deleted workflow with same name.
	if err := DB.Exec("DROP INDEX IF EXISTS idx_workflows_name").Error; err != nil {
		return err
	}
	if err := DB.Exec("DROP INDEX IF EXISTS idx_workflows_name_active").Error; err != nil {
		return err
	}

	createSQL := "CREATE UNIQUE INDEX IF NOT EXISTS idx_workflows_name_active " +
		"ON workflows(name COLLATE NOCASE) WHERE is_deleted = 0"
	if err := DB.Exec(createSQL).Error; err != nil {
		return fmt.Errorf("failed to enforce active workflow name uniqueness: %w", err)
	}

	return nil
}

func Close() {
	if DB != nil {
		sqlDB, _ := DB.DB()
		_ = sqlDB.Close()
	}
}

func BeginTransaction() (*gorm.DB, error) {
	tx := DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return tx, nil
}
