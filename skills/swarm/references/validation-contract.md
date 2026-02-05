# Validation Contract (Moved)

Source of truth: `skills/shared/validation-contract.md`

This file remains as a compatibility shim for older links and references.

3. Continue with other tasks (don't block entire swarm)

---

## Default Validation

When no explicit validation is specified, apply minimal checks:

```python
def default_validation(task_id):
    # Check agent didn't end with errors
    # (parse task notification for failure indicators)

    # Check for uncommitted changes
    result = subprocess.run("git status --porcelain", shell=True, capture_output=True)
    if result.stdout.strip():
        # Uncommitted changes - agent should have committed
        return WARN("Uncommitted changes detected")

    # Check most recent commit references task
    result = subprocess.run("git log -1 --oneline", shell=True, capture_output=True)
    if str(task_id) not in result.stdout.decode():
        return WARN("Recent commit doesn't reference task")

    return PASS
```

---

## Integration with Crank

When crank invokes swarm, it can specify validation at the epic level:

```python
# Crank creates tasks from beads issues
for issue in ready_issues:
    TaskCreate(
        subject=f"{issue.id}: {issue.title}",
        description=issue.description,
        metadata={
            "beads_id": issue.id,
            "validation": build_validation_from_issue(issue)
        }
    )
```

### Building Validation from Issue

```python
def build_validation_from_issue(issue):
    validation = {}

    # Check for test requirements in issue
    if "test" in issue.labels or "tests/" in issue.description:
        validation["tests"] = detect_test_command(issue)

    # Check for file creation requirements
    files_mentioned = extract_file_paths(issue.description)
    if files_mentioned:
        validation["files_exist"] = files_mentioned

    # Check for function/method requirements
    patterns = extract_code_patterns(issue.description)
    if patterns:
        validation["content_check"] = patterns

    return validation
```

---

## Examples

### Example 1: New Feature with Tests

```
TaskCreate(
  subject="Add user authentication",
  description="Implement JWT-based authentication...",
  metadata={
    "validation": {
      "files_exist": [
        "src/auth/jwt.py",
        "src/auth/__init__.py",
        "tests/test_auth.py"
      ],
      "content_check": [
        {"file": "src/auth/jwt.py", "pattern": "def create_token"},
        {"file": "src/auth/jwt.py", "pattern": "def verify_token"}
      ],
      "tests": "pytest tests/test_auth.py -v",
      "lint": "ruff check src/auth/"
    }
  }
)
```

### Example 2: Bug Fix

```
TaskCreate(
  subject="Fix null pointer in user lookup",
  description="Handle case where user not found...",
  metadata={
    "validation": {
      "content_check": {
        "file": "src/users/lookup.py",
        "pattern": "if user is None"
      },
      "tests": "pytest tests/test_users.py::test_user_not_found -v"
    }
  }
)
```

### Example 3: Documentation Update

```
TaskCreate(
  subject="Update API docs for v2",
  description="Update README with new endpoints...",
  metadata={
    "validation": {
      "files_exist": ["docs/api/v2.md"],
      "content_check": {
        "file": "docs/api/v2.md",
        "pattern": "## Authentication"
      }
    }
  }
)
```

### Example 4: Infrastructure Change

```
TaskCreate(
  subject="Add Redis caching layer",
  description="Configure Redis for session caching...",
  metadata={
    "validation": {
      "files_exist": ["docker-compose.yml", "src/cache/redis.py"],
      "command": "docker-compose config --quiet",
      "content_check": {
        "file": "docker-compose.yml",
        "pattern": "redis:"
      }
    }
  }
)
```

---

## Validation in Distributed Mode

In distributed mode (tmux + Agent Mail), validation works the same but with message-based coordination:

1. **Demigod completes work** -> sends `OFFERING_READY` message
2. **Mayor receives message** -> runs validation checks
3. **On PASS** -> sends `OFFERING_ACCEPTED`, closes beads issue
4. **On FAIL** -> sends `RETRY_REQUIRED` with failure context

See `skills/swarm/SKILL.md` for distributed mode details.

---

## See Also

- `skills/swarm/SKILL.md` - Main swarm skill with validation integration
- `skills/crank/SKILL.md` - Crank orchestration with validation loop
- `skills/crank/failure-taxonomy.md` - Comprehensive failure handling
- `skills/vibe/SKILL.md` - Comprehensive validation skill
