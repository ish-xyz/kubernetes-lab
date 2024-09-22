package config

type Config struct {
	Kubeconfig string `yaml:"kubeconfig"`
	NodeName   string `yaml:"nodeName"`
	Sync       struct {
		NodesCount int `yaml:"nodesCount"`
		Resources  struct {
			Namespace string `yaml:"namespace"`
			Prefix    string `yaml:"prefix"`
		} `yaml:"resources"`
	} `yaml:"sync"`
	PreMigration  []*PackageConfig   `yaml:"preMigration"`
	Migration     []*MigrationConfig `yaml:"migration"`
	PostMigration []*PackageConfig   `yaml:"postMigration"`
}

type MigrationConfig struct {
	SystemdUnit string `yaml:"systemUnit"`
	Manifest    string `yaml:"manifest"`
}

type PackageConfig struct {
	Name     string  `yaml:"name"`
	Driver   string  `yaml:"driver"`
	Manifest *string `yaml:"manifest"`
	Repo     struct {
		Name string `yaml:"name"`
	} `yaml:"repo"`
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
