package main

import (
	"os"
	"strings"
	"testing"
	"time"
)

// TestTruncatePromptShort verifies truncatePrompt passes short prompts through unchanged.
func TestTruncatePromptShort(t *testing.T) {
	result := truncatePrompt("short prompt", 80)
	if result != "short prompt" {
		t.Errorf("expected %q, got %q", "short prompt", result)
	}
}

// TestTruncatePromptLong verifies truncatePrompt truncates and appends ellipsis (…).
func TestTruncatePromptLong(t *testing.T) {
	prompt := strings.Repeat("x", 100)
	result := truncatePrompt(prompt, 80)
	// 80 chars + "…" (1 rune, 3 bytes UTF-8) = 81 runes total
	if len([]rune(result)) != 81 {
		t.Errorf("expected 81 runes, got %d", len([]rune(result)))
	}
	if !strings.HasSuffix(result, "…") {
		t.Errorf("expected result to end with '…', got: %q", result)
	}
}

// TestTruncatePromptExact verifies truncatePrompt does not truncate a prompt at exactly maxLen runes.
func TestTruncatePromptExact(t *testing.T) {
	prompt := strings.Repeat("y", 80)
	result := truncatePrompt(prompt, 80)
	if result != prompt {
		t.Errorf("expected prompt unchanged at exactly maxLen, got: %q", result)
	}
	if strings.HasSuffix(result, "…") {
		t.Errorf("expected no ellipsis for exact-length prompt, got: %q", result)
	}
}

// TestHistoryLogFormat validates the tab-separated log entry format written by logDispatchToPath.
func TestHistoryLogFormat(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := tmpDir + "/history.log"

	entry := HistoryEntry{
		Timestamp:  time.Date(2026, 5, 16, 10, 30, 0, 0, time.UTC),
		PromptText: "test prompt",
		Backend:    "cc",
	}
	if err := logDispatchToPath(entry, logPath); err != nil {
		t.Fatalf("logDispatchToPath returned error: %v", err)
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	line := strings.TrimSuffix(string(data), "\n")
	parts := strings.Split(line, "\t")
	if len(parts) != 3 {
		t.Errorf("expected 3 tab-separated parts, got %d: %v", len(parts), parts)
	}
	if parts[0] != "2026-05-16T10:30:00Z" {
		t.Errorf("expected timestamp %q, got %q", "2026-05-16T10:30:00Z", parts[0])
	}
	if parts[1] != "test prompt" {
		t.Errorf("expected prompt %q, got %q", "test prompt", parts[1])
	}
	if !strings.Contains(parts[2], "cc") {
		t.Errorf("expected backend part to contain %q, got %q", "cc", parts[2])
	}
}

// TestHistoryLogPermissions verifies logDispatchToPath creates the history log with 0600 permissions.
func TestHistoryLogPermissions(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := tmpDir + "/history.log"

	entry := HistoryEntry{
		Timestamp:  time.Now(),
		PromptText: "perm check",
		Backend:    "cl",
	}
	if err := logDispatchToPath(entry, logPath); err != nil {
		t.Fatalf("logDispatchToPath returned error: %v", err)
	}

	stat, err := os.Stat(logPath)
	if err != nil {
		t.Fatalf("failed to stat log file: %v", err)
	}
	if stat.Mode().Perm() != 0600 {
		t.Errorf("expected file permissions 0600, got %04o", stat.Mode().Perm())
	}
}

// TestHistoryPathNotBlocking is a placeholder for the full end-to-end XDG path test.
// Integration: verified by running `ai` and checking ~/.local/share/ai-router/history.log.
func TestHistoryPathNotBlocking(t *testing.T) {
	t.Skip("integration: verified by running ai and checking ~/.local/share/ai-router/history.log")
}
