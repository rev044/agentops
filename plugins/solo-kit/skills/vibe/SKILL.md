---
name: vibe
description: >
  Talos-class comprehensive code validation. Use for "validate code",
  "run vibe", "check quality", "security review", "architecture review",
  "accessibility audit", "complexity check", or any validation need.
  One skill to validate them all.
allowed-tools: "Read,Bash,Glob,Grep,Write,TodoWrite,Task"
version: 3.0.0
tier: solo
author: "AI Platform Team"
license: "MIT"
context: inline
context-budget:
  skill-md: 10KB
  references-total: 12KB
  typical-session: 15KB
skills:
  - beads
  - standards
---

# Vibe - Talos Comprehensive Validation

**One skill to validate them all.**

Vibe is a Talos-class validator that combines fast static analysis with deep
semantic verification across all quality dimensions: code quality, security,
architecture, accessibility, complexity, and more.

## Role in the Brownian Ratchet

> **Vibe is THE filter.**

In the Brownian Ratchet pipeline (chaos + filter + ratchet = progress), vibe serves as
the primary validation gate that determines what can proceed:

| Severity | Gate Decision | Ratchet Status |
|----------|---------------|----------------|
| 0 CRITICAL | **PASS** | Can ratchet forward (merge allowed) |
| 1+ CRITICAL | **BLOCK** | Must fix before proceeding |
| HIGH findings | **WARN** | Creates follow-up issues, can proceed |

**Gate Mode** (CI/automation):
```bash
/vibe recent --gate      # Exit non-zero on CRITICAL findings
```

Without the filter, chaos produces garbage. Vibe ensures only valid work ratchets.

## Philosophy

> **Mono over Micro**: Instead of chaining small skills, Vibe provides comprehensive
> validation in one invocation. Trade-off: larger context, but simpler mental model
> and guaranteed coverage.

> **Evidence over Scores**: All claims must be verifiable with specific evidence.
> Use letter grades + findings for quality assessments. Never use numeric scores
> (X/100) for subjective qualities. No "100%" claims without specific context.

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

Vibe validates across 9 aspects. By default, all aspects run.

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
| **Config Drift** | DRIFT-xxx | Config/code mismatch, dead config, multiple sources of truth | architecture-expert |
| **Prose** | PROSE-xxx | Writing quality, voice, Medium optimization (blog/docs) | writing skill |

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

---

## Execution Flow

### Phase 1: Prescan (Static, No LLM)

Fast pattern detection using static analysis tools.

```bash
~/.claude/skills/vibe/scripts/prescan.sh "$TARGET"
```

**Patterns Detected:**

| ID | Pattern | Severity | Language | Tool |
|----|---------|----------|----------|------|
| P1 | Phantom modifications | CRITICAL | All | git diff |
| P2 | Hardcoded secrets | CRITICAL | All | gitleaks/grep |
| P3 | SQL injection patterns | CRITICAL | All | regex |
| P4 | TODO/FIXME/commented code | HIGH | All | grep |
| P5 | Cyclomatic complexity >15 | HIGH | All | radon/gocyclo |
| P6 | Functions >50 lines | HIGH | All | wc/ast |
| P7 | Bare except/empty catch | HIGH | Python/Bash | ast/shellcheck |
| P8 | Unused imports/functions | MEDIUM | All | ast |
| P9 | Docstring mismatches | MEDIUM | Python | ast |
| P10 | Missing error handling | MEDIUM | All | ast |
| **P11** | **Shellcheck violations** | **HIGH** | **Shell** | **shellcheck** |
| **P12** | **Missing cluster connectivity gate** | **HIGH** | **Shell** | **grep** |
| **P13** | **Undocumented error ignores** | **HIGH** | **Go** | **grep/sed** |
| **P14** | **Error wrapping with %v** | **MEDIUM** | **Go** | **grep** |
| **P15** | **golangci-lint violations** | **HIGH** | **Go** | **golangci-lint** |
| **P16** | **Catch-all pattern evaluated first** | **HIGH** | **All** | **ast/grep** |
| **P17** | **String comparison on version strings** | **HIGH** | **Python** | **ast** |
| **P18** | **Unused CLI flag/variable** | **MEDIUM** | **All** | **grep** |
| **P19** | **YAML field never read by code** | **HIGH** | **YAML+Code** | **cross-ref** |

