package service

import (
	"fmt"

	"github.com/cloud-barista/cm-cicada/dao"
	"github.com/cloud-barista/cm-cicada/lib/airflow"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/util"
	"github.com/google/uuid"
	"github.com/jollaman999/utils/logger"
)

// WorkflowService interface defines the contract for workflow business logic
type WorkflowService interface {
	CreateWorkflow(req model.CreateWorkflowReq) (*model.Workflow, error)
	GetWorkflow(wfId string) (*model.Workflow, error)
	GetWorkflowByName(wfName string) (*model.Workflow, error)
	ListWorkflow(filter *model.Workflow, page, row int) (*[]model.Workflow, error)
	UpdateWorkflow(wfId string, req model.CreateWorkflowReq) (*model.Workflow, error)
	DeleteWorkflow(wfId string) error
	RunWorkflow(wfId string) error
	GetWorkflowRuns(wfId string) ([]model.WorkflowRun, error)
	GetWorkflowStatus(wfId string) ([]model.WorkflowStatus, error)
	ListWorkflowVersion(wfId string, page, row int) (*[]model.WorkflowVersion, error)
	GetWorkflowVersion(wfId, verId string) (*model.WorkflowVersion, error)
}

// workflowService is the concrete implementation of WorkflowService
type workflowService struct {
	airflowClient airflow.Client
}

// NewWorkflowService creates a new instance of WorkflowService
func NewWorkflowService() WorkflowService {
	return &workflowService{}
}

// CreateWorkflow creates a new workflow with all related entities
func (s *workflowService) CreateWorkflow(req model.CreateWorkflowReq) (*model.Workflow, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("please provide the name")
	}

	var specVersion = model.WorkflowSpecVersion_LATEST
	if req.SpecVersion != "" {
		specVersion = req.SpecVersion
	}

	workflowData, err := util.CreateDataReqToData(specVersion, req.Data)
	if err != nil {
		return nil, err
	}

	var workflow model.Workflow
	workflow.ID = uuid.New().String()
	workflow.SpecVersion = specVersion
	workflow.Name = req.Name
	workflow.Data = workflowData

	var success bool
	_, err = dao.WorkflowCreate(&workflow)
	if err != nil {
		return nil, err
	}
	defer func() {
		if !success {
			_ = dao.WorkflowDelete(&workflow)
		}
	}()

	client, err := airflow.GetClient()
	if err != nil {
		return nil, err
	}

	err = client.CreateDAG(&workflow)
	if err != nil {
		return nil, fmt.Errorf("failed to create the workflow: %w", err)
	}

	// Create task groups and tasks in database
	for _, tg := range workflow.Data.TaskGroups {
		_, err = dao.TaskGroupCreate(&model.TaskGroupDBModel{
			ID:         tg.ID,
			Name:       tg.Name,
			WorkflowID: workflow.ID,
		})
		if err != nil {
			return nil, err
		}

		for _, t := range tg.Tasks {
			_, err = dao.TaskCreate(&model.TaskDBModel{
				ID:          t.ID,
				Name:        t.Name,
				WorkflowID:  workflow.ID,
				TaskGroupID: tg.ID,
			})
			if err != nil {
				return nil, err
			}
		}
	}
	success = true

	return &workflow, nil
}

// GetWorkflow retrieves a workflow by ID and enriches it with task group/task IDs
func (s *workflowService) GetWorkflow(wfId string) (*model.Workflow, error) {
	if wfId == "" {
		return nil, fmt.Errorf("please provide the wfId")
	}

	workflow, err := dao.WorkflowGet(wfId)
	if err != nil {
		return nil, err
	}

	// Enrich workflow with task group and task IDs from database
	for i, tg := range workflow.Data.TaskGroups {
		_, err = dao.TaskGroupGetByWorkflowIDAndName(wfId, tg.Name)
		if err != nil {
			logger.Println(logger.ERROR, true, err)
		}

		workflow.Data.TaskGroups[i].ID = tg.ID

		for j, t := range tg.Tasks {
			_, err = dao.TaskGetByWorkflowIDAndName(wfId, tg.Name)
			if err != nil {
				logger.Println(logger.ERROR, true, err)
			}

			workflow.Data.TaskGroups[i].Tasks[j].ID = t.ID
		}
	}

	// Verify workflow exists in Airflow
	client, err := airflow.GetClient()
	if err != nil {
		return nil, err
	}

	_, err = client.GetDAG(wfId)
	if err != nil {
		return nil, fmt.Errorf("failed to get the workflow from the airflow server")
	}

	return workflow, nil
}

