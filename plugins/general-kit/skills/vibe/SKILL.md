---
name: vibe
description: >
  Talos-class comprehensive code validation. Use for "validate code",
  "run vibe", "check quality", "security review", "architecture review",
  "accessibility audit", "complexity check", or any validation need.
  One skill to validate them all.
allowed-tools: "Read,Bash,Glob,Grep,Write,TodoWrite,Task"
version: 3.0.0
author: "AI Platform Team"
license: "MIT"
context: fork
context-budget:
  skill-md: 10KB
  references-total: 12KB
  typical-session: 15KB
skills:
  - beads
  - standards
---

# Vibe - Talos Comprehensive Validation

> **Portability Note:** This skill is standalone with no external dependencies.
> For beads integration (issue creation from findings), see `vibe-kit`.

**One skill to validate them all.**

Vibe is a Talos-class validator that combines fast static analysis with deep
semantic verification across all quality dimensions: code quality, security,
architecture, accessibility, complexity, and more.

## Philosophy

> **Mono over Micro**: Instead of chaining small skills, Vibe provides comprehensive
> validation in one invocation. Trade-off: larger context, but simpler mental model
> and guaranteed coverage.

## Quick Start

```bash
/vibe                     # Auto-detect target (recent changes or staged files)
/vibe recent              # Full validation of recent changes
/vibe services/           # Validate a directory
/vibe --fast recent       # Prescan only (no LLM, CI-friendly)
/vibe --security recent   # Security-focused deep dive
/vibe --all-aspects all   # Nuclear option: everything on everything
```

## Argument Inference

When invoked without an explicit target, infer from context:

### Priority 1: Conversational Context

If the user mentions a topic, file, or directory in the same message (e.g., "/vibe the auth changes"),
use that as the target:

```bash
# User said "/vibe the auth changes" -> validate auth-related files
git diff --name-only | grep -i auth
# Or search for auth directory
find . -type d -name "*auth*" | head -1
```

**Extract keywords** from the user's message and match against changed files or directories.

### Priority 2: Git State Discovery

```bash
# 1. Check for staged changes
STAGED=$(git diff --cached --name-only 2>/dev/null | head -20)
if [[ -n "$STAGED" ]]; then
    TARGET="staged"
    echo "[VIBE] Auto-selected target: staged changes"
    echo "$STAGED" | head -5
    exit 0
fi

# 2. Check for unstaged changes
UNSTAGED=$(git diff --name-only 2>/dev/null | head -20)
if [[ -n "$UNSTAGED" ]]; then
    TARGET="recent"
    echo "[VIBE] Auto-selected target: recent changes (unstaged)"
    echo "$UNSTAGED" | head -5
    exit 0
fi

# 3. Check for recent commits (last 24h)
RECENT_COMMITS=$(git log --since="24 hours ago" --oneline 2>/dev/null | head -5)
if [[ -n "$RECENT_COMMITS" ]]; then
    TARGET="recent"
    echo "[VIBE] Auto-selected target: recent commits"
    echo "$RECENT_COMMITS"
    exit 0
fi

# 4. No changes found - ask user
echo "[VIBE] No recent changes detected. Please specify a target:"
echo "  /vibe services/        # Validate a directory"
echo "  /vibe path/to/file.py  # Validate specific file"
echo "  /vibe all              # Validate entire codebase"
```

**Key**: Conversational keywords > staged > unstaged > recent commits > ask user.

---

## Aspects

Vibe validates across 8 aspects. By default, all aspects run.

| Aspect | Prefix | Focus | Agent |
|--------|--------|-------|-------|
| **Quality** | QUAL-xxx | Code smells, patterns, maintainability | code-quality-expert |
| **Security** | SEC-xxx | OWASP, injection, auth, crypto | security-expert |
| **Architecture** | ARCH-xxx | Boundaries, coupling, scalability | architecture-expert |
| **Accessibility** | A11Y-xxx | WCAG, keyboard, screen reader | ux-expert |
| **Complexity** | CMPLX-xxx | Cyclomatic, cognitive, function size | code-quality-expert |
| **Semantic** | SEM-xxx | Docstrings, names, claims vs reality | code-quality-expert |
| **Performance** | PERF-xxx | N+1, unbounded loops, resource leaks | architecture-expert |
| **Slop** | SLOP-xxx | AI artifacts, hallucinations, cargo cult | code-quality-expert |

### Aspect Selection

```bash
/vibe recent                          # All aspects (default)
/vibe recent --only security          # Security only
/vibe recent --only security,arch     # Security + Architecture
/vibe recent --exclude slop,a11y      # Skip slop and accessibility
/vibe recent --fast                   # Prescan only (no LLM aspects)
```

---

## Modes

