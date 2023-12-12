package model

type DAGDefaultArgs struct {
	Owner         string `json:"owner"`
	StartDate     string `json:"start_date"`
	Retries       int    `json:"retries"`
	RetryDelaySec int    `json:"retry_delay_sec"`
}

type DAGOperatorOption struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type DAGTaskGroup struct {
	TaskGroupName string `json:"task_group_name"`
	Tooltip       string `yaml:"tooltip"`
}

type DAGTask struct {
	TaskName        string              `json:"task_name"`
	Operator        string              `json:"operator"`
	OperatorOptions []DAGOperatorOption `json:"operator_options"`
	TaskGroupName   string              `json:"task_group_name"`
	Dependencies    []string            `json:"dependencies"`
}

type DAG struct {
	DAGId       string         `json:"dag_id"`
	DefaultArgs DAGDefaultArgs `json:"default_args"`
	DefaultView string         `json:"default_view"` // default: 'graph', or 'tree', 'duration', 'gantt', 'landing_times'
	Orientation string         `json:"orientation"`  // default: 'LR', or 'TB', 'RL', 'BT'
	Description string         `json:"description"`
	TaskGroups  []DAGTaskGroup `json:"task_groups"`
	Tasks       []DAGTask      `json:"tasks"`
}
