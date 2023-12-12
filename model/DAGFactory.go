package model

type DAGFactoryDAGDefaultArgs struct {
	Owner         string `yaml:"owner"`
	StartDate     string `yaml:"start_date"`
	Retries       int    `yaml:"retries"`
	RetryDelaySec int    `yaml:"retry_delay_sec"`
}

type DAGFactoryDAGTaskGroupStruct struct {
	Tooltip string `yaml:"tooltip"`
}

type DAGFactoryDAGTaskGroup struct {
	DAGFactoryDAGTaskGroupStruct DAGFactoryDAGTaskGroupStruct `yaml:"###dag_factory_dag_task_group_struct###"`
}

type DAGFactoryDAGTask struct {
	DAGTaskStruct map[string]any `yaml:"###dag_task_struct###"`
}

type DAGFactoryDAGStruct struct {
	DefaultArgs DAGFactoryDAGDefaultArgs `yaml:"default_args"`
	DefaultView string                   `yaml:"default_view"`
	Orientation string                   `yaml:"orientation"`
	Description string                   `yaml:"description"`
	TaskGroups  []DAGFactoryDAGTaskGroup `yaml:"task_groups"`
	Tasks       []DAGFactoryDAGTask      `yaml:"tasks"`
}

type DAGFactory struct {
	DAGStruct DAGFactoryDAGStruct `yaml:"###dag_struct###"`
}
