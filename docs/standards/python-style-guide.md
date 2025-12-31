# Python Style Guide

> **Purpose**: Unified Python coding standards for this repository.

## Quick Reference

| Aspect | Standard |
|--------|----------|
| **Version** | Python 3.11+ |
| **Formatter** | black (line-length=100) |
| **Linter** | ruff |
| **Complexity** | CC ≤ 10 per function |

## pyproject.toml Configuration

```toml
[tool.black]
line-length = 100
target-version = ['py311']

[tool.ruff]
line-length = 100
target-version = "py311"

[tool.ruff.lint]
select = ["E", "W", "F", "I", "N", "UP", "B", "C4", "SIM"]
ignore = ["E501"]
```

## Code Complexity

**Required:** Maximum cyclomatic complexity of 10 (Grade B) per function.

| Grade | CC Range | Action |
|-------|----------|--------|
| A | 1-5 | ✅ Ideal |
| B | 6-10 | ✅ Acceptable |
| C | 11-20 | ⚠️ Refactor when touching |
| D+ | 21+ | 🔴 Must refactor |

```bash
# Check complexity
radon cc scripts/ -s -a

# Fail on Grade C or worse
xenon scripts/ --max-absolute B
```

### Reducing Complexity

**Dispatch Pattern** - Replace if/elif chains:
```python
# ❌ Bad - if/elif chain (CC=18+)
def main():
    if args.mode == "read":
        # 50 lines
    elif args.mode == "write":
        # 50 lines

# ✅ Good - Dispatch pattern (CC=3)
def _handle_read(args): ...
def _handle_write(args): ...

HANDLERS = {"read": _handle_read, "write": _handle_write}

def main():
    handler = HANDLERS.get(args.mode)
    return handler(args) if handler else die("Unknown mode")
```

## Type Hints

**Required** for all public functions. Use modern syntax:

```python
from __future__ import annotations

def process(items: list[str], limit: int | None = None) -> dict[str, int]:
    ...
```

## Error Handling

```python
# ✅ Good - Specific exception, logged
try:
    result = process_data(payload)
except ValueError as exc:
    logging.warning(f"Invalid data: {exc}")
    return None

# ❌ Bad - Bare exception, swallowed
try:
    process_data()
except Exception:
    pass
```

## CLI Script Template

```python
#!/usr/bin/env python3
"""One-line description.

Usage:
    python3 script.py --config config.yaml --apply
"""
from __future__ import annotations
import argparse
import logging
import sys

logging.basicConfig(format="%(asctime)s %(levelname)s %(message)s", level=logging.INFO)

def die(msg: str) -> None:
    logging.error(msg)
    sys.exit(1)

def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--apply", action="store_true")
    return parser.parse_args()

def main() -> int:
    args = parse_args()
    # Main logic
    return 0

if __name__ == "__main__":
    sys.exit(main())
```

## Summary

1. Python 3.11+ required
2. Use black + ruff
3. **Cyclomatic complexity ≤ 10** per function
4. Type hints for public functions
5. Specific exception types, never bare `except Exception:`
6. Use logging, never print()
