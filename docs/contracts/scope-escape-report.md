# Scope Escape Report Template

When an agent determines a task cannot be completed within its given constraints,
it should produce a structured report instead of forcing a partial or incorrect fix.

## Template

### Task
[Original task description]

### Finding
[What the agent discovered that prevents completion]

### Audit Summary

| Metric | Value |
|--------|-------|
| Sites examined | [count] |
| Fixable within scope | [count] |
| Requires scope expansion | [count] |
| Blocking dependency | [description or "none"] |

### Root Cause
[Why the task exceeds the given scope -- 2-3 sentences]

### Recommended Approach
[What WOULD fix it, at what effort level -- 2-3 sentences]

### Scope Expansion Required
- [ ] Production code changes (list files)
- [ ] API/interface changes
- [ ] Cross-package refactor
- [ ] External dependency update
- [ ] Human decision needed

### Evidence
[Key data points, grep outputs, or code references supporting the finding]
