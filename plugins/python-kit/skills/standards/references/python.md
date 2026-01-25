# Python Style Guide - Tier 1 Quick Reference

<!-- Tier 1: Generic standards (~5KB), always loaded -->
<!-- Tier 2: Deep standards in vibe/references/python-standards.md (~20KB), loaded on --deep -->
<!-- Last synced: 2026-01-21 -->

> **Purpose:** Quick reference for Python coding standards. For comprehensive patterns, load Tier 2.

---

## Quick Reference

| Standard | Value | Validation |
|----------|-------|------------|
| **Python Version** | 3.12+ | `python --version` |
| **Formatter** | ruff format | `ruff format --check .` |
| **Linter** | ruff check | `ruff check .` |
| **Complexity** | CC <= 10 | `radon cc -s -n C .` |
| **Type Checker** | ruff / mypy | Modern syntax: `list[str]` |

### Package Management

| Tool | Use For | Command |
|------|---------|---------|
| **uv** | Project deps | `uv add requests` |
| **pipx** | Global CLI tools | `pipx install ruff` |
| **brew** | System tools | `brew install shellcheck` |
| ~~pip~~ | Avoid | Use uv instead |

---

## ruff Configuration (Minimum)

```toml
# pyproject.toml
[tool.ruff]
line-length = 100
target-version = "py312"

[tool.ruff.lint]
select = ["E", "W", "F", "I", "N", "UP", "B", "C4", "SIM"]
ignore = ["E501"]
```

---

## Common Errors

| Symptom | Cause | Fix |
|---------|-------|-----|
| `ruff check` fails | Style violations | Run `ruff check --fix` |
| `xenon` reports Grade C+ | CC > 10 | Apply dispatch pattern |
| `TypeError: NoneType` | Missing null check | Add guard clause |
| Import cycle error | Circular imports | Move imports inside function |
| `ModuleNotFoundError` | Missing dependency | Run `uv sync` |
| Type hint error | Old syntax | Use `list[str]` not `List[str]` |
| f-string in logging | Using f"..." | Use `%s`: `logging.info("x: %s", x)` |

---

## Anti-Patterns

| Name | Pattern | Instead |
|------|---------|---------|
| God Function | >100 lines in one function | Dispatch pattern, extract helpers |
| Bare Except | `except Exception: pass` | Specific exceptions, log warnings |
| Print Debugging | `print("debug:", x)` | Use `logging.debug()` |
| Deep Nesting | 4+ indent levels | Early returns, extract functions |
| Global State | Module-level mutable state | Pass state explicitly |
| Any Typed | `def foo(x: Any)` | Use generics or specific types |
| Magic Strings | `if mode == "patch"` | Define constants |
| Stringly Typed | Strings for enums | Use `enum.Enum` or Literal |

---

## Naming Conventions

| Element | Convention | Example |
|---------|------------|---------|
| **Modules** | `snake_case.py` | `my_script.py` |
| **Classes** | `PascalCase` | `MyClient` |
| **Functions** | `snake_case()` | `get_secret()` |
| **Variables** | `snake_case` | `mount_point` |
| **Constants** | `UPPER_SNAKE_CASE` | `MAX_RETRIES` |
| **Private** | `_leading_underscore` | `_internal_helper()` |

---

## Type Hints

**Required:** Type hints for all public functions

```python
from __future__ import annotations

# Modern syntax (3.12+)
def process(
    items: list[str],
    config: dict[str, Any] | None = None,
) -> list[str]:
    """Process items with optional config."""
    ...
```

| Old Syntax | Modern Syntax |
|------------|---------------|
| `List[str]` | `list[str]` |
| `Dict[str, int]` | `dict[str, int]` |
| `Optional[str]` | `str \| None` |
| `Union[str, int]` | `str \| int` |

---

## Complexity Grades

| Grade | CC Range | Action |
|-------|----------|--------|
| A | 1-5 | Ideal |
| B | 6-10 | Acceptable |
| C | 11-20 | Refactor when touching |
| D | 21-30 | Must refactor |
| F | 31+ | Block merge |

---

## Summary Checklist

| Category | Requirement |
|----------|-------------|
| **Version** | Python 3.12+ |
| **Formatting** | All code passes `ruff format --check` |
| **Linting** | All code passes `ruff check` |
| **Complexity** | CC <= 10 per function |
| **Types** | All public functions have type hints |
| **Docstrings** | Google-style for public APIs |
| **Errors** | Specific exceptions, never bare except |
| **Logging** | Use `logging`, never `print()` |
| **Tests** | pytest with coverage |

---

## Talos Prescan Checks

| Check | Pattern | Rationale |
|-------|---------|-----------|
| PRE-002 | TODO/FIXME markers | Track technical debt |
| PRE-003 | Complexity > 10 | High CC correlates with bugs |
| PRE-004 | Bare except blocks | Silent failures |
| PRE-008 | print() in non-CLI | Use structured logging |
| PRE-015 | f-string in logging | Use % formatting |

---

## JIT Loading

**Tier 2 (Deep Standards):** For comprehensive patterns including:
- Project structure and package management details
- Reducing complexity (dispatch, guards, lookup tables)
- Error handling patterns with examples
- Testing with testcontainers
- CLI script template
- Validation & evidence requirements

Load: `vibe/references/python-standards.md`
