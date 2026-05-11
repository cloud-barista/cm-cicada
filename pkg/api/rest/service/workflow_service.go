package service

import (
	"errors"
	"strconv"

	"github.com/cloud-barista/cm-cicada/dao"
	"github.com/cloud-barista/cm-cicada/lib/airflow"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/mapper"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
	"github.com/google/uuid"
	"github.com/jollaman999/utils/logger"
)

type WorkflowService struct{}

func NewWorkflowService() *WorkflowService {
	return &WorkflowService{}
}

func (s *WorkflowService) CreateWorkflow(req model.CreateWorkflowReq) (*model.Workflow, error) {
	if req.Name == "" {
		return nil, errors.New("please provide the name")
	}

	specVersion := model.WorkflowSpecVersion_LATEST
	if req.SpecVersion != "" {
		specVersion = req.SpecVersion
	}

	sourceType, sourceTemplateID, err := mapper.ResolveCreateSourceType(specVersion, req.Data)
	if err != nil {
		return nil, err
	}

	return s.createWorkflowInternal(req, specVersion, sourceType, sourceTemplateID)
}

// CloneWorkflow creates a brand-new Workflow by deep-copying an existing one
// looked up by srcWfID. The new workflow's name is auto-generated as
// "<source name>_copy" (with _copy_2, _copy_3 ... suffix on collision).
// task_groups / tasks are pulled from the source and all IDs are re-issued.
// Source's runs and snapshots are not copied. The new workflow's first
// snapshot records source_type="clone" and source_template_id=<source
// workflow ID> so the lineage is traceable.
func (s *WorkflowService) CloneWorkflow(srcWfID string) (*model.Workflow, error) {
	if srcWfID == "" {
		return nil, errors.New("please provide the source workflow id")
	}

	src, err := mapper.GetWorkflowFromDB(srcWfID)
	if err != nil {
		return nil, errors.New("source workflow not found: " + err.Error())
	}

	createReq := model.CreateWorkflowReq{
		SpecVersion: src.SpecVersion,
		Name:        nextCloneName(src.Name),
		Data:        mapper.DataToCreateDataReq(src.Data),
	}

	return s.createWorkflowInternal(createReq, src.SpecVersion, "clone", srcWfID)
}

// nextCloneName returns the first non-colliding "<base>_copy" / "<base>_copy_N"
// name. Workflow.name has no DB unique constraint but ambiguous duplicates
// confuse name-based lookups, so we probe until WorkflowGetByName misses.
func nextCloneName(baseName string) string {
	candidate := baseName + "_copy"
	for i := 2; ; i++ {
		if existing, _ := dao.WorkflowGetByName(candidate); existing == nil {
			return candidate
		}
		candidate = baseName + "_copy_" + strconv.Itoa(i)
	}
}

// createWorkflowInternal persists a new Workflow (DB rows + DAG) using the
// provided CreateWorkflowReq. Caller decides sourceType / sourceTemplateID for
// the initial snapshot. Used by both CreateWorkflow and CloneWorkflow.
func (s *WorkflowService) createWorkflowInternal(req model.CreateWorkflowReq, specVersion, sourceType, sourceTemplateID string) (*model.Workflow, error) {
	workflowData, err := mapper.CreateDataReqToData(specVersion, req.Data)
	if err != nil {
		return nil, err
	}

	workflow := &model.Workflow{}
	workflow.ID = uuid.New().String()
	workflow.WorkflowKey = uuid.New().String()
	workflow.SpecVersion = specVersion
	workflow.Name = req.Name
	workflow.Data = workflowData

	var success bool
	_, err = dao.WorkflowCreate(workflow)
	if err != nil {
		return nil, err
	}
	defer func() {
		if !success {
			_ = dao.TaskSoftDeleteByWorkflowID(workflow.ID)
			_ = dao.TaskGroupSoftDeleteByWorkflowID(workflow.ID)
			_ = dao.WorkflowDelete(workflow)
		}
	}()

	for _, tg := range workflow.Data.TaskGroups {
		_, err = dao.TaskGroupCreate(&model.TaskGroupDBModel{
			ID:           tg.ID,
			Name:         tg.Name,
			WorkflowID:   workflow.ID,
			WorkflowKey:  workflow.WorkflowKey,
			TaskGroupKey: tg.ID,
		})
		if err != nil {
			return nil, err
		}

		for _, t := range tg.Tasks {
			_, err = dao.TaskCreate(&model.TaskDBModel{
				ID:           t.ID,
				Name:         t.Name,
				WorkflowID:   workflow.ID,
				WorkflowKey:  workflow.WorkflowKey,
				TaskGroupID:  tg.ID,
				TaskGroupKey: tg.ID,
				TaskKey:      t.ID,
			})
			if err != nil {
				return nil, err
			}
		}
	}

	_, err = dao.WorkflowCreateSnapshot(workflow, "create", sourceType, sourceTemplateID)
	if err != nil {
		return nil, err
	}

	client, err := airflow.GetClient()
	if err != nil {
		return nil, err
	}

	err = client.CreateDAG(workflow)
	if err != nil {
		return nil, errors.New("failed to create the workflow (error: " + err.Error() + ")")
	}
	success = true

	return workflow, nil
}

