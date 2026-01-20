# AgentOps Kit Architecture Proposal

> Scaling from solo developer to multi-agent orchestration.

## Problem Statement

Current state:
- 65 skills across 9 kits
- Unknown which skills are actually used
- No clear progression from solo → team → multi-agent
- Domain-kit is a grab-bag (21 skills mixing languages, tools, patterns)

Goal:
- Clear tiered architecture
- Language-agnostic core + modular language plugins
- Usage tracking to prune unused skills
- Works for any developer, any stack

---

## Proposed Architecture

### Tier 1: solo-kit (Any Developer, Any Project)

**Purpose:** Everything a single developer needs. Language-agnostic. Zero configuration.

**Skills:**
| Skill | Purpose | From |
|-------|---------|------|
| `research` | Codebase exploration | core-kit |
| `vibe` | Code validation (all aspects) | vibe-kit |
| `bug-hunt` | Git archaeology for root cause | general-kit |
| `complexity` | Find refactoring targets | general-kit |
| `doc` | Generate documentation | docs-kit |
| `oss-docs` | OSS scaffolding (README, CONTRIBUTING) | general-kit |
| `golden-init` | Initialize repo with best practices | general-kit |

**Hooks (portable):**
| Hook | Trigger | Purpose |
|------|---------|---------|
| `prettier-format` | PostToolUse: Edit *.{ts,tsx,js,jsx} | Auto-format JS/TS |
| `lint-warning` | PostToolUse: Edit | Warn about lint issues |
| `console-log-warn` | PostToolUse: Edit *.{ts,tsx,js,jsx} | Warn about console.log |
| `git-push-review` | PreToolUse: git push | Pause for review |

**Agents:**
| Agent | Purpose |
|-------|---------|
| `code-reviewer` | Quality review (read-only) |
| `security-reviewer` | Security scan (read-only) |

**Install:**
```bash
/plugin install solo-kit@agentops
```

---

### Tier 2: Language Kits (Plug-in Based on Project)

Each kit provides:
- Language standards (reference docs)
- Linting/formatting hooks
- Testing patterns
- Common errors and anti-patterns

#### python-kit
```bash
/plugin install python-kit@agentops
```
- Standards: `references/python.md`
- Hooks: `ruff --fix`, `mypy --check`
- Testing: pytest patterns
- Depends on: solo-kit

#### go-kit
```bash
/plugin install go-kit@agentops
```
- Standards: `references/go.md`
- Hooks: `gofmt`, `golangci-lint`
- Testing: table-driven test patterns
- Prescan: P13 (undocumented error ignores), P14 (error wrapping %v)
- Depends on: solo-kit

#### typescript-kit
```bash
/plugin install typescript-kit@agentops
```
- Standards: `references/typescript.md`
- Hooks: `prettier`, `tsc --noEmit`
- Testing: Jest/Vitest patterns
- Depends on: solo-kit

#### shell-kit
```bash
/plugin install shell-kit@agentops
```
- Standards: `references/shell.md`
- Hooks: `shellcheck`
- Testing: bats patterns
- Depends on: solo-kit

---

### Tier 3: Workflow Kits (Team Collaboration)

#### beads-kit (Issue Tracking)
```bash
/plugin install beads-kit@agentops
```
- Skills: `beads`, `bd-routing`
- Integration: `bd create`, `bd ready`, `bd close`
- For: Teams tracking work across sessions

#### pr-kit (Pull Request Workflows)
```bash
/plugin install pr-kit@agentops
```
- Skills: `pr-research`, `pr-plan`, `pr-implement`, `pr-validate`, `pr-retro`
- For: Open source contributions, team PRs

#### dispatch-kit (Multi-Agent Coordination)
```bash
/plugin install dispatch-kit@agentops
```
- Skills: `mail`, `handoff`, `dispatch`
- For: Handing off work between sessions/agents

---

### Tier 4: Orchestration Kits (Multi-Agent Systems)

#### gastown-kit (Full Orchestration)
```bash
/plugin install gastown-kit@agentops
```
- Skills: `gastown`, `roles`, `crew`, `polecat-lifecycle`
- Roles: Mayor, Crew, Polecat, Witness, Refinery
- For: Enterprise multi-agent workflows

#### crank-kit (Autonomous Execution)
```bash
/plugin install crank-kit@agentops
```
- Skills: `crank`, `implement`, `implement-wave`, `formulate`, `plan`, `product`
- For: Autonomous epic execution
- Depends on: beads-kit

---

## Migration Path

