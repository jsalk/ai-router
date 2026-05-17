package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/adrg/xdg"
)

// HistoryEntry is the data for one routing dispatch log line.
type HistoryEntry struct {
	Timestamp       time.Time
	PromptText      string
	Backend         string
	PrivateKeywords []string
}

// truncatePrompt returns prompt unchanged if len <= maxLen, else truncates to maxLen bytes and appends "…" (U+2026).
func truncatePrompt(prompt string, maxLen int) string {
	if len(prompt) <= maxLen {
		return prompt
	}
	return prompt[:maxLen] + "…"
}

// logDispatch writes one entry to the XDG Data Home history log.
func logDispatch(entry HistoryEntry) error {
	path, err := xdg.DataFile("ai-router/history.log")
	if err != nil {
		return fmt.Errorf("history: resolve path: %w", err)
	}
	return logDispatchToPath(entry, path)
}

// logDispatchToPath writes a history entry to the specified path (testable helper).
func logDispatchToPath(entry HistoryEntry, logPath string) error {
	dir := filepath.Dir(logPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("history: create dir: %w", err)
	}
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("history: open log: %w", err)
	}
	defer file.Close()

	promptPreview := truncatePrompt(entry.PromptText, 80)
	for _, kw := range entry.PrivateKeywords {
		if strings.Contains(strings.ToLower(entry.PromptText), strings.ToLower(kw)) {
			promptPreview = "[private]"
			break
		}
	}

	logLine := fmt.Sprintf("%s\t%s\t%s\n", entry.Timestamp.Format(time.RFC3339), promptPreview, entry.Backend)
	if _, err := file.WriteString(logLine); err != nil {
		return fmt.Errorf("history: write: %w", err)
	}
	if err := file.Sync(); err != nil {
		return fmt.Errorf("history: sync: %w", err)
	}
	return nil
}
