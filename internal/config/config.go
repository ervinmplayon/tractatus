package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config represents the application configuration
type Config struct {
	Accounts map[string]Account `json:"accounts"`
}

// Represents a single AWS application configuration
type Account struct {
	AccountID       string `json:"account_id"`
	Region          string `json:"region"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	SessionToken    string `json:"session_token"`
}

var LoadConfig = func(filepath string) (*Config, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("loadConfig: failed to open config file: %w", err)
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("loadConfig: failed to parse config file: %w", err)
	}

	// validate config
	if len(config.Accounts) == 0 {
		return nil, fmt.Errorf("loadConfig: no accounts defined in config")
	}

	for name, account := range config.Accounts {
		if account.AccountID == "" {
			return nil, fmt.Errorf("account '%s' missing account_id", name)
		}
		if account.Region == "" {
			return nil, fmt.Errorf("account '%s' missing region", name)
		}
		if account.AccessKeyID == "" {
			return nil, fmt.Errorf("account '%s' missing access_key_id", name)
		}
		if account.SecretAccessKey == "" {
			return nil, fmt.Errorf("account '%s' missing secret_access_key", name)
		}
		if account.SessionToken == "" {
			return nil, fmt.Errorf("account '%s' missing session_token", name)
		}
	}

	return &config, nil
}
