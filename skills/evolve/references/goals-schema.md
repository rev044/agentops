# GOALS.yaml Schema

```yaml
version: 1
mission: "What this repo does"

goals:
  - id: unique-identifier
    description: "Human-readable description"
    check: "shell command — exit 0 = pass, non-zero = fail"
    weight: 1-10  # Higher = fix first
```

Goals are checked in weight order (highest first). The first failing goal with the highest weight is selected for improvement.
