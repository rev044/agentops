---
description: Isolate→Locate→Fix with historical pattern search and interactive JIT loading
---

# /prime-debug - Debugging & Troubleshooting Context Router

**Purpose:** Guide systematic debugging using historical patterns and JIT context loading.

**Workflow:** Search History → Isolate → Locate → Fix → Verify → Document

**Token budget:** <40% total (80k of 200k for entire session)

---

## Step 1: Constitutional Baseline (Always Enforced)

{{cat .claude/CONSTITUTION.md}}

**Status:** ✅ Constitutional foundation loaded (2k tokens)

You are now operating under AgentOps constitution with debug workflow:
- **Search first:** 80% of bugs have been seen before
- **Isolate:** Reproduce consistently
- **Locate:** Find root cause (not symptom)
- **Fix:** Minimal change
- **Document:** Add pattern to codex

**Key insight:** Search institutional memory BEFORE deep investigation. Time saved: hours.

---

## Step 2: Describe the Problem

**What are you debugging?**

Please provide:
1. **Error message** (full text if possible)
2. **When it started** (recent commit? specific action?)
3. **What changed** (git log, recent deployments?)
4. **Impact** (who/what is affected?)

### Common Debug Categories:

**Deployment Issues**
- **ArgoCD sync failure** - App won't sync to cluster
- **Pod crashes** - CrashLoopBackOff, ImagePullBackOff
- **Resource exhaustion** - OOM, CPU throttling

**Configuration Issues**
- **YAML syntax error** - Validation failing
- **Missing values** - Undefined variables
- **Schema violation** - Invalid resource specs

**Runtime Issues**
- **Performance problem** - Slow response, high latency
- **Intermittent failure** - Works sometimes, fails other times
- **Dependency failure** - Service can't reach dependency

**Build/Pipeline Issues**
- **CI/CD failure** - Pipeline won't complete
- **Test failures** - Tests passing locally, failing in CI
- **Build errors** - Compilation or packaging issues

### Other
- Describe your specific problem

---

## Step 3: Historical Pattern Search (Do This FIRST!)

**Before deep investigation, search institutional memory:**

```bash
# Search codex for similar errors (most valuable source)
grep -C 5 "ERROR_KEYWORD" docs/reference/sessions/codex/codex-ops-notebook.md | tail -100

# Check recent fixes in git history
git log --all --grep="fix:" --grep="bug:" --since="90 days ago" --oneline | head -20

# Search for this specific error message
git log --all -S "ERROR_MESSAGE" --oneline

# Look for troubleshooting documentation
grep -r "Troubleshoot|Debug|Fix" docs/how-to/ | grep -i "KEYWORD"
```

**Why search first?**
- 80% of bugs have been seen before
- Pattern reuse saves hours of investigation
- Institutional memory compounds value
- Previous fix might apply directly

**If found:** Load that pattern, apply fix, validate
**If not found:** Proceed with systematic debugging

---

## Step 4: JIT Load Debug Context

**[After you describe the problem, I will:]**

1. Search historical patterns first (codex, git history)
2. If found: Load pattern, guide you to fix
3. If not found: Load relevant troubleshooting pattern
4. Guide systematic debugging workflow

**Examples of what gets loaded:**

→ **Issue:** "ArgoCD sync failing with Helm error"
  Search: codex + git history (10 sec)
  If found: Load previous fix (0.5k)
  If not: Load `argocd-troubleshooting.md` (0.8k)
  Total: 2.8-3.3k (1.4-1.6%)

→ **Issue:** "YAML validation error on line 42"
  Load: `yaml-validation.md` (0.3k)
  Quick fix: Identify syntax error
  Total: 2.3k (1.15%)

→ **Issue:** "Pod crashes with OOM"
  Search: codex for memory issues
  Load: `performance-debugging.md` (0.7k)
  Total: 2.7k (1.35%)

→ **Issue:** "Intermittent 503 errors"
  Search: codex + recent changes
  Load: `performance-debugging.md` (0.7k)
  Total: 2.7k (1.35%)

---

## Step 5: Systematic Debug Workflow

### 1. Isolate (20% of time)
**Goal:** Reproduce consistently

```bash
# Capture error
[Full error message + stack trace]

# Check recent changes
git log --oneline -20
git diff HEAD~5

# Try to reproduce
[Minimal steps to trigger error]
```

**Output:** Reliable reproduction steps

### 2. Locate (40% of time)
**Goal:** Find root cause (not symptom)

