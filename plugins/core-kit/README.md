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
| `/plan` | `/plan <goal>` | Create beads issues from a goal (alias for formulate) |
| `/formulate` | `/formulate <goal>` | Create reusable `.formula.toml` templates |
| `/product` | `/product <goal>` | Customer-first PR/FAQ brief |
| `/implement` | `/implement <issue>` | Execute a single beads issue |
| `/implement-wave` | `/implement-wave` | Execute multiple issues in parallel |
| `/crank` | `/crank` | Autonomous epic execution |
| `/retro` | `/retro` | Extract learnings |

## When to Use Which

### Planning Skills (plan vs formulate vs product)

| Skill | Use When | Output | Reusable? |
|-------|----------|--------|-----------|
| **`/product`** | Multi-day work, user-facing impact, unclear "why" | PR/FAQ brief in `.agents/products/` | No |
| **`/formulate`** | Repeatable pattern (auth, CRUD, etc.) | `.formula.toml` template | **Yes** |
| **`/plan`** | One-off decomposition | Beads issues directly | No |

**Decision tree:**

```
Is this a repeatable pattern you'll do again?
├─ Yes → /formulate (creates reusable template)
└─ No
   ├─ Is the "why" unclear or user-facing? → /product first, then /plan
   └─ Otherwise → /plan directly
```

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
/research → /product → /formulate → /crank → /retro
    ↓          ↓           ↓          ↓        ↓
 explore    clarify     template   execute   learn
```

**Shortcuts:**

- Simple feature: `/plan → /implement`
- Quick fix: `/implement` directly
- Repeatable work: `/formulate` once, reuse template

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

## Related Kits

- **vibe-kit** - Validate your implementation
- **beads-kit** - Track issues created by planning
- **dispatch-kit** - Hand off work to other agents
