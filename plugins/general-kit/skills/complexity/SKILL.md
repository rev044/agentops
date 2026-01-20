---
name: complexity
description: Analyze code complexity and find refactor targets using radon/gocyclo
version: 1.1.0
context: fork
triggers:
  - "analyze complexity"
  - "find complex functions"
  - "check complexity"
  - "refactor targets"
  - "code quality"
  - "radon"
  - "gocyclo"
  - "cyclomatic complexity"
allowed-tools: Bash, Read, Grep, Glob
agent: code-quality-expert
skills:
  - standards
---

# Complexity Analysis Skill

Analyze code complexity using `radon` (Python) and `gocyclo` (Go) to identify functions that need refactoring.

**Supported Languages:** Python, Go

## Overview

This skill provides comprehensive code complexity analysis:
- Run `radon cc` on target paths
- Interpret complexity grades (A-F)
- Identify refactoring candidates
- Generate actionable recommendations
- Support enforcement checks via `xenon`

---

## Phase 1: Determine Scope

### If Path Provided

```bash
# Validate the path exists
ls -d "$PATH"
```

### If No Path (Default)

Analyze all Python services:
```
services/
```

---

## Phase 2: Run Complexity Analysis

### Python (radon)

```bash
# Show complexity for all Python files with summary
radon cc <path> -s -a

# For single file
radon cc <path/to/file.py> -s
```

**Output Format** - `radon cc` outputs lines like:
```
F 179:0 sync_gitlab - F (84)
C 35:0 validate_input - C (12)
A 117:4 embed_document - A (3)
```
Format: `<Grade> <line>:<col> <function_name> - <Grade> (<complexity>)`

### Go (gocyclo)

```bash
# Show all functions with CC > 10
gocyclo -over 10 <path>

# Top 10 most complex functions
gocyclo -top 10 <path>

# All functions sorted by complexity
gocyclo <path> | sort -rn
```

**Output Format** - `gocyclo` outputs lines like:
```
58 cmd runSling /path/to/sling.go:128:1
28 cmd formatLogLine /path/to/logger.go:110:1
```
Format: `<complexity> <package> <function> <file>:<line>:<col>`

**Install**:
```bash
go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
```

---

## Phase 3: Interpret Results

### Complexity Grades Reference

| Grade | CC Range | Status | Action |
|-------|----------|--------|--------|
| **A** | 1-5 | Ideal | No action needed |
| **B** | 6-10 | Acceptable | No action needed |
| **C** | 11-20 | Complex | Refactor when touching this code |
| **D** | 21-30 | Very Complex | Should refactor soon |
| **E** | 31-40 | Extremely Complex | Urgent refactor needed |
| **F** | 41+ | Unmaintainable | Critical - blocks merges |

### Summary Output

Provide a formatted report:

```markdown
## Complexity Analysis Report

**Path:** <analyzed path>
**Date:** <current date>

### Summary
- Total functions analyzed: X
- Grade distribution: A(X) B(X) C(X) D(X) E(X) F(X)
- Average complexity: X.X

### Functions Needing Refactoring (Grade C or worse)

| File | Function | Grade | CC | Priority |
|------|----------|-------|----|---------
| path/file.py | function_name | C | 12 | Medium |
| path/file.py | function_name | D | 25 | High |
| path/file.py | function_name | F | 45 | Critical |

### Top 5 Most Complex Functions

1. `path/file.py:function_name` - F(84)
2. `path/file.py:function_name` - E(35)
...
```

---

## Phase 4: Recommendations

For each function Grade C or worse, provide:

### Refactoring Recommendations

**Function:** `<file>:<function_name>` (Grade <grade>, CC=<value>)

**Likely Issues:**
- Nested conditionals
- Long if/elif chains
- Multiple loop levels
- Mixed concerns

**Recommended Patterns:** (from `domain-kit/skills/standards/references/python.md`)