```bash
# Narrow down to component
kubectl get events -n NAMESPACE --sort-by='.lastTimestamp' | tail -20
kubectl logs POD_NAME -n NAMESPACE

# Find exact location
grep -r "ERROR_PATTERN" apps/

# Understand why it's happening
[Root cause analysis]
```

**Output:** File:line where problem originates + why

### 3. Fix (20% of time)
**Goal:** Minimal change to fix root cause

```bash
# Make smallest possible change
[Edit specific file:line]

# Don't refactor while debugging
# Don't fix other issues
# Focus on THIS problem only
```

**Output:** Targeted fix

### 4. Verify (15% of time)
**Goal:** Confirm fix works, no regressions

```bash
# Reproduce error (should fail before fix)
[Original reproduction steps]

# Apply fix
[Your changes]

# Verify fix (should pass now)
make quick
make test-app APP=name
make ci-all

# Check for regressions
[Test related functionality]
```

**Output:** Validated fix

### 5. Document (5% of time)
**Goal:** Capture pattern for future

```bash
# Add to codex (for future searches)
[Pattern description in commit]

# Update troubleshooting docs if new
[New pattern → docs/how-to/troubleshooting/]

# Commit with learning
[Context/Solution/Learning/Impact format]
```

**Output:** Institutional memory

---

## Step 6: Debug Metrics (Track Learning)

**Capture these for pattern building:**

```text
Time to isolate:  [X minutes]
Time to locate:   [X minutes]
Time to fix:      [X minutes]
Total:            [X minutes]

Previous similar issue: [commit hash or "first occurrence"]
Pattern applicable elsewhere: [yes/no - where?]
Should be automated: [yes/no - how?]
```

**If locate time > 40% of total:** Problem was poorly understood, needed more research

---

## Token Budget Tracking

```text
Context Window: 200,000 tokens
Target: <40% (80k) for entire debug session

Typical allocation:
  Constitution:       ████░░░░░░  2k/200k   (1%)
  Historical search:  ██░░░░░░░░  1k/200k   (0.5%)
  Debug pattern:      ██░░░░░░░░  0.8k/200k (0.4%)
  Investigation:      ████████████ 30k/200k  (15%)
  Fix & validate:     ████░░░░░░  10k/200k  (5%)
  Reserved:           ░░░░░░░░░░  156k/200k (78%)

Status: 🟢 GREEN - Single-phase debugging
```

**Monitor continuously:**
- 🟢 <35% (70k): GREEN - continue
- ⚡ 35-40% (70-80k): YELLOW - prepare to wrap up
- ⚠️ 40-60% (80-120k): RED - document findings, fresh session
- 🔴 >60% (120k+): CRITICAL - save state, reset immediately

---

## Common Debug Patterns (Quick Reference)

### YAML Syntax Errors
```bash
make quick  # Identifies line number
# Common: indentation, missing quotes, invalid characters
```

### ArgoCD Sync Failures
```bash
make test-app APP=name  # Shows rendering error
argocd app diff APP_NAME  # Shows what's different
# Common: missing config.env values, invalid Helm template
```

### Pod Crashes
```bash
kubectl describe pod POD_NAME -n NAMESPACE
kubectl logs POD_NAME -n NAMESPACE --previous
# Common: OOM, missing env var, failed health check
```

### Performance Issues
```bash
kubectl top pods -n NAMESPACE
kubectl get hpa -n NAMESPACE
# Common: CPU throttling, memory leaks, slow queries
```

---

## What Happens Next?

**I will:**
1. ✅ Search historical patterns first (codex + git)
2. ✅ If found: Guide you to apply previous fix
3. ✅ If not found: Load relevant troubleshooting pattern
4. ✅ Guide you through Isolate→Locate→Fix workflow
5. ✅ Track debug metrics (time breakdown)
6. ✅ Document pattern for future searches

**You do:**
1. Describe the problem (error, when, what changed, impact)
2. Provide error messages and logs
3. Review proposed fix
4. Validate fix works
5. Approve commit with learning

---

## Why Search First?

**Data from our codex:**
- 80% of bugs are repeats (seen before)
- Average time saved: 45 minutes per debug session
- Pattern reuse compounds (faster each time)
- Institutional memory = competitive moat

**Investment:**
- 2 minutes to search
- Potential savings: 45 minutes
- ROI: 22.5x

**"Don't debug, search and apply"**

---

## Related Patterns

- **Historical pattern search:** codex, git log, grep
- **Systematic debugging:** Isolate→Locate→Fix workflow
- **Debug metrics:** Track time distribution
- **Pattern documentation:** Add to codex for future

---

**Ready! What are you debugging?** (Describe the problem or choose from categories above)
