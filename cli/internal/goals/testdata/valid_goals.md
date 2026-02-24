# Goals

Ship reliable, well-tested software that improves continuously.

## North Stars

- Automated quality gates catch regressions before merge
- Every change improves or maintains fitness score

## Anti Stars

- Manual testing as the only safety net
- Shipping without running the gate suite

## Directives

### 1. Expand Test Coverage

Add integration tests for all CLI subcommands that currently lack them.

**Steer:** increase

### 2. Reduce Complexity Budget

Refactor functions exceeding cyclomatic complexity 15 into smaller units.

**Steer:** decrease

## Gates

| ID | Check | Weight | Description |
|----|-------|--------|-------------|
| build-passing | `cd cli && make build` | 8 | CLI builds without errors |
| test-passing | `cd cli && make test` | 7 | All unit tests pass |
