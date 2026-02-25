---
name: standards
description: 'Language-specific coding standards and validation rules. Provides Python, Go, Rust, TypeScript, Shell, YAML, JSON, and Markdown standards. Auto-loaded by $vibe, $implement, $doc, $bug-hunt, $complexity based on file types.'
---


# Standards Skill

Language-specific coding standards loaded on-demand by other skills.

## Purpose

This is a **library skill** - it doesn't run standalone but provides standards
references that other skills load based on file types being processed.

## Standards Available

| Standard | Reference | Loaded By |
|----------|-----------|-----------|
| Skill Structure | `references/skill-structure.md` | vibe (skill audits), doc (skill creation) |
| Python | `references/python.md` | vibe, implement, complexity |
| Go | `references/go.md` | vibe, implement, complexity |
| Rust | `references/rust.md` | vibe, implement, complexity |
| TypeScript | `references/typescript.md` | vibe, implement |
| Shell | `references/shell.md` | vibe, implement |
| YAML | `references/yaml.md` | vibe |
| JSON | `references/json.md` | vibe |
| Markdown | `references/markdown.md` | vibe, doc |

## How It Works

Skills declare `standards` as a dependency:

```yaml
skills:
  - standards
```

Then load the appropriate reference based on file type:

```python
# Pseudo-code for standard loading
if file.endswith('.py'):
    load('standards/references/python.md')
elif file.endswith('.go'):
    load('standards/references/go.md')
elif file.endswith('.rs'):
    load('standards/references/rust.md')
# etc.
```

## Deep Standards

For comprehensive audits, skills can load extended standards from
`vibe/references/*-standards.md` which contain full compliance catalogs.

| Standard | Size | Use Case |
|----------|------|----------|
| Tier 1 (this skill) | ~5KB each | Normal validation |
| Tier 2 (vibe/references) | ~15-20KB each | Deep audits, `--deep` flag |

## Integration

Skills that use standards:
- `$vibe` - Loads based on changed file types
- `$implement` - Loads for files being modified
- `$doc` - Loads markdown standards
- `$bug-hunt` - Loads for root cause analysis
- `$complexity` - Loads for refactoring recommendations

## Examples

### Vibe Loads Python Standards

**User says:** `$vibe` (detects changed Python files)

**What happens:**
1. Vibe skill checks git diff for file types
2. Vibe finds `auth.py` in changeset
3. Vibe loads `standards/references/python.md` automatically
4. Vibe validates against Python standards (type hints, docstrings, error handling)
5. Vibe reports findings with standard references

**Result:** Python code validated against language-specific standards without manual reference loading.

### Implement Loads Go Standards

**User says:** `$implement ag-xyz-123` (issue modifies Go files)

**What happens:**
1. Implement skill reads issue metadata to identify file targets
2. Implement finds `server.go` in implementation scope
3. Implement loads `standards/references/go.md` for context
4. Implement writes code following Go standards (error handling, naming, package structure)
5. Implement validates output against loaded standards before committing

**Result:** Go code generated conforming to standards, reducing post-implementation vibe findings.

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| Standards not loaded | File type not detected or standards skill missing | Check file extension matches reference; verify standards in dependencies |
| Wrong standard loaded | File type misidentified (e.g., .sh as .bash) | Manually specify standard; update file type detection logic |
| Deep standards missing | Vibe needs extended catalog, not found | Check `vibe/references/*-standards.md` exists; use `--deep` flag |
| Standard conflicts | Multiple languages in same changeset | Load all relevant standards; prioritize by primary language |

## Reference Documents

- [references/common-standards.md](references/common-standards.md)
- [references/examples-troubleshooting-template.md](references/examples-troubleshooting-template.md)
- [references/go.md](references/go.md)
- [references/json.md](references/json.md)
- [references/markdown.md](references/markdown.md)
- [references/python.md](references/python.md)
- [references/rust.md](references/rust.md)
- [references/shell.md](references/shell.md)
- [references/skill-structure.md](references/skill-structure.md)
- [references/typescript.md](references/typescript.md)
- [references/yaml.md](references/yaml.md)

---

## References

### common-standards.md

# Common Standards Catalog - Cross-Language Patterns

**Version:** 1.0.0
**Last Updated:** 2026-02-09
**Purpose:** Universal coding standards shared across all languages. Language-specific files reference this document for philosophical and cross-cutting patterns, keeping language-specific implementation details in their own catalogs.

---

## Table of Contents