| Pattern | When to Use |
|---------|-------------|
| **Phase-Based Orchestration** | Setup + processing + cleanup phases mixed together |
| **Per-Resource Processor Extraction** | Loop with complex per-item logic |
| **`from_raw()` Factory Methods** | Dataclass parsing dicts/YAML/JSON |
| **Async Iterator Abstraction** | Pagination + per-item processing |
| **Shared Helper Extraction** | Duplicated logic across functions |
| **Keyword-Only Arguments (`*`)** | 3+ parameters, order confusion |

**Example Fix:**
```python
# Before: CC=25 (nested loops + conditionals)
def process_items(items):
    for item in items:
        if item.type == "a":
            # 20 lines
        elif item.type == "b":
            # 20 lines

# After: CC=3 (dispatch to focused handlers)
def process_items(items):
    handlers = {"a": _process_a, "b": _process_b}
    for item in items:
        handlers[item.type](item)
```

---

## Phase 5: Enforcement Check (Optional)

To verify code meets quality gates:

```bash
# Fail if any function exceeds Grade B (CC > 10)
xenon <path> --max-absolute B

# This is what CI runs - will exit non-zero if thresholds exceeded
```

---

## Quick Reference

### Python Commands (radon)

```bash
# Analyze single file
radon cc services/etl/app/routers/sync.py -s

# Analyze entire service
radon cc services/etl/app/ -s -a

# Analyze all services
radon cc services/ -s -a

# JSON output for tooling
radon cc services/ -j > complexity.json

# Enforcement (CI style)
xenon services/etl/app/ --max-absolute B
```

### Go Commands (gocyclo)

```bash
# Analyze single file
gocyclo cmd/main.go

# Find functions with CC > 15
gocyclo -over 15 ./...

# Top 10 most complex functions
gocyclo -top 10 ./...

# Average complexity
gocyclo -avg ./...
```

### Interpreting Average Complexity

| Average | Status |
|---------|--------|
| < 5 | Excellent - well-factored code |
| 5-10 | Good - maintainable |
| 10-15 | Moderate - some refactoring needed |
| > 15 | Poor - significant refactoring needed |

---

## Creating Follow-Up Tasks (Optional)

For Grade D or worse functions, you can track refactoring work:

**With beads installed:**
```bash
bd create "Refactor <function_name> to reduce complexity" --type task --priority P2
```

**Without beads:** Create a GitHub issue or add to your project's issue tracker.

---

## Related Documentation

- **Python Standards:** `domain-kit/skills/standards/references/python.md`
- **Go Standards:** `domain-kit/skills/standards/references/go.md`
- **CI Integration:** Complexity checks run in GitLab CI via `xenon`

---

## Example Session

```
User: /complexity services/etl/app/routers/

Claude: Running complexity analysis on services/etl/app/routers/...

$ radon cc services/etl/app/routers/ -s -a

## Complexity Analysis Report

**Path:** services/etl/app/routers/
**Date:** 2025-12-29

### Summary
- Total functions analyzed: 45
- Grade distribution: A(32) B(8) C(3) D(2) E(0) F(0)
- Average complexity: 5.2

### Functions Needing Refactoring (Grade C or worse)

| File | Function | Grade | CC | Priority |
|------|----------|-------|----|---------
| sync.py | sync_slack | C | 11 | Medium |
| search.py | advanced_search | C | 14 | Medium |
| upload.py | process_batch | C | 13 | Medium |
| sync.py | sync_jira | D | 22 | High |
| sync.py | sync_confluence | D | 21 | High |

### Recommendations

**sync.py:sync_jira** (Grade D, CC=22)
- Apply **Per-Resource Processor Extraction** pattern
- Extract per-issue processing to dedicated module
- See: services/etl/app/routers/sync/jira_issue_processor.py for pattern

**sync.py:sync_confluence** (Grade D, CC=21)
- Apply **Per-Resource Processor Extraction** pattern
- Extract per-page processing to dedicated module

To track this work (optional, with beads):
$ bd create "Refactor sync_jira to reduce complexity" --type task --priority P2
$ bd create "Refactor sync_confluence to reduce complexity" --type task --priority P2

Or create GitHub issues / add to your preferred issue tracker.
```
