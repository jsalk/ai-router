package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
)

func main() {
	cfgDefault, err := defaultConfigPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	verbose := flag.Bool("v", false, "print routing decision to stderr")
	dryRun := flag.Bool("dry-run", false, "print routing decision to stderr and exit without dispatching")
	pick := flag.Bool("pick", false, "interactively select backend via fzf before dispatch")
	configPath := flag.String("config", cfgDefault, "path to config.yaml")
	flag.Parse()

	// Require at least one positional argument (the prompt).
	if flag.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "usage: ai-router [-v] [--dry-run] [--pick] [--config path] <prompt>\n")
		os.Exit(1)
	}

	prompt := strings.Join(flag.Args(), " ")

	// Load and validate config.
	config, err := loadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if err := validateConfig(config); err != nil {
		fmt.Fprintf(os.Stderr, "error: invalid config: %v\n", err)
		os.Exit(1)
	}

	// Sort routing rules by priority ascending so lower numbers win (CR-01).
	sort.Slice(config.RoutingRules, func(i, j int) bool {
		return config.RoutingRules[i].Priority < config.RoutingRules[j].Priority
	})

	// Keyword routing — returns ("", "") when no rule matches.
	backendName, matchedKeyword := routeByKeyword(prompt, config.RoutingRules)
	if backendName == "" {
		backendName = config.DefaultBackend
		matchedKeyword = "default"
	}

	// Health check + fallback selection (calls os.Exit(1) internally if none available).
	backendName = selectAvailableBackend(backendName, config)

	// --pick: interactive fzf backend selector overrides auto-routed backendName.
	if *pick {
		healthResults := checkAllHealth(config)
		selected, err := fzfPicker(config, healthResults)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: picker failed: %v\n", err)
			os.Exit(1)
		}
		// Guard: selected name must be a valid Config.Backends key (T-05-05-01).
		if _, ok := config.Backends[selected]; !ok {
			fmt.Fprintf(os.Stderr, "error: selected backend %q not found in config\n", selected)
			os.Exit(1)
		}
		backendName = selected
	}

	// Verbose output to stderr (stdout stays clean for scripting, per D-15).
	// Security: never print env var values — only backend names and health booleans (T-04-14).
	if *verbose || *dryRun {
		verboseBackend := config.Backends[backendName]
		verboseLine := fmt.Sprintf("ai: routing to %s (matched keyword: %s)", backendName, matchedKeyword)
		var tierParts []string
		if verboseBackend.CostTier != "" {
			tierParts = append(tierParts, fmt.Sprintf("cost=%s", verboseBackend.CostTier))
		}
		if verboseBackend.LatencyTier != "" {
			tierParts = append(tierParts, fmt.Sprintf("latency=%s", verboseBackend.LatencyTier))
		}
		if len(tierParts) > 0 {
			verboseLine += " | " + strings.Join(tierParts, " ")
		}
		fmt.Fprintf(os.Stderr, "%s\n", verboseLine)

		results := checkAllHealth(config)

		// Sort names for deterministic output.
		names := make([]string, 0, len(results))
		for name := range results {
			names = append(names, name)
		}
		sort.Strings(names)

		parts := make([]string, 0, len(results))
		for _, name := range names {
			parts = append(parts, fmt.Sprintf("%s=%v", name, results[name]))
		}
		fmt.Fprintf(os.Stderr, "ai: health: %s\n", strings.Join(parts, ", "))
	}

	// Log dispatch decision (including dry-run — log the decision even without executing, per D-07).
	// Non-fatal: a log failure prints a warning but does not block dispatch (T-05-05-05).
	if err := logDispatch(HistoryEntry{
		Timestamp:       time.Now(),
		PromptText:      prompt,
		Backend:         backendName,
		PrivateKeywords: config.PrivateKeywords,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to log history: %v\n", err)
	}

	// Dry-run: print routing decision and exit without dispatching (WR-05).
	if *dryRun {
		os.Exit(0)
	}

	// Build command and dispatch (T-04-13: exec.Command, NOT sh -c to prevent shell injection).
	backend := config.Backends[backendName]
	binary, args := buildCommand(backend, prompt)

	cmd := exec.Command(binary, args...) // #nosec G204 — binary comes from user-owned config file (T-04-15: accept)
	cmd.Env = setupEnv(backend)          // Fresh env copy — child process only, not parent shell
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "error: backend exited with error: %v\n", err)
		os.Exit(1)
	}
}
