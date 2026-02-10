# ao CLI Reference

> Complete command reference for the `ao` CLI tool.

## Global Flags

These flags are available on all commands:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--config` | string | `~/.agentops/config.yaml` | Config file path |
| `--dry-run` | bool | false | Show what would happen without executing |
| `-h, --help` | bool | false | Help for the command |
| `-o, --output` | string | `table` | Output format (json, table, yaml) |
| `-v, --verbose` | bool | false | Enable verbose output |

---

## Getting Started

### ao demo

Run an interactive demonstration of AgentOps capabilities.

**Usage:** `ao demo [flags]`

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--concepts` | bool | false | Just explain core concepts |
| `--quick` | bool | false | 2-minute quick overview |

**Examples:**

```bash
ao demo              # Interactive walkthrough
ao demo --quick      # 2-minute overview
ao demo --concepts   # Just explain concepts
```

---

### ao quick-start

Initialize AgentOps in your current project.

**Usage:** `ao quick-start [flags]`

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--minimal` | bool | false | Minimal setup (just directories) |
| `--no-beads` | bool | false | Skip beads initialization |

**Examples:**

```bash
ao quick-start              # Full setup with beads
ao quick-start --no-beads   # Skip beads initialization
ao quick-start --minimal    # Just .agents/ structure
```

---

### ao init

Create the `.agents/ao` directory structure for knowledge storage.

**Usage:** `ao init [flags]`

Creates:
- `.agents/ao/sessions/` -- Session markdown and JSONL files
- `.agents/ao/index/` -- Session index for quick lookup
- `.agents/ao/provenance/` -- Provenance tracking graph

**Examples:**

```bash
ao init
```

---

## Core Commands

### ao status

Display the current state of the AgentOps knowledge base.

**Usage:** `ao status [flags]`

Shows session counts, recent sessions, provenance statistics, and storage locations.

**Examples:**

```bash
ao status
ao status -o json
```

---

### ao version

Display version, build information, and runtime details.

**Usage:** `ao version [flags]`

**Examples:**

```bash
ao version
```

---

### ao config

View and manage AgentOps configuration.

**Usage:** `ao config [flags]`

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--show` | bool | false | Show resolved configuration with sources |

Configuration priority (highest to lowest):
1. Command-line flags
2. Environment variables (`AGENTOPS_*`)
3. Project config (`.agentops/config.yaml`)
4. Home config (`~/.agentops/config.yaml`)
5. Defaults

**Environment Variables:**

| Variable | Description |
|----------|-------------|
| `AGENTOPS_OUTPUT` | Default output format (table, json, yaml) |
| `AGENTOPS_BASE_DIR` | Data directory path |
| `AGENTOPS_VERBOSE` | Enable verbose output (true/1) |
| `AGENTOPS_NO_SC` | Disable Smart Connections (true/1) |

**Examples:**

```bash
ao config --show           # Show resolved configuration
ao config --show -o json   # Output as JSON
```

---

### ao badge

Display a visual badge showing knowledge flywheel health status.

**Usage:** `ao badge [flags]`

Status levels:
- **ESCAPE VELOCITY** -- knowledge compounds (sigma x rho > delta)
- **APPROACHING** -- almost there (sigma x rho > delta x 0.8)
- **BUILDING** -- making progress (sigma x rho > delta x 0.5)
- **STARTING** -- early stage (sigma x rho <= delta x 0.5)

**Examples:**

```bash
ao badge
```

---

### ao completion

Generate shell completion scripts.

**Usage:** `ao completion [bash|zsh|fish]`

**Examples:**

```bash
ao completion bash > /etc/bash_completion.d/ao
ao completion zsh > "${fpath[1]}/_ao"
ao completion fish > ~/.config/fish/completions/ao.fish
```

---

## Knowledge Flywheel

### ao inject

Inject relevant knowledge into session context.

**Usage:** `ao inject [context] [flags]`

Searches recent learnings, active patterns, and session summaries using Two-Phase retrieval (freshness + utility scoring). CASS integration adds maturity weighting and confidence decay.

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--apply-decay` | bool | false | Apply confidence decay before ranking |
| `--context` | string | | Context query for filtering (alternative to positional arg) |
| `--format` | string | `markdown` | Output format: markdown, json |
| `--max-tokens` | int | 1500 | Maximum tokens to output |
| `--no-cite` | bool | false | Disable citation recording |
| `--session` | string | | Session ID for citation tracking (auto-generated if empty) |

**Examples:**

```bash
ao inject                     # Inject general knowledge
ao inject "authentication"    # Inject knowledge about auth
ao inject --max-tokens 2000   # Larger budget
ao inject --format json       # JSON output
ao inject --apply-decay       # Apply confidence decay before ranking
```

---

### ao search

Search AgentOps knowledge base.

**Usage:** `ao search <query> [flags]`

Searches markdown and JSONL files in `.agents/ao/sessions/` by default. Optionally uses Smart Connections for semantic search or CASS for session-aware search.

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--cass` | bool | false | Enable CASS session-aware search with maturity weighting |
| `--limit` | int | 10 | Maximum results to return |
| `--type` | string | | Filter by type: decisions, knowledge, sessions |
| `--use-sc` | bool | false | Enable Smart Connections semantic search (requires Obsidian) |

