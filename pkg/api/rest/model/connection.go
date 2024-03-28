package model

type Connection struct {
	ID          string `json:"id" yaml:"id" mapstructure:"id"`
	Type        string `json:"type" yaml:"type" mapstructure:"type"`
	Description string `json:"description" yaml:"description" mapstructure:"description"`
	Host        string `json:"host" yaml:"host" mapstructure:"host"`
	Port        int32  `json:"port" yaml:"port" mapstructure:"port"`
	Schema      string `json:"schema" yaml:"schema" mapstructure:"schema"`
	Login       string `json:"login" yaml:"login" mapstructure:"login"`
	Password    string `json:"password" yaml:"password" mapstructure:"password"`
}
