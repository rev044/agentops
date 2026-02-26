# Audit Report Template

**Write to:** `.agents/research/YYYY-MM-DD-bug-<scope-slug>.md`

```markdown
# Bug Hunt: <Scope Description>

**Date:** YYYY-MM-DD
**Scope:** <files/directories audited>
**Failures:** N (hypothesis tests that didn't confirm)

## Summary

| # | Bug | Severity | File | Fix |
|---|-----|----------|------|-----|
| 1 | <short description> | HIGH | `<file:line>` | <proposed fix> |
| 2 | <short description> | MEDIUM | `<file:line>` | <proposed fix> |

## Findings

### BUG-1: <Title> (SEVERITY)

**File:** `<path:line>`
**Root cause:** <what's wrong and why>

**Observed:** <concrete evidence — error output, test failure, code path trace>

**Fix:** <specific change needed>

### BUG-2: <Title> (SEVERITY)

...
```

## Severity Criteria

| Severity | Criteria | Examples |
|----------|----------|---------|
| **HIGH** | Data loss, security, resource leak, process orphaning | Zombie processes, SQL injection, file handle leak |
| **MEDIUM** | Wrong output, incorrect defaults, silent data corruption | UTF-8 truncation, hardcoded paths, wrong error code |
| **LOW** | Dead code, cosmetic, minor inconsistency | Unreachable branch, unused import, style violation |