// GetWorkflowByName retrieves a workflow by name
func (s *workflowService) GetWorkflowByName(wfName string) (*model.Workflow, error) {
	if wfName == "" {
		return nil, fmt.Errorf("please provide the wfName")
	}

	workflow, err := dao.WorkflowGetByName(wfName)
	if err != nil {
		return nil, err
	}

	// Enrich workflow with task group and task IDs from database
	for i, tg := range workflow.Data.TaskGroups {
		_, err = dao.TaskGroupGetByWorkflowIDAndName(workflow.ID, tg.Name)
		if err != nil {
			logger.Println(logger.ERROR, true, err)
		}

		workflow.Data.TaskGroups[i].ID = tg.ID

		for j, t := range tg.Tasks {
			_, err = dao.TaskGetByWorkflowIDAndName(workflow.ID, tg.Name)
			if err != nil {
				logger.Println(logger.ERROR, true, err)
			}

			workflow.Data.TaskGroups[i].Tasks[j].ID = t.ID
		}
	}

	// Verify workflow exists in Airflow
	client, err := airflow.GetClient()
	if err != nil {
		return nil, err
	}

	_, err = client.GetDAG(workflow.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get the workflow from the airflow server")
	}

	return workflow, nil
}

// ListWorkflow retrieves a list of workflows with pagination
func (s *workflowService) ListWorkflow(filter *model.Workflow, page, row int) (*[]model.Workflow, error) {
	workflows, err := dao.WorkflowGetList(filter, page, row)
	if err != nil {
		return nil, err
	}

	// Enrich each workflow with task group and task IDs
	for i, w := range *workflows {
		for j, tg := range w.Data.TaskGroups {
			_, err = dao.TaskGroupGetByWorkflowIDAndName(w.ID, tg.Name)
			if err != nil {
				logger.Println(logger.ERROR, true, err)
			}

			(*workflows)[i].Data.TaskGroups[j].ID = tg.ID

			for k, t := range tg.Tasks {
				_, err = dao.TaskGetByWorkflowIDAndName(w.ID, tg.Name)
				if err != nil {
					logger.Println(logger.ERROR, true, err)
				}

				(*workflows)[i].Data.TaskGroups[j].Tasks[k].ID = t.ID
			}
		}
	}

	return workflows, nil
}

// UpdateWorkflow updates an existing workflow
func (s *workflowService) UpdateWorkflow(wfId string, req model.CreateWorkflowReq) (*model.Workflow, error) {
	oldWorkflow, err := dao.WorkflowGet(wfId)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		oldWorkflow.Name = req.Name
	}

	var specVersion = model.WorkflowSpecVersion_LATEST
	if req.SpecVersion != "" {
		specVersion = req.SpecVersion
	}

	workflowData, err := util.CreateDataReqToData(specVersion, req.Data)
	if err != nil {
		return nil, err
	}

	// Remove old task groups and tasks from the database
	for _, tg := range oldWorkflow.Data.TaskGroups {
		taskGroup, err := dao.TaskGroupGet(tg.ID)
		if err != nil {
			logger.Println(logger.ERROR, true, err)
		}
		err = dao.TaskGroupDelete(taskGroup)
		if err != nil {
			logger.Println(logger.ERROR, true, err)
		}

		for _, t := range tg.Tasks {
			task, err := dao.TaskGet(t.ID)
			if err != nil {
				logger.Println(logger.ERROR, true, err)
			}
			err = dao.TaskDelete(task)
			if err != nil {
				logger.Println(logger.ERROR, true, err)
			}
		}
	}

	// Create new task groups and tasks in the database
	for _, tg := range workflowData.TaskGroups {
		_, err = dao.TaskGroupCreate(&model.TaskGroupDBModel{
			ID:         tg.ID,
			Name:       tg.Name,
			WorkflowID: wfId,
		})
		if err != nil {
			return nil, err
		}

		for _, t := range tg.Tasks {
			_, err = dao.TaskCreate(&model.TaskDBModel{
				ID:          t.ID,
				Name:        t.Name,
				WorkflowID:  wfId,
				TaskGroupID: tg.ID,
			})
			if err != nil {
				return nil, err
			}
		}
	}

	oldWorkflow.Data = workflowData

	err = dao.WorkflowUpdate(oldWorkflow)
	if err != nil {
		return nil, err
	}

	// Update workflow in Airflow
	client, err := airflow.GetClient()
	if err != nil {
		return nil, err
	}

	err = client.DeleteDAG(oldWorkflow.ID, true)
	if err != nil {
		return nil, fmt.Errorf("failed to update the workflow: %w", err)
	}

	err = client.CreateDAG(oldWorkflow)
	if err != nil {
		return nil, fmt.Errorf("failed to update the workflow: %w", err)
	}

	return oldWorkflow, nil
}

