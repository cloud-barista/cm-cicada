package model

type DefaultArgs struct {
	Owner          string `json:"owner" mapstructure:"owner"`
	StartDate      string `json:"start_date" mapstructure:"start_date"`
	Retries        int    `json:"retries" mapstructure:"retries"`                 // default: 1
	RetryDelaySec  int    `json:"retry_delay_sec" mapstructure:"retry_delay_sec"` // default: 300
	Email          string `json:"email" mapstructure:"email"`
	EmailOnFailure bool   `json:"email_on_failure" mapstructure:"email_on_failure"`
	EmailOnRetry   bool   `json:"email_on_retry" mapstructure:"email_on_retry"`
}

type OperatorOptions []struct {
	Name  string `json:"name" mapstructure:"name"`
	Value any    `json:"value" mapstructure:"value"`
}

type Task struct {
	TaskName        string          `json:"task_name" mapstructure:"task_name"`
	TaskComponent   string          `json:"task_component" mapstructure:"task_component"`
	Operator        string          `json:"operator" mapstructure:"operator"`
	OperatorOptions OperatorOptions `json:"operator_options" mapstructure:"operator_options"`
	Dependencies    []string        `json:"dependencies" mapstructure:"dependencies"`
}

type TaskGroup struct {
	TaskGroupName string `json:"task_group_name" mapstructure:"task_group_name"`
	Description   string `json:"description" mapstructure:"description"`
	Tasks         []Task `json:"tasks" mapstructure:"tasks"`
}

type DAG struct {
	DagID       string      `json:"dag_id" mapstructure:"dag_id"`
	DefaultArgs DefaultArgs `json:"default_args" mapstructure:"default_args"`
	Description string      `json:"description" mapstructure:"description"`
	TaskGroups  []TaskGroup `json:"task_groups" mapstructure:"task_groups"`
}
