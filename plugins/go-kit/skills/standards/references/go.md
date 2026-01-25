# Go Style Guide - Tier 1 Quick Reference

<!-- Tier 1: Generic standards (~5KB), always loaded -->
<!-- Tier 2: Deep standards in vibe/references/go-standards.md (~20KB), loaded on --deep -->
<!-- Last synced: 2026-01-21 -->

> **Purpose:** Quick reference for Go coding standards. For comprehensive patterns, load Tier 2.

---

## Quick Reference

| Standard | Value | Validation |
|----------|-------|------------|
| **Go Version** | 1.21+ | `go version` |
| **Formatter** | `gofmt` / `goimports` | `gofmt -l .` |
| **Linter** | `golangci-lint` | `.golangci.yml` at repo root |
| **Module** | Required | `go.mod` in project root |
| **Complexity** | CC ≤ 10 | `gocyclo -over 10 ./...` |

### Version-Specific Features

| Go Version | Feature | Example |
|------------|---------|---------|
| **1.21+** | `log/slog`, `min()`, `max()`, `clear()` | `slog.Info("msg", "key", val)` |
| **1.22+** | Range over integers | `for i := range 10 { }` |
| **1.23+** | Iterator functions (`iter` package) | `for v := range seq { }` |

---

## golangci-lint (Minimum)

```yaml
# .golangci.yml
linters:
  enable:
    - errcheck      # Check error returns
    - govet         # Vet examines Go source
    - staticcheck   # Static analysis
    - gosimple      # Simplify code
    - ineffassign   # Detect ineffective assignments
    - unused        # Check for unused code
    - gofmt         # Check formatting
    - errorlint     # Error wrapping checks (%w vs %v)
```

---

## Common Errors

| Symptom | Cause | Fix |
|---------|-------|-----|
| `undefined: X` | Missing import or typo | Check imports, spelling |
| `cannot use X as Y` | Type mismatch | Check interface compliance |
| `nil pointer dereference` | Uninitialized pointer | Add nil check before use |
| `deadlock` | Goroutine waiting forever | Check channel/mutex usage |
| `race detected` | Data race | Use mutex or channels |
| `context canceled` | Parent context done | Handle `ctx.Err()` |
| `go.mod outdated` | Dependency drift | Run `go mod tidy` |

---

## Anti-Patterns

| Name | Pattern | Instead |
|------|---------|---------|
| Naked Returns | `return` without values | Explicit: `return result, nil` |
| Init Abuse | Heavy logic in `init()` | Explicit initialization |
| Interface Pollution | Interfaces with 10+ methods | Small interfaces, compose |
| Error Strings | `errors.New("User not found")` | Sentinel errors or types |
| Ignoring Errors | `result, _ := fn()` | Handle or document with `//nolint:errcheck` |
| Wrapping with %v | `fmt.Errorf("x: %v", err)` | Use `%w`: `fmt.Errorf("x: %w", err)` |
| Premature Channel | Channels for simple sync | Use mutex for simple cases |
| Global Mutable State | `var globalConfig *Config` | Pass config as dependency |
| fmt.Print in Libraries | Debug statements in production | Use `log/slog` structured logging |

---

## Summary Checklist

| Category | Requirement |
|----------|-------------|
| **Tooling** | Go 1.21+, `gofmt`, `golangci-lint` |
| **Formatting** | All code passes `gofmt -l .` |
| **Linting** | All code passes `golangci-lint run` |
| **Complexity** | CC ≤ 10 per function |
| **Errors** | Wrap with `%w`, use `errors.Is`/`errors.As` |
| **Interfaces** | Accept interfaces, return structs |
| **Concurrency** | Use `context.Context` for cancellation |
| **Logging** | Use `log/slog` (Go 1.21+) |
| **Tests** | Table-driven tests with subtests |

---

## Complexity Grades

| Grade | CC Range | Action |
|-------|----------|--------|
| A | 1-5 | Ideal |
| B | 6-10 | Acceptable |
| C | 11-15 | Refactor when touching |
| D | 16-20 | Must refactor |
| F | 21+ | Block merge |

---

## Talos Prescan Checks

| Check | Pattern | Rationale |
|-------|---------|-----------|
| PRE-002 | TODO/FIXME markers | Track technical debt |
| PRE-003 | Complexity > 15 | High CC correlates with bugs |
| PRE-004 | Empty error blocks | Silent failures |
| PRE-007 | fmt.Print in libraries | Use structured logging |
| PRE-013 | `_ = err` without nolint | Explicit acknowledgment required |
| PRE-014 | `%v` instead of `%w` | Preserves error chain |

---

## JIT Loading

**Tier 2 (Deep Standards):** For comprehensive patterns including:
- Concurrency (mutex, channels, errgroup, worker pools)
- Testing (mocking, benchmarks, profiling)
- Security (HMAC, timing attacks, TLS)
- HTTP API patterns
- K8s Operator patterns
- Configuration management

Load: `vibe/references/go-standards.md`

**Quick Reference Examples:** `vibe/references/go-patterns.md`
