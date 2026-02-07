<div align="center">

# AgentOps

[![License: Apache-2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Skills](https://img.shields.io/badge/Skills-npx%20skills-7c3aed)](https://skills.sh/)
[![Claude Code](https://img.shields.io/badge/Claude_Code-Plugin-blueviolet)](https://github.com/anthropics/claude-code)

**Your AI agent forgets everything between sessions. AgentOps fixes that.**

Cross-session memory, validation gates, and orchestrated parallel execution for AI coding agents.

</div>

## Install

```bash
npx skills@latest add boshu2/agentops --all -g
```

Or: `claude plugin add boshu2/agentops`

## See It Work

**Catch risks before you code:**
```
> /pre-mortem "add OAuth integration"

[council] Spawning 2 judges: pragmatist, skeptic
[pragmatist] RISK: Token storage — localStorage is XSS-vulnerable
[skeptic] RISK: Refresh token rotation — silent failure on expiry
Verdict: 3 risks found, 1 critical → fix before implementing
```

**Validate before you commit:**
```
> /vibe

[toolchain] gitleaks ✓  semgrep ✓  shellcheck ✓
[council] Spawning 2 judges...
[complexity] All functions within threshold
Verdict: PASS — no critical findings, ready to ship
```

**Parallelize with fresh context per agent:**
```
> /swarm

[swarm] Creating team with 3 workers...
[worker-1] ✓ implement-auth (2m 14s)
[worker-2] ✓ add-rate-limiter (1m 48s)
[worker-3] ✓ update-tests (3m 02s)
All 3 tasks complete. Run /vibe to validate.
```

## What You Get

- **Cross-session memory** — learnings persist to `.agents/` and inject into future sessions (`/inject`, automatic with `ao` CLI)
- **Validation gates** — multi-aspect code review (security, complexity, architecture, and more) that blocks bad merges (`/vibe`)
- **Parallel execution** — fresh-context agents work simultaneously without stepping on each other (`/swarm`)
- **Shift-left risk analysis** — simulate failures before writing code (`/pre-mortem`)
- **Progress locks** — once a gate passes, it stays passed — no regression (`/ratchet`)
- **Autonomous execution** — orchestrate multi-issue work across waves of parallel agents (`/crank`)

## Skills

| Skill | What it does |
|-------|-------------|
| `/pre-mortem` | Simulate failures before coding |
| `/vibe` | Complexity analysis + multi-model validation gate |
| `/swarm` | Parallel agents with fresh context |
| `/crank` | Autonomous epic execution (orchestrates `/swarm` waves) |
| `/council` | Multi-model consensus — validate, brainstorm, research |
| `/implement` | Single issue, full lifecycle |
| `/post-mortem` | Extract learnings for future sessions |
| `/research` | Deep codebase exploration |
| `/plan` | Break a goal into tracked issues |
| `/release` | Pre-flight, changelog, version bumps, tag |
| `/ratchet` | Progress gates that lock |

<details>
<summary>More skills</summary>

| Skill | What it does |
|-------|-------------|
| `/retro` | Quick retrospective |
| `/inject` | Manually load prior knowledge |
| `/knowledge` | Query knowledge base |
| `/bug-hunt` | Root cause analysis |
| `/complexity` | Code complexity metrics |
| `/doc` | Documentation generation |
| `/standards` | Language-specific coding rules |
| `/beads` | Git-native issue tracking |
</details>

## How It Works

| Pattern | What it solves |
|---------|---------------|
| **Fresh context per agent** | Context bloat degrades performance — each agent starts clean |
| **Validation gates** | Work regresses silently — must pass `/vibe` to commit |
| **Orchestrated execution** | Multiple agents cause chaos — one orchestrator owns the loop, agents work atomically |
| **Compounding memory** | Same bugs rediscovered — `/post-mortem` → `.agents/` → next session |

See [docs/reference.md](docs/reference.md) for architecture diagrams, execution modes, and the full pipeline.

## Going Deeper

**Optional: `ao` CLI** — adds session hooks that auto-inject prior knowledge at session start and auto-extract learnings at session end. Without it, you run `/inject` manually.

```bash
brew tap boshu2/agentops https://github.com/boshu2/homebrew-agentops
brew install agentops
ao hooks install
```

**Optional: `beads`** — git-native issue tracking. Lets `/crank` orchestrate multi-issue work from a tracked backlog.

See [docs/reference.md](docs/reference.md) for per-agent install options, CLI reference, execution modes (local vs distributed), and tool dependencies.

## FAQ

**Why not just use Claude Code directly?**
Claude Code has agent spawning built in. AgentOps adds what it lacks:
- Cross-session memory (agents forget everything when the session ends)
- Codified patterns (isolation, validation contracts, debate protocol) that agents won't discover on their own
- Cross-vendor validation (`--mixed` mode adds Codex judges alongside Claude)

**What data leaves my machine?**
Nothing. All state lives in `.agents/` (git-tracked, local). No telemetry, no cloud sync.

**How do I uninstall?**
```bash
npx skills@latest remove boshu2/agentops -g
```

## Built On

These ideas shaped AgentOps — they're baked in, not extra dependencies.

| Project | What we use it for |
|---------|-------------------|
| [Fresh context per agent](https://ghuntley.com/ralph/) | Each spawned agent gets clean context — no bleed-through |
| [Validation gates that lock](https://github.com/dlorenc/multiclaude) | Once a stage passes, it stays passed — no regression |
| [beads](https://github.com/steveyegge/beads) | Git-native issue tracking |
| [MemRL](https://arxiv.org/abs/2502.06173) | Two-phase retrieval for cross-session memory |

## License

Apache-2.0 · [Documentation](docs/) · [Reference](docs/reference.md) · [Changelog](CHANGELOG.md)