**P11 Details:** Runs shellcheck on shell scripts, flags common issues (SC2086, SC2164, etc.)
**P12 Details:** Detects shell scripts using `oc`/`kubectl` without connectivity check (`oc whoami` or equivalent)
**P13 Details:** Detects `_ =` without `nolint:errcheck` comment (intentional ignores must be documented)
**P14 Details:** Detects `fmt.Errorf.*%v` when wrapping errors (should use `%w` to preserve error chains)
**P15 Details:** Runs golangci-lint if available, parses JSON output, reports up to 10 most severe issues
**P16 Details:** Detects catch-all patterns (`.*`, `*`, `default`) evaluated before specific patterns in switch/match logic
**P17 Details:** Detects string comparison operators (`>=`, `<=`, `>`, `<`) on version-like strings (e.g., `4.18`, `v2.0.1`)
**P18 Details:** Detects CLI flags/variables that are set but never referenced in output or logic
**P19 Details:** Cross-references YAML config fields against code that reads them - flags fields never accessed

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

**Config Drift (DRIFT-xxx):**
- DRIFT-001: Comment/doc claims don't match code behavior
- DRIFT-002: YAML/config fields not read by code (dead config)
- DRIFT-003: Multiple sources of truth for same concept (wave defs, etc.)
- DRIFT-004: Hardcoded values that should come from config
- DRIFT-005: Config schema evolution without code update