**Examples:**

```bash
ao search "mutex pattern"
ao search "authentication" --limit 20
ao search "database migration" --type decisions
ao search "config" --use-sc
ao search "auth" --cass
```

---

### ao flywheel

Knowledge flywheel operations and status.

**Usage:** `ao flywheel [command]`

**Subcommands:**

| Command | Description |
|---------|-------------|
| `nudge` | Combined flywheel + ratchet + pool status for hooks |
| `status` | Show flywheel health status |

**Examples:**

```bash
ao flywheel status
ao flywheel status -o json
```

---

### ao metrics

Track and report on knowledge flywheel metrics.

**Usage:** `ao metrics [command]`

**Subcommands:**

| Command | Description |
|---------|-------------|
| `baseline` | Capture current flywheel state |
| `cite` | Record a citation event |
| `report` | Show flywheel metrics report |

The flywheel equation: `dK/dt = I(t) - d*K + s*r*K - B(K, K_crit)`

Escape velocity: `s * r > d` means knowledge compounds.

**Examples:**

```bash
ao metrics report
ao metrics baseline
ao metrics cite
```

---

## Forge / Temper / Store Pipeline

### ao forge

Extract knowledge from sources.

**Usage:** `ao forge [command]`

**Subcommands:**

| Command | Description |
|---------|-------------|
| `batch` | Process multiple transcripts at once |
| `transcript` | Extract knowledge from Claude Code transcripts |

**Examples:**

```bash
ao forge transcript ~/.claude/projects/**/*.jsonl
ao forge batch
```

---

### ao temper

Validate and lock knowledge artifacts (TEMPER phase).

**Usage:** `ao temper [command]`

**Subcommands:**

| Command | Description |
|---------|-------------|
| `lock` | Lock validated artifacts (engage ratchet) |
| `status` | Show tempered vs pending artifacts |
| `validate` | Validate artifact structure |

**Examples:**

```bash
ao temper validate
ao temper lock
ao temper status
```

---

### ao store

Index artifacts for retrieval and search (STORE phase).

**Usage:** `ao store [command]`

**Subcommands:**

| Command | Description |
|---------|-------------|
| `index` | Add files to search index |
| `rebuild` | Rebuild search index |
| `search` | Search the index |
| `stats` | Show index statistics |

**Examples:**

```bash
ao store index .agents/ao/sessions/*.md
ao store search "authentication"
ao store rebuild
ao store stats
```

---

### ao extract

Check for pending session extractions and output a prompt for Claude to process.

**Usage:** `ao extract [flags]`

Designed to be called from a SessionStart hook. Outputs a structured prompt asking Claude to extract learnings from pending sessions.

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--all` | bool | false | Process all pending entries |
| `--clear` | bool | false | Clear pending queue without processing |
| `--max-content` | int | 3000 | Maximum characters of session content to include |

**Examples:**

```bash
ao extract                    # Process most recent pending extraction
ao extract --all              # Process all pending extractions
ao extract --clear            # Clear pending queue without processing
ao extract --all --dry-run    # Preview what would be processed
```

---

## Quality Pools and Gates

### ao pool

Manage knowledge candidates in quality pools.

**Usage:** `ao pool [command]`

Pools organize candidates by processing status: pending, staged, promoted, rejected.

**Subcommands:**

| Command | Description |
|---------|-------------|
| `auto-promote` | Auto-promote silver candidates older than threshold |
| `batch-promote` | Bulk promote pending candidates to knowledge base |
| `list` | List candidates in pools |
| `promote` | Promote candidate to knowledge base |
| `reject` | Reject candidate |
| `show` | Show candidate details |
| `stage` | Stage candidate for promotion |

**Examples:**

```bash
ao pool list --tier=gold
ao pool show <candidate-id>
ao pool stage <candidate-id>
ao pool promote <candidate-id>
```

---

### ao gate

Manage human review gates for bronze-tier candidates.

**Usage:** `ao gate [command]`

Bronze-tier candidates (score 0.50-0.69) require human review before promotion.

**Subcommands:**

| Command | Description |
|---------|-------------|
| `approve` | Approve candidate for promotion |
| `bulk-approve` | Bulk approve silver candidates |
| `pending` | List candidates pending review |
| `reject` | Reject candidate |

**Examples:**

```bash
ao gate pending
ao gate approve <candidate-id>
ao gate reject <candidate-id> --reason="Too vague"
```

---

## Provenance and Tracing

### ao trace

Trace the provenance of an artifact back to its source transcript.

**Usage:** `ao trace <artifact-path> [flags]`

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--graph` | bool | false | Show ASCII provenance graph |

