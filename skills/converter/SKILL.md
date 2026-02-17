---
name: converter
description: 'Cross-platform skill converter. Parse AgentOps skills into a universal bundle format, then convert to target platforms (Codex, Cursor). Triggers: convert, converter, convert skill, export skill, cross-platform.'
metadata:
  tier: solo
  dependencies: []
---

# /converter -- Cross-Platform Skill Converter

Parse AgentOps skills into a universal SkillBundle format, then convert to target agent platforms.

## Quick Start

```bash
/converter skills/council codex     # Convert council skill to Codex format
/converter skills/vibe cursor       # Convert vibe skill to Cursor format
/converter --all codex              # Convert all skills to Codex
```

## Pipeline

The converter runs a three-stage pipeline:

```
parse --> convert --> write
```

### Stage 1: Parse

Read the source skill directory and produce a SkillBundle:

- Extract YAML frontmatter from SKILL.md (between `---` markers)
- Collect the markdown body (everything after the closing `---`)
- Enumerate all files in `references/` and `scripts/`
- Assemble into a SkillBundle (see `references/skill-bundle-schema.md`)

### Stage 2: Convert

Transform the SkillBundle into the target platform's format:

| Target | Output Format | Status |
|--------|---------------|--------|
| `codex` | Codex agents.md + instructions | Planned (ag-hm6.5) |
| `cursor` | Cursor rules directory | Planned (ag-hm6.6) |

Until a target adapter is implemented, the converter outputs the raw parsed SkillBundle as structured markdown for inspection.

### Stage 3: Write

Write the converted output to disk.

- **Default output directory:** `.agents/converter/<target>/<skill-name>/`
- **Write semantics:** Clean-write. The target directory is deleted before writing. No merge with existing content.

## CLI Usage

```bash
# Convert a single skill
bash skills/converter/scripts/convert.sh <skill-dir> <target> [output-dir]

# Convert all skills
bash skills/converter/scripts/convert.sh --all <target> [output-dir]
```

### Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `skill-dir` | Yes (or `--all`) | Path to skill directory (e.g. `skills/council`) |
| `target` | Yes | Target platform: `codex`, `cursor`, or `test` |
| `output-dir` | No | Override output location. Default: `.agents/converter/<target>/<skill-name>/` |
| `--all` | No | Convert all skills in `skills/` directory |

## Supported Targets

- **codex** -- Convert to OpenAI Codex agent format (instructions file + metadata). See ag-hm6.5.
- **cursor** -- Convert to Cursor rules format (`.cursor/rules/` directory layout). See ag-hm6.6.
- **test** -- Emit the raw SkillBundle as structured markdown. Useful for debugging the parse stage.

## Extending

To add a new target platform:

1. Add a conversion function to `scripts/convert.sh` (pattern: `convert_<target>`)
2. Update the target table above
3. Add reference docs to `references/` if the target format needs documentation

## References

- `references/skill-bundle-schema.md` -- SkillBundle interchange format specification
