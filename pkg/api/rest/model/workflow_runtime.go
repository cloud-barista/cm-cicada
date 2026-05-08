package model

import "time"

// WorkflowRun mirrors an Airflow DAG run snapshot returned from the scheduler.
// Not persisted in cm-cicada's DB — these records live in Airflow.
type WorkflowRun struct {
	WorkflowRunID          string                 `json:"workflow_run_id,omitempty"`
	WorkflowID             *string                `json:"workflow_id,omitempty"`
	DagID                  *string                `json:"dag_id,omitempty"`
	LogicalDate            string                 `json:"logical_date,omitempty"`
	ExecutionDate          time.Time              `json:"execution_date,omitempty"`
	StartDate              time.Time              `json:"start_date,omitempty"`
	EndDate                time.Time              `json:"end_date,omitempty"`
	DurationDate           float64                `json:"duration_date,omitempty"`
	DataIntervalStart      time.Time              `json:"data_interval_start,omitempty"`
	DataIntervalEnd        time.Time              `json:"data_interval_end,omitempty"`
	LastSchedulingDecision time.Time              `json:"last_scheduling_decision,omitempty"`
	RunType                string                 `json:"run_type,omitempty"`
	State                  string                 `json:"state,omitempty"`
	ExternalTrigger        *bool                  `json:"external_trigger,omitempty"`
	Conf                   map[string]interface{} `json:"conf,omitempty"`
	Note                   string                 `json:"note,omitempty"`
}

// TaskInstance mirrors an Airflow task instance state.
type TaskInstance struct {
	WorkflowRunID                string    `json:"workflow_run_id,omitempty"`
	WorkflowID                   *string   `json:"workflow_id,omitempty"`
	DagID                        *string   `json:"dag_id,omitempty"`
	IsDeletedTask                bool      `json:"is_deleted_task"`
	TaskID                       string    `json:"task_id,omitempty"`
	TaskName                     string    `json:"task_name,omitempty"`
	State                        string    `json:"state,omitempty"`
	StartDate                    time.Time `json:"start_date,omitempty"`
	EndDate                      time.Time `json:"end_date,omitempty"`
	DurationDate                 float64   `json:"duration_date"`
	ExecutionDate                time.Time `json:"execution_date,omitempty"`
	TryNumber                    int       `json:"try_number"`
	IsSoftwareMigrationTask      bool      `json:"is_software_migration_task"`
	SoftwareMigrationExecutionID string    `json:"software_migration_execution_id,omitempty"`
}

type TaskInstanceReference struct {
	// The task ID.
	TaskId   *string `json:"task_id,omitempty"`
	TaskName string  `json:"task_name,omitempty"`
	// DB workflow ID.
	WorkflowID *string `json:"workflow_id,omitempty"`
	// The DAG ID.
	DagID *string `json:"dag_id,omitempty"`
	// The DAG run ID.
	WorkflowRunID *string `json:"workflow_run_id,omitempty"`
	ExecutionDate *string `json:"execution_date,omitempty"`
}

type TaskLog struct {
	Content string `json:"content,omitempty"`
}

type EventLogs struct {
	EventLogs    []EventLog `json:"event_logs"`
	TotalEntries int        `json:"total_entries"`
}

type EventLog struct {
	WorkflowRunID string    `json:"workflow_run_id"`
	RunID         string    `json:"run_id,omitempty"`
	WorkflowID    string    `json:"workflow_id"`
	TaskID        string    `json:"task_id"`
	TaskName      string    `json:"task_name"`
	IsDeletedTask bool      `json:"is_deleted_task"`
	Event         string    `json:"event,omitempty"`
	When          time.Time `json:"when,omitempty"`
	Extra         string    `json:"extra,omitempty"`
}

type TaskClearOption struct {
	DryRun            bool     `json:"dryRun"`
	TaskIds           []string `json:"taskIds"`
	IncludeDownstream bool     `json:"includeDownstream"`
	//IncludeFuture     bool     `json:"includeFuture"`
	//IncludeParentdag  bool     `json:"includeParentdag"`
	//IncludePast       bool     `json:"includePast"`
	//IncludeSubdags    bool     `json:"includeSubdags"`
	IncludeUpstream bool `json:"includeUpstream"`
	OnlyFailed      bool `json:"onlyFailed"`
	OnlyRunning     bool `json:"onlyRunning"`
	ResetDagRuns    bool `json:"resetDagRuns"`
}

type WorkflowStatus struct {
	State string `json:"state"`
	Count int    `json:"count"`
}