1. [Error Handling Philosophy](#error-handling-philosophy)
2. [Testing Best Practices](#testing-best-practices)
3. [Security Principles](#security-principles)
4. [Documentation Standards](#documentation-standards)
5. [Code Organization Principles](#code-organization-principles)
6. [Dedup Manifest](#dedup-manifest)

---

## Error Handling Philosophy

Errors are first-class citizens. Every language has different mechanisms (Result types, exceptions, error returns), but the underlying principles are universal.

### Core Rules

| Rule | ALWAYS | NEVER |
|------|--------|-------|
| Visibility | Log or propagate every error | Suppress errors silently |
| Specificity | Use specific error types/exceptions | Catch-all without re-raising |
| Context | Add context when propagating | Lose the original error chain |
| Recovery | Distinguish recoverable vs fatal | Treat all errors the same |
| Documentation | Document error behavior in public APIs | Assume callers know failure modes |
| Libraries | Log before raising in library boundaries | Swallow errors inside libraries |

### Error Chain Preservation

Every language provides a mechanism for preserving error chains. Use it.

| Language | Mechanism | Example |
|----------|-----------|---------|
| Go | `fmt.Errorf("context: %w", err)` | Preserves `errors.Is()` / `errors.As()` |
| Python | `raise NewError("context") from exc` | Preserves `__cause__` chain |
| Rust | `?` with `.context()` / `#[source]` | Preserves `Error::source()` chain |
| TypeScript | `new AppError("context", { cause: err })` | Preserves `Error.cause` chain |
| Shell | `err "context: $cmd failed"; return $exit_code` | Preserves exit code semantics |

### Intentional Error Ignores

When errors are intentionally ignored (e.g., best-effort cleanup), document the reason:

| Language | Pattern |
|----------|---------|
| Go | `_ = conn.Close() // nolint:errcheck - best effort cleanup` |
| Python | `except SpecificError: pass  # best effort cleanup` with comment |
| Rust | `let _ = conn.close(); // Intentional ignore: best effort cleanup` |
| TypeScript | `void promise.catch(() => {}); // fire-and-forget, logged elsewhere` |
| Shell | `rm -rf "$TMPDIR" 2>/dev/null \|\| true` |

### Error Aggregation

When multiple operations can fail independently (parallel execution, multi-step cleanup), use the language's error aggregation mechanism rather than discarding all but the first error.

| Language | Mechanism |
|----------|-----------|
| Go | `errors.Join(err1, err2)` (1.20+) |
| Python | `ExceptionGroup` (3.11+) |
| Rust | Custom `Vec<Error>` or `anyhow` context chain |
| TypeScript | `AggregateError` |

### Custom Error Hierarchies

Define a base error type per project/crate/package. Subtypes encode categories.

**Principles:**
- Base type enables catch-all at API boundaries
- Subtypes enable programmatic handling by callers
- Machine-readable codes (where applicable) enable telemetry
- Human-readable messages enable debugging

### Severity Classification

| Level | Definition | Action |
|-------|-----------|--------|
| Fatal | Process cannot continue | Log, clean up, exit non-zero |
| Recoverable | Operation failed, process continues | Log, retry or degrade gracefully |
| Warning | Non-ideal but not broken | Log at warning level, continue |
| Informational | Expected alternative path | Log at debug level |

### Anti-Patterns (Universal)

| Anti-Pattern | Why It's Bad | Instead |
|--------------|-------------|---------|
| Silent suppression (`catch {}`, `except: pass`, `_ =` without comment) | Hides bugs, makes debugging impossible | Log, propagate, or document the ignore |
| String-only errors | Not matchable, no programmatic handling | Use typed/structured errors |
| Catching too broadly | Masks unrelated failures | Catch the most specific type possible |
| Logging AND re-raising the same error | Duplicate log entries at every layer | Log at the boundary, propagate elsewhere |
| Panic/throw in library code for expected failures | Crashes callers unexpectedly | Return error types; reserve panic for invariant violations |

---

## Testing Best Practices

### Test Organization

| Layer | Scope | Speed | When to Run |
|-------|-------|-------|-------------|
| Unit | Single function/method | < 100ms | Every commit |
| Integration | Multiple components, real I/O | < 30s | Every PR |
| End-to-end | Full system with real deps | < 5min | Pre-release |
| Property-based | Invariant fuzzing | Varies | CI nightly or on critical paths |

### Table-Driven / Parameterized Tests

The table-driven pattern is universal. Define inputs and expected outputs in a data structure, then iterate.

| Language | Mechanism |
|----------|-----------|
| Go | `[]struct{ name, input, want }` + `t.Run()` |
| Python | `@pytest.mark.parametrize("input,expected", [...])` |
| Rust | `#[test]` with loop or `proptest!` macro |
| TypeScript | `test.each([...])` or `describe.each([...])` |
| Shell | BATS `@test` with parameterized fixtures |

**Benefits:**
- Easy to add new cases (one line per case)
- Clear test naming
- DRY -- assertion logic written once

### Fixtures and Mocking Philosophy

| Principle | ALWAYS | NEVER |
|-----------|--------|-------|
| External boundaries | Mock external services, APIs, databases | Let tests hit real external services in unit tests |
| Internal code | Test real internal implementations | Mock internal functions (couples tests to implementation) |
| Test isolation | Each test sets up its own state | Share mutable state between tests |
| Cleanup | Clean up resources (files, containers, connections) | Leave test artifacts behind |

### Test Double Types

| Type | Purpose | When to Use |
|------|---------|-------------|
| Stub | Returns canned data | Simple happy/sad path |
| Mock | Verifies interactions were called | Behavior verification |
| Fake | Working lightweight implementation | Integration-like tests without real infra |
| Spy | Records calls for later assertion | Interaction counting/ordering |

### Coverage Targets

| Metric | Minimum | Target | Critical Paths |
|--------|---------|--------|----------------|
| Line coverage | 60% | 80% | 90%+ |
| Branch coverage | 50% | 70% | 85%+ |

**Coverage philosophy:**
- Coverage is a floor, not a ceiling -- low coverage signals under-testing, high coverage does not guarantee quality
- Prioritize critical paths (error handling, security, data integrity) over boilerplate
- Measure branch coverage, not just line coverage -- untested branches hide bugs

### Property-Based Testing

Test invariants that must hold for ALL inputs, not just hand-picked examples.

**When to use:**
- Serialization roundtrips (encode then decode = original)
- Mathematical properties (commutativity, associativity)
- Parser contracts (valid input always parses, invalid always fails)
- Boundary conditions (output never exceeds input, no negative values)

### Doc Tests / Example Tests

Code examples in documentation should be executable tests. Guarantees documentation accuracy.

| Language | Mechanism |
|----------|-----------|
| Go | `func Example*` in `_test.go` files |
| Python | Doctest in docstrings, or `>>> ` examples |
| Rust | Code blocks in `///` doc comments |
| TypeScript | JSDoc `@example` blocks (manual verification) |

---

## Security Principles

### No Hardcoded Secrets

| ALWAYS | NEVER |
|--------|-------|
| Load secrets from environment variables or secret stores | Hardcode API keys, tokens, passwords in source |
| Use `.env` files locally (gitignored) | Commit `.env` or credential files |
| Rotate secrets on exposure | Assume secrets are safe in private repos |
| Audit git history for leaked secrets | Rely on `.gitignore` alone for protection |

**Detection:** Prescan pattern P2 flags hardcoded secrets in all languages.

### Input Validation

Validate at system boundaries (user input, external APIs, file reads). Trust internal code within the same trust boundary.

| Rule | Description |
|------|-------------|
| Validate early | Check inputs at the entry point, not deep in business logic |
| Fail fast | Reject invalid input immediately with clear error messages |
| Allowlist over denylist | Define what IS valid, not what ISN'T |
| Type-safe parsing | Parse into typed structures, not raw strings |

### Injection Prevention

| Attack Vector | Prevention |
|---------------|-----------|
| SQL injection | Parameterized queries / prepared statements. NEVER string interpolation. |
| Command injection | Use array-based exec (no shell). Avoid `eval()`, `exec()`, `system()`. |
| Template injection | Use auto-escaping template engines. Escape user input in templates. |
| Path traversal | Resolve to absolute path, verify within allowed directory. Block `..` sequences. |
| JSON/YAML injection | Use proper serialization libraries (e.g., `jq` in shell). NEVER string interpolation for structured formats. |

### Cryptographic Best Practices

| ALWAYS | NEVER |
|--------|-------|
| Use timing-safe comparison for secrets | Use `==` for secret/token comparison |
| Use established crypto libraries | Roll your own cryptography |
| Use strong hash functions (SHA-256+, bcrypt, argon2) | Use MD5 or SHA-1 for security |
| Enforce TLS 1.2+ (prefer 1.3) | Disable certificate verification in production |
| Generate random values with crypto-grade RNG | Use math/random for security-sensitive values |

### Dependency Auditing

| Practice | Frequency |
|----------|-----------|
| Run `audit` command (`npm audit`, `cargo audit`, `pip-audit`, `govulncheck`) | Every CI build |
| Pin dependency versions with lock files | Always committed for applications |
| Review new dependencies before adding | Before merge |
| Monitor for CVEs in transitive dependencies | Automated via Dependabot/Renovate |

### eval/exec/system Avoidance

| Rule | Description |
|------|-------------|
| Avoid `eval()` in all languages | Executes arbitrary code; use structured dispatch instead |
| Avoid shell execution from application code | Use library APIs instead of shelling out |
| If shell execution is unavoidable | Use array-based exec with no interpolation |
| Shell scripts | Avoid `eval` for user-provided data; use functions for dispatch |

---

## Documentation Standards

### What to Document

| Document | Why |
|----------|-----|
| Public API signatures | Callers need to know parameters, return types, error behavior |
| Non-obvious logic | Future readers (including yourself) need to understand WHY, not WHAT |
| Error behavior | Callers must know what can fail and how |
| Security-sensitive decisions | Reviewers need to verify threat model compliance |
| Configuration options | Users need to know defaults, valid ranges, and effects |
| Architecture decisions | Teams need to understand trade-offs and constraints |

### What NOT to Document

| Skip | Why |
|------|-----|
| Obvious code (`i++`, `return nil`) | Comments add noise, not signal |
| Implementation details of private functions | Changes frequently; comments go stale |
| Type information already in signatures | Redundant with the type system |
| "What" the code does (when code is clear) | The code itself is the documentation |

### Examples in Documentation

- Include usage examples for public APIs
- Examples should be runnable (doc tests where supported)
- Show the common case first, edge cases second
- Include error handling in examples

### Keeping Documentation in Sync

| Practice | Description |
|----------|-------------|
| Doc tests | Executable examples catch staleness automatically |
| Review docs with code changes | PR reviews should include doc updates |
| Delete docs for deleted features | Stale docs are worse than no docs |
| Version documentation | Match docs to release versions |

### Cross-Reference Patterns

- Link to related concepts rather than duplicating content
- Use relative paths within a project
- Reference external standards by URL (e.g., RFC numbers, OWASP guides)

---

## Code Organization Principles

### Module/Package Naming

| Convention | Description |
|------------|-------------|
| Short, descriptive names | `config`, `handlers`, `models` -- not `configurationManager` |
| Lowercase with language-appropriate separators | `snake_case` (Python/Rust/Go), `kebab-case` (npm/crate names), `camelCase` (TS) |
| No stuttering | `config.Config` is fine; `config.ConfigConfig` is not |
| Domain-driven grouping | Group by feature/domain, not by technical layer |

### Public vs Private Visibility

| Rule | Description |
|------|-------------|
| Minimize public API surface | Export only what callers need |
| Default to private | Make things public only when required |
| Use explicit re-exports | Control the public API from a single entry point |
| Hide implementation details | Internal helpers, data structures, and algorithms stay private |

### Circular Dependency Avoidance

| Strategy | Description |
|----------|-------------|
| Dependency inversion | Depend on abstractions (interfaces/traits), not implementations |
| Extract shared types | Move shared types to a separate, leaf-level module |
| Event-based decoupling | Use events/callbacks instead of direct cross-module calls |
| Layer discipline | Higher layers depend on lower layers, never the reverse |

### File Size Heuristics

| Size | Status | Action |
|------|--------|--------|
| < 300 lines | Excellent | Maintain |
| 300-500 lines | Acceptable | Monitor |
| 500-800 lines | Warning | Consider splitting |
| 800+ lines | Critical | Split into submodules |

### Version-Aware Development

Language-specific standards SHOULD declare the target language/runtime version and organize modern features by version availability. This prevents using features unavailable in the target version and ensures developers adopt modern alternatives when available.

| Language | Version Source | Example Modern Features |
|----------|---------------|------------------------|
| Go | `go.mod` `go` directive | `slices` (1.21+), `range n` (1.22+), `t.Context()` (1.24+) |
| Python | `pyproject.toml` `requires-python` | `match` (3.10+), `tomllib` (3.11+), exception groups (3.11+) |
| Rust | `Cargo.toml` `edition` | `let-else` (2021+), `async fn in trait` (2024+) |
| TypeScript | `tsconfig.json` `target` | `satisfies` (4.9+), `using` (5.2+) |

### Import Ordering

All languages follow the same conceptual grouping:

1. **Standard library** imports
2. **External/third-party** imports
3. **Internal/project** imports

Separated by blank lines. Alphabetical within each group.

---

## Dedup Manifest

This table maps which sections in each language-specific file contain universal philosophical content that can be replaced with a cross-reference to this document, and which must remain because they contain language-specific implementation details.

| Language File | Section | Action | Rationale |
|---------------|---------|--------|-----------|
| `go.md` | Error Handling | keep-as-is | Go-specific `%w`, `errors.Is()`, `errors.Join()`, custom error types |
| `go.md` | Modern Standard Library | keep-as-is | Entirely Go-specific stdlib packages (`slices`, `maps`, `cmp`) and version-gated features |
| `go.md` | Concurrency | keep-as-is | Go-specific `sync.OnceFunc`, type-safe atomics, context patterns |
| `go.md` | Future Features | keep-as-is | Go-specific version-gated features for upgrade readiness |
| `python-standards.md` | Error Handling | keep-as-is | Python-specific exception hierarchy, `from exc` chaining, bare except rules |
| `python-standards.md` | Testing | keep-as-is | Pytest-specific fixtures, `conftest.py`, testcontainers, `parametrize` |
| `python-standards.md` | Docstrings | keep-as-is | Google style docstrings, Python-specific sections (Args, Returns, Raises) |
| `rust-standards.md` | Error Handling Patterns | keep-as-is | Rust-specific `thiserror`/`anyhow`, `?` operator, `Result` aliases |
| `rust-standards.md` | Testing Patterns | keep-as-is | Rust-specific `#[cfg(test)]`, doc tests, `proptest!`, criterion benchmarks |
| `rust-standards.md` | Unsafe Code | keep-as-is | Entirely Rust-specific (SAFETY comments, FFI, scope minimization) |
| `typescript-standards.md` | Error Handling | keep-as-is | TS-specific Result pattern, type guards, error classes with branded types |
| `typescript-standards.md` | Type System Patterns | keep-as-is | Entirely TS-specific (generics, utility types, conditional types) |
| `shell-standards.md` | Error Handling | keep-as-is | Shell-specific `set -eEuo pipefail`, ERR trap, exit codes |
| `shell-standards.md` | Security | keep-as-is | Shell-specific sed injection, `jq` for JSON, CLI secret handling |
| `shell-standards.md` | Testing | keep-as-is | BATS-specific test patterns, `shellcheck` integration |
| ALL | Compliance Assessment | keep-as-is | Grading scales are language-specific (different tool outputs, thresholds) |
| ALL | Vibe Integration | keep-as-is | Prescan patterns and JIT loading are language-specific |
| ALL | Anti-Patterns | trim-universal-keep-specific | Add cross-ref to common anti-patterns; keep language-specific examples |
| ALL | Code Quality Metrics | trim-universal-keep-specific | Add cross-ref to common coverage targets; keep language-specific tool commands |

**Legend:**
- **keep-as-is** -- Section contains primarily language-specific implementation details. No changes needed.
- **trim-universal-keep-specific** -- Section contains some universal philosophical content that overlaps with this document. Add a cross-reference note at the top of the section pointing here, but keep all language-specific examples and tool commands.
- **replace-with-ref** -- Section is entirely universal philosophy. Replace with a cross-reference. (None found -- all language sections contain significant implementation details.)

**Conservative approach:** All language-specific files retain their full content. Only two section categories get a small cross-reference header added. This ensures no loss of language-specific implementation guidance.

---

**Related:** Language-specific standards in `go.md`, `python.md`, `rust.md`, `typescript.md`, `shell.md`

### examples-troubleshooting-template.md

# Examples + Troubleshooting Template

> Reference template for adding `## Examples` and `## Troubleshooting` sections to skills. Workers MUST follow this format exactly.

## Section Placement

- `## Examples` goes BEFORE `## See Also` (or at end of file if no See Also)
- `## Troubleshooting` goes AFTER `## Examples`, BEFORE `## See Also`

## Append-vs-Create Rules (4 cases)

1. **Neither section exists:** CREATE both `## Examples` and `## Troubleshooting` before `## See Also`
2. **Only `## Examples` exists:** APPEND new examples below existing ones; CREATE `## Troubleshooting` after Examples
3. **Only `## Troubleshooting` exists:** CREATE `## Examples` before Troubleshooting; APPEND new rows to existing table
4. **Both exist:** APPEND new examples and new troubleshooting rows to existing sections (don't rewrite)

## Examples Format

```markdown
## Examples

### <Scenario Title>

**User says:** `/<skill> <args>`

**What happens:**
1. Agent does X
2. Agent does Y
3. Output written to Z

**Result:** Brief description of outcome
```

Each example MUST include:
- A realistic trigger phrase showing how a user invokes the skill
- Step-by-step behavior description (what the agent actually does)
- Expected output or result

## Troubleshooting Format

```markdown
## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| Error message or symptom | Why it happens | How to fix |
```

Each entry MUST include:
- A specific, recognizable error message or symptom
- The root cause explanation
- A concrete fix (command, config change, or workaround)

## Per-Tier Requirements

| Tier | Skills | Examples | Troubleshooting | Word Budget |
|------|--------|----------|-----------------|-------------|
| Tier 1 | council, crank, vibe | 3+ scenarios | 3+ entries | Examples: max 400 words |
| Tier 2 | research, plan, implement, pre-mortem, post-mortem, rpi | 2+ scenarios | 2+ entries (skip if exists) | Examples: max 250 words |
| Tier 3 | swarm, codex-team, evolve, release, quickstart, handoff | 2+ scenarios | 2+ entries (skip if exists) | Examples: max 250 words |
| Tier 4 | bug-hunt, complexity, doc, product, status, trace, inbox, knowledge, retro | 2 scenarios | 2-3 entries | Examples: max 250 words |
| Internal | extract, flywheel, forge, inject, provenance, ratchet, standards, using-agentops | 1-2 scenarios | 2 entries | Examples: max 200 words |

**Note:** `shared` is excluded — it's a reference collection, not a skill.

Troubleshooting: max 200 words per skill across all tiers.

Total SKILL.md word count MUST stay under 5000 words. Verify with `wc -w SKILL.md`.

## Quality Bar

- Examples must reflect actual skill behavior (not placeholder text)
- Troubleshooting entries must describe real failure modes users encounter
- Internal skills: show programmatic invocation (how other skills call them)
- User-facing skills: show natural language triggers a human would type

### go.md

# Go Standards (Tier 1)

## Target Version

Detect from `go.mod`. Use all features up to and including that version. Never use features from newer versions. Current project target: **Go 1.26**.

## Required

- `gofmt` (automatic)
- `golangci-lint run` passes
- All exported symbols documented

## Error Handling

- Always check errors: `if err != nil`
- Wrap errors with context: `fmt.Errorf("doing X: %w", err)`
- Never `_ = err` without `// nolint:errcheck` comment
- Use `errors.Is(err, target)` instead of `err == target` -- works with wrapped errors (1.13+)
- Use `errors.Join(err1, err2)` to aggregate errors from parallel operations or multi-step cleanup (1.20+)
- Use `context.WithCancelCause` / `context.Cause` to attach error reasons to cancellations (1.20+)

## Common Issues

| Pattern | Problem | Fix |
|---------|---------|-----|
| `%v` for errors | Breaks error chain | Use `%w` |
| `panic()` in library | Crashes caller | Return error |
| Naked goroutine | No error handling | errgroup or channels |
| `interface{}` | Type safety loss | Use `any` (1.18+), generics, or specific types |
| `err == target` | Misses wrapped errors | `errors.Is(err, target)` (1.13+) |
| `atomic.StoreInt32` | Type-unsafe | `atomic.Bool` / `atomic.Int64` / `atomic.Pointer[T]` (1.19+) |
| `for i := 0; i < n; i++` | Verbose | `for i := range n` (1.22+) |
| Manual loop for contains/sort | Error-prone, verbose | `slices.Contains`, `slices.SortFunc` (1.21+) |
| `sync.Once` + closure wrapper | Verbose, easy to misuse | `sync.OnceFunc` / `sync.OnceValue` (1.21+) |

## Interfaces

- Accept interfaces, return structs
- Keep interfaces small (1-3 methods)
- Define interfaces where used, not implemented

## Documentation

- All exported symbols must have godoc comments starting with the symbol name
- Package-level doc in `doc.go` for non-trivial packages
- Include runnable `Example_*` functions in `_test.go` files
- Run `go doc ./...` to verify documentation

## Concurrency

- Always pass `context.Context` as first param
- Use `sync.Mutex` for shared state; use type-safe atomics (`atomic.Bool`, `atomic.Int64`, `atomic.Pointer[T]`) for simple flags/counters (1.19+)
- Prefer channels for communication
- Use `sync.OnceFunc(fn)` instead of `sync.Once` + wrapper; `sync.OnceValue(fn)` when returning a value (1.21+)
- Use `context.AfterFunc(ctx, cleanup)` to register cleanup on cancellation (1.21+)
- Loop variables are safe to capture in goroutines since 1.22 (each iteration gets its own copy)

## Modern Standard Library

### slices package (1.21+)

Prefer `slices` over hand-written loops:

| Function | Replaces |
|----------|----------|
| `slices.Contains(items, x)` | Manual search loop |
| `slices.Index(items, x)` | Manual search loop returning index |
| `slices.IndexFunc(items, fn)` | Manual search loop with predicate |
| `slices.Sort(items)` | `sort.Slice` / `sort.Strings` |
| `slices.SortFunc(items, cmp)` | `sort.Slice` with less function |
| `slices.Max(items)` / `slices.Min(items)` | Manual loop tracking max/min |
| `slices.Reverse(items)` | Manual swap loop |
| `slices.Compact(items)` | Manual dedup of consecutive elements |
| `slices.Clip(s)` | `s[:len(s):len(s)]` to remove excess capacity |
| `slices.Clone(s)` | `append([]T(nil), s...)` |

Iterator consumption (1.23+):

| Function | Usage |
|----------|-------|
| `slices.Collect(iter)` | Build slice from iterator |
| `slices.Sorted(iter)` | Collect and sort in one step |

### maps package (1.21+; Keys/Values return iterators as of 1.23)

| Function | Replaces |
|----------|----------|
| `maps.Clone(m)` | Manual map copy loop |
| `maps.Copy(dst, src)` | Manual map merge loop |
| `maps.DeleteFunc(m, fn)` | Manual delete loop with predicate |
| `maps.Keys(m)` | Manual key collection loop (returns iterator, 1.23+) |
| `maps.Values(m)` | Manual value collection loop (returns iterator, 1.23+) |

### cmp package (1.22+)

- `cmp.Or(a, b, c)` -- returns first non-zero value. Replaces `if x == "" { x = default }` chains:
  ```go
  name := cmp.Or(os.Getenv("NAME"), config.Name, "default")
  ```

### strings / bytes improvements

| Function | Version | Replaces |
|----------|---------|----------|
| `strings.Cut(s, sep)` / `bytes.Cut(b, sep)` | 1.18+ | `Index` + slice arithmetic |
| `strings.CutPrefix(s, prefix)` / `strings.CutSuffix(s, suffix)` | 1.20+ | `HasPrefix` + `TrimPrefix` |
| `strings.Clone(s)` / `bytes.Clone(b)` | 1.20+ | Manual copy (prevents memory leaks from substring references) |

### net/http improvements (1.22+)

Enhanced `ServeMux` with method and path parameters:

```go
mux.HandleFunc("GET /api/users/{id}", func(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    // ...
})
```

May eliminate the need for third-party routers for simple APIs.

### Other stdlib

| Function | Version | Replaces |
|----------|---------|----------|
| `fmt.Appendf(buf, fmt, args...)` | 1.19+ | `[]byte(fmt.Sprintf(...))` -- avoids allocation |
| `time.Since(start)` | 1.0+ | `time.Now().Sub(start)` |
| `time.Until(deadline)` | 1.8+ | `deadline.Sub(time.Now())` |
| `errors.Join(err1, err2)` | 1.20+ | Discarding all but the first error (see Error Handling) |
| `reflect.TypeFor[T]()` | 1.22+ | `reflect.TypeOf((*T)(nil)).Elem()` |
| `min(a, b)` / `max(a, b)` | 1.21+ | `if a > b` patterns or custom helpers |
| `clear(m)` / `clear(s)` | 1.21+ | Manual map deletion loop / manual slice zeroing |

## Future Features (Go 1.24+)

This section tracks features by first-supported Go version and can be used to plan future target upgrades.

| Feature | Version | What It Replaces |
|---------|---------|------------------|
| `t.Context()` | 1.24+ | `context.WithCancel(context.Background())` in tests |
| `b.Loop()` | 1.24+ | `for i := 0; i < b.N; i++` in benchmarks |
| `omitzero` JSON tag | 1.24+ | `omitempty` (which fails for `time.Duration`, structs, slices, maps) |
| `strings.SplitSeq` / `FieldsSeq` | 1.24+ | `strings.Split` when iterating (avoids intermediate slice) |
| `wg.Go(fn)` | 1.25+ | `wg.Add(1)` + `go func() { defer wg.Done(); ... }()` |
| `new(val)` | 1.26+ | `x := val; &x` for pointer creation |
| `errors.AsType[T](err)` | 1.26+ | `var target T; errors.As(err, &target)` |

### json.md

# JSON Standards (Tier 1)

## Validation
- Valid JSON (use `jq .` to verify)
- Consistent formatting (2-space indent)
- No trailing commas

## Common Issues
| Pattern | Problem | Fix |
|---------|---------|-----|
| Trailing comma | Parse error | Remove |
| Single quotes | Invalid JSON | Double quotes only |
| Comments | Invalid JSON | Remove or use JSONC |
| Unquoted keys | Invalid JSON | Quote all keys |

## JSONL (newline-delimited)
- One JSON object per line
- No trailing newline on last line
- Each line must be valid JSON

## Schema Validation
- Use JSON Schema for validation
- Reference: `"$schema": "https://..."`
- Required fields should be explicit

## Security
- Never use `eval()` or `Function()` to parse JSON — use `JSON.parse()`
- Validate against JSON Schema before processing untrusted input
- Watch for prototype pollution in JavaScript/TypeScript JSON handling
- Sanitize keys and values when constructing JSON from user input

## Large Files
- Consider JSONL for append-only logs
- Use streaming parsers for large files
- Compress with gzip for storage

### markdown.md

# Markdown Standards (Tier 1)

## Structure
- Single H1 (`#`) at top
- Hierarchical headings (don't skip levels)
- Blank line before/after headings

## Common Issues
| Pattern | Problem | Fix |
|---------|---------|-----|
| Multiple H1s | Confusing structure | Single H1 |
| Skipped heading | H1 → H3 | H1 → H2 → H3 |
| No blank lines | Rendering issues | Blank before/after blocks |
| Hard line breaks | Formatting | Let text wrap naturally |

## Tables
```markdown
| Header | Header |
|--------|--------|
| Cell   | Cell   |
```
- Align `|` for readability
- Use `-` for header separator

## Code Blocks
- Always specify language: ` ```python `
- Use inline `` `code` `` for short refs
- 4-space indent also works (but fenced preferred)

## Links
- Use descriptive link text, not generic "click here"
- Use relative paths for local references
- Check links aren't broken

### python.md

# Python Standards (Tier 1)

## Required
- `ruff check` passes (or `flake8`)
- `ruff format` (or `black`) for formatting
- Type hints on public functions
- Docstrings on public classes/functions

## Error Handling
- Never bare `except:` - always specify exception type
- Use `raise ... from e` to preserve stack traces
- Log before raising in library code

## Common Issues
| Pattern | Problem | Fix |
|---------|---------|-----|
| `except Exception:` | Too broad | Catch specific exceptions |
| `# type: ignore` | Hiding problems | Fix the type error |
| `eval()` / `exec()` | Security risk | Use safer alternatives |
| Mutable default args | Shared state bugs | Use `None` + conditional |

## Security
- Never use `eval()`, `exec()`, or `__import__()` with untrusted input
- Use `secrets` module for tokens, not `random`
- Validate and sanitize all external input (user data, file paths, URLs)
- Use parameterized queries for SQL — never string formatting

## Testing
- pytest preferred
- `conftest.py` for shared fixtures
- Mock external services, not internal code

### rust.md

# Rust Standards (Tier 1)

## Required
- `cargo fmt` (automatic)
- `cargo clippy` passes (no warnings)
- All public items documented (rustdoc)

## Error Handling
- Use `Result<T, E>` for fallible operations
- Implement custom errors with `thiserror` or `anyhow`
- Never `unwrap()` in library code (OK in tests/bins)
- Use `?` operator for error propagation

## Ownership & Borrowing
- Prefer references over cloning
- Use `&str` in function params over `String`
- Add explicit lifetime annotations when needed
- Clone sparingly and document why

## Common Issues
| Pattern | Problem | Fix |
|---------|---------|-----|
| `unwrap()` | Panic on None/Err | Use `?` or pattern match |
| Mutable statics | Data races | Use `once_cell` or `Mutex` |
| String allocation | Performance | Use `&str` in function params |
| Lifetime errors | Borrow checker reject | Add explicit lifetimes |
| Unsafe block | Memory unsafety | Add `// SAFETY:` comment |
| Excessive `.clone()` | Performance waste | Use references or `Cow<T>` |

## Unsafe Code
- Always add `// SAFETY:` comment explaining invariants
- Minimize unsafe scope
- Prefer safe abstractions

## Security
- Minimize `unsafe` blocks — each needs `// SAFETY:` justification
- Use `secrecy::Secret<T>` for sensitive values (prevents accidental logging)
- Validate all external input before deserialization (`serde` validators)
- Prefer `ring` or `rustls` over OpenSSL bindings

## Documentation
- All public items must have rustdoc comments (`///`)
- Include `# Examples` section in doc comments for complex APIs
- Use `#![deny(missing_docs)]` in library crates
- Run `cargo doc --no-deps` to verify doc builds

## Testing
- `cargo test` (built-in)
- `cargo test --doc` (doc tests)
- Use `#[cfg(test)]` modules
- `cargo bench` for benchmarks

### shell.md

# Shell Standards (Tier 1)

## Required Header
```bash
#!/usr/bin/env bash
set -euo pipefail
```

## Validation
- `shellcheck` must pass
- Quote all variables: `"$var"` not `$var`

## Common Issues
| Pattern | Problem | Fix |
|---------|---------|-----|
| Unquoted `$var` | Word splitting | `"$var"` |
| `cd` without check | Silent failure | `cd dir \|\| exit 1` |
| `[ ]` vs `[[ ]]` | Portability | Use `[[ ]]` in bash |
| Backticks | Nesting issues | Use `$(command)` |

## Best Practices
- Use `local` for function variables
- Trap errors: `trap 'cleanup' ERR EXIT`
- Check command existence: `command -v foo >/dev/null`
- Use `readonly` for constants

## Cluster Scripts
- Always verify connectivity first:
  ```bash
  oc whoami &>/dev/null || { echo "Not logged in"; exit 1; }
  ```

### skill-structure.md

# Skill Structure Standard

**Version:** 2.0.0
**Last Updated:** 2026-02-20
**Source:** Codex official documentation (https://code.claude.com/docs/en/skills)
**Purpose:** Defines the required structure, frontmatter, and quality standards for all AgentOps skills.

---

## Table of Contents

1. [File Structure](#file-structure)
2. [YAML Frontmatter](#yaml-frontmatter)
3. [Description Field](#description-field)
4. [Body Structure](#body-structure)
5. [Progressive Disclosure](#progressive-disclosure)
6. [Quality Checklist](#quality-checklist)
7. [AgentOps Extensions](#agentops-extensions)

---

## File Structure

```
skill-name/
├── SKILL.md              # Required — exact case, no variations
├── scripts/              # Optional — executable code
├── references/           # Optional — progressive disclosure docs
└── assets/               # Optional — templates, fonts, icons
```

### Rules

| Rule | ALWAYS | NEVER |
|------|--------|-------|
| Entry point | `SKILL.md` (exact case) | `skill.md`, `SKILL.MD`, `Skill.md` |
| Folder name | kebab-case (`bug-hunt`) | spaces, underscores, capitals |
| Name match | Folder name = `name:` field | Mismatch between folder and frontmatter |
| README | None inside skill folder | `README.md` in skill directories |
| Reserved | Any valid kebab-case name | `claude-*` or `anthropic-*` prefixes |

---

## YAML Frontmatter

### Required Fields

```yaml
---
name: skill-name
description: 'What it does. When to use it. Trigger phrases.'
---
```

Only `description` is technically required (recommended). If `name` is omitted, the directory name is used.

### All Codex Frontmatter Fields

| Field | Required | Purpose |
|-------|----------|---------|
| `name` | No | Display name. Lowercase letters, numbers, hyphens only (max 64 chars). Defaults to directory name. |
| `description` | Recommended | What the skill does and when to use it. Claude uses this to decide when to load the skill. |
| `argument-hint` | No | Hint shown during autocomplete (e.g., `[issue-number]`, `[filename] [format]`). |
| `disable-model-invocation` | No | Set to `true` to prevent Claude from auto-loading. User must invoke with `/name`. Default: `false`. |
| `user-invocable` | No | Set to `false` to hide from `/` menu. Use for background knowledge. Default: `true`. |
| `allowed-tools` | No | Tools Claude can use without permission when skill is active (e.g., `Read, Grep, Glob`). |
| `model` | No | Model to use when skill is active (`sonnet`, `opus`, `haiku`, `inherit`). |
| `context` | No | Set to `fork` to run in a forked subagent context. **Only for worker spawner skills** (e.g., council, codex-team). Never set on orchestrators (evolve, rpi, crank) — they need visibility. See two-tier rule in SKILL-TIERS.md. |
| `agent` | No | Which subagent type to use when `context: fork` is set (e.g., `Explore`, `Plan`, `general-purpose`). |
| `hooks` | No | Hooks scoped to this skill's lifecycle. |

### Execution Mode (Two-Tier Rule)

Skills follow a two-tier execution model: **orchestrators stay in the main context, workers fork.**

| Mode | `context: fork` | When to use |
|------|-----------------|-------------|
| Orchestrator | Do NOT set | Skills that loop, gate phases, or report progress (evolve, rpi, crank, vibe, post-mortem, etc.) |
| Worker spawner | Set `context: fork` | Skills that fan out parallel workers and merge results (council, codex-team) |

Optionally add `execution_mode` to the `metadata` block for documentation (informational only — no tooling reads this field):

```yaml
metadata:
  tier: execution
  execution_mode: orchestrator  # informational — stays in main context
```

See `SKILL-TIERS.md` for the full classification table and the two-tier rule.

### Invocation Control Matrix

| Frontmatter | User can invoke | Claude can invoke | Context loading |
|-------------|----------------|-------------------|-----------------|
| (default) | Yes | Yes | Description always in context, full skill loads when invoked |
| `disable-model-invocation: true` | Yes | No | Description not in context, full skill loads when user invokes |
| `user-invocable: false` | No | Yes | Description always in context, full skill loads when invoked |

### String Substitutions

| Variable | Description |
|----------|-------------|
| `$ARGUMENTS` | All arguments passed when invoking the skill |
| `$ARGUMENTS[N]` | Specific argument by 0-based index |
| `$N` | Shorthand for `$ARGUMENTS[N]` |
| `${CLAUDE_SESSION_ID}` | Current session ID |

### Dynamic Context Injection

The `` !`command` `` syntax runs shell commands before skill content is sent to Claude:

```yaml
## Context
- Current branch: !`git branch --show-current`
- Recent changes: !`git log --oneline -5`
```

### AgentOps Extension Fields (under `metadata:`)

AgentOps uses these custom fields under `metadata:` for tooling integration:

```yaml
metadata:
  tier: solo          # solo, team, orchestration, library, background, meta
  dependencies:       # List of skill names this skill depends on
    - standards
    - council
  internal: true      # true for non-user-facing skills
  replaces: old-name  # Deprecated skill this replaces
```

**Tier values and their constraints:**

| Tier | Max Lines | Purpose |
|------|-----------|---------|
| `solo` | 200 | Single-agent, no spawning |
| `team` | 500 | Spawns workers |
| `orchestration` | 500 | Coordinates multiple skills/teams |
| `library` | 200 | Referenced by other skills, not invoked directly |
| `background` | 200 | Hooks/automation, not user-invoked |
| `meta` | 200 | Explains the system itself |

### Security Restrictions

- No XML angle brackets (`<` `>`) in frontmatter
- No `claude` or `anthropic` in skill names
- YAML safe parsing only (no code execution)

---

## Description Field

The description is the **most critical field** — it determines when Claude loads the skill.

### Structure

```
[What it does] + [When to use it] + [Key capabilities]
```

### Requirements

- Under 1024 characters
- MUST include trigger phrases users would actually say
- MUST explain what the skill does (not just when)
- No XML tags

### Good Examples

```yaml
# Specific + actionable + triggers
description: 'Investigate suspected bugs with git archaeology and root cause analysis. Triggers: "bug", "broken", "doesn''t work", "failing", "investigate bug".'

# Clear value prop + multiple triggers
description: 'Comprehensive code validation. Runs complexity analysis then multi-model council. Answer: Is this code ready to ship? Triggers: "vibe", "validate code", "check code", "review code", "is this ready".'
```

### Bad Examples

```yaml
# Too vague
description: Helps with projects.

# Missing triggers
description: Creates sophisticated multi-page documentation systems.

# Too technical, no user triggers
description: Implements the Project entity model with hierarchical relationships.
```

### Internal Skills Exception

Library/background/meta skills that are auto-loaded (not user-invoked) may describe their loading mechanism instead of user triggers:

```yaml
description: 'Auto-loaded by $vibe, $implement based on file types.'
```

---

## Body Structure

### Recommended Template

```markdown
---
name: skill-name
description: '...'
metadata:
  tier: solo
---

# Skill Name

## Quick Start

Example invocations showing common usage patterns.

## Instructions

### Step 1: [First Major Step]
Specific, actionable instructions with exact commands.

### Step 2: [Next Step]
...

## Examples

### Example 1: [Common scenario]
User says: "..."
Actions: ...
Result: ...

## Troubleshooting

### Error: [Common error]
Cause: ...
Solution: ...
```

### Requirements

| Aspect | Requirement |
|--------|-------------|
| Size | Under 5,000 words; keep SKILL.md under 500 lines |
| Instructions | Specific and actionable (exact commands, not "validate the data") |
| Examples | At least 2-3 usage examples for user-facing skills |
| Error handling | Troubleshooting section for common failures |
| References | Link to `references/` for detailed docs (don't inline everything) |

---

## Progressive Disclosure

Skills use three levels:

1. **Frontmatter** — Always in system prompt. Minimal: name + description.
2. **SKILL.md body** — Loaded when skill is relevant. Core instructions.
3. **references/** — Loaded on-demand. Detailed docs, schemas, examples.

### Rules

- Keep SKILL.md focused on core workflow
- Move detailed reference material to `references/`
- Explicitly link to references: "Read `references/api-patterns.md` for..."
- Move scripts >20 lines to `scripts/` directory
- Move inline bash >30 lines to `scripts/` or `references/`

---

## Quality Checklist

### Before Commit

- [ ] `SKILL.md` exists (exact case)
- [ ] Folder name matches `name:` field
- [ ] Folder name is kebab-case
- [ ] Description includes WHAT + WHEN (triggers)
- [ ] Description under 1024 characters
- [ ] No XML tags in frontmatter
- [ ] No `claude`/`anthropic` in name
- [ ] `metadata.tier` is set and valid
- [ ] SKILL.md under 5,000 words
- [ ] User-facing skills have examples section
- [ ] User-facing skills have troubleshooting section
- [ ] Detailed docs in references/, not inlined
- [ ] No README.md in skill folder

### Trigger Testing

- [ ] Triggers on 3+ obvious phrases
- [ ] Triggers on paraphrased requests
- [ ] Does NOT trigger on unrelated topics

---

## AgentOps Extensions

These are AgentOps-specific patterns not in the Codex spec:

### Tier System

Controls line limits and categorization. Enforced by `tests/skills/lint-skills.sh`.

### Dependencies

Declared under `metadata.dependencies`. Validated by `tests/skills/validate-skill.sh`.

### Skill Tiers Document

Full taxonomy at `skills/SKILL-TIERS.md`.

### Standards Loading

Language standards loaded JIT by `$vibe`, `$implement` — see `standards-index.md`.

### typescript.md

# TypeScript Standards (Tier 1)

## Required
- `strict: true` in tsconfig.json
- `prettier` for formatting
- `eslint` with recommended rules

## Type Safety
- No `any` - use `unknown` + type guards
- No `@ts-ignore` without explanation
- Prefer `interface` for objects, `type` for unions

## Common Issues
| Pattern | Problem | Fix |
|---------|---------|-----|
| `as Type` | Unsafe cast | Type guards or `satisfies` |
| `!` (non-null) | Runtime errors | Proper null checks |
| `== null` | Loose equality | `=== null \|\| === undefined` |
| Implicit `any` | Type safety loss | Enable `noImplicitAny` |

## React (if applicable)
- Functional components only
- `useState` / `useReducer` for state
- `useEffect` with proper deps array
- No inline object/function props (memo issues)

## Testing
- Jest or Vitest
- React Testing Library for components
- MSW for API mocking

### yaml.md

# YAML Standards (Tier 1)

## Validation
- `yamllint` must pass
- 2-space indentation
- No trailing whitespace

## Common Issues
| Pattern | Problem | Fix |
|---------|---------|-----|
| Tabs | Invalid YAML | 2 spaces |
| `yes`/`no` unquoted | Becomes boolean | Quote: `"yes"` |
| `:` in value | Parse error | Quote the value |
| Long lines | Readability | Use `>` or `\|` |

## Kubernetes/Helm
- Use `---` between documents
- Labels: `app.kubernetes.io/*`
- Always specify `resources.limits`
- Use ConfigMaps for config, Secrets for secrets

## Security
- Never use `yaml.load()` (Python) — always `yaml.safe_load()`
- Quote values that look like booleans (`"yes"`, `"no"`, `"true"`)
- Validate against schema before processing untrusted YAML
- Avoid anchors/aliases (`*`/`&`) in user-facing configs — confusing and exploitable

## Multiline Strings
```yaml
# Literal (preserves newlines)
description: |
  Line 1
  Line 2

# Folded (joins lines)
description: >
  This becomes
  one line
```


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
check "name is standards" "grep -q '^name: standards' '$SKILL_DIR/SKILL.md'"
check "mentions language-specific or coding standards" "grep -qiE 'language-specific|coding standards' '$SKILL_DIR/SKILL.md'"
check "has references directory" "[ -d '$SKILL_DIR/references' ]"

echo ""; echo "Results: $PASS passed, $FAIL failed"
[ $FAIL -eq 0 ] && exit 0 || exit 1
```


