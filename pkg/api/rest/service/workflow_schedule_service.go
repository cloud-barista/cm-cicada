package service

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/cloud-barista/cm-cicada/dao"
	"github.com/cloud-barista/cm-cicada/lib/airflow"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/common"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/mapper"
	"github.com/cloud-barista/cm-cicada/pkg/api/rest/model"
)

type WorkflowScheduleService struct{}

func NewWorkflowScheduleService() *WorkflowScheduleService {
	return &WorkflowScheduleService{}
}

// Schedule registers a scheduled execution. Exactly one of req.RunAt /
// req.Cron must be set:
//   - RunAt: one-shot future execution (Airflow schedule="@once" + start_date).
//   - Cron:  recurring execution (Airflow schedule=<cron>).
//
// Only one active schedule per workflow is allowed regardless of type;
// callers must Cancel the existing one before registering another. After
// persisting the row the DAG metadata is rewritten so Airflow picks up the
// new schedule on its next parse cycle.
func (s *WorkflowScheduleService) Schedule(workflowID string, req model.CreateWorkflowScheduleReq) (*model.WorkflowSchedule, error) {
	if workflowID == "" {
		return nil, errors.New("please provide the workflow id")
	}
	s.syncOverdueOnceSchedule(workflowID)

	hasRunAt := req.RunAt != nil && !req.RunAt.IsZero()
	hasCron := req.Cron != nil && strings.TrimSpace(*req.Cron) != ""
	switch {
	case hasRunAt && hasCron:
		return nil, errors.New("provide exactly one of run_at / cron, not both")
	case !hasRunAt && !hasCron:
		return nil, errors.New("provide one of run_at / cron")
	}

	workflow, err := mapper.GetWorkflowFromDB(workflowID)
	if err != nil {
		return nil, errors.New("workflow not found: " + err.Error())
	}

	existing, err := dao.WorkflowScheduleGetActive(workflowID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("workflow already has an active schedule; cancel it first (schedule_id=" + existing.ID + ")")
	}

	row := &model.WorkflowSchedule{
		ID:         uuid.New().String(),
		WorkflowID: workflowID,
		Status:     model.WorkflowScheduleStatusActive,
	}
	if hasRunAt {
		if !req.RunAt.After(time.Now()) {
			return nil, errors.New("run_at must be in the future")
		}
		runAtUTC := req.RunAt.UTC()
		row.Type = model.WorkflowScheduleTypeOnce
		row.RunAt = &runAtUTC
	} else {
		cron := strings.TrimSpace(*req.Cron)
		row.Type = model.WorkflowScheduleTypeCron
		row.Cron = &cron
	}

	if err := dao.WorkflowScheduleCreate(row); err != nil {
		return nil, err
	}

	if err := s.refreshDAG(workflow); err != nil {
		// rollback the schedule row so DAG and DB stay consistent
		_ = dao.WorkflowScheduleUpdateStatus(row.ID, model.WorkflowScheduleStatusCanceled)
		return nil, err
	}

	return row, nil
}

// Cancel marks the workflow's active schedule as canceled and rewrites the
// DAG metadata so the schedule line is dropped (DAG becomes manual-trigger
// only again). Since at most one active schedule per workflow is allowed,
// identifying the target by workflow id alone is unambiguous.
func (s *WorkflowScheduleService) Cancel(workflowID string) (*model.WorkflowSchedule, error) {
	if workflowID == "" {
		return nil, errors.New("please provide the workflow id")
	}
	s.syncOverdueOnceSchedule(workflowID)

	row, err := dao.WorkflowScheduleGetActive(workflowID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, errors.New("no active schedule for this workflow")
	}

	if err := dao.WorkflowScheduleUpdateStatus(row.ID, model.WorkflowScheduleStatusCanceled); err != nil {
		return nil, err
	}
	row.Status = model.WorkflowScheduleStatusCanceled

	workflow, err := mapper.GetWorkflowFromDB(workflowID)
	if err == nil {
		_ = s.refreshDAG(workflow)
	}

	return row, nil
}

// GetLatest returns the workflow's most recently created schedule row
// regardless of status (active / executed / canceled), or (nil, nil) when
// there's no schedule history. Callers branch on .status to interpret the
// result.
func (s *WorkflowScheduleService) GetLatest(workflowID string) (*model.WorkflowSchedule, error) {
	if workflowID == "" {
		return nil, errors.New("please provide the workflow id")
	}
	s.syncOverdueOnceSchedule(workflowID)
	return dao.WorkflowScheduleGetLatest(workflowID)
}

// syncOverdueOnceSchedule promotes the workflow's active one-shot schedule
// to "executed" when Airflow has already scheduler-triggered it. Matching is
// strict: only DAGRuns with run_type="scheduled" and logical_date within ±1s
// of schedule.run_at count — manual triggers and backfills are ignored so a
// stray POST /run can't flip the schedule.
//
// Cron schedules are intentionally left alone: a cron row represents a
// recurring rule, not a single execution, and stays active until canceled.
//
// Called lazily at every schedule API entry point (POST/GET/DELETE) so the
// system has no background worker. Failures are silent — sync is best-effort
// and never blocks the actual request.
func (s *WorkflowScheduleService) syncOverdueOnceSchedule(workflowID string) {
	row, err := dao.WorkflowScheduleGetActive(workflowID)
	if err != nil || row == nil {
		return
	}
	if row.Type != model.WorkflowScheduleTypeOnce || row.RunAt == nil {
		return
	}
	if row.RunAt.After(time.Now()) {
		return // not due yet
	}

	workflow, err := mapper.GetWorkflowFromDB(workflowID)
	if err != nil {
		return
	}
	client, err := airflow.GetClient()
	if err != nil {
		return
	}
	runs, err := client.GetDAGRuns(common.WorkflowDagID(workflow))
	if err != nil || runs.DagRuns == nil {
		return
	}

	target := row.RunAt.UTC()
	const tolerance = time.Second
	for _, run := range *runs.DagRuns {
		if run.GetRunType() != "scheduled" {
			continue
		}
		ld := run.GetLogicalDate()
		if ld.IsZero() {
			continue
		}
		diff := ld.UTC().Sub(target)
		if diff < 0 {
			diff = -diff
		}
		if diff <= tolerance {
			_ = dao.WorkflowScheduleUpdateStatus(row.ID, model.WorkflowScheduleStatusExecuted)
			return
		}
	}
}

// refreshDAG re-writes the DAG metadata so gusty.writeDAGMetadata picks up
// the latest schedule state from DB.
func (s *WorkflowScheduleService) refreshDAG(workflow *model.Workflow) error {
	client, err := airflow.GetClient()
	if err != nil {
		return err
	}
	if err := client.CreateDAG(workflow); err != nil {
		return errors.New("failed to refresh the DAG metadata (error: " + err.Error() + ")")
	}
	return nil
}
