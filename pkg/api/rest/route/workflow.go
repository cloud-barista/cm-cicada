package route

import (
	"strings"

	"github.com/cloud-barista/cm-cicada/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/controller/workflow"
	"github.com/labstack/echo/v4"
)

func Workflow(e *echo.Echo) {
	// Basic CRUD operations
	e.POST("/"+strings.ToLower(common.ShortModuleName)+"/workflow", workflow.CreateWorkflow)
	e.GET("/"+strings.ToLower(common.ShortModuleName)+"/workflow/:wfId", workflow.GetWorkflow)
	e.GET("/"+strings.ToLower(common.ShortModuleName)+"/workflow/name/:wfName", workflow.GetWorkflowByName)
	e.GET("/"+strings.ToLower(common.ShortModuleName)+"/workflow", workflow.ListWorkflow)
	e.PUT("/"+strings.ToLower(common.ShortModuleName)+"/workflow/:wfId", workflow.UpdateWorkflow)
	e.DELETE("/"+strings.ToLower(common.ShortModuleName)+"/workflow/:wfId", workflow.DeleteWorkflow)

	// Workflow execution and monitoring
	e.POST("/"+strings.ToLower(common.ShortModuleName)+"/workflow/:wfId/run", workflow.RunWorkflow)
	e.GET("/"+strings.ToLower(common.ShortModuleName)+"/workflow/:wfId/runs", workflow.GetWorkflowRuns)
	e.GET("/"+strings.ToLower(common.ShortModuleName)+"/workflow/:wfId/status", workflow.GetWorkflowStatus)
	e.GET("/"+strings.ToLower(common.ShortModuleName)+"/workflow/:wfId/eventlogs", workflow.GetEventLogs)
	e.GET("/"+strings.ToLower(common.ShortModuleName)+"/importErrors", workflow.GetImportErrors)

	// Task group operations
	e.GET("/"+strings.ToLower(common.ShortModuleName)+"/workflow/:wfId/task_group", workflow.ListTaskGroup)
	e.GET("/"+strings.ToLower(common.ShortModuleName)+"/workflow/:wfId/task_group/:tgId", workflow.GetTaskGroup)
	e.GET("/"+strings.ToLower(common.ShortModuleName)+"/workflow/:wfId/task_group/:tgId/task", workflow.ListTaskFromTaskGroup)
	e.GET("/"+strings.ToLower(common.ShortModuleName)+"/workflow/:wfId/task_group/:tgId/task/:taskId", workflow.GetTaskFromTaskGroup)
	e.GET("/"+strings.ToLower(common.ShortModuleName)+"/task_group/:tgId", workflow.GetTaskGroupDirectly)

	// Task operations
	e.GET("/"+strings.ToLower(common.ShortModuleName)+"/workflow/:wfId/task", workflow.ListTask)
	e.GET("/"+strings.ToLower(common.ShortModuleName)+"/workflow/:wfId/task/:taskId", workflow.GetTask)
	e.GET("/"+strings.ToLower(common.ShortModuleName)+"/task/:taskId", workflow.GetTaskDirectly)
	e.GET("/"+strings.ToLower(common.ShortModuleName)+"/workflow/:wfId/workflowRun/:wfRunId/task/:taskId/taskTryNum/:taskTyNum/logs", workflow.GetTaskLogs)
	e.GET("/"+strings.ToLower(common.ShortModuleName)+"/workflow/:wfId/workflowRun/:wfRunId/task/:taskId/taskTryNum/:taskTyNum/logs/download", workflow.GetTaskLogDownload)
	e.GET("/"+strings.ToLower(common.ShortModuleName)+"/workflow/:wfId/workflowRun/:wfRunId/taskInstances", workflow.GetTaskInstances)
	e.POST("/"+strings.ToLower(common.ShortModuleName)+"/workflow/:wfId/workflowRun/:wfRunId/range", workflow.ClearTaskInstances)

	// Version management
	e.GET("/"+strings.ToLower(common.ShortModuleName)+"/workflow/:wfId/version", workflow.ListWorkflowVersion)
	e.GET("/"+strings.ToLower(common.ShortModuleName)+"/workflow/:wfId/version/:verId", workflow.GetWorkflowVersion)
}
