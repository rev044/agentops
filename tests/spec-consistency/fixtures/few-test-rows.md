# Contract: Too Few Test Rows

---

```yaml
# --- Contract Frontmatter ---
issue:      ag-abc.2
framework:  python
category:   feature
```

---

## Problem

Only two test rows — intentionally below the minimum of 3.

## Invariants

1. Property A always holds.
2. Property B always holds.
3. Property C always holds.

## Test Cases

| # | Input | Expected | Validates Invariant |
|---|-------|----------|---------------------|
| 1 | input-a | success | #1 |
| 2 | input-b | error | #2 |
