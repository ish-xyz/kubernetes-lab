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
	Key         string `yaml:"key"`
	SystemdUnit string `yaml:"systemdUnit"`
	Manifest    string `yaml:"manifest"`
	LeaderOnly  bool   `yaml:"leaderOnly"`
	HTTPChecks  []struct {
		URL        string `yaml:"url"`
		CA         string `yaml:"ca"`
		Insecure   bool   `yaml:"insecure"`
		MaxRetries int    `yaml:"maxRetries"`
		Interval   int    `yaml:"interval"`
	} `yaml:"httpChecks"`
	KubectlChecks []struct {
		LabelSelector  string `yaml:"labelSelector"`
		Namespace      string `yaml:"namespace"`
		Node           string `yaml:"node"`
		ExpectedStatus string `yaml:"expectedStatus"`
		MaxRetries     int    `yaml:"maxRetries"`
		Interval       int    `yaml:"interval"`
	} `yaml:"kubectlChecks"`
}

type PackageConfig struct {
	Name       string `yaml:"name"`
	LeaderOnly bool   `yaml:"leaderOnly"`
	Driver     string `yaml:"driver"`
	Manifest   string `yaml:"manifest"`
	Repo       struct {
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
