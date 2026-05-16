package main

import (
	"strings"
	"testing"
)

func TestLoadConfigValid(t *testing.T) {
	cfg, err := loadConfig("testdata/valid.yaml")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	expectedBackends := []string{"cc", "cl", "claude-ollama-cloud", "perplexity", "gemini"}
	for _, name := range expectedBackends {
		if _, ok := cfg.Backends[name]; !ok {
			t.Errorf("expected backend %q to be present", name)
		}
	}
	if cfg.DefaultBackend != "cl" {
		t.Errorf("expected default_backend %q, got %q", "cl", cfg.DefaultBackend)
	}
	if len(cfg.RoutingRules) != 4 {
		t.Errorf("expected 4 routing rules, got %d", len(cfg.RoutingRules))
	}
}

func TestLoadConfigMissing(t *testing.T) {
	_, err := loadConfig("/nonexistent/path.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
	if !strings.Contains(err.Error(), "config file not found") {
		t.Errorf("expected error to contain %q, got: %v", "config file not found", err)
	}
	if !strings.Contains(err.Error(), "cp ai-router-config.yaml") {
		t.Errorf("expected error to contain copy hint %q, got: %v", "cp ai-router-config.yaml", err)
	}
}

func TestLoadConfigMalformed(t *testing.T) {
	_, err := loadConfig("testdata/malformed.yaml")
	if err == nil {
		t.Fatal("expected error for malformed YAML, got nil")
	}
	if !strings.Contains(err.Error(), "invalid config") {
		t.Errorf("expected error to contain %q, got: %v", "invalid config", err)
	}
}

func TestValidateConfigUndefinedBackend(t *testing.T) {
	cfg := &Config{
		Backends: map[string]Backend{
			"cc": {Command: "claude", HealthCheckType: "env_var", HealthCheckValue: "ANTHROPIC_API_KEY"},
		},
		RoutingRules: []RoutingRule{
			{Keywords: []string{"code"}, Backend: "ghost", Priority: 1},
		},
		DefaultBackend: "cc",
	}
	err := validateConfig(cfg)
	if err == nil {
		t.Fatal("expected error for undefined backend reference, got nil")
	}
	if !strings.Contains(err.Error(), "undefined backend") {
		t.Errorf("expected error to contain %q, got: %v", "undefined backend", err)
	}
	if !strings.Contains(err.Error(), "ghost") {
		t.Errorf("expected error to contain backend name %q, got: %v", "ghost", err)
	}
}

func TestValidateConfigBadDefault(t *testing.T) {
	cfg := &Config{
		Backends: map[string]Backend{
			"cc": {Command: "claude", HealthCheckType: "env_var", HealthCheckValue: "ANTHROPIC_API_KEY"},
		},
		RoutingRules:   []RoutingRule{},
		DefaultBackend: "ghost",
	}
	err := validateConfig(cfg)
	if err == nil {
		t.Fatal("expected error for undefined default_backend, got nil")
	}
	if !strings.Contains(err.Error(), "default_backend") {
		t.Errorf("expected error to contain %q, got: %v", "default_backend", err)
	}
}

func TestValidateConfigMissingCommand(t *testing.T) {
	cfg := &Config{
		Backends: map[string]Backend{
			"cc": {Command: "", HealthCheckType: "env_var", HealthCheckValue: "ANTHROPIC_API_KEY"},
		},
		RoutingRules:   []RoutingRule{},
		DefaultBackend: "cc",
	}
	err := validateConfig(cfg)
	if err == nil {
		t.Fatal("expected error for backend with missing command, got nil")
	}
	if !strings.Contains(err.Error(), "missing command") {
		t.Errorf("expected error to contain %q, got: %v", "missing command", err)
	}
}

func TestValidateConfigValid(t *testing.T) {
	cfg, err := loadConfig("testdata/valid.yaml")
	if err != nil {
		t.Fatalf("failed to load valid config: %v", err)
	}
	err = validateConfig(cfg)
	if err != nil {
		t.Errorf("expected nil error for valid config, got: %v", err)
	}
}
