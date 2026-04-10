package controller

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cloud-barista/cm-cicada/dao"
	"github.com/cloud-barista/cm-cicada/lib/airflow"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/mapper"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/jollaman999/utils/logger"
	"github.com/labstack/echo/v4"
	"github.com/mitchellh/mapstructure"
)

// GetTaskInstances godoc
//
//	@ID			get-task-instances
//	@Summary	Get taskInstances
//	@Description	Get the task Logs.
//	@Tags	[Workflow]
//	@Accept	json
//	@Produce	json
//	@Param	wfId path string true "DB workflow ID."
//	@Param	wfRunId path string true "DB workflow ID."
//	@Success	200	{object}	model.TaskInstance		"Successfully get the taskInstances."
//	@Failure	400	{object}	common.ErrorResponse	"Sent bad request."
//	@Failure	500	{object}	common.ErrorResponse	"Failed to get the taskInstances."
//	@Router	 /workflow/{wfId}/workflowRun/{wfRunId}/taskInstances [get]
func GetTaskInstances(c echo.Context) error {
	wfId, err := requireParam(c, "wfId", "wfId")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	wfRunId, err := requireParam(c, "wfRunId", "wfRunId")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	workflow, err := mapper.GetWorkflowFromDB(wfId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	client, err := airflow.GetClient()
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	runList, err := client.GetTaskInstances(workflowDagID(workflow), common.UrlDecode(wfRunId))
	if err != nil {
		return common.ReturnErrorMsg(c, "Failed to get the taskInstances: "+err.Error())
	}
	var taskInstances []model.TaskInstance
	layout := time.RFC3339Nano

	for _, taskInstance := range *runList.TaskInstances {
		taskDBInfo, err := dao.TaskGetByWorkflowIDAndTaskKey(workflow.ID, taskInstance.GetTaskId())
		if err != nil {
			taskDBInfo, err = dao.TaskGetByWorkflowKeyAndTaskKey(workflowDagID(workflow), taskInstance.GetTaskId())
		}
		if err != nil {
			taskDBInfo, err = dao.TaskGetByWorkflowIDAndName(workflow.ID, taskInstance.GetTaskId())
		}
		if err != nil {
			taskDBInfo, err = dao.TaskGetByWorkflowIDAndTaskKeyIncludeDeleted(workflow.ID, taskInstance.GetTaskId())
		}
		if err != nil {
			taskDBInfo, err = dao.TaskGetByWorkflowKeyAndTaskKeyIncludeDeleted(workflowDagID(workflow), taskInstance.GetTaskId())
		}
		if err != nil {
			taskDBInfo, err = dao.TaskGetByWorkflowIDAndNameIncludeDeleted(workflow.ID, taskInstance.GetTaskId())
		}
		if err != nil {
			return common.ReturnErrorMsg(c, "Failed to get the taskInstances: "+err.Error())
		}
		taskId := &taskDBInfo.ID
		executionDate, err := time.Parse(layout, taskInstance.GetExecutionDate())
		if err != nil {
			fmt.Println("Error parsing execution date:", err)
			continue
		}
		startDate, err := time.Parse(layout, taskInstance.GetExecutionDate())
		if err != nil {
			fmt.Println("Error parsing start date:", err)
			continue
		}
		endDate, err := time.Parse(layout, taskInstance.GetExecutionDate())
		if err != nil {
			fmt.Println("Error parsing end date:", err)
			continue
		}

		var isSoftwareMigrationTask bool
		var executionID string
		for _, tg := range workflow.Data.TaskGroups {
			for _, task := range tg.Tasks {
				if strings.Contains(task.TaskComponent, "grasshopper") &&
					strings.Contains(task.TaskComponent, "software") &&
					strings.Contains(task.TaskComponent, "migration") &&
					task.ID == *taskId {
					isSoftwareMigrationTask = true

					// software migration task인 경우 xcom에서 execution_id 조회
					xcomData, err := client.GetXComValue(
						taskInstance.GetDagId(),
						taskInstance.GetDagRunId(),
						taskInstance.GetTaskId(),
						"return_value",
					)
					if err != nil {
						logger.Println(logger.WARN, false,
							"Failed to get xcom data for task: "+taskInstance.GetTaskId()+" (Error: "+err.Error()+")")
					} else if xcomData != nil {
						if execID, ok := xcomData["execution_id"].(string); ok {
							executionID = execID
						}
					}
					break
				}
			}
		}

		taskInfo := model.TaskInstance{
			WorkflowID:                   taskInstance.DagId,
			WorkflowRunID:                taskInstance.GetDagRunId(),
			TaskID:                       *taskId,
			TaskName:                     taskDBInfo.Name,
			State:                        string(taskInstance.GetState()),
			ExecutionDate:                executionDate,
			StartDate:                    startDate,
			EndDate:                      endDate,
			DurationDate:                 float64(taskInstance.GetDuration()),
			TryNumber:                    int(taskInstance.GetTryNumber()),
			IsSoftwareMigrationTask:      isSoftwareMigrationTask,
			SoftwareMigrationExecutionID: executionID,
		}
		taskInstances = append(taskInstances, taskInfo)
	}
	return c.JSONPretty(http.StatusOK, taskInstances, " ")
}

// ClearTaskInstances godoc
//
//	@ID			clear-task-instances
//	@Summary	Clear taskInstances
//	@Description	Clear the task Instance.
//	@Tags	[Workflow]
//	@Accept	json
//	@Produce	json
//	@Param	wfId path string true "DB workflow ID."
//	@Param	wfRunId path string true "ID of the wfRunId."
//
// @Param		request body 	model.TaskClearOption true "Workflow content"
// @Success	200	{object}	model.TaskInstanceReference		"Successfully clear the taskInstances."
// @Failure	400	{object}	common.ErrorResponse	"Sent bad request."
// @Failure	500	{object}	common.ErrorResponse	"Failed to clear the taskInstances."
// @Router	 /workflow/{wfId}/workflowRun/{wfRunId}/range [post]
func ClearTaskInstances(c echo.Context) error {
	var taskClearOption model.TaskClearOption

	data, err := common.GetJSONRawBody(c)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Metadata: nil,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			toTimeHookFunc()),
		Result: &taskClearOption,
	})
	if err != nil {
		return err
	}
	err = decoder.Decode(data)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	wfId, err := requireParam(c, "wfId", "wfId")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	wfRunId, err := requireParam(c, "wfRunId", "wfRunId")
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	workflow, err := dao.WorkflowGet(wfId)
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}

	var taskKeyList []string
	for _, taskId := range taskClearOption.TaskIds {
		taskInfo, err := dao.TaskGet(taskId)
		if err != nil {
			return common.ReturnErrorMsg(c, fmt.Sprintf("failed to get task info for ID %s: %v", taskId, err))
		}
		taskKeyList = append(taskKeyList, taskAirflowID(taskInfo))
	}
	taskClearOption.TaskIds = taskKeyList
	if err := common.ValidateTaskClearOptions(taskClearOption); err != nil {
		fmt.Printf("옵션 검증 실패: %v\n", err)
		return common.ReturnErrorMsg(c, err.Error())
	}

	TaskInstanceReferences := make([]model.TaskInstanceReference, 0)
	client, err := airflow.GetClient()
	if err != nil {
		return common.ReturnErrorMsg(c, err.Error())
	}
	clearList, err := client.ClearTaskInstance(workflowDagID(workflow), common.UrlDecode(wfRunId), taskClearOption)
	if err != nil {
		return common.ReturnErrorMsg(c, "Failed to get the taskInstances: "+err.Error())
	}
	logger.Println(logger.DEBUG, false, "clearList 요청 내용 : {} ", &clearList)
	if clearList.TaskInstances == nil || len(*clearList.TaskInstances) == 0 {
		logger.Println(logger.DEBUG, false, "TaskInstances is nil or empty")

	}
	for _, taskInstance := range *clearList.TaskInstances {
		taskDBInfo, err := dao.TaskGetByWorkflowIDAndTaskKey(workflow.ID, taskInstance.GetTaskId())
		if err != nil {
			taskDBInfo, err = dao.TaskGetByWorkflowKeyAndTaskKey(workflowDagID(workflow), taskInstance.GetTaskId())
		}
		if err != nil {
			taskDBInfo, err = dao.TaskGetByWorkflowIDAndName(workflow.ID, taskInstance.GetTaskId())
		}
		if err != nil {
			taskDBInfo, err = dao.TaskGetByWorkflowIDAndTaskKeyIncludeDeleted(workflow.ID, taskInstance.GetTaskId())
		}
		if err != nil {
			taskDBInfo, err = dao.TaskGetByWorkflowKeyAndTaskKeyIncludeDeleted(workflowDagID(workflow), taskInstance.GetTaskId())
		}
		if err != nil {
			taskDBInfo, err = dao.TaskGetByWorkflowIDAndNameIncludeDeleted(workflow.ID, taskInstance.GetTaskId())
		}
		if err != nil {
			return common.ReturnErrorMsg(c, "Failed to get the taskInstances: "+err.Error())
		}
		taskId := &taskDBInfo.ID
		taskInfo := model.TaskInstanceReference{
			WorkflowID:    taskInstance.DagId,
			WorkflowRunID: taskInstance.DagRunId,
			TaskId:        taskId,
			TaskName:      taskDBInfo.Name,
			ExecutionDate: taskInstance.ExecutionDate,
		}
		logger.Println(logger.DEBUG, false, "TaskInstanceReferences  ", TaskInstanceReferences)
		TaskInstanceReferences = append(TaskInstanceReferences, taskInfo)
	}
	logger.Println(logger.DEBUG, false, "TaskInstanceReferences ", TaskInstanceReferences)

	return c.JSONPretty(http.StatusOK, TaskInstanceReferences, " ")
}
