# Contract: Too Few Invariants

---

```yaml
# --- Contract Frontmatter ---
issue:      ag-abc.1
framework:  shell
category:   ci
```

---

## Problem

Only two invariants — intentionally below the minimum of 3.

## Invariants

1. First invariant.
2. Second invariant.

## Test Cases

| # | Input | Expected | Validates Invariant |
|---|-------|----------|---------------------|
| 1 | input-a | success | #1 |
| 2 | input-b | error | #2 |
| 3 | input-c | success | #1 |
