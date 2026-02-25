---
name: vibe
description: 'Comprehensive code validation. Runs complexity analysis then multi-model council. Answer: Is this code ready to ship? Triggers: "vibe", "validate code", "check code", "review code", "code quality", "is this ready".'
---


# Vibe Skill

> **Purpose:** Is this code ready to ship?

Three steps:
1. **Complexity analysis** — Find hotspots (radon, gocyclo)
2. **Bug hunt audit** — Systematic sweep for concrete bugs
3. **Council validation** — Multi-model judgment

---

## Quick Start

```bash
$vibe                                    # validates recent changes
$vibe recent                             # same as above
$vibe src/auth/                          # validates specific path
$vibe --quick recent                     # fast inline check, no agent spawning
$vibe --deep recent                      # 3 judges instead of 2
$vibe --mixed recent                     # cross-vendor (Claude + Codex)
$vibe --preset=security-audit src/auth/  # security-focused review
$vibe --explorers=2 recent               # judges with explorer sub-agents
$vibe --debate recent                    # two-round adversarial review
```

---

## Execution Steps

### Step 1: Determine Target

**If target provided:** Use it directly.

**If no target or "recent":** Auto-detect from git:
```bash
# Check recent commits
git diff --name-only HEAD~3 2>/dev/null | head -20
```

If nothing found, ask user.

**Pre-flight: If no files found:**
Return immediately with: "PASS (no changes to review) — no modified files detected."
Do NOT spawn agents for empty file lists.

### Step 1.5: Fast Path (--quick mode)

**If `--quick` flag is set**, skip Steps 2a–2f (constraint tests, metadata checks, OL validation, codex review, knowledge search, bug hunt, product context) and jump directly to Step 4 with inline council. Complexity analysis (Step 2) still runs — it's cheap and informative.

**Why:** Steps 2a–2f add 30–90 seconds of pre-processing that feed multi-judge council packets. In --quick mode (single inline agent), these inputs aren't worth the cost — the inline reviewer reads files directly.

### Step 2: Run Complexity Analysis

**Detect language and run appropriate tool:**

**For Python:**
```bash
# Check if radon is available
mkdir -p .agents/council
echo "$(date -Iseconds) preflight: checking radon" >> .agents/council/preflight.log
if ! which radon >> .agents/council/preflight.log 2>&1; then
  echo "⚠️ COMPLEXITY SKIPPED: radon not installed (pip install radon)"
  # Record in report that complexity was skipped
else
  # Run cyclomatic complexity
  radon cc <path> -a -s 2>/dev/null | head -30
  # Run maintainability index
  radon mi <path> -s 2>/dev/null | head -30
fi
```

**For Go:**
```bash
# Check if gocyclo is available
echo "$(date -Iseconds) preflight: checking gocyclo" >> .agents/council/preflight.log
if ! which gocyclo >> .agents/council/preflight.log 2>&1; then
  echo "⚠️ COMPLEXITY SKIPPED: gocyclo not installed (go install github.com/fzipp/gocyclo/cmd/gocyclo@latest)"
  # Record in report that complexity was skipped
else
  # Run complexity analysis
  gocyclo -over 10 <path> 2>/dev/null | head -30
fi
```

**For other languages:** Skip complexity with explicit note: "⚠️ COMPLEXITY SKIPPED: No analyzer for <language>"

**Interpret results:**

| Score | Rating | Action |
|-------|--------|--------|
| A (1-5) | Simple | Good |
| B (6-10) | Moderate | OK |
| C (11-20) | Complex | Flag for council |
| D (21-30) | Very complex | Recommend refactor |
| F (31+) | Untestable | Must refactor |

**Include complexity findings in council context.**

### Step 2a: Run Constraint Tests

**Skip if `--quick` (see Step 1.5).**

**If the project has constraint tests, run them before council:**

```bash
# Check if constraint tests exist (Olympus pattern)
if [ -d "internal/constraints" ] && ls internal/constraints/*_test.go &>/dev/null; then
  echo "Running constraint tests..."
  go test ./internal/constraints/ -run TestConstraint -v 2>&1
  # If FAIL → include failures in council context as CRITICAL findings
  # If PASS → note "N constraint tests passed" in report
fi
```

**Why:** Constraint tests catch mechanical violations (ghost references, TOCTOU races, dead code at entry points) that council judges miss. Proven by Argus ghost ref in ol-571 — council gave PASS while constraint test caught it.

Include constraint test results in the council packet context. Failed constraint tests are CRITICAL findings that override council PASS verdict.

### Step 2b: Metadata Verification Checklist (MANDATORY)

**Skip if `--quick` (see Step 1.5).**

Run mechanical checks BEFORE council — catches errors LLMs estimate instead of measure:
1. **File existence** — every path in `git diff --name-only HEAD~3` must exist on disk
2. **Line counts** — if a file claims "N lines", verify with `wc -l`
3. **Cross-references** — internal markdown links resolve to existing files
4. **Diagram sanity** — files with >3 ASCII boxes should have matching labels

Include failures in council packet as `context.metadata_failures` (MECHANICAL findings). If all pass, note in report.

### Step 2c: Deterministic Validation (Olympus)

**Skip if `--quick` (see Step 1.5).**

**Guard:** Only run when `.ol/config.yaml` exists AND `which ol` succeeds. Skip silently otherwise.

**Implementation:**

```bash
# Run ol-validate.sh
skills/vibe/scripts/ol-validate.sh
ol_exit_code=$?

case $ol_exit_code in
  0)
    # Passed: include the validation report in vibe output
    echo "✅ Deterministic validation passed"
    # Append the report section to council context and vibe report
    ;;
  1)
    # Failed: abort vibe with FAIL verdict
    echo "❌ Deterministic validation FAILED"
    echo "VIBE FAILED — Olympus Stage1 validation did not pass"
    exit 1
    ;;
  2)
    # Skipped: note and continue
    echo "⚠️ OL validation skipped"
    # Continue to council
    ;;
esac
```

**Behavior:**
- **Exit 0 (passed):** Include the validation report section in vibe output and council context. Proceed normally.
- **Exit 1 (failed):** Auto-FAIL the vibe. Do NOT proceed to council.
- **Exit 2 (skipped):** Note "OL validation skipped" in report. Proceed to council.

### Step 2.5: Codex Review (if available)

**Skip if `--quick` (see Step 1.5).**

Run a fast, diff-focused code review via Codex CLI before council:

```bash
echo "$(date -Iseconds) preflight: checking codex" >> .agents/council/preflight.log
if which codex >> .agents/council/preflight.log 2>&1; then
  codex review --uncommitted > .agents/council/codex-review-pre.md 2>&1 && \
    echo "Codex review complete — output at .agents/council/codex-review-pre.md" || \
    echo "Codex review skipped (failed)"
else
  echo "Codex review skipped (CLI not found)"
fi
```

**If output exists**, summarize and include in council packet (cap at 2000 chars to prevent context bloat):
```json
"codex_review": {
  "source": "codex review --uncommitted",
  "content": "<first 2000 chars of .agents/council/codex-review-pre.md>"
}
```

**IMPORTANT:** The raw codex review can be 50k+ chars. Including the full text in every judge's packet multiplies token cost by N judges. Truncate to the first 2000 chars (covers the summary and top findings). Judges can read the full file from disk if they need more detail.

This gives council judges a Codex-generated review as pre-existing context — cheap, fast, diff-focused. It does NOT replace council judgment; it augments it.

**Skip conditions:**
- Codex CLI not on PATH → skip silently
- `codex review` fails → skip silently, proceed with council only
- No uncommitted changes → skip (nothing to review)

### Step 2d: Search Knowledge Flywheel

**Skip if `--quick` (see Step 1.5).**

```bash
if command -v ao &>/dev/null; then
    ao search "code review findings <target>" 2>/dev/null | head -10
fi
```
If ao returns prior code review patterns for this area, include them in the council packet context. Skip silently if ao is unavailable or returns no results.

### Step 2e: Bug Hunt Audit

**Skip if `--quick` (see Step 1.5).**

Run a proactive bug hunt on the target files before council review:

```
$bug-hunt --audit <target>
```

If bug-hunt produces findings, include them in the council packet as `context.bug_hunt`:
```json
"bug_hunt": {
  "source": "$bug-hunt --audit",
  "findings_count": 3,
  "high": 1,
  "medium": 1,
  "low": 1,
  "summary": "<first 2000 chars of bug hunt report>"
}
```

**Why:** Bug hunt catches concrete line-level bugs (resource leaks, truncation errors, dead code) that council judges — reviewing holistically — often miss. Proven: goals code audit found 5 real bugs with 0 hypothesis failures by systematic reading.

**Skip conditions:**
- `--quick` mode → skip (fast path)
- No source files in target → skip (nothing to audit)
- Target is non-code (pure docs/config) → skip

### Step 2f: Check for Product Context

**Skip if `--quick` (see Step 1.5).**

```bash
if [ -f PRODUCT.md ]; then
  # PRODUCT.md exists — include developer-experience perspectives
fi
```

When `PRODUCT.md` exists in the project root AND the user did NOT pass an explicit `--preset` override:
1. Read `PRODUCT.md` content and include in the council packet via `context.files`
2. Add a single consolidated `developer-experience` perspective to the council invocation:
   - **With spec:** `$council --preset=code-review --perspectives="developer-experience" validate <target>` (3 judges: 2 code-review + 1 DX)
   - **Without spec:** `$council --perspectives="developer-experience" validate <target>` (3 judges: 2 independent + 1 DX)
   The DX judge covers api-clarity, error-experience, and discoverability in a single review.
3. With `--deep`: adds 1 more judge per mode (4 judges total).

When `PRODUCT.md` exists BUT the user passed an explicit `--preset`: skip DX auto-include (user's explicit preset takes precedence).

When `PRODUCT.md` does not exist: proceed to Step 3 unchanged.

> **Tip:** Create `PRODUCT.md` from `docs/PRODUCT-TEMPLATE.md` to enable developer-experience-aware code review.

### Step 3: Load the Spec (New)

**Skip if `--quick` (see Step 1.5).**

Before invoking council, try to find the relevant spec/bead:

1. **If target looks like a bead ID** (e.g., `na-0042`): `bd show <id>` to get the spec
2. **Search for plan doc:** `ls .agents/plans/ | grep <target-keyword>`
3. **Check git log:** `git log --oneline | head -10` to find the relevant bead reference

If a spec is found, include it in the council packet's `context.spec` field:
```json
{
  "spec": {
    "source": "bead na-0042",
    "content": "<the spec/bead description text>"
  }
}
```

### Step 4: Run Council Validation

**With spec found — use code-review preset:**
```
$council --preset=code-review validate <target>
```
- `error-paths`: Trace every error handling path. What's uncaught? What fails silently?
- `api-surface`: Review every public interface. Is the contract clear? Breaking changes?
- `spec-compliance`: Compare implementation against the spec. What's missing? What diverges?

The spec content is injected into the council packet context so the `spec-compliance` judge can compare implementation against it.

**Without spec — 2 independent judges (no perspectives):**
```
$council validate <target>
```
2 independent judges (no perspective labels). Use `--deep` for 3 judges on high-stakes reviews. Override with `--quick` (inline single-agent check) or `--mixed` (cross-vendor with Codex).

**Council receives:**
- Files to review
- Complexity hotspots (from Step 2)
- Git diff context
- Spec content (when found, in `context.spec`)

All council flags pass through: `--quick` (inline), `--mixed` (cross-vendor), `--preset=<name>` (override perspectives), `--explorers=N`, `--debate` (adversarial 2-round). See Quick Start examples and `$council` docs.

### Step 5: Council Checks

Each judge reviews for:

| Aspect | What to Look For |
|--------|------------------|
| **Correctness** | Does code do what it claims? |
| **Security** | Injection, auth issues, secrets |
| **Edge Cases** | Null handling, boundaries, errors |
| **Quality** | Dead code, duplication, clarity |
| **Complexity** | High cyclomatic scores, deep nesting |
| **Architecture** | Coupling, abstractions, patterns |

### Step 6: Interpret Verdict

| Council Verdict | Vibe Result | Action |
|-----------------|-------------|--------|
| PASS | Ready to ship | Merge/deploy |
| WARN | Review concerns | Address or accept risk |
| FAIL | Not ready | Fix issues |

### Step 7: Write Vibe Report

**Write to:** `.agents/council/YYYYMMDDTHHMMSSZ-vibe-<target>.md` (use `date -u +%Y%m%dT%H%M%SZ`)

```markdown
# Vibe Report: <Target>

**Date:** YYYY-MM-DD
**Files Reviewed:** <count>

## Complexity Analysis

**Status:** ✅ Completed | ⚠️ Skipped (<reason>)

| File | Score | Rating | Notes |
|------|-------|--------|-------|
| src/auth.py | 15 | C | Consider breaking up |
| src/utils.py | 4 | A | Good |

**Hotspots:** <list files with C or worse>
**Skipped reason:** <if skipped, explain why - e.g., "radon not installed">

## Council Verdict: PASS / WARN / FAIL

| Judge | Verdict | Key Finding |
|-------|---------|-------------|
| Error-Paths | ... | ... (with spec — code-review preset) |
| API-Surface | ... | ... (with spec — code-review preset) |
| Spec-Compliance | ... | ... (with spec — code-review preset) |
| Judge 1 | ... | ... (no spec — 2 independent judges) |
| Judge 2 | ... | ... (no spec — 2 independent judges) |
| Judge 3 | ... | ... (no spec — 2 independent judges) |

## Shared Findings
- ...

## Concerns Raised
- ...

## Recommendation
<council recommendation>

## Decision

[ ] SHIP - Complexity acceptable, council passed
[ ] FIX - Address concerns before shipping
[ ] REFACTOR - High complexity, needs rework
```

### Step 8: Report to User

Tell the user:
1. Complexity hotspots (if any)
2. Council verdict (PASS/WARN/FAIL)
3. Key concerns
4. Location of vibe report

### Step 9: Record Ratchet Progress

After council verdict:
1. If verdict is PASS or WARN:
   - Run: `ao ratchet record vibe --output "<report-path>" 2>/dev/null || true`
   - Suggest: "Run $post-mortem to capture learnings and complete the cycle."
2. If verdict is FAIL:
   - Do NOT record ratchet progress.
   - Extract top 5 findings from the council report for structured retry context:
     ```
     Read the council report. For each finding (max 5), format as:
     FINDING: <description> | FIX: <fix or recommendation> | REF: <ref or location>

     Fallback for v1 findings (no fix/why/ref fields):
       fix = finding.fix || finding.recommendation || "No fix specified"
       ref = finding.ref || finding.location || "No reference"
     ```
   - Tell user to fix issues and re-run $vibe, including the formatted findings as actionable guidance.

### Step 9.5: Feed Findings to Flywheel

**If verdict is WARN or FAIL**, write top findings as a lightweight learning for future sessions:

```bash
if [[ "$VERDICT" == "WARN" || "$VERDICT" == "FAIL" ]]; then
  mkdir -p .agents/learnings
  LEARNING_FILE=".agents/learnings/$(date -u +%Y-%m-%d)-vibe-$(echo "$TARGET" | tr '/' '-' | head -c 40).md"
  cat > "$LEARNING_FILE" <<EOF
---
type: anti-pattern
source: vibe
date: $(date -Iseconds)
confidence: high
---

# Vibe findings: $TARGET

$(for finding in "${TOP_FINDINGS[@]:0:3}"; do
  echo "- **${finding.severity}:** ${finding.description} (${finding.location})"
done)

**Recommendation:** ${COUNCIL_RECOMMENDATION}
EOF

  # Index for flywheel if ao available
  if command -v ao &>/dev/null; then
    ao forge markdown "$LEARNING_FILE" 2>/dev/null || true
  fi
fi
```

**Why:** Vibe catches anti-patterns repeatedly across epics but they evaporate unless `$post-mortem` runs. This captures findings at the point of discovery — lightweight (one file write, no `$retro` invocation) and immediately available to future sessions via inject.

**Skip if:** PASS verdict (nothing to learn from clean code).

### Step 10: Test Bead Cleanup

After validation completes (regardless of verdict), clean up any stale test beads to prevent bead pollution:

```bash
# Test bead hygiene: close any beads created by test/validation runs
if command -v bd &>/dev/null; then
  test_beads=$(bd list --status=open 2>/dev/null | grep -iE "test bead|test quest|smoke test" | awk '{print $1}')
  if [ -n "$test_beads" ]; then
    echo "$test_beads" | xargs bd close 2>/dev/null || true
    log "Cleaned up $(echo "$test_beads" | wc -l | tr -d ' ') test beads"
  fi
fi
```

---

## Integration with Workflow

```
$implement issue-123
    │
    ▼
(coding, quick lint/test as you go)
    │
    ▼
$vibe                      ← You are here
    │
    ├── Complexity analysis (find hotspots)
    ├── Bug hunt audit (find concrete bugs)
    └── Council validation (multi-model judgment)
    │
    ├── PASS → ship it
    ├── WARN → review, then ship or fix
    └── FAIL → fix, re-run $vibe
```

---

## Examples

**User says:** "Run a quick validation on the latest changes."

**Do:**
```bash
$vibe recent
```

### Validate Recent Changes

```bash
$vibe recent
```

Runs complexity on recent changes, then council reviews.

### Validate Specific Directory

```bash
$vibe src/auth/
```

Complexity + council on auth directory.

### Deep Review

```bash
$vibe --deep recent
```

Complexity + 3 judges for thorough review.

### Cross-Vendor Consensus

```bash
$vibe --mixed recent
```

Complexity + Claude + Codex judges.

See `references/examples.md` for additional examples: security audit with spec compliance, developer-experience code review with PRODUCT.md, and fast inline checks.

---

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| "COMPLEXITY SKIPPED: radon not installed" | Python complexity analyzer missing | Install with `pip install radon` or skip complexity (council still runs). |
| "COMPLEXITY SKIPPED: gocyclo not installed" | Go complexity analyzer missing | Install with `go install github.com/fzipp/gocyclo/cmd/gocyclo@latest` or skip. |
| Vibe returns PASS but constraint tests fail | Council LLMs miss mechanical violations | Check `.agents/council/<timestamp>-vibe-*.md` for constraint test results. Failed constraints override council PASS. Fix violations and re-run. |
| Codex review skipped | Codex CLI not on PATH or no uncommitted changes | Install Codex CLI (`brew install codex`) or commit changes first. Vibe proceeds without codex review. |
| "No modified files detected" | Clean working tree, no recent commits | Make changes or specify target path explicitly: `$vibe src/auth/`. |
| Spec-compliance judge not spawned | No spec found in beads/plans | Reference bead ID in commit message or create plan doc in `.agents/plans/`. Without spec, vibe uses 2 independent judges (3 with `--deep`). |

---

## See Also

- `skills/council/SKILL.md` — Multi-model validation council
- `skills/complexity/SKILL.md` — Standalone complexity analysis
- `skills/bug-hunt/SKILL.md` — Proactive code audit and bug investigation
- `.agents/specs/conflict-resolution-algorithm.md` — Conflict resolution between agent findings

## Reference Documents

- [references/examples.md](references/examples.md)
- [references/go-patterns.md](references/go-patterns.md)
- [references/go-standards.md](references/go-standards.md)
- [references/json-standards.md](references/json-standards.md)
- [references/markdown-standards.md](references/markdown-standards.md)
- [references/patterns.md](references/patterns.md)
- [references/python-standards.md](references/python-standards.md)
- [references/report-format.md](references/report-format.md)
- [references/rust-standards.md](references/rust-standards.md)
- [references/shell-standards.md](references/shell-standards.md)
- [references/typescript-standards.md](references/typescript-standards.md)
- [references/vibe-coding.md](references/vibe-coding.md)
- [references/yaml-standards.md](references/yaml-standards.md)

---

## References

### examples.md

# $vibe Examples

## Security Audit with Spec Compliance

**User says:** `$vibe --preset=security-audit src/auth/`

**What happens:**
1. Agent searches for spec (checks `bd show`, `.agents/plans/`, git log)
2. Agent runs complexity analysis (radon/gocyclo) on `src/auth/`
3. Agent runs constraint tests (`internal/constraints/*_test.go`) if present
4. Agent runs `codex review --uncommitted` for diff-focused review
5. Agent invokes `$council --deep --preset=security-audit validate src/auth/` with spec in packet
6. Spec found: 3 judges use security-audit personas + spec-compliance judge added (4 total)
7. Report written to `.agents/council/<timestamp>-vibe-src-auth.md`

**Result:** Security-focused review with attacker/defender/compliance perspectives and spec validation.

## Developer-Experience Code Review (PRODUCT.md detected)

**User says:** `$vibe recent`

**What happens:**
1. Agent detects `PRODUCT.md` in project root
2. Agent searches for spec (found: bead na-0042)
3. Agent runs complexity + constraint tests + codex review
4. Agent invokes `$council --deep --preset=code-review --perspectives="api-clarity,error-experience,discoverability" validate recent`
5. Auto-escalation: 6 judges spawn (3 code-review + 3 DX perspectives)
6. Judges review against spec + developer experience criteria

**Result:** Code review augmented with API clarity, error messages, and discoverability checks.

## Fast Inline Check (No Spawning)

**User says:** `$vibe --quick recent`

**What happens:**
1. Agent runs complexity analysis inline (radon/gocyclo)
2. Agent runs constraint tests and codex review
3. Agent performs structured self-review using council schema (no subprocess spawning)
4. Report written to `.agents/council/<timestamp>-vibe-recent.md` labeled `Mode: quick (single-agent)`

**Result:** Sub-60s validation for routine pre-commit checks, no multi-agent overhead.

### go-patterns.md

# Go Patterns Quick Reference - Vibe

Quick reference for Go patterns. Copy-paste these examples when writing new code.

---

## Error Handling

### ✅ Wrap Errors with %w

```go
// DO THIS
if err != nil {
    return fmt.Errorf("failed to initialize: %w", err)
}

// NOT THIS - Breaks error chains (triggers P14)
if err != nil {
    return fmt.Errorf("failed to initialize: %v", err)
}
```

**Prescan:** P14 detects `%v` in `fmt.Errorf` when wrapping errors

### ✅ Custom Error Creation

```go
// Define custom error type
type AppError struct {
    Code    string
    Message string
    Cause   error
}

func (e *AppError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
    }
    return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error { return e.Cause }

func (e *AppError) Is(target error) bool {
    t, ok := target.(*AppError)
    return ok && e.Code == t.Code
}

// Use with errors.Is() and errors.As()
if errors.Is(err, &AppError{Code: "NOT_FOUND"}) {
    // Handle not found
}

var appErr *AppError
if errors.As(err, &appErr) {
    log.Printf("Error code: %s", appErr.Code)
}
```

### ✅ Document Intentional Error Ignores

```go
// DO THIS (passes P13)
defer func() {
    _ = conn.Close() // nolint:errcheck - best effort cleanup
}()

// NOT THIS (triggers P13)
defer func() {
    _ = conn.Close() // Silent ignore
}()
```

**Prescan:** P13 detects `_ =` without `nolint:errcheck` comment

---

## Concurrency

### ✅ Always Use context.Context

```go
// DO THIS
func (c *Client) SendTask(ctx context.Context, task *Task) error {
    req, err := http.NewRequestWithContext(ctx, "POST", url, body)
    if err != nil {
        return fmt.Errorf("creating request: %w", err)
    }
    // ...
}

// NOT THIS - No cancellation support
func (c *Client) SendTask(task *Task) error {
    req, err := http.NewRequest("POST", url, body)
    // ...
}
```

### ✅ WaitGroup Pattern

```go
var wg sync.WaitGroup
for name, agent := range agents {
    wg.Add(1)

    // CRITICAL: Capture loop variables
    name := name
    agent := agent

    go func() {
        defer wg.Done() // Always defer, protects against panic

        if err := agent.Process(ctx); err != nil {
            mu.Lock()
            results[name] = err
            mu.Unlock()
        }
    }()
}
wg.Wait()
```

**Common Mistake:** Forgetting to capture loop variables causes race conditions

### ✅ Mutex for Shared State

```go
type Registry struct {
    items map[string]Item
    mu    sync.RWMutex // Read-write mutex
}

// Read operations use RLock (concurrent reads OK)
func (r *Registry) Get(name string) (Item, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    item, ok := r.items[name]
    if !ok {
        return Item{}, ErrNotFound
    }
    return item, nil
}

// Write operations use Lock (exclusive)
func (r *Registry) Register(name string, item Item) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.items[name] = item
    return nil
}
```

### ✅ Backpressure in Channels

```go
select {
case eventChan <- event:
    // Event sent successfully
case <-time.After(30 * time.Second):
    return fmt.Errorf("event channel blocked - consumer too slow")
case <-ctx.Done():
    return ctx.Err()
}
```

**Why:** Prevents unbounded memory growth from fast producer, slow consumer

---

## Security

### ✅ Constant-Time Comparison

```go
import "crypto/subtle"

// DO THIS - Timing attack resistant
if subtle.ConstantTimeCompare([]byte(token), []byte(expectedToken)) != 1 {
    return ErrUnauthorized
}

// NOT THIS - Vulnerable to timing attacks
if token == expectedToken {
    // Attacker can brute-force byte-by-byte
}
```

**Use Cases:** API keys, tokens, passwords, secrets

### ✅ HMAC Signature Validation

```go
import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
)

func validateHMAC(payload []byte, signature, secret string) bool {
    expectedMAC := hmac.New(sha256.New, []byte(secret))
    expectedMAC.Write(payload)
    expected := hex.EncodeToString(expectedMAC.Sum(nil))

    // Use constant-time comparison
    return hmac.Equal([]byte(expected), []byte(signature))
}
```

**Use Cases:** Webhook signatures (GitHub, GitLab, Slack)

### ✅ Timestamp Validation (Replay Attack Prevention)

```go
func validateTimestamp(ts string, maxAge time.Duration) error {
    timestamp, err := time.Parse(time.RFC3339, ts)
    if err != nil {
        return fmt.Errorf("invalid timestamp: %w", err)
    }

    age := time.Since(timestamp)
    if age > maxAge {
        return fmt.Errorf("timestamp too old: %v > %v", age, maxAge)
    }
    if age < -1*time.Minute {
        return fmt.Errorf("timestamp in future: %v", age)
    }

    return nil
}
```

**Typical maxAge:** 5 minutes for webhooks, 1 minute for API requests

---

## HTTP Clients

### ✅ Proper Body Handling

```go
resp, err := client.Do(req)
if err != nil {
    return fmt.Errorf("request failed: %w", err)
}
defer func() {
    if err := resp.Body.Close(); err != nil {
        log.Printf("Failed to close response body: %v", err)
    }
}()

body, err := io.ReadAll(resp.Body)
if err != nil {
    return fmt.Errorf("reading response: %w", err)
}
```

**CRITICAL:** Always close response body, even on error paths

### ✅ Retry Logic with Exponential Backoff

```go
func (c *Client) doWithRetry(ctx context.Context, req *http.Request) (*http.Response, error) {
    var lastErr error

    for attempt := 0; attempt < c.maxRetries; attempt++ {
        if attempt > 0 {
            backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
            select {
            case <-time.After(backoff):
            case <-ctx.Done():
                return nil, ctx.Err()
            }
        }

        resp, err := c.httpClient.Do(req)
        if err != nil {
            lastErr = err
            continue
        }

        // Retry on 5xx
        if resp.StatusCode >= 500 {
            _ = resp.Body.Close() // nolint:errcheck - best effort
            lastErr = fmt.Errorf("server error: %d", resp.StatusCode)
            continue
        }

        return resp, nil
    }

    return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}
```

---

## Interface Design

### ✅ Accept Interfaces, Return Structs

```go
// Define interface
type Processor interface {
    Process(ctx context.Context, data []byte) error
}

// Functions accept interface (flexible for testing)
func RunPipeline(ctx context.Context, processor Processor, data []byte) error {
    if err := processor.Process(ctx, data); err != nil {
        return fmt.Errorf("processing failed: %w", err)
    }
    return nil
}

// Constructors return struct (concrete)
func NewDataProcessor() *DataProcessor {
    return &DataProcessor{
        cache: make(map[string][]byte),
        mu:    sync.RWMutex{},
    }
}
```

**Why:** Callers can pass any implementation (testability), return type can add methods

### ✅ Small, Focused Interfaces

```go
// DO THIS - Single responsibility
type Initializer interface {
    Initialize(ctx context.Context) error
}

type Processor interface {
    Process(ctx context.Context, data []byte) error
}

// Compose when needed
type Service interface {
    Initializer
    Processor
}

// NOT THIS - God interface
type Service interface {
    Initialize(ctx context.Context) error
    Process(ctx context.Context, data []byte) error
    Shutdown(ctx context.Context) error
    HealthCheck(ctx context.Context) error
    GetMetrics() *Metrics
    SetConfig(cfg *Config)
    // ... 20 more methods
}
```

---

## Testing

### ✅ Table-Driven Tests

```go
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {"valid", "user@example.com", false},
        {"missing @", "userexample.com", true},
        {"empty", "", true},
        {"no domain", "user@", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEmail(tt.email)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateEmail(%q) error = %v, wantErr %v",
                    tt.email, err, tt.wantErr)
            }
        })
    }
}
```

### ✅ Test Helpers

```go
func setupTestServer(t *testing.T) *httptest.Server {
    t.Helper() // Mark as helper - failures report caller line

    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Mock responses
    }))

    t.Cleanup(func() {
        server.Close()
    })

    return server
}

func TestClient(t *testing.T) {
    server := setupTestServer(t) // Failures report this line, not inside helper
    client := NewClient(server.URL)
    // ... test code
}
```

### ✅ Mock Interfaces

```go
// Define mockable interface
type Repository interface {
    GetUser(ctx context.Context, id string) (*User, error)
}

// Create mock
type MockRepository struct {
    GetUserFn func(ctx context.Context, id string) (*User, error)
}

func (m *MockRepository) GetUser(ctx context.Context, id string) (*User, error) {
    if m.GetUserFn != nil {
        return m.GetUserFn(ctx, id)
    }
    return nil, nil
}

// Use in tests
func TestService(t *testing.T) {
    mock := &MockRepository{
        GetUserFn: func(ctx context.Context, id string) (*User, error) {
            return &User{ID: id, Name: "Test"}, nil
        },
    }

    service := NewService(mock)
    user, err := service.FetchUser(ctx, "123")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if user.Name != "Test" {
        t.Errorf("got name %q, want %q", user.Name, "Test")
    }
}
```

---

## Common Mistakes to Avoid

### ❌ Don't: Ignore Context

```go
// BAD
func (c *Client) Send(req *Request) error {
    // No way to cancel or timeout
    return c.process(req)
}

// GOOD
func (c *Client) Send(ctx context.Context, req *Request) error {
    select {
    case result := <-c.process(req):
        return result
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

### ❌ Don't: Use Pointer to Interface

```go
// BAD
func Process(agent *Agent) error { // Interface is already a reference
    // ...
}

// GOOD
func Process(agent Agent) error {
    // ...
}
```

### ❌ Don't: Naked Returns with Named Results

```go
// BAD
func calculate() (result int, err error) {
    result = 42
    return // What's being returned? Unclear!
}

// GOOD
func calculate() (int, error) {
    result := 42
    return result, nil // Explicit and clear
}
```

### ❌ Don't: Use panic in Library Code

```go
// BAD
func GetItem(key string) Item {
    item, ok := registry[key]
    if !ok {
        panic("item not found") // Caller can't recover
    }
    return item
}

// GOOD
func GetItem(key string) (Item, error) {
    item, ok := registry[key]
    if !ok {
        return Item{}, ErrItemNotFound
    }
    return item, nil
}
```

### ❌ Don't: Forget to Capture Loop Variables

```go
// BAD
for _, item := range items {
    go func() {
        process(item) // Race condition! All goroutines see last item
    }()
}

// GOOD
for _, item := range items {
    item := item // Capture variable
    go func() {
        process(item) // Each goroutine has its own copy
    }()
}
```

### ❌ Don't: Leak Goroutines

```go
// BAD - Goroutine never exits
go func() {
    for {
        work() // No way to stop
    }
}()

// GOOD - Context-based cancellation
go func() {
    for {
        select {
        case <-ctx.Done():
            return
        default:
            work()
        }
    }
}()
```

---

## Code Review Checklist

Before submitting PR, verify:

- [ ] All errors wrapped with `%w` (P14 passing)
- [ ] All long operations accept `context.Context`
- [ ] All `defer` statements have error checking where needed
- [ ] Loop variables captured before goroutines
- [ ] HTTP response bodies closed with defer
- [ ] Secrets compared with `subtle.ConstantTimeCompare()`
- [ ] Intentional error ignores documented with `nolint:errcheck` (P13 passing)
- [ ] Tests use table-driven pattern
- [ ] Test helpers use `t.Helper()`
- [ ] Interfaces are small and focused
- [ ] Functions return concrete types (not interfaces)
- [ ] golangci-lint passes (P15)
- [ ] gofmt applied
- [ ] Complexity < 10 per function

---

## golangci-lint Commands

```bash
# Run all linters
golangci-lint run ./...

# Run specific linter
golangci-lint run --enable=errcheck ./...

# Fix auto-fixable issues
golangci-lint run --fix ./...

# Show configuration
golangci-lint linters
```

---

## Useful Commands

```bash
# Format code
gofmt -w .

# Check formatting
gofmt -l .

# Vet code
go vet ./...

# Run tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Check for race conditions
go test -race ./...

# Build all binaries
go build ./cmd/...

# Tidy dependencies
go mod tidy

# Check cyclomatic complexity
gocyclo -over 10 .
```

---

## Prescan Pattern Reference

| Pattern | Severity | Triggers On |
|---------|----------|-------------|
| P13 | HIGH | `_ =` without `nolint:errcheck` comment |
| P14 | MEDIUM | `fmt.Errorf.*%v` when wrapping errors |
| P15 | HIGH | golangci-lint violations (requires golangci-lint installed) |

Run prescan: `~/.codex/skills/vibe/scripts/prescan.sh recent`

---

## Additional Resources

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Go Proverbs](https://go-proverbs.github.io/)
- [golangci-lint Linters](https://golangci-lint.run/usage/linters/)
- [errcheck](https://github.com/kisielk/errcheck)

---

**See Also:** `go-standards.md` for comprehensive catalog with detailed explanations

### go-standards.md

# Go Standards Catalog - Vibe Canonical Reference

**Version:** 1.0.0
**Last Updated:** 2026-01-20
**Purpose:** Canonical Go standards for vibe skill validation

---

## Table of Contents

1. [Error Handling Patterns](#error-handling-patterns)
2. [Interface Design](#interface-design)
3. [Concurrency Patterns](#concurrency-patterns)
4. [Security Practices](#security-practices)
5. [Package Organization](#package-organization)
6. [Testing Patterns](#testing-patterns)
7. [Documentation Standards](#documentation-standards)
8. [Code Quality Metrics](#code-quality-metrics)
9. [Anti-Patterns Avoided](#anti-patterns-avoided)

---

## Error Handling Patterns

### ✅ **Custom Error Types**

Production-grade error types follow these patterns:

```go
type AppError struct {
    Code     string        // Machine-readable error code
    Message  string        // Human-readable message
    Cause    error         // Wrapped error (optional)
    Metadata map[string]any // Additional context
}

// Implements error interface
func (e *AppError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
    }
    return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Supports errors.Unwrap()
func (e *AppError) Unwrap() error {
    return e.Cause
}

// Supports errors.Is() for sentinel comparison
func (e *AppError) Is(target error) bool {
    t, ok := target.(*AppError)
    if !ok {
        return false
    }
    return e.Code == t.Code
}
```

**Requirements:**
- ✅ Implements `error` interface
- ✅ Implements `Unwrap()` for error chain inspection
- ✅ Implements `Is()` for sentinel error comparison
- ✅ Structured error codes enable programmatic handling
- ✅ Preserves context with metadata
- ✅ Proper nil-safety in `Unwrap()` and `Is()`

### ✅ **Error Wrapping with %w**

Use `fmt.Errorf` with `%w` verb for error wrapping:

```go
// CORRECT
resp, err := client.Do(req)
if err != nil {
    return nil, fmt.Errorf("sending request: %w", err)
}

// INCORRECT - Breaks error chains
if err != nil {
    return nil, fmt.Errorf("sending request: %v", err)
}
```

**Why This Matters:**
- `%w` preserves error chain for `errors.Is()` and `errors.As()`
- `%v` breaks the chain - root cause is lost
- Error context adds debugging information

### ⚠️ **Intentional Error Ignores**

Document why errors are intentionally ignored:

```go
// CORRECT
defer func() {
    _ = conn.Close() // nolint:errcheck - best effort cleanup
}()

// INCORRECT - Silent ignore
defer func() {
    _ = conn.Close()
}()
```

**Validation:** Prescan pattern P13 detects undocumented ignores

---

## Interface Design

### ✅ **Accept Interfaces, Return Structs**

**Pattern:**
```go
// Define interface
type Agent interface {
    Initialize(ctx context.Context) error
    Invoke(ctx context.Context, req *Request) (*Response, error)
}

// Functions accept interface (flexible)
func ProcessAgent(ctx context.Context, agent Agent) error {
    if err := agent.Initialize(ctx); err != nil {
        return fmt.Errorf("initialization failed: %w", err)
    }
    // ...
}

// Constructors return struct (concrete)
func NewRegistry() *Registry {
    return &Registry{
        agents: make(map[string]Agent),
        mu:     sync.RWMutex{},
    }
}
```

**Why This Matters:**
- Callers can pass any implementation (testability)
- Return type can add methods without breaking callers
- Follows Go proverb: "Be conservative in what you send, liberal in what you accept"

### ✅ **Small, Focused Interfaces**

**Good Example:**
```go
type Initializer interface {
    Initialize(ctx context.Context) error
}

type Invoker interface {
    Invoke(ctx context.Context, req *Request) (*Response, error)
}

// Compose interfaces
type Agent interface {
    Initializer
    Invoker
}
```

**Anti-Pattern (God Interface):**
```go
type Agent interface {
    Initialize(ctx context.Context) error
    Invoke(ctx context.Context, req *Request) (*Response, error)
    Shutdown(ctx context.Context) error
    HealthCheck(ctx context.Context) error
    GetMetrics() *Metrics
    SetConfig(cfg *Config)
    // ... 20 more methods
}
```

---

## Concurrency Patterns

### ✅ **Context Propagation** (Required)

Every I/O or long-running operation accepts `context.Context`:

```go
// HTTP Requests
req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)