func (s *WorkflowService) GetWorkflow(wfId string, includeDeleted bool) (*model.Workflow, error) {
	var (
		workflow *model.Workflow
		err      error
	)
	if includeDeleted {
		workflow, err = mapper.GetWorkflowFromDBIncludeDeleted(wfId)
	} else {
		workflow, err = mapper.GetWorkflowFromDB(wfId)
	}
	if err != nil {
		return nil, err
	}

	client, err := airflow.GetClient()
	if err != nil {
		return nil, err
	}

	_, err = client.GetDAG(common.WorkflowDagID(workflow))
	if err != nil {
		return nil, errors.New("failed to get the workflow from the airflow server")
	}

	return workflow, nil
}

func (s *WorkflowService) GetWorkflowByName(wfName string, includeDeleted bool) (*model.Workflow, error) {
	var (
		workflowByName *model.Workflow
		err            error
	)
	if includeDeleted {
		workflowByName, err = dao.WorkflowGetByNameIncludeDeleted(wfName)
	} else {
		workflowByName, err = dao.WorkflowGetByName(wfName)
	}
	if err != nil {
		return nil, err
	}

	return s.GetWorkflow(workflowByName.ID, includeDeleted)
}

func (s *WorkflowService) ListWorkflow(name string, includeDeleted bool, page int, row int) (*[]model.Workflow, error) {
	workflow := &model.Workflow{Name: name}
	if includeDeleted {
		return dao.WorkflowGetListIncludeDeleted(workflow, page, row)
	}
	return dao.WorkflowGetList(workflow, page, row)
}

func (s *WorkflowService) UpdateWorkflow(wfId string, req model.CreateWorkflowReq) (*model.Workflow, error) {
	oldWorkflow, err := dao.WorkflowGet(wfId)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		oldWorkflow.Name = req.Name
	}

	specVersion := model.WorkflowSpecVersion_LATEST
	if req.SpecVersion != "" {
		specVersion = req.SpecVersion
	}

	workflowData, err := mapper.CreateDataReqToData(specVersion, req.Data)
	if err != nil {
		return nil, err
	}

	diff, err := mapper.BuildWorkflowGraphDiff(oldWorkflow, workflowData)
	if err != nil {
		return nil, err
	}

	for _, tg := range diff.TaskGroupsToUpsert {
		taskGroup := tg
		if err := dao.TaskGroupSave(&taskGroup); err != nil {
			return nil, err
		}
	}
	for _, t := range diff.TasksToUpsert {
		task := t
		if err := dao.TaskSave(&task); err != nil {
			return nil, err
		}
	}
	if err := captureSoftDroppedTaskSnapshots(oldWorkflow, diff.TasksToSoftDrop, "update_delete"); err != nil {
		return nil, err
	}
	for _, t := range diff.TasksToSoftDrop {
		task := t
		if err := dao.TaskDelete(&task); err != nil {
			return nil, err
		}
	}
	for _, tg := range diff.TaskGroupsToSoftDrop {
		taskGroup := tg
		if err := dao.TaskGroupDelete(&taskGroup); err != nil {
			return nil, err
		}
	}

	oldWorkflow.SpecVersion = specVersion
	oldWorkflow.Data = diff.WorkflowData

	err = dao.WorkflowUpdate(oldWorkflow)
	if err != nil {
		return nil, err
	}

	_, err = dao.WorkflowCreateSnapshot(oldWorkflow, "update", "modified", "")
	if err != nil {
		return nil, err
	}

	client, err := airflow.GetClient()
	if err != nil {
		return nil, err
	}

	err = client.CreateDAG(oldWorkflow)
	if err != nil {
		return nil, errors.New("failed to update the workflow (error: " + err.Error() + ")")
	}

	return oldWorkflow, nil
}

