package model

type DAGOperatorOption struct {
	Name  string `json:"name"`
	Value any    `json:"value"`
}

type DAGTask struct {
	TaskName        string              `json:"task_name"`
	Operator        string              `json:"operator"`
	OperatorOptions []DAGOperatorOption `json:"operator_options"`
	Dependencies    []string            `json:"dependencies"`
}

type DAG struct {
	DAGId       string `json:"dag_name"`
	DefaultArgs struct {
		Owner         string `json:"owner"`
		StartDate     string `json:"start_date"`
		EndDate       string `json:"end_date"`
		Retries       int    `json:"retries"`
		RetryDelaySec int    `json:"retry_delay_sec"`
	} `json:"default_args"`
	ScheduleInterval      string    `json:"schedule_interval"`
	Concurrency           int       `json:"concurrency"`
	MaxActiveRuns         int       `json:"max_active_runs"`
	DagrunTimeoutSec      int       `json:"dagrun_timeout_sec"`
	DefaultView           string    `json:"default_view"`
	Orientation           string    `json:"orientation"`
	Description           string    `json:"description"`
	OnSuccessCallbackName string    `json:"on_success_callback_name"`
	OnSuccessCallbackFile string    `json:"on_success_callback_file"`
	OnFailureCallbackName string    `json:"on_failure_callback_name"`
	OnFailureCallbackFile string    `json:"on_failure_callback_file"`
	Tasks                 []DAGTask `json:"tasks"`
}
