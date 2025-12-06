package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func GetConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".gatorconfig.json"), nil
}

func Read(configFilePath string) (Config, error) {
	byteValue, err := os.ReadFile(configFilePath)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	err = json.Unmarshal(byteValue, &cfg)
	if err != nil {
		return Config{}, nil
	}
	return cfg, nil
}
