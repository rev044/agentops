---
name: pre-mortem
description: 'Validate a plan or spec before implementation using multi-model council. Answer: Is this good enough to implement? Triggers: "pre-mortem", "validate plan", "validate spec", "is this ready".'
---


# Pre-Mortem Skill

> **Purpose:** Is this plan/spec good enough to implement?

> **Mandatory for 3+ issue epics.** Pre-mortem is enforced by hook when `$crank` is invoked on epics with 3+ child issues. 6/6 consecutive positive ROI. Bypass: `--skip-pre-mortem` flag or `AGENTOPS_SKIP_PRE_MORTEM_GATE=1`.

Run `$council validate` on a plan or spec to get multi-model judgment before committing to implementation.

---

## Quick Start

```bash
$pre-mortem                                         # validates most recent plan (inline, no spawning)
$pre-mortem path/to/PLAN.md                         # validates specific plan (inline)
$pre-mortem --deep path/to/SPEC.md                  # 4 judges (thorough review, spawns agents)
$pre-mortem --mixed path/to/PLAN.md                 # cross-vendor (Claude + Codex)
$pre-mortem --preset=architecture path/to/PLAN.md   # architecture-focused review
$pre-mortem --explorers=3 path/to/SPEC.md           # deep investigation of plan
$pre-mortem --debate path/to/PLAN.md                # two-round adversarial review
```

---

## Execution Steps

### Step 1: Find the Plan/Spec

**If path provided:** Use it directly.

**If no path:** Find most recent plan:
```bash
ls -lt .agents/plans/ 2>/dev/null | head -3
ls -lt .agents/specs/ 2>/dev/null | head -3
```

Use the most recent file. If nothing found, ask user.

### Step 1.5: Default Inline Mode

**By default, pre-mortem runs inline (`--quick`)** — single-agent structured review, no spawning. This catches real implementation issues at ~10% of full council cost (proven in ag-nsx: 3 actionable bugs found inline that would have caused runtime failures).

**Skip Steps 1a and 1b** (knowledge search, product context) unless `--deep`, `--mixed`, `--debate`, or `--explorers` is set. These pre-processing steps are for multi-judge council packets only.

To escalate to full multi-judge council, use `--deep` (4 judges) or `--mixed` (cross-vendor).

### Step 1a: Search Knowledge Flywheel

**Skip unless `--deep`, `--mixed`, or `--debate`.**

```bash
if command -v ao &>/dev/null; then
    ao know search "plan validation lessons <goal>" 2>/dev/null | head -10
fi
```
If ao returns prior plan review findings, include them as context for the council packet. Skip silently if ao is unavailable or returns no results.

### Step 1b: Check for Product Context

**Skip unless `--deep`, `--mixed`, or `--debate`.**

```bash
if [ -f PRODUCT.md ]; then
  # PRODUCT.md exists — include product perspectives alongside plan-review
fi
```

When `PRODUCT.md` exists in the project root AND the user did NOT pass an explicit `--preset` override:
1. Read `PRODUCT.md` content and include in the council packet via `context.files`
2. Add a single consolidated `product` perspective to the council invocation:
   ```
   $council --preset=plan-review --perspectives="product" validate <plan-path>
   ```
   This yields 3 judges total (2 plan-review + 1 product). The product judge covers user-value, adoption-barriers, and competitive-position in a single review.
3. With `--deep`: 5 judges (4 plan-review + 1 product).

