---
name: vibe
description: >
  This skill should be used when the user asks to "validate code",
  "check semantic faithfulness", "run vibe", "prescan for patterns",
  "validate plugins", or needs L13 semantic verification of code behavior.
allowed-tools: "Read,Bash,Glob,Grep,Write,TodoWrite,Skill"
version: 2.1.0
author: "AI Platform Team"
license: "MIT"
context: fork
agent: code-quality-expert
skills:
  - beads
  - complexity
---

# Vibe - Semantic Code Validation

L13 semantic verification: validates that code does what it claims.

## Overview

**Vibe** combines fast static analysis with deep semantic verification.

```
vibe-check    -> Did the AI DEVELOP well? (git history, L9)
vibe-validate -> Did the AI PRODUCE good code? (semantic analysis, L13)
```

## Modes

| Mode | Command | Purpose |
|------|---------|---------|
| **Standard** | `/vibe <target>` | Full validation (prescan + semantic) |
| **Prescan** | `/vibe-prescan <target>` | Fast static checks only |
| **Semantic** | `/vibe-semantic <target>` | Deep semantic analysis |
| **Plugin** | `/vibe-plugin <target>` | Plugin validation |

## Arguments

| Argument | Purpose | Examples |
|----------|---------|----------|
| `recent` | Files from last commit | Default for CI |
| `all` | Full codebase (sampled) | Deep audit |
| `<path>` | Specific directory or file | Targeted check |

---

## Mode 1: Standard Vibe (`/vibe`)

Full orchestrated validation.

### Execution Flow

```
TodoWrite([
  "Run pre-scan (fast static checks)",
  "Run semantic analysis",
  "Generate validation report",
  "Create beads issues for findings"
])
```

**Phase 1: Pre-Scan**
```bash
~/.claude/skills/vibe/scripts/prescan.sh "$TARGET"
```

Detects 6 patterns: P1, P4, P5, P8, P9, P12 (see `references/patterns.md`).

**Phase 2: Semantic Analysis**

For each file, analyze across 5 dimensions:
- **Docstrings**: Do docstrings match implementation?
- **Names**: Do function names match behavior?
- **Security**: Is there security theater?
- **Pragmatic**: Does code follow pragmatic principles?
- **Slop**: Is there AI-generated slop?

**Phase 3: Generate Report**

Output to (see `references/report-format.md`):
- `reports/vibe-report.json`
- `reports/vibe-junit.xml`
- `.agents/assessments/{date}-vibe-validate-{target}.md`

**Phase 4: Create Issues**

For CRITICAL/HIGH findings, create beads issues when `VIBE_VALIDATE_CREATE_ISSUES=true`.

---

## Mode 2: Prescan Only (`/vibe-prescan`)

Fast static detection - no LLM required.

```bash
~/.claude/skills/vibe/scripts/prescan.sh "$TARGET"
```

| Pattern | Severity | What |
|---------|----------|------|
| P1 | CRITICAL | Phantom modifications |
| P4 | HIGH | TODO/FIXME, commented code |
| P5 | HIGH | CC > 15, functions > 50 lines |
| P8 | HIGH | except: pass, bare except |
| P9 | MEDIUM | Docstring claims vs reality |
| P12 | MEDIUM | Unused functions, unreachable |

Full details: `references/patterns.md`

---

## Mode 3: Semantic Only (`/vibe-semantic`)

Deep LLM-powered analysis.

| Analysis | Prefix | Focus |
|----------|--------|-------|
| Docstrings | FAITH-xxx | Param mismatches, return lies |
| Names | NAME-xxx | validate_*, auth_*, encrypt_* |
| Security | SEC-xxx | Injection, auth bypass, crypto |
| Pragmatic | PRAG-xxx | DRY, orthogonality |
| Slop | SLOP-xxx | Hallucinations, cargo cult |

### Process

1. Read target files
2. For each function, check name/docstring vs implementation
3. Flag mismatches with severity
4. Generate findings report

---

## Mode 4: Plugin Validation (`/vibe-plugin`)

Validate Claude Code plugins (commands, skills, agents).

```bash
~/.claude/scripts/validate-plugin.sh "$TARGET"
```

### Semantic Checks

1. **Description Truthfulness**: Does description match body?
2. **Trigger Accuracy** (skills): Do triggers match capabilities?
3. **Argument Consistency** (commands): Are declared args used?
4. **Progressive Disclosure** (skills): Is content properly layered?
5. **Painted Doors**: Documented features that don't exist?

---

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `VIBE_VALIDATE_CREATE_ISSUES` | `false` | Auto-create beads issues |
| `VIBE_VALIDATE_FAIL_ON` | `critical` | Exit non-zero threshold |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Usage error |
| 2 | CRITICAL findings |
| 3 | HIGH findings |

---

## Examples

```bash
# Full validation of recent changes
/vibe recent

# Fast prescan only
/vibe-prescan services/etl/

# Deep semantic analysis
/vibe-semantic services/gateway --only security,names

# Plugin validation
/vibe-plugin ~/.claude/skills/beads

# CI usage
VIBE_VALIDATE_CREATE_ISSUES=true /vibe recent
```

---

## References

- **Pattern Details**: `references/patterns.md`
- **Report Formats**: `references/report-format.md`
- **Prescan Script**: `scripts/prescan.sh`
