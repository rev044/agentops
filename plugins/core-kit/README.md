# Core Kit

The complete workflow from exploration to execution. 8 skills covering research, planning, and implementation.

## Install

```bash
/plugin install core-kit@boshu2-agentops
```

## Skills

| Skill | Invoke | Purpose |
|-------|--------|---------|
| `/research` | `/research <topic>` | Deep codebase exploration |
| `/formulate` | `/formulate <goal>` | Create beads issues (and optional `.formula.toml` templates) |
| `/product` | `/product <goal>` | Customer-first PR/FAQ brief |
| `/implement` | `/implement <issue>` | Execute a single beads issue |
| `/implement-wave` | `/implement-wave` | Execute multiple issues in parallel |
| `/crank` | `/crank` | Autonomous epic execution |
| `/retro` | `/retro` | Extract learnings |

## When to Use Which

### Planning Skills (formulate vs product)

| Skill | Use When | Output | Reusable? |
|-------|----------|--------|-----------|
| **`/product`** | Multi-day work, user-facing impact, unclear "why" | PR/FAQ brief in `.agents/products/` | No |
| **`/formulate`** | Break work into beads issues | Beads issues (+ optional `.formula.toml` template) | Optional |

**Decision tree:**

```
Is the "why" unclear or user-facing?
├─ Yes → /product first, then /formulate
└─ No → /formulate directly
```

**Note:** For planning approach decisions, use Claude's built-in plan mode (`EnterPlanMode`).

### Implementation Skills (implement vs implement-wave vs crank)

| Skill | Scope | Human Involvement | When to Use |
|-------|-------|-------------------|-------------|
| **`/implement`** | Single issue | High - review each | Bug fixes, learning, unfamiliar code |
| **`/implement-wave`** | Wave of issues | Medium - review waves | Independent features, parallelizable work |
| **`/crank`** | Entire epic | None - fully autonomous | Well-planned epics, trusted decomposition |

**Key differences:**

- **`/implement <issue-id>`** - Works on ONE beads issue. You see the work, review it, then continue. Best for:
  - First time working in a codebase
  - Complex bugs requiring investigation
  - Learning how the system works
  - When you want to stay in control

- **`/implement-wave`** - Executes a "wave" of independent issues in parallel. Checkpoints between waves. Best for:
  - Multiple features that don't conflict
  - Batch refactoring (rename X in 10 files)
  - When issues are well-isolated

- **`/crank`** - Runs the ENTIRE epic autonomously until ALL children are CLOSED. No stopping. Best for:
  - Well-decomposed epics with clear acceptance criteria
  - Repeatable patterns (CRUD, migrations)
  - When you trust the planning phase completely

**Decision tree:**

```
Is this a single issue or quick fix?
├─ Yes → /implement (one issue, full control)
└─ No, it's a multi-issue epic
   ├─ Do you need to review between issues?
   │  ├─ Yes → /implement-wave (parallel with checkpoints)
   │  └─ No, run to completion → /crank (fully autonomous)
```

**Trust escalation:**
```
/implement → /implement-wave → /crank
   ↑              ↑              ↑
 learning    comfortable    confident
```

## Standard Workflow

```
/research → plan mode → /formulate → execute → /retro
    ↓           ↓            ↓           ↓         ↓
 explore   approach      beads      run it     learn
```

### The Planning Bridge (Native Plan Mode)

Claude Code's built-in plan mode (`Shift+Tab` x2) handles approach decisions:

1. Enter plan mode after `/research`
2. Claude explores, asks clarifying questions
3. Creates `plan.md` with approach decisions
4. On acceptance: **context clears**, execution begins with just the plan
5. Previous session accessible via `/resume`

**Key benefit:** Fresh context for execution, but full research accessible if needed.

### From Plan to Beads

After plan acceptance, create trackable issues:

```bash
/formulate <goal>           # Creates beads issues from plan
bd ready                    # See what's ready to work
```

### Execution Tiers

| Skill | Scope | Parallelism | When to Use |
|-------|-------|-------------|-------------|
| `/implement` | Single issue | None | Learning, complex bugs, unfamiliar code |
| `/implement-wave` | Wave of issues | Task() subagents | Independent features, batch work |
| `/crank` | Entire epic | Auto-detects mode | Well-planned epics, overnight runs |

**Trust escalation:**
```
/implement → /implement-wave → /crank
   ↑              ↑              ↑
 learning    comfortable    confident
```

### Session Continuity

Claude Code's native session features:

| Command | Purpose |
|---------|---------|
| `/rename <name>` | Name session for later reference |
| `/resume` | Pick from recent sessions (same repo) |
| `--continue` | Continue most recent session |

**Pattern:** Name sessions after phase: `research-oauth`, `implement-wave1-auth`

**Shortcuts:**

- Simple feature: `/formulate → /implement`
- Quick fix: `/implement` directly
- Repeatable work: `/formulate` once, reuse template
- Full auto: `/formulate → /crank`

## Examples

### Research then implement

```bash
/research "authentication flow"
# Creates .agents/research/YYYY-MM-DD-authentication-flow.md

/implement gt-1234
# Executes the issue with research context
```

### Full workflow for new feature

```bash
/product "user dashboard"
# Creates PR/FAQ brief, clarifies requirements

/formulate "dashboard"
# Creates .formula.toml template, generates beads issues

/crank
# Runs autonomously until epic is complete

/retro
# Captures learnings for next time
```

## Philosophy

- **Research before coding** - understand the codebase
- **Plan before implementing** - break work into manageable pieces
- **Iterate and learn** - capture insights for continuous improvement

## Configuration

### plansDirectory (Optional)

Store native plan mode outputs in your project:

```json
// .claude/settings.json
{
  "plansDirectory": ".claude/plans"
}
```

This makes plans versioned with your repo instead of global storage.

**Gas Town pattern:**
```json
// ~/gt/<rig>/crew/boden/.claude/settings.json
{
  "plansDirectory": "../../../.agents/{rig}/plans"
}
```

## Related Kits

- **vibe-kit** - Validate your implementation
- **beads-kit** - Track issues created by planning
- **dispatch-kit** - Hand off work to other agents
