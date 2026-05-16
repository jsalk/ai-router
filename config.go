package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config is the top-level configuration for ai-router.
type Config struct {
	Backends         map[string]Backend `yaml:"backends"`
	RoutingRules     []RoutingRule      `yaml:"routing_rules"`
	DefaultBackend   string             `yaml:"default_backend"`
	FallbackBackends []string           `yaml:"fallback_backends"`
}

// Backend defines how to invoke and health-check a single AI backend.
type Backend struct {
	Command          string            `yaml:"command"`
	Args             []string          `yaml:"args"`
	SetEnv           map[string]string `yaml:"set_env"`
	UnsetEnv         []string          `yaml:"unset_env"`
	URLTemplate      string            `yaml:"url_template"`
	HealthCheckType  string            `yaml:"health_check_type"`
	HealthCheckValue string            `yaml:"health_check_value"`
	Description      string            `yaml:"description"`
}

// RoutingRule maps a set of keywords to a backend with a given priority.
type RoutingRule struct {
	Keywords []string `yaml:"keywords"`
	Backend  string   `yaml:"backend"`
	Priority int      `yaml:"priority"`
}

// defaultConfigPath returns the canonical location for the ai-router config file.
func defaultConfigPath() string {
	return filepath.Join(os.Getenv("HOME"), ".config", "ai-router", "config.yaml")
}

// loadConfig reads and unmarshals the YAML config file at path.
// Returns a helpful error if the file is missing or the YAML is malformed.
func loadConfig(path string) (*Config, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf(
			"config file not found: %s\nPlease create the file by copying: cp ai-router-config.yaml ~/.config/ai-router/config.yaml",
			path,
		)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("invalid config at %s: %w", path, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("invalid config at %s: %w", path, err)
	}

	return &config, nil
}

// validateConfig checks that all routing rules and the default backend reference
// defined backends, and that every backend has a non-empty command.
func validateConfig(c *Config) error {
	// Check every backend has a command field.
	for name, backend := range c.Backends {
		if backend.Command == "" {
			return fmt.Errorf("backend %q: missing command field", name)
		}
	}

	// Check each routing rule references a defined backend.
	for _, rule := range c.RoutingRules {
		if _, ok := c.Backends[rule.Backend]; !ok {
			return fmt.Errorf("routing rule references undefined backend: %q", rule.Backend)
		}
	}

	// Check default_backend is defined.
	if _, ok := c.Backends[c.DefaultBackend]; !ok {
		return fmt.Errorf("default_backend %q is not defined in backends", c.DefaultBackend)
	}

	return nil
}
