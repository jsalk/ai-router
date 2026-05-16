package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
)

func main() {
	verbose := flag.Bool("v", false, "print routing decision to stderr")
	configPath := flag.String("config", defaultConfigPath(), "path to config.yaml")
	flag.Parse()

	// Require at least one positional argument (the prompt).
	if flag.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "usage: ai-router [-v] [--config path] <prompt>\n")
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

	// Verbose output to stderr (stdout stays clean for scripting, per D-15).
	// Security: never print env var values — only backend names and health booleans (T-04-14).
	if *verbose {
		fmt.Fprintf(os.Stderr, "ai: routing to %s (matched keyword: %s)\n", backendName, matchedKeyword)
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
