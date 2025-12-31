# Example Session: Content Creation

**Profile**: content-creation
**Scenario**: Extract pattern from production code, create documentation
**Duration**: ~3 hours

---

## Session Flow

### 1. Research Phase (45 min)

```
User: /research "authentication patterns in our codebase"
```

**Claude loads**: code-explorer, history-explorer, doc-explorer

**Actions**:
- Scan codebase for auth implementations
- Trace git history for evolution
- Find existing documentation
- Identify common patterns

**Findings**:
- 3 different auth approaches used
- JWT most common (12 services)
- OAuth for external integrations
- API keys for internal services

---

### 2. Pattern Extraction (30 min)

```
User: Extract the JWT auth pattern
```

**Claude loads**: meta-retro-analyzer, spec-architect

**Actions**:
- Identify core components
- Document decision rationale
- Capture variations
- Note failure modes

**Output**: Pattern specification

---

### 3. Documentation Writing (60 min)

```
User: Create documentation for this pattern
```

**Claude loads**: documentation-create-docs, api-documenter

**Actions**:
- Write pattern overview
- Create usage examples
- Document configuration options
- Add troubleshooting section

**Outputs**:
1. `docs/patterns/jwt-authentication.md`
2. `docs/how-to/implement-jwt-auth.md`
3. `docs/reference/auth-api.md`

---

### 4. Quality Review (20 min)

```
User: Review docs for Diátaxis compliance
```

**Claude loads**: documentation-diataxis-auditor

**Actions**:
- Check document placement
- Verify content type matches location
- Ensure cross-references work
- Validate completeness

**Feedback**: Minor adjustments to how-to guide

---

### 5. Retrospective (25 min)

```
User: /retro "pattern extraction session"
```

**Claude loads**: meta-retro-analyzer

**Actions**:
- Capture what worked
- Document challenges
- Extract reusable insights
- Update institutional memory

**Output**: Retrospective in `.agents/retros/`

---

## Agents Used Summary

| Agent | When | Purpose |
|-------|------|---------|
| code-explorer | Research | Find implementations |
| history-explorer | Research | Trace evolution |
| doc-explorer | Research | Find existing docs |
| meta-retro-analyzer | Extract | Pattern synthesis |
| spec-architect | Extract | Specification |
| documentation-create-docs | Write | Create docs |
| api-documenter | Write | API reference |
| documentation-diataxis-auditor | Review | Compliance check |

---

## Session Outcome

- ✅ Pattern extracted and documented
- ✅ 3 documentation files created
- ✅ Diátaxis compliant
- ✅ Retrospective captured

**Time**: ~3 hours (pattern now reusable for team)

---

## Artifacts Created

```
docs/
├── patterns/
│   └── jwt-authentication.md     # Pattern overview
├── how-to/
│   └── implement-jwt-auth.md     # Step-by-step guide
└── reference/
    └── auth-api.md               # API documentation

.agents/
├── research/
│   └── 2025-12-30-auth-patterns.md
├── patterns/
│   └── jwt-auth-pattern.md
└── retros/
    └── 2025-12-30-pattern-extraction.md
```
