package model

import "time"

// WorkflowScheduleStatus is the lifecycle state of a WorkflowSchedule row.
// Transitions: active -> canceled (user cancel) | active -> executed
// (Airflow ran it; only applies to type="once" — cron schedules stay active
// until canceled).
type WorkflowScheduleStatus string

const (
	WorkflowScheduleStatusActive   WorkflowScheduleStatus = "active"
	WorkflowScheduleStatusCanceled WorkflowScheduleStatus = "canceled"
	WorkflowScheduleStatusExecuted WorkflowScheduleStatus = "executed"
)

// WorkflowScheduleType discriminates one-shot vs recurring schedules.
type WorkflowScheduleType string

const (
	WorkflowScheduleTypeOnce WorkflowScheduleType = "once"
	WorkflowScheduleTypeCron WorkflowScheduleType = "cron"
)

// WorkflowSchedule records a scheduled execution of a workflow. The actual
// scheduling is delegated to Airflow via DAG metadata — cm-cicada only
// persists the intent so it survives restarts and is queryable.
//
// One of RunAt / Cron is set depending on Type:
//   - Type="once": RunAt set, Cron nil. Airflow gets schedule="@once" +
//     start_date=RunAt + catchup=false.
//   - Type="cron": Cron set, RunAt nil. Airflow gets schedule=<cron> +
//     catchup=false.
//
// At most one row per workflow_id is in active state at a time, regardless
// of type, because the Airflow DAG only carries a single schedule value.
type WorkflowSchedule struct {
	ID         string                 `gorm:"primaryKey;column:id" json:"id"`
	WorkflowID string                 `gorm:"column:workflow_id;index;not null" json:"workflow_id"`
	Type       WorkflowScheduleType   `gorm:"column:type;not null;default:once" json:"type"`
	RunAt      *time.Time             `gorm:"column:run_at" json:"run_at,omitempty"`
	Cron       *string                `gorm:"column:cron" json:"cron,omitempty"`
	Status     WorkflowScheduleStatus `gorm:"column:status;not null;default:active" json:"status"`
	CreatedAt  time.Time              `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time              `gorm:"autoUpdateTime" json:"updated_at"`
}

func (WorkflowSchedule) TableName() string {
	return "workflow_schedules"
}

// CreateWorkflowScheduleReq is the body for POST /workflow/{wfId}/schedule.
// Exactly one of RunAt / Cron must be provided. RunAt creates a one-shot
// schedule, Cron creates a recurring schedule.
type CreateWorkflowScheduleReq struct {
	RunAt *time.Time `json:"run_at,omitempty" mapstructure:"run_at"`
	Cron  *string    `json:"cron,omitempty" mapstructure:"cron"`
}
