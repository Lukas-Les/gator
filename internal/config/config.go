package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const configFileName = ".gatorconfig.json"

func GetConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, configFileName), nil
}

func Read(cfgFilePath string) (Config, error) {
	byteValue, err := os.ReadFile(cfgFilePath)
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

func write(cfgFilePath string, cfg Config) error {
	byteValue, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	f, err := os.Create(cfgFilePath)
	defer func() {
		closeErr := f.Close()
		if closeErr != nil && err == nil {
			err = closeErr
		}
	}()
	_, err = f.Write(byteValue)
	if err != nil {
		return err
	}
	return nil
}

func SetUser(cfgFilePath, userName string) error {
	cfg, err := Read(cfgFilePath)
	if err != nil {
		fmt.Printf("Failed to open config file")
		return err
	}
	cfg.CurrentUserName = userName
	err = write(cfgFilePath, cfg)
	if err != nil {
		return err
	}
	return nil
}
