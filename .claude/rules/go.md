# Go Conventions

> Canonical source with full examples: `skills/standards/references/go.md`
> This file is kept self-contained for sessions that don't invoke skills.

## Complexity Budget

- Warn at cyclomatic complexity 15, fail at 25.
- Run `golangci-lint run` to check.

## Before Committing Go Changes

```bash
cd cli && go build ./... && go vet ./... && go test ./...
```

Or equivalently: `cd cli && make build && make test`

## Testing (AI-Native Test Shape)

**L2 first, L1 always.** Write L2 integration tests first (where bugs are found), then L1 unit tests for regression safety. AI agents write both. See `skills/standards/references/test-pyramid.md` for the full AI-native test shape.

- Test file naming: `<source>_test.go` (e.g., `goals_test.go`). NEVER `cov*_test.go` or `*_extra_test.go`.
- Test function naming: `Test<Uppercase>` (e.g., `TestFoo_Bar`). Go requires uppercase after `Test`.
- No coverage-padding tests: trivial `!= ""` or `!= nil` assertions are banned.
- No zero-assertion smoke tests: every test must assert behavioral correctness, not just "doesn't panic".
- For print/output functions, use `captureStdout` and assert output contains expected strings.
- Assert exact expected values (`== expected`), not just "not the wrong one" (`!= wrong`).
- Prefer table-driven tests for multi-case functions.
- Test low-level functions directly; don't depend on external CLIs (`bd`, `ao`) in tests.
- **Prefer L2 integration tests** that call a command/workflow entry point over L1 tests that mock dependencies.

## Error Handling

- Always check errors: `if err != nil`.
- Wrap with context: `fmt.Errorf("doing X: %w", err)`.
- Use `errors.Is(err, target)` not `err == target`.

## Struct Fields

- When adding a field, grep all `StructName{` literals and verify each sets the new field.
- Check factory functions and synthesized/summary instances.

## Style

- `gofmt` is automatic. All exported symbols must have godoc comments.
- Accept interfaces, return structs. Keep interfaces small (1-3 methods).
- Detect Go version from `go.mod`; never use features from newer versions.