| Mode | Flag | LLM | Speed | Use Case |
|------|------|-----|-------|----------|
| **Full** | (default) | Yes | Slow | PR review, pre-merge |
| **Fast** | `--fast` | No | Fast | CI, quick check |
| **Deep** | `--deep` | Yes | Slowest | Audit, new codebase |
| **Security** | `--security` | Yes | Medium | Security-focused |
| **Arch** | `--arch` | Yes | Medium | Architecture review |

### Two-Tier Standards Loading

Vibe uses a two-tier JIT loading strategy for language standards:

| Tier | Location | Size | Loaded When |
|------|----------|------|-------------|
| **Tier 1** | `standards/references/*.md` | ~4-5KB | Always (via standards skill) |
| **Tier 2** | `vibe/references/*-standards.md` | ~15-25KB | With `--deep` flag |

**Tier 1 (Quick Reference):** Slim refs (~150 lines) with:
- Quick reference tables
- Common errors and anti-patterns
- Summary checklist
- Prescan checks

**Tier 2 (Deep Standards):** Comprehensive standards (~400-1000 lines) with:
- Full table of contents
- Detailed patterns and examples
- Project structure guides
- Compliance assessment with grading scale

**Languages Covered:** Python, TypeScript, Shell, Go, YAML, JSON, Markdown

```bash
/vibe recent                  # Tier 1 only (default)
/vibe --deep recent           # Tier 1 + Tier 2 (comprehensive audit)
/vibe --deep all              # Full codebase audit with all standards
```

**When to use --deep:**
- New codebase onboarding
- Security/compliance audits
- Architecture reviews
- Training new team members

---

## Execution Flow

### Phase 1: Prescan (Static, No LLM)

Fast pattern detection using static analysis tools.

```bash
./scripts/prescan.sh "$TARGET"
```

**Patterns Detected:**

| ID | Pattern | Severity | Tool |
|----|---------|----------|------|
| P1 | Phantom modifications | CRITICAL | git diff |
| P2 | Hardcoded secrets | CRITICAL | gitleaks/grep |
| P3 | SQL injection patterns | CRITICAL | regex |
| P4 | TODO/FIXME/commented code | HIGH | grep |
| P5 | Cyclomatic complexity >15 | HIGH | radon/gocyclo |
| P6 | Functions >50 lines | HIGH | wc/ast |
| P7 | Bare except/empty catch | HIGH | ast/shellcheck |
| P8 | Unused imports/functions | MEDIUM | ast |
| P9 | Docstring mismatches | MEDIUM | ast |
| P10 | Missing error handling | MEDIUM | ast |

**Exit Codes:** 0=clean, 2=CRITICAL, 3=HIGH

### Phase 2: Semantic Analysis (LLM-Powered)

Deep analysis requiring semantic understanding.

**Quality (QUAL-xxx):**
- QUAL-001: Dead code paths
- QUAL-002: Inconsistent naming
- QUAL-003: Magic numbers/strings
- QUAL-004: Missing tests for complexity
- QUAL-005: Copy-paste with slight changes

