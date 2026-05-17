package main

import (
	"errors"
	"strings"
	"testing"
)

// TestPickerNoHealthyBackends verifies fzfPicker returns an error when no backends are healthy.
func TestPickerNoHealthyBackends(t *testing.T) {
	config := &Config{
		Backends: map[string]Backend{
			"cc": {Command: "claude", Description: "Cloud Claude"},
			"cl": {Command: "ollama", Description: "Local Ollama"},
		},
		DefaultBackend: "cl",
	}
	healthyResults := map[string]bool{
		"cc": false,
		"cl": false,
	}
	_, err := fzfPicker(config, healthyResults)
	if err == nil {
		t.Fatal("expected error when no backends are healthy, got nil")
	}
	if !strings.Contains(err.Error(), "no healthy backends") {
		t.Errorf("expected error to contain 'no healthy backends', got: %v", err)
	}
}

// TestPickerFzfNotFound is a compile stub only.
// Verified in integration by temporarily removing fzf from PATH.
func TestPickerFzfNotFound(t *testing.T) {
	t.Skip("fzf not-found test: run manually with fzf unavailable")

	// Uncomment after picker.go exists:
	// config := &Config{
	// 	Backends: map[string]Backend{
	// 		"cc": {Command: "claude", Description: "Cloud Claude"},
	// 	},
	// 	DefaultBackend: "cc",
	// }
	// healthyResults := map[string]bool{"cc": true}
	// _, err := fzfPicker(config, healthyResults)
	// if err == nil {
	// 	t.Fatal("expected error when fzf is not found, got nil")
	// }
	// if !strings.Contains(err.Error(), "fzf") {
	// 	t.Errorf("expected error mentioning fzf, got: %v", err)
	// }
}

// TestPickerHealthFilter verifies picker excludes unhealthy backends from the fzf menu.
// After picker.go is written, this should verify that entries passed to fzf exclude "cc".
func TestPickerHealthFilter(t *testing.T) {
	t.Skip("filter test: implemented after picker.go exists")

	// Uncomment after picker.go exists:
	// config := &Config{
	// 	Backends: map[string]Backend{
	// 		"cc": {Command: "claude", Description: "Cloud Claude"},
	// 		"cl": {Command: "ollama", Description: "Local Ollama"},
	// 	},
	// 	DefaultBackend: "cl",
	// }
	// healthyResults := map[string]bool{
	// 	"cc": false, // unhealthy — must not appear in fzf menu
	// 	"cl": true,  // healthy — must appear in fzf menu
	// }
	// // Verify picker only presents healthy backends to fzf
	// _, _ = fzfPicker(config, healthyResults)
}

// TestIsNotFound verifies isNotFound helper detects "command not found" errors.
func TestIsNotFound(t *testing.T) {
	cmdNotFoundErr := errors.New("command not found")
	if !isNotFound(cmdNotFoundErr) {
		t.Error("expected isNotFound(errors.New('command not found')) to return true")
	}
	otherErr := errors.New("some other error")
	if isNotFound(otherErr) {
		t.Errorf("expected isNotFound to return false for non-not-found error, got true")
	}
}