### From everything-claude-code
```bash
/plugin install solo-kit@agentops
/plugin install typescript-kit@agentops  # if JS/TS project
```

### For Solo Developer (General)
```bash
/plugin install solo-kit@agentops
/plugin install python-kit@agentops      # pick your language(s)
/plugin install go-kit@agentops
```

### For Team Workflows
```bash
/plugin install solo-kit@agentops
/plugin install beads-kit@agentops       # issue tracking
/plugin install pr-kit@agentops          # PR workflows
```

### For Multi-Agent Orchestration
```bash
/plugin install solo-kit@agentops
/plugin install beads-kit@agentops
/plugin install crank-kit@agentops
/plugin install gastown-kit@agentops
```

---

## Skill Usage Tracking

### Proposal: Invocation Logging

Add to `~/.claude/hooks/skill-tracker.sh`:
```bash
#!/bin/bash
# Called on Skill invocation
SKILL_NAME="$1"
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
echo "$TIMESTAMP,$SKILL_NAME" >> ~/.claude/telemetry/skill-usage.csv
```

Hook configuration:
```json
{
  "Skill": [{
    "matcher": "*",
    "hooks": [{
      "type": "command",
      "command": "~/.claude/hooks/skill-tracker.sh \"$SKILL_NAME\""
    }]
  }]
}
```

### Analysis
```bash
# Most used skills
cut -d',' -f2 ~/.claude/telemetry/skill-usage.csv | sort | uniq -c | sort -rn

# Skills never used (candidates for removal)
comm -23 <(ls ~/.claude/skills | sort) <(cut -d',' -f2 skill-usage.csv | sort -u)
```

---

## Current Kit Refactoring

### Keep As-Is
- **core-kit** → Split into solo-kit + crank-kit
- **vibe-kit** → Merge into solo-kit (it's essential)
- **general-kit** → Merge into solo-kit
- **beads-kit** → Keep (Tier 3)
- **pr-kit** → Keep (Tier 3)
- **dispatch-kit** → Keep (Tier 3)
- **gastown-kit** → Keep (Tier 4)

### Refactor
- **domain-kit** → Split into:
  - `python-kit`
  - `go-kit`
  - `typescript-kit`
  - `shell-kit`
  - `tekton-kit` (specialized)
  - `container-kit` (specialized)

### Deprecate
- **docs-kit** → Merge `doc` into solo-kit, rest into domain kits

---

## File Structure (Post-Refactor)

```
plugins/
├── solo-kit/                    # Tier 1: Any developer
│   ├── .claude-plugin/
│   │   └── plugin.json
│   ├── skills/
│   │   ├── research/
│   │   ├── vibe/
│   │   ├── bug-hunt/
│   │   ├── complexity/
│   │   ├── doc/
│   │   ├── oss-docs/
│   │   └── golden-init/
│   ├── agents/
│   │   ├── code-reviewer.md
│   │   └── security-reviewer.md
│   ├── hooks/
│   │   └── hooks.json
│   └── README.md
│
├── python-kit/                  # Tier 2: Language
│   ├── .claude-plugin/
│   │   └── plugin.json          # depends: ["solo-kit"]
│   ├── skills/
│   │   └── standards/
│   │       └── references/
│   │           └── python.md
│   ├── hooks/
│   │   └── hooks.json           # ruff, mypy
│   └── README.md
│
├── go-kit/                      # Tier 2: Language
├── typescript-kit/              # Tier 2: Language
├── shell-kit/                   # Tier 2: Language
│
├── beads-kit/                   # Tier 3: Team
├── pr-kit/                      # Tier 3: Team
├── dispatch-kit/                # Tier 3: Team
│
├── crank-kit/                   # Tier 4: Orchestration
├── gastown-kit/                 # Tier 4: Orchestration
│
└── specialized/                 # Domain-specific
    ├── tekton-kit/
    └── container-kit/
```

---

## Success Metrics

1. **Solo developer can start in < 2 min**
   - `git clone && /plugin install solo-kit`

2. **Language kits are modular**
   - Install only what you need
   - No cross-language pollution

3. **Usage tracking shows real adoption**
   - Know which skills to maintain
   - Know which to deprecate

4. **Clear progression path**
   - Solo → add beads → add crank → add gastown
   - Each tier unlocks new capabilities

---

## Next Steps

1. [ ] Create solo-kit from core-kit + vibe-kit + general-kit
2. [ ] Split domain-kit into language kits
3. [ ] Implement skill usage tracking
4. [ ] Update README with tiered install instructions
5. [ ] Deprecate unused skills after 30 days of tracking
