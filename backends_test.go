package main

import (
	"os"
	"testing"
)

func TestRemoveEnvVarPresent(t *testing.T) {
	env := []string{"FOO=1", "BAR=2", "FOO=3"}
	result := removeEnvVar(env, "FOO")
	if len(result) != 1 || result[0] != "BAR=2" {
		t.Errorf("expected [BAR=2], got %v", result)
	}
}

func TestRemoveEnvVarAbsent(t *testing.T) {
	env := []string{"FOO=1"}
	result := removeEnvVar(env, "MISSING")
	if len(result) != 1 || result[0] != "FOO=1" {
		t.Errorf("expected [FOO=1] (no-op), got %v", result)
	}
}

func TestRemoveEnvVarEmpty(t *testing.T) {
	result := removeEnvVar([]string{}, "X")
	if len(result) != 0 {
		t.Errorf("expected [], got %v", result)
	}
}

func TestSetEnvVarReplace(t *testing.T) {
	env := []string{"FOO=old", "BAR=2"}
	result := setEnvVar(env, "FOO", "new")
	hasNew := false
	hasOld := false
	hasBar := false
	for _, e := range result {
		if e == "FOO=new" {
			hasNew = true
		}
		if e == "FOO=old" {
			hasOld = true
		}
		if e == "BAR=2" {
			hasBar = true
		}
	}
	if !hasNew {
		t.Errorf("result missing FOO=new: %v", result)
	}
	if hasOld {
		t.Errorf("result still contains FOO=old: %v", result)
	}
	if !hasBar {
		t.Errorf("result missing BAR=2: %v", result)
	}
}

func TestSetEnvVarAdd(t *testing.T) {
	env := []string{"BAR=2"}
	result := setEnvVar(env, "FOO", "new")
	hasNew := false
	hasBar := false
	for _, e := range result {
		if e == "FOO=new" {
			hasNew = true
		}
		if e == "BAR=2" {
			hasBar = true
		}
	}
	if !hasNew {
		t.Errorf("result missing FOO=new: %v", result)
	}
	if !hasBar {
		t.Errorf("result missing BAR=2: %v", result)
	}
}

func TestSetupEnvCC(t *testing.T) {
	// Pre-set the env vars that cc backend should unset, so we can verify removal.
	os.Setenv("ANTHROPIC_BASE_URL", "http://should-be-removed")
	os.Setenv("ANTHROPIC_AUTH_TOKEN", "should-be-removed")
	os.Setenv("OLLAMA_MODEL", "should-be-removed")
	t.Cleanup(func() {
		os.Unsetenv("ANTHROPIC_BASE_URL")
		os.Unsetenv("ANTHROPIC_AUTH_TOKEN")
		os.Unsetenv("OLLAMA_MODEL")
	})

	ccBackend := Backend{
		Command:  "claude",
		Args:     []string{},
		UnsetEnv: []string{"ANTHROPIC_BASE_URL", "ANTHROPIC_AUTH_TOKEN", "OLLAMA_MODEL"},
		SetEnv:   map[string]string{},
	}

	result := setupEnv(ccBackend)

	for _, e := range result {
		if len(e) >= len("ANTHROPIC_BASE_URL=") && e[:len("ANTHROPIC_BASE_URL=")] == "ANTHROPIC_BASE_URL=" {
			t.Errorf("result must NOT contain ANTHROPIC_BASE_URL=, but got: %s", e)
		}
		if len(e) >= len("ANTHROPIC_AUTH_TOKEN=") && e[:len("ANTHROPIC_AUTH_TOKEN=")] == "ANTHROPIC_AUTH_TOKEN=" {
			t.Errorf("result must NOT contain ANTHROPIC_AUTH_TOKEN=, but got: %s", e)
		}
		if len(e) >= len("OLLAMA_MODEL=") && e[:len("OLLAMA_MODEL=")] == "OLLAMA_MODEL=" {
			t.Errorf("result must NOT contain OLLAMA_MODEL=, but got: %s", e)
		}
	}
}

