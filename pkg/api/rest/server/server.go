package server

import (
	"fmt"
	"github.com/cloud-barista/cm-cicada/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/middlewares"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/route"
	"net"
	"strings"

	"github.com/cloud-barista/cm-cicada/lib/config"
	_ "github.com/cloud-barista/cm-cicada/pkg/api/rest/docs" // Cicada Documentation
	"github.com/jollaman999/utils/logger"
	"github.com/labstack/echo/v4"
)

const (
	infoColor   = "\033[1;34m%s\033[0m"
	noticeColor = "\033[1;36m%s\033[0m"
)

const (
	website = " https://github.com/cloud-barista/cm-cicada"
)

func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		logger.Println(logger.ERROR, true, err)
		return ""
	}
	defer func() {
		_ = conn.Close()
	}()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	localIP := strings.Split(localAddr.String(), ":")
	if len(localIP) == 0 {
		logger.Println(logger.ERROR, true, "Failed to get local IP.")
		return ""
	}

	return localIP[0]
}

// @title CM-Cicada REST API
// @version latest
// @description Workflow management module

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath /cicada

func Init() {
	e := echo.New()

	e.Use(middlewares.CustomLogger())

	// Hide Echo Banner
	e.HideBanner = true

	route.TaskComponent(e)
	route.WorkflowTemplate(e)
	route.Workflow(e)
	route.RegisterSwagger(e)
	route.RegisterUtility(e)

	// Display API Docs Dashboard when server starts
	endpoint := getLocalIP() + ":" + config.CMCicadaConfig.CMCicada.Listen.Port
	apiDocsDashboard := " http://" + endpoint + "/" + strings.ToLower(common.ShortModuleName) + "/api/index.html"

	fmt.Println("\n ")
	fmt.Println(" CM-Cicada repository:")
	fmt.Printf(infoColor, website)
	fmt.Println("\n ")
	fmt.Println(" API Docs Dashboard:")
	fmt.Printf(noticeColor, apiDocsDashboard)
	fmt.Println("\n ")

	err := e.Start(":" + config.CMCicadaConfig.CMCicada.Listen.Port)
	logger.Panicln(logger.ERROR, true, err)
}
