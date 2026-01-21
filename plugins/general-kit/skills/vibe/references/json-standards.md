# JSON/JSONL Standards Catalog - Vibe Canonical Reference

**Version:** 1.0.0
**Last Updated:** 2026-01-21
**Purpose:** Canonical JSON/JSONL standards for vibe skill validation

---

## Table of Contents

1. [JSON Formatting](#json-formatting)
2. [JSONL Format](#jsonl-format)
3. [Configuration Files](#configuration-files)
4. [JSON Schema](#json-schema)
5. [Tooling](#tooling)
6. [Compliance Assessment](#compliance-assessment)

---

## JSON Formatting

### Standard Format

```json
{
  "name": "example",
  "version": "1.0.0",
  "config": {
    "timeout": 30,
    "retries": 3,
    "enabled": true
  }
}
```

### Formatting Rules

| Rule | Example | Why |
|------|---------|-----|
| 2-space indent | `  "key": "value"` | Readability |
| Double quotes only | `"key"` not `'key'` | JSON spec |
| No trailing commas | `["a", "b"]` | JSON spec |
| Trailing newline | File ends with `\n` | POSIX, git diffs |
| UTF-8 encoding | Always | Compatibility |

### Key Naming Conventions

| Convention | Use For | Example |
|------------|---------|---------|
| `camelCase` | JavaScript/TypeScript | `"apiVersion"` |
| `snake_case` | Python | `"issue_type"` |

---

## JSONL Format

### What is JSONL?

JSON Lines: one valid JSON object per line, newline-delimited.

```jsonl
{"id": "abc-123", "status": "open", "title": "First issue"}
{"id": "abc-124", "status": "closed", "title": "Second issue"}
```

### When to Use

| Use JSONL | Use JSON |
|-----------|----------|
| Append-only data | Single config |
| Streaming ingestion | Nested data |
| Line-by-line processing | Small datasets |
| Large datasets | Human-edited |

### Processing JSONL

```bash
# Count records
wc -l issues.jsonl

# Filter by field
jq -c 'select(.status == "open")' issues.jsonl

# Extract field
jq -r '.title' issues.jsonl

# Convert JSON array to JSONL
jq -c '.[]' array.json > data.jsonl
```

---

## Configuration Files

### tsconfig.json

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "NodeNext",
    "strict": true,
    "outDir": "./dist"
  },
  "include": ["src/**/*"],
  "exclude": ["node_modules"]
}
```

### package.json

```json
{
  "name": "package-name",
  "version": "1.0.0",
  "description": "Brief description",
  "main": "dist/index.js",
  "scripts": {
    "build": "tsc",
    "test": "jest",
    "lint": "eslint ."
  }
}
```

---

## JSON Schema

### Defining Schemas

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://example.com/config.schema.json",
  "title": "Configuration",
  "type": "object",
  "required": ["name", "version"],
  "properties": {
    "name": {
      "type": "string",
      "minLength": 1
    },
    "version": {
      "type": "string",
      "pattern": "^\\d+\\.\\d+\\.\\d+$"
    }
  }
}
```

---

## Tooling

### Formatting

```bash
# jq - Format and validate
jq . config.json > formatted.json

# Prettier - Format with config
npx prettier --write '**/*.json'

# Python - Format
python -m json.tool config.json
```

### Validation

```bash
# jq - Check valid JSON
jq empty config.json && echo "Valid"

# Python - Check valid JSON
python -c "import json; json.load(open('config.json'))"
```

---

## Compliance Assessment

**Use letter grades + evidence, NOT numeric scores.**

### Grading Scale

| Grade | Criteria |
|-------|----------|
| A+ | All files validate, 2-space, UTF-8, schema valid |
| A | Valid JSON, consistent formatting |
| A- | Minor formatting inconsistencies |
| B | Valid but poorly formatted |
| C | Parse errors |

### Validation Commands

```bash
# Validate JSON
find . -name '*.json' -exec jq empty {} \; 2>&1 | grep -c "parse error"

# JSONL: validate line count
wc -l data.jsonl
jq -c '.' data.jsonl | wc -l
```

---

## Additional Resources

- [JSON Spec](https://www.json.org/)
- [JSON Lines](https://jsonlines.org/)
- [JSON Schema](https://json-schema.org/)
- [jq Manual](https://stedolan.github.io/jq/manual/)
