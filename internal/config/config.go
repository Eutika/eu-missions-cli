package config

import (
	"os"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	mu                sync.RWMutex
	keyringService    string
	clientID          string
	deviceCodeURL     string
	tokenURL          string
	remoteURL         string
	dangerousPatterns []string
}

func NewConfig() *Config {
	return &Config{
		keyringService: "missions-cli",
		clientID:       "missions",
		deviceCodeURL:  "https://missions.eutika.com/api/auth/device/code",
		tokenURL:       "https://missions.eutika.com/api/auth/device/token",
		dangerousPatterns: []string{
			"rm -rf", "sudo", "dd ", ":(){ :|:& };:", "mkfs", "format",
		},
	}
}

func (c *Config) GetKeyringService() string {
	return c.keyringService
}

func (c *Config) GetClientID() string {
	return c.clientID
}

func (c *Config) GetDeviceCodeURL() string {
	if err := godotenv.Load(); err != nil {

	}
	// Check for environment variable override
	if envURL := os.Getenv("MISSIONS_CLI_DEVICE_CODE_URL"); envURL != "" {
		return envURL
	}
	return c.deviceCodeURL
}

func (c *Config) GetTokenURL() string {
	if err := godotenv.Load(); err != nil {

	}
	// Check for environment variable override
	if envURL := os.Getenv("MISSIONS_CLI_TOKEN_URL"); envURL != "" {

		return envURL
	}
	return c.tokenURL
}

func (c *Config) GetRemoteURL() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Load .env file if possible, but don't fail if it doesn't exist
	if err := godotenv.Load(); err != nil {

	}

	// Default URL
	defaultURL := "https://missions.eutika.com/api/cli"

	// Check for environment variable, use it if set
	if envURL := os.Getenv("MISSIONS_CLI_URL"); envURL != "" {

		return envURL
	}

	return defaultURL
}

func (c *Config) GetDangerousPatterns() []string {
	return c.dangerousPatterns
}
