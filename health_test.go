package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

// TestCheckHealthHTTPSuccess verifies that an HTTP backend returning 200 is healthy.
func TestCheckHealthHTTPSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	backend := Backend{HealthCheckType: "http", HealthCheckValue: srv.URL}
	if !checkHealth(backend) {
		t.Error("expected checkHealth to return true for reachable HTTP server, got false")
	}
}

// TestCheckHealthHTTPFailure verifies that an unreachable HTTP backend is unhealthy.
// Uses port 1 which is always connection-refused on Linux.
func TestCheckHealthHTTPFailure(t *testing.T) {
	start := time.Now()
	backend := Backend{HealthCheckType: "http", HealthCheckValue: "http://localhost:1"}
	if checkHealth(backend) {
		t.Error("expected checkHealth to return false for unreachable URL, got true")
	}
	elapsed := time.Since(start)
	if elapsed >= 5*time.Second {
		t.Errorf("checkHealth took too long (%v); expected < 5s", elapsed)
	}
}

// TestCheckHealthHTTPTimeout verifies the 3-second timeout is enforced for slow HTTP backends.
func TestCheckHealthHTTPTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use request context so the handler exits quickly when the client disconnects.
		select {
		case <-r.Context().Done():
		case <-time.After(5 * time.Second):
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer srv.Close()

	backend := Backend{HealthCheckType: "http", HealthCheckValue: srv.URL}
	start := time.Now()
	result := checkHealth(backend)
	elapsed := time.Since(start)

	if result {
		t.Error("expected checkHealth to return false for slow server (timeout), got true")
	}
	if elapsed >= 4*time.Second {
		t.Errorf("checkHealth took %v; expected to timeout before 4s", elapsed)
	}
}

// TestCheckHealthEnvVarPresent verifies that a non-empty env var reports as healthy.
func TestCheckHealthEnvVarPresent(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "sk-test")
	backend := Backend{HealthCheckType: "env_var", HealthCheckValue: "ANTHROPIC_API_KEY"}
	if !checkHealth(backend) {
		t.Error("expected checkHealth to return true when ANTHROPIC_API_KEY is set, got false")
	}
}

// TestCheckHealthEnvVarEmpty verifies that an empty env var reports as unhealthy.
func TestCheckHealthEnvVarEmpty(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	backend := Backend{HealthCheckType: "env_var", HealthCheckValue: "ANTHROPIC_API_KEY"}
	if checkHealth(backend) {
		t.Error("expected checkHealth to return false when ANTHROPIC_API_KEY is empty, got true")
	}
}

// TestCheckHealthEnvVarUnset verifies that an unset env var reports as unhealthy.
func TestCheckHealthEnvVarUnset(t *testing.T) {
	os.Unsetenv("ANTHROPIC_API_KEY") //nolint:errcheck
	backend := Backend{HealthCheckType: "env_var", HealthCheckValue: "ANTHROPIC_API_KEY"}
	if checkHealth(backend) {
		t.Error("expected checkHealth to return false when ANTHROPIC_API_KEY is unset, got true")
	}
}

// TestCheckHealthUnknownType verifies that an unknown health check type defaults to available.
func TestCheckHealthUnknownType(t *testing.T) {
	backend := Backend{HealthCheckType: "unknown"}
	if !checkHealth(backend) {
		t.Error("expected checkHealth to return true for unknown type (assume available), got false")
	}
}

// TestCheckHealthPerplexityKey verifies Perplexity env var check works.
func TestCheckHealthPerplexityKey(t *testing.T) {
	t.Setenv("PERPLEXITY_API_KEY", "pplx-abc")
	backend := Backend{HealthCheckType: "env_var", HealthCheckValue: "PERPLEXITY_API_KEY"}
	if !checkHealth(backend) {
		t.Error("expected checkHealth to return true when PERPLEXITY_API_KEY is set, got false")
	}
}

// TestSelectAvailableBackendChosenHealthy verifies that the chosen backend is returned when healthy.
func TestSelectAvailableBackendChosenHealthy(t *testing.T) {
	t.Setenv("TEST_CC_KEY", "sk-test-value")

	cfg := &Config{
		Backends: map[string]Backend{
			"cc": {Command: "claude", HealthCheckType: "env_var", HealthCheckValue: "TEST_CC_KEY"},
			"cl": {Command: "ollama", HealthCheckType: "env_var", HealthCheckValue: "TEST_MISSING_KEY"},
		},
		FallbackBackends: []string{"cl", "cc"},
		DefaultBackend:   "cl",
	}

	result := selectAvailableBackend("cc", cfg)
	if result != "cc" {
		t.Errorf("expected selectAvailableBackend to return %q, got %q", "cc", result)
	}
}

// TestSelectAvailableBackendFallback verifies that fallback is used when chosen is unhealthy.
func TestSelectAvailableBackendFallback(t *testing.T) {
	// cc is unhealthy (env var not set)
	os.Unsetenv("TEST_CC_KEY_FALLBACK") //nolint:errcheck

	// cl is healthy via httptest server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cfg := &Config{
		Backends: map[string]Backend{
			"cc": {Command: "claude", HealthCheckType: "env_var", HealthCheckValue: "TEST_CC_KEY_FALLBACK"},
			"cl": {Command: "ollama", HealthCheckType: "http", HealthCheckValue: srv.URL},
		},
		FallbackBackends: []string{"cl", "cc"},
		DefaultBackend:   "cl",
	}

	result := selectAvailableBackend("cc", cfg)
	if result != "cl" {
		t.Errorf("expected selectAvailableBackend to return %q (fallback), got %q", "cl", result)
	}
}

// TestCheckAllHealth verifies that checkAllHealth returns correct status for each backend.
func TestCheckAllHealth(t *testing.T) {
	t.Setenv("TEST_PRESENT_KEY", "somevalue")
	os.Unsetenv("TEST_ABSENT_KEY") //nolint:errcheck

	cfg := &Config{
		Backends: map[string]Backend{
			"present": {Command: "foo", HealthCheckType: "env_var", HealthCheckValue: "TEST_PRESENT_KEY"},
			"absent":  {Command: "bar", HealthCheckType: "env_var", HealthCheckValue: "TEST_ABSENT_KEY"},
		},
		DefaultBackend: "present",
	}

	results := checkAllHealth(cfg)

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
	if !results["present"] {
		t.Error("expected 'present' backend to be healthy, got false")
	}
	if results["absent"] {
		t.Error("expected 'absent' backend to be unhealthy, got true")
	}
}
