package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/Renal37/goph-keeper/internal/yaml"
)

const (
	DefaultAppDirName = ".goph-keeper"
	hostportDefault   = "http://localhost:34555"
	dbDefault         = "goph-keeper.db"
)

type Config struct {
	Hostport string `yaml:"hostport"`
	Database string `yaml:"database"`
}

func ReadConfig(filename string) (*Config, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	confFilePath := filepath.Join(dir, DefaultAppDirName, filename)

	conf, err := yaml.ReadYaml[Config](confFilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// write default config
			conf = defaultConfig()

			return conf, yaml.WriteYaml(confFilePath, conf)
		}

		return nil, err
	}

	if conf.Hostport == "" {
		conf.Hostport = hostportDefault
	}

	if conf.Database == "" {
		conf.Database = dbDefault
	}

	return conf, nil
}

func UpdateConfig(filename string, params map[string]string) error {
	dir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	fullPath := filepath.Join(dir, DefaultAppDirName, filename)

	config, err := ReadConfig(fullPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			config = defaultConfig()
		} else {
			return err
		}
	}

	if hostport, ok := params["hostport"]; ok {
		config.Hostport = hostport
	}

	if database, ok := params["database"]; ok {
		config.Database = database
	}

	return yaml.WriteYaml(fullPath, config)
}

func defaultConfig() *Config {
	return &Config{
		Hostport: hostportDefault,
		Database: dbDefault,
	}
}