// Database Operations
rows, err := db.QueryContext(ctx, query)

// Custom Functions
func (c *Client) Invoke(ctx context.Context, req *Request) (*Response, error)
```

**Benefits:**
- Timeout propagation
- Cancellation support
- Request-scoped values (tracing)

### ✅ **Proper WaitGroup Usage**

```go
var wg sync.WaitGroup
for name, agent := range agents {
    wg.Add(1)

    // Capture loop variables
    name := name
    agent := agent

    go func() {
        defer wg.Done() // Always defer, protects against panic

        if err := agent.Process(ctx); err != nil {
            mu.Lock()
            results[name] = err
            mu.Unlock()
        }
    }()
}
wg.Wait()
```

**Requirements:**
- ✅ Variables captured before goroutine (avoids closure bug)
- ✅ `defer wg.Done()` ensures decrement on panic
- ✅ Mutex protects shared data structures
- ✅ Context cancellation checked in each goroutine

### ✅ **Thread-Safe Data Structures**

```go
type Registry struct {
    items map[string]Item
    mu    sync.RWMutex // Read-write mutex
}

// Read operations use RLock
func (r *Registry) Get(key string) (Item, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    // ...
}

// Write operations use Lock
func (r *Registry) Set(key string, item Item) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    // ...
}
```

**Pattern Benefits:**
- Multiple concurrent reads
- Exclusive writes
- Zero race conditions

### ✅ **Backpressure in Streaming**

```go
select {
case eventChan <- event:
    // Event sent successfully
case <-time.After(30 * time.Second):
    return fmt.Errorf("event channel blocked - consumer too slow (backpressure triggered)")
case <-ctx.Done():
    return ctx.Err()
}
```

**Why This Matters:**
- Prevents unbounded memory growth
- Handles fast producer, slow consumer scenario
- Explicit timeout for debugging

---

## Security Practices

### ✅ **Constant-Time Comparison** (Timing Attack Prevention)

```go
import "crypto/subtle"

// CORRECT - Timing attack resistant
token := r.Header.Get("Authorization")
if subtle.ConstantTimeCompare([]byte(token), []byte(expectedToken)) != 1 {
    return ErrUnauthorized
}

// INCORRECT - Vulnerable to timing attacks
if token == expectedToken {
    // Attacker can brute-force byte-by-byte
}
```

**Why This Matters:**
- String comparison (`==`) leaks timing information
- Attacker can brute-force secrets byte-by-byte
- `subtle.ConstantTimeCompare()` runs in constant time
- Critical for API keys, tokens, passwords

### ✅ **HMAC Signature Validation**

```go
import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
)

func validateHMAC(payload []byte, signature, secret string) bool {
    if !strings.HasPrefix(signature, "sha256=") {
        return false
    }

    expectedMAC := hmac.New(sha256.New, []byte(secret))
    expectedMAC.Write(payload)
    expected := "sha256=" + hex.EncodeToString(expectedMAC.Sum(nil))

    return hmac.Equal([]byte(expected), []byte(signature))
}
```

**Security Features:**
- ✅ HMAC prevents payload tampering
- ✅ Uses `hmac.Equal()` (constant-time)
- ✅ Verifies signature format first
- ✅ SHA-256 (secure hash function)

### ✅ **Replay Attack Prevention**

```go
func validateTimestamp(timestamp string, maxAge time.Duration) error {
    ts, err := time.Parse(time.RFC3339, timestamp)
    if err != nil {
        return fmt.Errorf("invalid timestamp format")
    }

    age := time.Since(ts)
    if age > maxAge || age < -1*time.Minute {
        return fmt.Errorf("request too old or in future: age=%v max=%v", age, maxAge)
    }

    return nil
}
```

**Protection Against:**
- Replay attacks (old requests resubmitted)
- Clock skew (1 minute tolerance for future timestamps)
- DoS via timestamp manipulation

### ✅ **TLS Configuration**

```go
tlsConfig := &tls.Config{
    MinVersion: tls.VersionTLS13, // Only TLS 1.3+
    // No InsecureSkipVerify - validates certificates
}
```

---

## Package Organization

### ✅ **Layered Architecture**

```
project/
├── cmd/                    # Binaries (main packages)
│   ├── server/            # Server binary
│   ├── worker/            # Worker binary
│   └── cli/               # CLI tool
├── internal/              # Private packages (cannot be imported externally)
│   ├── domain/            # Business logic
│   ├── handlers/          # HTTP handlers
│   ├── repository/        # Data access
│   └── sdk/               # External SDK clients
├── pkg/                   # Public packages (can be imported)
│   ├── api/              # API types
│   └── client/           # Client library
└── tests/                # Test suites
    ├── e2e/              # End-to-end tests
    └── integration/      # Integration tests
```

**Principles:**
- ✅ `cmd/` for binaries (no importable code)
- ✅ `internal/` prevents external imports
- ✅ `pkg/` for public APIs
- ✅ Domain-driven structure
- ✅ Tests at package level, e2e/integration separate

### ✅ **Import Grouping** (Go Convention)

```go
import (
    // Standard library
    "context"
    "fmt"
    "time"

    // External dependencies
    "github.com/external/package"

    // Internal packages
    "myproject.com/internal/domain"
)
```

---

## Testing Patterns

### ✅ **Table-Driven Tests**

```go
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {"valid", "user@example.com", false},
        {"missing @", "userexample.com", true},
        {"empty", "", true},
        {"no domain", "user@", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEmail(tt.email)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateEmail(%q) error = %v, wantErr %v",
                    tt.email, err, tt.wantErr)
            }
        })
    }
}
```

**Benefits:**
- Easy to add test cases
- Clear test names with `t.Run()`
- DRY (Don't Repeat Yourself)

### ✅ **Test Helpers with t.Helper()**

```go
func setupTestServer(t *testing.T) *httptest.Server {
    t.Helper() // Marks this as a helper function

    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Mock responses
    }))

    t.Cleanup(func() {
        server.Close()
    })

    return server
}

func TestClient(t *testing.T) {
    server := setupTestServer(t) // Failures report this line, not inside helper
    // ... test code
}
```

**Why t.Helper() Matters:**
- Test failures report the *calling* line, not helper line
- Makes test output more useful
- Standard Go testing pattern

### ✅ **Mock Interfaces**

```go
// Define mockable interface
type Invoker interface {
    Invoke(ctx context.Context, req *Request) (*Response, error)
}

// Create mock
type MockInvoker struct {
    InvokeFn func(ctx context.Context, req *Request) (*Response, error)
}

func (m *MockInvoker) Invoke(ctx context.Context, req *Request) (*Response, error) {
    if m.InvokeFn != nil {
        return m.InvokeFn(ctx, req)
    }
    return nil, nil
}

// Use in tests
func TestProcessor(t *testing.T) {
    mock := &MockInvoker{
        InvokeFn: func(ctx context.Context, req *Request) (*Response, error) {
            return &Response{Status: "success"}, nil
        },
    }

    processor := NewProcessor(mock)
    // ... test with mock
}
```

### Test Double Types

| Type | Purpose | When to Use |
|------|---------|-------------|
| **Stub** | Returns canned data | Simple happy path |
| **Mock** | Verifies interactions | Behavior verification |
| **Fake** | Working implementation | Integration-like tests |
| **Spy** | Records calls | Interaction counting |

---

## Documentation Standards

### ✅ **Godoc Format**

Document all exported symbols with a comment directly above the declaration:

```go
// Registry manages agent lifecycle and discovery.
// It is safe for concurrent use.
type Registry struct {
    agents map[string]Agent
    mu     sync.RWMutex
}

// Get returns the agent registered under the given key.
// It returns ErrNotFound if no agent is registered with that key.
func (r *Registry) Get(key string) (Agent, error) {
    // ...
}
```

**Rules:**
- Comment starts with the name of the symbol
- First sentence is a complete summary (used by `go doc -short`)
- Use `//` comments, not `/* */` blocks (except for package-level docs)

### ✅ **Package-Level Comments**

For packages with significant public API, use a `doc.go` file:

```go
// Package registry provides agent lifecycle management
// including registration, discovery, and health monitoring.
//
// Basic usage:
//
//	reg := registry.New()
//	reg.Register("my-agent", agent)
//	a, err := reg.Get("my-agent")
package registry
```

**When to use `doc.go`:**
- Package has 3+ exported symbols
- Package is part of a public API (`pkg/`)
- Package needs usage examples beyond a single line

### ✅ **Testable Examples**

Write `Example*` functions in `_test.go` files — they appear in generated docs and are compiled/run by `go test`:

```go
func ExampleRegistry_Get() {
    reg := registry.New()
    reg.Register("agent-1", &MyAgent{Name: "alpha"})

    agent, err := reg.Get("agent-1")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(agent.Name)
    // Output: alpha
}
```

**Naming Convention:**
- `ExampleTypeName` — type-level example
- `ExampleTypeName_MethodName` — method-level example
- `Example` — package-level example

### ✅ **Doc Generation**

```bash
# View docs in terminal
go doc ./pkg/registry
go doc ./pkg/registry.Registry.Get

# Run local doc server (pkgsite)
go install golang.org/x/pkgsite/cmd/pkgsite@latest
pkgsite -open .
```

### ✅ **Interface Documentation**

Interfaces define contracts — document the behavioral expectations, not just the signature:

```go
// Store persists agent state across restarts.
//
// Implementations must be safe for concurrent use.
// All methods must respect context cancellation.
type Store interface {
    // Save persists the agent. It returns ErrConflict if the agent
    // was modified since it was last read (optimistic locking).
    Save(ctx context.Context, agent *Agent) error

    // Load retrieves an agent by ID. It returns ErrNotFound if
    // no agent exists with the given ID.
    Load(ctx context.Context, id string) (*Agent, error)
}
```

**Guidelines:**
- Document concurrency guarantees on the interface comment
- Document error contracts on each method (which sentinel errors are returned)
- Document preconditions and postconditions when non-obvious

### ✅ **Internal Package Documentation**

`internal/` packages cannot be imported outside the module, but still need documentation for team maintainability:

```go
// Package repository implements data access for agent storage.
//
// This is an internal package — it should not be imported outside
// the module. Use pkg/client for the public API.
package repository
```

**Guidelines:**
- Every `internal/` package needs a package comment explaining its role
- Note the public alternative if one exists (e.g., `pkg/client`)
- Document non-obvious design constraints (e.g., "not safe for concurrent use")

### ✅ **Package README Files**

For packages with significant scope, include a `README.md` alongside the Go source:

```
pkg/registry/
├── README.md          # Setup instructions, architecture notes
├── doc.go             # Godoc package comment
├── registry.go        # Implementation
└── registry_test.go   # Tests with examples
```

**When to include a README:**
- Package requires setup steps (config, env vars, migrations)
- Package has architecture or design decisions worth explaining
- Package is a top-level entry point (`cmd/`, major `pkg/` packages)

**README vs doc.go:**
- `doc.go` → API usage shown in `go doc` output
- `README.md` → Setup, architecture, diagrams, non-API context

### ✅ **Comment Style**

Follow Go's documentation conventions for consistent, tooling-friendly comments:

```go
// ProcessBatch sends all queued events to the remote collector.
// It returns the number of events successfully delivered and
// a non-nil error if the connection to the collector fails.
//
// ProcessBatch is safe for concurrent use. Each call acquires
// a connection from the pool and releases it on return.
func (c *Client) ProcessBatch(ctx context.Context) (int, error) {
    // ...
}

// ErrRateLimited is returned when the collector rejects a request
// due to rate limiting. Callers should back off and retry.
var ErrRateLimited = errors.New("rate limited")
```

**Rules:**
- Write complete sentences with proper punctuation
- First word is the name of the declared thing (`ProcessBatch sends...`, `ErrRateLimited is...`)
- First sentence stands alone as a summary — `go doc -short` shows only this
- Use third-person declarative ("ProcessBatch sends...") not imperative ("Send...")
- Separate paragraphs with a blank `//` line
- Use `[Registry.Get]` syntax (Go 1.19+) to link to other symbols in doc comments
- Keep line length under 80 characters for readability in terminals

**Anti-Patterns:**
```go
// BAD - Doesn't start with symbol name
// This function processes a batch of events.
func (c *Client) ProcessBatch(ctx context.Context) (int, error)

// BAD - Not a complete sentence
// process batch
func (c *Client) ProcessBatch(ctx context.Context) (int, error)

// BAD - Imperative instead of declarative
// Send all queued events to the remote collector.
func (c *Client) ProcessBatch(ctx context.Context) (int, error)
```

### ALWAYS / NEVER Rules

| Rule | Rationale |
|------|-----------|
| **ALWAYS** document exported types, functions, and methods | Required by `revive` linter, enables `go doc` |
| **ALWAYS** start doc comments with the symbol name | Standard godoc convention, enables tooling |
| **ALWAYS** write doc comments as complete sentences | Consistent style, readable in `go doc` output |
| **ALWAYS** include `// Output:` in Example functions | Makes examples testable by `go test` |
| **ALWAYS** document interface contracts (thread-safety, errors, lifecycle) | Callers depend on the contract, not the implementation |
| **NEVER** document unexported symbols unless logic is non-obvious | Noise — internal code changes frequently |
| **NEVER** use `@param` / `@return` javadoc-style annotations | Not idiomatic Go — godoc ignores them |
| **NEVER** duplicate the function signature in prose | Redundant — the signature is right below |
| **NEVER** use imperative voice in doc comments | Go convention is declarative third-person |

---

## Structured Logging (slog)

### ✅ **Use log/slog (Go 1.21+)**

```go
import "log/slog"

func main() {
    // Production: JSON handler for log aggregation
    logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    }))
    slog.SetDefault(logger)

    // Include correlation IDs for tracing
    slog.Info("request processed",
        "request_id", reqID,
        "user_id", userID,
        "duration_ms", duration.Milliseconds(),
    )
}
```

### Handler Selection

| Environment | Handler | Use Case |
|-------------|---------|----------|
| Production | `slog.JSONHandler` | Elasticsearch, Loki, CloudWatch |
| Development | `slog.TextHandler` | Human-readable console output |

### ❌ **Logging Anti-Patterns**

| Pattern | Problem | Instead |
|---------|---------|---------|
| `fmt.Println` in library | Not parseable, no levels | Use `slog.Info` |
| `log.Printf` | No structure | Use `slog` with attributes |
| Logging secrets | Security risk | Use `ReplaceAttr` to redact |
| Missing correlation ID | Can't trace requests | Always include request_id |

> **Talos check:** PRE-007 detects `fmt.Print*` debug statements in non-CLI code.

---

## Benchmarking and Profiling

### ✅ **Writing Benchmarks**

```go
func BenchmarkProcess(b *testing.B) {
    data := setupTestData()
    b.ResetTimer() // Exclude setup from timing

    for i := 0; i < b.N; i++ {
        Process(data)
    }
}

// Memory allocation benchmark
func BenchmarkProcessAllocs(b *testing.B) {
    data := setupTestData()
    b.ResetTimer()
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        Process(data)
    }
}
```

### Running Benchmarks

```bash
# Run benchmarks
go test -bench=. -benchmem ./...

# Compare before/after
go test -bench=. -count=10 > old.txt
# make changes
go test -bench=. -count=10 > new.txt
benchstat old.txt new.txt
```

### ✅ **Profiling with pprof**

```go
import _ "net/http/pprof"

// Profiles available at:
// /debug/pprof/profile  - CPU profile
// /debug/pprof/heap     - Memory profile
// /debug/pprof/goroutine - Goroutine stacks
```

**Analyze Profiles:**
```bash
# CPU profile (30 seconds)
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Memory profile
go tool pprof http://localhost:6060/debug/pprof/heap

# Interactive commands
(pprof) top10      # Top 10 functions
(pprof) web        # Open flame graph in browser
```

---

## Configuration Management

### ✅ **Single Config Struct Pattern**

```go
type Config struct {
    Server   ServerConfig   `yaml:"server"`
    Database DatabaseConfig `yaml:"database"`
    Log      LogConfig      `yaml:"log"`
}

type ServerConfig struct {
    Port         int           `yaml:"port" env:"PORT"`
    ReadTimeout  time.Duration `yaml:"read_timeout"`
    WriteTimeout time.Duration `yaml:"write_timeout"`
}

// Load with precedence: flags > env > file > defaults
func Load() (*Config, error) {
    cfg := &Config{}
    setDefaults(cfg)
    if err := loadFromFile(cfg); err != nil {
        return nil, fmt.Errorf("load config file: %w", err)
    }
    loadFromEnv(cfg)
    if err := cfg.Validate(); err != nil {
        return nil, fmt.Errorf("validate config: %w", err)
    }
    return cfg, nil
}
```

### ❌ **Configuration Anti-Patterns**

| Pattern | Problem | Instead |
|---------|---------|---------|
| Global config var | Hard to test | Pass as dependency |
| Reading env in functions | Scattered config | Centralize in Load() |
| No validation | Runtime errors | Validate at startup |
| Secrets in config files | Security risk | Use env vars or vault |

---

## HTTP API Standards

### ✅ **API Versioning**

```go
mux := http.NewServeMux()

// Health endpoints (unversioned - K8s standard)
mux.HandleFunc("/health", healthHandler)
mux.HandleFunc("/healthz", healthHandler)   // K8s liveness
mux.HandleFunc("/readyz", readyHandler)     // K8s readiness

// API v1
mux.HandleFunc("/v1/webhook/gitlab", handler.ServeHTTP)

// API documentation
mux.HandleFunc("/openapi.json", openAPIHandler)
```

### ✅ **Health Endpoints**

```go
func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"status":"healthy"}`))
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
    if !dependenciesReady() {
        w.WriteHeader(http.StatusServiceUnavailable)
        w.Write([]byte(`{"status":"not ready"}`))
        return
    }
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"status":"ready"}`))
}
```

### ✅ **Server Configuration**

```go
server := &http.Server{
    Addr:         ":" + port,
    Handler:      loggingMiddleware(mux),
    ReadTimeout:  15 * time.Second,
    WriteTimeout: 15 * time.Second,
    IdleTimeout:  60 * time.Second,
}

// Graceful shutdown
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

go func() {
    if err := server.ListenAndServe(); err != http.ErrServerClosed {
        log.Fatalf("Server failed: %v", err)
    }
}()

<-quit
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
server.Shutdown(ctx)
```

### ❌ **HTTP API Anti-Patterns**

| Pattern | Problem | Instead |
|---------|---------|---------|
| Unversioned API | Breaking changes affect all | `/v1/webhook/gitlab` |
| No Health Endpoint | K8s can't probe | Add `/health`, `/readyz` |
| No OpenAPI Spec | Undocumented API | Serve OpenAPI 3.0 |
| No Timeout Config | Slow clients block | Set Read/Write timeouts |
| No Graceful Shutdown | Dropped requests | Catch signals, drain |

---

## Kubernetes Operator Patterns

### ✅ **Controller Reconciliation**

```go
func (r *MyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    var resource myv1.MyResource
    if err := r.Get(ctx, req.NamespacedName, &resource); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    // Handle deletion with finalizer
    if !resource.DeletionTimestamp.IsZero() {
        return r.handleDeletion(ctx, &resource)
    }

    // Add finalizer if not present
    if !controllerutil.ContainsFinalizer(&resource, myFinalizer) {
        controllerutil.AddFinalizer(&resource, myFinalizer)
        if err := r.Update(ctx, &resource); err != nil {
            return ctrl.Result{}, err
        }
        return ctrl.Result{Requeue: true}, nil
    }

    // State machine based on desired state
    switch resource.Spec.DesiredState {
    case myv1.StateActive:
        return r.ensureActive(ctx, &resource)
    case myv1.StateIdle:
        return r.ensureIdle(ctx, &resource)
    }
    return ctrl.Result{}, nil
}
```

### Return Patterns

| Result | Meaning |
|--------|---------|
| `ctrl.Result{}, nil` | Success, no requeue |
| `ctrl.Result{Requeue: true}, nil` | Requeue immediately |
| `ctrl.Result{RequeueAfter: time.Minute}, nil` | Requeue after duration |
| `ctrl.Result{}, err` | Error, controller-runtime handles backoff |

### ❌ **Operator Anti-Patterns**

| Pattern | Problem | Instead |
|---------|---------|---------|
| Status as Spec | Status is observed, not desired | Use Spec for desired |
| Missing Finalizer | Orphaned external resources | Add finalizer first |
| No Context Timeout | Hung operations | `context.WithTimeout` |
| Condition Storms | Triggers unnecessary watches | Update only on change |
| Direct Status Update | Conflicts with spec updates | Use `r.Status().Update()` |

---

## Code Quality Metrics

> See `common-standards.md` for universal coverage targets and testing principles.

### ✅ **golangci-lint Configuration**

Minimum recommended linters:

```yaml
# .golangci.yml
linters:
  enable:
    - errcheck      # Check error returns
    - govet         # Go vet
    - staticcheck   # Advanced static analysis
    - unused        # Detect unused code
    - gosimple      # Simplification suggestions
    - gocritic      # Opinionated checks
    - misspell      # Spell checking
    - errorlint     # Error wrapping checks
    - goimports     # Auto-organize imports
    - revive        # Exported name checks

linters-settings:
  gocyclo:
    min-complexity: 10  # Cyclomatic complexity threshold
```

### 📊 **Complexity Thresholds**

| Complexity Range | Status | Action |
|-----------------|--------|--------|
| CC 1-5 (Simple) | ✅ Excellent | Maintain |
| CC 6-10 (OK) | ✅ Acceptable | Monitor |
| CC 11-15 (High) | ⚠️ Warning | Refactor recommended |
| CC 16+ (Very High) | ❌ Critical | Refactor required |

**Refactoring Strategies:**
- Strategy maps (replace switch statements)
- Guard clauses (early returns)
- Helper functions (extract validation)
- Interface composition

---

## Anti-Patterns Avoided

> See `common-standards.md` for universal anti-patterns across all languages.

### ❌ **No Naked Returns**
```go
// BAD
func bad() (err error) {
    err = doSomething()
    return // Naked return
}

// GOOD
func good() error {
    err := doSomething()
    return err // Explicit return
}
```

### ❌ **No init() Abuse**
- No `init()` functions with side effects
- Configuration via constructors
- Explicit initialization with error handling

### ❌ **No Panics in Library Code**
- All errors returned via `error` interface
- `panic` only used in tests for assertion failures
- No `panic` in production paths

### ❌ **No Global Mutable State**
```go
// BAD
var globalRegistry *Registry

// GOOD
type Server struct {
    registry *Registry // Instance field
}
```

### ❌ **No Pointer to Interface**
```go
// BAD
func bad(agent *Agent) // Interface is already a reference

// GOOD
func good(agent Agent)
```

### ❌ **No Goroutine Leaks**
```go
// BAD - Goroutine never exits
go func() {
    for {
        work() // No way to stop
    }
}()

// GOOD - Context-based cancellation
go func() {
    for {
        select {
        case <-ctx.Done():
            return
        default:
            work()
        }
    }
}()
```

---

## Compliance Assessment

**Use letter grades + evidence, NOT numeric scores.**

| Category | Assessment Criteria | Evidence Required |
|----------|-------------------|-------------------|
| Error Handling | Custom errors, %w wrapping, documented ignores | Count proper wrappings, undocumented ignores |
| Interface Design | Accept interfaces, return structs, small interfaces | Count interfaces, methods per interface |
| Concurrency | Context propagation, WaitGroups, mutexes | Activities with context, race condition count |
| Security | Constant-time comparison, HMAC, replay prevention | Prescan P2 findings, hardcoded secrets count |
| Code Organization | Layered architecture, import grouping | Package structure review, import violations |
| Testing | Table-driven, helpers, mocks | Test pattern count, coverage percentage |

**Grading Scale:**

| Grade | Finding Threshold | Description |
|-------|------------------|-------------|
| A+ | 0-2 minor findings | Exemplary - industry best practices |
| A | <5 HIGH findings | Excellent - strong practices |
| A- | 5-15 HIGH findings | Very Good - solid practices |
| B+ | 15-25 HIGH findings | Good - acceptable practices |
| B | 25-40 HIGH findings | Satisfactory - needs improvement |
| C+ | 40-60 HIGH findings | Needs Improvement - multiple issues |
| C | 60+ HIGH findings | Significant Issues - major refactoring |
| D | 1+ CRITICAL findings | Major Problems - not production-ready |
| F | Multiple CRITICAL | Critical Issues - complete rewrite |

**Example Assessment:**

| Category | Grade | Evidence |
|----------|-------|----------|
| Error Handling | A- | 131 proper %w wrappings, 5 undocumented ignores, 0 %v issues |
| Interface Design | A+ | 9 small interfaces (avg 4 methods), proper composition |
| Concurrency | A | 24/24 activities use context, 0 race conditions (go test -race) |
| Security | A | 0 CRITICAL, 2 HIGH (P2 findings), timing-safe comparisons |
| **OVERALL** | **A- (Excellent)** | **12 HIGH, 34 MEDIUM findings** |

---

## Vibe Integration

### Prescan Patterns

| Pattern | Severity | Detection |
|---------|----------|-----------|
| P13: Undocumented Error Ignores | HIGH | `_ =` without `nolint:errcheck` |
| P14: Error Wrapping with %v | MEDIUM | `fmt.Errorf.*%v` with error args |
| P15: golangci-lint Violations | HIGH | JSON output parsing |

### Semantic Analysis

Deep validation includes:
- Error chain inspection (`errors.Is`, `errors.As` usage)
- Interface segregation (ISP compliance)
- Goroutine lifecycle analysis
- Security vulnerability detection

### JIT Loading

**Tier 1 (Fast):** Load `~/.agents/skills/standards/references/go.md` (5KB)
**Tier 2 (Deep):** Load this document (16KB) for comprehensive audit
**Override:** Use `.agents/validation/GO_*.md` if project-specific standards exist

---