**Security (SEC-xxx):**
- SEC-001: Injection vulnerabilities (SQL, command, XSS)
- SEC-002: Authentication bypass potential
- SEC-003: Authorization missing/weak
- SEC-004: Cryptographic weaknesses
- SEC-005: Sensitive data exposure
- SEC-006: Security theater (looks secure, isn't)

**Architecture (ARCH-xxx):**
- ARCH-001: Layer boundary violations
- ARCH-002: Circular dependencies
- ARCH-003: God classes/functions
- ARCH-004: Missing abstraction
- ARCH-005: Inappropriate coupling
- ARCH-006: Scalability concerns

**Accessibility (A11Y-xxx):**
- A11Y-001: Missing ARIA labels
- A11Y-002: Keyboard navigation broken
- A11Y-003: Color contrast insufficient
- A11Y-004: Missing alt text
- A11Y-005: Focus management issues

**Complexity (CMPLX-xxx):**
- CMPLX-001: Cyclomatic complexity >10
- CMPLX-002: Cognitive complexity high
- CMPLX-003: Nesting depth >4
- CMPLX-004: Parameter count >5
- CMPLX-005: File too long (>500 lines)

**Semantic (SEM-xxx):**
- SEM-001: Docstring lies (claims vs implementation)
- SEM-002: Misleading function names
- SEM-003: Misleading variable names
- SEM-004: Comment rot (outdated comments)
- SEM-005: API contract violations

**Performance (PERF-xxx):**
- PERF-001: N+1 query patterns
- PERF-002: Unbounded loops/recursion
- PERF-003: Missing pagination
- PERF-004: Resource leaks (unclosed handles)
- PERF-005: Blocking in async context

**Slop (SLOP-xxx):**
- SLOP-001: Hallucinated imports/APIs
- SLOP-002: Cargo cult patterns
- SLOP-003: Excessive boilerplate
- SLOP-004: AI conversation artifacts
- SLOP-005: Over-engineering for simple tasks

### Phase 3: Expert Routing

For CRITICAL/HIGH findings, optionally invoke expert agents via explicit request:

```markdown
# Security findings → security-expert
"Use the security-expert agent to deep dive on SEC-xxx findings..."

# Architecture findings → architecture-expert
"Use the architecture-expert agent to validate ARCH-xxx concerns..."

# Accessibility findings → ux-expert
"Use the ux-expert agent to audit A11Y-xxx issues..."
```

> **Note:** Custom agents are invoked via explicit request, not Task().
> Task() only supports built-in subagent_types (Explore, Plan, etc.).

Enable with `--expert-routing` flag.

### Phase 4: Report Generation

Output to multiple formats:

| Format | Path | Use |
|--------|------|-----|
| Markdown | `.agents/assessments/{date}-vibe-{target}.md` | Human review |
| JSON | `reports/vibe-report.json` | Tooling/CI |
| JUnit XML | `reports/vibe-junit.xml` | CI integration |
| SARIF | `reports/vibe.sarif` | Security tools |

### Phase 5: Issue Creation

When `--create-issues` is set, create beads issues for findings:

```bash
/vibe recent --create-issues   # Create issues for CRITICAL/HIGH
/vibe recent --create-issues --threshold medium  # Include MEDIUM
```

---

## Severity Matrix

| Level | Definition | Action | Exit Code |
|-------|------------|--------|-----------|
| **CRITICAL** | Security vuln, data loss, broken build | Block merge | 2 |
| **HIGH** | Significant quality/security gap | Fix before merge | 3 |
| **MEDIUM** | Technical debt, minor issues | Follow-up issue | 0 |
| **LOW** | Nitpicks, style preferences | Optional | 0 |

---

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `VIBE_CREATE_ISSUES` | `false` | Auto-create beads issues |
| `VIBE_FAIL_THRESHOLD` | `critical` | Exit non-zero threshold |
| `VIBE_EXPERT_ROUTING` | `false` | Spawn expert agents for deep dive |
| `VIBE_OUTPUT_FORMAT` | `markdown` | Report format (markdown/json/junit/sarif) |
| `VIBE_SKIP_PRESCAN` | `false` | Skip static analysis phase |
| `VIBE_SKIP_SEMANTIC` | `false` | Skip LLM analysis phase |

---

## Examples

```bash
# Standard full validation
/vibe recent

# CI pipeline (fast, fail on critical)
/vibe --fast recent

# Security audit
/vibe --security services/auth/

# Architecture review before major change
/vibe --arch services/ --expert-routing

# Accessibility audit for frontend
/vibe --only a11y frontend/src/components/

# New codebase deep audit
/vibe --deep all --create-issues

# Pre-commit hook style
VIBE_FAIL_THRESHOLD=high /vibe recent

# Specific aspects
/vibe recent --only security,complexity,semantic

# Exclude noisy aspects
/vibe recent --exclude slop
```

---

## Plugin Validation Mode

Validate Claude Code plugins (commands, skills, agents):

```bash
/vibe --plugin ./my-skill
```

**Plugin Checks:**
- Description matches implementation
- Triggers are accurate (skills)
- Arguments are used (commands)
- Tools declared are used
- No painted doors (documented but missing features)

---

## Integration with Workflow

### Pre-commit
```bash
# .git/hooks/pre-commit
/vibe --fast staged || exit 1
```

### CI Pipeline
```yaml
- name: Vibe Validation
  run: |
    VIBE_FAIL_THRESHOLD=high /vibe recent
```

### PR Review
```bash
# Full review with issue creation
/vibe recent --create-issues --expert-routing
```

### Crank Integration
```bash
# In /crank workflow, vibe runs automatically after implementation
/implement <issue> && /vibe recent --create-issues
```

---

## Output Format

```markdown
## Vibe Validation Report

**Target:** services/gateway/
**Date:** 2026-01-18
**Mode:** Full (all aspects)

### Summary
| Aspect | Critical | High | Medium | Low |
|--------|----------|------|--------|-----|
| Security | 1 | 2 | 0 | 0 |
| Quality | 0 | 3 | 5 | 2 |
| Architecture | 0 | 1 | 2 | 0 |
| Complexity | 0 | 2 | 3 | 0 |
| ... | ... | ... | ... | ... |

### Critical Findings (Block Merge)
- **SEC-001** `auth/handler.go:42` - SQL injection in user lookup
  - Fix: Use parameterized query

### High Priority (Fix Before Merge)
- **CMPLX-001** `services/processor.py:89` - CC=23
  - Fix: Extract validation logic to separate function
...

### Beads Issues Created
- sec-0001: Fix SQL injection in auth handler
- cmplx-0002: Refactor processor.py complexity
```

---

## References

- **Pattern Details**: `references/patterns.md`
- **Report Formats**: `references/report-format.md`
- **Prescan Script**: `scripts/prescan.sh`
- **Expert Agents**: `~/.claude/agents/`