**Examples:**

```bash
ao trace .agents/ao/sessions/2026-01-20-my-session.md
ao trace .agents/ao/sessions/*.md --graph
ao trace session-abc123 -o json
```

---

## Ratchet Workflow

### ao ratchet

Track progress through the RPI (Research-Plan-Implement) workflow.

**Usage:** `ao ratchet [command]`

The Brownian Ratchet ensures progress cannot be lost: Chaos x Filter = Ratchet = Progress.

**Subcommands:**

| Command | Description |
|---------|-------------|
| `check` | Check if step gate is met |
| `find` | Search for artifacts |
| `migrate` | Migrate legacy chain |
| `migrate-artifacts` | Add schema_version to artifacts |
| `next` | Show next pending RPI step |
| `promote` | Record tier promotion |
| `record` | Record step completion |
| `skip` | Record intentional skip |
| `spec` | Get current spec path |
| `status` | Show ratchet chain state |
| `trace` | Trace provenance backward |
| `validate` | Validate step requirements |

The ratchet chain is stored in `.agents/ao/chain.jsonl`.

**Examples:**

```bash
ao ratchet status
ao ratchet status -o json
ao ratchet record --step research
ao ratchet check --step plan
ao ratchet next
```

---

## Session Lifecycle

### ao session

Session lifecycle operations.

**Usage:** `ao session [command]`

**Subcommands:**

| Command | Description |
|---------|-------------|
| `close` | Forge transcript, extract learnings, measure flywheel impact |

**Examples:**

```bash
ao session close
ao session close --session abc123
ao session close --dry-run
ao session close -o json
```

---

### ao session-outcome

Analyze session transcript to derive a composite reward signal (0.0-1.0).

**Usage:** `ao session-outcome [transcript-path] [flags]`

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--output` | string | `text` | Output format: text, json |
| `--session` | string | | Session ID (extracted from transcript if not provided) |

**Positive signals:** tests pass (+0.30), git push (+0.20), git commit (+0.15), beads closed (+0.15), ratchet lock (+0.10), no errors (+0.10).

**Penalties:** test failure (-0.20), exceptions (-0.15), no commits (-0.10).

**Examples:**

```bash
ao session-outcome ~/.claude/projects/*/transcript.jsonl
ao session-outcome --session abc123
ao session-outcome --output json
```

---

## MemRL Feedback Loop

### ao feedback

Record reward feedback for a learning to update its utility value.

**Usage:** `ao feedback <learning-id> [flags]`

Implements the MemRL EMA update rule: `u_{t+1} = (1 - alpha) * u_t + alpha * r`

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--alpha` | float | 0.1 | EMA learning rate |
| `--harmful` | bool | false | Mark as harmful (shortcut for --reward 0.0) |
| `--helpful` | bool | false | Mark as helpful (shortcut for --reward 1.0) |
| `--reward` | float | -1 | Reward value (0.0 to 1.0) |

**Examples:**

```bash
ao feedback L001 --helpful        # Learning was helpful
ao feedback L001 --harmful        # Learning was harmful
ao feedback L001 --reward 0.75    # Partial success
ao feedback L001 --reward 1.0 --alpha 0.2   # Faster learning rate
```

---

### ao feedback-loop

Automatically close the MemRL feedback loop by updating utilities of cited learnings.

**Usage:** `ao feedback-loop [flags]`

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--alpha` | float | 0.1 | EMA learning rate |
| `--citation-type` | string | `retrieved` | Filter citations by type (retrieved, applied, all) |
| `--reward` | float | -1 | Override reward value (0.0-1.0); -1 = compute from transcript |
| `--session` | string | | Session ID to process |
| `--transcript` | string | | Path to transcript for reward computation |

**Examples:**

```bash
ao feedback-loop --session session-20260125-120000
ao feedback-loop --session abc123 --reward 0.85
ao feedback-loop --transcript ~/.claude/projects/*/abc.jsonl
```

---

### ao batch-feedback

Process feedback loop for all sessions with citations but no feedback.

**Usage:** `ao batch-feedback [flags]`

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--days` | int | 7 | Process sessions from the last N days |

**Examples:**

```bash
ao batch-feedback
ao batch-feedback --days 7
ao batch-feedback --dry-run
```

---

## CASS Maturity

### ao maturity

Check and manage CASS maturity levels for learnings.

**Usage:** `ao maturity [learning-id] [flags]`

