package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type LLMConfig struct {
	APIKey      string  `yaml:"api_key"`
	BaseURL     string  `yaml:"base_url"`
	Model       string  `yaml:"model"`
	MaxTokens   int     `yaml:"max_tokens"`
	Temperature float32 `yaml:"temperature"`
	Stream      bool    `yaml:"stream"`
}

type Config struct {
	LLM LLMConfig `yaml:"llm"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Override with env var if set
	if apiKey := os.Getenv("BIGMODEL_API_KEY"); apiKey != "" {
		cfg.LLM.APIKey = apiKey
	}

	return &cfg, nil
}