func (s *WorkflowService) DeleteWorkflow(wfId string) error {
	workflow, err := dao.WorkflowGet(wfId)
	if err != nil {
		return err
	}

	client, err := airflow.GetClient()
	if err != nil {
		return err
	}

	err = client.DeleteDAG(common.WorkflowDagID(workflow), true)
	if err != nil {
		logger.Println(logger.ERROR, true, "AIRFLOW: "+err.Error())
	}

	activeTasks, err := dao.TaskGetListByWorkflowID(workflow.ID, false)
	if err != nil {
		return err
	}
	if err := captureSoftDroppedTaskSnapshots(workflow, activeTasks, "workflow_delete"); err != nil {
		return err
	}

	if err := dao.TaskSoftDeleteByWorkflowID(workflow.ID); err != nil {
		return err
	}
	if err := dao.TaskGroupSoftDeleteByWorkflowID(workflow.ID); err != nil {
		return err
	}

	err = dao.WorkflowDelete(workflow)
	if err != nil {
		return err
	}

	workflow.IsDeleted = true
	_, err = dao.WorkflowCreateSnapshot(workflow, "delete", "custom", "")
	if err != nil {
		return err
	}

	return nil
}

func (s *WorkflowService) RunWorkflow(wfId string) error {
	workflow, err := dao.WorkflowGet(wfId)
	if err != nil {
		return err
	}

	client, err := airflow.GetClient()
	if err != nil {
		return err
	}

	conf := map[string]interface{}{
		"workflow_id":   workflow.ID,
		"workflow_key":  common.WorkflowDagID(workflow),
		"workflow_name": workflow.Name,
	}
	_, err = client.RunDAG(common.WorkflowDagID(workflow), conf)
	if err != nil {
		return err
	}

	return nil
}

func captureSoftDroppedTaskSnapshots(workflow *model.Workflow, droppedTasks []model.TaskDBModel, snapshotType string) error {
	taskMap := workflowTaskByID(workflow)
	for _, taskDB := range droppedTasks {
		rawTask, ok := taskMap[taskDB.ID]
		if !ok {
			rawTask = model.Task{
				ID:            taskDB.ID,
				Name:          taskDB.Name,
				TaskComponent: "",
				Spec:          nil,
				Dependencies:  []string{},
			}
		}
		if rawTask.Dependencies == nil {
			rawTask.Dependencies = []string{}
		}
		if err := dao.TaskSnapshotCreateFromTask(&taskDB, rawTask, snapshotType); err != nil {
			return err
		}
	}
	return nil
}

func (s *WorkflowService) ListWorkflowVersions(wfID string, page, row int) (*[]model.WorkflowVersion, error) {
	filter := &model.WorkflowVersion{WorkflowID: wfID}
	return dao.WorkflowVersionGetList(filter, page, row)
}

func (s *WorkflowService) GetWorkflowVersion(wfID, versionID string) (*model.WorkflowVersion, error) {
	return dao.WorkflowVersionGet(versionID, wfID)
}

func workflowTaskByID(workflow *model.Workflow) map[string]model.Task {
	tasks := make(map[string]model.Task)
	if workflow == nil {
		return tasks
	}
	for _, tg := range workflow.Data.TaskGroups {
		for _, task := range tg.Tasks {
			tasks[task.ID] = task
		}
	}
	return tasks
}
