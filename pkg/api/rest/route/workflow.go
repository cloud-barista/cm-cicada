package route

import (
	"strings"

	"github.com/cloud-barista/cm-cicada/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/controller"
	"github.com/labstack/echo/v4"
)

func Workflow(e *echo.Echo) {
	e.POST("/"+strings.ToLower(common.ShortModuleName)+"/workflow", controller.CreateWorkflow)
	e.GET("/"+strings.ToLower(common.ShortModuleName)+"/workflow/:wfId", controller.GetWorkflow)
	e.GET("/"+strings.ToLower(common.ShortModuleName)+"/workflow/name/:wfName", controller.GetWorkflowByName)
	e.GET("/"+strings.ToLower(common.ShortModuleName)+"/workflow", controller.ListWorkflow)
	e.PUT("/"+strings.ToLower(common.ShortModuleName)+"/workflow/:wfId", controller.UpdateWorkflow)
	e.POST("/"+strings.ToLower(common.ShortModuleName)+"/workflow/:wfId/run", controller.RunWorkflow)
	e.DELETE("/"+strings.ToLower(common.ShortModuleName)+"/workflow/:wfId", controller.DeleteWorkflow)

	e.GET("/"+strings.ToLower(common.ShortModuleName)+"/workflow/:wfId/task_group", controller.ListTaskGroup)
	e.GET("/"+strings.ToLower(common.ShortModuleName)+"/workflow/:wfId/task_group/:tgId", controller.GetTaskGroup)
	e.GET("/"+strings.ToLower(common.ShortModuleName)+"/workflow/:wfId/task_group/:tgId/task", controller.ListTaskFromTaskGroup)
	e.GET("/"+strings.ToLower(common.ShortModuleName)+"/workflow/:wfId/task_group/:tgId/task/:taskId", controller.GetTaskFromTaskGroup)

	e.GET("/"+strings.ToLower(common.ShortModuleName)+"/workflow/:wfId/task", controller.ListTask)

	e.GET("/"+strings.ToLower(common.ShortModuleName)+"/workflow/:wfId/task/:taskId", controller.GetTask)

	e.GET("/"+strings.ToLower(common.ShortModuleName)+"/task_group/:tgId", controller.GetTaskGroupDirectly)
	e.GET("/"+strings.ToLower(common.ShortModuleName)+"/task/:taskId", controller.GetTaskDirectly)
	e.GET("/"+strings.ToLower(common.ShortModuleName)+"/workflow/:wfId/workflowRun/:wfRunId/task/:taskId/taskTryNum/:taskTyNum/logs", controller.GetTaskLogs)
	e.GET("/"+strings.ToLower(common.ShortModuleName)+"/workflow/:wfId/runs", controller.GetWorkflowRuns)
}
