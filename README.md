# ai-router

A zero-latency AI dispatch layer for multi-backend LLM environments. Routes prompts to the optimal inference endpoint — local GPU, cloud API, or web — based on configurable keyword rules and real-time health checks.

## Overview

`ai-router` eliminates the cognitive overhead of backend selection in heterogeneous LLM stacks. It intercepts each prompt, applies priority-ordered keyword classification, performs live health checks on candidate backends, and dispatches to the highest-quality available endpoint — all with < 100ms overhead on the dispatch path.

Designed for developer workstations running mixed local/cloud inference stacks (Ollama + Claude API + Gemini + Perplexity). Integrates with Fish/bash/zsh via a single shell alias.

## Key Features

- **Priority-based keyword router** — YAML-configurable rules map prompt keywords to backends with numeric priority ordering; first match wins
- **Multi-modal health checks** — HTTP reachability, binary presence in PATH, and env var existence checks before each dispatch
- **Automatic fallback chain** — degraded backend? Routes to next-best available without user intervention
- **`fzf`-powered interactive picker** — `--pick` drops into a fuzzy-select TUI listing all healthy backends with cost/latency tiers
- **XDG-compliant dispatch history** — every routing decision logged to `$XDG_DATA_HOME/ai-router/history.log` with timestamp, truncated prompt, and backend
- **Privacy-aware logging** — prompts containing configured private keywords are redacted to `[private]` in history
- **Zero runtime dependencies** — single statically-linkable Go binary, no daemon, no sidecar

## Quick Start

```bash
# Build
go build -o ~/go/bin/ai-router .

# Install config
mkdir -p ~/.config/ai-router
cp ../ai-router-config.yaml ~/.config/ai-router/config.yaml

# Wire to Fish shell
# ~/.config/fish/functions/ai.fish
function ai
    ~/go/bin/ai-router $argv
end
```

## Common Commands

| Command | What happens |
|---|---|
| `ai "<prompt>"` | Keyword-match → health check → dispatch |
| `ai -v "<prompt>"` | Same, with routing decision printed to stderr |
| `ai --dry-run "<prompt>"` | Print routing decision only, no dispatch |
| `ai --pick "<prompt>"` | Open fzf backend picker before dispatch |
| `ai --config /path/config.yaml "<prompt>"` | Use alternate config file |

### Example Dispatches

```bash
ai "help me debug this segfault"         # → cc  (keyword: debug → Claude)
ai "what's the latest on Rust 2025"      # → perplexity (keyword: latest → web)
ai "private: summarize my notes"         # → cl  (keyword: private → local Ollama)
ai "summarize this 200-page PDF"         # → gemini (keyword: summarize + doc)
ai --pick "do something"                 # fzf menu of healthy backends
ai -v "write a unit test"                # stderr: routing to cc (matched: test)
ai --dry-run "classify this prompt"      # print decision, skip dispatch
```

## Backend Configuration

Backends are defined in `~/.config/ai-router/config.yaml`:

```yaml
backends:
  cc:
    command: claude
    unset_env: [ANTHROPIC_BASE_URL, ANTHROPIC_AUTH_TOKEN, OLLAMA_MODEL]
    description: "Cloud Claude (Anthropic API) — complex coding, multi-step reasoning"
    health_check_type: binary
    health_check_value: claude
    cost_tier: high
    latency_tier: medium

  cl:
    command: ollama
    args: [run, qwen3-code]
    description: "Local Ollama (qwen3-code) — private, fast, offline"
    health_check_type: http
    health_check_value: "http://localhost:11434/"
    cost_tier: free
    latency_tier: fast
```

Supported `health_check_type` values: `http`, `binary`, `env_var`.

## Routing Rules

```yaml
routing_rules:
  - keywords: [code, debug, implement, refactor, test, review]
    backend: cc
    priority: 3

  - keywords: [private, local, offline, sensitive, secret]
    backend: cl
    priority: 2

  - keywords: [search, news, today, current, latest, weather]
    backend: perplexity
    priority: 1       # lower number = evaluated first

default_backend: cl
fallback_backends: [cl, cc]
```

Rules are evaluated in ascending priority order. First keyword match wins. Matching is case-insensitive. Unmatched prompts go to `default_backend`.

## Architecture

```
prompt
  │
  ▼
keyword classifier ──── no match ──── default_backend
  │                                         │
  ▼                                         ▼
health check ──── unhealthy ──── fallback chain
  │                                         │
  ▼                                         ▼
dispatch                               dispatch
```

**[Full user guide →](USER_GUIDE.md)**

## Installation

### Requirements

- Go ≥ 1.21
- fzf (for `--pick` mode): `sudo pacman -S fzf` / `apt install fzf`

### Build & Test

```bash
go build -o ~/go/bin/ai-router .
go test ./...
```

### Verify

```bash
ai-router --dry-run "debug my code"     # → routing to cc
ai-router --dry-run "private notes"     # → routing to cl
ai-router -v "hello"                    # → health check results + routing decision
```

## Project

Part of [multi-llm-stack](https://github.com/jsalk/multi-llm-stack) — a native Linux AI routing stack for RTX 4060 Ti class hardware.
