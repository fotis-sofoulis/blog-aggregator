package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	DbUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func (c *Config)SetUser(currUserName string) error {
	c.CurrentUserName = currUserName

	return write(*c)
}

const configFileName = ".gatorconfig.json"

func Read() (Config, error) {

	confPath, err := getConfigFilePath()
	if err != nil {
		return Config{}, fmt.Errorf("could not get the config file path: %w", err)
	}

	data, err := os.ReadFile(confPath)
	if err != nil {
		return Config{}, fmt.Errorf("could not read file: %w", err)
	}
	
	var gatorConfig Config
	if err := json.Unmarshal(data, &gatorConfig); err != nil {
		return Config{}, fmt.Errorf("Error unmarshalling JSON: %w", err)
	}

	return gatorConfig, nil

}

func getConfigFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not get the home directory: %w", err)
	}

	configURL := filepath.Join(home, configFileName)
	if _, err := os.Stat(configURL); err != nil {
		return "", fmt.Errorf("could not locate the config file")
	}

	return configURL, nil

}

func write(cfg Config) error {
	confPath, err := getConfigFilePath()
	if err != nil {
		return fmt.Errorf("could not get the config file path: %w", err)
	}

	jsonData, err := json.MarshalIndent(cfg, "", " ")
	if err != nil {
		return fmt.Errorf("could not marshal the config: %w", err)
	}

	if err := os.WriteFile(confPath, jsonData, 0644); err != nil {
		return fmt.Errorf("could not write config file: %w", err)
	}

	return nil

}