When `PRODUCT.md` exists BUT the user passed an explicit `--preset`: skip product auto-include (user's explicit preset takes precedence).

When `PRODUCT.md` does not exist: proceed to Step 2 unchanged.

> **Tip:** Create `PRODUCT.md` from `docs/PRODUCT-TEMPLATE.md` to enable product-aware plan validation.

### Step 2: Run Council Validation

**Default (inline, no spawning):**
```
$council --quick validate <plan-path>
```
Single-agent structured review. Catches real implementation issues at ~10% of full council cost. Sufficient for most plans (proven across 6+ epics).

**With --deep (4 judges with plan-review perspectives):**
```
$council --deep --preset=plan-review validate <plan-path>
```
Spawns 4 judges:
- `missing-requirements`: What's not in the spec that should be? What questions haven't been asked?
- `feasibility`: What's technically hard or impossible here? What will take 3x longer than estimated?
- `scope`: What's unnecessary? What's missing? Where will scope creep?
- `spec-completeness`: Are boundaries defined? Do conformance checks cover all acceptance criteria? Is the plan mechanically verifiable?

Use `--deep` for high-stakes plans (migrations, security, multi-service, 7+ issues).

**With --mixed (cross-vendor):**
```
$council --mixed --preset=plan-review validate <plan-path>
```
3 Claude + 3 Codex agents for cross-vendor plan validation with plan-review perspectives.

**With explicit preset override:**
```
$pre-mortem --preset=architecture path/to/PLAN.md
```
Explicit `--preset` overrides the automatic plan-review preset. Uses architecture-focused personas instead.

**With explorers:**
```
$council --deep --preset=plan-review --explorers=3 validate <plan-path>
```
Each judge spawns 3 explorers to investigate aspects of the plan's feasibility against the codebase. Useful for complex migration or refactoring plans.

**With debate mode:**
```
$pre-mortem --debate
```
Enables adversarial two-round review for plan validation. Use for high-stakes plans where multiple valid approaches exist. See `$council` docs for full --debate details.

### Step 3: Interpret Council Verdict

| Council Verdict | Pre-Mortem Result | Action |
|-----------------|-------------------|--------|
| PASS | Ready to implement | Proceed |
| WARN | Review concerns | Address warnings or accept risk |
| FAIL | Not ready | Fix issues before implementing |

### Step 4: Write Pre-Mortem Report

**Write to:** `.agents/council/YYYY-MM-DD-pre-mortem-<topic>.md`

```markdown
# Pre-Mortem: <Topic>

**Date:** YYYY-MM-DD
**Plan/Spec:** <path>

## Council Verdict: PASS / WARN / FAIL

| Judge | Verdict | Key Finding |
|-------|---------|-------------|
| Missing-Requirements | ... | ... |
| Feasibility | ... | ... |
| Scope | ... | ... |

## Shared Findings
- ...

## Concerns Raised
- ...

## Recommendation
<council recommendation>

## Decision Gate

[ ] PROCEED - Council passed, ready to implement
[ ] ADDRESS - Fix concerns before implementing
[ ] RETHINK - Fundamental issues, needs redesign
```

### Step 5: Record Ratchet Progress

```bash
ao work ratchet record pre-mortem 2>/dev/null || true
```

### Step 6: Report to User

Tell the user:
1. Council verdict (PASS/WARN/FAIL)
2. Key concerns (if any)
3. Recommendation
4. Location of pre-mortem report

---

## Integration with Workflow

```
$plan epic-123
    │
    ▼
$pre-mortem                    ← You are here
    │
    ├── PASS → $implement
    ├── WARN → Review, then $implement or fix
    └── FAIL → Fix plan, re-run $pre-mortem
```

---

## Examples

### Validate a Plan (Default — Inline)

**User says:** `$pre-mortem .agents/plans/2026-02-05-auth-system.md`

**What happens:**
1. Agent reads the auth system plan
2. Runs `$council --quick validate <plan-path>` (inline, no spawning)
3. Single-agent structured review finds missing error handling for token expiry
4. Council verdict: WARN
5. Output written to `.agents/council/2026-02-13-pre-mortem-auth-system.md`

**Result:** Fast pre-mortem report with actionable concerns. Use `--deep` for high-stakes plans needing multi-judge consensus.

### Cross-Vendor Plan Validation

**User says:** `$pre-mortem --mixed .agents/plans/2026-02-05-auth-system.md`

**What happens:**
1. Agent runs mixed-vendor council (3 Claude + 3 Codex)
2. Cross-vendor perspectives catch platform-specific issues
3. Verdict: PASS with 2 warnings

**Result:** Higher confidence from cross-vendor validation before committing resources.

### Auto-Find Recent Plan

**User says:** `$pre-mortem`

**What happens:**
1. Agent scans `.agents/plans/` for most recent plan
2. Finds `2026-02-13-add-caching-layer.md`
3. Runs inline council validation (no spawning, ~10% of full council cost)
4. Records ratchet progress

**Result:** Frictionless validation of most recent planning work.

### Deep Review for High-Stakes Plan

**User says:** `$pre-mortem --deep .agents/plans/2026-02-05-migration-plan.md`

**What happens:**
1. Agent reads the migration plan
2. Searches knowledge flywheel for prior migration learnings
3. Checks PRODUCT.md for product context
4. Runs `$council --deep --preset=plan-review validate <plan-path>` (4 judges)
5. Council verdict with multi-perspective consensus

**Result:** Thorough multi-judge review for plans where the stakes justify spawning agents.

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| Council times out | Plan too large or complex for judges to review in allocated time | Split plan into smaller epics or increase timeout via council config |
| FAIL verdict on valid plan | Judges misunderstand domain-specific constraints | Add context via `--perspectives-file` with domain explanations |
| Product perspectives missing | PRODUCT.md exists but not included in council packet | Verify PRODUCT.md is in project root and no explicit `--preset` override was passed |
| Pre-mortem gate blocks $crank | Epic has 3+ issues and no pre-mortem ran | Run `$pre-mortem` before `$crank`, or use `--skip-pre-mortem` flag (not recommended) |
| Spec-completeness judge warns | Plan lacks Boundaries or Conformance Checks sections | Add SDD sections or accept WARN (backward compatibility — not a failure) |
| Mandatory for epics enforcement | Hook blocks $crank on 3+ issue epic without pre-mortem | Run `$pre-mortem` first, or set `AGENTOPS_SKIP_PRE_MORTEM_GATE=1` to bypass |

---

## See Also

- `skills/council/SKILL.md` — Multi-model validation council
- `skills/plan/SKILL.md` — Create implementation plans
- `skills/vibe/SKILL.md` — Validate code after implementation

## Reference Documents

- [references/enhancement-patterns.md](references/enhancement-patterns.md)
- [references/failure-taxonomy.md](references/failure-taxonomy.md)
- [references/simulation-prompts.md](references/simulation-prompts.md)
- [references/spec-verification-checklist.md](references/spec-verification-checklist.md)

---

## References

### enhancement-patterns.md

# Enhancement Patterns for Spec Simulation

Patterns for transforming findings into concrete spec improvements.

---

## Pattern 1: Schema from Code

**Problem**: Spec has example JSON that doesn't match actual output.

**Before (Bad)**:
```markdown
The API returns status like:
{
  "status": "running",
  "progress": "50%"
}
```

**After (Good)**:
```markdown
## Status Response Schema

| Field | Type | Values | Description |
|-------|------|--------|-------------|
| status | enum | pending, running, complete, failed | Current state |
| progress | object | {completed: int, total: int} | Progress counts |
| error | string? | null or error message | Present only on failure |

Source: `lib/models.py:WaveStatus`
```

**How to apply**:
1. Find the actual model class in code
2. Extract all fields with types
3. Document enum values explicitly
4. Reference source file

---

## Pattern 2: Error Recovery Matrix

**Problem**: Spec says "handle errors appropriately" without specifics.

**Before (Bad)**:
```markdown
If the operation fails, the system will display an error message.
```

**After (Good)**:
```markdown
## Error Recovery Matrix

| Error Type | User Sees | AI Action | Human Action |
|------------|-----------|-----------|--------------|
| Timeout | "Operation timed out" | Retry once | Check cluster health |
| Auth failure | "401 Unauthorized" | Escalate immediately | Refresh credentials |
| Partial sync | "3/5 apps synced" | Show failed apps | Manual sync remaining |
| Network error | "Connection refused" | Retry 3x with backoff | Check connectivity |
```

**How to apply**:
1. List all error types from code
2. Map each to user-visible message
3. Define AI behavior for each
4. Define human escalation for each

---

## Pattern 3: Mandatory Safety Display

**Problem**: Safety information is optional or buried.

**Before (Bad)**:
```markdown
Wave 6 is forward-only and cannot be rolled back.
```

**After (Good)**:
```markdown
## Safety Classification (ALWAYS DISPLAY)

Every command suggestion MUST show safety level:

| Level | Emoji | Meaning | Display Pattern |
|-------|-------|---------|-----------------|
| Safe | :green_circle: | Read-only, no changes | "**Safe** - This only reads data" |
| Caution | :yellow_circle: | Reversible changes | "**Caution** - This makes changes (reversible)" |
| Dangerous | :red_circle: | Irreversible changes | "**DANGEROUS** - Cannot be undone!" |
| Escalate | :warning: | Requires human expert | "**ESCALATE** - Contact platform team" |

**Wave 6 Example**:
:red_circle: **DANGEROUS - FORWARD ONLY**
This wave migrates Crossplane to v2. There is NO rollback procedure.
```

**How to apply**:
1. Define safety level enum
2. Create visual differentiation (emoji, formatting)
3. Make display MANDATORY in spec
4. Show example for each level

---

## Pattern 4: Per-Tool Timeout Configuration

**Problem**: Single timeout for all operations.

**Before (Bad)**:
```markdown
Tool timeout: 300 seconds
```

**After (Good)**:
```markdown
## Timeout Configuration

| Tool | Operation | Timeout | On Timeout |
|------|-----------|---------|------------|
| get_upgrade_status | Query only | 30s | Retry once |
| preview_wave | Dry run | 60s | Show partial results |
| execute_wave | Wave 1-4 | 300s | Continue polling |
| execute_wave | Wave 5 (operators) | 600s | Show operator status |
| execute_wave | Wave 6 (Crossplane) | 900s | Never auto-retry |
| run_diagnostics | Full scan | 120s | Limit to critical issues |
```

**How to apply**:
1. Profile actual operation durations
2. Add 2x buffer for worst case
3. Define behavior on timeout (retry, partial, escalate)
4. Document differently for different scenarios

---

## Pattern 5: Progress Feedback Specification

**Problem**: Long operations with no feedback.

**Before (Bad)**:
```markdown
Execute the wave and wait for completion.
```

**After (Good)**:
```markdown
## Progress Feedback Requirements

### During Execution
- Status update every 30 seconds minimum
- Show current step: "Approving InstallPlan 3/16..."
- Show elapsed time

### Expected Durations

| Wave | Typical | Maximum | Progress Pattern |
|------|---------|---------|------------------|
| 2-4 | 30s | 2min | Fast, show completion |
| 5 | 5min | 15min | Per-operator status |
| 6 | 10min | 30min | Per-migration step |

### User Guidance
"Wave 5 typically takes 5 minutes. You can check progress with 'get status' or wait for completion notification."
```

**How to apply**:
1. Measure actual durations
2. Define update frequency
3. Specify what to show during wait
4. Give users expected timeframes

---

## Pattern 6: Wave-Specific Handling

**Problem**: All waves treated the same.

**Before (Bad)**:
```markdown
For each wave, call execute_wave with the wave number.
```

**After (Good)**:
```markdown
## Wave-Specific Handling

### Wave 1 (OCP Upgrade)
:red_circle: **HUMAN GATE**
- Requires explicit `human_approved: true` flag
- Display: Full impact assessment before any action
- Duration: 2+ hours, suggest monitoring link
- Never auto-execute

### Wave 6 (Crossplane v2)
:red_circle: **FORWARD ONLY**
- No rollback procedure exists
- Pre-flight: Verify backup exists
- Confirm: Require typed "I understand there is no rollback"
- Post-execution: Verify migration succeeded

### Waves 2-5, 7-10
:yellow_circle: **STANDARD WAVES**
- Preview available (dry-run)
- Confirm before execution
- Rollback guidance available if failed
```

**How to apply**:
1. Identify waves with special requirements
2. Document specific pre-conditions
3. Define confirmation patterns
4. Include post-execution verification

---

## Pattern 7: Escalation Flow

**Problem**: AI tries to handle everything, gets stuck.

**Before (Bad)**:
```markdown
The assistant will help troubleshoot any issues.
```

**After (Good)**:
```markdown
## Escalation Flow

### When to Escalate
1. Error not in known issues database
2. RAG returns no relevant results (score < 0.7)
3. Same error after 2 fix attempts
4. User requests human help
5. Dangerous operation without clear procedure

### Escalation Response Template
:warning: **I need human expertise for this**

I've encountered [specific issue] that I'm not confident handling.

**What I tried:**
- [Action 1 and result]
- [Action 2 and result]

**What I found in docs:**
- [Relevant doc or "no matching documentation"]

**Recommended escalation:**
- Contact: Platform team (#platform-support)
- Include: [specific diagnostic output]
```

**How to apply**:
1. Define escalation triggers
2. Create template response
3. Require "what I tried" summary
4. Provide specific contact/channel

---

## Pattern 8: Audit Trail Requirements

**Problem**: Can't investigate what happened.

**Before (Bad)**:
```markdown
Tool results are displayed to the user.
```

**After (Good)**:
```markdown
## Audit Trail Requirements

### Session Tracking
- Generate unique session_id on first tool call
- Include session_id in all tool calls and responses

### Required Log Fields

| Field | Type | Example |
|-------|------|---------|
| timestamp | ISO8601 | 2026-01-22T10:30:00Z |
| session_id | UUID | abc123-def456 |
| tool_name | string | execute_wave |
| input_params | object | {wave: 5, confirm: true} |
| output_summary | string | "Wave 5 complete, 16 operators updated" |
| duration_ms | int | 45000 |
| safety_level | enum | caution |

### Export Format
```json
{
  "session_id": "abc123",
  "started": "2026-01-22T10:00:00Z",
  "events": [...]
}
```
```

**How to apply**:
1. Define required fields
2. Specify format (JSON, structured log)
3. Include correlation IDs
4. Enable export for post-mortem

---

## Applying Patterns

When enhancing a spec:

1. **Match finding to pattern**: Which pattern addresses this failure mode?
2. **Extract concrete details**: What are the actual values, not placeholders?
3. **Apply pattern**: Copy structure, fill in specifics
4. **Verify completeness**: Does this fully address the failure mode?

### Enhancement Checklist

For each finding:
- [ ] Pattern identified
- [ ] Concrete values extracted from code/testing
- [ ] Pattern applied with specifics
- [ ] Cross-referenced with related sections
- [ ] Example included showing pattern in use

### failure-taxonomy.md

# Failure Taxonomy for Spec Validation

Comprehensive catalog of failure modes to check during pre-mortem validation.

---

## How to Use This Taxonomy

When running pre-mortem, use each category as a checklist item:

1. For each category (Interface Mismatch, Timing, Error Handling, etc.)
2. Ask the Detection Question against the spec
3. If the answer is "no" or "unclear", that's a GAP
4. Apply the Enhancement Pattern to fix it

The taxonomy covers 10 categories. Minimum viable pre-mortem covers at least:
- **Interface Mismatch** - API/schema defined?
- **Error Handling** - Error states defined?
- **Safety** - Rollback defined?
- **Integration** - Dependencies defined?

For comprehensive validation, walk through all 10 categories. See `enhancement-patterns.md` for how to fix gaps.

---

## Category 1: Interface Mismatch

**Description**: What the spec says vs what the system actually does.

| Failure Mode | Detection Question | Enhancement Pattern |
|--------------|-------------------|---------------------|
| Wrong JSON schema | "What does the actual output look like?" | Extract schema from code |
| Missing fields | "What fields are we assuming exist?" | Document all expected fields |
| Different types | "Is this a string or enum?" | Add type constraints |
| Versioning issues | "What if API version changes?" | Add version handling |

**Simulation Prompt**: "What if I actually run this command right now and compare output to spec?"

---

## Category 2: Timing & Performance

**Description**: Operations take longer or behave differently under load.

| Failure Mode | Detection Question | Enhancement Pattern |
|--------------|-------------------|---------------------|
| Timeout | "What if this takes 10x longer?" | Per-operation timeouts |
| Race condition | "What if two requests overlap?" | Add locking/ordering |
| Resource exhaustion | "What if we hit rate limits?" | Add backoff/retry |
| Cascading delays | "What if dependency is slow?" | Add circuit breakers |

**Simulation Prompt**: "What happens if I run this during peak load with degraded network?"

---

## Category 3: Error Handling

**Description**: What happens when things go wrong.

| Failure Mode | Detection Question | Enhancement Pattern |
|--------------|-------------------|---------------------|
| Unclear error message | "Can user understand this?" | Add actionable messages |
| Missing recovery | "What does user do after error?" | Add recovery steps |
| Silent failure | "How do we know this failed?" | Add explicit error states |
| Partial failure | "What if step 3 of 5 fails?" | Add checkpoint/resume |

**Simulation Prompt**: "What if every external call fails? What does the user see?"

---

## Category 4: Safety & Security

**Description**: Dangerous operations without adequate protection.

| Failure Mode | Detection Question | Enhancement Pattern |
|--------------|-------------------|---------------------|
| Missing confirmation | "Can this delete prod data?" | Add explicit confirm gate |
| Unclear severity | "Does user know this is dangerous?" | Add visual safety levels |
| No rollback | "What if we need to undo?" | Document rollback procedure |
| Privilege escalation | "Can this exceed permissions?" | Add permission checks |

**Simulation Prompt**: "What's the worst thing a user could do by accident?"

---

## Category 5: User Experience

**Description**: How users interact with the system.

| Failure Mode | Detection Question | Enhancement Pattern |
|--------------|-------------------|---------------------|
| Skipped instructions | "What if user doesn't read?" | Put warnings before actions |
| Confusing flow | "Is the next step obvious?" | Add explicit next actions |
| Missing feedback | "Does user know it's working?" | Add progress indicators |
| Information overload | "Is this scannable?" | Limit to 2-3 sentences |

**Simulation Prompt**: "What if the user is stressed and just wants to copy-paste?"

---

## Category 6: Integration Points

**Description**: Interactions with external systems.

| Failure Mode | Detection Question | Enhancement Pattern |
|--------------|-------------------|---------------------|
| Dependency unavailable | "What if API is down?" | Add fallback behavior |
| Changed behavior | "What if upstream updates?" | Version pin dependencies |
| Auth failure | "What if token expires?" | Add re-auth flow |
| Data format change | "What if schema evolves?" | Add schema validation |

**Simulation Prompt**: "What if every external system is having a bad day?"

---

## Category 7: State Management

**Description**: Keeping track of where we are.

| Failure Mode | Detection Question | Enhancement Pattern |
|--------------|-------------------|---------------------|
| Lost state | "What if session ends mid-operation?" | Add checkpointing |
| Inconsistent state | "What if DB and cache differ?" | Add reconciliation |
| Stale state | "What if data changed since read?" | Add refresh/optimistic locking |
| Orphaned resources | "What if create succeeds but record fails?" | Add cleanup procedures |

**Simulation Prompt**: "What if power goes out halfway through?"

---

## Category 8: Documentation Gap

**Description**: Spec doesn't match reality.

| Failure Mode | Detection Question | Enhancement Pattern |
|--------------|-------------------|---------------------|
| Outdated example | "Does this actually work?" | Test all examples |
| Missing prerequisite | "What else needs to be true?" | Document prerequisites |
| Implicit assumption | "What am I assuming is already done?" | Make assumptions explicit |
| Wrong version | "Does this work on current version?" | Add version requirements |

**Simulation Prompt**: "Could a new team member follow this spec from scratch?"

---

## Category 9: Tooling & CLI

**Description**: Command-line and tool behavior.

| Failure Mode | Detection Question | Enhancement Pattern |
|--------------|-------------------|---------------------|
| Different flags | "Are these the actual flags?" | Verify against --help |
| Path issues | "What if running from different dir?" | Use absolute paths |
| Missing tools | "Is this tool installed?" | Add tool prerequisites |
| Output format varies | "Is output consistent?" | Parse defensively |

**Simulation Prompt**: "What if I run this on a fresh machine?"

---

## Category 10: Operational

**Description**: Running in production.

| Failure Mode | Detection Question | Enhancement Pattern |
|--------------|-------------------|---------------------|
| No audit trail | "Can we investigate later?" | Add structured logging |
| Missing metrics | "How do we know it's healthy?" | Add observability |
| No runbook | "What do we do at 2 AM?" | Add troubleshooting guide |
| Unclear ownership | "Who gets paged?" | Add escalation path |

**Simulation Prompt**: "What if this breaks on Sunday at 3 AM?"

---

## Quick Reference

During validation, for each category:

| Step | Action |
|------|--------|
| 1 | Ask the Detection Question against the spec |
| 2 | Answer: yes (present), no (missing), partial (incomplete) |
| 3 | If missing/partial: log as GAP with line number |
| 4 | Apply Enhancement Pattern from `enhancement-patterns.md` |

Not every category will yield findings for every spec. Focus on categories relevant to your spec's domain.

### simulation-prompts.md

# Simulation Prompts for Spec Simulation

Prompts to ask yourself during each iteration of spec simulation.

---

## The Core 10 Prompts

Use these to drive each iteration:

### 1. Input Validation
> "What if the input isn't what we expect?"

- Wrong format
- Missing fields
- Null values
- Out of range
- Malicious input

### 2. Dependency Failure
> "What if the external dependency fails?"

- API returns 500
- Service is down
- Network timeout
- Auth token expired
- Rate limited

### 3. Scale Issues
> "What if this takes 10x longer or 10x more resources?"

- Slow network
- Large dataset
- Many concurrent users
- Resource exhaustion
- Memory pressure

### 4. User Behavior
> "What if the user skips reading instructions?"

- Copy-paste without reading
- Clicks confirm without checking
- Ignores warnings
- Does steps out of order
- Cancels mid-operation

### 5. Rollback Need
> "What if we need to undo this?"

- Partial completion
- Wrong environment
- Changed requirements
- Bug discovered after
- User regret

### 6. Debugging Scenario
> "What if we're debugging this at 2 AM?"

- No access to original user
- Logs are unclear
- State is inconsistent
- Multiple possible causes
- Time pressure

### 7. Partial Failure
> "What happens on partial failure?"

- Step 3 of 5 fails
- Some items succeed, some fail
- Inconsistent state
- Unknown progress
- Recovery unclear

### 8. Repetition
> "What if the user does this 100 times?"

- Accumulating errors
- Resource leaks
- State drift
- Performance degradation
- User fatigue

### 9. Environment Difference
> "What if the environment is different?"

- Different version
- Missing tool
- Different permissions
- Different config
- Different network

### 10. Audit & Compliance
> "What does the audit trail look like?"

- Who did what when
- What was the state before/after
- Why was this action taken
- Can we prove compliance
- Can we investigate later

---

## Domain-Specific Prompts

### For API Specs

- "What if the client sends an older API version?"
- "What if response size exceeds limits?"
- "What if pagination is needed?"
- "What if the API is called twice rapidly?"

### For CLI Tool Specs

- "What if run from wrong directory?"
- "What if environment variables missing?"
- "What if output is piped vs TTY?"
- "What if user Ctrl+C mid-execution?"

### For Workflow Specs

- "What if user skips a required step?"
- "What if workflow is interrupted and resumed?"
- "What if two users run simultaneously?"
- "What if approval times out?"

### For Integration Specs

- "What if the other system's schema changed?"
- "What if webhook delivery fails?"
- "What if timestamps are in different zones?"
- "What if retry creates duplicates?"

### For AI/LLM Specs

- "What if the model hallucinates?"
- "What if context window exceeded?"
- "What if model refuses the request?"
- "What if RAG returns wrong context?"

---

## Iteration Structure

For each prompt, document:

```markdown
## Iteration N: [Category]

**Prompt used**: "[The question asked]"

**Scenario imagined**:
[Specific failure scenario in detail]

**What goes wrong**:
- [Specific symptom 1]
- [Specific symptom 2]

**Root cause**:
[Why this happens]

**Lesson learned**:
[What assumption was wrong or what was missing]

**Enhancement needed**:
- [ ] [Concrete spec change with details]
```

---

## Quick Iteration Guide

| Iteration | Primary Focus | Secondary Focus |
|-----------|---------------|-----------------|
| 1 | Input validation | Interface mismatch |
| 2 | Timeout/performance | Scale issues |
| 3 | Error handling | Dependency failure |
| 4 | Safety/security | Rollback need |
| 5 | User experience | User behavior |
| 6 | Integration | Environment difference |
| 7 | State management | Partial failure |
| 8 | Documentation | Debugging scenario |
| 9 | Tooling/CLI | Repetition |
| 10 | Operational | Audit & compliance |

---

## Severity Assessment

After each iteration, assess:

| Question | If Yes → |
|----------|----------|
| Blocks basic functionality? | Critical |
| Causes data loss? | Critical |
| Significant UX degradation? | Important |
| Hard to debug/fix later? | Important |
| Would be nice to have? | Nice-to-have |

---

## Output Summary Template

After all iterations:

```markdown
# Simulation Summary

**Spec**: [Name]
**Iterations**: 10
**Date**: YYYY-MM-DD

## Findings by Severity

### Critical (4)
1. Iteration 1: [Brief description]
2. Iteration 4: [Brief description]
3. Iteration 6: [Brief description]
4. Iteration 7: [Brief description]

### Important (3)
5. Iteration 2: [Brief description]
6. Iteration 5: [Brief description]
7. Iteration 9: [Brief description]

### Nice-to-Have (3)
8. Iteration 3: [Brief description]
9. Iteration 8: [Brief description]
10. Iteration 10: [Brief description]

## Top Enhancements Needed

1. [Most impactful enhancement]
2. [Second most impactful]
3. [Third most impactful]

## Questions to Answer Before Implementation

1. [Question from simulation that needs real answer]
2. [Another question]
```

### spec-verification-checklist.md

# Spec Verification Checklist

Use this checklist to verify spec completeness before implementation.

## Mandatory Items

Every spec MUST have answers to these questions:

### 1. Interface Definition
- [ ] Input format defined (schema/types)
- [ ] Output format defined
- [ ] Error response format defined
- [ ] API versioning strategy

### 2. Error Handling
- [ ] What errors can occur?
- [ ] How is each error communicated?
- [ ] What should user do for each error?
- [ ] Retry logic (if applicable)

### 3. Timing
- [ ] Timeout values specified
- [ ] Rate limits (if applicable)
- [ ] Expected latency bounds
- [ ] What happens on timeout?

### 4. Safety
- [ ] Destructive operations require confirmation
- [ ] Rollback procedure documented
- [ ] Data backup strategy (if applicable)
- [ ] Permission requirements

### 5. Dependencies
- [ ] External services listed
- [ ] Version requirements
- [ ] Fallback behavior if dependency unavailable
- [ ] Authentication/authorization requirements

### 6. State Management
- [ ] Initial state defined
- [ ] State transitions listed
- [ ] How to recover from inconsistent state
- [ ] Cleanup procedures

## Verification Template

| Category | Checklist Item | Present? | Location | Notes |
|----------|----------------|----------|----------|-------|
| Interface | Input schema | yes/no | line N | |
| Interface | Output schema | yes/no | line N | |
| Interface | Error format | yes/no | line N | |
| Error | Error list | yes/no | line N | |
| Error | Recovery steps | yes/no | line N | |
| Timing | Timeouts | yes/no | line N | |
| Timing | Rate limits | yes/no | line N | |
| Safety | Rollback | yes/no | line N | |
| Safety | Confirmation | yes/no | line N | |
| Deps | Dep list | yes/no | line N | |
| Deps | Fallbacks | yes/no | line N | |
| State | Initial state | yes/no | line N | |
| State | Transitions | yes/no | line N | |

## Gap Severity

| Missing Item | Severity | Rationale |
|--------------|----------|-----------|
| Rollback procedure | CRITICAL | Can't recover from failures |
| Error handling | CRITICAL | Users stranded on errors |
| Input validation | HIGH | Security and reliability risk |
| Timeouts | HIGH | Can hang indefinitely |
| Dependencies | HIGH | Silent failures when deps unavailable |
| Rate limits | MEDIUM | Performance issues at scale |
| Cleanup procedures | MEDIUM | Resource leaks |
| Version strategy | LOW | Future compatibility |

## Quick Reference

**Minimum Viable Spec** must have:
1. Input/output schema (what goes in, what comes out)
2. Error handling (what can go wrong, what user does)
3. Rollback procedure (how to undo)
4. Dependencies (what this needs to work)

If any of these 4 are missing → CRITICAL gap, spec is not ready for implementation.


---

## Scripts

### validate.sh

```bash
#!/usr/bin/env bash
set -euo pipefail
SKILL_DIR="$(cd "$(dirname "$0")/.." && pwd)"
PASS=0; FAIL=0

check() { if bash -c "$2"; then echo "PASS: $1"; PASS=$((PASS + 1)); else echo "FAIL: $1"; FAIL=$((FAIL + 1)); fi; }

check "SKILL.md exists" "[ -f '$SKILL_DIR/SKILL.md' ]"
check "SKILL.md has YAML frontmatter" "head -1 '$SKILL_DIR/SKILL.md' | grep -q '^---$'"
check "SKILL.md has name: pre-mortem" "grep -q '^name: pre-mortem' '$SKILL_DIR/SKILL.md'"
check "references/ directory exists" "[ -d '$SKILL_DIR/references' ]"
check "references/ has at least 3 files" "[ \$(ls '$SKILL_DIR/references/' | wc -l) -ge 3 ]"
check "SKILL.md mentions $council delegation" "grep -qi '$council' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions plan-review preset" "grep -qi 'plan-review' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions PASS/WARN/FAIL verdicts" "grep -q 'PASS.*WARN.*FAIL\|PASS | WARN | FAIL' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions .agents/council/ output path" "grep -q '\.agents/council/' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions pre-mortem report format" "grep -qi 'pre-mortem report\|Pre-Mortem:' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions --deep mode" "grep -q '\-\-deep' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions --mixed mode" "grep -q '\-\-mixed' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions --debate mode" "grep -q '\-\-debate' '$SKILL_DIR/SKILL.md'"

echo ""; echo "Results: $PASS passed, $FAIL failed"
[ $FAIL -eq 0 ] && exit 0 || exit 1
```


