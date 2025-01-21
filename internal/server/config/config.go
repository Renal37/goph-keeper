package config

type Config struct {
	Hostport       string `yaml:"hostport"`
	DataSourceName string `yaml:"dataSourceName"`
	DisableLogging bool   `yaml:"disableLogging"`
}
