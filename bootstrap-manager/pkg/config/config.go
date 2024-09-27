package config

type Config struct {
	Kubeconfig string `yaml:"kubeconfig" validate:"required"`
	NodeName   string `yaml:"nodeName" validate:"required"`
	Sync       struct {
		NodesCount int `yaml:"nodesCount" validate:"required,gt=0"`
		Resources  struct {
			Namespace string `yaml:"namespace" validate:"required"`
			Prefix    string `yaml:"prefix" validate:"required"`
		} `yaml:"resources" validate:"dive,required"`
	} `yaml:"sync" validate:"dive,required"`
	PreMigration  []*PackageConfig   `yaml:"preMigration" validate:"dive,required"`
	Migration     []*MigrationConfig `yaml:"migration" validate:"dive,required"`
	PostMigration []*PackageConfig   `yaml:"postMigration" validate:"dive,required"`
}

type MigrationConfig struct {
	Key         string `yaml:"key" validate:"required"`
	SystemdUnit string `yaml:"systemdUnit" validate:"required"`
	Manifest    string `yaml:"manifest" validate:"required"`
	LeaderOnly  bool   `yaml:"leaderOnly"`
	HTTPChecks  []struct {
		URL        string `yaml:"url" validate:"url,required"`
		CA         string `yaml:"ca" validate:"filepath,required"`
		Insecure   bool   `yaml:"insecure"`
		MaxRetries int    `yaml:"maxRetries" validate:"gt=1,required"`
		Interval   int    `yaml:"interval" validate:"gt=1,required"`
	} `yaml:"httpChecks" validate:"dive,required"`
	KubectlChecks []struct {
		LabelSelector  string `yaml:"labelSelector"`
		Namespace      string `yaml:"namespace"`
		Node           string `yaml:"node"`
		ExpectedStatus string `yaml:"expectedStatus"`
		MaxRetries     int    `yaml:"maxRetries"`
		Interval       int    `yaml:"interval"`
	} `yaml:"kubectlChecks" validate:"dive,required"`
}

type PackageConfig struct {
	Name       string  `yaml:"name"`
	LeaderOnly bool    `yaml:"leaderOnly"`
	Driver     string  `yaml:"driver"`
	Manifest   *string `yaml:"manifest"`
	Repo       struct {
		Name string `yaml:"name"`
	} `yaml:"repo" validate:"dive,required"`
	Chart *ChartConfig `yaml:"chart"`
}

type ChartConfig struct {
	Url         string                 `yaml:"url"`
	Name        string                 `yaml:"name"`
	ReleaseName string                 `yaml:"releaseName"`
	Namespace   string                 `yaml:"namespace"`
	Version     string                 `yaml:"version"`
	Values      map[string]interface{} `yaml:"values" json:"values"`
}