## Additional Resources

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Go Proverbs](https://go-proverbs.github.io/)
- [golangci-lint Linters](https://golangci-lint.run/usage/linters/)
- [OWASP Go Secure Coding](https://owasp.org/www-project-go-secure-coding-practices-guide/)

---

**Related:** `go-patterns.md` for quick reference examples

### json-standards.md

# JSON/JSONL Standards Catalog - Vibe Canonical Reference

**Version:** 1.0.0
**Last Updated:** 2026-01-21
**Purpose:** Canonical JSON/JSONL standards for vibe skill validation

---

## Table of Contents

1. [JSON Formatting](#json-formatting)
2. [JSONL Format](#jsonl-format)
3. [Beads JSONL Schema](#beads-jsonl-schema)
4. [Configuration Files](#configuration-files)
5. [JSON Schema](#json-schema)
6. [Tooling](#tooling)
7. [Anti-Patterns](#anti-patterns)
8. [Code Quality Metrics](#code-quality-metrics)
9. [Prescan Patterns](#prescan-patterns)
10. [Compliance Assessment](#compliance-assessment)

---

## JSON Formatting

### Standard Format

```json
{
  "name": "example",
  "version": "1.0.0",
  "config": {
    "timeout": 30,
    "retries": 3,
    "enabled": true
  },
  "items": [
    "first",
    "second",
    "third"
  ]
}
```

### Formatting Rules

| Rule | Example | Why |
|------|---------|-----|
| 2-space indent | `  "key": "value"` | Readability |
| Double quotes only | `"key"` not `'key'` | JSON spec |
| No trailing commas | `["a", "b"]` | JSON spec |
| Trailing newline | File ends with `\n` | POSIX, git diffs |
| UTF-8 encoding | Always | Compatibility |

### Key Naming Conventions

| Convention | Use For | Example |
|------------|---------|---------|
| `camelCase` | JavaScript/TypeScript | `"apiVersion"` |
| `snake_case` | Python, beads | `"issue_type"` |
| `kebab-case` | Avoid | - |
| `UPPER_CASE` | Environment vars | `"DATABASE_URL"` |

**Rule:** Be consistent within a file. Match ecosystem convention.

---

## JSONL Format

### What is JSONL?

JSON Lines: one valid JSON object per line, newline-delimited.

```jsonl
{"id": "abc-123", "status": "open", "title": "First issue"}
{"id": "abc-124", "status": "closed", "title": "Second issue"}
{"id": "abc-125", "status": "open", "title": "Third issue"}
```

### When to Use

| Use JSONL | Use JSON |
|-----------|----------|
| Append-only data | Single config |
| Streaming ingestion | Nested data |
| Line-by-line processing | Small datasets |
| Beads issues | API responses |
| Large datasets | Human-edited |

### JSONL Rules

| Rule | Rationale |
|------|-----------|
| One object per line | Enables grep/head/tail |
| No trailing comma | Each line is complete |
| No array wrapper | Not `[{...}, {...}]` |
| Newline after last | Append-friendly |
| UTF-8, no BOM | Compatibility |

### Processing JSONL

```bash
# Count records
wc -l issues.jsonl

# Filter by field
jq -c 'select(.status == "open")' issues.jsonl

# Extract field
jq -r '.title' issues.jsonl

# Pretty-print one record
head -1 issues.jsonl | jq .

# Append new record
echo '{"id": "new", "status": "open"}' >> issues.jsonl

# Convert JSON array to JSONL
jq -c '.[]' array.json > data.jsonl

# Convert JSONL to JSON array
jq -s '.' data.jsonl > array.json
```

---

## Beads JSONL Schema

### Issue Record Schema

```json
{
  "id": "prefix-xxxx",
  "title": "Issue title",
  "status": "open",
  "priority": 2,
  "issue_type": "task",
  "owner": "user@example.com",
  "created_at": "2026-01-15T08:18:34.317984-05:00",
  "created_by": "User Name",
  "updated_at": "2026-01-15T08:42:39.253689-05:00",
  "closed_at": null,
  "close_reason": null,
  "dependencies": []
}
```

### Field Reference

| Field | Type | Required | Values |
|-------|------|----------|--------|
| `id` | string | Yes | `prefix-xxxx` |
| `title` | string | Yes | Brief description |
| `status` | string | Yes | `open`, `in_progress`, `closed` |
| `priority` | integer | Yes | 0-4 (0=critical) |
| `issue_type` | string | Yes | `task`, `bug`, `feature`, `epic` |
| `owner` | string | No | Email address |
| `created_at` | string | Yes | ISO 8601 |
| `updated_at` | string | Yes | ISO 8601 |
| `closed_at` | string | No | ISO 8601 or null |
| `dependencies` | array | No | Dependency objects |

### Dependency Object

```json
{
  "issue_id": "prefix-child",
  "depends_on_id": "prefix-parent",
  "type": "parent-child",
  "created_at": "2026-01-15T08:19:32.440350-05:00"
}
```

---

## Configuration Files

### tsconfig.json

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "NodeNext",
    "strict": true,
    "outDir": "./dist"
  },
  "include": ["src/**/*"],
  "exclude": ["node_modules"]
}
```

### package.json

```json
{
  "name": "package-name",
  "version": "1.0.0",
  "description": "Brief description",
  "main": "dist/index.js",
  "scripts": {
    "build": "tsc",
    "test": "jest",
    "lint": "eslint ."
  },
  "dependencies": {},
  "devDependencies": {}
}
```

### VS Code settings.json

```json
{
  "editor.formatOnSave": true,
  "editor.defaultFormatter": "esbenp.prettier-vscode",
  "files.insertFinalNewline": true,
  "files.trimTrailingWhitespace": true
}
```

---

## JSON Schema

### Defining Schemas

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://example.com/config.schema.json",
  "title": "Configuration",
  "type": "object",
  "required": ["name", "version"],
  "properties": {
    "name": {
      "type": "string",
      "description": "Project name",
      "minLength": 1
    },
    "version": {
      "type": "string",
      "pattern": "^\\d+\\.\\d+\\.\\d+$"
    },
    "enabled": {
      "type": "boolean",
      "default": true
    }
  },
  "additionalProperties": false
}
```

### Schema Validation

```bash
# Using ajv-cli
npx ajv validate -s schema.json -d config.json

# Using Python jsonschema
python -c "
import json
from jsonschema import validate
with open('schema.json') as s, open('config.json') as c:
    validate(json.load(c), json.load(s))
"
```

---

## Tooling

### Formatting

```bash
# jq - Format and validate
jq . config.json > formatted.json

# Prettier - Format with config
npx prettier --write '**/*.json'

# Python - Format
python -m json.tool config.json
```

### Validation

```bash
# jq - Check valid JSON
jq empty config.json && echo "Valid"

# Python - Check valid JSON
python -c "import json; json.load(open('config.json'))"

# Node - Check valid JSON
node -e "require('./config.json')"
```

### Editor Configuration

**.editorconfig:**
```ini
[*.json]
indent_style = space
indent_size = 2
insert_final_newline = true
charset = utf-8

[*.jsonl]
indent_style = space
indent_size = 0
insert_final_newline = true
```

**.prettierrc:**
```json
{
  "tabWidth": 2,
  "useTabs": false,
  "trailingComma": "none",
  "singleQuote": false
}
```

---

## Anti-Patterns

### Deeply Nested Objects

Nesting beyond 4 levels indicates missing abstraction or flattening opportunity.

```json
// BAD - 5 levels deep
{"config": {"server": {"auth": {"oauth": {"scopes": ["read"]}}}}}

// GOOD - flattened
{"auth_oauth_scopes": ["read"]}
```

### Inconsistent Key Naming

Mixing `camelCase` and `snake_case` within a single file breaks grep-ability and signals multiple authors without review.

### Missing Schema References

JSON config files without a `$schema` field cannot be validated automatically. Always include `$schema` when a schema exists.

### Magic Values Without Documentation

Undocumented numeric or string constants embedded in JSON (e.g., `"timeout": 86400`) should use descriptive keys or adjacent comments in the referencing code.

### Oversized Arrays or Objects

Arrays with >1000 elements or objects with >100 keys in a single file suggest the data belongs in JSONL or a database, not a monolithic JSON file.

### Duplicate Keys

JSON parsers silently drop earlier values when duplicate keys exist. This is always a bug.

---

## Code Quality Metrics

### Validation Thresholds

| Metric | Threshold | Severity |
|--------|-----------|----------|
| Schema coverage | 100% of config files have `$schema` | Warning |
| Nesting depth | ≤4 levels | Error above 4 |
| File size | ≤100KB per JSON file | Warning above 100KB |
| Key consistency | Single naming convention per file | Error if mixed |
| JSONL line validity | 100% lines parse | Error on any failure |
| Duplicate keys | 0 per file | Error |

### Grading Impact

| Violation | Grade Impact |
|-----------|-------------|
| Parse failure | Automatic C |
| Nesting >4 levels | Cap at B+ |
| Mixed key naming | Cap at A- |
| Missing schema ref | -0.5 grade step |
| File >100KB | -0.5 grade step |

---

## Prescan Patterns

Automated detection commands for CI or pre-commit validation.

### P01: Nesting Depth Check

| Field | Value |
|-------|-------|
| **Pattern** | Nesting depth exceeds 4 levels |
| **Detection** | `jq '[paths \| length] \| max' file.json` — fails if result >4 |
| **Severity** | Error |

### P02: Inconsistent Key Naming

| Field | Value |
|-------|-------|
| **Pattern** | Mixed camelCase and snake_case keys in same file |
| **Detection** | `jq '[paths \| .[] \| strings] \| unique \| map(test("_")) \| unique \| length' file.json` — fails if result >1 |
| **Severity** | Error |

### P03: Duplicate Keys

| Field | Value |
|-------|-------|
| **Pattern** | Same key appears twice in an object |
| **Detection** | `python -c "import json,sys; json.load(open(sys.argv[1]),object_pairs_hook=lambda p: (_ for k,v in p if sum(1 for k2,_ in p if k2==k)>1).__next__())" file.json` or use `jq --jsonargs` strict mode |
| **Severity** | Error |

### P04: Missing Schema Reference

| Field | Value |
|-------|-------|
| **Pattern** | Config file lacks `$schema` field |
| **Detection** | `jq 'has("$schema")' file.json` — fails if result is `false` |
| **Severity** | Warning |

### P05: Oversized Values

| Field | Value |
|-------|-------|
| **Pattern** | File exceeds 100KB or arrays exceed 1000 elements |
| **Detection** | `stat -f%z file.json` (macOS) or `stat -c%s file.json` (Linux) — fails if >102400; `jq '[.. \| arrays \| length] \| max' file.json` — fails if >1000 |
| **Severity** | Warning |

---

## Compliance Assessment

**Use letter grades + evidence, NOT numeric scores.**

### Assessment Categories

| Category | Evidence Required |
|----------|------------------|
| **Formatting** | jq validation, indentation, newlines |
| **Schema** | Validation errors, required fields |
| **Key Naming** | Consistency check |
| **JSONL Integrity** | Line count = record count |

### Grading Scale

| Grade | Criteria |
|-------|----------|
| A+ | All files validate, 2-space, UTF-8, schema valid |
| A | Valid JSON, consistent formatting |
| A- | Minor formatting inconsistencies |
| B | Valid but poorly formatted |
| C | Parse errors |

### Validation Commands

```bash
# Validate JSON
find . -name '*.json' -exec jq empty {} \; 2>&1 | grep -c "parse error"
# Should be 0

# Check indentation
jq . config.json | head -5

# JSONL: validate line count
wc -l data.jsonl
jq -c '.' data.jsonl | wc -l
# Should match

# JSONL: validate each line
while IFS= read -r line; do echo "$line" | jq empty; done < data.jsonl
```

### Example Assessment

```markdown
## JSON/JSONL Standards Compliance

| Category | Grade | Evidence |
|----------|-------|----------|
| Formatting | A+ | 18/18 validate, 2-space |
| Schema | A+ | 1247/1247 records pass |
| Key Naming | A | Consistent snake_case |
| JSONL | A+ | Line count matches |
| **OVERALL** | **A+** | **0 findings** |
```

---

## Additional Resources

- [JSON Spec](https://www.json.org/)
- [JSON Lines](https://jsonlines.org/)
- [JSON Schema](https://json-schema.org/)
- [jq Manual](https://stedolan.github.io/jq/manual/)

---

**Related:** Quick reference in Tier 1 `json.md`

### markdown-standards.md

# Markdown Standards Catalog - Vibe Canonical Reference

**Version:** 1.0.0
**Last Updated:** 2026-01-21
**Purpose:** Canonical Markdown standards for vibe skill validation

---

## Table of Contents

1. [AI-Agent Optimization](#ai-agent-optimization)
2. [Document Structure](#document-structure)
3. [Heading Conventions](#heading-conventions)
4. [Code Blocks](#code-blocks)
5. [Tables](#tables)
6. [Links](#links)
7. [Lists](#lists)
8. [Emphasis and Blockquotes](#emphasis-and-blockquotes)
9. [Validation](#validation)
10. [Compliance Assessment](#compliance-assessment)

---

## AI-Agent Optimization

### Principles

| Principle | Implementation | Why |
|-----------|----------------|-----|
| **Tables over prose** | Use tables for comparisons | Parallel parsing, scannable |
| **Explicit rules** | ALWAYS/NEVER, not "try to" | Removes ambiguity |
| **Decision trees** | If/then logic in lists | Executable reasoning |
| **Named patterns** | Anti-patterns with names | Recognizable error states |
| **Progressive disclosure** | Quick ref → details JIT | Context window efficiency |
| **Copy-paste ready** | Complete examples | Reduces inference errors |

---

## Document Structure

### SKILL.md Template

```markdown
# Skill Name

> **Triggers:** "phrase 1", "phrase 2", "phrase 3"

## Quick Reference

| Action | Command | Notes |
|--------|---------|-------|
| ... | ... | ... |

## When to Use

| Scenario | Action |
|----------|--------|
| Condition A | Do X |
| Condition B | Do Y |

## Workflow

1. Step one
2. Step two
3. Step three

## Common Errors

| Symptom | Cause | Fix |
|---------|-------|-----|
| Error message | Root cause | Solution |

## References

- [Reference 1](./references/detail1.md) - Load when needed
- [Reference 2](./references/detail2.md) - Load when needed
```

### Reference Doc Template

```markdown
# Reference: Topic Name

<!-- Load JIT when skill needs deep context -->

## Context

Brief overview of when this reference applies.

## Details

### Section 1

...

## Decision Tree

```text
Is X true?
├─ Yes → Do A
│   └─ Did A fail? → Try B
└─ No → Do C
```

## Anti-Patterns

| Name | Pattern | Why Bad | Instead |
|------|---------|---------|---------|
| ... | ... | ... | ... |
```

---

## Heading Conventions

### Hierarchy Rules

| Level | Use For | Example |
|-------|---------|---------|
| `#` | Document title (one per file) | `# Style Guide` |
| `##` | Major sections | `## Installation` |
| `###` | Subsections | `### macOS Setup` |
| `####` | Minor divisions (sparingly) | `#### Homebrew` |

**NEVER:**
- Skip heading levels (`#` → `###`)
- Use bold text as fake headings
- Start with `##` (missing `#` title)

### Heading Text

```markdown
# Good - Title Case for Title
## Good - Sentence case for sections
### Good - Sentence case continues

# Bad - all lowercase title
## Bad - ALL CAPS SECTION
### Bad - Using: Colons: Everywhere
```

---

## Code Blocks

### Language Hints (Required)

ALWAYS specify language for syntax highlighting:

````markdown
```python
def hello():
    print("world")
```
````

### Common Language Hints

| Language | Fence | Use For |
|----------|-------|---------|
| `bash` | ` ```bash ` | Shell commands |
| `python` | ` ```python ` | Python code |
| `go` | ` ```go ` | Go code |
| `typescript` | ` ```typescript ` | TypeScript |
| `yaml` | ` ```yaml ` | YAML config |
| `json` | ` ```json ` | JSON data |
| `text` | ` ```text ` | Plain text, diagrams |
| `diff` | ` ```diff ` | Code diffs |

### Command Output

```markdown
```bash
$ kubectl get pods
NAME         READY   STATUS    RESTARTS   AGE
my-pod       1/1     Running   0          5m
```
```

---

## Tables

### When to Use

| Situation | Use Table? | Alternative |
|-----------|------------|-------------|
| Comparing 3+ items | Yes | - |
| Key-value mappings | Yes | - |
| Command reference | Yes | - |
| Step-by-step | No | Numbered list |
| Narrative | No | Paragraphs |
| Two items only | No | Inline comparison |

### Table Formatting

```markdown
# Good - Aligned, readable
| Column A | Column B | Column C |
|----------|----------|----------|
| Value 1  | Value 2  | Value 3  |

# Bad - Misaligned
|Column A|Column B|Column C|
|-|-|-|
|Value 1|Value 2|Value 3|
```

### Table Cell Content

| Content Type | Formatting |
|--------------|------------|
| Code/commands | Backticks: `` `cmd` `` |
| Emphasis | Bold: `**required**` |
| Links | Inline: `[text](url)` |
| Long text | Under 50 chars |

---

## Links

### Internal Links

```markdown
# Good - Relative paths
[Guide](./other-doc.md)

# Good - Anchor links
[Code Blocks](#code-blocks)

# Bad - Absolute paths
[Guide](/Users/me/project/docs/guide.md)
```

### Reference Links

For repeated URLs:

```markdown
See the [official docs][k8s-docs] for more info.
The [Kubernetes documentation][k8s-docs] covers this.

[k8s-docs]: https://kubernetes.io/docs/
```

---

## Lists

### Unordered Lists

Use `-` consistently:

```markdown
# Good
- Item one
- Item two
  - Nested item

# Bad - Mixed markers
* Item one
+ Item two
- Item three
```

### Ordered Lists

Use `1.` for all items:

```markdown
# Good - All 1s
1. First step
1. Second step
1. Third step

# Acceptable - Explicit numbering
1. First step
2. Second step
3. Third step
```

### Task Lists

```markdown
- [ ] Incomplete task
- [x] Completed task
- [ ] Another incomplete
```

---

## Emphasis and Blockquotes

### Emphasis

| Purpose | Syntax | Example |
|---------|--------|---------|
| Important terms | `**bold**` | **required** |
| File names, commands | `` `backticks` `` | `config.yaml` |
| Titles, emphasis | `*italic*` | *optional* |
| Keyboard keys | `<kbd>` | <kbd>Ctrl</kbd>+<kbd>C</kbd> |

**NEVER use bold for:**
- Entire paragraphs
- Headings (use `#`)
- Code (use backticks)

### Callout Patterns

```markdown
> **Note:** Supplementary information.

> **Warning:** Something that could cause issues.

> **Important:** Critical information.

> **Tip:** Helpful suggestion.
```

---

## Validation

### markdownlint Configuration

```yaml
# .markdownlint.yml
default: true

MD013:
  line_length: 100
  code_blocks: false
  tables: false

MD033:
  allowed_elements:
    - kbd
    - br
    - details
    - summary

MD034: false

MD004:
  style: dash

MD003:
  style: atx
```

### Validation Commands

```bash
# Lint Markdown files
npx markdownlint '**/*.md' --ignore node_modules

# Check links
npx markdown-link-check README.md

# Format with Prettier
npx prettier --write '**/*.md'
```

---

## Compliance Assessment

**Use letter grades + evidence, NOT numeric scores.**

### Assessment Categories

| Category | Evidence Required |
|----------|------------------|
| **Structure** | Heading hierarchy, single H1 |
| **Formatting** | markdownlint violations, code fence hints |
| **Links** | Broken link count, relative paths |
| **AI Optimization** | Table usage, explicit rules |
| **Accessibility** | Alt text, semantic markup |

### Grading Scale

| Grade | Criteria |
|-------|----------|
| A+ | 0 errors, single H1, 100% code hints, 0 broken links |
| A | <5 warnings, good structure |
| A- | <15 warnings, mostly correct |
| B | <30 warnings |
| C | Significant issues |

### Validation Commands

```bash
# Lint Markdown
npx markdownlint '**/*.md' --ignore node_modules

# Check heading hierarchy
grep -r "^# " docs/*.md | wc -l
ls docs/*.md | wc -l
# Should match (1 H1 per file)

# Code blocks without language
grep -rP '```\s*$' docs/ | wc -l
# Should be 0

# Check links
npx markdown-link-check docs/**/*.md
```

### Example Assessment

```markdown
## Markdown Standards Compliance

| Category | Grade | Evidence |
|----------|-------|----------|
| Structure | A+ | 47/47 single H1, 0 skipped |
| Formatting | A- | 18 warnings (MD013) |
| Links | A | 0 broken, 93% relative |
| AI Optimization | A | 85 tables, 23 decision trees |
| **OVERALL** | **A** | **18 MEDIUM findings** |
```

---

## Additional Resources

- [CommonMark Spec](https://spec.commonmark.org/)
- [markdownlint Rules](https://github.com/DavidAnson/markdownlint)
- [GitHub Flavored Markdown](https://github.github.com/gfm/)

---

**Related:** Quick reference in Tier 1 `markdown.md`

### patterns.md

# Vibe Pattern Reference

Comprehensive pattern catalog for Talos validation.

## Pattern Categories

| Category | Prefix | Phase | Description |
|----------|--------|-------|-------------|
| Prescan | P1-P10 | Static | Fast, no LLM required |
| Quality | QUAL-xxx | Semantic | Code smells, patterns |
| Security | SEC-xxx | Semantic | OWASP, injection, auth |
| Architecture | ARCH-xxx | Semantic | Boundaries, coupling |
| Accessibility | A11Y-xxx | Semantic | WCAG, keyboard |
| Complexity | CMPLX-xxx | Both | Cyclomatic, cognitive |
| Semantic | SEM-xxx | Semantic | Names, docstrings |
| Performance | PERF-xxx | Semantic | N+1, leaks |
| Slop | SLOP-xxx | Semantic | AI artifacts |

---

## Prescan Patterns (Static Detection)

Fast static analysis - no LLM required.

**Supported Languages:** Python, Go, Bash, TypeScript, JavaScript

### P1: Phantom Modifications (CRITICAL)

**What**: Committed lines that don't exist in current file.

**Why Critical**: Indicates broken git workflow - changes were committed but then removed or lost.

**Detection**: Compare `git show HEAD -- <file>` with actual file content.

**Fix**: Re-commit or investigate git history.

---

### P2: Hardcoded Secrets (CRITICAL)

**What**: API keys, passwords, tokens in source code.

**Why Critical**: Credential exposure leads to immediate compromise.

**Detection**:
- gitleaks scan
- Regex for common patterns (AWS keys, JWT, password=)

**Patterns**:
```regex
(password|secret|api_key|token)\s*[=:]\s*["'][^"']{8,}["']
AKIA[0-9A-Z]{16}
eyJ[a-zA-Z0-9_-]+\.eyJ[a-zA-Z0-9_-]+
```

**Fix**: Use environment variables or secrets manager.

---

### P3: SQL Injection Patterns (CRITICAL)

**What**: String concatenation in SQL queries.

**Why Critical**: Direct path to data breach.

**Detection**:
```regex
(execute|query)\s*\(\s*f?["'].*\{.*\}
cursor\.(execute|query)\s*\([^)]*%
```

**Fix**: Use parameterized queries.

---

### P4: TODO/FIXME/Commented Code (HIGH)

**What**: TODO markers, FIXME, commented-out code blocks.

**Why High**: Incomplete work or tech debt markers.

**Detection**:
```bash
grep -E "TODO|FIXME|XXX|HACK|BUG"
grep -E "^\s*#\s*(def |class |if |for |while )"  # Commented code
```

**Fix**: Complete or remove with explanation.

---

### P5: Cyclomatic Complexity (HIGH)

**What**: Functions with CC > 15.

**Why High**: Too complex to maintain safely.

**Detection by Language**:

| Language | Tool | Command |
|----------|------|---------|
| Python | radon | `radon cc <file> -s -n E` |
| Go | gocyclo | `gocyclo -over 15 <file>` |
| TypeScript | escomplex | `escomplex <file>` |

**Thresholds**:
- CC > 10: Warning
- CC > 15: Flag as complex
- CC > 20: Critical

**Fix**: Extract functions, simplify logic.

---

### P6: Long Functions (HIGH)

**What**: Functions exceeding 50 lines.

**Why High**: Long functions are hard to test and maintain.

**Detection**: AST parsing, line counting.

**Thresholds**:
- Lines > 30: Warning
- Lines > 50: Flag
- Lines > 100: Critical

**Fix**: Extract helper functions.

---

### P7: Cargo Cult Error Handling (HIGH)

**What**: Empty except blocks, pass-only handlers, bare except.

**Why High**: Swallowed errors hide bugs.

**Detection by Language**:

| Language | Pattern |
|----------|---------|
| Python | `except: pass`, `except Exception: pass` |
| Go | `if err != nil { }` (empty block) |
| Bash | shellcheck SC2181 |

**Fix**: Handle or propagate errors explicitly.

---

### P8: Unused Imports/Functions (MEDIUM)

**What**: Imported modules or defined functions never used.

**Why Medium**: Dead code clutters codebase.

**Detection**: AST analysis, import tracking.

**Fix**: Remove unused code.

---

### P9: Docstring Mismatches (MEDIUM)

**What**: Docstrings claiming behavior not implemented.

**Why Medium**: False security from lying documentation.

**Detection**: Match docstring claims vs implementation:
- "validates" but no raise/ValueError
- "encrypts" but no crypto imports
- "authenticates" but no token handling

**Fix**: Update docs or implement claimed behavior.

---

### P10: Missing Error Handling (MEDIUM)

**What**: Operations that can fail without error handling.

**Why Medium**: Silent failures cause hard-to-debug issues.

**Detection**:
- File operations without try/except
- Network calls without timeout/retry
- Parsing without validation

**Fix**: Add appropriate error handling.

---

## Semantic Patterns (LLM-Powered)

Deep analysis requiring semantic understanding.

### Quality (QUAL-xxx)

| Code | Pattern | Severity |
|------|---------|----------|
| QUAL-001 | Dead code paths | MEDIUM |
| QUAL-002 | Inconsistent naming | MEDIUM |
| QUAL-003 | Magic numbers/strings | MEDIUM |
| QUAL-004 | Missing tests for complex code | HIGH |
| QUAL-005 | Copy-paste with variations | HIGH |
| QUAL-006 | Feature envy (method uses another class more) | MEDIUM |
| QUAL-007 | Primitive obsession | LOW |
| QUAL-008 | Long parameter lists | MEDIUM |

---

### Security (SEC-xxx)

| Code | Pattern | Severity |
|------|---------|----------|
| SEC-001 | Injection (SQL, command, XSS, template) | CRITICAL |
| SEC-002 | Authentication bypass | CRITICAL |
| SEC-003 | Authorization missing/weak | CRITICAL |
| SEC-004 | Cryptographic weakness | HIGH |
| SEC-005 | Sensitive data exposure | HIGH |
| SEC-006 | Security theater (looks secure, isn't) | HIGH |
| SEC-007 | Insecure deserialization | HIGH |
| SEC-008 | SSRF/path traversal | HIGH |
| SEC-009 | Race conditions | MEDIUM |
| SEC-010 | Debug mode in production | MEDIUM |

---

### Architecture (ARCH-xxx)

| Code | Pattern | Severity |
|------|---------|----------|
| ARCH-001 | Layer boundary violation | HIGH |
| ARCH-002 | Circular dependency | HIGH |
| ARCH-003 | God class/function | HIGH |
| ARCH-004 | Missing abstraction | MEDIUM |
| ARCH-005 | Inappropriate coupling | MEDIUM |
| ARCH-006 | Scalability concern | MEDIUM |
| ARCH-007 | Single point of failure | HIGH |
| ARCH-008 | Hardcoded configuration | MEDIUM |
| ARCH-009 | Missing retry/circuit breaker | MEDIUM |
| ARCH-010 | Synchronous where async needed | MEDIUM |

---

### Accessibility (A11Y-xxx)

| Code | Pattern | Severity |
|------|---------|----------|
| A11Y-001 | Missing ARIA labels | HIGH |
| A11Y-002 | Keyboard navigation broken | CRITICAL |
| A11Y-003 | Color contrast insufficient | HIGH |
| A11Y-004 | Missing alt text | HIGH |
| A11Y-005 | Focus management issues | HIGH |
| A11Y-006 | Missing skip links | MEDIUM |
| A11Y-007 | Form labels missing | HIGH |
| A11Y-008 | Dynamic content not announced | MEDIUM |
| A11Y-009 | Touch target too small | MEDIUM |
| A11Y-010 | Motion without reduced-motion support | LOW |

---

### Complexity (CMPLX-xxx)

| Code | Pattern | Severity | Threshold |
|------|---------|----------|-----------|
| CMPLX-001 | Cyclomatic complexity | HIGH | CC > 10 |
| CMPLX-002 | Cognitive complexity | HIGH | > 15 |
| CMPLX-003 | Nesting depth | MEDIUM | > 4 |
| CMPLX-004 | Parameter count | MEDIUM | > 5 |
| CMPLX-005 | File too long | MEDIUM | > 500 lines |
| CMPLX-006 | Class too large | MEDIUM | > 20 methods |
| CMPLX-007 | Inheritance depth | LOW | > 3 |
| CMPLX-008 | Fan-out too high | MEDIUM | > 10 dependencies |

---

### Semantic (SEM-xxx)

| Code | Pattern | Severity |
|------|---------|----------|
| SEM-001 | Docstring lies | HIGH |
| SEM-002 | Misleading function name | HIGH |
| SEM-003 | Misleading variable name | MEDIUM |
| SEM-004 | Comment rot | MEDIUM |
| SEM-005 | API contract violation | HIGH |
| SEM-006 | Type annotation mismatch | MEDIUM |
| SEM-007 | Inconsistent return types | MEDIUM |
| SEM-008 | Side effects in getter | HIGH |

---

### Performance (PERF-xxx)

| Code | Pattern | Severity |
|------|---------|----------|
| PERF-001 | N+1 query | HIGH |
| PERF-002 | Unbounded loop/recursion | CRITICAL |
| PERF-003 | Missing pagination | HIGH |
| PERF-004 | Resource leak | HIGH |
| PERF-005 | Blocking in async context | HIGH |
| PERF-006 | Inefficient algorithm | MEDIUM |
| PERF-007 | Repeated computation | MEDIUM |
| PERF-008 | Missing caching | LOW |
| PERF-009 | Large object in memory | MEDIUM |
| PERF-010 | Excessive logging | LOW |

---

### Slop (SLOP-xxx)

| Code | Pattern | Severity |
|------|---------|----------|
| SLOP-001 | Hallucinated imports/APIs | CRITICAL |
| SLOP-002 | Cargo cult patterns | HIGH |
| SLOP-003 | Excessive boilerplate | MEDIUM |
| SLOP-004 | AI conversation artifacts | HIGH |
| SLOP-005 | Over-engineering | MEDIUM |
| SLOP-006 | Unnecessary abstractions | MEDIUM |
| SLOP-007 | Copy-paste from tutorials | MEDIUM |
| SLOP-008 | Sycophantic comments | LOW |
| SLOP-009 | Redundant type annotations | LOW |
| SLOP-010 | Verbose where concise works | LOW |

---

## Severity Mapping

| Level | Definition | Action | Exit Code |
|-------|------------|--------|-----------|
| **CRITICAL** | Security vuln, data loss, broken build | Block merge | 2 |
| **HIGH** | Significant quality/security gap | Fix before merge | 3 |
| **MEDIUM** | Technical debt, minor issues | Follow-up issue | 0 |
| **LOW** | Nitpicks, style preferences | Optional | 0 |

---

## Tool Requirements

| Tool | Languages | Install |
|------|-----------|---------|
| radon | Python | `pip install radon` |
| gocyclo | Go | `go install github.com/fzipp/gocyclo/cmd/gocyclo@latest` |
| shellcheck | Bash | `brew install shellcheck` |
| gitleaks | All | `brew install gitleaks` |
| eslint | JS/TS | `npm install eslint` |

### python-standards.md

# Python Standards Catalog - Vibe Canonical Reference

**Version:** 1.0.0
**Last Updated:** 2026-01-21
**Purpose:** Canonical Python standards for vibe skill validation

---

## Table of Contents

1. [Project Structure](#project-structure)
2. [Package Management](#package-management)
3. [Code Formatting](#code-formatting)
4. [Reducing Complexity](#reducing-complexity)
5. [Type Hints](#type-hints)
6. [Docstrings](#docstrings)
7. [Error Handling](#error-handling)
8. [Logging](#logging)
9. [Testing](#testing)
10. [CLI Script Template](#cli-script-template)
11. [Code Quality Metrics](#code-quality-metrics)
12. [Security Practices](#security-practices)
13. [Anti-Patterns Avoided](#anti-patterns-avoided)
14. [Compliance Assessment](#compliance-assessment)

---

## Project Structure

### Standard Layout

```text
project/
├── pyproject.toml           # Project metadata and dependencies
├── uv.lock                  # Lock file (commit this!)
├── src/
│   └── mypackage/           # Source code
│       ├── __init__.py
│       ├── core.py
│       └── utils.py
├── scripts/                 # CLI tools
│   └── my_script.py
├── tests/                   # Test suite
│   ├── __init__.py
│   ├── conftest.py          # Pytest fixtures
│   ├── test_core.py
│   └── e2e/                 # End-to-end tests
│       ├── conftest.py      # Testcontainers fixtures
│       └── test_integration.py
└── docs/                    # Documentation
```

**Key Principles:**
- Use `src/` layout for packages (prevents import issues)
- CLI scripts are standalone files in `scripts/`
- Tests mirror source structure
- Always commit `uv.lock` for reproducibility

---

## Package Management

### uv - Project Dependencies

Use `uv` for all project-level Python dependencies. It's 10-100x faster than pip and creates deterministic builds.

```bash
# Initialize a new project
uv init my-project
cd my-project

# Add dependencies
uv add requests pyyaml        # Runtime deps
uv add --dev pytest ruff      # Dev deps

# Install from existing pyproject.toml
uv sync                       # Creates/updates uv.lock

# Run a script with project deps
uv run python my_script.py
```

### pipx - Global CLI Tools

Use `pipx` for Python CLI tools you want available everywhere.

```bash
# Install CLI tools globally
pipx install ruff             # Linter/formatter
pipx install radon            # Complexity analysis
pipx install xenon            # Complexity enforcement
pipx install pre-commit       # Git hooks

# Upgrade all
pipx upgrade-all

# Run without installing
pipx run cowsay "hello"
```

### When to Use What

| Need | Tool | Command |
|------|------|---------|
| Install project deps | uv | `uv sync` |
| Add library to project | uv | `uv add requests` |
| Install CLI globally | pipx | `pipx install ruff` |
| Install system tool | brew | `brew install shellcheck` |
| Quick script run | uv | `uv run script.py` |

### What NOT to Do

```python
# DON'T use pip globally
pip install requests          # Pollutes system Python
sudo pip install anything     # Even worse

# DON'T mix package managers
pip install requests          # Now you have pip AND uv deps
uv add pyyaml                 # Conflicts likely

# DON'T commit venv/
git add .venv/                # Use .gitignore
```

---

## Code Formatting

### ruff Configuration

**Full recommended configuration:**

```toml
# pyproject.toml
[tool.ruff]
line-length = 100
target-version = "py312"
exclude = [
    ".git",
    ".venv",
    "__pycache__",
    "build",
    "dist",
]

[tool.ruff.lint]
select = [
    "E",   # pycodestyle errors
    "W",   # pycodestyle warnings
    "F",   # pyflakes
    "I",   # isort
    "N",   # pep8-naming
    "UP",  # pyupgrade
    "B",   # flake8-bugbear
    "C4",  # flake8-comprehensions
    "SIM", # flake8-simplify
    "S",   # flake8-bandit (security)
    "A",   # flake8-builtins
    "PT",  # flake8-pytest-style
]
ignore = [
    "E501",  # line-too-long (handled by formatter)
    "S101",  # assert (OK in tests)
]

[tool.ruff.lint.per-file-ignores]
"tests/**/*.py" = ["S101"]  # Allow assert in tests

[tool.ruff.format]
quote-style = "double"
indent-style = "space"
```

### Usage

```bash
# Check linting
ruff check src/

# Auto-fix issues
ruff check --fix src/

# Format code
ruff format src/

# Check formatting only
ruff format --check src/
```

---

## Reducing Complexity

**Target:** Maximum cyclomatic complexity of 10 (Grade B) per function

### Why Complexity Matters

- CC = number of independent paths through code
- CC > 10 means exponentially more test cases for coverage
- High complexity correlates with defect density
- Humans (and LLMs) struggle with deeply nested logic

### Pattern 1: Dispatch Pattern (Handler Registry)

**When to use:** Functions with if/elif chains that dispatch based on mode or type.

```python
# Bad - if/elif chain (CC=18+)
def main():
    if args.patch:
        # 90 lines of patch logic
    elif args.read:
        # 20 lines of read logic
    else:
        # 100 lines of write logic

# Good - Dispatch pattern (CC=6)
def _handle_patch_mode(args: Args, client: Client) -> None:
    """Handle --patch mode."""
    # Focused patch logic

def _handle_read_mode(args: Args, client: Client) -> None:
    """Handle --read mode."""
    # Focused read logic

def main() -> int:
    args = parse_args()
    client = build_client()

    handlers = {
        "patch": _handle_patch_mode,
        "read": _handle_read_mode,
        "write": _handle_write_mode,
    }

    handler = handlers.get(args.mode, _handle_write_mode)
    handler(args, client)
    return 0
```

### Pattern 2: Early Returns (Guard Clauses)

```python
# Bad - Deep nesting (CC=8)
def validate_document(doc: Document) -> bool:
    if doc:
        if doc.content:
            if len(doc.content) > 0:
                if doc.tenant:
                    return True
    return False

# Good - Guard clauses (CC=4)
def validate_document(doc: Document | None) -> bool:
    if not doc:
        return False
    if not doc.content:
        return False
    if len(doc.content) == 0:
        return False
    if not doc.tenant:
        return False
    return True
```

### Pattern 3: Lookup Tables

```python
# Bad - Each 'or' adds +1 CC
def normalize_field(key: str, value: str) -> str:
    if key == "tls.crt" or key == "tls.key" or key == "ca":
        return normalize_cert_field(value)
    elif key == "config.json":
        return normalize_pull_secret_json(value)
    else:
        return value

# Good - O(1) lookup
NORMALIZERS: dict[str, Callable[[str], str]] = {
    "tls.crt": normalize_cert_field,
    "tls.key": normalize_cert_field,
    "ca": normalize_cert_field,
    "config.json": normalize_pull_secret_json,
}

def normalize_field(key: str, value: str) -> str:
    normalizer = NORMALIZERS.get(key)
    return normalizer(value) if normalizer else value
```

### Pattern 4: Strategy Pattern (Class-Based)

```python
# Bad - Type checking with isinstance
def process(item: Item) -> Result:
    if isinstance(item, TypeA):
        # TypeA logic
    elif isinstance(item, TypeB):
        # TypeB logic
    elif isinstance(item, TypeC):
        # TypeC logic
    # ... many more types

# Good - Strategy pattern
from abc import ABC, abstractmethod

class ItemProcessor(ABC):
    @abstractmethod
    def process(self, item: Item) -> Result:
        pass

class TypeAProcessor(ItemProcessor):
    def process(self, item: Item) -> Result:
        # TypeA logic

class TypeBProcessor(ItemProcessor):
    def process(self, item: Item) -> Result:
        # TypeB logic

# Registry
PROCESSORS: dict[type, ItemProcessor] = {
    TypeA: TypeAProcessor(),
    TypeB: TypeBProcessor(),
}

def process(item: Item) -> Result:
    processor = PROCESSORS.get(type(item))
    if not processor:
        raise ValueError(f"No processor for {type(item)}")
    return processor.process(item)
```

### Helper Naming Convention

| Prefix | Meaning | Example |
|--------|---------|---------|
| `_handle_` | Mode/dispatch handler | `_handle_patch_mode()` |
| `_process_` | Processing helper | `_process_secret()` |
| `_validate_` | Validation helper | `_validate_cert()` |
| `_setup_` | Initialization helper | `_setup_mount_point()` |
| `_normalize_` | Data normalization | `_normalize_cert_field()` |
| `_build_` | Construction | `_build_audit_metadata()` |

### Measuring Complexity

```bash
# Check specific file
radon cc scripts/my_script.py -s -a

# Fail if any function exceeds Grade B (CC > 10)
xenon scripts/ --max-absolute B

# Show only Grade C or worse
radon cc scripts/ -s -n C
```

---

## Type Hints

### Modern Syntax (Python 3.12+)

```python
from __future__ import annotations
from typing import Any, Callable, TypeVar

# Basic types - use lowercase
items: list[str] = []
mapping: dict[str, int] = {}
coords: tuple[int, int, int] = (0, 0, 0)

# Union with pipe operator
value: str | int = "hello"
optional: str | None = None

# Function signatures
def process(
    items: list[str],
    config: dict[str, Any] | None = None,
    callback: Callable[[str], bool] | None = None,
) -> list[str]:
    """Process items with optional config."""
    ...

# Generics
T = TypeVar("T")

def first(items: list[T]) -> T | None:
    return items[0] if items else None
```

### Type Hint Anti-Patterns

| Anti-Pattern | Problem | Better |
|--------------|---------|--------|
| `Any` everywhere | Defeats type checking | Use generics or specific types |
| `# type: ignore` without comment | Hides real issues | Add explanation |
| Old syntax `List[str]` | Deprecated | Use `list[str]` |
| Missing return type | Incomplete signature | Always add return type |

---

## Docstrings

### Google Style (Required)

```python
def verify_secret_after_write(
    client: hvac.Client,
    mount_point: str,
    name: str,
    expected_payload: dict[str, Any],
) -> bool:
    """Verify secret was written correctly.

    Args:
        client: Vault client connection.
        mount_point: KV v2 mount point path.
        name: Secret name/key.
        expected_payload: Expected secret data to verify against.

    Returns:
        True if verification passed, False if any check failed.

    Raises:
        hvac.exceptions.InvalidPath: If secret path is invalid.
        ConnectionError: If Vault connection fails.

    Example:
        >>> client = hvac.Client(url="http://localhost:8200")
        >>> verify_secret_after_write(client, "secret", "mykey", {"foo": "bar"})
        True
    """
    pass
```

### When to Include Each Section

| Section | When to Include |
|---------|-----------------|
| **Args** | Always if function has parameters |
| **Returns** | Always if function returns non-None |
| **Raises** | If function can raise exceptions |
| **Example** | For complex or non-obvious usage |
| **Note** | For important caveats or warnings |

---

## Error Handling

### Good Patterns

```python
# Good - Specific exception, logged
try:
    cert_info = validate_certificate(payload["tls.crt"])
except subprocess.CalledProcessError as exc:
    logging.warning("Certificate validation failed: %s", exc)

# Good - Multiple specific types for format detection
try:
    decoded = base64.b64decode(data)
except (UnicodeDecodeError, base64.binascii.Error, ValueError) as exc:
    logging.debug("Not base64, assuming PEM format: %s", exc)
    decoded = data

# Good - Re-raise with context
try:
    result = subprocess.run(cmd, check=True, capture_output=True)
except subprocess.CalledProcessError as exc:
    raise RuntimeError(f"Command failed: {cmd}") from exc

# Good - Custom exception with context
class ConfigError(Exception):
    """Configuration validation error."""
    def __init__(self, key: str, message: str):
        self.key = key
        super().__init__(f"Config '{key}': {message}")
```

### Bad Patterns

```python
# Bad - Bare exception, swallowed
try:
    validate_something()
except Exception:
    pass  # Silent failure!

# Bad - Catching Exception without re-raising
try:
    process_data()
except Exception as e:
    logging.error("Error: %s", e)
    return None  # Hides the problem

# Bad - Too broad, catches KeyboardInterrupt
try:
    long_running_task()
except:  # noqa: E722
    pass
```

### Exception Hierarchy for Custom Errors

```python
class MyAppError(Exception):
    """Base exception for application errors."""

class ValidationError(MyAppError):
    """Input validation failed."""

class ConnectionError(MyAppError):
    """External service connection failed."""

class ConfigError(MyAppError):
    """Configuration error."""
```

---

## Logging

### Standard Setup

```python
import logging

# Basic setup for scripts
logging.basicConfig(
    format="%(asctime)s %(levelname)s %(message)s",
    level=logging.INFO,
)

# Module logger for libraries
log = logging.getLogger(__name__)
```

### Log Levels

| Level | When to Use |
|-------|-------------|
| `DEBUG` | Detailed diagnostic (development only) |
| `INFO` | Key events, progress |
| `WARNING` | Recoverable issues |
| `ERROR` | Operation failed |
| `CRITICAL` | Application cannot continue |

### Good Patterns

```python
# Good - Use % formatting (lazy evaluation)
logging.info("Processing secret: %s", secret_name)
logging.warning("Retry %d of %d: %s", attempt, max_retries, error)

# Good - Include context
logging.info("Prepared %s: %s", secret_name, preview)
logging.warning("Security policy check failed for %s: %s", key, exc)

# Good - Structured for parsing
logging.info("event=secret_prepared name=%s preview=%s", secret_name, preview)
```

### Bad Patterns

```python
# Bad - f-string (evaluated even if level disabled)
logging.info(f"Processing {expensive_to_compute()}")

# Bad - No context
logging.info("Processing...")
logging.error(str(e))

# Bad - print() instead of logging
print("DEBUG: value is", value)
```

---

## Testing

### Pytest Structure

```text
tests/
├── conftest.py           # Shared fixtures
├── test_core.py          # Unit tests for core module
├── test_utils.py         # Unit tests for utils
└── e2e/                  # End-to-end tests
    ├── conftest.py       # Testcontainers fixtures
    └── test_integration.py
```

### Configuration

```toml
# pyproject.toml
[tool.pytest.ini_options]
testpaths = ["tests"]
markers = [
    "e2e: marks tests as end-to-end (require Docker)",
    "slow: marks tests as slow",
]
addopts = "-v --tb=short"

[tool.coverage.run]
source = ["src"]
branch = true

[tool.coverage.report]
exclude_lines = [
    "pragma: no cover",
    "if TYPE_CHECKING:",
    "raise NotImplementedError",
]
```

### Testcontainers for E2E Tests

Use testcontainers for tests that need real infrastructure.

```python
# tests/e2e/conftest.py
import pytest
from testcontainers.postgres import PostgresContainer

@pytest.fixture(scope="session")
def postgres_container():
    """Spin up PostgreSQL for E2E tests."""
    with PostgresContainer("postgres:16") as postgres:
        yield postgres
    # Container automatically cleaned up

@pytest.fixture
def db_connection(postgres_container):
    """Get connection to test database."""
    import psycopg
    conn_str = postgres_container.get_connection_url()
    with psycopg.connect(conn_str) as conn:
        yield conn
```

### Test Patterns

```python
# Table-driven tests
import pytest

@pytest.mark.parametrize("input,expected", [
    ("valid@example.com", True),
    ("invalid", False),
    ("", False),
    ("@nodomain", False),
])
def test_validate_email(input: str, expected: bool):
    assert validate_email(input) == expected

# Fixtures for setup/teardown
@pytest.fixture
def temp_config(tmp_path):
    """Create temporary config file."""
    config_file = tmp_path / "config.yaml"
    config_file.write_text("key: value")
    return config_file

def test_load_config(temp_config):
    config = load_config(temp_config)
    assert config["key"] == "value"

# Mock external services
from unittest.mock import patch, MagicMock

def test_api_call():
    with patch("mymodule.requests.get") as mock_get:
        mock_get.return_value = MagicMock(status_code=200, json=lambda: {"data": "test"})
        result = my_api_function()
        assert result == {"data": "test"}
```

### Running Tests

```bash
# Run all tests
pytest

# Run with coverage
pytest --cov=src --cov-report=term-missing

# Run only E2E tests
pytest -m e2e

# Run excluding slow tests
pytest -m "not slow"
```

---

## CLI Script Template

```python
#!/usr/bin/env python3
"""One-line description of what this script does.

Usage:
    python3 script_name.py --config config.yaml --apply

Exit Codes:
    0 - Success
    1 - Argument/configuration error
    2 - Runtime error
"""

from __future__ import annotations

import argparse
import logging
import sys
from pathlib import Path
from typing import Any

logging.basicConfig(
    format="%(asctime)s %(levelname)s %(message)s",
    level=logging.INFO,
)


def die(message: str) -> None:
    """Print error message and exit with code 1."""
    logging.error(message)
    sys.exit(1)


def parse_args() -> argparse.Namespace:
    """Parse command-line arguments."""
    parser = argparse.ArgumentParser(
        description=__doc__,
        formatter_class=argparse.RawDescriptionHelpFormatter,
    )
    parser.add_argument(
        "--config",
        default="config.yaml",
        type=Path,
        help="Path to config file (default: config.yaml)",
    )
    parser.add_argument(
        "--apply",
        action="store_true",
        help="Apply changes (default: dry-run)",
    )
    parser.add_argument(
        "-v", "--verbose",
        action="store_true",
        help="Enable debug logging",
    )
    return parser.parse_args()


def main() -> int:
    """Main entry point."""
    args = parse_args()

    if args.verbose:
        logging.getLogger().setLevel(logging.DEBUG)

    if not args.apply:
        logging.info("Dry-run mode (use --apply to make changes)")

    # Validate config exists
    if not args.config.exists():
        die(f"Config file not found: {args.config}")

    # Main logic here
    try:
        # ... implementation
        logging.info("Processing complete")
        return 0
    except Exception as exc:
        logging.error("Failed: %s", exc)
        return 2


if __name__ == "__main__":
    sys.exit(main())
```

---

## Code Quality Metrics

> See `common-standards.md` for universal coverage targets and testing principles.

### Complexity Thresholds

| Grade | CC Range | Action |
|-------|----------|--------|
| A | 1-5 | Ideal - simple, low risk |
| B | 6-10 | Acceptable - moderate complexity |
| C | 11-20 | Refactor when touching |
| D | 21-30 | Must refactor before merge |
| F | 31+ | Block merge |

### Validation Commands

```bash
# Code quality + style
ruff check src/ --statistics
# Output: "10 errors, 5 warnings" → Count these

# Complexity analysis
radon cc src/ -s -a
# Output includes per-function CC and average → Report both

# Enforce complexity limit
xenon src/ --max-absolute B
# Fails if any function exceeds CC=10

# Test coverage
pytest --cov=src --cov-report=term-missing
# Output: "87% line, 71% branch" → Report both

# Docstring coverage
interrogate src/
# Output: "85% (45/53 functions)" → Report fraction + %
```

---

## Security Practices

### eval/exec Avoidance

Never use `eval()` or `exec()` on user-controlled input:

```python
# DANGEROUS - Remote code execution
user_expr = request.args["expr"]
result = eval(user_expr)  # Attacker sends: __import__('os').system('rm -rf /')

# SAFE - Use ast.literal_eval for data literals
import ast
result = ast.literal_eval(user_input)  # Only parses strings, numbers, tuples, lists, dicts

# SAFE - Use a mapping for dynamic dispatch
OPERATIONS = {"add": operator.add, "mul": operator.mul}
func = OPERATIONS.get(user_input)
if func:
    result = func(a, b)
```

**Validation:** Prescan pattern P16 detects `eval(` and `exec(` calls

### Pickle Safety

Never unpickle untrusted data — `pickle.loads()` executes arbitrary code:

```python
# DANGEROUS - Arbitrary code execution on load
import pickle
data = pickle.loads(untrusted_bytes)  # Attacker crafts payload to run code

# SAFE - Use JSON for data interchange
import json
data = json.loads(untrusted_bytes)

# SAFE - Use msgpack for binary efficiency
import msgpack
data = msgpack.unpackb(untrusted_bytes, raw=False)
```

If pickle is unavoidable (e.g., ML model loading), load only from trusted, integrity-verified sources.

### YAML Deserialization

`yaml.load()` with the default loader executes arbitrary Python objects:

```python
# DANGEROUS - Arbitrary code execution
import yaml
data = yaml.load(untrusted_string)  # Can execute __reduce__, !!python/object, etc.

# SAFE - Use safe_load (only basic YAML types)
data = yaml.safe_load(untrusted_string)

# SAFE - Explicit SafeLoader
data = yaml.load(untrusted_string, Loader=yaml.SafeLoader)
```

**Rule:** Always use `yaml.safe_load()` or `yaml.safe_load_all()`. Never use `yaml.load()` without `Loader=yaml.SafeLoader`.

### SQL Injection Prevention

Always use parameterized queries:

```python
# DANGEROUS - SQL injection
cursor.execute(f"SELECT * FROM users WHERE name = '{user_input}'")

# SAFE - Parameterized query
cursor.execute("SELECT * FROM users WHERE name = %s", (user_input,))

# SAFE - SQLAlchemy ORM
user = session.query(User).filter(User.name == user_input).first()

# SAFE - SQLAlchemy text with bind params
from sqlalchemy import text
stmt = text("SELECT * FROM users WHERE name = :name")
result = conn.execute(stmt, {"name": user_input})
```

### SSRF Prevention

Validate URLs before making outbound requests:

```python
# DANGEROUS - Server-Side Request Forgery
url = request.args["url"]
resp = requests.get(url)  # Attacker sends: http://169.254.169.254/metadata

# SAFE - URL allowlist validation
from urllib.parse import urlparse

ALLOWED_HOSTS = {"api.example.com", "cdn.example.com"}

def validate_url(url: str) -> bool:
    parsed = urlparse(url)
    if parsed.scheme not in ("http", "https"):
        return False
    if parsed.hostname not in ALLOWED_HOSTS:
        return False
    return True

if validate_url(url):
    resp = requests.get(url, timeout=10)
```

### Path Traversal Prevention

User-controlled path components can escape intended directories:

```python
# DANGEROUS - Path traversal
user_file = request.args["filename"]
path = os.path.join("/data/uploads", user_file)  # "../../../etc/passwd" escapes!
content = open(path).read()

# SAFE - Resolve and check prefix
from pathlib import Path

UPLOAD_DIR = Path("/data/uploads").resolve()

def safe_read(filename: str) -> str:
    target = (UPLOAD_DIR / filename).resolve()
    if not target.is_relative_to(UPLOAD_DIR):
        raise ValueError(f"Path traversal blocked: {filename!r}")
    return target.read_text()

# SAFE - Strip directory components entirely
from pathlib import PurePosixPath

def sanitize_filename(filename: str) -> str:
    """Extract only the final filename component."""
    return PurePosixPath(filename).name
```

**Key pitfalls:**
- `os.path.join("/base", "/etc/passwd")` returns `/etc/passwd` (absolute path overrides base)
- Symlinks can bypass prefix checks — use `.resolve()` before comparison
- Always use `pathlib.Path.is_relative_to()` (Python 3.9+) for containment checks
- Never construct file paths from user input without validation

### Input Validation

Validate all external input at system boundaries:

```python
# Pydantic (recommended for structured data)
from pydantic import BaseModel, Field, field_validator

class CreateUserRequest(BaseModel):
    name: str = Field(min_length=1, max_length=100)
    email: str = Field(pattern=r"^[\w.+-]+@[\w-]+\.[\w.]+$")
    age: int = Field(ge=0, le=150)

    @field_validator("name")
    @classmethod
    def no_script_tags(cls, v: str) -> str:
        if "<script" in v.lower():
            raise ValueError("HTML not allowed in name")
        return v.strip()

# Manual validation for simple cases
def validate_port(port: str) -> int:
    try:
        p = int(port)
    except ValueError:
        raise ValueError(f"Invalid port: {port!r}")
    if not (1 <= p <= 65535):
        raise ValueError(f"Port out of range: {p}")
    return p
```

### Secrets Management

Never hardcode secrets in source code:

```python
# DANGEROUS - Hardcoded secrets
API_KEY = "REDACTED"  # Leaked in git history forever
db_url = "postgresql://user:REDACTED@prod-db:5432/app"

# SAFE - Environment variables
import os
API_KEY = os.environ["API_KEY"]  # Fails loudly if missing

# SAFE - With default for optional config
DEBUG = os.environ.get("DEBUG", "false").lower() == "true"

# SAFE - Vault/secrets manager for production
from hvac import Client
vault = Client(url=os.environ["VAULT_ADDR"])
secret = vault.secrets.kv.v2.read_secret_version(path="myapp/creds")
```

**Validation:** Prescan pattern P17 detects common secret patterns (API keys, passwords in strings)

### Subprocess Safety

Avoid `shell=True` — it enables command injection:

```python
# DANGEROUS - Shell injection
filename = request.args["file"]
subprocess.run(f"cat {filename}", shell=True)  # Attacker sends: "; rm -rf /"

# SAFE - List arguments, no shell
subprocess.run(["cat", filename], check=True, capture_output=True)

# SAFE - For complex pipelines, use Python instead of shell
from pathlib import Path
content = Path(filename).read_text()

# If shell=True is truly needed, validate input strictly
import shlex
safe_arg = shlex.quote(user_input)
```

### ALWAYS / NEVER Rules

| Rule | Category | Detail |
|------|----------|--------|
| **ALWAYS** use parameterized queries | SQL | Never interpolate user input into SQL strings |
| **ALWAYS** validate URLs before fetch | SSRF | Check scheme, hostname against allowlist |
| **ALWAYS** resolve and check path prefix | Path Traversal | Use `pathlib.resolve()` + `is_relative_to()` |
| **ALWAYS** use `secrets` module for tokens | Crypto | `secrets.token_urlsafe()`, not `random` |
| **ALWAYS** set request timeouts | Network | `requests.get(url, timeout=10)` |
| **NEVER** use `eval()`/`exec()` on user input | Injection | Use `ast.literal_eval` or dispatch maps |
| **NEVER** unpickle untrusted data | Deserialization | Use JSON or msgpack instead |
| **NEVER** use `shell=True` with user input | Command injection | Use list args with `subprocess.run` |
| **NEVER** hardcode secrets | Secrets | Use env vars or vault |
| **NEVER** disable TLS verification | TLS | No `verify=False` in production |
| **NEVER** log secrets or tokens | Logging | Redact sensitive fields before logging |

---

## Anti-Patterns Avoided

> See `common-standards.md` for universal anti-patterns across all languages.

### No God Functions

```python
# Bad - Single function doing everything
def process_all(data):
    # 200+ lines of validation, transformation, saving, logging...
    pass

# Good - Separated concerns
def validate(data: Data) -> ValidationResult:
    ...

def transform(data: Data) -> TransformedData:
    ...

def save(data: TransformedData) -> None:
    ...
```

### No Bare Except

```python
# Bad
try:
    risky_operation()
except:
    pass

# Good
try:
    risky_operation()
except SpecificError as e:
    logging.warning("Operation failed: %s", e)
```

### No Global Mutable State

```python
# Bad
config = {}  # Module-level mutable

def load_config(path):
    global config
    config = load_yaml(path)

# Good
@dataclass
class Config:
    setting_a: str
    setting_b: int

def load_config(path: Path) -> Config:
    data = load_yaml(path)
    return Config(**data)
```

### No Magic Strings

```python
# Bad
if status == "pending":
    ...
elif status == "complete":
    ...

# Good
class Status(str, Enum):
    PENDING = "pending"
    COMPLETE = "complete"

if status == Status.PENDING:
    ...
```

---

## Compliance Assessment

**Use letter grades + evidence, NOT numeric scores.**

### Assessment Categories

| Category | Evidence Required |
|----------|------------------|
| **Code Quality** | ruff violations count, auto-fixable count |
| **Complexity** | radon cc output, functions >CC10 count |
| **Type Safety** | % public functions with hints, missing count |
| **Error Handling** | Bare except count, specific exception count |
| **Testing** | pytest coverage (line/branch %), test count |
| **Documentation** | Docstring coverage %, missing count |

### Grading Scale

| Grade | Criteria |
|-------|----------|
| A+ | 0 ruff violations, 0 functions >CC10, 95%+ hints, 90%+ coverage |
| A | <5 ruff violations, <3 functions >CC10, 85%+ hints, 80%+ coverage |
| A- | <15 ruff violations, <8 functions >CC10, 75%+ hints, 70%+ coverage |
| B+ | <30 ruff violations, <15 functions >CC10, 60%+ hints, 60%+ coverage |
| B | <50 ruff violations, <25 functions >CC10, 50%+ hints, 50%+ coverage |
| C | Significant issues, major refactoring needed |
| D | Not production-ready |
| F | Critical issues |

### Example Assessment

```markdown
## Python Standards Compliance

**Target:** src/
**Date:** 2026-01-21

| Category | Grade | Evidence |
|----------|-------|----------|
| Code Quality | A- | 8 ruff violations (6 fixable), 0 security |
| Complexity | B+ | 12 functions >CC10, avg CC=6.8 (radon) |
| Type Safety | A | 47/52 public functions typed (90%) |
| Error Handling | A- | 0 bare except, 2 broad catches |
| Testing | B | 73% line, 58% branch (pytest) |
| Documentation | A | 48/52 documented (92%, interrogate) |
| **OVERALL** | **A-** | **8 HIGH, 15 MEDIUM findings** |

### High Priority Findings

- **CMPLX-001** - `processor.py:89` CC=15 - Refactor dispatch
- **TYPE-001** - `utils.py` - 5 functions missing hints
```

---

## Vibe Integration

### Prescan Patterns

| Pattern | Severity | Detection |
|---------|----------|-----------|
| P04: Bare Except | HIGH | `except:` or `except Exception:` without re-raise |
| P08: print() Debug | MEDIUM | `print(` in non-CLI modules |
| P15: f-string Logging | LOW | `logging.*\(f"` pattern |

### JIT Loading

**Tier 1 (Fast):** Load `~/.agents/skills/standards/references/python.md` (5KB)
**Tier 2 (Deep):** Load this document (20KB) for comprehensive audit

---

## Additional Resources

- [PEP 8 - Style Guide](https://peps.python.org/pep-0008/)
- [Google Python Style Guide](https://google.github.io/styleguide/pyguide.html)
- [ruff Documentation](https://docs.astral.sh/ruff/)
- [radon Complexity](https://radon.readthedocs.io/)
- [pytest Documentation](https://docs.pytest.org/)
- [testcontainers-python](https://testcontainers-python.readthedocs.io/)

---

**Related:** `python-patterns.md` for quick reference examples (if needed)

### report-format.md

# Vibe Report Formats

## Output Files

| File | Purpose |
|------|---------|
| `reports/vibe-report.json` | Full JSON findings |
| `reports/vibe-junit.xml` | CI integration (JUnit XML) |
| `.agents/assessments/{date}-vibe-validate-{target}.md` | Knowledge artifact |

---

## JSON Report Structure

```json
{
  "summary": {
    "critical": 0,
    "high": 2,
    "medium": 5,
    "low": 1,
    "total": 8
  },
  "prescan": [
    {
      "id": "P4",
      "pattern": "Invisible Undone",
      "severity": "HIGH",
      "file": "services/auth/main.py",
      "line": 42,
      "message": "TODO marker"
    }
  ],
  "semantic": [
    {
      "id": "FAITH-001",
      "category": "docstrings",
      "severity": "HIGH",
      "file": "services/auth/main.py",
      "function": "validate_token",
      "message": "Docstring claims validation but no raise/return False"
    }
  ]
}
```

---

## JUnit XML Format

For CI integration:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<testsuites name="vibe-validate" tests="8" failures="7" errors="1">
  <testsuite name="prescan" tests="3" failures="3">
    <testcase name="P4-services/auth/main.py:42" classname="prescan.invisible_undone">
      <failure message="TODO marker" type="HIGH"/>
    </testcase>
  </testsuite>
  <testsuite name="semantic" tests="5" failures="4">
    <testcase name="FAITH-001-validate_token" classname="semantic.docstrings">
      <failure message="Docstring mismatch" type="HIGH"/>
    </testcase>
  </testsuite>
</testsuites>
```

---

## Assessment Artifact Format

Saved to `.agents/assessments/`:

```yaml
---
date: 2025-01-03
type: Assessment
assessment_type: vibe-validate
scope: recent
target: HEAD~1..HEAD
status: PASS_WITH_WARNINGS
severity: HIGH
findings:
  critical: 0
  high: 2
  medium: 5
  low: 1
  total: 8
tags: [assessment, vibe-validate, validation, recent]
---

# Vibe Validation: recent

## Summary

| Severity | Count |
|----------|-------|
| CRITICAL | 0 |
| HIGH | 2 |
| MEDIUM | 5 |
| LOW | 1 |

## Critical Findings

None.

## High Findings

1. **P4** `services/auth/main.py:42` - TODO marker
2. **FAITH-001** `validate_token()` - Docstring mismatch

## Recommendations

1. Complete or remove TODO at services/auth/main.py:42
2. Update validate_token() docstring to match implementation
```

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success, no CRITICAL findings |
| 1 | Argument/usage error |
| 2 | CRITICAL findings detected |
| 3 | HIGH findings detected (no CRITICAL) |

### rust-standards.md

# Rust Standards Catalog - Vibe Canonical Reference

**Version:** 1.0.0
**Last Updated:** 2026-02-09
**Purpose:** Canonical Rust standards for vibe skill validation

---

## Table of Contents

1. [Project Structure](#project-structure)
2. [Cargo Configuration](#cargo-configuration)
3. [Code Formatting](#code-formatting)
4. [Ownership & Borrowing](#ownership--borrowing)
5. [Error Handling Patterns](#error-handling-patterns)
6. [Trait & Type System Design](#trait--type-system-design)
7. [Concurrency Patterns](#concurrency-patterns)
8. [Unsafe Code](#unsafe-code)
9. [Testing Patterns](#testing-patterns)
10. [Security Practices](#security-practices)
11. [Documentation Standards](#documentation-standards)
12. [Code Quality Metrics & Anti-Patterns](#code-quality-metrics--anti-patterns)

---

## Project Structure

### ✅ **Standard Crate Layout**

```
my-project/
├── Cargo.toml             # Package manifest
├── Cargo.lock             # Dependency lock (commit for binaries, .gitignore for libs)
├── src/
│   ├── lib.rs             # Library root (public API surface)
│   ├── main.rs            # Binary entrypoint (or use src/bin/)
│   ├── bin/
│   │   ├── server.rs      # Additional binary
│   │   └── cli.rs         # Additional binary
│   ├── models/
│   │   └── mod.rs         # Domain types
│   └── handlers/
│       └── mod.rs         # Request handlers
├── tests/                 # Integration tests (each file is a separate crate)
│   ├── integration.rs
│   └── e2e.rs
├── examples/              # Runnable examples (`cargo run --example`)
│   └── basic_usage.rs
├── benches/               # Benchmarks (`cargo bench`)
│   └── throughput.rs
└── build.rs               # Build script (optional)
```

**Principles:**
- ✅ `src/lib.rs` defines the public API; `src/main.rs` consumes it
- ✅ `src/bin/` for multiple binaries within one crate
- ✅ `tests/` for integration tests (compiled as separate crates)
- ✅ `examples/` for documentation-as-code
- ✅ Commit `Cargo.lock` for binaries, omit for libraries

### ⚠️ **Module Organization**

```rust
// GOOD - Explicit re-exports in lib.rs
pub mod config;
pub mod handlers;
pub mod models;

pub use config::Config;
pub use models::AppError;

// BAD - Deep nesting with no re-exports
// Forces users to write: my_crate::handlers::http::v1::webhook::process
```

**Module Size Thresholds:**

| File Size | Status | Action |
|-----------|--------|--------|
| < 300 lines | ✅ Excellent | Maintain |
| 300-500 lines | ✅ Acceptable | Monitor |
| 500-800 lines | ⚠️ Warning | Consider splitting |
| 800+ lines | ❌ Critical | Split into submodules |

---

## Cargo Configuration

### ✅ **Dependency Management**

```toml
[package]
name = "my-service"
version = "0.1.0"
edition = "2021"
rust-version = "1.75"       # MSRV - minimum supported Rust version

[dependencies]
tokio = { version = "1", features = ["full"] }
serde = { version = "1", features = ["derive"] }
serde_json = "1"
thiserror = "2"
tracing = "0.1"

[dev-dependencies]
tokio = { version = "1", features = ["test-util", "macros"] }
proptest = "1"
criterion = { version = "0.5", features = ["html_reports"] }

[build-dependencies]
prost-build = "0.13"        # Only if needed at build time
```

**Requirements:**
- ✅ Pin `edition` and `rust-version` for reproducibility
- ✅ Use feature flags to minimize compile-time and binary size
- ✅ Separate `dev-dependencies` from production deps
- ✅ Never use wildcard versions (`*`)

### ✅ **Feature Flags**

```toml
[features]
default = ["json"]
json = ["dep:serde_json"]
tls = ["dep:rustls"]
full = ["json", "tls"]

# Optional dependencies gated by feature
[dependencies]
serde_json = { version = "1", optional = true }
rustls = { version = "0.23", optional = true }
```

**Why This Matters:**
- Users opt into functionality they need
- Reduces compile time and binary size
- Avoids pulling transitive dependencies unnecessarily

### ✅ **Profile Configuration**

```toml
[profile.release]
lto = true          # Link-time optimization
codegen-units = 1   # Single codegen unit for max optimization
strip = true        # Strip debug symbols from binary
panic = "abort"     # Smaller binary, no unwinding

[profile.dev]
opt-level = 0       # Fast compile
debug = true        # Full debug info

[profile.test]
opt-level = 1       # Slight optimization for faster test runs
```

### ✅ **Workspace Configuration**

```toml
# Root Cargo.toml
[workspace]
members = [
    "crates/core",
    "crates/api",
    "crates/cli",
]

[workspace.dependencies]
serde = { version = "1", features = ["derive"] }
tokio = { version = "1", features = ["full"] }

# In member Cargo.toml
[dependencies]
serde = { workspace = true }
tokio = { workspace = true }
```

**Benefits:**
- Single lockfile across all crates
- Unified dependency versions
- `cargo test --workspace` runs all tests

---

## Code Formatting

### ✅ **rustfmt Configuration**

```toml
# rustfmt.toml
edition = "2021"
max_width = 100
tab_spaces = 4
use_field_init_shorthand = true
use_try_shorthand = true
imports_granularity = "Module"
group_imports = "StdExternalCrate"
```

**Requirements:**
- ✅ Run `cargo fmt --check` in CI (zero-tolerance for formatting drift)
- ✅ `group_imports = "StdExternalCrate"` enforces import order: std, external, crate-local

### ✅ **Import Grouping**

```rust
// GOOD - Grouped and ordered
use std::collections::HashMap;
use std::sync::Arc;

use serde::{Deserialize, Serialize};
use tokio::sync::Mutex;

use crate::config::Config;
use crate::models::AppError;

// BAD - Unorganized imports
use crate::config::Config;
use std::collections::HashMap;
use serde::Serialize;
use std::sync::Arc;
use crate::models::AppError;
use tokio::sync::Mutex;
```

### ✅ **Naming Conventions**

| Item | Convention | Example |
|------|-----------|---------|
| Types, Traits | `UpperCamelCase` | `HttpClient`, `Serialize` |
| Functions, Methods | `snake_case` | `process_request` |
| Local Variables | `snake_case` | `retry_count` |
| Constants | `SCREAMING_SNAKE_CASE` | `MAX_RETRIES` |
| Modules | `snake_case` | `error_handling` |
| Type Parameters | Single uppercase or `CamelCase` | `T`, `Item` |
| Lifetimes | Short lowercase | `'a`, `'ctx` |
| Crate Names | `kebab-case` (Cargo.toml) | `my-service` |
| Feature Flags | `kebab-case` | `full-json` |

### ⚠️ **Naming Anti-Patterns**

| Pattern | Problem | Instead |
|---------|---------|---------|
| `get_` prefix on getters | Redundant in Rust | `fn name(&self)` not `fn get_name(&self)` |
| `FooStruct` suffix | Redundant | `Foo` |
| `IFoo` prefix on traits | Not idiomatic Rust | `Foo` trait, `FooImpl` if needed |
| Single-letter variable names | Unreadable (except in closures/iterators) | Descriptive names |

---

## Ownership & Borrowing

### ✅ **Prefer Borrowing Over Ownership**

```rust
// GOOD - Borrows the string, caller retains ownership
fn validate_email(email: &str) -> bool {
    email.contains('@') && email.contains('.')
}

// BAD - Takes ownership unnecessarily
fn validate_email(email: String) -> bool {
    email.contains('@') && email.contains('.')
}
```

**Why This Matters:**
- Borrowing avoids unnecessary allocations and clones
- Caller retains ownership for reuse
- `&str` accepts both `String` and `&str` via deref coercion

### ✅ **Lifetime Annotations**

```rust
// GOOD - Explicit lifetime ties output to input
fn first_word(s: &str) -> &str {
    s.split_whitespace().next().unwrap_or("")
}

// GOOD - Multiple lifetimes when inputs have different scopes
fn longest<'a>(x: &'a str, y: &'a str) -> &'a str {
    if x.len() > y.len() { x } else { y }
}

// GOOD - Struct borrowing data
struct Config<'a> {
    name: &'a str,
    version: &'a str,
}
```

**Lifetime Elision Rules (when annotations are NOT needed):**
1. Each reference parameter gets its own lifetime
2. If exactly one input lifetime, it applies to all output lifetimes
3. If `&self` or `&mut self`, its lifetime applies to all output lifetimes

### ✅ **Copy vs Clone**

```rust
// GOOD - Small, stack-only types implement Copy
#[derive(Debug, Clone, Copy, PartialEq)]
struct Point {
    x: f64,
    y: f64,
}

// GOOD - Types with heap data implement Clone only
#[derive(Debug, Clone)]
struct Config {
    name: String,       // String is Clone but NOT Copy
    retries: u32,
}
```

| Trait | Behavior | Use When |
|-------|----------|----------|
| `Copy` | Implicit bitwise copy | Small stack-only types (integers, bools, tuples of Copy types) |
| `Clone` | Explicit `.clone()` | Heap-allocated or expensive-to-copy types |
| Neither | Move semantics | Unique resources (file handles, connections) |

### ⚠️ **Common Ownership Mistakes**

```rust
// BAD - Unnecessary clone to satisfy borrow checker
let name = config.name.clone();
process(&name);
process2(&config.name); // Could have borrowed directly

// GOOD - Borrow instead of clone
process(&config.name);
process2(&config.name);

// BAD - Moving out of a shared reference
fn take_name(config: &Config) -> String {
    config.name // ERROR: cannot move out of borrowed content
}

// GOOD - Clone when you truly need ownership from a borrow
fn take_name(config: &Config) -> String {
    config.name.clone()
}
```

---

## Error Handling Patterns

### ✅ **Custom Error Types with thiserror**

```rust
use thiserror::Error;

#[derive(Debug, Error)]
pub enum AppError {
    #[error("configuration error: {0}")]
    Config(String),

    #[error("database query failed: {source}")]
    Database {
        #[source]
        source: sqlx::Error,
    },

    #[error("HTTP request failed: {url}")]
    Http {
        url: String,
        #[source]
        source: reqwest::Error,
    },

    #[error("not found: {entity} with id {id}")]
    NotFound { entity: &'static str, id: String },

    #[error(transparent)]
    Unexpected(#[from] anyhow::Error),
}
```

**Requirements:**
- ✅ Use `thiserror` for library error types (structured, matchable)
- ✅ Use `anyhow` for application-level errors (ergonomic, context-rich)
- ✅ Implement `#[source]` for error chain inspection
- ✅ Implement `#[from]` for automatic conversion via `?`
- ✅ Human-readable display messages

### ✅ **The ? Operator and Error Propagation**

```rust
// GOOD - Clean error propagation with context
use anyhow::{Context, Result};

fn load_config(path: &str) -> Result<Config> {
    let contents = std::fs::read_to_string(path)
        .with_context(|| format!("failed to read config from {path}"))?;

    let config: Config = toml::from_str(&contents)
        .with_context(|| format!("failed to parse config from {path}"))?;

    config.validate()
        .context("config validation failed")?;

    Ok(config)
}

// BAD - Manual match on every error
fn load_config(path: &str) -> Result<Config, Box<dyn std::error::Error>> {
    let contents = match std::fs::read_to_string(path) {
        Ok(c) => c,
        Err(e) => return Err(Box::new(e)),
    };
    // ... tedious repetition
}
```

### ✅ **Result Type Aliases**

```rust
// GOOD - Crate-level Result alias
pub type Result<T> = std::result::Result<T, AppError>;

// Usage throughout the crate
pub fn get_user(id: u64) -> Result<User> {
    // AppError is the implicit error type
    Ok(User { id, name: "Alice".into() })
}
```

### ⚠️ **Error Handling Anti-Patterns**

| Pattern | Problem | Instead |
|---------|---------|---------|
| `.unwrap()` in production | Panics on None/Err | Use `?`, `.unwrap_or()`, or match |
| `Box<dyn Error>` everywhere | Loses type info | Use `thiserror` enums |
| String errors | Not matchable | Use typed errors |
| Swallowing errors silently | Hides bugs | Log or propagate |
| `panic!()` for expected failures | Crashes the process | Return `Result` |

**Unwrap Threshold:**

| Context | `.unwrap()` Allowed? |
|---------|---------------------|
| Tests | ✅ Yes |
| Examples | ✅ Yes |
| Build scripts | ⚠️ Acceptable with comment |
| Library code | ❌ Never |
| Binary (main) | ⚠️ Only after validation |

---

## Trait & Type System Design

### ✅ **Trait Design**

```rust
// GOOD - Small, focused traits
pub trait Validate {
    fn validate(&self) -> Result<(), ValidationError>;
}

pub trait Persist {
    fn save(&self, store: &dyn Store) -> Result<()>;
    fn load(id: &str, store: &dyn Store) -> Result<Self>
    where
        Self: Sized;
}

// Compose traits via supertraits
pub trait Entity: Validate + Persist + std::fmt::Debug {}
```

**Anti-Pattern (God Trait):**
```rust
// BAD - Too many methods, forces implementors to define everything
pub trait Service {
    fn start(&self) -> Result<()>;
    fn stop(&self) -> Result<()>;
    fn health(&self) -> HealthStatus;
    fn metrics(&self) -> Metrics;
    fn configure(&mut self, config: Config);
    fn validate(&self) -> Result<()>;
    // ... 15 more methods
}
```

### ✅ **Generics vs Trait Objects**

```rust
// GOOD - Static dispatch (monomorphized, zero-cost abstraction)
fn process<T: Serialize + Send>(item: T) -> Result<()> {
    let json = serde_json::to_string(&item)?;
    send_to_queue(&json)
}

// GOOD - Dynamic dispatch (runtime polymorphism, smaller binary)
fn process_any(item: &dyn Serialize) -> Result<()> {
    let json = serde_json::to_value(item)?;
    send_to_queue(&json.to_string())
}
```

| Approach | Binary Size | Performance | Use When |
|----------|------------|-------------|----------|
| Generics (`T: Trait`) | Larger (monomorphized) | Faster (inlined) | Hot paths, known types at compile time |
| Trait Objects (`dyn Trait`) | Smaller | Vtable overhead | Collections of mixed types, plugin systems |
| `impl Trait` (return) | Smaller | Inlined | Returning closures or iterators |

### ✅ **Associated Types vs Generics**

```rust
// GOOD - Associated type when there's ONE natural choice per impl
pub trait Iterator {
    type Item;
    fn next(&mut self) -> Option<Self::Item>;
}

// GOOD - Generic parameter when impl can work with MANY types
pub trait From<T> {
    fn from(value: T) -> Self;
}
```

### ✅ **Derive Macros**

```rust
// GOOD - Derive common traits
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct User {
    pub id: u64,
    pub name: String,
    pub email: String,
}
```

**Standard Derive Order:**

| Priority | Traits | Purpose |
|----------|--------|---------|
| 1 | `Debug` | Always derive for debugging |
| 2 | `Clone`, `Copy` | If semantically appropriate |
| 3 | `PartialEq`, `Eq` | If comparison is needed |
| 4 | `Hash` | If used as HashMap key |
| 5 | `Serialize`, `Deserialize` | If serialized |
| 6 | `Default` | If zero-value makes sense |

---

## Concurrency Patterns

### ✅ **Shared State with Arc<Mutex<T>>**

```rust
use std::sync::Arc;
use tokio::sync::Mutex;

#[derive(Clone)]
struct AppState {
    db: Arc<Mutex<Database>>,
    cache: Arc<dashmap::DashMap<String, String>>,
}

// GOOD - Lock scope is minimal
async fn get_user(state: &AppState, id: u64) -> Result<User> {
    let db = state.db.lock().await;
    let user = db.query_user(id).await?;
    drop(db); // Explicit drop releases lock before further processing
    Ok(user)
}

// BAD - Holding lock across await points
async fn bad_get_user(state: &AppState, id: u64) -> Result<User> {
    let db = state.db.lock().await;
    let user = db.query_user(id).await?; // Lock held across .await!
    let enriched = enrich_user(user).await; // Still holding lock!
    Ok(enriched)
}
```

**Lock Duration Thresholds:**

| Duration | Status | Action |
|----------|--------|--------|
| < 1 ms | ✅ Excellent | Maintain |
| 1-10 ms | ⚠️ Warning | Review scope |
| > 10 ms | ❌ Critical | Refactor (clone-and-release pattern) |

### ✅ **Channel Patterns**

```rust
use tokio::sync::mpsc;

// GOOD - Bounded channel with backpressure
let (tx, mut rx) = mpsc::channel::<Event>(100);

// Producer
tokio::spawn(async move {
    for event in events {
        if tx.send(event).await.is_err() {
            tracing::warn!("receiver dropped, stopping producer");
            break;
        }
    }
});

// Consumer
tokio::spawn(async move {
    while let Some(event) = rx.recv().await {
        process_event(event).await;
    }
});
```

### ✅ **Send and Sync Bounds**

```rust
// GOOD - Explicit Send + Sync bounds for spawned futures
fn spawn_worker<F>(task: F) -> tokio::task::JoinHandle<()>
where
    F: Future<Output = ()> + Send + 'static,
{
    tokio::spawn(task)
}

// GOOD - Ensure types are thread-safe
struct SharedConfig {
    data: Arc<RwLock<HashMap<String, String>>>,  // Send + Sync
}
```

| Marker | Meaning | NOT Send/Sync |
|--------|---------|---------------|
| `Send` | Can be transferred across threads | `Rc<T>`, `*const T` |
| `Sync` | Can be shared between threads via `&T` | `Cell<T>`, `RefCell<T>` |
| Both | Safe for concurrent access | `Arc<Mutex<T>>` is both |

### ✅ **Async/Await Best Practices**

```rust
// GOOD - Select for racing multiple futures
tokio::select! {
    result = process_request(&req) => {
        handle_response(result).await;
    }
    _ = tokio::time::sleep(Duration::from_secs(30)) => {
        return Err(AppError::Timeout);
    }
    _ = shutdown_signal.recv() => {
        tracing::info!("shutting down gracefully");
        return Ok(());
    }
}

// GOOD - Spawn blocking work off the async runtime
let hash = tokio::task::spawn_blocking(move || {
    argon2::hash_encoded(password.as_bytes(), &salt, &config)
}).await??;
```

---

## Unsafe Code

### ✅ **SAFETY Comments (Required)**

```rust
// GOOD - Every unsafe block has a SAFETY comment
let value = unsafe {
    // SAFETY: We verified that `ptr` is non-null and properly aligned
    // in the check above (line 42). The pointed-to data is initialized
    // by `init_buffer()` called on line 38 and has not been freed.
    *ptr
};

// BAD - Unsafe with no justification
let value = unsafe { *ptr };
```

**Requirements:**
- ✅ Every `unsafe` block must have a `// SAFETY:` comment
- ✅ Comment must explain WHY the invariants are upheld
- ✅ Reference the specific preconditions being satisfied

### ✅ **Minimizing Unsafe Scope**

```rust
// GOOD - Minimal unsafe block, safe wrapper
pub fn get_element(slice: &[u8], index: usize) -> Option<u8> {
    if index < slice.len() {
        // SAFETY: We just verified index is within bounds
        Some(unsafe { *slice.get_unchecked(index) })
    } else {
        None
    }
}

// BAD - Entire function is unsafe when only one operation needs it
pub unsafe fn get_element(slice: &[u8], index: usize) -> u8 {
    let ptr = slice.as_ptr().add(index);
    let extra = compute_offset(ptr); // This doesn't need unsafe!
    let result = *ptr;
    log_access(result);               // This doesn't need unsafe!
    result
}
```

### ✅ **FFI (Foreign Function Interface)**

```rust
// GOOD - Safe wrapper around FFI
mod ffi {
    extern "C" {
        fn c_process(data: *const u8, len: usize) -> i32;
    }
}

/// Process data using the C library.
///
/// # Errors
/// Returns `Err` if the C function returns a non-zero exit code.
pub fn process(data: &[u8]) -> Result<(), FfiError> {
    // SAFETY: `data.as_ptr()` is valid for `data.len()` bytes.
    // The C function does not retain the pointer after returning.
    let result = unsafe { ffi::c_process(data.as_ptr(), data.len()) };
    if result == 0 {
        Ok(())
    } else {
        Err(FfiError::ExitCode(result))
    }
}
```

### ⚠️ **Unsafe Code Thresholds**

| Metric | Status | Action |
|--------|--------|--------|
| 0 unsafe blocks | ✅ Ideal | Maintain |
| 1-5 with SAFETY comments | ✅ Acceptable | Audit quarterly |
| 6-15 with SAFETY comments | ⚠️ Warning | Justify each, seek safe alternatives |
| Any without SAFETY comments | ❌ Critical | Add comments immediately |
| `#[allow(unsafe_code)]` crate-wide | ❌ Critical | Remove, audit all unsafe |

---

## Testing Patterns

### ✅ **Unit Tests (Inline Modules)**

```rust
pub fn calculate_discount(price: f64, tier: &str) -> f64 {
    match tier {
        "gold" => price * 0.20,
        "silver" => price * 0.10,
        _ => 0.0,
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn gold_tier_gets_twenty_percent() {
        let discount = calculate_discount(100.0, "gold");
        assert!((discount - 20.0).abs() < f64::EPSILON);
    }

    #[test]
    fn unknown_tier_gets_no_discount() {
        assert_eq!(calculate_discount(100.0, "bronze"), 0.0);
    }

    #[test]
    fn zero_price_returns_zero() {
        assert_eq!(calculate_discount(0.0, "gold"), 0.0);
    }
}
```

**Requirements:**
- ✅ `#[cfg(test)]` module in the same file as the code
- ✅ Test names describe the expected behavior
- ✅ `use super::*` imports the parent module

### ✅ **Integration Tests**

```rust
// tests/api_integration.rs
// Each file in tests/ is compiled as a separate crate

use my_service::{Config, Server};

#[tokio::test]
async fn server_responds_to_health_check() {
    let config = Config::test_default();
    let server = Server::start(config).await.unwrap();

    let resp = reqwest::get(&format!("{}/health", server.url()))
        .await
        .unwrap();

    assert_eq!(resp.status(), 200);
    server.shutdown().await;
}
```

### ✅ **Doc Tests**

```rust
/// Parses a duration string like "5s", "100ms", "2m".
///
/// # Examples
///
/// ```
/// use my_crate::parse_duration;
///
/// let d = parse_duration("5s").unwrap();
/// assert_eq!(d, std::time::Duration::from_secs(5));
///
/// let d = parse_duration("100ms").unwrap();
/// assert_eq!(d, std::time::Duration::from_millis(100));
/// ```
///
/// # Errors
///
/// Returns `Err` if the string is not a valid duration format.
pub fn parse_duration(s: &str) -> Result<Duration, ParseError> {
    // ...
}
```

**Why Doc Tests Matter:**
- Examples in documentation are compiled and tested
- Guarantees documentation stays accurate
- `cargo test` runs doc tests by default

### ✅ **Property-Based Testing with proptest**

```rust
use proptest::prelude::*;

proptest! {
    #[test]
    fn roundtrip_serialization(input in "\\PC{1,100}") {
        let serialized = serde_json::to_string(&input).unwrap();
        let deserialized: String = serde_json::from_str(&serialized).unwrap();
        prop_assert_eq!(input, deserialized);
    }

    #[test]
    fn discount_never_exceeds_price(price in 0.0f64..10000.0, tier in "gold|silver|bronze") {
        let discount = calculate_discount(price, &tier);
        prop_assert!(discount <= price);
        prop_assert!(discount >= 0.0);
    }
}
```

### ✅ **Benchmarks with Criterion**

```rust
// benches/throughput.rs
use criterion::{black_box, criterion_group, criterion_main, Criterion};
use my_service::process;

fn benchmark_process(c: &mut Criterion) {
    let data = setup_test_data();

    c.bench_function("process_1000_items", |b| {
        b.iter(|| process(black_box(&data)))
    });
}

criterion_group!(benches, benchmark_process);
criterion_main!(benches);
```

**Running:**
```bash
cargo bench                       # Run all benchmarks
cargo bench -- process            # Run matching benchmarks
```

### Test Type Summary

| Type | Location | Runs With | Purpose |
|------|----------|-----------|---------|
| Unit | `#[cfg(test)]` inline | `cargo test` | Test private functions |
| Integration | `tests/` directory | `cargo test` | Test public API |
| Doc | `///` comments | `cargo test` | Verify examples |
| Property | Inline or `tests/` | `cargo test` | Fuzz invariants |
| Benchmark | `benches/` | `cargo bench` | Performance regression |

---

## Security Practices

### ✅ **Minimize Unsafe Code**

```rust
// CORRECT — isolate unsafe behind a safe API
pub fn read_buffer(ptr: *const u8, len: usize) -> &[u8] {
    // SAFETY: caller guarantees ptr is valid for `len` bytes,
    // properly aligned, and the memory won't be mutated during
    // the lifetime of the returned slice.
    unsafe { std::slice::from_raw_parts(ptr, len) }
}

// INCORRECT — unsafe scattered through business logic
pub fn process(data: *const u8) {
    unsafe {
        // Multiple unsafe operations without justification
        let val = *data;
        let next = *data.add(1);
    }
}
```

**Unsafe Audit Criteria:**
- Every `unsafe` block MUST have a `// SAFETY:` comment explaining the invariant
- Minimize the scope of `unsafe` — wrap in safe abstractions
- Prefer safe alternatives: `Vec`, `Box`, `Rc`/`Arc` over raw pointers
- Audit all `unsafe impl Send` and `unsafe impl Sync` for correctness

### ✅ **FFI Safety**

```rust
// CORRECT — safe wrapper around FFI
extern "C" {
    fn c_process(data: *const u8, len: usize) -> i32;
}

/// Process data through the C library.
///
/// # Panics
/// Panics if `data` is empty.
pub fn process(data: &[u8]) -> Result<(), FfiError> {
    assert!(!data.is_empty(), "data must not be empty");
    // SAFETY: data.as_ptr() is valid for data.len() bytes,
    // and c_process does not retain the pointer.
    let result = unsafe { c_process(data.as_ptr(), data.len()) };
    match result {
        0 => Ok(()),
        code => Err(FfiError::ReturnCode(code)),
    }
}
```

**FFI Rules:**
- Always validate inputs before crossing the FFI boundary
- Wrap every `extern "C"` function in a safe Rust API
- Never expose raw pointers in public APIs
- Use `CStr`/`CString` for string interchange, never cast directly

### ✅ **Integer Overflow**

```rust
// CORRECT — use checked arithmetic for untrusted inputs
fn allocate_buffer(count: usize, item_size: usize) -> Result<Vec<u8>, AllocError> {
    let total = count.checked_mul(item_size)
        .ok_or(AllocError::Overflow)?;
    Ok(vec![0u8; total])
}

// CORRECT — use saturating arithmetic for counters/metrics
fn increment_retry(count: u32) -> u32 {
    count.saturating_add(1) // Caps at u32::MAX instead of wrapping
}

// INCORRECT — silent wrapping in release mode
fn total_size(count: usize, item_size: usize) -> usize {
    count * item_size // Wraps silently in release, panics in debug
}
```

**Integer Overflow Rules:**

| Context | Strategy | Method |
|---------|----------|--------|
| Untrusted input | Checked | `checked_add`, `checked_mul` — returns `None` on overflow |
| Counters / metrics | Saturating | `saturating_add` — caps at MAX |
| Bit manipulation | Wrapping | `wrapping_add` — intentional modular arithmetic |
| Debug assertions | Default | Panics in debug, wraps in release |

- Prefer `checked_*` for any arithmetic involving external data
- Use `saturating_*` when capping is acceptable (progress bars, counters)
- Use `wrapping_*` only for intentional modular arithmetic (hashing, crypto)
- Add `#[deny(clippy::arithmetic_side_effects)]` for high-assurance modules

### ✅ **Panic Handling**

```rust
// CORRECT — catch_unwind at FFI and thread boundaries
use std::panic;

pub extern "C" fn ffi_entry_point(input: *const u8, len: usize) -> i32 {
    let result = panic::catch_unwind(|| {
        // SAFETY: caller guarantees valid pointer and length
        let data = unsafe { std::slice::from_raw_parts(input, len) };
        process(data)
    });
    match result {
        Ok(Ok(())) => 0,
        Ok(Err(_)) => -1,
        Err(_panic) => -2, // Caught a panic — do not unwind into C
    }
}

// CORRECT — prefer Result over panic for recoverable errors
pub fn parse_port(s: &str) -> Result<u16, ParseError> {
    let port: u16 = s.parse().map_err(|_| ParseError::InvalidPort)?;
    if port == 0 {
        return Err(ParseError::PortZero);
    }
    Ok(port)
}

// INCORRECT — panic for expected failure
pub fn parse_port(s: &str) -> u16 {
    s.parse().expect("invalid port") // Crashes on bad input
}
```

**Panic Rules:**
- Use `catch_unwind` at FFI boundaries to prevent unwinding into C/C++
- Use `catch_unwind` in thread pool workers to prevent poisoning the pool
- Return `Result` for all recoverable errors — reserve `panic!` for programmer bugs
- Set `panic = "abort"` in release profile if unwinding is not needed (smaller binary)
- Use `assert!` only for invariants that indicate a bug if violated

### ✅ **Input Validation**

```rust
use std::net::IpAddr;

pub fn parse_config(input: &str) -> Result<Config, ConfigError> {
    let config: Config = toml::from_str(input)
        .map_err(ConfigError::Parse)?;

    // Validate bounds after deserialization
    if config.port == 0 || config.port > 65535 {
        return Err(ConfigError::InvalidPort(config.port));
    }
    if config.max_connections > 10_000 {
        return Err(ConfigError::ExceedsLimit("max_connections", 10_000));
    }

    Ok(config)
}
```

**Validation Rules:**
- Validate all external data at system boundaries (CLI args, env vars, files, network)
- Use newtypes to enforce invariants at the type level
- Prefer `TryFrom` over unchecked conversions

### ✅ **Dependency Auditing**

```bash
# Audit for known vulnerabilities
cargo audit

# Check for unmaintained or yanked crates
cargo audit --deny warnings

# In CI, fail the build on any advisory
cargo audit --deny vulnerability --deny unmaintained --deny yanked
```

**Dependency Rules:**
- Run `cargo audit` in CI on every PR
- Pin major versions in `Cargo.toml` (e.g., `serde = "1"`)
- Review new dependencies for `unsafe` usage and maintenance status
- Prefer crates from the RustSec-reviewed ecosystem

### Security ALWAYS / NEVER

| ALWAYS | NEVER |
|--------|-------|
| Add `// SAFETY:` comment on every `unsafe` block | Use `unsafe` without documenting the invariant |
| Wrap FFI calls in safe Rust abstractions | Expose raw pointers in public APIs |
| Validate external inputs at system boundaries | Trust deserialized data without bounds checks |
| Run `cargo audit` in CI | Ignore advisory warnings on dependencies |
| Use `CStr`/`CString` for C string interchange | Cast `*const u8` to `&str` without validation |
| Minimize `unsafe` block scope | Scatter `unsafe` through business logic |

---

## Documentation Standards

### ✅ **Rustdoc Conventions**

```rust
//! # mycrate
//!
//! A library for processing widgets efficiently.
//!
//! ## Quick Start
//!
//! ```rust
//! use mycrate::Widget;
//! let w = Widget::new("example");
//! assert!(w.is_valid());
//! ```

/// A widget that can be processed.
///
/// Widgets are the core data type. They must be created
/// via [`Widget::new`] to ensure invariants are upheld.
///
/// # Examples
///
/// ```
/// use mycrate::Widget;
///
/// let widget = Widget::new("test");
/// assert_eq!(widget.name(), "test");
/// ```
pub struct Widget {
    name: String,
}

impl Widget {
    /// Creates a new widget with the given name.
    ///
    /// # Panics
    ///
    /// Panics if `name` is empty.
    ///
    /// # Examples
    ///
    /// ```
    /// use mycrate::Widget;
    /// let w = Widget::new("example");
    /// ```
    pub fn new(name: &str) -> Self {
        assert!(!name.is_empty(), "name must not be empty");
        Self { name: name.to_string() }
    }
}
```

### ✅ **Doc Test Patterns**

```rust
/// Parses a duration string like "5s", "100ms".
///
/// # Examples
///
/// ```
/// # use mycrate::parse_duration;
/// assert_eq!(parse_duration("5s").unwrap().as_secs(), 5);
/// assert_eq!(parse_duration("100ms").unwrap().as_millis(), 100);
/// ```
///
/// # Errors
///
/// Returns [`ParseError`] if the format is unrecognized.
///
/// ```
/// # use mycrate::parse_duration;
/// assert!(parse_duration("invalid").is_err());
/// ```
pub fn parse_duration(s: &str) -> Result<Duration, ParseError> {
    // ...
}
```

**Doc Test Rules:**
- Doc tests compile and run with `cargo test` — treat them as real tests
- Use `# ` prefix to hide setup lines (imports, boilerplate)
- Use `no_run` for examples that need network/filesystem
- Use `should_panic` for examples demonstrating failure

### ✅ **Module Documentation**

```rust
//! # handlers
//!
//! HTTP request handlers for the API.
//!
//! Each handler follows the pattern:
//! 1. Parse and validate input
//! 2. Call domain logic
//! 3. Map result to HTTP response
//!
//! See [`crate::domain`] for business logic.
```

### ✅ **`#[doc(hidden)]` Usage**

```rust
// Hide implementation details from public docs
#[doc(hidden)]
pub mod __internal {
    // Used by macros, not part of public API
}

// Hide trait impls that are required but not user-facing
#[doc(hidden)]
pub fn __macro_helper() {}
```

**When to use `#[doc(hidden)]`:**
- Macro support functions that must be `pub` but are not API
- Trait implementations required by the compiler but meaningless to users
- Never hide things to avoid documenting them

### ✅ **README Integration**

```bash
# Generate README.md from lib.rs module-level docs
cargo install cargo-readme
cargo readme > README.md

# Verify README stays in sync (CI check)
cargo readme | diff - README.md
```

```rust
// src/lib.rs — module-level docs become README content
//! # mycrate
//!
//! A fast, safe widget processor.
//!
//! ## Features
//!
//! - Zero-copy parsing
//! - Async support via tokio
//! - Type-safe configuration
//!
//! ## Usage
//!
//! ```rust
//! use mycrate::process;
//!
//! let result = process("input").unwrap();
//! assert!(!result.is_empty());
//! ```
```

**README Rules:**
- Use `cargo-readme` to generate README.md from `//!` docs in `src/lib.rs`
- Keep the single source of truth in `lib.rs` — README is a derived artifact
- Add a CI step to verify README stays in sync with `lib.rs` docs
- Include: crate purpose, features, usage example, MSRV, license

### Documentation ALWAYS / NEVER

| ALWAYS | NEVER |
|--------|-------|
| Use `///` for public items, `//!` for module-level docs | Leave public API items undocumented |
| Include `# Examples` section on public functions | Write doc tests that don't actually assert behavior |
| Document `# Panics`, `# Errors`, `# Safety` sections | Use `#[doc(hidden)]` to avoid writing documentation |
| Run `cargo test` to verify doc examples compile | Assume doc examples stay correct without CI |
| Link to related items with [`ident`] syntax | Duplicate information already in type signatures |
| Use `# ` to hide boilerplate in doc tests | Write doc tests that require external services |

---

## Code Quality Metrics & Anti-Patterns

> See `common-standards.md` for universal coverage targets, testing principles, and anti-patterns across all languages.

### ✅ **Clippy Lint Levels**

```toml
# Cargo.toml or clippy.toml
[lints.clippy]
# Deny — treat as errors
unwrap_used = "deny"
expect_used = "deny"
panic = "deny"
todo = "deny"

# Warn — flag for review
clone_on_ref_ptr = "warn"
large_enum_variant = "warn"
needless_pass_by_value = "warn"
implicit_clone = "warn"
missing_errors_doc = "warn"
missing_panics_doc = "warn"
```

**Recommended CI Command:**
```bash
cargo clippy --all-targets --all-features -- -D warnings
```

### ✅ **Lint Category Enforcement**

| Category | CI Policy | Rationale |
|----------|-----------|-----------|
| `clippy::correctness` | ❌ Deny (fail build) | Likely bugs |
| `clippy::suspicious` | ❌ Deny (fail build) | Probably wrong |
| `clippy::pedantic` | ⚠️ Warn | Style improvements |
| `clippy::nursery` | ⚠️ Optional | Experimental lints |
| `clippy::cargo` | ⚠️ Warn | Cargo.toml hygiene |

### 📊 **Complexity Thresholds**

| Complexity Range | Status | Action |
|-----------------|--------|--------|
| CC 1-5 (Simple) | ✅ Excellent | Maintain |
| CC 6-10 (OK) | ✅ Acceptable | Monitor |
| CC 11-15 (High) | ⚠️ Warning | Refactor recommended |
| CC 16+ (Very High) | ❌ Critical | Refactor required |

**Coverage Targets:**

| Metric | Minimum | Target |
|--------|---------|--------|
| Line coverage | 60% | 80%+ |
| Branch coverage | 50% | 70%+ |
| Critical path coverage | 90% | 100% |

### ❌ **Named Anti-Patterns**

**1. Stringly-Typed Code**
```rust
// BAD - Strings for everything
fn set_status(status: &str) { /* "active", "idle", "error" */ }

// GOOD - Enums encode valid states
enum Status { Active, Idle, Error }
fn set_status(status: Status) { /* ... */ }
```

**2. Clone-Happy Code**
```rust
// BAD - Cloning to avoid borrow checker fights
fn process(data: &Data) {
    let owned = data.clone();   // Unnecessary allocation
    compute(&owned);
}

// GOOD - Work with references
fn process(data: &Data) {
    compute(data);
}
```

**3. Typestate Neglect**
```rust
// BAD - Runtime checks for compile-time invariants
struct Connection { is_authenticated: bool }
fn query(conn: &Connection) {
    assert!(conn.is_authenticated); // Runtime panic
}

// GOOD - Typestate pattern enforces at compile time
struct Unauthenticated;
struct Authenticated;
struct Connection<State> { _state: std::marker::PhantomData<State> }

impl Connection<Unauthenticated> {
    fn authenticate(self, creds: &Credentials) -> Result<Connection<Authenticated>> {
        // ...
    }
}

impl Connection<Authenticated> {
    fn query(&self, sql: &str) -> Result<Rows> {
        // Can only be called on authenticated connections
    }
}
```

**4. Arc<Mutex<T>> Everywhere**
```rust
// BAD - Mutex when only reads happen
let config = Arc::new(Mutex::new(load_config()));

// GOOD - Use RwLock for read-heavy workloads
let config = Arc::new(RwLock::new(load_config()));

// BETTER - Use Arc<T> if config is immutable after init
let config = Arc::new(load_config());
```

**5. Ignoring Must-Use Types**
```rust
// BAD - Ignoring a Result
fn fire_and_forget() {
    send_notification(); // Warning: unused Result
}

// GOOD - Explicitly acknowledge
fn fire_and_forget() {
    let _ = send_notification(); // Intentional ignore
}
```

**6. Unbounded Collections**
```rust
// BAD - No size limit on cache
let mut cache: HashMap<String, Data> = HashMap::new();
// Grows forever...

// GOOD - Bounded with eviction
let cache = lru::LruCache::new(NonZeroUsize::new(10_000).unwrap());
```

---

## Compliance Assessment

**Use letter grades + evidence, NOT numeric scores.**

| Category | Assessment Criteria | Evidence Required |
|----------|-------------------|-------------------|
| Project Structure | Standard layout, module sizes, re-exports | File count per module, module line counts |
| Cargo Config | MSRV set, features used, profiles configured | Cargo.toml audit, dep count |
| Code Formatting | rustfmt clean, naming conventions followed | `cargo fmt --check` output, naming violations |
| Ownership & Borrowing | Minimal clones, correct lifetimes, no unnecessary ownership | Clone count, borrow checker workarounds |
| Error Handling | thiserror/anyhow usage, no unwrap in prod, context added | Unwrap count, error type audit |
| Traits & Types | Small traits, appropriate dispatch, derive usage | Methods per trait, dyn vs generic ratio |
| Concurrency | Minimal lock scope, bounded channels, Send/Sync correct | Lock duration, channel audit |
| Unsafe Code | SAFETY comments, minimal scope, safe wrappers | Unsafe block count, comment coverage |
| Testing | Unit + integration + doc tests, property tests | Coverage %, test type distribution |
| Security | Unsafe minimized, FFI wrapped, inputs validated, deps audited | Unsafe count, audit output, overflow handling |
| Documentation | Rustdoc on public API, doc tests, module docs, README sync | `cargo doc` warnings, doc test count, README freshness |
| Code Quality | Clippy clean, low complexity, no named anti-patterns | Clippy findings, CC distribution |

**Grading Scale:**

| Grade | Finding Threshold | Description |
|-------|------------------|-------------|
| A+ | 0-2 minor findings | Exemplary - industry best practices |
| A | <5 HIGH findings | Excellent - strong practices |
| A- | 5-15 HIGH findings | Very Good - solid practices |
| B+ | 15-25 HIGH findings | Good - acceptable practices |
| B | 25-40 HIGH findings | Satisfactory - needs improvement |
| C+ | 40-60 HIGH findings | Needs Improvement - multiple issues |
| C | 60+ HIGH findings | Significant Issues - major refactoring |
| D | 1+ CRITICAL findings | Major Problems - not production-ready |
| F | Multiple CRITICAL | Critical Issues - complete rewrite |

**Example Assessment:**

| Category | Grade | Evidence |
|----------|-------|----------|
| Error Handling | A | 0 unwraps in lib code, 45 proper `?` propagations, thiserror enums |
| Ownership | A- | 3 unnecessary clones flagged, all lifetimes correct |
| Concurrency | A+ | All locks < 1ms scope, bounded channels, no deadlock paths |
| Unsafe Code | A+ | 0 unsafe blocks in application code |
| Testing | B+ | 72% line coverage, doc tests on public API, no property tests |
| **OVERALL** | **A- (Excellent)** | **8 HIGH, 22 MEDIUM findings** |

---

## Vibe Integration

### Prescan Patterns

| Pattern | Severity | Detection |
|---------|----------|-----------|
| PR-01: `.unwrap()` in library code | HIGH | grep for `.unwrap()` outside `#[cfg(test)]` |
| PR-02: Missing SAFETY comments | CRITICAL | `unsafe` blocks without `// SAFETY:` |
| PR-03: Clippy warnings | HIGH | `cargo clippy` JSON output parsing |
| PR-04: Unformatted code | MEDIUM | `cargo fmt --check` exit code |
| PR-05: Unused dependencies | LOW | `cargo machete` output |

### Semantic Analysis

Deep validation includes:
- Ownership pattern analysis (clone frequency, lifetime correctness)
- Trait design review (ISP compliance, dispatch appropriateness)
- Concurrency safety audit (lock scope, Send/Sync bounds)
- Unsafe code audit (SAFETY comments, scope minimization)

### JIT Loading

**Tier 1 (Fast):** Load `~/.agents/skills/standards/references/rust.md` (5KB)
**Tier 2 (Deep):** Load this document (~20KB) for comprehensive audit
**Override:** Use `.agents/validation/RUST_*.md` if project-specific standards exist

---

## Additional Resources

- [The Rust Programming Language](https://doc.rust-lang.org/book/)
- [Rust API Guidelines](https://rust-lang.github.io/api-guidelines/)
- [Rust Design Patterns](https://rust-unofficial.github.io/patterns/)
- [Clippy Lints](https://rust-lang.github.io/rust-clippy/master/)
- [Rust Performance Book](https://nnethercote.github.io/perf-book/)
- [The Rustonomicon](https://doc.rust-lang.org/nomicon/) (unsafe Rust)

---

**Related:** `rust-patterns.md` for quick reference examples

### shell-standards.md

# Shell Script Standards Catalog - Vibe Canonical Reference

**Version:** 1.0.0
**Last Updated:** 2026-01-21
**Purpose:** Canonical shell scripting standards for vibe skill validation

---

## Table of Contents

1. [Required Patterns](#required-patterns)
2. [Shellcheck Integration](#shellcheck-integration)
3. [Error Handling](#error-handling)
4. [Logging Functions](#logging-functions)
5. [Script Organization](#script-organization)
6. [Security](#security)
7. [Common Patterns](#common-patterns)
8. [Testing](#testing)
9. [Documentation Standards](#documentation-standards)
10. [Code Quality Metrics](#code-quality-metrics)
11. [Anti-Patterns Avoided](#anti-patterns-avoided)
12. [Compliance Assessment](#compliance-assessment)

---

## Required Patterns

### Shebang and Flags

Every shell script MUST start with:

```bash
#!/usr/bin/env bash
set -eEuo pipefail
```

**Flag explanation:**

| Flag | Effect | Failure without |
|------|--------|-----------------|
| `-e` | Exit on error | Silent failures, continued execution |
| `-E` | ERR trap inherited | Traps don't fire in functions |
| `-u` | Exit on undefined | Empty variables cause silent bugs |
| `-o pipefail` | Pipe fails propagate | `false \| true` returns 0 |

### Variable Quoting

Always quote variables to prevent word splitting and globbing:

```bash
# GOOD - Quoted variables, safe defaults
namespace="${NAMESPACE:-default}"
kubectl get pods -n "${namespace}"

# GOOD - Array expansion
files=("file with spaces.txt" "another file.txt")
cat "${files[@]}"

# BAD - Unquoted variables (word splitting, globbing risks)
kubectl get pods -n $namespace
cat $files
```

### Safe Defaults

```bash
# Pattern: ${VAR:-default}
namespace="${NAMESPACE:-default}"
timeout="${TIMEOUT:-300}"
log_level="${LOG_LEVEL:-INFO}"

# Pattern: ${VAR:?error message}
api_key="${API_KEY:?API_KEY must be set}"
```

---

## Shellcheck Integration

### Repository Configuration

Create `.shellcheckrc` at repo root:

```ini
# .shellcheckrc
# Shell variant
shell=bash

# Can't follow non-constant source
disable=SC1090
# Not following sourced files
disable=SC1091
# Consider invoking separately (pipefail handles this)
disable=SC2312
```

### Common Shellcheck Fixes

| Code | Issue | Fix |
|------|-------|-----|
| SC2086 | Word splitting | Quote: `"$var"` |
| SC2164 | cd can fail | `cd /path \|\| exit 1` |
| SC2046 | Word splitting in $() | Quote: `"$(command)"` |
| SC2181 | Checking $? | Use `if command; then` |
| SC2155 | declare/local hides exit | Split: `local x; x=$(cmd)` |
| SC2034 | Unused variable | Remove or use |
| SC2206 | Word splitting in array | `read -ra arr <<< "$var"` |

### Disable Rules Sparingly

```bash
# Only disable when truly necessary
# shellcheck disable=SC2086
# Reason: Word splitting is intentional for flag array
$tool_cmd $flags_array "$input_file"
```

---

## Error Handling

### ERR Trap for Debug Context

```bash
#!/usr/bin/env bash
set -eEuo pipefail

on_error() {
    local exit_code=$?
    local line_no=$1
    echo "ERROR: Script failed on line $line_no with exit code $exit_code" >&2
    echo "Command: ${BASH_COMMAND}" >&2
    exit "$exit_code"
}
trap 'on_error $LINENO' ERR
```

### Exit Code Documentation

Document exit codes in script headers:

```bash
#!/usr/bin/env bash
# ===================================================================
# Script: deploy.sh
# Purpose: Deploy application to Kubernetes cluster
# Usage: ./deploy.sh <namespace> [--dry-run]
#
# Exit Codes:
#   0 - Success
#   1 - Argument error
#   2 - Missing dependency
#   3 - Configuration error
#   4 - Validation failed
#   5 - User cancelled
#   6 - Deployment failed
# ===================================================================

set -eEuo pipefail
```

### Cleanup Pattern

```bash
#!/usr/bin/env bash
set -eEuo pipefail

# Create temp directory
TMPDIR=$(mktemp -d)

cleanup() {
    local exit_code=$?
    rm -rf "$TMPDIR" 2>/dev/null || true
    exit "$exit_code"
}
trap cleanup EXIT

# Your code here - cleanup runs on exit or error
```

### Checking Command Success

```bash
# GOOD - Direct conditional
if kubectl get namespace "$ns" &>/dev/null; then
    echo "Namespace exists"
else
    echo "Creating namespace"
    kubectl create namespace "$ns"
fi

# GOOD - Negation
if ! kubectl get namespace "$ns" &>/dev/null; then
    kubectl create namespace "$ns"
fi

# BAD - Capturing exit code (unnecessary)
kubectl get namespace "$ns" &>/dev/null
result=$?
if [[ $result -eq 0 ]]; then
    ...
fi
```

---

## Logging Functions

### Standard Logging

```bash
# Logging functions
log()  { echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*"; }
warn() { echo "[$(date '+%Y-%m-%d %H:%M:%S')] WARNING: $*" >&2; }
err()  { echo "[$(date '+%Y-%m-%d %H:%M:%S')] ERROR: $*" >&2; }
die()  { err "$*"; exit 1; }

# Debug logging (controlled by variable)
debug() {
    [[ "${DEBUG:-false}" == "true" ]] && echo "[DEBUG] $*" >&2
}

# Usage
log "Processing namespace: ${NAMESPACE}"
warn "Timeout exceeded, retrying..."
die "Required tool 'kubectl' not found"
debug "Variable state: foo=$foo"
```

### Colored Output (Optional)

```bash
# Color codes (check if terminal supports colors)
if [[ -t 1 ]]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    NC='\033[0m'  # No Color
else
    RED=''
    GREEN=''
    YELLOW=''
    NC=''
fi

log_success() { echo -e "${GREEN}[OK]${NC} $*"; }
log_warning() { echo -e "${YELLOW}[WARN]${NC} $*" >&2; }
log_error()   { echo -e "${RED}[ERR]${NC} $*" >&2; }
```

---

## Script Organization

### Full Template

```bash
#!/usr/bin/env bash
# ===================================================================
# Script: <name>.sh
# Purpose: <one-line description>
# Usage: ./<script>.sh [args]
#
# Environment Variables:
#   NAMESPACE - Target namespace (default: default)
#   DRY_RUN   - If "true", don't make changes (default: false)
#
# Exit Codes:
#   0 - Success
#   1 - Argument error
#   2 - Missing dependency
# ===================================================================

set -eEuo pipefail

# -------------------------------------------------------------------
# Configuration
# -------------------------------------------------------------------

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_NAME="$(basename "${BASH_SOURCE[0]}")"

NAMESPACE="${NAMESPACE:-default}"
DRY_RUN="${DRY_RUN:-false}"
TIMEOUT="${TIMEOUT:-300}"

# -------------------------------------------------------------------
# Functions
# -------------------------------------------------------------------

log()  { echo "[$(date '+%H:%M:%S')] $*"; }
warn() { echo "[$(date '+%H:%M:%S')] WARNING: $*" >&2; }
err()  { echo "[$(date '+%H:%M:%S')] ERROR: $*" >&2; }
die()  { err "$*"; exit 1; }

on_error() {
    local exit_code=$?
    err "Script failed on line $1 with exit code $exit_code"
    exit "$exit_code"
}
trap 'on_error $LINENO' ERR

cleanup() {
    rm -rf "${TMPDIR:-}" 2>/dev/null || true
}
trap cleanup EXIT

usage() {
    cat <<EOF
Usage: $SCRIPT_NAME [options] <required-arg>

Options:
    -h, --help      Show this help message
    -n, --namespace Kubernetes namespace (default: $NAMESPACE)
    --dry-run       Don't make changes, just show what would happen

Environment:
    NAMESPACE       Same as --namespace
    DRY_RUN         Same as --dry-run (set to "true")
EOF
}

validate_args() {
    if [[ $# -lt 1 ]]; then
        usage
        exit 1
    fi
}

check_dependencies() {
    local missing=()
    for cmd in kubectl jq; do
        if ! command -v "$cmd" &>/dev/null; then
            missing+=("$cmd")
        fi
    done
    if [[ ${#missing[@]} -gt 0 ]]; then
        die "Missing dependencies: ${missing[*]}"
    fi
}

# -------------------------------------------------------------------
# Main
# -------------------------------------------------------------------

main() {
    local required_arg=""

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case "$1" in
            -h|--help)
                usage
                exit 0
                ;;
            -n|--namespace)
                NAMESPACE="$2"
                shift 2
                ;;
            --dry-run)
                DRY_RUN="true"
                shift
                ;;
            -*)
                die "Unknown option: $1"
                ;;
            *)
                required_arg="$1"
                shift
                ;;
        esac
    done

    validate_args "${required_arg:-}"
    check_dependencies

    log "Starting with namespace: $NAMESPACE"
    [[ "$DRY_RUN" == "true" ]] && log "DRY RUN MODE - no changes will be made"

    # Main logic here
}

# Setup
TMPDIR=$(mktemp -d)

# Run
main "$@"
```

---

## Security

### Secret Handling

**Never pass secrets as CLI arguments** - they're visible in `ps aux`:

```bash
# BAD - Secrets visible in process list
kubectl create secret generic my-secret --from-literal=token="$TOKEN"

# GOOD - Pass via stdin
kubectl create secret generic my-secret --from-literal=token=- <<< "$TOKEN"

# GOOD - Use file-based approach
echo "$SECRET" > "$TMPDIR/secret"
chmod 600 "$TMPDIR/secret"
kubectl create secret generic my-secret --from-file=token="$TMPDIR/secret"

# GOOD - Environment variable (some tools support this)
export TOKEN
some_tool --token-env=TOKEN
```

### Input Validation

#### Kubernetes Resource Names (RFC 1123)

```bash
validate_namespace() {
    local ns="$1"
    # RFC 1123: lowercase alphanumeric, hyphens, max 63 chars
    if [[ ! "$ns" =~ ^[a-z0-9][a-z0-9-]{0,61}[a-z0-9]$ ]] && \
       [[ ! "$ns" =~ ^[a-z0-9]$ ]]; then
        die "Invalid namespace format: $ns (must be RFC 1123 label)"
    fi
}
```

#### Path Traversal Prevention

```bash
validate_path() {
    local path="$1"
    # Block path traversal
    case "$path" in
        *..*)
            die "Path traversal detected: $path"
            ;;
    esac
    # Optionally ensure path is under allowed directory
    local resolved
    resolved=$(realpath -m "$path" 2>/dev/null) || die "Invalid path: $path"
    if [[ "$resolved" != "$ALLOWED_DIR"/* ]]; then
        die "Path outside allowed directory: $path"
    fi
}
```

### Sed Injection Prevention

```bash
# User input in sed can have special meaning
# BAD - Injection possible
NAME="test&id"  # & inserts matched text
sed "s/{{NAME}}/$NAME/g" template.txt  # & causes issues

# GOOD - Escape special characters
escape_sed_replacement() {
    printf '%s' "$1" | sed -e 's/[&/\]/\\&/g'
}

escaped_name=$(escape_sed_replacement "$NAME")
sed "s/{{NAME}}/$escaped_name/g" template.txt
```

### JSON Construction

```bash
# BAD - String interpolation (injection risk, quoting issues)
json="{\"name\": \"$NAME\", \"value\": \"$VALUE\"}"

# GOOD - Use jq for proper escaping
json=$(jq -n --arg name "$NAME" --arg value "$VALUE" \
    '{name: $name, value: $value}')

# GOOD - Building complex JSON
json=$(jq -n \
    --arg name "$NAME" \
    --arg ns "$NAMESPACE" \
    --argjson replicas "$REPLICAS" \
    '{
        metadata: {name: $name, namespace: $ns},
        spec: {replicas: $replicas}
    }')
```

---

## Common Patterns

### Polling with Timeout

> **SECURITY WARNING:** The pattern below uses `eval` which can lead to command injection
> if `condition_cmd` contains unsanitized user input. **Never pass untrusted input to this function.**
> For safer alternatives, use bash functions instead of string evaluation:
> ```bash
> # SAFER: Pass a function name instead of a command string
> wait_for_condition 300 10 check_pod_running
> ```

```bash
wait_for_condition() {
    local timeout=${1:-300}
    local interval=${2:-10}
    local condition_cmd="$3"

    # WARNING: eval is dangerous with untrusted input - see security note above
    local elapsed=0
    while ! eval "$condition_cmd" &>/dev/null; do
        if [[ $elapsed -ge $timeout ]]; then
            err "Timeout waiting for condition after ${timeout}s"
            return 1
        fi
        log "Waiting... (${elapsed}s/${timeout}s)"
        sleep "$interval"
        elapsed=$((elapsed + interval))
    done
    return 0
}

# Usage
wait_for_condition 300 10 \
    "kubectl get pod my-pod -o jsonpath='{.status.phase}' | grep -q Running"
```

### Parallel Execution

```bash
run_parallel() {
    local pids=()
    local failures=()

    for item in "$@"; do
        process_item "$item" &
        pids+=($!)
    done

    for pid in "${pids[@]}"; do
        if ! wait "$pid"; then
            failures+=("$pid")
        fi
    done

    if [[ ${#failures[@]} -gt 0 ]]; then
        err "Failed jobs: ${#failures[@]}"
        return 1
    fi
}
```

### Kubernetes Resource Checks

```bash
resource_exists() {
    local kind="$1"
    local name="$2"
    local ns="${3:-}"

    local ns_flag=""
    [[ -n "$ns" ]] && ns_flag="-n $ns"

    # shellcheck disable=SC2086
    kubectl get "$kind" "$name" $ns_flag &>/dev/null
}

# Usage
if resource_exists deployment my-app my-namespace; then
    log "Deployment exists"
fi
```

### Retry Pattern

```bash
retry() {
    local max_attempts=${1:-3}
    local delay=${2:-5}
    shift 2
    local cmd=("$@")

    local attempt=1
    while [[ $attempt -le $max_attempts ]]; do
        if "${cmd[@]}"; then
            return 0
        fi
        warn "Attempt $attempt/$max_attempts failed, retrying in ${delay}s..."
        sleep "$delay"
        attempt=$((attempt + 1))
    done

    err "All $max_attempts attempts failed"
    return 1
}

# Usage
retry 3 5 kubectl apply -f manifest.yaml
```

---

## Testing

### Manual Testing

```bash
# Test with shellcheck
shellcheck ./script.sh

# Test with bash
bash ./script.sh --help

# Test with debug output
bash -x ./script.sh

# Test with verbose mode
bash -v ./script.sh
```

### BATS Framework

For complex scripts, use [BATS](https://github.com/bats-core/bats-core):

```bash
# test/test_script.bats
#!/usr/bin/env bats

setup() {
    load 'test_helper/bats-support/load'
    load 'test_helper/bats-assert/load'
    SCRIPT="$BATS_TEST_DIRNAME/../script.sh"
}

@test "script requires argument" {
    run bash "$SCRIPT"
    assert_failure 1
    assert_output --partial "Usage:"
}

@test "script validates namespace format" {
    run bash "$SCRIPT" --namespace "INVALID_NS"
    assert_failure
    assert_output --partial "Invalid namespace"
}

@test "script succeeds with valid input" {
    run bash "$SCRIPT" valid-namespace
    assert_success
}
```

---

## Documentation Standards

### Header Comments

Every script MUST have a header block describing purpose, usage, and exit codes:

```bash
#!/usr/bin/env bash
# ===================================================================
# Script: deploy.sh
# Purpose: Deploy application to target Kubernetes cluster
# Usage: ./deploy.sh [options] <namespace>
#
# Options:
#   -h, --help      Show this help message
#   -n, --namespace Target namespace (default: default)
#   --dry-run       Preview changes without applying
#
# Environment Variables:
#   NAMESPACE - Target namespace (default: default)
#   DRY_RUN   - If "true", don't make changes
#
# Exit Codes:
#   0 - Success
#   1 - Argument error
#   2 - Missing dependency
# ===================================================================
```

### Inline Help (--help / -h)

Every user-facing script MUST support `--help` and `-h` flags:

```bash
usage() {
    cat <<EOF
Usage: ${SCRIPT_NAME} [options] <required-arg>

Options:
    -h, --help      Show this help message
    -v, --verbose   Enable verbose output
    --dry-run       Preview mode

Examples:
    ${SCRIPT_NAME} my-namespace
    ${SCRIPT_NAME} --dry-run my-namespace
EOF
}
```

**Requirements:**
- Print to stdout (not stderr) so output is pipeable
- Exit 0 on `--help`, exit 1 on missing/invalid args
- Include at least one usage example

### Function Documentation

Document non-trivial functions with a comment above the declaration:

```bash
# Wait for a Kubernetes resource to reach the desired state.
# Arguments:
#   $1 - Resource kind (e.g., deployment, pod)
#   $2 - Resource name
#   $3 - Desired condition (e.g., Available, Ready)
#   $4 - Timeout in seconds (default: 300)
# Returns: 0 on success, 1 on timeout
wait_for_resource() {
    local kind="$1" name="$2" condition="$3" timeout="${4:-300}"
    # ...
}
```

### ALWAYS / NEVER Rules

| Rule | Rationale |
|------|-----------|
| **ALWAYS** include a header block with Purpose, Usage, Exit Codes | First thing a reader sees — orients them |
| **ALWAYS** support `--help` / `-h` in user-facing scripts | Discoverability — no need to read source |
| **ALWAYS** document functions with 3+ parameters | Prevents misuse of positional args |
| **ALWAYS** document non-obvious exit codes (beyond 0/1) | Callers need to handle specific failures |
| **NEVER** omit the shebang line (`#!/usr/bin/env bash`) | Portability — don't assume `/bin/bash` |
| **NEVER** put usage info only in comments (not in `--help`) | Users shouldn't need to read source |

---

## Code Quality Metrics

> See `common-standards.md` for universal coverage targets and testing principles.

### Validation Commands

```bash
# Shellcheck all scripts
shellcheck scripts/*.sh
# Output: "X issues" → Count by severity

# Check set flags
grep -r "^set -" scripts/ | grep -c "eEuo pipefail"
# Compare to total script count

# Count unquoted variables (SC2086)
shellcheck scripts/*.sh -f json | jq '[.[] | select(.code == 2086)] | length'

# Check ERR trap presence
grep -r "trap.*ERR" scripts/ | wc -l

# Check exit code documentation
grep -r "# Exit Codes:" scripts/ | wc -l

# Security: Check for secrets in args
grep -rE "\-\-(password|token|secret)=" scripts/
# Should return nothing
```

---

## Anti-Patterns Avoided

> See `common-standards.md` for universal anti-patterns across all languages.

### No Parsing ls Output

```bash
# Bad
for f in $(ls); do
    echo "$f"
done

# Good
for f in *; do
    [[ -e "$f" ]] || continue
    echo "$f"
done

# Good - find with null separator
while IFS= read -r -d '' f; do
    echo "$f"
done < <(find . -type f -print0)
```

### No Useless Cat

```bash
# Bad
cat file.txt | grep pattern

# Good
grep pattern file.txt
```

### No Backticks

```bash
# Bad
result=`command`

# Good
result=$(command)

# Good - nested
result=$(echo $(date))
```

---

## Compliance Assessment

**Use letter grades + evidence, NOT numeric scores.**

### Assessment Categories

| Category | Evidence Required |
|----------|------------------|
| **Safety** | set flags count, SC2086 violations |
| **Code Quality** | shellcheck total violations, function count |
| **Security** | Secrets in CLI, input validation |
| **Error Handling** | ERR trap, exit code docs |
| **Logging** | Log function usage |

### Grading Scale

| Grade | Criteria |
|-------|----------|
| A+ | 0 shellcheck errors, set flags, ERR trap, 0 security issues |
| A | <5 shellcheck warnings, set flags, ERR trap, quoted vars |
| A- | <15 shellcheck warnings, set flags, mostly quoted |
| B+ | <30 shellcheck warnings, set flags present |
| B | <50 shellcheck warnings, some flags |
| C | Significant safety issues |
| D | Not production-ready |
| F | Critical issues |

### Example Assessment

```markdown
## Shell Script Standards Compliance

**Target:** scripts/
**Date:** 2026-01-21

| Category | Grade | Evidence |
|----------|-------|----------|
| Safety | A+ | 12/12 have set -eEuo pipefail, 0 SC2086 |
| Code Quality | A- | 8 shellcheck warnings (SC2312), 15 functions |
| Security | A | 0 secrets in CLI, 8/8 inputs validated |
| Error Handling | A | 12/12 ERR trap, 11/12 exit codes |
| **OVERALL** | **A** | **5 MEDIUM findings** |
```

---

## Vibe Integration

### Prescan Patterns

| Pattern | Severity | Detection |
|---------|----------|-----------|
| P05: Missing set flags | HIGH | No `set -eEuo pipefail` |
| P06: Unquoted variables | MEDIUM | SC2086 violations |
| P09: Secrets in CLI | CRITICAL | `--password=` patterns |

### JIT Loading

**Tier 1 (Fast):** Load `~/.agents/skills/standards/references/shell.md` (5KB)
**Tier 2 (Deep):** Load this document (18KB) for comprehensive audit

---

## Additional Resources

- [Bash Manual](https://www.gnu.org/software/bash/manual/)
- [ShellCheck Wiki](https://www.shellcheck.net/wiki/)
- [Google Shell Style Guide](https://google.github.io/styleguide/shellguide.html)
- [BATS Testing Framework](https://github.com/bats-core/bats-core)

---

**Related:** Quick reference in Tier 1 `shell.md`

### typescript-standards.md

# TypeScript Standards Catalog - Vibe Canonical Reference

**Version:** 1.0.0
**Last Updated:** 2026-01-21
**Purpose:** Canonical TypeScript standards for vibe skill validation

---

## Table of Contents

1. [Strict Configuration](#strict-configuration)
2. [ESLint Configuration](#eslint-configuration)
3. [Type System Patterns](#type-system-patterns)
4. [Generic Constraints](#generic-constraints)
5. [Utility Types](#utility-types)
6. [Conditional Types](#conditional-types)
7. [Error Handling](#error-handling)
8. [Module Template](#module-template)
9. [Code Quality Metrics](#code-quality-metrics)
10. [Testing Patterns](#testing-patterns)
    - [Test Frameworks](#test-frameworks)
    - [Test Organization](#test-organization)
    - [React Testing Library Patterns](#react-testing-library-patterns)
    - [MSW (Mock Service Worker)](#msw-mock-service-worker-for-api-mocking)
    - [Mocking](#mocking)
    - [Async Testing](#async-testing)
    - [Type-safe Testing](#type-safe-testing)
    - [Snapshot Testing](#snapshot-testing)
    - [Coverage Expectations](#coverage-expectations)
11. [Anti-Patterns Avoided](#anti-patterns-avoided)
12. [Compliance Assessment](#compliance-assessment)

---

## Strict Configuration

### Full tsconfig.json

Every TypeScript project MUST use strict mode:

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "NodeNext",
    "moduleResolution": "NodeNext",
    "lib": ["ES2022"],
    "outDir": "./dist",
    "rootDir": "./src",

    "strict": true,
    "noUncheckedIndexedAccess": true,
    "noImplicitReturns": true,
    "noFallthroughCasesInSwitch": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "exactOptionalPropertyTypes": true,

    "declaration": true,
    "declarationMap": true,
    "sourceMap": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true
  },
  "include": ["src/**/*"],
  "exclude": ["node_modules", "dist"]
}
```

### Why Strict Matters

| Option | Effect |
|--------|--------|
| `strict: true` | Enables all strict type-checking options |
| `noUncheckedIndexedAccess` | Adds `undefined` to index signatures |
| `exactOptionalPropertyTypes` | Distinguishes `undefined` from missing |
| `noImplicitReturns` | All code paths must return |
| `noFallthroughCasesInSwitch` | Prevents accidental case fallthrough |

---

## ESLint Configuration

### eslint.config.js (Flat Config)

```javascript
import eslint from '@eslint/js';
import tseslint from 'typescript-eslint';

export default tseslint.config(
  eslint.configs.recommended,
  ...tseslint.configs.strictTypeChecked,
  ...tseslint.configs.stylisticTypeChecked,
  {
    languageOptions: {
      parserOptions: {
        project: true,
        tsconfigRootDir: import.meta.dirname,
      },
    },
    rules: {
      '@typescript-eslint/no-unused-vars': ['error', { argsIgnorePattern: '^_' }],
      '@typescript-eslint/explicit-function-return-type': 'error',
      '@typescript-eslint/no-explicit-any': 'error',
      '@typescript-eslint/prefer-nullish-coalescing': 'error',
      '@typescript-eslint/prefer-optional-chain': 'error',
      '@typescript-eslint/no-floating-promises': 'error',
      '@typescript-eslint/await-thenable': 'error',
    },
  },
  {
    ignores: ['dist/', 'node_modules/', '*.js'],
  }
);
```

### Usage

```bash
# Lint check
npx eslint . --ext .ts,.tsx

# Fix auto-fixable issues
npx eslint . --ext .ts,.tsx --fix

# Type check only (no emit)
npx tsc --noEmit
```

---

## Type System Patterns

### Prefer Type Inference

Let TypeScript infer types when obvious:

```typescript
// Good - inference is clear
const users = ['alice', 'bob'];
const count = users.length;

// Good - explicit when non-obvious or API boundary
function getUser(id: string): User | undefined {
  return userMap.get(id);
}

// Bad - redundant annotation
const name: string = 'alice';
```

### Discriminated Unions

Use discriminated unions for state modeling:

```typescript
// Good - exhaustive pattern matching
type Result<T, E> =
  | { status: 'success'; data: T }
  | { status: 'error'; error: E };

function handleResult<T, E>(result: Result<T, E>): void {
  switch (result.status) {
    case 'success':
      console.log(result.data);
      break;
    case 'error':
      console.error(result.error);
      break;
    // TypeScript enforces exhaustiveness
  }
}
```

### Const Assertions

Use `as const` for literal types:

```typescript
// Good - preserves literal types
const CONFIG = {
  apiVersion: 'v1',
  retries: 3,
  endpoints: ['primary', 'fallback'],
} as const;

// Type: { readonly apiVersion: "v1"; readonly retries: 3; ... }
```

### Branded Types

Use branded types for type-safe IDs:

```typescript
type UserId = string & { readonly __brand: 'UserId' };
type OrderId = string & { readonly __brand: 'OrderId' };

function createUserId(id: string): UserId {
  return id as UserId;
}

function getUser(id: UserId): User { ... }
function getOrder(id: OrderId): Order { ... }

// Type error: can't pass UserId where OrderId expected
const user = getUser(createUserId('123'));
const order = getOrder(createUserId('123')); // Error!
```

---

## Generic Constraints

### Constrained Generics

Always constrain generics when possible:

```typescript
// Good - constrained generic
function getProperty<T, K extends keyof T>(obj: T, key: K): T[K] {
  return obj[key];
}

// Good - multiple constraints
function merge<T extends object, U extends object>(a: T, b: U): T & U {
  return { ...a, ...b };
}

// Bad - unconstrained (allows any)
function unsafe<T>(value: T): T {
  return value;
}
```

### Generic Defaults

Provide defaults for optional type parameters:

```typescript
interface ApiResponse<T = unknown, E = Error> {
  data?: T;
  error?: E;
  status: number;
}

// Uses defaults
const response: ApiResponse = { status: 200 };

// Override defaults
const typed: ApiResponse<User, ApiError> = { status: 200 };
```

### Generic Inference

Let TypeScript infer generic types when possible:

```typescript
// Good - infers T from argument
function identity<T>(value: T): T {
  return value;
}

const str = identity('hello'); // T inferred as string
const num = identity(42);      // T inferred as number

// Bad - unnecessary explicit type
const str2 = identity<string>('hello'); // Redundant
```

---

## Utility Types

### Built-in Utilities

Use built-in utility types over manual definitions:

```typescript
// Partial - all properties optional
type PartialUser = Partial<User>;

// Required - all properties required
type RequiredConfig = Required<Config>;

// Pick - select properties
type UserPreview = Pick<User, 'id' | 'name'>;

// Omit - exclude properties
type UserWithoutPassword = Omit<User, 'password'>;

// Record - typed object
type UserMap = Record<string, User>;

// Extract/Exclude - union manipulation
type StringOrNumber = Extract<string | number | boolean, string | number>;
```

### Custom Type Helpers

Create reusable type utilities:

```typescript
// Deep partial
type DeepPartial<T> = {
  [P in keyof T]?: T[P] extends object ? DeepPartial<T[P]> : T[P];
};

// Non-nullable object values
type NonNullableValues<T> = {
  [K in keyof T]: NonNullable<T[K]>;
};

// Extract function return types from object
type ReturnTypes<T extends Record<string, (...args: never[]) => unknown>> = {
  [K in keyof T]: ReturnType<T[K]>;
};

// Make specific keys required
type WithRequired<T, K extends keyof T> = T & Required<Pick<T, K>>;
```

---

## Conditional Types

### Type-Level Logic

Use conditional types for dynamic typing:

```typescript
// Infer array element type
type ElementOf<T> = T extends readonly (infer E)[] ? E : never;

// Flatten promise type
type Awaited<T> = T extends Promise<infer U> ? Awaited<U> : T;

// Function parameter extraction
type FirstParam<T> = T extends (first: infer P, ...args: never[]) => unknown
  ? P
  : never;

// Conditional return type
type ApiResult<T> = T extends 'user'
  ? User
  : T extends 'order'
  ? Order
  : never;
```

### Template Literal Types

Use template literals for string manipulation:

```typescript
// Event handler naming
type EventName = 'click' | 'change' | 'submit';
type HandlerName = `on${Capitalize<EventName>}`;
// Result: "onClick" | "onChange" | "onSubmit"

// Path building
type ApiPath<T extends string> = `/api/v1/${T}`;
type UserPath = ApiPath<'users'>; // "/api/v1/users"

// Property getters/setters
type Getters<T> = {
  [K in keyof T as `get${Capitalize<string & K>}`]: () => T[K];
};
```

---

## Error Handling

### Result Pattern

Prefer explicit error handling over exceptions:

```typescript
type Result<T, E = Error> =
  | { ok: true; value: T }
  | { ok: false; error: E };

function parseJson<T>(json: string): Result<T, SyntaxError> {
  try {
    return { ok: true, value: JSON.parse(json) as T };
  } catch (e) {
    return { ok: false, error: e as SyntaxError };
  }
}

// Usage
const result = parseJson<User>(input);
if (result.ok) {
  console.log(result.value.name);
} else {
  console.error(result.error.message);
}
```

### Type Guards

Use type guards for runtime type narrowing:

```typescript
// User-defined type guard
function isUser(value: unknown): value is User {
  return (
    typeof value === 'object' &&
    value !== null &&
    'id' in value &&
    'name' in value
  );
}

// Assertion function
function assertUser(value: unknown): asserts value is User {
  if (!isUser(value)) {
    throw new Error('Invalid user');
  }
}

// Usage
function processData(data: unknown): void {
  if (isUser(data)) {
    // data is User here
    console.log(data.name);
  }

  // Or with assertion
  assertUser(data);
  // data is User from here on
  console.log(data.id);
}
```

### Error Classes

Create typed error classes:

```typescript
class AppError extends Error {
  constructor(
    message: string,
    public readonly code: string,
    public readonly statusCode: number = 500,
  ) {
    super(message);
    this.name = 'AppError';
  }
}

class ValidationError extends AppError {
  constructor(
    message: string,
    public readonly field: string,
  ) {
    super(message, 'VALIDATION_ERROR', 400);
    this.name = 'ValidationError';
  }
}

// Type guard for error handling
function isAppError(error: unknown): error is AppError {
  return error instanceof AppError;
}
```

---

## Module Template

Standard template for TypeScript modules:

```typescript
/**
 * Module description.
 * @module module-name
 */

// Types first
export interface Config {
  readonly apiUrl: string;
  readonly timeout: number;
}

export type Handler<T> = (data: T) => Promise<void>;

// Type guards
export function isConfig(value: unknown): value is Config {
  return (
    typeof value === 'object' &&
    value !== null &&
    'apiUrl' in value &&
    'timeout' in value
  );
}

// Constants
const DEFAULT_TIMEOUT = 5000;

// Private helpers (not exported)
function validateConfig(config: Config): void {
  if (!config.apiUrl) {
    throw new Error('apiUrl is required');
  }
}

// Public API
export function createClient(config: Config): Client {
  validateConfig(config);
  return new Client(config);
}

export class Client {
  readonly #config: Config;

  constructor(config: Config) {
    this.#config = config;
  }

  async fetch<T>(path: string): Promise<T> {
    const response = await fetch(`${this.#config.apiUrl}${path}`);
    return response.json() as Promise<T>;
  }
}
```

---

## Code Quality Metrics

> See `common-standards.md` for universal coverage targets and testing principles.

### Type Coverage Metrics

| Metric | Target | Validation |
|--------|--------|------------|
| tsc errors | 0 | `tsc --noEmit` |
| any types | 0 | `grep -r ": any"` |
| Explicit returns | 100% on exports | `grep "^export function"` |
| Type-only imports | 100% | Check `import type` usage |

### Validation Commands

```bash
# Type check (no emit)
tsc --noEmit
# Output: "Found X errors" → Count these

# ESLint violations
npx eslint . --ext .ts,.tsx
# Output: "X problems (Y errors, Z warnings)" → Report all

# Count any types
grep -r ": any" src/ | wc -l
# Report: "5 any types found"

# Count explicit return types on exports
grep -r "^export function" src/ | grep -c ": .* {"
# Compare to total export function count

# Type-only imports check
grep -r "^import {" src/ | grep -vc "import type"
# Report: "12 value imports (should be type-only)"
```

---

## Testing Patterns

### Test Frameworks

#### Vitest (Preferred)

Vitest is the recommended test runner for TypeScript projects. It shares Vite's config and transform pipeline, supports ESM natively, and runs significantly faster than Jest for TypeScript codebases.

```typescript
// vitest.config.ts
import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    globals: true,
    environment: 'jsdom',          // For React; use 'node' for backend
    setupFiles: ['./src/test/setup.ts'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'lcov'],
      exclude: ['**/*.d.ts', '**/*.test.ts', '**/test/**'],
    },
    typecheck: {
      enabled: true,               // Run type-level tests via expect-type
    },
  },
});
```

#### Jest 29+ with ts-jest

When Vitest is not an option (legacy codebase, specific CI constraints):

```typescript
// jest.config.ts
import type { Config } from 'jest';

const config: Config = {
  preset: 'ts-jest',
  testEnvironment: 'node',
  roots: ['<rootDir>/src'],
  moduleNameMapper: {
    '^@/(.*)$': '<rootDir>/src/$1',
  },
  collectCoverageFrom: [
    'src/**/*.ts',
    '!src/**/*.d.ts',
    '!src/**/*.test.ts',
  ],
};

export default config;
```

| Setting | Recommendation |
|---------|---------------|
| Runner | Vitest (preferred) or Jest 29+ with ts-jest |
| Environment | `jsdom` for UI, `node` for backend/CLI |
| Globals | `true` — avoids `import { describe, it }` boilerplate |
| Coverage provider | `v8` (fast) or `istanbul` (precise) |
| Transform | Vitest uses esbuild (fast); Jest uses ts-jest or `@swc/jest` |

### Test Organization

```typescript
describe('UserService', () => {
  // Group by method
  describe('createUser', () => {
    it('creates user with valid data', async () => { /* ... */ });
    it('throws ValidationError for duplicate email', async () => { /* ... */ });
    it('hashes password before storing', async () => { /* ... */ });
  });

  describe('deleteUser', () => {
    it('soft-deletes user by setting deletedAt', async () => { /* ... */ });
    it('throws NotFoundError for unknown id', async () => { /* ... */ });
  });
});
```

**Nesting guidelines:**

- Top-level `describe` = class or module name
- Second-level `describe` = method or function name
- `it` blocks = single behavior, stated as expected outcome
- Limit nesting to 3 levels maximum — deeper nesting signals the unit under test is too complex

**File naming and placement:**

| Convention | Example |
|-----------|---------|
| Co-located tests (preferred) | `user-service.test.ts` next to `user-service.ts` |
| Test directory | `__tests__/user-service.test.ts` (when co-location is impractical) |
| Test utilities | `src/test/setup.ts` for global setup, `src/test/factories.ts` for test data |
| Integration tests | `tests/integration/` at project root |
| E2E tests | `tests/e2e/` at project root |
| Describe blocks | Class/module name: `describe('UserService', ...)` |
| Test names | Behavior: `it('throws ValidationError for duplicate email')` |

### React Testing Library Patterns

Test components by user behavior, not implementation:

```typescript
// Good - tests user-visible behavior
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

test('submits form with valid data', async () => {
  const onSubmit = vi.fn();
  const user = userEvent.setup();

  render(<LoginForm onSubmit={onSubmit} />);

  await user.type(screen.getByLabelText('Email'), 'alice@example.com');
  await user.type(screen.getByLabelText('Password'), 'secret123');
  await user.click(screen.getByRole('button', { name: /sign in/i }));

  expect(onSubmit).toHaveBeenCalledWith({
    email: 'alice@example.com',
    password: 'secret123',
  });
});

// Bad - tests implementation details
test('sets state on input change', () => {
  const { container } = render(<LoginForm />);
  const input = container.querySelector('input[name="email"]')!;
  fireEvent.change(input, { target: { value: 'alice@example.com' } });
  // Brittle: relies on DOM structure and internal state
});
```

**Query Priority (prefer top to bottom):**

| Priority | Query | When |
|----------|-------|------|
| 1 | `getByRole` | Interactive elements (buttons, inputs, headings) |
| 2 | `getByLabelText` | Form fields |
| 3 | `getByText` | Non-interactive text content |
| 4 | `getByTestId` | Last resort — no accessible selector available |

### MSW (Mock Service Worker) for API Mocking

Mock API calls at the network level, not the implementation level:

```typescript
import { http, HttpResponse } from 'msw';
import { setupServer } from 'msw/node';

const handlers = [
  http.get('/api/users/:id', ({ params }) => {
    return HttpResponse.json({
      id: params.id,
      name: 'Alice',
      email: 'alice@example.com',
    });
  }),

  http.post('/api/users', async ({ request }) => {
    const body = await request.json();
    return HttpResponse.json(body, { status: 201 });
  }),
];

const server = setupServer(...handlers);

beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

// Override for specific test
test('handles server error', async () => {
  server.use(
    http.get('/api/users/:id', () => {
      return HttpResponse.json({ message: 'Internal error' }, { status: 500 });
    }),
  );
  // ... test error handling
});
```

### Async Testing Patterns

Use `waitFor` and async queries for asynchronous UI updates:

```typescript
// Good - waits for async state updates
import { render, screen, waitFor } from '@testing-library/react';

test('loads and displays user data', async () => {
  render(<UserProfile userId="123" />);

  // findBy* waits for element to appear (combines getBy + waitFor)
  const name = await screen.findByText('Alice');
  expect(name).toBeInTheDocument();

  // waitFor for assertions on async state
  await waitFor(() => {
    expect(screen.getByRole('status')).toHaveTextContent('Active');
  });
});

// Bad - manual timers and arbitrary delays
test('loads data', async () => {
  render(<UserProfile userId="123" />);
  await new Promise((r) => setTimeout(r, 1000)); // Flaky!
  expect(screen.getByText('Alice')).toBeInTheDocument();
});
```

### Snapshot Testing

| Use Snapshots For | Avoid Snapshots For |
|-------------------|---------------------|
| Serialized data structures (API responses, configs) | Full component trees (too brittle) |
| Error message formatting | Styled components (CSS changes break snapshots) |
| CLI output strings | Large objects (unreadable diffs) |

```typescript
// Good - small, focused snapshot
test('formats error response', () => {
  const error = formatApiError(404, 'User not found');
  expect(error).toMatchInlineSnapshot(`
    {
      "code": 404,
      "message": "User not found",
      "type": "NOT_FOUND",
    }
  `);
});

// Bad - entire component tree snapshot
test('renders dashboard', () => {
  const { container } = render(<Dashboard />);
  expect(container).toMatchSnapshot(); // 500+ line snapshot nobody reviews
});
```

### Mocking

#### Function Mocks (`vi.fn` / `jest.fn`)

Use function mocks to verify interactions and control return values:

```typescript
// Basic function mock
const onSave = vi.fn();
render(<Form onSave={onSave} />);
await user.click(screen.getByRole('button', { name: /save/i }));
expect(onSave).toHaveBeenCalledOnce();
expect(onSave).toHaveBeenCalledWith({ name: 'Alice', email: 'alice@example.com' });

// Mock with return value
const fetchUser = vi.fn().mockResolvedValue({ id: '1', name: 'Alice' });

// Mock with implementation
const hash = vi.fn((input: string) => `hashed_${input}`);

// Spy on existing method (preserves original by default)
const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
// ... test code ...
expect(consoleSpy).toHaveBeenCalledWith('Connection failed');
consoleSpy.mockRestore();
```

#### Module Mocks

Mock entire modules when you need to replace dependencies:

```typescript
// Vitest — mock a module
vi.mock('./database', () => ({
  getConnection: vi.fn().mockReturnValue({
    query: vi.fn().mockResolvedValue([]),
    close: vi.fn(),
  }),
}));

// Jest — mock a module
jest.mock('./database', () => ({
  getConnection: jest.fn().mockReturnValue({
    query: jest.fn().mockResolvedValue([]),
    close: jest.fn(),
  }),
}));

// Import after mock declaration — the import gets the mocked version
import { getConnection } from './database';
```

#### Manual Mocks (`__mocks__/`)

Use manual mocks for complex dependencies shared across many test files:

```
src/
  services/
    __mocks__/
      email-service.ts    # Manual mock — auto-used when vi.mock('./email-service') is called
    email-service.ts      # Real implementation
    email-service.test.ts
```

```typescript
// src/services/__mocks__/email-service.ts
export const sendEmail = vi.fn().mockResolvedValue({ messageId: 'mock-id' });
export const validateAddress = vi.fn().mockReturnValue(true);

// In test file — just declare the mock, implementation comes from __mocks__/
vi.mock('./email-service');
```

**When to use each mocking approach:**

| Approach | When |
|----------|------|
| `vi.fn()` / `jest.fn()` | Callbacks, event handlers, simple dependency injection |
| `vi.spyOn()` / `jest.spyOn()` | Observing calls without replacing behavior (or with `mockImplementation`) |
| `vi.mock()` / `jest.mock()` | Replacing an imported module for a single test file |
| Manual mocks (`__mocks__/`) | Shared mock used by 3+ test files |
| MSW | HTTP/API calls — always prefer over mocking `fetch`/`axios` directly |

### Async Testing

#### Promises and Async/Await

```typescript
// Good — async/await with proper assertion
it('fetches user by id', async () => {
  const user = await userService.getById('123');
  expect(user).toEqual({ id: '123', name: 'Alice' });
});

// Good — testing rejected promises
it('throws NotFoundError for missing user', async () => {
  await expect(userService.getById('nonexistent')).rejects.toThrow(NotFoundError);
});

// Good — testing promise resolution value
it('resolves with created user', async () => {
  await expect(userService.create({ name: 'Bob' })).resolves.toMatchObject({
    name: 'Bob',
    id: expect.any(String),
  });
});
```

#### Fake Timers

```typescript
// Vitest
beforeEach(() => {
  vi.useFakeTimers();
});

afterEach(() => {
  vi.useRealTimers();
});

it('retries after delay', async () => {
  const fetchData = vi.fn()
    .mockRejectedValueOnce(new Error('timeout'))
    .mockResolvedValue({ data: 'ok' });

  const promise = retryWithDelay(fetchData, { delay: 1000, retries: 2 });

  // Advance past the retry delay
  await vi.advanceTimersByTimeAsync(1000);

  const result = await promise;
  expect(result).toEqual({ data: 'ok' });
  expect(fetchData).toHaveBeenCalledTimes(2);
});

it('debounces input handler', async () => {
  const handler = vi.fn();
  const debounced = debounce(handler, 300);

  debounced('a');
  debounced('ab');
  debounced('abc');

  expect(handler).not.toHaveBeenCalled();

  await vi.advanceTimersByTimeAsync(300);

  expect(handler).toHaveBeenCalledOnce();
  expect(handler).toHaveBeenCalledWith('abc');
});
```

#### React `act()` and Async State Updates

```typescript
import { act, render, screen } from '@testing-library/react';

// act() is needed when triggering state updates outside of RTL helpers
it('updates count on external event', async () => {
  const eventBus = new EventEmitter();
  render(<Counter eventBus={eventBus} />);

  await act(async () => {
    eventBus.emit('increment');
  });

  expect(screen.getByText('Count: 1')).toBeInTheDocument();
});

// RTL's userEvent and findBy* wrap act() automatically — prefer those
// Only use act() directly when dealing with non-RTL async triggers
```

### Type-safe Testing

#### Typed Mocks

```typescript
// Type-safe mock function
const onSubmit = vi.fn<[FormData], Promise<void>>();

// Type-safe mock of an interface
interface UserRepository {
  findById(id: string): Promise<User | null>;
  save(user: User): Promise<User>;
  delete(id: string): Promise<void>;
}

function createMockRepository(): { [K in keyof UserRepository]: ReturnType<typeof vi.fn> } & UserRepository {
  return {
    findById: vi.fn<[string], Promise<User | null>>().mockResolvedValue(null),
    save: vi.fn<[User], Promise<User>>().mockImplementation(async (user) => user),
    delete: vi.fn<[string], Promise<void>>().mockResolvedValue(undefined),
  };
}

it('returns null for unknown user', async () => {
  const repo = createMockRepository();
  const service = new UserService(repo);

  const result = await service.getById('unknown');
  expect(result).toBeNull();
  expect(repo.findById).toHaveBeenCalledWith('unknown');
});
```

#### Assertion Helpers and Custom Matchers

```typescript
// Custom matcher for domain-specific assertions
expect.extend({
  toBeValidEmail(received: string) {
    const pass = /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(received);
    return {
      pass,
      message: () => `expected ${received} ${pass ? 'not ' : ''}to be a valid email`,
    };
  },
});

// Declare the matcher type
declare module 'vitest' {
  interface Assertion<T> {
    toBeValidEmail(): T;
  }
}

// Usage
it('generates valid email for new user', () => {
  const user = createUser({ name: 'Alice' });
  expect(user.email).toBeValidEmail();
});
```

#### Type-level Testing with `expect-type`

```typescript
import { expectTypeOf } from 'vitest';

it('returns correct types', () => {
  // Verify function signature types
  expectTypeOf(getUser).parameter(0).toBeString();
  expectTypeOf(getUser).returns.resolves.toMatchTypeOf<User>();

  // Verify discriminated union exhaustiveness
  expectTypeOf<Result<string>>().toMatchTypeOf<
    { ok: true; value: string } | { ok: false; error: Error }
  >();

  // Verify type utilities produce correct types
  expectTypeOf<WithRequired<Partial<User>, 'id'>>().toHaveProperty('id');
});
```

### Snapshot Testing

| Use Snapshots For | Avoid Snapshots For |
|-------------------|---------------------|
| Serialized data structures (API responses, configs) | Full component trees (too brittle) |
| Error message formatting | Styled components (CSS changes break snapshots) |
| CLI output strings | Large objects (unreadable diffs) |

```typescript
// Good - small, focused snapshot
test('formats error response', () => {
  const error = formatApiError(404, 'User not found');
  expect(error).toMatchInlineSnapshot(`
    {
      "code": 404,
      "message": "User not found",
      "type": "NOT_FOUND",
    }
  `);
});

// Bad - entire component tree snapshot
test('renders dashboard', () => {
  const { container } = render(<Dashboard />);
  expect(container).toMatchSnapshot(); // 500+ line snapshot nobody reviews
});
```

### Coverage Expectations

| Level | Minimum | Target | Notes |
|-------|---------|--------|-------|
| Overall | 60% | 80% | Enforced in CI |
| Critical paths | 80% | 90% | Auth, payments, data mutations |
| Utility functions | 80% | 95% | Pure functions are easy to test |
| Type guards | 100% | 100% | Runtime type safety boundary |

```bash
# Run tests with coverage
npx vitest run --coverage

# Check coverage thresholds (in vitest.config.ts)
# coverage.thresholds: { lines: 60, branches: 60, functions: 60 }
```

**What to cover vs. what to skip:**

| Cover | Skip |
|-------|------|
| Business logic and domain rules | Generated code (protobuf, GraphQL types) |
| Error handling paths | Third-party library internals |
| Type guards (runtime boundary) | Trivial getters/setters |
| State transitions | Framework boilerplate (module re-exports) |
| Edge cases in parsing/validation | Static configuration objects |

### ALWAYS / NEVER Rules

| Rule | Type | Rationale |
|------|------|-----------|
| ALWAYS use `userEvent` over `fireEvent` | ALWAYS | `userEvent` simulates real browser behavior (focus, hover, keystrokes) |
| ALWAYS use `findBy*` for async elements | ALWAYS | Avoids race conditions; auto-retries until timeout |
| ALWAYS set `onUnhandledRequest: 'error'` in MSW | ALWAYS | Catches unmocked API calls that indicate missing test setup |
| ALWAYS co-locate test files with source | ALWAYS | Easier navigation; test dies when source is deleted |
| ALWAYS type your mock functions | ALWAYS | Catches incorrect call signatures at compile time |
| ALWAYS clean up mocks in `afterEach` | ALWAYS | Prevents test pollution; use `vi.restoreAllMocks()` or `jest.restoreAllMocks()` |
| NEVER use `container.querySelector` in RTL tests | NEVER | Bypasses accessibility queries; tests implementation not behavior |
| NEVER use `setTimeout` / manual delays in tests | NEVER | Flaky; use `waitFor` or `findBy*` instead |
| NEVER snapshot full component trees | NEVER | Unreadable diffs; nobody reviews 500-line snapshots |
| NEVER mock what you don't own without MSW | NEVER | Direct `jest.mock('axios')` couples tests to HTTP library choice |
| NEVER use `as any` to silence mock type errors | NEVER | Hides real type mismatches; use proper typed mocks instead |

---

## Anti-Patterns Avoided

> See `common-standards.md` for universal anti-patterns across all languages.

### No Any Escape

```typescript
// Bad - defeats type safety
const data = response as any;
const typed = data as User;

// Good - use unknown + type guard
const data: unknown = response;
if (isUser(data)) {
  const typed: User = data;
}
```

### No Non-null Assertion Spam

```typescript
// Bad - runtime errors if assumption wrong
const name = user!.profile!.displayName!;

// Good - proper null handling
const name = user?.profile?.displayName ?? 'Anonymous';
```

### No Index Signature Abuse

```typescript
// Bad - no type safety
interface Config {
  [key: string]: any;
}

// Good - explicit properties
interface Config {
  apiUrl: string;
  timeout: number;
  features: string[];
}

// Or generic when truly dynamic
type Config<T extends string> = Record<T, string>;
```

### No Enum for Strings

```typescript
// Bad - verbose, poor tree-shaking
enum Color {
  Red = 'RED',
  Blue = 'BLUE',
}

// Good - union type
type Color = 'RED' | 'BLUE';

// Or const object for runtime values
const Color = {
  Red: 'RED',
  Blue: 'BLUE',
} as const;
type Color = typeof Color[keyof typeof Color];
```

---

## Compliance Assessment

**Use letter grades + evidence, NOT numeric scores.**

### Assessment Categories

| Category | Evidence Required |
|----------|------------------|
| **Type Safety** | tsc error count, any usage count, strict mode enabled |
| **Code Quality** | ESLint violations count, unused variables |
| **Type Coverage** | Explicit return types on exports (count), any/unknown ratio |
| **Best Practices** | Discriminated union usage, type guard count |
| **Testing** | Test file count, coverage % |

### Grading Scale

| Grade | Criteria |
|-------|----------|
| A+ | 0 tsc errors, 0 any types, strict mode, 0 ESLint errors, 100% return types |
| A | 0 tsc errors, <3 any types (justified), <5 ESLint errors, 95%+ return types |
| A- | <5 tsc errors, <10 any types, <15 ESLint errors, 85%+ return types |
| B+ | <15 tsc errors, <20 any types, <30 ESLint errors, 75%+ return types |
| B | <30 tsc errors, <40 any types, <50 ESLint errors, 60%+ return types |
| C | Significant type safety issues |
| D | Not production-ready |
| F | Critical issues |

### Example Assessment

```markdown
## TypeScript Standards Compliance

**Target:** src/
**Date:** 2026-01-21

| Category | Grade | Evidence |
|----------|-------|----------|
| Type Safety | A+ | 0 tsc errors, 0 any types, strict mode |
| Code Quality | A- | 8 ESLint violations (6 auto-fixable) |
| Type Coverage | A | 48/52 exports typed (92%) |
| Best Practices | A | 12 discriminated unions, 8 type guards |
| **OVERALL** | **A** | **2 HIGH, 6 MEDIUM findings** |
```

---

## Vibe Integration

### Prescan Patterns

| Pattern | Severity | Detection |
|---------|----------|-----------|
| P10: any type usage | HIGH | `: any` without justification |
| P11: Non-null assertion spam | MEDIUM | Multiple `!` in chain |
| P12: Missing import type | LOW | `import {` for type-only |

### JIT Loading

**Tier 1 (Fast):** Load `~/.agents/skills/standards/references/typescript.md` (5KB)
**Tier 2 (Deep):** Load this document (18KB) for comprehensive audit

---

## Additional Resources

- [TypeScript Handbook](https://www.typescriptlang.org/docs/handbook/)
- [typescript-eslint](https://typescript-eslint.io/)
- [Total TypeScript](https://www.totaltypescript.com/)
- [Type Challenges](https://github.com/type-challenges/type-challenges)

---

**Related:** Quick reference in Tier 1 `typescript.md`

### vibe-coding.md

# Vibe-Coding Science Reference

**JIT-loaded by $vibe skill and validation agents**

---

## Vibe Levels (Trust Calibration)

| Level | Trust | Verify | Use For | Tracer Test |
|:-----:|:-----:|:------:|---------|-------------|
| **5** | 95% | Final only | Format, lint, imports | Smoke (2m) |
| **4** | 80% | Spot check | Boilerplate, renames | Environment (5m) |
| **3** | 60% | Key outputs | CRUD, tests, known patterns | Integration (10m) |
| **2** | 40% | Every change | New features, integrations | Components (15m) |
| **1** | 20% | Every line | Architecture, security | All assumptions (30m) |
| **0** | 0% | N/A | Novel research | Feasibility (15m) |

**Most tasks are L3.** When in doubt, go lower.

---

## The 5 Core Metrics

| Metric | Target | Red Flag | What It Means |
|--------|:------:|:--------:|---------------|
| **Iteration Velocity** | >3/hr | <1/hr | Feedback loop frequency |
| **Rework Ratio** | <30% | >50% | Building vs debugging |
| **Trust Pass Rate** | >80% | <60% | Code acceptance rate |
| **Debug Spiral Duration** | <30m | >60m | Time stuck on issues |
| **Flow Efficiency** | >75% | <50% | Productive time ratio |

**The key number:** Trust Pass Rate. If >80%, building. If <60%, debugging.

---

## Rating Thresholds

| Metric | ELITE | HIGH | MEDIUM | LOW |
|--------|:-----:|:----:|:------:|:---:|
| Velocity | >5 | ≥3 | ≥1 | <1 |
| Rework | <30% | <50% | <70% | ≥70% |
| Trust Pass | >95% | ≥80% | ≥60% | <60% |
| Spiral | <15m | <30m | <60m | ≥60m |
| Flow | >90% | ≥75% | ≥50% | <50% |

---

## PDC Framework

| Phase | Question | Actions |
|-------|----------|---------|
| **Prevent** | Could we have avoided this? | Specs, checkpoints, tests, 40% rule |
| **Detect** | How did we catch it? | TDD, verify claims, monitor |
| **Correct** | How do we fix it? | Fresh session, rollback, modularize |

**Investment ratio:** Prevention (1x) > Detection (10x) > Correction (100x)

---

## The 12 Failure Patterns

### Inner Loop (Seconds-Minutes)

| # | Pattern | Symptom | Fix |
|:-:|---------|---------|-----|
| 1 | **Tests Lie** | AI says "pass" but broken | Run tests yourself |
| 2 | **Amnesia** | Forgets constraints | Fresh session (>40%) |
| 3 | **Drift** | "Improving" undirected | Smaller tasks |
| 4 | **Debug Spiral** | 3rd log, no fix | Real debugger |

### Middle Loop (Hours-Days)

| # | Pattern | Symptom | Fix |
|:-:|---------|---------|-----|
| 5 | **Eldritch Horror** | 3000-line function | Test. Modularize |
| 6 | **Collision** | Same files | Clear territories |
| 7 | **Memory Decay** | Re-solving | Bundle maintenance |
| 8 | **Deadlock** | Agents waiting | Break cycle |

### Outer Loop (Weeks-Months)

| # | Pattern | Symptom | Fix |
|:-:|---------|---------|-----|
| 9 | **Bridge Torch** | API broke downstream | Roll back |
| 10 | **Deletion** | "Unused" removed | Approval required |
| 11 | **Gridlock** | PRs backed up | Fast lane |
| 12 | **Stewnami** | Half-done pile | Limit WIP |

---

## Code Review Calibration

| Task | Max Level | Notes |
|------|:---------:|-------|
| Generate review comments | L4 | Suggestions only |
| Apply review suggestions | L3 | Verify applies |
| Security review findings | L2 | Higher risk |
| Automated linting | L5 | Fully automated |

---

## Grade Mapping

| Vibe Grade | Trust Pass | Verdict |
|:----------:|:----------:|---------|
| **A** | >95% | ELITE - ship it |
| **B** | ≥80% | HIGH - minor fixes |
| **C** | ≥60% | MEDIUM - needs work |
| **D** | <60% | LOW - significant issues |
| **F** | <40% | BLOCK - systemic problems |

---

## 40% Context Rule

| Utilization | Effect | Action |
|:-----------:|--------|--------|
| 0-40% | Optimal | Continue |
| 40-60% | Degradation | Checkpoint |
| 60-80% | Instruction loss | Save state |
| 80-100% | Confabulation | STOP |

---

**Source:** gitops/docs/methodology/vibe-ecosystem/vibe-coding/

### yaml-standards.md

# YAML/Helm Standards Catalog - Vibe Canonical Reference

**Version:** 1.0.0
**Last Updated:** 2026-01-21
**Purpose:** Canonical YAML/Helm standards for vibe skill validation

---

## Table of Contents

1. [yamllint Configuration](#yamllint-configuration)
2. [Formatting Rules](#formatting-rules)
3. [Helm Chart Conventions](#helm-chart-conventions)
4. [Kustomize Patterns](#kustomize-patterns)
5. [Template Best Practices](#template-best-practices)
6. [Validation Workflow](#validation-workflow)
7. [Compliance Assessment](#compliance-assessment)
8. [Anti-Patterns Avoided](#anti-patterns-avoided)
9. [Code Quality Metrics](#code-quality-metrics)
10. [Prescan Patterns](#prescan-patterns)

---

## yamllint Configuration

### Full Configuration

```yaml
# .yamllint.yml
extends: default
rules:
  line-length:
    max: 120
    allow-non-breakable-inline-mappings: true
  indentation:
    spaces: 2
    indent-sequences: consistent
  truthy:
    check-keys: false
  comments:
    min-spaces-from-content: 1
  document-start: disable
  empty-lines:
    max: 2
  brackets:
    min-spaces-inside: 0
    max-spaces-inside: 0
  colons:
    max-spaces-before: 0
    max-spaces-after: 1
  commas:
    max-spaces-before: 0
    min-spaces-after: 1
  hyphens:
    max-spaces-after: 1
```

### Usage

```bash
# Lint all YAML files
yamllint .

# Lint specific directory
yamllint apps/

# Lint with format output
yamllint -f parsable .
```

---

## Formatting Rules

### Quoting Strings

```yaml
# Quote strings that look like other types
enabled: "true"      # String, not boolean
port: "8080"         # String, not integer
version: "1.0"       # String, not float

# No quotes for actual typed values
enabled: true        # Boolean
port: 8080           # Integer
replicas: 3          # Integer
```

### Multi-line Strings

```yaml
# Literal block scalar (preserves newlines)
script: |
  #!/bin/bash
  set -euo pipefail
  echo "Hello"

# Folded block scalar (folds newlines to spaces)
description: >
  This is a long description that will be
  folded into a single line with spaces.

# BAD - Escaped newlines (hard to read)
script: "#!/bin/bash\nset -euo pipefail\necho \"Hello\""
```

### Comments

```yaml
# Section header (full line)
# =============================================================================
# Database Configuration
# =============================================================================

database:
  host: localhost      # Inline comment (1 space before #)
  port: 5432
  # Subsection comment
  credentials:
    username: admin
```

---

## Helm Chart Conventions

### Chart Structure

```text
charts/<chart-name>/
├── Chart.yaml
├── values.yaml
├── values.schema.json    # Optional: JSON Schema for values
├── templates/
│   ├── _helpers.tpl
│   ├── deployment.yaml
│   ├── service.yaml
│   └── ...
└── charts/               # Nested charts (if needed)
```

### Chart.yaml

```yaml
apiVersion: v2
name: my-app
description: A Helm chart for my application
type: application
version: 1.0.0
appVersion: "2.0.0"

dependencies:
  - name: postgresql
    version: "12.x.x"
    repository: https://charts.bitnami.com/bitnami
    condition: postgresql.enabled
```

### values.yaml Conventions

```yaml
# =============================================================================
# Application Configuration
# =============================================================================

app:
  name: my-app
  replicas: 3

# Resource limits (adjust for environment)
resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 512Mi

# =============================================================================
# Image Configuration
# =============================================================================

image:
  repository: myregistry/my-app
  tag: ""  # Defaults to appVersion
  pullPolicy: IfNotPresent
```

### Validation Commands

```bash
# Lint chart
helm lint charts/<chart-name>/

# Template with values (dry-run)
helm template <release> charts/<chart-name>/ -f values.yaml

# Validate rendered output
helm template <release> charts/<chart-name>/ | kubectl apply --dry-run=client -f -

# Debug template rendering
helm template <release> charts/<chart-name>/ --debug
```

---

## Kustomize Patterns

### Overlay Structure

```text
apps/<app>/
├── base/
│   ├── kustomization.yaml
│   ├── deployment.yaml
│   └── service.yaml
└── overlays/
    ├── dev/
    │   └── kustomization.yaml
    ├── staging/
    │   └── kustomization.yaml
    └── prod/
        └── kustomization.yaml
```

### kustomization.yaml Template

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - deployment.yaml
  - service.yaml

# Environment-specific patches
patches:
  - path: ./patches/replicas.yaml
    target:
      kind: Deployment
      name: my-app
```

### Patch Types

**Strategic Merge Patch:**
```yaml
# patches/extend-rbac.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: my-role
rules:
  - apiGroups: ["custom.io"]
    resources: ["widgets"]
    verbs: ["get", "list"]
```

**JSON Patch:**
```yaml
# patches/add-annotation.yaml
- op: add
  path: /metadata/annotations/custom.io~1managed
  value: "true"
```

**Delete Patch:**
```yaml
# patches/delete-resource.yaml
$patch: delete
apiVersion: v1
kind: ConfigMap
metadata:
  name: unused-config
```

---

## Template Best Practices

### Use include for Reusable Snippets

```yaml
# templates/_helpers.tpl
{{- define "app.labels" -}}
app.kubernetes.io/name: {{ .Chart.Name }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}

# templates/deployment.yaml
metadata:
  labels:
    {{- include "app.labels" . | nindent 4 }}
```

### Whitespace Control

```yaml
# GOOD - Use {{- and -}} to control whitespace
{{- if .Values.enabled }}
apiVersion: v1
kind: ConfigMap
{{- end }}

# BAD - Extra blank lines in output
{{ if .Values.enabled }}

apiVersion: v1

{{ end }}
```

### Required Values

```yaml
# Fail fast if required value missing
image: {{ required "image.repository is required" .Values.image.repository }}:{{ .Values.image.tag | default .Chart.AppVersion }}
```

### Default Values

```yaml
# Safe defaults
replicas: {{ .Values.replicas | default 1 }}

# Nested defaults
resources:
  {{- with .Values.resources }}
  {{- toYaml . | nindent 2 }}
  {{- else }}
  requests:
    cpu: 100m
    memory: 128Mi
  {{- end }}
```

---

## Validation Workflow

### Pre-commit Checks

```bash
# 1. Lint YAML
yamllint .

# 2. Lint Helm charts
for chart in charts/*/Chart.yaml; do
    helm lint "$(dirname "$chart")"
done

# 3. Build Kustomize overlays
kustomize build apps/<app>/ --enable-helm > /dev/null
```

### CI Pipeline Example

```yaml
# .github/workflows/validate.yaml
- name: Lint YAML
  run: yamllint .

- name: Lint Helm
  run: |
    for chart in charts/*/Chart.yaml; do
      helm lint "$(dirname "$chart")"
    done

- name: Validate Kustomize
  run: |
    for kust in apps/*/kustomization.yaml; do
      kustomize build "$(dirname "$kust")" --enable-helm > /dev/null
    done
```

---

## Compliance Assessment

**Use letter grades + evidence, NOT numeric scores.**

### Assessment Categories

| Category | Evidence Required |
|----------|------------------|
| **Formatting** | yamllint violations, tab count, indentation |
| **Helm Charts** | helm lint output, template rendering |
| **Kustomize** | kustomize build success, patch correctness |
| **Documentation** | values.yaml comments, section headers |
| **Security** | Hardcoded secrets, external secret refs |

### Grading Scale

| Grade | Criteria |
|-------|----------|
| A+ | 0 yamllint errors, 0 helm lint errors, documented, 0 secrets |
| A | <3 yamllint warnings, <3 helm lint warnings, documented |
| A- | <10 warnings, partial docs |
| B+ | <20 warnings |
| B | <40 warnings, templates render |
| C | Significant issues |

### Validation Commands

```bash
# Lint YAML
yamllint .
# Output: "X error(s), Y warning(s)"

# Check for tabs
grep -rP '\t' --include='*.yaml' --include='*.yml' . | wc -l
# Should be 0

# Helm lint
for chart in charts/*/Chart.yaml; do
  helm lint "$(dirname "$chart")"
done

# Check for hardcoded secrets
grep -r "password:\|secret:\|token:" --include='*.yaml' apps/
# Should only return external references
```

### Example Assessment

```markdown
## YAML/Helm Standards Compliance

| Category | Grade | Evidence |
|----------|-------|----------|
| Formatting | A+ | 0 yamllint errors, 0 tabs |
| Helm Charts | A- | 3 lint warnings (docs) |
| Kustomize | A | All overlays build |
| Security | A | 0 hardcoded secrets |
| **OVERALL** | **A** | **3 MEDIUM findings** |
```

---

## Anti-Patterns Avoided

### ❌ **Implicit Typing Traps (Norway Problem)**

Unquoted values silently coerced to unexpected types:

```yaml
# BAD - These become booleans (false, true)
country: NO       # false
feature: YES      # true
enabled: on       # true
disabled: off     # false

# GOOD - Quote ambiguous strings
country: "NO"
feature: "YES"
enabled: "on"
disabled: "off"
```

### ❌ **Anchor/Alias Abuse**

Overuse of `&` anchors and `*` aliases creates unreadable configs:

```yaml
# BAD - Excessive aliasing obscures intent
defaults: &defaults
  timeout: 30
  retries: 3

service_a:
  <<: *defaults
  name: a

service_b:
  <<: *defaults
  name: b

# GOOD - Explicit values for clarity (or use Kustomize overlays)
service_a:
  timeout: 30
  retries: 3
  name: a
```

**Rule:** Anchors acceptable for DRY in 2-3 references. Beyond that, use templating (Helm, Kustomize).

### ❌ **Deeply Nested Configs**

Nesting beyond 6 levels signals structural problems:

```yaml
# BAD - 7+ levels deep
app:
  server:
    routes:
      api:
        v1:
          users:
            endpoints:
              list:
                timeout: 30

# GOOD - Flatten with dotted keys or restructure
app:
  server:
    routes:
      api-v1-users-list:
        timeout: 30
```

### ❌ **Missing Document Markers**

Multi-document YAML files without `---` separators cause parse failures:

```yaml
# BAD - Two documents, no separator
apiVersion: v1
kind: ConfigMap
metadata:
  name: config-a
apiVersion: v1
kind: ConfigMap
metadata:
  name: config-b

# GOOD - Explicit document markers
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: config-a
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: config-b
```

### ❌ **Mixed Indentation**

Tabs or inconsistent indent widths cause silent parse errors:

```yaml
# BAD - Tabs mixed with spaces (invisible breakage)
app:
	name: broken    # Tab character

# BAD - Inconsistent indent width
app:
  name: my-app     # 2 spaces
  config:
      port: 8080   # 4 spaces

# GOOD - Consistent 2-space indentation throughout
app:
  name: my-app
  config:
    port: 8080
```

---

## Code Quality Metrics

### Validation Thresholds

| Metric | Threshold | Status | Action |
|--------|-----------|--------|--------|
| yamllint errors | 0 | ✅ Required | Fix before merge |
| yamllint warnings | <5 | ✅ Acceptable | Fix in next PR |
| yamllint warnings | 5-20 | ⚠️ Warning | Refactor recommended |
| yamllint warnings | 20+ | ❌ Critical | Block merge |
| Nesting depth | ≤6 levels | ✅ Acceptable | Flatten if deeper |
| Line length | ≤256 chars | ✅ Maximum | Prefer ≤120 |
| Helm lint errors | 0 | ✅ Required | Fix before merge |
| Kustomize build | Pass | ✅ Required | All overlays must build |
| Hardcoded secrets | 0 | ✅ Required | Use external refs |

### Tool Commands

```bash
# Full validation pass
yamllint -f parsable . | wc -l          # Total findings
yamllint -f parsable . | grep error     # Errors only
helm lint charts/*/                      # Helm validation
grep -rP '\t' --include='*.yaml' . | wc -l  # Tab detection
```

---

## Prescan Patterns

| ID | Pattern | Detection Command | Severity |
|----|---------|-------------------|----------|
| P01 | Implicit boolean detection | `yamllint -d '{extends: default, rules: {truthy: {check-keys: true}}}' .` | HIGH |
| P02 | Duplicate keys | `yamllint -d '{extends: default, rules: {key-duplicates: enable}}' .` | HIGH |
| P03 | Excessive nesting (>6 levels) | `awk '/^( ){14}[^ ]/' *.yaml` | MEDIUM |
| P04 | Long lines (>256 chars) | `yamllint -d '{extends: default, rules: {line-length: {max: 256}}}' .` | MEDIUM |
| P05 | Missing document marker | `grep -rL '^---' --include='*.yaml' .` | LOW |

### Pattern Details

**P01: Implicit Boolean Detection**
Catches the Norway problem — unquoted values like `NO`, `YES`, `on`, `off` silently become booleans. The `truthy` rule with `check-keys: true` flags these in both keys and values.

**P02: Duplicate Keys**
Duplicate keys in the same mapping silently overwrite earlier values. YAML spec allows it but most parsers keep only the last value, causing hard-to-debug configuration drift.

**P03: Excessive Nesting**
Detects indentation at 14+ spaces (7+ levels at 2-space indent). Deep nesting indicates config structure should be flattened or split into overlays.

**P04: Long Lines**
Lines beyond 256 characters indicate inline lists or values that should use block scalars. Default yamllint threshold is 120; 256 is the hard maximum.

**P05: Missing Document Marker**
Multi-resource YAML files (common in Kubernetes) require `---` separators. Missing markers cause concatenation errors during apply.

---

## Additional Resources

- [YAML Spec](https://yaml.org/spec/)
- [Helm Documentation](https://helm.sh/docs/)
- [Kustomize Documentation](https://kustomize.io/)
- [yamllint Documentation](https://yamllint.readthedocs.io/)

---

**Related:** Quick reference in Tier 1 `yaml.md`


---

## Scripts

### ol-validate.sh

```bash
#!/usr/bin/env bash
# Olympus (OL) deterministic validation for vibe checks.
# Parses Stage1Result JSON from `ol validate stage1`.
set -euo pipefail

# --- Guard: detect OL environment ---
has_config=false
has_binary=false

if [ -f ".ol/config.yaml" ]; then
  has_config=true
fi

if command -v ol >/dev/null 2>&1; then
  has_binary=true
fi

if [ "$has_config" = false ] && [ "$has_binary" = false ]; then
  echo "SKIPPED (ol not detected)"
  exit 2
fi

# OL is detected (at least one of config/binary found).
# From here, errors should be reported but exit 2 (skip), not crash.

if [ "$has_binary" = false ]; then
  echo "SKIPPED (ol error: config found but ol binary not on PATH)"
  exit 2
fi

if [ "$has_config" = false ]; then
  echo "SKIPPED (ol error: ol binary found but .ol/config.yaml missing)"
  exit 2
fi

# --- Run ol validate stage1 ---
json=""
if ! json="$(ol validate stage1 -o json 2>&1)"; then
  echo "SKIPPED (ol error: ol validate stage1 failed: ${json})"
  exit 2
fi

# --- Parse Stage1Result JSON ---
passed=""
if ! passed="$(echo "$json" | jq -r 'if has("passed") then .passed else error("missing .passed") end' 2>&1)"; then
  echo "SKIPPED (ol error: failed to parse .passed from Stage1Result: ${passed})"
  exit 2
fi

summary=""
if ! summary="$(echo "$json" | jq -re '.summary' 2>&1)"; then
  echo "SKIPPED (ol error: failed to parse .summary from Stage1Result: ${summary})"
  exit 2
fi

# --- Build report ---
if [ "$passed" = "true" ]; then
  status="PASSED"
else
  status="FAILED"
fi

echo "## Deterministic Validation (Olympus)"
echo ""
echo "**Status:** ${status}"
echo ""
echo "| Step | Duration | Exit Code | Passed |"
echo "|------|----------|-----------|--------|"

# Parse steps array
step_count=""
if ! step_count="$(echo "$json" | jq -re '.steps | length' 2>&1)"; then
  echo "SKIPPED (ol error: failed to parse .steps from Stage1Result: ${step_count})"
  exit 2
fi

for ((i = 0; i < step_count; i++)); do
  step_name="$(echo "$json" | jq -r ".steps[$i].name")"
  step_duration="$(echo "$json" | jq -r ".steps[$i].duration")"
  step_exit="$(echo "$json" | jq -r ".steps[$i].exit_code")"
  step_passed="$(echo "$json" | jq -r ".steps[$i].passed")"
  echo "| ${step_name} | ${step_duration} | ${step_exit} | ${step_passed} |"
done

echo ""
echo "**Summary:** ${summary}"

# --- Exit ---
if [ "$passed" = "true" ]; then
  exit 0
else
  exit 1
fi
```

### prescan.sh

```bash
#!/usr/bin/env bash
# Vibe Pre-Scan: Fast static detection for 7 failure patterns
# Usage: prescan.sh <target>
#   target: recent | all | <directory> | <file>

set -euo pipefail

TARGET="${1:-recent}"

# Validate TARGET to prevent argument injection
if [[ "$TARGET" =~ ^- ]]; then
    echo "Error: TARGET cannot start with a dash (prevents argument injection)" >&2
    exit 1
fi
if [[ "$TARGET" != "recent" && "$TARGET" != "all" && ! -e "$TARGET" ]]; then
    echo "Error: TARGET '$TARGET' does not exist" >&2
    exit 1
fi

# File filtering (exclude generated code, build artifacts, test fixtures)
filter_files() {
  grep -v '__pycache__\|\.venv\|venv/\|node_modules\|\.git/\|test_fixtures\|/fixtures/\|\.eggs\|egg-info\|/dist/\|/build/\|\.tox\|\.mypy_cache\|\.pytest_cache' \
  | grep -v '\.gen\.go$\|zz_generated\|_generated\.go$\|\.pb\.go$\|mock_.*\.go$\|/generated/\|/gen/\|deepcopy'
}

# Resolve target to file lists (Python, Go, Bash)
case "$TARGET" in
  recent)
    PY_FILES=$(git diff --name-only HEAD~1 HEAD 2>/dev/null | grep '\.py$' | filter_files || true)
    GO_FILES=$(git diff --name-only HEAD~1 HEAD 2>/dev/null | grep '\.go$' | filter_files || true)
    SH_FILES=$(git diff --name-only HEAD~1 HEAD 2>/dev/null | grep '\.sh$' | filter_files || true)
    MODE="Recent"
    ;;
  all)
    PY_FILES=$(find . -name "*.py" -type f 2>/dev/null | filter_files | grep -v 'test_' || true)
    GO_FILES=$(find . -name "*.go" -type f 2>/dev/null | filter_files | grep -v '_test\.go$' || true)
    SH_FILES=$(find . -name "*.sh" -type f 2>/dev/null | filter_files || true)
    MODE="All"
    ;;
  *)
    if [ -d "$TARGET" ]; then
      PY_FILES=$(find "$TARGET" -name "*.py" -type f 2>/dev/null | filter_files || true)
      GO_FILES=$(find "$TARGET" -name "*.go" -type f 2>/dev/null | filter_files || true)
      SH_FILES=$(find "$TARGET" -name "*.sh" -type f 2>/dev/null | filter_files || true)
      MODE="Dir"
    elif [ -f "$TARGET" ]; then
      case "$TARGET" in
        *.py) PY_FILES="$TARGET"; GO_FILES=""; SH_FILES="" ;;
        *.go) GO_FILES="$TARGET"; PY_FILES=""; SH_FILES="" ;;
        *.sh) SH_FILES="$TARGET"; PY_FILES=""; GO_FILES="" ;;
        *) PY_FILES="$TARGET"; GO_FILES=""; SH_FILES="" ;;
      esac
      MODE="File"
    else
      echo "ERROR: Target not found: $TARGET" >&2
      exit 1
    fi
    ;;
esac

# Combine for backwards compatibility
FILES="$PY_FILES"
[ -n "$GO_FILES" ] && FILES=$(printf "%s\n%s" "$FILES" "$GO_FILES")
[ -n "$SH_FILES" ] && FILES=$(printf "%s\n%s" "$FILES" "$SH_FILES")

# Count files (handle empty strings properly)
count_lines() {
  local input="$1"
  [ -z "$input" ] && echo 0 && return
  echo "$input" | wc -l | tr -d ' '
}
PY_COUNT=$(count_lines "$PY_FILES")
GO_COUNT=$(count_lines "$GO_FILES")
SH_COUNT=$(count_lines "$SH_FILES")
FILE_COUNT=$((PY_COUNT + GO_COUNT + SH_COUNT))
if [ "$FILE_COUNT" -eq 0 ]; then
  echo "No files found for target: $TARGET"
  exit 0
fi

echo "Pre-Scan Target: $TARGET"
echo "Mode: $MODE | Files: $FILE_COUNT (py:$PY_COUNT go:$GO_COUNT sh:$SH_COUNT)"
echo ""

# Initialize counters
P1_COUNT=0
P2_COUNT=0
P4_COUNT=0
P5_COUNT=0
P8_COUNT=0
P9_COUNT=0
P12_COUNT=0

# P1: Phantom Modifications (CRITICAL)
# Committed lines not in current file
echo "[P1] Phantom Modifications"
if [ "$TARGET" = "recent" ]; then
  for file in $FILES; do
    [ -f "$file" ] || continue
    while IFS= read -r line; do
      clean=$(echo "$line" | sed 's/^+//' | xargs)
      if [ ${#clean} -gt 10 ] && ! grep -qF "$clean" "$file" 2>/dev/null; then
        echo "  - $file: Committed line missing: \"${clean:0:50}...\""
        P1_COUNT=$((P1_COUNT + 1))
      fi
    done < <(git show HEAD -- "$file" 2>/dev/null | grep '^+[^+]' || true)
  done
fi
echo "  $P1_COUNT findings"

# P2: Hardcoded Secrets (CRITICAL)
# Uses path-based filtering to exclude test directories
echo ""
echo "[P2] Hardcoded Secrets"
for file in $FILES; do
  [ -f "$file" ] || continue
  # Skip test directories
  case "$file" in
    */test/*|*/tests/*|*_test.*|*/example/*|*/examples/*|*.example.*) continue ;;
  esac
  while IFS= read -r match; do
    line_num=$(echo "$match" | cut -d: -f1)
    echo "  - $file:$line_num: Possible hardcoded secret"
    P2_COUNT=$((P2_COUNT + 1))
  done < <(grep -n -E "(password|secret|api_key|apikey|token)\s*=\s*['\"][^'\"]+['\"]" "$file" 2>/dev/null | head -5 || true)
done
echo "  $P2_COUNT findings"

# P4: Invisible Undone (HIGH)
# Detects: unfinished work markers, commented-out code
echo ""
echo "[P4] Invisible Undone"
for file in $FILES; do
  [ -f "$file" ] || continue
  # TODO/FIXME markers
  while IFS= read -r match; do
    line_num=$(echo "$match" | cut -d: -f1)
    echo "  - $file:$line_num: TODO marker"
    P4_COUNT=$((P4_COUNT + 1))
  done < <(grep -n "TODO\|FIXME\|XXX\|HACK" "$file" 2>/dev/null | head -3 || true)
  # Commented code
  while IFS= read -r match; do
    line_num=$(echo "$match" | cut -d: -f1)
    echo "  - $file:$line_num: Commented code"
    P4_COUNT=$((P4_COUNT + 1))
  done < <(grep -n "^\s*#\s*\(def \|class \|if \|for \)" "$file" 2>/dev/null | head -2 || true)
done
echo "  $P4_COUNT findings"

# P5: Eldritch Horror (HIGH)
# Complexity CC > 15 or function > 50 lines
echo ""
echo "[P5] Eldritch Horror"

# Python: radon for cyclomatic complexity
if [ -n "$PY_FILES" ]; then
  if command -v radon &>/dev/null; then
    for file in $PY_FILES; do
      [ -f "$file" ] || continue
      while IFS= read -r line; do
        cc=$(echo "$line" | grep -oE '\([0-9]+\)' | tr -d '()')
        if [ -n "$cc" ] && [ "$cc" -gt 15 ]; then
          func=$(echo "$line" | awk '{print $3}')
          echo "  - $file: $func CC=$cc (py)"
          P5_COUNT=$((P5_COUNT + 1))
        fi
      done < <(radon cc "$file" -s -n E 2>/dev/null | grep -E "^\s*[EF]\s+[0-9]+" || true)
    done
  else
    echo "  WARNING: radon not installed (Python CC skipped)"
  fi
fi

# Go: gocyclo for cyclomatic complexity
if [ -n "$GO_FILES" ]; then
  if command -v gocyclo &>/dev/null; then
    for file in $GO_FILES; do
      [ -f "$file" ] || continue
      while IFS= read -r line; do
        # gocyclo output: "15 pkg funcName file.go:42:1"
        cc=$(echo "$line" | awk '{print $1}')
        func=$(echo "$line" | awk '{print $3}')
        loc=$(echo "$line" | awk '{print $4}')
        if [ -n "$cc" ] && [ "$cc" -gt 15 ]; then
          echo "  - $loc: $func CC=$cc (go)"
          P5_COUNT=$((P5_COUNT + 1))
        fi
      done < <(gocyclo -over 15 "$file" 2>/dev/null || true)
    done
  else
    echo "  WARNING: gocyclo not installed (Go CC skipped)"
  fi
fi

# Python: Function length > 50 lines
for file in $PY_FILES; do
  [ -f "$file" ] || continue
  python3 -c '
import ast, sys
fname = sys.argv[1]
try:
    with open(fname) as f: tree = ast.parse(f.read())
    for n in ast.walk(tree):
        if isinstance(n, (ast.FunctionDef, ast.AsyncFunctionDef)) and hasattr(n, "end_lineno"):
            lines = n.end_lineno - n.lineno + 1
            if lines > 50: print(f"  - {fname}:{n.lineno}: {n.name}() is {lines} lines (py)")
except: pass
' "$file" 2>/dev/null || true
done

# Go: Function length > 50 lines (simple heuristic)
# Limitation: This awk-based parser only detects `func ` at line start and `}` alone on a line.
# Multi-line signatures, nested braces, or unusual formatting may cause false positives/negatives.
# For production Go codebases, consider using gocyclo or go/ast for accurate metrics.
for file in $GO_FILES; do
  [ -f "$file" ] || continue
  awk '
    /^func / { fname=$0; start=NR; in_func=1 }
    in_func && /^}$/ {
      lines = NR - start + 1
      if (lines > 50) {
        # Extract function name
        match(fname, /func[[:space:]]+(\([^)]+\)[[:space:]]+)?([a-zA-Z_][a-zA-Z0-9_]*)/, arr)
        print "  - '"$file"':" start ": " arr[2] "() is " lines " lines (go)"
      }
      in_func=0
    }
  ' "$file" 2>/dev/null || true
done
echo "  $P5_COUNT findings"

# P8: Cargo Cult Error Handling (HIGH)
# Empty except, pass-only handlers, bare except
echo ""
echo "[P8] Cargo Cult Error Handling"

# Python: except:pass, bare except
for file in $PY_FILES; do
  [ -f "$file" ] || continue
  python3 -c '
import ast, sys
fname = sys.argv[1]
try:
    with open(fname) as f: tree = ast.parse(f.read())
    for n in ast.walk(tree):
        if isinstance(n, ast.Try):
            for h in n.handlers:
                if len(h.body) == 1 and isinstance(h.body[0], ast.Pass):
                    print(f"  - {fname}:{h.lineno}: except: pass (swallowed) (py)")
                if h.type is None:
                    print(f"  - {fname}:{h.lineno}: bare except (catches SystemExit) (py)")
except: pass
' "$file" 2>/dev/null || true
done

# Bash: shellcheck for error handling issues
if [ -n "$SH_FILES" ]; then
  if command -v shellcheck &>/dev/null; then
    for file in $SH_FILES; do
      [ -f "$file" ] || continue
      # SC2181: Check exit code directly, not via $?
      # SC2086: Double quote to prevent globbing/splitting
      # SC2046: Quote to prevent word splitting
      # SC2155: Declare and assign separately to avoid masking return values
      while IFS= read -r line; do
        echo "  - $line (sh)"
        P8_COUNT=$((P8_COUNT + 1))
      done < <(shellcheck -f gcc -S warning "$file" 2>/dev/null | head -5 || true)
    done
  else
    echo "  WARNING: shellcheck not installed (Bash checks skipped)"
  fi
fi
echo "  $P8_COUNT findings"

# P9: Documentation Phantom (MEDIUM)
# Docstrings claiming behavior not implemented (Python only)
echo ""
echo "[P9] Documentation Phantom"
for file in $PY_FILES; do
  [ -f "$file" ] || continue
  python3 -c '
import ast, re, sys
fname = sys.argv[1]
try:
    with open(fname) as f: src = f.read()
    tree = ast.parse(src)
    PATTERNS = [
        (r"\bvalidates?\b", ["raise", "ValueError", "return False"]),
        (r"\bensures?\b", ["assert", "raise"]),
        (r"\bencrypts?\b", ["crypto", "cipher"]),
        (r"\bauthenticat", ["token", "password"]),
        (r"\bsanitiz", ["escape", "strip"])
    ]
    for n in ast.walk(tree):
        if isinstance(n, (ast.FunctionDef, ast.AsyncFunctionDef)):
            if n.body and isinstance(n.body[0], ast.Expr) and isinstance(getattr(n.body[0], "value", None), ast.Constant):
                doc = str(n.body[0].value.value).lower()
                fsrc = (ast.get_source_segment(src, n) or "").lower()
                for pat, impl in PATTERNS:
                    if re.search(pat, doc) and not any(i in fsrc for i in impl):
                        print(f"  - {fname}:{n.lineno}: {n.name}() docstring mismatch")
                        break
except: pass
' "$file" 2>/dev/null || true
done
echo "  $P9_COUNT findings"

# P12: Zombie Code (MEDIUM)
# Unused functions, unreachable code after return (Python only)
echo ""
echo "[P12] Zombie Code"
for file in $PY_FILES; do
  [ -f "$file" ] || continue
  python3 -c '
import ast, sys
fname = sys.argv[1]
try:
    with open(fname) as f: src = f.read()
    tree = ast.parse(src)
    defined, called = set(), set()
    for n in ast.walk(tree):
        if isinstance(n, (ast.FunctionDef, ast.AsyncFunctionDef)) and not n.name.startswith("_"):
            defined.add(n.name)
        if isinstance(n, ast.Call):
            if isinstance(n.func, ast.Name): called.add(n.func.id)
            elif isinstance(n.func, ast.Attribute): called.add(n.func.attr)
    for fn in (defined - called):
        if fn not in ("main", "setup", "teardown") and not fn.startswith("test_"):
            print(f"  - {fname}: {fn}() may be unused")
    # Unreachable code
    for n in ast.walk(tree):
        if isinstance(n, (ast.FunctionDef, ast.AsyncFunctionDef)):
            for i, s in enumerate(n.body[:-1]):
                if isinstance(s, (ast.Return, ast.Raise)) and n.body[i+1:]:
                    nxt = n.body[i+1]
                    if not (isinstance(nxt, ast.Expr) and isinstance(getattr(nxt, "value", None), ast.Constant)):
                        print(f"  - {fname}:{nxt.lineno}: Unreachable after return/raise")
except: pass
' "$file" 2>/dev/null || true
done
echo "  $P12_COUNT findings"

# Summary
echo ""
echo "=============================================="
echo "Pre-Scan Results:"
CRITICAL=$((P1_COUNT + P2_COUNT))
HIGH=$((P4_COUNT + P5_COUNT + P8_COUNT))
MEDIUM=$((P9_COUNT + P12_COUNT))
TOTAL=$((CRITICAL + HIGH + MEDIUM))

echo "[P1] Phantom Modifications: $P1_COUNT findings"
echo "[P2] Hardcoded Secrets: $P2_COUNT findings"
echo "[P4] Invisible Undone: $P4_COUNT findings"
echo "[P5] Eldritch Horror: $P5_COUNT findings"
echo "[P8] Cargo Cult Error Handling: $P8_COUNT findings"
echo "[P9] Documentation Phantom: $P9_COUNT findings"
echo "[P12] Zombie Code: $P12_COUNT findings"
echo "----------------------------------------------"
echo "Summary: $TOTAL findings ($CRITICAL CRITICAL, $HIGH HIGH, $MEDIUM MEDIUM)"
echo ""

[ "$CRITICAL" -gt 0 ] && echo "CRITICAL: Fix P1, P2 immediately"
[ "$HIGH" -gt 0 ] && echo "HIGH: Review P4, P5, P8"
[ "$MEDIUM" -gt 0 ] && echo "MEDIUM: Consider P9, P12"
[ "$TOTAL" -eq 0 ] && echo "All clear - no violations"
echo "=============================================="

# Exit code based on findings
[ "$CRITICAL" -gt 0 ] && exit 2
[ "$HIGH" -gt 0 ] && exit 3
exit 0
```

### validate.sh

```bash
#!/usr/bin/env bash
set -euo pipefail
SKILL_DIR="$(cd "$(dirname "$0")/.." && pwd)"
PASS=0; FAIL=0

check() { if bash -c "$2"; then echo "PASS: $1"; PASS=$((PASS + 1)); else echo "FAIL: $1"; FAIL=$((FAIL + 1)); fi; }

check "SKILL.md exists" "[ -f '$SKILL_DIR/SKILL.md' ]"
check "SKILL.md has YAML frontmatter" "head -1 '$SKILL_DIR/SKILL.md' | grep -q '^---$'"
check "SKILL.md has name: vibe" "grep -q '^name: vibe' '$SKILL_DIR/SKILL.md'"
check "references/ has at least 5 files" "[ \$(ls '$SKILL_DIR/references/' | wc -l) -ge 5 ]"
check "scripts/prescan.sh exists" "[ -f '$SKILL_DIR/scripts/prescan.sh' ]"
check "scripts/prescan.sh is executable" "[ -x '$SKILL_DIR/scripts/prescan.sh' ]"
check "SKILL.md mentions complexity" "grep -qi 'complexity' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions council" "grep -qi 'council' '$SKILL_DIR/SKILL.md'"

echo ""; echo "Results: $PASS passed, $FAIL failed"
[ $FAIL -eq 0 ] && exit 0 || exit 1
```


