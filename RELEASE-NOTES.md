# Release Notes

User-friendly highlights for each release. For full technical details, see [CHANGELOG.md](CHANGELOG.md).

## v1.2.0 - Parallel Wave Execution

**Released:** 2026-01-31

Run multiple issues simultaneously. `/crank` just got 3x faster.

### Highlights

- **Parallel waves** - Issues with no dependencies run concurrently (up to 3 at a time)
- **Automatic wave detection** - `bd ready` returns parallelizable work, `/crank` dispatches subagents
- **Context-aware limits** - Capped at 3 agents per wave to prevent context overflow

### The Model

```
/plan creates:     issue-1 (no deps)
                   issue-2 (no deps)
                   issue-3 (blocked by 1, 2)

/crank executes:   Wave 1: [issue-1, issue-2] → parallel
                   Wave 2: [issue-3] → after wave 1
```

Dependencies form waves. Waves run in parallel. Simple.

### What's Coming

This is single-session parallelism via subagents. **Olympus** will bring true multi-session orchestration with persistent workers and direct context management.

---

## v1.1.0 - Agent Farm

**Released:** 2026-01-26

Multi-agent execution arrives. Spawn a farm of Claude agents to parallelize work across your epic.

### Highlights

- **`/farm` skill** - Parallel multi-agent execution with `ao farm start --agents N`
- **Witness monitoring** - Background observer tracks agent progress
- **Agent messaging** - `ao inbox` and `ao mail send` for inter-agent communication
- **Circuit breaker** - Auto-stops if >50% agents fail

---

## v0.4.0 - Professional Polish

**Released:** 2026-01-25

Repository restructured for cleaner organization and better discoverability.

### Highlights

- **Minimal README** - 47 lines vs 350. One install, 4 skills, done.
- **Simplified structure** - Reduced root directories from 22 to 13
- **Progressive disclosure** - Start with solo-kit, discover more as needed
- **Multi-platform support** - Added Codex and OpenCode setup guides
- **Session hooks** - Auto-creates `.agents/` directories on startup

### Migration Notes

No breaking changes. The `plugins/` directory (core architecture) is unchanged.

---

## v0.3.1 - Standardized Paths

**Released:** 2026-01-24

### Highlights

- **Portable paths** - All skills now use relative `.agents/` instead of hardcoded paths
- **RAG formatting standard** - Knowledge artifacts optimized for retrieval (200-400 char sections)
- **Mermaid diagrams** - README uses GitHub-native diagrams instead of ASCII art

---

## v0.2.0 - Skill Context Fix

**Released:** 2026-01-24

### Highlights

- **Fixed conversation awareness** - Skills like `/vibe` and `/crank` now see chat context
- **Two-tier standards** - Tier 1 (quick, ~5KB) vs Tier 2 (deep, ~20KB) language validation
- **Marketplace release skill** - Workflow for releasing Claude Code plugins

---

## v0.1.2 - Tiered Architecture

**Released:** 2026-01-20

### Highlights

- **4-tier plugin system**:
  - Tier 1: `solo-kit` - Any developer, any project
  - Tier 2: Language kits (python, go, typescript, shell)
  - Tier 3: Team workflows (beads-kit, pr-kit)
  - Tier 4: Multi-agent (crank-kit, gastown-kit)

- **solo-kit** - 7 skills for solo developers with zero dependencies

---

## v0.1.0 - Unix Philosophy

**Released:** 2026-01-19

### Highlights

- **8 focused kits** following Unix philosophy (do one thing well):
  - core-kit, vibe-kit, docs-kit, beads-kit
  - dispatch-kit, pr-kit, gastown-kit, domain-kit

- **50 total skills** across all kits

---

## v1.0.0 - Initial Release

**Released:** 2025-11-10

First public release with core-workflow, devops-operations, and software-development plugins.

---

## Upgrade Path

```
solo-kit → language-kit → beads-kit → core-kit → gastown-kit
   │            │             │           │            │
   └── Tier 1 ──┴── Tier 2 ───┴── Tier 3 ─┴── Tier 4 ──┘
```

**Recommended starting point:** `solo-kit` for any project.
