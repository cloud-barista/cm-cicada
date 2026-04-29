package catalog

// TaskTypeDef는 카탈로그의 한 task type 정의.
// conf/task_types.yaml의 각 항목과 매핑된다.
type TaskTypeDef struct {
	ID              string                 `yaml:"id"`
	Label           string                 `yaml:"label"`
	Category        string                 `yaml:"category"`
	OperatorClass   string                 `yaml:"operator_class"`
	ComponentSchema map[string]FieldSchema `yaml:"component_schema,omitempty"`
	TaskSchema      map[string]FieldSchema `yaml:"task_schema,omitempty"`
}

// FieldSchema는 spec 한 필드의 검증 규칙.
type FieldSchema struct {
	Type     string   `yaml:"type"`
	Required bool     `yaml:"required"`
	Enum     []string `yaml:"enum,omitempty"`
	Min      *int     `yaml:"min,omitempty"`
	Max      *int     `yaml:"max,omitempty"`
}

// catalogFile은 yaml 최상위 구조 (내부용).
type catalogFile struct {
	TaskTypes []TaskTypeDef `yaml:"task_types"`
}
