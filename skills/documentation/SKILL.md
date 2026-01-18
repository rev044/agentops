---
name: documentation
description: >
  Use when: "document", "docs", "README", "API docs", "OpenAPI", "Swagger", "Diátaxis",
  "tutorial", "how-to", "reference", "explanation", "SDK", "developer docs",
  "knowledge OS", "7-pattern stack".
version: 1.0.0
author: "AgentOps Team"
license: "MIT"
---

# Documentation Skill

Documentation creation, optimization, and API documentation patterns.

## Quick Reference

| Area | Key Patterns | When to Use |
|------|--------------|-------------|
| **Create** | 7-pattern stack, Diátaxis | New documentation |
| **Optimize** | Knowledge OS standards | Refactoring docs |
| **Audit** | Diátaxis compliance | Quality checks |
| **API** | OpenAPI, SDKs, examples | Developer docs |

---

## Diátaxis Framework

Documentation organized by user need:

| Type | Purpose | User Need |
|------|---------|-----------|
| **Tutorial** | Learning-oriented | "Teach me" |
| **How-to** | Task-oriented | "Help me do X" |
| **Reference** | Information-oriented | "Tell me about X" |
| **Explanation** | Understanding-oriented | "Explain why" |

### Placement Rules

```
docs/
├── tutorials/     # Learning journeys (ordered steps)
├── how-to/        # Task guides (problem → solution)
├── reference/     # Technical specs (accurate, complete)
└── explanation/   # Concepts (why things work)
```

### Content Types

**Tutorial** (teaches):
- Ordered steps for beginners
- "Build your first X"
- Focus on learning, not speed

**How-to** (guides):
- Problem-focused
- "How to configure X for Y"
- Assumes basic knowledge

**Reference** (describes):
- Complete, accurate specs
- API documentation
- Configuration options

**Explanation** (explains):
- Conceptual understanding
- Architecture decisions
- Trade-offs and alternatives

---

## 7-Pattern Documentation Stack

| Pattern | Purpose | Location |
|---------|---------|----------|
| **README** | Entry point | Root |
| **CLAUDE.md** | AI context | Root |
| **Tutorials** | Learning | docs/tutorials/ |
| **How-to** | Tasks | docs/how-to/ |
| **Reference** | Specs | docs/reference/ |
| **Explanation** | Concepts | docs/explanation/ |
| **ADRs** | Decisions | docs/decisions/ |

---

## Documentation Creation

### Approach
1. Identify documentation type (Diátaxis)
2. Match content to user need
3. Use consistent structure
4. Include working examples
5. Cross-reference related docs

### Templates

**Tutorial Template:**
```markdown
# Tutorial: Build Your First X

## Prerequisites
- [List what they need]

## What You'll Learn
- [Learning objectives]

## Step 1: [First step]
[Detailed instructions]

## Step 2: [Second step]
[Detailed instructions]

## Summary
[What they accomplished]

## Next Steps
[Where to go next]
```

**How-to Template:**
```markdown
# How to [Do Thing]

## Problem
[What problem this solves]

## Solution
[Step-by-step solution]

## Example
[Working example]

## Troubleshooting
[Common issues]
```

**Reference Template:**
```markdown
# [Component] Reference

## Overview
[Brief description]

## Configuration
| Option | Type | Default | Description |
|--------|------|---------|-------------|

## Methods/API
### `methodName(params)`
[Description and examples]

## Examples
[Usage examples]
```

---

## Documentation Optimization

### Knowledge OS Standards

1. **Single source of truth** - No duplicate information
2. **Discoverable** - Clear navigation and search
3. **Maintainable** - Easy to update
4. **Versioned** - Tied to code versions
5. **Tested** - Examples that work

### Optimization Checklist

- [ ] Correct Diátaxis placement
- [ ] No content duplication
- [ ] Working code examples
- [ ] Cross-references work
- [ ] Consistent formatting
- [ ] Clear headings

### Common Issues

| Issue | Fix |
|-------|-----|
| Tutorial in reference/ | Move to tutorials/ |
| Duplicate content | Link instead |
| Broken examples | Test and fix |
| Missing context | Add prerequisites |

---

## API Documentation

### OpenAPI Best Practices

```yaml
openapi: 3.0.3
info:
  title: API Name
  version: 1.0.0
  description: |
    Brief description of what the API does.

    ## Authentication
    Use Bearer token in Authorization header.

paths:
  /resource:
    get:
      summary: List resources
      description: Returns paginated list of resources
      parameters:
        - name: limit
          in: query
          schema:
            type: integer
            default: 20
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ResourceList'
              example:
                items: [...]
                total: 100
```

### SDK Documentation

1. **Quick Start** - Get working in 5 minutes
2. **Installation** - All platforms
3. **Authentication** - How to configure
4. **Examples** - Common use cases
5. **API Reference** - All methods
6. **Error Handling** - What can go wrong
7. **Changelog** - Version history

### Example Structure

```markdown
# SDK Name

## Installation
```bash
npm install @org/sdk
```

## Quick Start
```typescript
import { Client } from '@org/sdk';

const client = new Client({ apiKey: 'your-key' });
const result = await client.resource.list();
```

## Authentication
[Authentication details]

## Resources
### resource.list()
[Method documentation]

## Error Handling
[Error types and handling]
```

---

## Diátaxis Audit Checklist

### Content Type Validation

| Check | Pass | Fail |
|-------|------|------|
| Tutorials teach, don't reference | ✅ | ❌ Has reference tables |
| How-tos solve problems | ✅ | ❌ Teaches basics |
| Reference is complete | ✅ | ❌ Missing options |
| Explanations explain why | ✅ | ❌ Just describes what |

### Structure Validation

| Check | Pass | Fail |
|-------|------|------|
| Files in correct directory | ✅ | ❌ Tutorial in reference/ |
| Cross-references work | ✅ | ❌ Broken links |
| No duplicate content | ✅ | ❌ Same info in 2 places |
| Consistent formatting | ✅ | ❌ Mixed styles |

### Completeness

| Check | Pass | Fail |
|-------|------|------|
| README exists | ✅ | ❌ Missing |
| All features documented | ✅ | ❌ Gaps |
| Examples work | ✅ | ❌ Broken |
| Up to date | ✅ | ❌ Stale |
