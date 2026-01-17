package config

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
}
