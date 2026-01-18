# Vibe Pattern Reference

## Pre-Scan Patterns (Static Detection)

Fast static analysis - no LLM required.

**Supported Languages:** Python, Go, Bash

| ID | Pattern | Severity | Detection Method | Languages |
|----|---------|----------|------------------|-----------|
| P1 | Phantom Modifications | CRITICAL | git show vs file content | All |
| P4 | Invisible Undone | HIGH | TODO/FIXME, commented code | All |
| P5 | Eldritch Horror | HIGH | CC > 15, functions > 50 lines | py, go |
| P8 | Cargo Cult Error Handling | HIGH | except: pass, shellcheck | py, sh |
| P9 | Documentation Phantom | MEDIUM | docstrings vs implementation | py |
| P12 | Zombie Code | MEDIUM | unused functions, unreachable | py |

### P1: Phantom Modifications (CRITICAL)

**What**: Committed lines that don't exist in current file.

**Why Critical**: Indicates broken git workflow - changes were committed but then removed or lost.

**Detection**: Compare `git show HEAD -- <file>` with actual file content.

**Fix**: Re-commit or investigate git history for what happened.

---

### P4: Invisible Undone (HIGH)

**What**: TODO markers, FIXME, commented-out code, revert comments.

**Why High**: Incomplete work or tech debt markers that may be forgotten.

**Detection**:
- Grep for `TODO|FIXME|XXX|HACK`
- Grep for `^\s*#\s*(def |class |if |for )` (commented code)

**Fix**: Complete the work or remove the markers with explanation.

---

### P5: Eldritch Horror (HIGH)

**What**: Functions with cyclomatic complexity > 15 or > 50 lines.

**Why High**: Too complex to maintain safely - high bug probability.

**Detection by Language**:

| Language | Tool | Command |
|----------|------|---------|
| Python | radon | `radon cc <file> -s -n E` |
| Go | gocyclo | `gocyclo -over 15 <file>` |

**Thresholds** (same for all languages):
- CC > 15: Flag as complex
- Lines > 50: Flag as long

**Install Tools**:
```bash
pip install radon                    # Python
go install github.com/fzipp/gocyclo/cmd/gocyclo@latest  # Go
```

**Fix**: Extract functions, simplify logic, break into smaller units.

---

### P8: Cargo Cult Error Handling (HIGH)

**What**: Empty except blocks, pass-only handlers, bare except, shell anti-patterns.

**Why High**: Swallowed errors hide bugs and make debugging impossible.

**Detection by Language**:

| Language | Tool | What |
|----------|------|------|
| Python | AST | `ast.Try` with empty/pass handlers |
| Bash | shellcheck | Warning-level issues (quoting, exit codes) |

**Python Patterns**:
```python
# BAD
try:
    risky()
except:          # Bare except catches SystemExit!
    pass         # Silently swallows everything

# GOOD
try:
    risky()
except SpecificError as e:
    logger.error(f"Failed: {e}")
    raise
```

**Bash Patterns** (shellcheck detects):
```bash
# BAD: SC2181 - Check exit code directly
cmd
if [ $? -ne 0 ]; then ...

# GOOD
if ! cmd; then ...

# BAD: SC2086 - Unquoted variable
rm $file

# GOOD
rm "$file"
```

**Install**: `brew install shellcheck` or `apt install shellcheck`

**Fix**: Handle or propagate errors explicitly.

---

### P9: Documentation Phantom (MEDIUM)

**What**: Docstrings claiming behavior not implemented.

**Why Medium**: False security from lying documentation.

**Detection**: Match docstring claims against implementation:
- "validates" but no raise/ValueError
- "ensures" but no assert
- "encrypts" but no crypto imports
- "authenticates" but no token handling
- "sanitizes" but no escape/strip

**Fix**: Update docs to match reality or implement claimed behavior.

---

### P12: Zombie Code (MEDIUM)

**What**: Unused functions, unreachable code after return/raise.

**Why Medium**: Dead code clutters codebase and may have vulnerabilities.

**Detection**:
- Track defined vs called functions
- Find statements after return/raise

**Exclusions**: `main`, `setup`, `teardown`, `test_*`, `_private`

**Fix**: Remove unused code entirely.

---

## Semantic Patterns (LLM-Powered)

Deep analysis requiring semantic understanding.

| Analysis | Code Prefix | Checks |
|----------|-------------|--------|
| Docstrings | FAITH-xxx | Parameter mismatches, return lies, behavioral claims |
| Names | NAME-xxx | validate_*, sanitize_*, get_*, auth_*, encrypt_* |
| Security | SEC-xxx | Validation theater, auth bypass, crypto theater |
| Pragmatic | PRAG-xxx | DRY, orthogonality, reversibility |
| Slop | SLOP-xxx | Verbose boilerplate, hallucinations, cargo cult |

### Severity Mapping

**Docstrings (FAITH-xxx)**:
- THEATER (CRITICAL): Appears to work but doesn't
- MISMATCH (HIGH): Significant claim/reality gap
- PARTIAL (MEDIUM): Minor discrepancies

**Names (NAME-xxx)**:
- Security lie (HIGH): auth/encrypt/validate that doesn't
- Behavioral mismatch (MEDIUM): get with side effects
- Generic name (LOW): process/handle/helper

**Security (SEC-xxx)**:
- CRITICAL: Direct exploit (injection, auth bypass)
- HIGH: Security gap (weak crypto, swallowed exceptions)
- MEDIUM: Defense weakness (TOCTOU, debug mode)

**Pragmatic (PRAG-xxx)**:
- DRY >20 lines (HIGH): Large duplication
- Orthogonality 4+ (HIGH): God function
- Disabled tests (HIGH): Coverage gap
- Hardcoded config (MEDIUM): Deployment friction

**Slop (SLOP-xxx)**:
- Hallucination (CRITICAL): Non-existent imports
- Cargo cult (HIGH): async without await
- Verbose (MEDIUM): Docstring > function
- Sycophantic (LOW): AI conversation artifacts
