# Python Kit

> Python development standards and tooling for AgentOps.

## Install

```bash
/plugin install python-kit@agentops
```

Requires: `solo-kit`

## What's Included

### Standards

Comprehensive Python coding standards in `skills/standards/references/python.md`:
- PEP 8 conventions
- Type hints and mypy
- Error handling patterns
- Testing with pytest
- Async/await patterns
- Common anti-patterns to avoid

### Hooks

| Hook | Trigger | What It Does |
|------|---------|--------------|
| `ruff-format` | Edit *.py | Auto-format with ruff |
| `ruff-check` | Edit *.py | Lint with ruff |
| `mypy-check` | Edit *.py | Type check with mypy |

### Patterns

**Testing (pytest)**
```python
import pytest

@pytest.fixture
def client():
    return TestClient()

def test_endpoint(client):
    response = client.get("/")
    assert response.status_code == 200
```

**Error Handling**
```python
class DomainError(Exception):
    """Base for domain errors."""
    pass

def process(data: dict) -> Result:
    try:
        return Result(success=True, data=transform(data))
    except ValidationError as e:
        return Result(success=False, error=str(e))
```

**Type Hints**
```python
from typing import Optional, List

def find_users(
    query: str,
    limit: Optional[int] = None
) -> List[User]:
    ...
```

## Requirements

- Python 3.10+
- Optional: ruff, mypy (for hooks)

## License

MIT
