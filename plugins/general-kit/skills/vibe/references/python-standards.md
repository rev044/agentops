# Python Standards Catalog - Vibe Canonical Reference

**Version:** 1.0.0
**Last Updated:** 2026-01-21
**Purpose:** Canonical Python standards for vibe skill validation

---

## Table of Contents

1. [Project Structure](#project-structure)
2. [Package Management](#package-management)
3. [Code Formatting](#code-formatting)
4. [Reducing Complexity](#reducing-complexity)
5. [Type Hints](#type-hints)
6. [Docstrings](#docstrings)
7. [Error Handling](#error-handling)
8. [Logging](#logging)
9. [Testing](#testing)
10. [CLI Script Template](#cli-script-template)
11. [Code Quality Metrics](#code-quality-metrics)
12. [Anti-Patterns Avoided](#anti-patterns-avoided)
13. [Compliance Assessment](#compliance-assessment)

---

## Project Structure

### Standard Layout

```text
project/
├── pyproject.toml           # Project metadata and dependencies
├── uv.lock                  # Lock file (commit this!)
├── src/
│   └── mypackage/           # Source code
│       ├── __init__.py
│       ├── core.py
│       └── utils.py
├── scripts/                 # CLI tools
│   └── my_script.py
├── tests/                   # Test suite
│   ├── __init__.py
│   ├── conftest.py          # Pytest fixtures
│   ├── test_core.py
│   └── e2e/                 # End-to-end tests
│       ├── conftest.py      # Testcontainers fixtures
│       └── test_integration.py
└── docs/                    # Documentation
```

**Key Principles:**
- Use `src/` layout for packages (prevents import issues)
- CLI scripts are standalone files in `scripts/`
- Tests mirror source structure
- Always commit `uv.lock` for reproducibility

---

## Package Management

### uv - Project Dependencies

Use `uv` for all project-level Python dependencies. It's 10-100x faster than pip and creates deterministic builds.

```bash
# Initialize a new project
uv init my-project
cd my-project

# Add dependencies
uv add requests pyyaml        # Runtime deps
uv add --dev pytest ruff      # Dev deps

# Install from existing pyproject.toml
uv sync                       # Creates/updates uv.lock

# Run a script with project deps
uv run python my_script.py
```

### pipx - Global CLI Tools

Use `pipx` for Python CLI tools you want available everywhere.

```bash
# Install CLI tools globally
pipx install ruff             # Linter/formatter
pipx install radon            # Complexity analysis
pipx install xenon            # Complexity enforcement
pipx install pre-commit       # Git hooks

# Upgrade all
pipx upgrade-all

# Run without installing
pipx run cowsay "hello"
```

### When to Use What

| Need | Tool | Command |
|------|------|---------|
| Install project deps | uv | `uv sync` |
| Add library to project | uv | `uv add requests` |
| Install CLI globally | pipx | `pipx install ruff` |
| Install system tool | brew | `brew install shellcheck` |
| Quick script run | uv | `uv run script.py` |

---

## Code Formatting

### ruff Configuration

**Full recommended configuration:**

```toml
# pyproject.toml
[tool.ruff]
line-length = 100
target-version = "py312"
exclude = [
    ".git",
    ".venv",
    "__pycache__",
    "build",
    "dist",
]

[tool.ruff.lint]
select = [
    "E",   # pycodestyle errors
    "W",   # pycodestyle warnings
    "F",   # pyflakes
    "I",   # isort
    "N",   # pep8-naming
    "UP",  # pyupgrade
    "B",   # flake8-bugbear
    "C4",  # flake8-comprehensions
    "SIM", # flake8-simplify
    "S",   # flake8-bandit (security)
    "A",   # flake8-builtins
    "PT",  # flake8-pytest-style
]
ignore = [
    "E501",  # line-too-long (handled by formatter)
    "S101",  # assert (OK in tests)
]

[tool.ruff.lint.per-file-ignores]
"tests/**/*.py" = ["S101"]  # Allow assert in tests

[tool.ruff.format]
quote-style = "double"
indent-style = "space"
```

### Usage

```bash
# Check linting
ruff check src/

# Auto-fix issues
ruff check --fix src/

# Format code
ruff format src/

# Check formatting only
ruff format --check src/
```

---

## Reducing Complexity

**Target:** Maximum cyclomatic complexity of 10 (Grade B) per function

### Pattern 1: Dispatch Pattern (Handler Registry)

**When to use:** Functions with if/elif chains that dispatch based on mode or type.

```python
# Bad - if/elif chain (CC=18+)
def main():
    if args.patch:
        # 90 lines of patch logic
    elif args.read:
        # 20 lines of read logic
    else:
        # 100 lines of write logic

# Good - Dispatch pattern (CC=6)
def _handle_patch_mode(args: Args, client: Client) -> None:
    """Handle --patch mode."""
    # Focused patch logic

def _handle_read_mode(args: Args, client: Client) -> None:
    """Handle --read mode."""
    # Focused read logic

def main() -> int:
    args = parse_args()
    client = build_client()

    handlers = {
        "patch": _handle_patch_mode,
        "read": _handle_read_mode,
        "write": _handle_write_mode,
    }

    handler = handlers.get(args.mode, _handle_write_mode)
    handler(args, client)
    return 0
```

### Pattern 2: Early Returns (Guard Clauses)

```python
# Bad - Deep nesting (CC=8)
def validate_document(doc: Document) -> bool:
    if doc:
        if doc.content:
            if len(doc.content) > 0:
                if doc.tenant:
                    return True
    return False

# Good - Guard clauses (CC=4)
def validate_document(doc: Document | None) -> bool:
    if not doc:
        return False
    if not doc.content:
        return False
    if len(doc.content) == 0:
        return False
    if not doc.tenant:
        return False
    return True
```

### Pattern 3: Lookup Tables

```python
# Bad - Each 'or' adds +1 CC
def normalize_field(key: str, value: str) -> str:
    if key == "tls.crt" or key == "tls.key" or key == "ca":
        return normalize_cert_field(value)
    elif key == "config.json":
        return normalize_pull_secret_json(value)
    else:
        return value

# Good - O(1) lookup
NORMALIZERS: dict[str, Callable[[str], str]] = {
    "tls.crt": normalize_cert_field,
    "tls.key": normalize_cert_field,
    "ca": normalize_cert_field,
    "config.json": normalize_pull_secret_json,
}

def normalize_field(key: str, value: str) -> str:
    normalizer = NORMALIZERS.get(key)
    return normalizer(value) if normalizer else value
```

### Measuring Complexity

```bash
# Check specific file
radon cc scripts/my_script.py -s -a

# Fail if any function exceeds Grade B (CC > 10)
xenon scripts/ --max-absolute B

# Show only Grade C or worse
radon cc scripts/ -s -n C
```

---

## Type Hints

### Modern Syntax (Python 3.12+)

```python
from __future__ import annotations
from typing import Any, Callable, TypeVar

# Basic types - use lowercase
items: list[str] = []
mapping: dict[str, int] = {}
coords: tuple[int, int, int] = (0, 0, 0)

# Union with pipe operator
value: str | int = "hello"
optional: str | None = None

# Function signatures
def process(
    items: list[str],
    config: dict[str, Any] | None = None,
    callback: Callable[[str], bool] | None = None,
) -> list[str]:
    """Process items with optional config."""
    ...

# Generics
T = TypeVar("T")

def first(items: list[T]) -> T | None:
    return items[0] if items else None
```

---

## Docstrings

### Google Style (Required)

```python
def verify_secret_after_write(
    client: hvac.Client,
    mount_point: str,
    name: str,
    expected_payload: dict[str, Any],
) -> bool:
    """Verify secret was written correctly.

    Args:
        client: Vault client connection.
        mount_point: KV v2 mount point path.
        name: Secret name/key.
        expected_payload: Expected secret data to verify against.

    Returns:
        True if verification passed, False if any check failed.

    Raises:
        hvac.exceptions.InvalidPath: If secret path is invalid.
        ConnectionError: If Vault connection fails.
    """
    pass
```

---

## Error Handling

### Good Patterns

```python
# Good - Specific exception, logged
try:
    cert_info = validate_certificate(payload["tls.crt"])
except subprocess.CalledProcessError as exc:
    logging.warning("Certificate validation failed: %s", exc)

# Good - Re-raise with context
try:
    result = subprocess.run(cmd, check=True, capture_output=True)
except subprocess.CalledProcessError as exc:
    raise RuntimeError(f"Command failed: {cmd}") from exc
```

### Bad Patterns

```python
# Bad - Bare exception, swallowed
try:
    validate_something()
except Exception:
    pass  # Silent failure!

# Bad - Too broad, catches KeyboardInterrupt
try:
    long_running_task()
except:  # noqa: E722
    pass
```

---

## Logging

### Standard Setup

```python
import logging

# Basic setup for scripts
logging.basicConfig(
    format="%(asctime)s %(levelname)s %(message)s",
    level=logging.INFO,
)

# Module logger for libraries
log = logging.getLogger(__name__)
```

### Good Patterns

```python
# Good - Use % formatting (lazy evaluation)
logging.info("Processing secret: %s", secret_name)
logging.warning("Retry %d of %d: %s", attempt, max_retries, error)

# Bad - f-string (evaluated even if level disabled)
logging.info(f"Processing {expensive_to_compute()}")
```

---

## Testing

### Pytest Structure

```text
tests/
├── conftest.py           # Shared fixtures
├── test_core.py          # Unit tests for core module
├── test_utils.py         # Unit tests for utils
└── e2e/                  # End-to-end tests
    ├── conftest.py       # Testcontainers fixtures
    └── test_integration.py
```

### Test Patterns

```python
# Table-driven tests
import pytest

@pytest.mark.parametrize("input,expected", [
    ("valid@example.com", True),
    ("invalid", False),
    ("", False),
    ("@nodomain", False),
])
def test_validate_email(input: str, expected: bool):
    assert validate_email(input) == expected
```

---

## CLI Script Template

```python
#!/usr/bin/env python3
"""One-line description of what this script does.

Usage:
    python3 script_name.py --config config.yaml --apply

Exit Codes:
    0 - Success
    1 - Argument/configuration error
    2 - Runtime error
"""

from __future__ import annotations

import argparse
import logging
import sys
from pathlib import Path

logging.basicConfig(
    format="%(asctime)s %(levelname)s %(message)s",
    level=logging.INFO,
)


def die(message: str) -> None:
    """Print error message and exit with code 1."""
    logging.error(message)
    sys.exit(1)


def main() -> int:
    """Main entry point."""
    # Main logic here
    return 0


if __name__ == "__main__":
    sys.exit(main())
```

---

## Code Quality Metrics

### Complexity Thresholds

| Grade | CC Range | Action |
|-------|----------|--------|
| A | 1-5 | Ideal - simple, low risk |
| B | 6-10 | Acceptable - moderate complexity |
| C | 11-20 | Refactor when touching |
| D | 21-30 | Must refactor before merge |
| F | 31+ | Block merge |

### Validation Commands

```bash
# Code quality + style
ruff check src/ --statistics

# Complexity analysis
radon cc src/ -s -a

# Enforce complexity limit
xenon src/ --max-absolute B

# Test coverage
pytest --cov=src --cov-report=term-missing
```

---

## Anti-Patterns Avoided

### No God Functions

```python
# Bad - Single function doing everything
def process_all(data):
    # 200+ lines of validation, transformation, saving, logging...
    pass

# Good - Separated concerns
def validate(data: Data) -> ValidationResult: ...
def transform(data: Data) -> TransformedData: ...
def save(data: TransformedData) -> None: ...
```

### No Bare Except

```python
# Bad
try:
    risky_operation()
except:
    pass

# Good
try:
    risky_operation()
except SpecificError as e:
    logging.warning("Operation failed: %s", e)
```

### No Global Mutable State

```python
# Bad
config = {}  # Module-level mutable

# Good
@dataclass
class Config:
    setting_a: str
    setting_b: int

def load_config(path: Path) -> Config:
    data = load_yaml(path)
    return Config(**data)
```

---

## Compliance Assessment

**Use letter grades + evidence, NOT numeric scores.**

### Grading Scale

| Grade | Criteria |
|-------|----------|
| A+ | 0 ruff violations, 0 functions >CC10, 95%+ hints, 90%+ coverage |
| A | <5 ruff violations, <3 functions >CC10, 85%+ hints, 80%+ coverage |
| A- | <15 ruff violations, <8 functions >CC10, 75%+ hints, 70%+ coverage |
| B+ | <30 ruff violations, <15 functions >CC10, 60%+ hints, 60%+ coverage |
| B | <50 ruff violations, <25 functions >CC10, 50%+ hints, 50%+ coverage |
| C | Significant issues, major refactoring needed |
| D | Not production-ready |
| F | Critical issues |

---

## Additional Resources

- [PEP 8 - Style Guide](https://peps.python.org/pep-0008/)
- [Google Python Style Guide](https://google.github.io/styleguide/pyguide.html)
- [ruff Documentation](https://docs.astral.sh/ruff/)
- [radon Complexity](https://radon.readthedocs.io/)
- [pytest Documentation](https://docs.pytest.org/)
