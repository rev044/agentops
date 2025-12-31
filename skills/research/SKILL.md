---
name: research
description: >
  Use when: "explore", "find", "search", "analyze", "understand", "git history",
  "pattern", "dependency", "documentation search", "specification", "file:line",
  "archive", "duplication detection", "structure analysis".
version: 1.0.0
---

# Research Skill

Codebase exploration, pattern detection, and specification design.

## Quick Reference

| Area | Key Patterns | When to Use |
|------|--------------|-------------|
| **Code Explorer** | Structure, dependencies, patterns | Understanding codebase |
| **Doc Explorer** | Documentation search, synthesis | Finding existing docs |
| **History Explorer** | Git archaeology, decisions | Understanding evolution |
| **Archive Researcher** | Duplication, archival analysis | Repository cleanup |
| **Structure Analyzer** | Layout, hierarchy | Document analysis |
| **Spec Architect** | file:line specifications | Precise planning |

---

## Code Exploration

### Approach
1. Map high-level structure (directories, modules)
2. Identify entry points and core components
3. Trace dependencies and data flow
4. Document patterns and conventions
5. Note areas of complexity

### Commands

```bash
# Structure overview
tree -L 2 -d

# Find entry points
grep -r "if __name__" --include="*.py"
grep -r "func main" --include="*.go"

# Dependency analysis
grep -r "import" --include="*.py" | cut -d: -f2 | sort | uniq -c | sort -rn

# Pattern detection
grep -rn "class.*:" --include="*.py" | head -20

# Complexity hotspots
find . -name "*.py" -exec wc -l {} \; | sort -rn | head -10
```

### Output Template

```markdown
# Code Exploration: [Repository/Component]

## Structure
```
src/
├── core/          # [purpose]
├── api/           # [purpose]
└── utils/         # [purpose]
```

## Entry Points
- `main.py` - Application entry
- `api/server.py` - HTTP server

## Key Dependencies
- [library] - [usage]
- [library] - [usage]

## Patterns Identified
- [Pattern 1] - [where used]
- [Pattern 2] - [where used]

## Complexity Notes
- [High complexity area] - [why]
```

---

## Documentation Search

### Approach
1. Locate all documentation files
2. Index by topic and type
3. Identify gaps and overlaps
4. Synthesize findings

### Commands

```bash
# Find all docs
find . -name "*.md" -type f

# Search for topic
grep -rn "authentication" --include="*.md"

# List READMEs
find . -name "README.md"

# Find code comments
grep -rn "TODO\|FIXME\|NOTE" --include="*.py"
```

### Output Template

```markdown
# Documentation Search: [Topic]

## Relevant Documents
| File | Relevance | Summary |
|------|-----------|---------|
| docs/auth.md | High | Authentication flow |
| README.md | Medium | Mentions setup |

## Key Findings
- [Finding 1]
- [Finding 2]

## Gaps Identified
- No documentation for [X]
- Outdated info in [Y]

## Synthesis
[Combined understanding from all sources]
```

---

## Git History Analysis

### Approach
1. Identify significant commits
2. Track file evolution
3. Find decision points
4. Document patterns over time

### Commands

```bash
# Recent history
git log --oneline -20

# File history
git log --oneline -- [file]

# Who worked on what
git shortlog -sn

# Recent changes to file
git log -p --follow -- [file] | head -100

# Find when something was added
git log -S "search term" --oneline

# Blame for context
git blame [file] | head -20
```

### Output Template

```markdown
# Git History Analysis: [Component]

## Timeline
| Date | Commit | Change |
|------|--------|--------|
| 2024-01 | abc123 | Initial implementation |
| 2024-03 | def456 | Major refactor |

## Key Decisions
- [Date]: Switched from X to Y because [reason]
- [Date]: Added feature Z for [purpose]

## Contributors
- [Name] - [area of focus]

## Patterns Over Time
- [Evolution pattern]
- [Recurring issues]
```

---

## Archive Analysis

### Approach
1. Analyze repository structure
2. Detect duplication with other repos
3. Check references and dependencies
4. Assess archival candidacy

### Duplication Detection

```bash
# Compare directories
diff -rq [repo1]/commands/ [repo2]/commands/

# Find similar files
find [repo] -name "*.md" -exec basename {} \; | sort > files1.txt
find [other] -name "*.md" -exec basename {} \; | sort > files2.txt
comm -12 files1.txt files2.txt

# Content similarity
diff [repo]/file.md [other]/file.md
```

### Output Template

```markdown
# Archive Analysis: [Repository]

## Recommendation: [ARCHIVE | KEEP | CONSOLIDATE]

## Duplication
| Content | This Repo | Also In | Status |
|---------|-----------|---------|--------|
| commands/ | 10 files | ~/.claude | 100% dup |
| agents/ | 5 files | work/ops | 80% dup |

## Unique Content
- [file] - [why unique]

## References to This Repo
- [file:line] - [type of reference]

## Risk Assessment
- **Impact**: [None | Low | Medium | High]
- **Recovery**: Git archive branch available
```

---

## Specification Design

### file:line Precision

Specifications should reference exact locations:

```markdown
## Implementation Spec

### File Changes

1. **src/api/routes.py:45-60**
   - Add new endpoint handler
   - Follow existing pattern at line 30

2. **src/models/user.py:12**
   - Add new field to User model
   - Pattern: see line 8-11

3. **tests/test_api.py:EOF**
   - Add test cases for new endpoint
```

### Spec Template

```markdown
# Specification: [Feature Name]

## Overview
[Brief description]

## Requirements
1. [Requirement 1]
2. [Requirement 2]

## Implementation

### Phase 1: [Name]
| File | Line | Change |
|------|------|--------|
| src/file.py | 45 | Add function |
| src/file.py | 100-110 | Modify class |

### Phase 2: [Name]
[Same format]

## Testing Strategy
| Test | File | Purpose |
|------|------|---------|
| Unit | tests/test_x.py | Core logic |
| Integration | tests/integration/ | API flow |

## Dependencies
- [Dependency 1] - [why needed]

## Risks
- [Risk 1] - [mitigation]
```

---

## Document Structure Analysis

### Approach
1. Parse document hierarchy
2. Map heading structure
3. Identify semantic patterns
4. Detect inconsistencies

### Analysis Points

| Aspect | Check |
|--------|-------|
| Hierarchy | Proper H1 → H2 → H3 nesting |
| Consistency | Same patterns across sections |
| Completeness | All sections have content |
| Links | Internal references valid |

### Output Template

```markdown
# Structure Analysis: [Document]

## Hierarchy Map
```
H1: Title
├── H2: Section 1
│   ├── H3: Subsection
│   └── H3: Subsection
└── H2: Section 2
```

## Patterns Detected
- [Pattern 1]
- [Pattern 2]

## Issues Found
- [ ] Missing H2 section for [topic]
- [ ] Inconsistent heading levels at [location]

## Recommendations
- [Improvement 1]
- [Improvement 2]
```
