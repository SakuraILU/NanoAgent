package Config

import (
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Reactor AgentConfig  `yaml:"reactor"`
	LLM     LLMConfig    `yaml:"llm"`
	Serper  SerperConfig `yaml:"serper"`
}

type AgentConfig struct {
	SystemPrompt string `yaml:"system_prompt"`
	UserPrompt   string `yaml:"user_prompt"`
	MaxEpoch     int    `yaml:"max_epoch"`
}

type LLMConfig struct {
	APIKey      string  `yaml:"api_key"`
	BaseURL     string  `yaml:"base_url"`
	Model       string  `yaml:"model"`
	MaxTokens   int     `yaml:"max_tokens"`
	Temperature float32 `yaml:"temperature"`
	Stream      bool    `yaml:"stream"`
	Thinking    bool    `yaml:"thinking"`
}

type SerperConfig struct {
	APIKey  string `yaml:"api_key"`
	BaseURL string `yaml:"base_url"`
	Limit   int    `yaml:"limit"`
}

var (
	globalConfig *Config
	configOnce   sync.Once
	configPath   string
)

func LoadConfig(path string) (*Config, error) {
	var err error
	configOnce.Do(func() {
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			err = readErr
			return
		}

		var cfg Config
		if unmarshalErr := yaml.Unmarshal(data, &cfg); unmarshalErr != nil {
			err = unmarshalErr
			return
		}

		// Override with env var if set
		if apiKey := os.Getenv("BIGMODEL_API_KEY"); apiKey != "" {
			cfg.LLM.APIKey = apiKey
		}

		// Set defaults if not provided
		if cfg.Reactor.MaxEpoch == 0 {
			cfg.Reactor.MaxEpoch = 50
		}
		if cfg.LLM.MaxTokens == 0 {
			cfg.LLM.MaxTokens = 4096
		}
		if cfg.Serper.Limit == 0 {
			cfg.Serper.Limit = 10
		}

		globalConfig = &cfg
		configPath = path
	})

	if err != nil {
		return nil, err
	}

	return globalConfig, nil
}

func GetConfig() *Config {
	if globalConfig == nil {
		panic("Config not loaded. Call LoadConfig first.")
	}
	return globalConfig
}