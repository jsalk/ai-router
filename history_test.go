package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

// TestTruncatePromptShort verifies truncatePrompt passes short prompts through unchanged.
// RED gate: remove t.Skip after Wave 2 implements truncatePrompt in history.go.
func TestTruncatePromptShort(t *testing.T) {
	t.Skip("skipping until history.go exists — RED gate: will remove skip and confirm after Wave 2 implements truncatePrompt")

	// Uncomment after history.go exists:
	// result := truncatePrompt("short prompt", 80)
	// if result != "short prompt" {
	// 	t.Errorf("expected %q, got %q", "short prompt", result)
	// }
}

// TestTruncatePromptLong verifies truncatePrompt truncates and appends ellipsis (…).
// RED gate: remove t.Skip after Wave 2 implements truncatePrompt in history.go.
func TestTruncatePromptLong(t *testing.T) {
	t.Skip("skipping until history.go exists — RED gate: will remove skip and confirm after Wave 2 implements truncatePrompt")

	// Uncomment after history.go exists:
	// prompt := strings.Repeat("x", 100)
	// result := truncatePrompt(prompt, 80)
	// // 80 chars + "…" (1 rune, 3 bytes UTF-8) = 81 runes total
	// if len([]rune(result)) != 81 {
	// 	t.Errorf("expected 81 runes, got %d", len([]rune(result)))
	// }
	// if !strings.HasSuffix(result, "…") {
	// 	t.Errorf("expected result to end with '…', got: %q", result)
	// }
}

// TestTruncatePromptExact verifies truncatePrompt does not truncate a prompt at exactly maxLen runes.
// RED gate: remove t.Skip after Wave 2 implements truncatePrompt in history.go.
func TestTruncatePromptExact(t *testing.T) {
	t.Skip("skipping until history.go exists — RED gate: will remove skip and confirm after Wave 2 implements truncatePrompt")

	// Uncomment after history.go exists:
	// prompt := strings.Repeat("y", 80)
	// result := truncatePrompt(prompt, 80)
	// if result != prompt {
	// 	t.Errorf("expected prompt unchanged at exactly maxLen, got: %q", result)
	// }
	// if strings.HasSuffix(result, "…") {
	// 	t.Errorf("expected no ellipsis for exact-length prompt, got: %q", result)
	// }
}

// TestHistoryLogFormat validates the tab-separated log entry format contract.
// Uses pure format logic (no unimplemented functions). This test defines the
// expected format that logDispatch must produce in history.go (Wave 2).
func TestHistoryLogFormat(t *testing.T) {
	// Define expected format: timestamp\tprompt\tbackend\n
	// This mirrors the format logDispatch will produce.
	timestamp := time.Date(2026, 5, 16, 10, 30, 0, 0, time.UTC).Format(time.RFC3339)
	promptText := "test prompt"
	backend := "cc"

	logLine := fmt.Sprintf("%s\t%s\t%s\n", timestamp, promptText, backend)

	// Strip trailing newline before splitting
	parts := strings.Split(strings.TrimSuffix(logLine, "\n"), "\t")
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

// TestHistoryLogPermissions verifies that 0600 is the correct file permission constant
// for the history log file. Tests the os.OpenFile + stat flow without calling logDispatch.
func TestHistoryLogPermissions(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := tmpDir + "/history.log"

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	file.Close()

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
