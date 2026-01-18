---
name: code-quality
description: >
  Use when: "code review", "review", "PR review", "test", "testing", "unit test",
  "integration test", "test generation", "coverage", "quality", "maintainability",
  "security review", "edge cases".
version: 1.0.0
author: "AgentOps Team"
license: "MIT"
---

# Code Quality Skill

Code review, test generation, and quality assurance patterns.

## Quick Reference

| Area | Key Patterns | When to Use |
|------|--------------|-------------|
| **Review** | Security, quality, maintainability | After code changes |
| **Improve** | Systematic analysis, refactoring | Code optimization |
| **Test** | Unit, integration, edge cases | Test creation |

---

## Code Review

### Approach
1. Run git diff to see recent changes
2. Focus on modified files
3. Begin review immediately

### Review Checklist

| Category | Checks |
|----------|--------|
| **Readability** | Simple, well-named, no duplication |
| **Correctness** | Proper error handling, edge cases |
| **Security** | No secrets, input validation |
| **Performance** | Efficient algorithms, no N+1 |
| **Testing** | Good coverage, meaningful tests |

### Feedback Organization

**Critical** (must fix):
- Security vulnerabilities
- Data loss risks
- Breaking bugs

**Warnings** (should fix):
- Performance issues
- Maintainability concerns
- Missing tests

**Suggestions** (consider):
- Style improvements
- Alternative approaches
- Documentation additions

### Review Template

```markdown
## Code Review: [PR/commit]

### Summary
[Brief overview of changes]

### Critical Issues
1. **[Issue]** - [File:line]
   - Problem: [Description]
   - Fix: [How to fix]

### Warnings
1. **[Issue]** - [File:line]
   - Problem: [Description]
   - Suggestion: [How to improve]

### Suggestions
1. **[Suggestion]** - [File:line]
   - [Optional improvement]

### Positive Notes
- [What's done well]
```

---

## Systematic Code Improvement

### Analysis Dimensions

| Dimension | Checks |
|-----------|--------|
| **Correctness** | Logic bugs, edge cases, error handling |
| **Maintainability** | Complexity, coupling, naming |
| **Security** | OWASP Top 10, input validation |
| **Performance** | Algorithms, memory, I/O |

### Complexity Metrics

| Metric | Good | Warning | Bad |
|--------|------|---------|-----|
| Cyclomatic Complexity | < 10 | 10-20 | > 20 |
| Function Length | < 50 lines | 50-100 | > 100 |
| Nesting Depth | < 4 | 4-6 | > 6 |
| Parameters | < 5 | 5-7 | > 7 |

### Refactoring Patterns

| Problem | Pattern |
|---------|---------|
| Long function | Extract method |
| Deep nesting | Early return, guard clauses |
| Parameter explosion | Parameter object |
| Duplicated code | Extract shared function |
| Complex conditional | Strategy pattern |

---

## Test Generation

### Test Types

| Type | Purpose | Scope |
|------|---------|-------|
| **Unit** | Single function/class | Isolated |
| **Integration** | Component interaction | Combined |
| **E2E** | Full user flow | System |

### Coverage Targets

| Type | Minimum | Target |
|------|---------|--------|
| Unit | 80% | 90% |
| Integration | 60% | 80% |
| Critical paths | 100% | 100% |

### Test Structure

```python
# Arrange-Act-Assert pattern
def test_should_do_expected_behavior():
    # Arrange - Set up test conditions
    input_data = create_test_data()

    # Act - Execute the code under test
    result = function_under_test(input_data)

    # Assert - Verify expectations
    assert result == expected_output
```

### Edge Cases to Test

| Category | Examples |
|----------|----------|
| **Boundary** | 0, 1, max, min, empty |
| **Invalid** | null, undefined, wrong type |
| **Error** | Network failure, timeout |
| **Concurrent** | Race conditions, deadlocks |
| **State** | Initial, transitional, final |

### Test Generation Template

```python
import pytest

class TestFeatureName:
    """Tests for FeatureName functionality."""

    # Happy path tests
    def test_basic_functionality(self):
        """Test normal expected behavior."""
        pass

    # Edge cases
    @pytest.mark.parametrize("input,expected", [
        (0, "zero case"),
        (-1, "negative case"),
        (MAX_INT, "max case"),
    ])
    def test_boundary_conditions(self, input, expected):
        """Test boundary conditions."""
        pass

    # Error handling
    def test_handles_invalid_input(self):
        """Test error handling for invalid input."""
        with pytest.raises(ValueError):
            function_under_test(invalid_input)

    # Integration points
    def test_integration_with_dependency(self, mock_dependency):
        """Test interaction with external dependency."""
        pass
```

---

## Security Review

### OWASP Top 10 Checks

| Vulnerability | Check |
|---------------|-------|
| **Injection** | Parameterized queries, input sanitization |
| **Broken Auth** | Session management, password handling |
| **Sensitive Data** | Encryption, no secrets in code |
| **XXE** | Disable external entities |
| **Access Control** | Authorization checks |
| **Misconfig** | Secure defaults |
| **XSS** | Output encoding |
| **Deserialization** | Validate untrusted data |
| **Components** | Updated dependencies |
| **Logging** | Audit logs, no sensitive data |

### Security Checklist

```markdown
## Security Review

### Input Validation
- [ ] All user input validated
- [ ] Parameterized queries used
- [ ] File uploads restricted

### Authentication
- [ ] Strong password requirements
- [ ] Session timeout configured
- [ ] MFA available

### Authorization
- [ ] Role-based access control
- [ ] Resource-level permissions
- [ ] Principle of least privilege

### Data Protection
- [ ] Sensitive data encrypted
- [ ] No secrets in code
- [ ] HTTPS enforced

### Dependencies
- [ ] No known vulnerabilities
- [ ] Pinned versions
- [ ] Regular updates planned
```

---

## Quality Metrics

### Code Health Dashboard

| Metric | Threshold | Measurement |
|--------|-----------|-------------|
| Test Coverage | > 80% | coverage report |
| Complexity | CC < 10 | radon, complexity tools |
| Duplication | < 5% | sonar, jscpd |
| Tech Debt | < 2h/kloc | time to fix issues |
| Security | 0 critical | security scanner |

### Automated Quality Gates

```yaml
# Example CI quality gate
quality:
  - coverage: 80%
  - complexity: 10
  - duplication: 5%
  - security: 0-critical
  - linting: 0-errors
```
