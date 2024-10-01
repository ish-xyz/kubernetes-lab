package config

type Config struct {
	Kubeconfig    string             `yaml:"kubeconfig" validate:"required"`
	NodeName      string             `yaml:"nodeName" validate:"required"`
	Sync          *SyncConfig        `yaml:"sync" validate:"required"`
	PreMigration  []*PackageConfig   `yaml:"preMigration" validate:"required,dive"`
	Migration     []*MigrationConfig `yaml:"migration" validate:"required,dive"`
	PostMigration []*PackageConfig   `yaml:"postMigration" validate:"required,dive"`
}

type SyncConfig struct {
	NodesCount int    `yaml:"nodesCount" validate:"required,gt=0"`
	Namespace  string `yaml:"namespace" validate:"required"`
	Prefix     string `yaml:"prefix" validate:"required"`
}

type PackageConfig struct {
	Name       string       `yaml:"name" validate:"required"`
	LeaderOnly bool         `yaml:"leaderOnly"`
	Driver     string       `yaml:"driver" validate:"required,oneof=kubectl helm"`
	Manifest   string       `yaml:"manifest" validate:"required_if=Driver kubectl"`
	Chart      *ChartConfig `yaml:"chart" validate:"required_if=Driver helm"`
}

type MigrationConfig struct {
	Key         string             `yaml:"key" validate:"required"`
	SystemdUnit string             `yaml:"systemdUnit" validate:"required"`
	Driver      string             `yaml:"driver" validate:"required,oneof=kubectl helm"`
	Manifest    string             `yaml:"manifest" validate:"required_if=Driver kubectl"`
	Chart       *ChartConfig       `yaml:"chart" validate:"required_if=Driver helm"`
	LeaderOnly  bool               `yaml:"leaderOnly"`
	HTTPChecks  []*HTTPCheckConfig `yaml:"httpChecks" validate:"dive,required"`
	PodChecks   []*PodCheckConfig  `yaml:"podChecks" validate:"dive,required"`
}

type ChartConfig struct {
	Url         string                 `yaml:"url" validate:"required"`
	Name        string                 `yaml:"name" validate:"required"`
	ReleaseName string                 `yaml:"releaseName" validate:"required"`
	Namespace   string                 `yaml:"namespace" validate:"required"`
	Version     string                 `yaml:"version" validate:"required"`
	Values      map[string]interface{} `yaml:"values" json:"values"`
}

type PodCheckConfig struct {
	LabelSelector  string `yaml:"labelSelector" validate:"required"`
	Namespace      string `yaml:"namespace" validate:"required"`
	Node           string `yaml:"node" validate:"required"`
	ExpectedStatus string `yaml:"expectedStatus" validate:"required"`
	MaxRetries     int    `yaml:"maxRetries" validate:"gt=1,required"`
	Interval       int    `yaml:"interval" validate:"gt=1,required"`
}

type HTTPCheckConfig struct {
	URL        string `yaml:"url" validate:"url,required"`
	CA         string `yaml:"ca" validate:"filepath"`
	Insecure   bool   `yaml:"insecure"`
	MaxRetries int    `yaml:"maxRetries" validate:"gt=1,required"`
	Interval   int    `yaml:"interval" validate:"gt=1,required"`
}
