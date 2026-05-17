package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Cost tier constants — valid values for Backend.CostTier (D-04).
const (
	CostFree   = "free"
	CostLow    = "low"
	CostMedium = "medium"
	CostHigh   = "high"

	LatencyInstant = "instant"
	LatencyFast    = "fast"
	LatencyMedium  = "medium"
	LatencySlow    = "slow"
)

var validCostTiers = map[string]bool{
	CostFree: true, CostLow: true, CostMedium: true, CostHigh: true,
}

var validLatencyTiers = map[string]bool{
	LatencyInstant: true, LatencyFast: true, LatencyMedium: true, LatencySlow: true,
}

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
	CostTier         string            `yaml:"cost_tier"`    // optional; one of: free, low, medium, high
	LatencyTier      string            `yaml:"latency_tier"` // optional; one of: instant, fast, medium, slow
}

// RoutingRule maps a set of keywords to a backend with a given priority.
type RoutingRule struct {
	Keywords []string `yaml:"keywords"`
	Backend  string   `yaml:"backend"`
	Priority int      `yaml:"priority"`
}

// defaultConfigPath returns the canonical location for the ai-router config file.
// Returns an error when the home directory cannot be determined (WR-02).
func defaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".config", "ai-router", "config.yaml"), nil
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

// validateURLTemplate checks that a URL template has an http or https scheme (CR-04).
func validateURLTemplate(tmpl string) error {
	parsed, err := url.Parse(strings.Replace(tmpl, "{prompt}", "x", -1))
	if err != nil {
		return fmt.Errorf("invalid url_template: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return fmt.Errorf("url_template scheme must be http or https, got %q", parsed.Scheme)
	}
	return nil
}

// validateConfig checks that all routing rules and the default backend reference
// defined backends, and that every backend has a non-empty command or url_template.
func validateConfig(c *Config) error {
	// Check every backend has a command (when url_template is not set) and
	// validate any url_template scheme (CR-04, WR-04).
	for name, backend := range c.Backends {
		if backend.URLTemplate == "" && backend.Command == "" {
			return fmt.Errorf("backend %q: missing command field (required when url_template is not set)", name)
		}
		if backend.URLTemplate != "" {
			if err := validateURLTemplate(backend.URLTemplate); err != nil {
				return fmt.Errorf("backend %q: %w", name, err)
			}
		}
	}

	// Check each routing rule references a defined backend.
	for _, rule := range c.RoutingRules {
		if _, ok := c.Backends[rule.Backend]; !ok {
			return fmt.Errorf("routing rule references undefined backend: %q", rule.Backend)
		}
	}

	// Check default_backend is explicitly set and defined (CR-02).
	if c.DefaultBackend == "" {
		return fmt.Errorf("default_backend is required but not set")
	}
	if _, ok := c.Backends[c.DefaultBackend]; !ok {
		return fmt.Errorf("default_backend %q is not defined in backends", c.DefaultBackend)
	}

	// Validate fallback_backends entries against defined backends (CR-02).
	for i, name := range c.FallbackBackends {
		if _, ok := c.Backends[name]; !ok {
			return fmt.Errorf("fallback_backends[%d] %q is not defined in backends", i, name)
		}
	}

	// Validate optional cost_tier and latency_tier fields (T-05-02-01).
	// Empty values are allowed (D-06: graceful degradation).
	for name, backend := range c.Backends {
		if backend.CostTier != "" && !validCostTiers[backend.CostTier] {
			return fmt.Errorf("backend %q: invalid cost_tier %q (must be one of: free, low, medium, high)", name, backend.CostTier)
		}
		if backend.LatencyTier != "" && !validLatencyTiers[backend.LatencyTier] {
			return fmt.Errorf("backend %q: invalid latency_tier %q (must be one of: instant, fast, medium, slow)", name, backend.LatencyTier)
		}
	}

	return nil
}
