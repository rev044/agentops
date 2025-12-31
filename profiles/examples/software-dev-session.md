# Example Session: Software Development

**Profile**: software-dev
**Scenario**: Build a REST API endpoint with tests
**Duration**: ~2 hours

---

## Session Flow

### 1. Research Phase (20 min)

```
User: /research "REST API authentication patterns"
```

**Claude loads**: code-explorer, doc-explorer

**Actions**:
- Explore existing auth patterns in codebase
- Find similar implementations
- Document constraints and requirements

**Output**: Research bundle with recommended approach

---

### 2. Planning Phase (15 min)

```
User: /plan "implement JWT auth endpoint"
```

**Claude loads**: backend-architect, spec-architect

**Actions**:
- Design API endpoint structure
- Define request/response schemas
- Plan middleware integration
- Specify file locations with line numbers

**Output**: Implementation plan with file:line specs

---

### 3. Implementation Phase (60 min)

```
User: /implement
```

**Claude loads**: python-pro (or language-specific agent)

**Actions**:
- Create auth middleware
- Implement JWT token generation
- Add route handlers
- Write error handling

**Agents used**:
1. `python-pro` - Core implementation
2. `backend-architect` - API design decisions
3. `deployment-engineer` - Docker/CI updates

---

### 4. Quality Phase (20 min)

```
User: Run code review and generate tests
```

**Claude loads**: code-reviewer, test-generator

**Actions**:
- Review implementation for security
- Check for edge cases
- Generate unit tests
- Generate integration tests

**Output**: Review feedback + test files

---

### 5. Commit Phase (5 min)

```
User: Commit these changes
```

**Actions**:
- Stage files
- Create semantic commit message
- Push to branch

---

## Agents Used Summary

| Agent | When | Purpose |
|-------|------|---------|
| code-explorer | Research | Understand existing patterns |
| doc-explorer | Research | Find documentation |
| backend-architect | Plan | API design |
| spec-architect | Plan | File:line specifications |
| python-pro | Implement | Write Python code |
| code-reviewer | Quality | Review for issues |
| test-generator | Quality | Create tests |

---

## Session Outcome

- ✅ JWT auth endpoint implemented
- ✅ Middleware integrated
- ✅ Tests passing
- ✅ Code reviewed
- ✅ Committed with semantic message

**Time**: ~2 hours (vs ~6 hours without AI assistance)
