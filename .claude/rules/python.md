# Python Conventions

> Canonical source with full examples: `skills/standards/references/python.md`
> This file is kept self-contained for sessions that don't invoke skills.

## Formatting and Linting

- **Black** formatter with 100-character line length.
- **ruff** linter (`ruff check` must pass).
- **mypy** for type checking.
- Config lives in `pyproject.toml`.

## Style

- Type hints on all public functions.
- Docstrings on all public classes and functions.
- pytest for testing; `conftest.py` for shared fixtures.

## Error Handling

- Never bare `except:` -- always specify the exception type.
- Use `raise ... from e` to preserve stack traces.
- Catch specific exceptions, not `except Exception:`.

## Security

- Never use `eval()`, `exec()`, or `__import__()` with untrusted input.
- Use `secrets` module for tokens, not `random`.
- Validate all external input.

## Testing

- Assert exact expected values (`== expected`), not just `!= wrong`.
- Mock external services, not internal code.
- Add structural invariant tests when adding fields to dataclasses/models.
