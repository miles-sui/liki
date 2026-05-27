package llm

import (
	_ "embed"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

//go:embed models.yaml
var defaultModelsYAML []byte

// ModelConfig holds LLM provider and model configuration.
type ModelConfig struct {
	Provider  string `yaml:"provider"`
	Model     string `yaml:"model"`
	MaxTokens int    `yaml:"max_tokens"`
	Timeout   string `yaml:"timeout"`
}

// LoadModelConfig returns the effective model config.
// If overridePath is non-empty and the file exists, it overrides the embedded defaults.
func LoadModelConfig(overridePath string) ModelConfig {
	cfg := ModelConfig{Model: "claude-haiku-4-5", MaxTokens: 2048, Timeout: "15s"}
	if err := yaml.Unmarshal(defaultModelsYAML, &cfg); err != nil {
		// Use hardcoded defaults above
	}
	if overridePath != "" {
		if data, err := os.ReadFile(overridePath); err == nil {
			yaml.Unmarshal(data, &cfg)
		}
	}
	return cfg
}

// ParseTimeout parses a timeout string, returning the fallback on error.
func ParseTimeout(s string, fallback time.Duration) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return fallback
	}
	return d
}