func TestSetupEnvClaudeOllamaCloud(t *testing.T) {
	// Pre-set ANTHROPIC_API_KEY so we can verify it's removed.
	os.Setenv("ANTHROPIC_API_KEY", "should-be-removed")
	t.Cleanup(func() {
		os.Unsetenv("ANTHROPIC_API_KEY")
	})

	cloudBackend := Backend{
		Command:  "claude",
		Args:     []string{"--model", "qwen3.5:cloud"},
		UnsetEnv: []string{"ANTHROPIC_API_KEY"},
		SetEnv: map[string]string{
			"ANTHROPIC_BASE_URL":   "http://localhost:11434",
			"ANTHROPIC_AUTH_TOKEN": "ollama",
		},
	}

	result := setupEnv(cloudBackend)

	hasBaseURL := false
	hasAuthToken := false
	hasAPIKey := false
	for _, e := range result {
		if e == "ANTHROPIC_BASE_URL=http://localhost:11434" {
			hasBaseURL = true
		}
		if e == "ANTHROPIC_AUTH_TOKEN=ollama" {
			hasAuthToken = true
		}
		if len(e) >= len("ANTHROPIC_API_KEY=") && e[:len("ANTHROPIC_API_KEY=")] == "ANTHROPIC_API_KEY=" {
			hasAPIKey = true
		}
	}

	if !hasBaseURL {
		t.Errorf("result must contain ANTHROPIC_BASE_URL=http://localhost:11434, got: %v", result)
	}
	if !hasAuthToken {
		t.Errorf("result must contain ANTHROPIC_AUTH_TOKEN=ollama, got: %v", result)
	}
	if hasAPIKey {
		t.Errorf("result must NOT contain ANTHROPIC_API_KEY=, but it does")
	}
}

func TestBuildCommandStandard(t *testing.T) {
	ccBackend := Backend{
		Command: "claude",
		Args:    []string{},
	}
	binary, args := buildCommand(ccBackend, "fix this")
	if binary != "claude" {
		t.Errorf("expected binary=claude, got %q", binary)
	}
	if len(args) != 1 || args[0] != "fix this" {
		t.Errorf("expected args=[fix this], got %v", args)
	}
}

func TestBuildCommandWithArgs(t *testing.T) {
	clBackend := Backend{
		Command: "ollama",
		Args:    []string{"run", "qwen3-code"},
	}
	binary, args := buildCommand(clBackend, "hello")
	if binary != "ollama" {
		t.Errorf("expected binary=ollama, got %q", binary)
	}
	if len(args) != 3 || args[0] != "run" || args[1] != "qwen3-code" || args[2] != "hello" {
		t.Errorf("expected args=[run qwen3-code hello], got %v", args)
	}
}

func TestBuildCommandURLTemplate(t *testing.T) {
	perplexityBackend := Backend{
		Command:     "xdg-open",
		URLTemplate: "https://www.perplexity.ai/search?q={prompt}",
	}
	binary, args := buildCommand(perplexityBackend, "latest news")
	if binary != "xdg-open" {
		t.Errorf("expected binary=xdg-open, got %q", binary)
	}
	if len(args) != 1 || args[0] != "https://www.perplexity.ai/search?q=latest+news" {
		t.Errorf("expected args=[https://www.perplexity.ai/search?q=latest+news], got %v", args)
	}
}

func TestBuildCommandURLTemplateGemini(t *testing.T) {
	geminiBackend := Backend{
		Command:     "xdg-open",
		URLTemplate: "https://gemini.google.com/app?q={prompt}",
	}
	binary, args := buildCommand(geminiBackend, "hello world")
	if binary != "xdg-open" {
		t.Errorf("expected binary=xdg-open, got %q", binary)
	}
	if len(args) != 1 || args[0] != "https://gemini.google.com/app?q=hello+world" {
		t.Errorf("expected args=[https://gemini.google.com/app?q=hello+world], got %v", args)
	}
}
