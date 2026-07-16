package client

import "time"

// Types mirror the Airflow 3.3.0 v2 schemas, limited to the fields cm-cicada
// reads. Nullable fields are pointers so "absent" stays distinguishable from
// the zero value.
//
// The biggest rename from the Airflow 2 API: execution_date no longer exists
// anywhere; it is logical_date, and it is nullable.

// DagRunStates are the states a DAG run can be in (schema: DagRunState).
// Airflow 3 has exactly these four.
var DagRunStates = []string{"queued", "running", "success", "failed"}

// DAG is a subset of DAGResponse.
type DAG struct {
	DagID           string   `json:"dag_id"`
	DagDisplayName  string   `json:"dag_display_name"`
	IsPaused        bool     `json:"is_paused"`
	IsStale         bool     `json:"is_stale"`
	Description     *string  `json:"description"`
	Fileloc         string   `json:"fileloc"`
	RelativeFileloc string   `json:"relative_fileloc"`
	BundleName      *string  `json:"bundle_name"`
	Tags            []DAGTag `json:"tags"`
	HasImportErrors bool     `json:"has_import_errors"`
	Owners          []string `json:"owners"`
}

type DAGTag struct {
	Name  string `json:"name"`
	DagID string `json:"dag_id"`
}

// DAGRun is a subset of DAGRunResponse.
type DAGRun struct {
	DagRunID               string                 `json:"dag_run_id"`
	DagID                  string                 `json:"dag_id"`
	LogicalDate            *time.Time             `json:"logical_date"`
	QueuedAt               *time.Time             `json:"queued_at"`
	StartDate              *time.Time             `json:"start_date"`
	EndDate                *time.Time             `json:"end_date"`
	Duration               *float64               `json:"duration"`
	DataIntervalStart      *time.Time             `json:"data_interval_start"`
	DataIntervalEnd        *time.Time             `json:"data_interval_end"`
	RunAfter               *time.Time             `json:"run_after"`
	LastSchedulingDecision *time.Time             `json:"last_scheduling_decision"`
	RunType                string                 `json:"run_type"`
	State                  string                 `json:"state"`
	TriggeredBy            *string                `json:"triggered_by"`
	Conf                   map[string]interface{} `json:"conf"`
	Note                   *string                `json:"note"`
}

type DAGRunCollection struct {
	DagRuns      []DAGRun `json:"dag_runs"`
	TotalEntries int      `json:"total_entries"`
}

// TriggerDAGRunRequest is the POST /dags/{dag_id}/dagRuns body.
//
// LogicalDate is a required key that accepts null: omitting it entirely is a
// 422. A nil pointer marshals to null, which is what a manual trigger wants.
type TriggerDAGRunRequest struct {
	LogicalDate *time.Time             `json:"logical_date"`
	Conf        map[string]interface{} `json:"conf,omitempty"`
	Note        *string                `json:"note,omitempty"`
}

// TaskInstance is a subset of TaskInstanceResponse.
type TaskInstance struct {
	ID              string     `json:"id"`
	TaskID          string     `json:"task_id"`
	DagID           string     `json:"dag_id"`
	DagRunID        string     `json:"dag_run_id"`
	MapIndex        int        `json:"map_index"`
	LogicalDate     *time.Time `json:"logical_date"`
	RunAfter        *time.Time `json:"run_after"`
	StartDate       *time.Time `json:"start_date"`
	EndDate         *time.Time `json:"end_date"`
	Duration        *float64   `json:"duration"`
	State           *string    `json:"state"`
	TryNumber       int        `json:"try_number"`
	MaxTries        int        `json:"max_tries"`
	TaskDisplayName string     `json:"task_display_name"`
	Operator        *string    `json:"operator"`
	Note            *string    `json:"note"`
}

type TaskInstanceCollection struct {
	TaskInstances []TaskInstance `json:"task_instances"`
	TotalEntries  int            `json:"total_entries"`
}

// ClearTaskInstancesRequest is the POST /dags/{dag_id}/clearTaskInstances body.
//
// Airflow 3 dropped include_subdags and include_parentdag along with SubDags.
type ClearTaskInstancesRequest struct {
	DryRun            bool     `json:"dry_run"`
	TaskIDs           []string `json:"task_ids,omitempty"`
	DagRunID          *string  `json:"dag_run_id,omitempty"`
	IncludeUpstream   bool     `json:"include_upstream"`
	IncludeDownstream bool     `json:"include_downstream"`
	IncludeFuture     bool     `json:"include_future"`
	IncludePast       bool     `json:"include_past"`
	OnlyFailed        bool     `json:"only_failed"`
	OnlyRunning       bool     `json:"only_running"`
	ResetDagRuns      bool     `json:"reset_dag_runs"`
}

// StructuredLogMessage is one entry of a task log. Airflow 3 returns logs as
// structured records rather than the single blob the Airflow 2 API returned.
type StructuredLogMessage struct {
	Timestamp *time.Time `json:"timestamp"`
	Event     string     `json:"event"`
}

// XComEntry is a subset of the xcomEntries response.
type XComEntry struct {
	Key       string      `json:"key"`
	Timestamp *time.Time  `json:"timestamp"`
	TaskID    string      `json:"task_id"`
	DagID     string      `json:"dag_id"`
	RunID     string      `json:"run_id"`
	MapIndex  int         `json:"map_index"`
	Value     interface{} `json:"value"`
}

type ImportError struct {
	ImportErrorID int        `json:"import_error_id"`
	Timestamp     *time.Time `json:"timestamp"`
	Filename      string     `json:"filename"`
	BundleName    *string    `json:"bundle_name"`
	StackTrace    string     `json:"stack_trace"`
}

type ImportErrorCollection struct {
	ImportErrors []ImportError `json:"import_errors"`
	TotalEntries int           `json:"total_entries"`
}

type EventLog struct {
	EventLogID  int        `json:"event_log_id"`
	When        time.Time  `json:"when"`
	DagID       *string    `json:"dag_id"`
	TaskID      *string    `json:"task_id"`
	RunID       *string    `json:"run_id"`
	MapIndex    *int       `json:"map_index"`
	TryNumber   *int       `json:"try_number"`
	Event       string     `json:"event"`
	LogicalDate *time.Time `json:"logical_date"`
	Owner       *string    `json:"owner"`
	Extra       *string    `json:"extra"`
}

type EventLogCollection struct {
	EventLogs    []EventLog `json:"event_logs"`
	TotalEntries int        `json:"total_entries"`
}

// Connection mirrors ConnectionResponse/ConnectionBody. Airflow 3 uses one
// shape for both, so there is no separate collection-item type as in v1.
type Connection struct {
	ConnectionID string  `json:"connection_id"`
	ConnType     string  `json:"conn_type"`
	Description  *string `json:"description,omitempty"`
	Host         *string `json:"host,omitempty"`
	Login        *string `json:"login,omitempty"`
	Schema       *string `json:"schema,omitempty"`
	Port         *int32  `json:"port,omitempty"`
	Password     *string `json:"password,omitempty"`
	Extra        *string `json:"extra,omitempty"`
}

type ConnectionCollection struct {
	Connections  []Connection `json:"connections"`
	TotalEntries int          `json:"total_entries"`
}