// DeleteWorkflow deletes a workflow and all related entities
func (s *workflowService) DeleteWorkflow(wfId string) error {
	if wfId == "" {
		return fmt.Errorf("please provide the wfId")
	}

	workflow, err := dao.WorkflowGet(wfId)
	if err != nil {
		return err
	}

	// Delete from Airflow first
	client, err := airflow.GetClient()
	if err != nil {
		return err
	}

	err = client.DeleteDAG(workflow.ID, false)
	if err != nil {
		logger.Println(logger.ERROR, true, "AIRFLOW: "+err.Error())
	}

	// Delete task groups and tasks from database
	for _, tg := range workflow.Data.TaskGroups {
		taskGroup, err := dao.TaskGroupGet(tg.ID)
		if err != nil {
			logger.Println(logger.ERROR, true, err)
		}
		err = dao.TaskGroupDelete(taskGroup)
		if err != nil {
			logger.Println(logger.ERROR, true, err)
		}

		for _, t := range tg.Tasks {
			task, err := dao.TaskGet(t.ID)
			if err != nil {
				logger.Println(logger.ERROR, true, err)
			}
			err = dao.TaskDelete(task)
			if err != nil {
				logger.Println(logger.ERROR, true, err)
			}
		}
	}

	// Delete workflow from database
	err = dao.WorkflowDelete(workflow)
	if err != nil {
		return err
	}

	return nil
}

// RunWorkflow executes a workflow in Airflow
func (s *workflowService) RunWorkflow(wfId string) error {
	if wfId == "" {
		return fmt.Errorf("please provide the id")
	}

	workflow, err := dao.WorkflowGet(wfId)
	if err != nil {
		return err
	}

	client, err := airflow.GetClient()
	if err != nil {
		return err
	}

	_, err = client.RunDAG(workflow.ID)
	if err != nil {
		return fmt.Errorf("failed to run the workflow: %w", err)
	}

	return nil
}

// GetWorkflowRuns retrieves the runs for a specific workflow
func (s *workflowService) GetWorkflowRuns(wfId string) ([]model.WorkflowRun, error) {
	if wfId == "" {
		return nil, fmt.Errorf("please provide the wfId")
	}

	client, err := airflow.GetClient()
	if err != nil {
		return nil, err
	}

	runList, err := client.GetDAGRuns(wfId)
	if err != nil {
		return nil, fmt.Errorf("failed to get the workflow runs: %w", err)
	}

	var transformedRuns []model.WorkflowRun

	for _, dagRun := range *runList.DagRuns {
		transformedRun := model.WorkflowRun{
			WorkflowID:             dagRun.DagId,
			WorkflowRunID:          dagRun.GetDagRunId(),
			DataIntervalStart:      dagRun.GetDataIntervalStart(),
			DataIntervalEnd:        dagRun.GetDataIntervalEnd(),
			State:                  string(dagRun.GetState()),
			ExecutionDate:          dagRun.GetExecutionDate(),
			StartDate:              dagRun.GetStartDate(),
			EndDate:                dagRun.GetEndDate(),
			RunType:                dagRun.GetRunType(),
			LastSchedulingDecision: dagRun.GetLastSchedulingDecision(),
			DurationDate:           (dagRun.GetEndDate().Sub(dagRun.GetStartDate()).Seconds()),
		}
		transformedRuns = append(transformedRuns, transformedRun)
	}

	return transformedRuns, nil
}

// GetWorkflowStatus retrieves the status counts for a workflow
func (s *workflowService) GetWorkflowStatus(wfId string) ([]model.WorkflowStatus, error) {
	if wfId == "" {
		return nil, fmt.Errorf("please provide the wfId")
	}

	client, err := airflow.GetClient()
	if err != nil {
		return nil, err
	}

	enumStatus := client.GetAllowedDagStateEnumValues()
	var statusList []model.WorkflowStatus

	for _, v := range enumStatus {
		resp, err := client.GetDagStatus(wfId, string(*v.Ptr()))
		if err != nil {
			logger.Println(logger.ERROR, false,
				"AIRFLOW: Error occurred while getting DAGRuns. (Error: "+err.Error()+").")
		}
		statusList = append(statusList, model.WorkflowStatus{
			State: string(*v.Ptr()),
			Count: len(*resp.DagRuns),
		})
	}

	return statusList, nil
}

// ListWorkflowVersion retrieves a list of workflow versions
func (s *workflowService) ListWorkflowVersion(wfId string, page, row int) (*[]model.WorkflowVersion, error) {
	workflow := &model.WorkflowVersion{
		WorkflowID: wfId,
	}

	workflows, err := dao.WorkflowVersionGetList(workflow, page, row)
	if err != nil {
		return nil, err
	}

	return workflows, nil
}

// GetWorkflowVersion retrieves a specific workflow version
func (s *workflowService) GetWorkflowVersion(wfId, verId string) (*model.WorkflowVersion, error) {
	if wfId == "" {
		return nil, fmt.Errorf("please provide the wfId")
	}
	if verId == "" {
		return nil, fmt.Errorf("please provide the verId")
	}

	workflow, err := dao.WorkflowVersionGet(verId, wfId)
	if err != nil {
		return nil, err
	}

	return workflow, nil
}
