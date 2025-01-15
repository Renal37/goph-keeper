package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	env "github.com/caarlos0/env/v6"
)

// ConfigENV contains app settings.
type ConfigENV struct {
	JWTkey             string `json:"jwt_key" env:"JWT_KEY"`
	Host               string `json:"host" env:"HOST"`
	DSN                string `json:"dsn" env:"DSN"`
	CertificatePath    string `json:"certificate"`
	CertificateKeyPath string `json:"certificate_key"`
	MasterKey          string
}

// GetConfig get app settings.
func GetConfig() (*ConfigENV, error) {
	var eCfg ConfigENV
	configPath := "config/server.json"

	flag.StringVar(&eCfg.MasterKey, "mk", "", "master key for encryption keys")
	flag.Parse()

	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&eCfg); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	if err := file.Close(); err != nil {
		return nil, fmt.Errorf("failed close config file: %w", err)
	}

	err = env.Parse(&eCfg)
	if err != nil {
		return nil, fmt.Errorf("failed parsing environment variables: %w", err)
	}

	return &eCfg, nil
}