Maturity stages: `provisional` -> `candidate` -> `established` -> `anti-pattern`

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--apply` | bool | false | Apply maturity transitions |
| `--scan` | bool | false | Scan all learnings for pending transitions |

**Transition Rules:**
- provisional -> candidate: utility >= 0.7 AND reward_count >= 3
- candidate -> established: utility >= 0.7 AND reward_count >= 5 AND helpful > harmful
- any -> anti-pattern: utility <= 0.2 AND harmful_count >= 5
- established -> candidate: utility < 0.5 (demotion)
- candidate -> provisional: utility < 0.3 (demotion)

**Examples:**

```bash
ao maturity L001                    # Check maturity status
ao maturity L001 --apply            # Check and apply transition
ao maturity --scan                  # Scan all learnings
ao maturity --scan --apply          # Apply all pending transitions
```

---

### ao anti-patterns

List learnings that have been marked as anti-patterns.

**Usage:** `ao anti-patterns [flags]`

Anti-patterns are learnings with utility <= 0.2 and harmful_count >= 5.

**Examples:**

```bash
ao anti-patterns
ao anti-patterns -o json
```

---

### ao promote-anti-patterns

Scan learnings and promote those meeting anti-pattern criteria.

**Usage:** `ao promote-anti-patterns [flags]`

**Examples:**

```bash
ao promote-anti-patterns
ao promote-anti-patterns --dry-run
```

---

## Task Integration

### ao task-sync

Import and sync tasks from Claude Code's Task tool to CASS maturity tracking.

**Usage:** `ao task-sync [flags]`

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--promote` | bool | false | Promote completed tasks to learnings |
| `--session` | string | | Filter tasks by session ID |
| `--transcript` | string | | Path to Claude Code transcript |

**Examples:**

```bash
ao task-sync
ao task-sync --transcript ~/.claude/projects/*/abc.jsonl
ao task-sync --session session-20260125
ao task-sync --promote
```

---

### ao task-feedback

Apply task completion signals to CASS feedback loop.

**Usage:** `ao task-feedback [flags]`

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--all` | bool | false | Process all tasks without feedback |
| `--session` | string | | Session ID to process |

**Examples:**

```bash
ao task-feedback --session session-20260125
ao task-feedback --all
```

---

### ao task-status

Show task status and CASS maturity distribution.

**Usage:** `ao task-status [flags]`

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--session` | string | | Filter by session ID |

**Examples:**

```bash
ao task-status
ao task-status --session session-20260125
```

---

## Agent Messaging

### ao mail

Inter-agent messaging for the Agent Farm.

**Usage:** `ao mail [command]`

**Subcommands:**

| Command | Description |
|---------|-------------|
| `send` | Send a message |

**Examples:**

```bash
ao mail send --to mayor --body "Issue complete"
ao mail send --to mayor --body "FARM COMPLETE" --type farm_complete
```

---

### ao inbox

View messages from the Agent Farm.

**Usage:** `ao inbox [flags]`

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--from` | string | | Filter by sender |
| `--limit` | int | 100 | Maximum messages to display (0 for all) |
| `--mark-read` | bool | false | Mark displayed messages as read |
| `--since` | string | | Show messages from last duration (e.g., 5m, 1h) |
| `--unread` | bool | false | Show only unread messages |

**Examples:**

```bash
ao inbox
ao inbox --since 5m
ao inbox --from witness
ao inbox --unread
ao inbox --limit 50
```

---

## Hooks Management

### ao hooks

Manage Claude Code hooks that automate the CASS knowledge flywheel.

**Usage:** `ao hooks [command]`

**Subcommands:**

| Command | Description |
|---------|-------------|
| `init` | Generate hooks configuration |
| `install` | Install hooks to Claude Code settings |
| `show` | Display current hook configuration |
| `test` | Test hooks configuration |

**Examples:**

```bash
ao hooks init
ao hooks install
ao hooks test
ao hooks show
```

---

## Plans Management

### ao plans

Manage the plan manifest at `.agents/plans/manifest.jsonl`.

**Usage:** `ao plans [command]`

**Subcommands:**

| Command | Description |
|---------|-------------|
| `diff` | Show drift between manifest and beads |
| `list` | List all registered plans |
| `register` | Register a plan in the manifest |
| `search` | Search plans by name or project |
| `sync` | Sync manifest with beads (beads is source of truth) |
| `update` | Update a plan's status or metadata |

**Examples:**

```bash
ao plans list
ao plans register --path plans/my-plan.md
ao plans search "auth"
ao plans sync
ao plans diff
```

---

## Migration Utilities

### ao migrate

Migrate existing learnings to include MemRL utility field.

**Usage:** `ao migrate memrl [flags]`

Adds `utility: 0.5` to learnings that do not have it.

**Examples:**

```bash
ao migrate memrl
ao migrate memrl --dry-run
```
