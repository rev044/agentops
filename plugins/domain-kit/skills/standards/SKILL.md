---
name: standards
description: Library skill providing language standards, style guides, and best practices. Not invoked directly - other skills reference these docs.
version: 1.0.0
triggers: []
allowed-tools:
  - Read
context: fork
---

# Standards Library

> **This is a library skill.** It provides reference documentation for other skills.
> Not typically invoked directly by users.

## Purpose

Central repository of coding standards, style guides, and best practices that other skills can reference for language-specific validation, implementation guidance, and quality checks.

## Available References

### Language Standards

| Reference | Path | Use When |
|-----------|------|----------|
| **Python** | `references/python.md` | `.py` files |
| **Go** | `references/go.md` | `.go` files |
| **TypeScript** | `references/typescript.md` | `.ts`, `.tsx` files |
| **Shell** | `references/shell.md` | `.sh`, `.bash` files |
| **YAML/Helm** | `references/yaml.md` | `.yaml`, `.yml` files |
| **Markdown** | `references/markdown.md` | `.md` files |
| **JSON/JSONL** | `references/json.md` | `.json`, `.jsonl` files |
| **Tags** | `references/tags.md` | `.agents/` documents |

### Knowledge Artifact Standards

| Reference | Path | Use When |
|-----------|------|----------|
| **RAG Formatting** | `references/rag-formatting.md` | Research, learnings, retros, patterns |

### API & Platform Standards

| Reference | Path | Use When |
|-----------|------|----------|
| **OpenAI** | `references/openai.md` | Building OpenAI-powered agents |
| **OpenAI Prompts** | `references/openai-prompts.md` | Prompt engineering for OpenAI models |
| **OpenAI Functions** | `references/openai-functions.md` | Function calling, tool definitions |
| **OpenAI Responses** | `references/openai-responses.md` | Responses API, agent orchestration |
| **OpenAI Reasoning** | `references/openai-reasoning.md` | o3/o4-mini reasoning models |
| **OpenAI GPT-OSS** | `references/openai-gptoss.md` | GPT-OSS-120B/20B open-weight models |

## How Other Skills Use This

### Declare Dependency

```yaml
---
name: my-skill
skills:
  - standards
---
```

### Load Relevant Reference

```markdown
## When Validating Python

Load `domain-kit/skills/standards/references/python.md` for:
- Common Errors table (symptom -> cause -> fix)
- Anti-Patterns (named patterns to avoid)
- AI Agent Guidelines (ALWAYS/NEVER rules)
```

### Detection Pattern

```markdown
## Language Detection

| File Pattern | Load Reference |
|--------------|----------------|
| `*.py` | `domain-kit/skills/standards/references/python.md` |
| `*.go` | `domain-kit/skills/standards/references/go.md` |
| `*.ts`, `*.tsx` | `domain-kit/skills/standards/references/typescript.md` |
| `*.sh` | `domain-kit/skills/standards/references/shell.md` |
| `*.yaml`, `*.yml` | `domain-kit/skills/standards/references/yaml.md` |
| `*.md` | `domain-kit/skills/standards/references/markdown.md` |
| `*.json`, `*.jsonl` | `domain-kit/skills/standards/references/json.md` |

## OpenAI Integration Detection

| Context | Load Reference |
|---------|----------------|
| Building OpenAI agents | `domain-kit/skills/standards/references/openai.md` (overview) |
| Prompt engineering | `domain-kit/skills/standards/references/openai-prompts.md` |
| Function definitions | `domain-kit/skills/standards/references/openai-functions.md` |
| Responses API usage | `domain-kit/skills/standards/references/openai-responses.md` |
| o3/o4-mini models | `domain-kit/skills/standards/references/openai-reasoning.md` |
| GPT-OSS-120B/20B models | `domain-kit/skills/standards/references/openai-gptoss.md` |

## Knowledge Artifact Detection

| Output Type | Load Reference | Apply To |
|-------------|----------------|----------|
| Research artifacts | `domain-kit/skills/standards/references/rag-formatting.md` | `.agents/research/` |
| Learning artifacts | `domain-kit/skills/standards/references/rag-formatting.md` | `.agents/learnings/` |
| Retro artifacts | `domain-kit/skills/standards/references/rag-formatting.md` | `.agents/retros/` |
| Pattern artifacts | `domain-kit/skills/standards/references/rag-formatting.md` | `.agents/patterns/` |

**Key requirements for knowledge artifacts:**
- 200-400 chars per H2 section (embedding sweet spot)
- Frontmatter with `type` and `tier` fields
- NO `confidence` or `relevance` fields (query-time, not storage-time)
- Action-oriented headings, front-loaded key terms
```

## Reference Structure

Each reference follows a consistent format optimized for AI agent consumption:

```markdown
# Language Standard

## Quick Reference
[Key rules table]

## Common Errors
| Symptom | Cause | Fix |
[Troubleshooting lookup]

## Anti-Patterns
| Name | Pattern | Why Bad | Instead |
[Named patterns to recognize and avoid]

## AI Agent Guidelines
| Guideline | Rationale |
[ALWAYS/NEVER rules for agents]
```

## Adding New Standards

1. Create `references/<language>.md`
2. Follow the structure above (Quick Reference, Common Errors, Anti-Patterns, AI Guidelines)
3. Update this SKILL.md's reference table
4. Update dependent skills to reference the new standard
