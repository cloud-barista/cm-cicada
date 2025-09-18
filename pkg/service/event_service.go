package service

import (
	"encoding/json"
	"fmt"

	"github.com/apache/airflow-client-go/airflow"
	"github.com/cloud-barista/cm-cicada/dao"
	airflowLib "github.com/cloud-barista/cm-cicada/lib/airflow"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
)

// EventService interface defines the contract for event and log business logic
type EventService interface {
	GetEventLogs(wfId, wfRunId, taskId string) ([]model.EventLog, error)
	GetImportErrors() (*airflow.ImportErrorCollection, error)
}

// eventService is the concrete implementation of EventService
type eventService struct{}

// NewEventService creates a new instance of EventService
func NewEventService() EventService {
	return &eventService{}
}

// GetEventLogs retrieves event logs for a workflow, optionally filtered by run and task
func (s *eventService) GetEventLogs(wfId, wfRunId, taskId string) ([]model.EventLog, error) {
	if wfId == "" {
		return nil, fmt.Errorf("please provide the wfId")
	}

	var taskName string
	if taskId != "" {
		taskDBInfo, err := dao.TaskGet(taskId)
		if err != nil {
			return nil, fmt.Errorf("failed to get the taskInstances: %w", err)
		}
		taskName = taskDBInfo.Name
	}

	client, err := airflowLib.GetClient()
	if err != nil {
		return nil, err
	}

	logs, err := client.GetEventLogs(wfId, wfRunId, taskName)
	if err != nil {
		return nil, fmt.Errorf("failed to get the taskInstances: %w", err)
	}

	var eventLogs model.EventLogs
	err = json.Unmarshal(logs, &eventLogs)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal event logs: %w", err)
	}

	var logList []model.EventLog
	for _, eventlog := range eventLogs.EventLogs {
		var taskID, runId string

		if eventlog.TaskID != "" {
			taskDBInfo, err := dao.TaskGetByWorkflowIDAndName(wfId, eventlog.TaskID)
			if err != nil {
				return nil, fmt.Errorf("failed to get the taskInstances: %w", err)
			}
			taskID = taskDBInfo.ID
		}

		eventlog.WorkflowID = wfId
		if eventlog.RunID != "" {
			runId = eventlog.RunID
		}

		log := model.EventLog{
			WorkflowID:    eventlog.WorkflowID,
			WorkflowRunID: runId,
			TaskID:        taskID,
			TaskName:      eventlog.TaskID,
			Extra:         eventlog.Extra,
			Event:         eventlog.Event,
			When:          eventlog.When,
		}
		logList = append(logList, log)
	}

	return logList, nil
}

// GetImportErrors retrieves import errors from Airflow
func (s *eventService) GetImportErrors() (*airflow.ImportErrorCollection, error) {
	client, err := airflowLib.GetClient()
	if err != nil {
		return nil, err
	}

	logs, err := client.GetImportErrors()
	if err != nil {
		return nil, fmt.Errorf("failed to get the taskInstances: %w", err)
	}

	return &logs, nil
}
