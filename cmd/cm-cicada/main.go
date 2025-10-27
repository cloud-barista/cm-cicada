package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/cloud-barista/cm-cicada/common"
	"github.com/cloud-barista/cm-cicada/db"
	"github.com/cloud-barista/cm-cicada/lib/airflow"
	"github.com/cloud-barista/cm-cicada/lib/config"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/controller"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/server"
	"github.com/jollaman999/utils/logger"
	"github.com/jollaman999/utils/syscheck"
)

func init() {
	err := syscheck.CheckRoot()
	if err != nil {
		log.Fatalln(err)
	}

	err = config.PrepareConfigs()
	if err != nil {
		log.Fatalln(err)
	}

	logger.Println(logger.DEBUG, false, "Base path:", common.RootPath)
	err = logger.InitLogFile(common.RootPath+"/log", strings.ToLower(common.ModuleName))
	if err != nil {
		log.Panicln(err)
	}

	controller.OkMessage.Message = "API server is not ready"

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
		}()
		server.Init()
	}()

	controller.OkMessage.Message = "Database is not ready"
	err = db.Open()
	if err != nil {
		logger.Panicln(logger.ERROR, true, err.Error())
	}

	controller.OkMessage.Message = "Airflow connection is not ready"
	airflow.Init()

	controller.OkMessage.Message = "CM-Cicada API server is ready"
	controller.IsReady = true

	wg.Wait()
}

func end() {
	db.Close()
	logger.CloseLogFile()
}

func main() {
	// Catch the exit signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		logger.Println(logger.INFO, false, "Exiting "+common.ModuleName+" module...")
		end()
		os.Exit(0)
	}()
}
