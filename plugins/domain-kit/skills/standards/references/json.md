# JSON/JSONL Standards - Tier 1 Quick Reference

<!-- Tier 1: Generic standards (~4KB), always loaded -->
<!-- Tier 2: Deep standards in vibe/references/json-standards.md (~15KB), loaded on --deep -->
<!-- Last synced: 2026-01-21 -->

> **Purpose:** Quick reference for JSON/JSONL standards. For comprehensive patterns, load Tier 2.

---

## Quick Reference

| Standard | Value | Validation |
|----------|-------|------------|
| **Indentation** | 2 spaces | `jq .` or Prettier |
| **Trailing Newline** | Required | Editor config |
| **Trailing Commas** | Not allowed | JSON spec |
| **Comments** | Not allowed | Use JSONC |
| **JSONL Delimiter** | Newline (`\n`) | One object per line |

---

## Format Decision

| Purpose | Format |
|---------|--------|
| Config (needs comments) | JSONC or YAML |
| Config (no comments) | JSON |
| Single record/object | JSON |
| Append-only data | JSONL |
| Logs/events | JSONL |
| Large datasets | JSONL |

---

## Common Errors

| Symptom | Cause | Fix |
|---------|-------|-----|
| `Unexpected token` | Trailing comma | Remove last comma |
| `Unexpected token '` | Single quotes | Use double quotes |
| `Unexpected token /` | Comments in JSON | Use JSONC |
| `Invalid character` | BOM/wrong encoding | UTF-8 without BOM |
| `Unexpected end` | Truncated file | Check structure |
| JSONL parse error | Multi-line object | One object per line |

---

## Anti-Patterns

| Name | Pattern | Instead |
|------|---------|---------|
| Minified Config | `{"a":1,"b":2}` | Pretty-print, 2 spaces |
| Comments in JSON | `// comment` | Use JSONC |
| Mixed Key Naming | camelCase + snake_case | Pick one per file |
| Magic Numbers | `"priority": 2` | Document or use enums |
| No Schema | Large config | Add JSON Schema |

---

## Key Naming

| Convention | Use For | Example |
|------------|---------|---------|
| `camelCase` | JavaScript/TypeScript | `"apiVersion"` |
| `snake_case` | Python, beads | `"issue_type"` |
| `UPPER_CASE` | Environment vars only | `"DATABASE_URL"` |

---

## JSONL Format

```jsonl
{"id": "abc-123", "status": "open", "title": "First"}
{"id": "abc-124", "status": "closed", "title": "Second"}
```

**Rules:**
- One valid JSON object per line
- No array wrapper
- Newline after last record
- UTF-8, no BOM

---

## Processing Commands

```bash
# Validate JSON
jq empty config.json && echo "Valid"

# Format JSON
jq . config.json > formatted.json

# JSONL: filter by field
jq -c 'select(.status == "open")' data.jsonl

# JSONL: count records
wc -l data.jsonl

# JSONL: validate each line
while read -r line; do echo "$line" | jq empty; done < data.jsonl
```

---

## Summary Checklist

| Category | Requirement |
|----------|-------------|
| **Indentation** | 2 spaces |
| **Quotes** | Double quotes only |
| **Trailing** | No commas, yes newline |
| **Encoding** | UTF-8, no BOM |
| **Keys** | Consistent naming |
| **JSONL** | One object per line |
| **Schema** | For large/shared config |

---

## Talos Prescan Checks

| Check | Pattern | Rationale |
|-------|---------|-----------|
| PRE-002 | TODO/FIXME markers | Track technical debt |
| PRE-018 | Invalid JSON syntax | Parse errors |
| PRE-019 | Trailing commas | JSON spec violation |

---

## JIT Loading

**Tier 2 (Deep Standards):** For comprehensive patterns including:
- Beads JSONL schema reference
- JSON Schema definitions
- Configuration file templates
- Validation commands
- Compliance assessment details

Load: `vibe/references/json-standards.md`
