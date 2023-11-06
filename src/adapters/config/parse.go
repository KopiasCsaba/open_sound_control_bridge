package config

import (
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

// LoadConfig returns the app config.
func LoadConfig() (*MainConfig, error) {
	cfg := &MainConfig{}
	configPath := "config.yml"
	configPathFromEnv := os.Getenv("APP_CONFIG_FILE")

	if configPathFromEnv != "" {
		configPath = configPathFromEnv
	}

	err := cleanenv.ReadConfig(configPath, cfg)
	if err != nil {
		return nil, fmt.Errorf("config error: %w (APP_CONFIG_FILE env variable: '%s', final config path: %s)", err, configPathFromEnv, configPath)
	}

	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to readEnv: %w", err)
	}

	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
