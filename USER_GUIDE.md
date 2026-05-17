# ai-router User Guide

Complete reference for configuring and using ai-router in a multi-backend LLM environment.

## Contents

- [Installation](#installation)
- [Configuration File](#configuration-file)
- [Routing Rules](#routing-rules)
- [Backend Definitions](#backend-definitions)
- [Health Checks](#health-checks)
- [CLI Reference](#cli-reference)
- [Interactive Picker](#interactive-picker)
- [History Log](#history-log)
- [Troubleshooting](#troubleshooting)
- [Advanced: Custom Backends](#advanced-custom-backends)

---

## Installation

### 1. Build the binary

```bash
cd ai-router
go build -o ~/go/bin/ai-router .
```

Verify it works:
```bash
~/go/bin/ai-router --dry-run "hello"
# stderr: routing to cl  (or your configured default)
```

### 2. Install config

```bash
mkdir -p ~/.config/ai-router
cp ai-router-config.yaml ~/.config/ai-router/config.yaml
```

Edit the installed copy to match your backends. The reference config ships with five backends: `cc` (Claude API), `cl` (local Ollama), `claude-ollama-cloud` (Ollama Cloud), `perplexity`, and `gemini`.

### 3. Wire to your shell

**Fish** (`~/.config/fish/functions/ai.fish`):
```fish
function ai --description "Auto-route prompt to correct AI backend"
    ~/go/bin/ai-router $argv
end
```

**Bash / Zsh** (`~/.bashrc` or `~/.zshrc`):
```bash
alias ai='~/go/bin/ai-router'
```

---

## Configuration File

Default location: `~/.config/ai-router/config.yaml`
Override with: `ai-router --config /path/to/other.yaml`

### Top-level keys

| Key | Type | Description |
|---|---|---|
| `backends` | map | Named backend definitions |
| `routing_rules` | list | Keyword → backend mappings |
| `default_backend` | string | Used when no rule matches |
| `fallback_backends` | list | Ordered fallbacks when chosen backend is unhealthy |
| `private_keywords` | list | Keywords that redact the prompt in the history log |

---

## Routing Rules

```yaml
routing_rules:
  - keywords: [code, debug, implement, refactor]
    backend: cc
    priority: 3
```

Rules are evaluated in **ascending priority order** (priority 1 is checked before priority 3). Within a rule, keywords are checked left-to-right; the first match wins.

Matching is **case-insensitive** and uses substring search — partial matches work (`"implement"` matches `"reimplementation"`).

### Resolution order

1. Sort all rules by `priority` ascending
2. For each rule, check if any keyword appears anywhere in the lowercased prompt
3. First match → select that backend
4. No match → use `default_backend`
5. Chosen backend unhealthy → walk `fallback_backends` in order
6. No fallbacks available → print diagnostic and exit 1

---

## Backend Definitions

```yaml
backends:
  my-backend:
    command: some-cli              # executable name or absolute path
    args: [--flag, value]          # args prepended before the prompt
    set_env:                       # env vars set for this invocation only
      MY_KEY: "value"
    unset_env:                     # env vars unset for this invocation only
      - OTHER_KEY
    url_template: "https://example.com/?q={prompt}"  # browser-launch backends
    description: "Shown in --pick menu and -v output"
    health_check_type: http        # http | binary | env_var
    health_check_value: "http://localhost:8080/"
    cost_tier: medium              # free | low | medium | high
    latency_tier: fast             # instant | fast | medium | slow
```

### Dispatch behavior

- If `url_template` is set: `{prompt}` is URL-encoded and substituted; backend is launched via `xdg-open <url>`
- Otherwise: `command [args...] <prompt>` is executed with the modified environment

`set_env` and `unset_env` are scoped to the child process — your shell environment is never modified.

---

## Health Checks

A health check runs before each dispatch. If the chosen backend fails, ai-router walks the fallback chain.

| Type | `health_check_value` field | Passes when |
|---|---|---|
| `http` | URL string | GET returns 2xx within 3 seconds |
| `binary` | Command name | Command found in PATH (`which <value>`) |
| `env_var` | Variable name | Env var is set and non-empty |

### Viewing health check results

```bash
ai -v "hello"
# stderr: health: cc=true cl=true perplexity=false gemini=false
# stderr: routing to cl (default)
```

---

## CLI Reference

```
ai-router [flags] <prompt>

Flags:
  -v              Print routing decision to stderr before dispatching
  --dry-run       Print routing decision to stderr and exit without dispatching
  --pick          Open fzf interactive backend selector before dispatch
  --config path   Use alternate config (default: ~/.config/ai-router/config.yaml)
```

Flags may appear before or after the prompt. The prompt is everything after the last recognized flag, joined with spaces.

---

## Interactive Picker

`--pick` opens an fzf menu listing all **healthy** backends:

```
cc [high/medium]         — Cloud Claude (Anthropic API) — complex coding, reasoning
cl [free/fast]           — Local Ollama (qwen3-code) — private, fast, offline
claude-ollama-cloud [low/fast] — Ollama Cloud Pro — mid-tier coding
```

Navigate with arrow keys or type to fuzzy-filter. Press **Enter** to dispatch to the selected backend. Press **Esc** or **Ctrl-C** to cancel.

**Requires fzf:**
```bash
sudo pacman -S fzf     # Arch Linux
sudo apt install fzf   # Debian / Ubuntu
```

---

## History Log

Every dispatch is appended to `$XDG_DATA_HOME/ai-router/history.log` (typically `~/.local/share/ai-router/history.log`).

**Format:** `<RFC3339 timestamp>\t<prompt preview>\t<backend>`

```
2026-05-17T14:23:01Z    debug this segfault                      cc
2026-05-17T14:25:40Z    [private]                                cl
2026-05-17T15:01:12Z    what's the latest on Rust 2025...        perplexity
```

- Prompts longer than 80 bytes are truncated with `…`
- Prompts containing any `private_keywords` are replaced with `[private]`

**View recent history:**
```bash
tail -20 ~/.local/share/ai-router/history.log
```

**Search by backend:**
```bash
grep $'\tcc$' ~/.local/share/ai-router/history.log | tail -10
```

---

## Troubleshooting

### "config file not found"

```bash
mkdir -p ~/.config/ai-router
cp /path/to/ai-router-config.yaml ~/.config/ai-router/config.yaml
```

### "fzf not installed"

```bash
sudo pacman -S fzf     # Arch
sudo apt install fzf   # Debian/Ubuntu
```

### Routing to wrong backend

Use `-v` to see the decision and which keyword matched:

```bash
ai -v "your prompt here"
# stderr: routing to cc (matched keyword: code)
```

Adjust the relevant rule in `~/.config/ai-router/config.yaml` — change keywords, adjust priority, or add a higher-priority rule.

### Backend always shows as unhealthy

```bash
ai -v "test"
# stderr: health: cc=false cl=true ...
```

Diagnose by check type:
- **`http`**: `curl -v http://localhost:11434/` — is Ollama running?
- **`binary`**: `which claude` — is the binary in PATH?
- **`env_var`**: `echo $ANTHROPIC_API_KEY` — is the var set?

### No backends available (exit 1)

ai-router prints each backend's status before exiting. The chosen backend and all fallbacks are down. Fix at least one backend or add a healthy backend to `fallback_backends`.

---

## Advanced: Custom Backends

Add any command-line tool as a backend in `~/.config/ai-router/config.yaml`:

```yaml
backends:
  llm-cli:
    command: llm
    args: [-m, mistral]
    description: "Simon Willison's llm CLI with mistral"
    health_check_type: binary
    health_check_value: llm
    cost_tier: free
    latency_tier: fast

routing_rules:
  - keywords: [story, creative, fiction, poem]
    backend: llm-cli
    priority: 5
```

### Browser-based backend (URL template)

```yaml
backends:
  kagi:
    command: xdg-open
    url_template: "https://kagi.com/search?q={prompt}"
    description: "Kagi search"
    health_check_type: env_var
    health_check_value: KAGI_SESSION_TOKEN
    cost_tier: low
    latency_tier: slow
```

`{prompt}` is URL-encoded before substitution. The backend command must be `xdg-open` for URL-launch behavior.

### Environment-scoped backends

Use `set_env` / `unset_env` to isolate API keys per invocation without touching your shell environment:

```yaml
backends:
  openai:
    command: llm
    args: [-m, gpt-4o]
    set_env:
      OPENAI_API_KEY: "sk-..."
    unset_env:
      - ANTHROPIC_API_KEY
    description: "OpenAI GPT-4o via llm CLI"
    health_check_type: env_var
    health_check_value: OPENAI_API_KEY
    cost_tier: high
    latency_tier: medium
```
