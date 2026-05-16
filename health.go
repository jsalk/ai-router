package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

// healthCheckTimeout is the maximum time allowed for an HTTP health check.
const healthCheckTimeout = 3 * time.Second

// HealthResults maps backend name to availability boolean.
type HealthResults map[string]bool

// checkHealth returns true if backend is currently available.
// For "http" type: performs a GET request with a 3-second timeout; healthy if 2xx response.
// For "env_var" type: checks that the named env var is non-empty; no network call.
// For unknown types: returns true (assume available).
func checkHealth(backend Backend) bool {
	switch backend.HealthCheckType {
	case "http":
		client := &http.Client{Timeout: healthCheckTimeout}
		resp, err := client.Get(backend.HealthCheckValue)
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode >= 200 && resp.StatusCode < 300
	case "env_var":
		return os.Getenv(backend.HealthCheckValue) != ""
	default:
		return true
	}
}

// checkAllHealth returns availability for every backend in config.
func checkAllHealth(config *Config) HealthResults {
	results := make(HealthResults)
	for name, backend := range config.Backends {
		results[name] = checkHealth(backend)
	}
	return results
}

// selectAvailableBackend returns the name of an available backend.
// Prefers chosen; falls back through config.FallbackBackends in order (skipping chosen).
// If no backend is available, prints a diagnostic to stderr listing each backend's
// status and calls os.Exit(1).
// Health is checked once upfront via checkAllHealth to avoid redundant network
// calls in the failure diagnostic path (WR-01).
func selectAvailableBackend(chosen string, config *Config) string {
	results := checkAllHealth(config)

	if results[chosen] {
		return chosen
	}

	for _, name := range config.FallbackBackends {
		if name != chosen && results[name] {
			return name
		}
	}

	// No backends available — print diagnostic and exit.
	fmt.Fprintf(os.Stderr, "error: no backends available\n")
	for name, up := range results {
		status := "up"
		if !up {
			status = "down"
		}
		fmt.Fprintf(os.Stderr, "  %s: %s\n", name, status)
	}
	os.Exit(1)
	return "" // unreachable
}