**Prose (PROSE-xxx):** *(for blog posts, docs, Medium articles)*
- PROSE-001: AI slop (uniform sentence length, soulless tone)
- PROSE-002: LinkedIn energy (buzzwords, self-congratulatory)
- PROSE-003: Vague claims without evidence ("significant improvement")
- PROSE-004: Wall of text (paragraphs >3 sentences)
- PROSE-005: Missing hook (doesn't open with insight)
- PROSE-006: Academic tone ("Furthermore", "It should be noted")
- PROSE-007: No actionable ending (no "Try It" section)
- PROSE-008: Length mismatch (not in 5-7 min range for Medium)

*Reference: `/writing` skill for full voice patterns and Medium optimization*

### Phase 3: Expert Routing

For CRITICAL/HIGH findings, optionally spawn expert agents:

```markdown
# Security findings → security-expert
Task(subagent_type="security-expert", prompt="Deep dive on SEC-xxx findings...")

# Architecture findings → architecture-expert
Task(subagent_type="architecture-expert", prompt="Validate ARCH-xxx concerns...")

# Accessibility findings → ux-expert
Task(subagent_type="ux-expert", prompt="Audit A11Y-xxx issues...")
```

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
/vibe --plugin ~/.claude/skills/my-skill
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

### Claim Validation Requirements

**CRITICAL: All claims in vibe reports MUST be verifiable with evidence.**

#### Score Inflation Anti-Patterns (FORBIDDEN)

❌ **DON'T use numeric scores for subjective assessments:**
- `Interface Design | 100/100` - No scale definition, unfalsifiable
- `Security | 95/100` - Cannot be objectively measured
- `Overall Score: 91.8/100` - False precision

✅ **DO use letter grades + evidence:**
- `Interface Design | A+ | 9 small focused interfaces, proper composition`
- `Security | A | Timing-safe auth, HMAC validation, 0 hardcoded secrets`
- `Overall Grade: A- (Excellent) | 12 HIGH, 34 MEDIUM findings`

❌ **DON'T make percentage claims without context:**
- `100% context propagation` - Misleading (not all functions need context)
- `100% test coverage` - What does 100% mean? Line? Branch? Path?

✅ **DO provide specific metrics:**
- `Context propagation: 24/24 activities use context.Context (100% compliant)`
- `Test coverage: 87% line, 71% branch (go test -cover)`

#### Evidence Requirements

Every claim needs:
1. **Measurement method** - How was it determined?
2. **Specific numbers** - Actual counts, not percentages
3. **Context** - What was measured?
4. **Location** - Where to verify?

**Examples:**

| ❌ Bad Claim | ✅ Good Claim |
|-------------|--------------|
| "100% naming conventions" | "92/92 Go files use lowercase (find . -name '*.go')" |
| "Excellent error handling" | "131 proper %w wrappings, 5 undocumented ignores (grep)" |
| "Perfect security" | "0 CRITICAL, 2 HIGH findings (gitleaks + prescan P2)" |

#### Grading Scale

Use letter grades for subjective quality assessments:

| Grade | Numeric | Criteria |
|-------|---------|----------|
| A+ | 95-100 | Exemplary - industry best practices, 0-2 minor findings |
| A | 90-94 | Excellent - strong practices, <5 HIGH findings |
| A- | 85-89 | Very Good - solid practices, 5-15 HIGH findings |
| B+ | 80-84 | Good - acceptable practices, 15-25 HIGH findings |
| B | 75-79 | Satisfactory - needs improvement, 25-40 HIGH findings |
| C+ | 70-74 | Needs Improvement - multiple issues, 40-60 HIGH findings |
| C | 65-69 | Significant Issues - major refactoring needed |
| D | 60-64 | Major Problems - not production-ready |
| F | <60 | Critical Issues - complete rewrite recommended |

**Note:** Numeric ranges are guidelines only. Always include finding counts as primary evidence.

### Report Self-Validation

**Before outputting any vibe report, validate it against these anti-patterns:**

#### Automated Checks

```bash
# Check for numeric score anti-patterns
grep -E '\d+/100|[0-9]+\.[0-9]+/100' report.md
# Should return: nothing

# Check for unqualified percentage claims
grep -E '100%|[0-9]+%' report.md | grep -v 'test coverage|line coverage|branch'
# Review each match for context

# Check for subjective perfection claims
grep -iE 'perfect|flawless|zero issues|no problems' report.md
# Should return: nothing (except in negative context like "not perfect")
```

#### Manual Review Checklist

Before finalizing report:
- [ ] All claims have measurement method specified
- [ ] Percentage claims include both numerator and denominator
- [ ] Subjective assessments use letter grades, not numbers
- [ ] Every grade includes supporting evidence
- [ ] No "100%" claims without specific context
- [ ] No "X/100" scores for non-measurable qualities
- [ ] Finding counts included in summary

**If any check fails, revise the report before output.**

---

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

### Go-Specific Validation Example

```markdown
## Vibe Validation Report - Go

**Target:** internal/agents/
**Date:** 2026-01-20
**Mode:** Deep (with go-standards.md)

### Prescan Results (P13-P15)
| Pattern | Findings | Severity |
|---------|----------|----------|
| P13: Undocumented Error Ignores | 5 | HIGH |
| P14: Error Wrapping with %v | 5 | MEDIUM |
| P15: golangci-lint Violations | 87 | HIGH |

### Go Standards Compliance
| Category | Grade | Evidence |
|----------|-------|----------|
| Error Handling | A- | 131 proper wrappings, 5 undocumented ignores, 5 %v instances |
| Interface Design | A+ | 9 small focused interfaces, accept interfaces/return structs |
| Concurrency | A | 24/24 activities use context, proper mutex usage, 0 race conditions |
| Security | A | Timing-safe comparisons, HMAC validation, 0 hardcoded secrets |
| **OVERALL** | **A- (Excellent)** | **12 HIGH, 34 MEDIUM findings** |

### High Priority (Fix Before Merge)
- **P13-001** `registry.go:114` - Error ignored without documentation
  - Fix: `_ = agent.Shutdown(ctx) // nolint:errcheck - best effort cleanup`

- **P14-001** `workflow.go:596` - Error wrapping with %v breaks chain
  - Fix: Change `fmt.Errorf("failed: %v", err)` to `fmt.Errorf("failed: %w", err)`

- **CMPLX-001** `handler.go:239` - CC=14 (handleGitHub)
  - Fix: Extract validation helpers, use strategy map for events

### Recommendations
- Refactor 10 functions with CC > 10
- Add nolint:errcheck comments to 5 error ignores
- Fix 5 error wrapping instances
- Target: Grade A+ (100/100)
```

---

## References

- **Pattern Details**: `references/patterns.md`
- **Report Formats**: `references/report-format.md`
- **Prescan Script**: `scripts/prescan.sh`
- **Expert Agents**: `~/.claude/agents/`

---

## Standards Loading

Load language-specific standards during semantic analysis (Phase 2) based on file patterns detected:

| File Pattern | Tier 1 (Generic) | Tier 2 (Deep) | Override |
|--------------|------------------|---------------|----------|
| `*.py` | `standards/references/python.md` | - | `.agents/validation/PYTHON_*.md` |
| `*.go` | `standards/references/go.md` (5KB) | `vibe/references/go-standards.md` (16KB) | `.agents/validation/GO_*.md` |
| `*.ts`, `*.tsx` | `standards/references/typescript.md` | - | `.agents/validation/TS_*.md` |
| `*.sh` | `standards/references/shell.md` | - | `.agents/validation/SHELL_*.md` |
| `*.yaml`, `*.yml` | `standards/references/yaml.md` | - | - |
| `*.md` | `standards/references/markdown.md` | `writing/SKILL.md` (blog/Medium) | `.agents/validation/PROSE_*.md` |
| `*.json`, `*.jsonl` | `standards/references/json.md` | - | - |

### Platform Standards (Auto-Detected)

| Detection Pattern | Standard | Checks |
|-------------------|----------|--------|
| `**/api/**/groupversion_info.go` | `standards/references/olympus.md` | API group domain, version |
| `**/config/crd/**/*.yaml` | `standards/references/olympus.md` | CRD naming, labels |
| Content contains `olympus.io` | `standards/references/olympus.md` | Domain migration |
| Content contains `mt-olympus.io` | `standards/references/olympus.md` | Domain compliance |

**Olympus Prescan Checks (OL-xxx):**
- OL-001: `olympus.io` without `mt-` prefix (HIGH)
- OL-002: `{god}.olympus` without TLD (HIGH)
- OL-003: Legacy `kagent.dev` group (HIGH)
- OL-004: `fractal.io` references (CRITICAL)
- OL-005: Mixed domain patterns in same file (MEDIUM)

### Two-Tier Loading (Go Example)

**Tier 1 (Fast, Generic):**
- Load `standards/references/go.md` for ALL Go validation
- 5KB generic standards (error handling, concurrency basics)
- Used for normal vibe runs

**Tier 2 (Deep, Comprehensive):**
- Load `vibe/references/go-standards.md` when `--deep` flag used
- 16KB comprehensive catalog (security, testing, anti-patterns)
- Full compliance scoring
- Used for audits and new codebases

**Project Override:**
- If `.agents/validation/GO_*.md` exists, use instead of generic
- Allows project-specific standards
- Example: Apollo has `GO_STANDARDS_CATALOG.md` summary

**Quick Reference:**
- `vibe/references/go-patterns.md` provides copy-paste examples
- Loaded JIT when fixing issues

**Usage**: When validating files, load Tier 1 for normal runs, Tier 2 for `--deep`, Override if present. This keeps context under 40% while enabling comprehensive validation on demand.

---

## Cross-Reference Validation (Config Drift)

The Config Drift aspect detects mismatches between documentation/config and actual code behavior.

### DRIFT-001: Comment Claims vs Code

**Detection:**
```bash
# Find structured comments claiming behavior
grep -n "Wave [0-9]:" scripts/specialist.sh
# Output: "Wave 9: validation"

# Cross-reference with actual config
grep -A5 "number: 9" waves/v1.16.yaml
# Output: "name: Remaining Apps" -- MISMATCH!
```

**Common Patterns:**
- Architecture comments that don't match implementation
- Wave/phase definitions in multiple places
- API documentation out of sync with handlers

### DRIFT-002: Dead Config Fields

**Detection:**
```python
# Extract all YAML keys defined
yaml_fields = extract_keys("config.yaml")
# ['name', 'phase', 'apps', 'validation_script']

# Find what code actually reads
code_reads = grep_patterns("*.py", r'\.get\(["\'](\w+)')
# ['name', 'apps']

# Dead fields = defined but never read
dead_fields = yaml_fields - code_reads
# {'phase', 'validation_script'}  # BUG!
```

### DRIFT-003: Multiple Sources of Truth

**Detection:**
```bash
# Find all wave definitions
grep -rn "wave.*[0-9]" --include="*.sh" --include="*.yaml" --include="*.py"

# If same concept (e.g., "wave 9") defined differently in multiple files:
# - upgrade-specialist.sh says "validation"
# - waves.yaml says "auto-discover"
# - upgrade.py hardcodes different mapping
# = DRIFT-003 violation
```

### Integration with Prescan

For P19 (YAML field never read), the prescan does:

```bash
# 1. Extract YAML structure
yq eval 'keys' config.yaml > /tmp/yaml_keys.txt

# 2. Find code access patterns
grep -r "\.get\|getattr\|\[" *.py | grep -oE '["\'][a-z_]+["\']' > /tmp/code_reads.txt

# 3. Diff
comm -23 <(sort /tmp/yaml_keys.txt) <(sort /tmp/code_reads.txt)
# Outputs fields defined in YAML but never accessed in code
```

### Shell Script Safety Checks (P11, P12)

**P11 - Shellcheck Integration:**
```bash
# Run shellcheck on all shell scripts
find . -name "*.sh" -exec shellcheck -f json {} \; 2>/dev/null | jq -s 'flatten'

# Filter to HIGH severity
jq '[.[] | select(.level == "error" or .level == "warning")]'
```

**P12 - Cluster Connectivity Gate:**
```bash
# Detect scripts using oc/kubectl without connectivity check
for script in $(find . -name "*.sh"); do
  if grep -q "oc get\|oc patch\|kubectl" "$script"; then
    if ! grep -q "oc whoami\|kubectl cluster-info\|oc cluster-info" "$script"; then
      echo "P12: $script uses cluster commands without connectivity check"
    fi
  fi
done
```

### Prose Validation (PROSE-xxx)

For blog posts, Medium articles, and documentation (`.md` files in `.agents/drafts/`, `content/`, `public/`):

**Trigger Detection:**
```bash
# Detect blog/Medium content
if [[ "$FILE" =~ \.agents/drafts/.*\.md$ ]] || \
   [[ "$FILE" =~ content/.*\.md$ ]] || \
   [[ "$FILE" =~ devlog.*\.md$ ]] || \
   [[ "$FILE" =~ public/medium/.*\.md$ ]]; then
    ENABLE_PROSE_VALIDATION=true
fi
```

**Checks Applied:**
```bash
# PROSE-001: AI slop detection (uniform sentence length)
awk '{print length}' "$FILE" | sort -u | wc -l  # Should have variance

# PROSE-002: LinkedIn energy (buzzword detection)
grep -ciE 'leverage|synergy|excited to announce|thrilled|game-changing' "$FILE"

# PROSE-003: Vague claims
grep -ciE 'significant|much (faster|better)|improved dramatically' "$FILE"

# PROSE-004: Wall of text (paragraph length)
# Check for paragraphs with >3 sentences

# PROSE-008: Length check for Medium
wc -w "$FILE"  # Target: 1,400-1,960 words for 5-7 min read
```

**Integration with /writing skill:**

When prose validation finds issues, reference the writing skill:
```markdown
### PROSE Findings

- **PROSE-002** `devlog-4.md:42` - LinkedIn energy detected: "excited to announce"
  - Fix: State facts, let numbers speak. See `/writing` skill → Anti-Patterns

- **PROSE-004** `devlog-4.md:78` - Paragraph has 6 sentences
  - Fix: Max 3 sentences per paragraph. See `/writing` skill → Medium Optimization
```

---

### Version Comparison Check (P17)

**Detection:**
```python
# AST-based detection for Python
import ast

class VersionCompareVisitor(ast.NodeVisitor):
    def visit_Compare(self, node):
        # Check if comparing version-like strings
        for comparator in node.comparators:
            if isinstance(comparator, ast.Constant):
                if re.match(r'^\d+\.\d+', str(comparator.value)):
                    # Found version string in comparison
                    emit_warning(f"P17: String comparison on version at line {node.lineno}")
```

### Pattern Precedence Check (P16)

**Detection:**
```bash
# Find catch-all patterns in switch/case/match
grep -n '".*".*:' file.sh | head -1  # If catch-all is first = P16

# For Python dicts with regex patterns
grep -A20 "PATTERNS = {" file.py | grep -n '".\*"'
# If ".*" appears before more specific patterns = P16
```
