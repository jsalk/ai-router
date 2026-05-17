package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// isNotFound returns true if err indicates the binary was not found in PATH.
func isNotFound(err error) bool {
	return err == exec.ErrNotFound || strings.Contains(err.Error(), "not found")
}

// fzfPicker invokes fzf as a subprocess to let the user select a backend.
// Returns the backend name selected, or an error on failure.
// Only healthy backends (healthy[name] == true) are presented in the menu.
// Menu entries are tab-separated "name\tdescription" so fzf displays only the
// name column (--with-nth=1) while the full line is returned on selection.
func fzfPicker(config *Config, healthy map[string]bool) (string, error) {
	// 1. Build entries for healthy backends only.
	var entries []string
	for name, backend := range config.Backends {
		if healthy[name] {
			entries = append(entries, fmt.Sprintf("%s\t%s", name, backend.Description))
		}
	}

	// 2. No healthy backends — fail immediately without spawning fzf.
	if len(entries) == 0 {
		return "", fmt.Errorf("no healthy backends available")
	}

	// 3. Build fzf command.
	cmd := exec.Command("fzf", "--with-nth=1", "--delimiter=\t", "--no-sort") // #nosec G204 — fixed args, no user input in command

	// 4. Acquire stdin/stdout pipes.
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create fzf stdin pipe: %w", err)
	}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create fzf stdout pipe: %w", err)
	}

	// 5. Let fzf render its TUI on the real terminal.
	cmd.Stderr = os.Stderr

	// 6. Start fzf subprocess.
	if err := cmd.Start(); err != nil {
		if isNotFound(err) {
			return "", fmt.Errorf("fzf not installed; install with: sudo pacman -S fzf (or apt install fzf)")
		}
		return "", fmt.Errorf("failed to start fzf: %w", err)
	}

	// 7. Write menu entries to fzf stdin in a goroutine so cmd.Wait() can proceed
	// without deadlocking on a full pipe buffer (T-05-03-04: stdin close mitigates hang).
	go func() {
		defer stdinPipe.Close()
		for _, entry := range entries {
			fmt.Fprintln(stdinPipe, entry)
		}
	}()

	// 8. Read fzf's selected line from stdout.
	var selected string
	scanner := bufio.NewScanner(stdoutPipe)
	if scanner.Scan() {
		line := scanner.Text()
		// Extract backend name (tab field 0).
		selected = strings.SplitN(line, "\t", 2)[0]
	}

	// 9. Check for scanner error.
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading fzf output: %w", err)
	}

	// 10. Wait for fzf to exit.
	if err := cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 130 {
			// Exit code 130 = user pressed Esc or Ctrl-C in fzf.
			return "", fmt.Errorf("picker cancelled")
		}
	}

	// 11. Empty selection means user cancelled (e.g., fzf closed with no choice).
	if selected == "" {
		return "", fmt.Errorf("picker cancelled")
	}

	// 12. Return the selected backend name (T-05-03-01: caller validates against config map).
	return selected, nil
}
