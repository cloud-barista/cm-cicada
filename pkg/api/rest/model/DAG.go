package model

type DAGDefaultArgs struct {
	Owner         string `json:"owner" mapstructure:"owner"`
	StartDate     string `json:"start_date" mapstructure:"start_date"`
	Retries       int    `json:"retries" mapstructure:"retries"`
	RetryDelaySec int    `json:"retry_delay_sec" mapstructure:"retry_delay_sec"`
}

type DAGOperatorOption struct {
	Name  string `json:"name" mapstructure:"name"`
	Value string `json:"value" mapstructure:"value"`
}

type DAGTaskGroup struct {
	TaskGroupName string `json:"task_group_name" mapstructure:"task_group_name"`
	Tooltip       string `yaml:"tooltip" mapstructure:"tooltip"`
}

type DAGTask struct {
	TaskName        string              `json:"task_name" mapstructure:"task_name"`
	Operator        string              `json:"operator" mapstructure:"operator"`
	OperatorOptions []DAGOperatorOption `json:"operator_options" mapstructure:"operator_options"`
	TaskGroupName   string              `json:"task_group_name" mapstructure:"task_group_name"`
	Dependencies    []string            `json:"dependencies" mapstructure:"dependencies"`
}

type DAG struct {
	DAGId       string         `json:"dag_id" mapstructure:"dag_id"`
	DefaultArgs DAGDefaultArgs `json:"default_args" mapstructure:"default_args"`
	DefaultView string         `json:"default_view" mapstructure:"default_view"` // default: 'graph', or 'tree', 'duration', 'gantt', 'landing_times'
	Orientation string         `json:"orientation" mapstructure:"orientation"`   // default: 'LR', or 'TB', 'RL', 'BT'
	Description string         `json:"description" mapstructure:"description"`
	TaskGroups  []DAGTaskGroup `json:"task_groups" mapstructure:"task_groups"`
	Tasks       []DAGTask      `json:"tasks" mapstructure:"tasks"`
}
