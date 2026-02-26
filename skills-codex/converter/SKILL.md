---
name: converter
description: 'Cross-platform skill converter. Parse AgentOps skills into a universal bundle format, then convert to target platforms (Codex, Cursor). Triggers: convert, converter, convert skill, export skill, cross-platform.'
---


# $converter -- Cross-Platform Skill Converter

Parse AgentOps skills into a universal SkillBundle format, then convert to target agent platforms.

## Quick Start

```bash
$converter skills/council codex     # Convert council skill to Codex format
$converter skills/vibe cursor       # Convert vibe skill to Cursor format
$converter --all codex              # Convert all skills to Codex
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
| `codex` | Codex SKILL.md + prompt.md | Implemented |
| `cursor` | Cursor .mdc rule + optional mcp.json | Implemented |

The Codex adapter produces a `SKILL.md` with YAML frontmatter (`name`, `description`) plus body content, inlined references, and scripts as code blocks. It also emits a `prompt.md` (Codex prompt referencing the skill). Codex output rewrites known slash-skill references (for example `$plan`) to dollar-skill syntax (`$plan`), replaces Claude-specific paths/labels, rewrites Claude-only primitive labels to runtime-neutral wording, and rewrites flat `ao` command references to namespace-qualified forms expected by Codex-native lint. Descriptions are truncated to 1024 chars at a word boundary if needed.

The Cursor adapter produces a `<name>.mdc` rule file with YAML frontmatter (`description`, `globs`, `alwaysApply: false`) and body content. References are inlined into the body, scripts are included as code blocks. Output is budget-fitted to 100KB max -- references are omitted largest-first if the total exceeds the limit. If the skill references MCP servers, a `mcp.json` stub is also generated.

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

- **codex** -- Convert to OpenAI Codex format (frontmatter-preserving `SKILL.md` + `prompt.md`) with codex-native rewrites (slash-to-dollar skills, `~/.codex/` to `~/.codex/`, Claude primitive label neutralization, and namespace-qualified `ao` command references). Output: `<dir>/SKILL.md` and `<dir>/prompt.md`.
- **cursor** -- Convert to Cursor rules format (`.mdc` rule file + optional `mcp.json`). Output: `<dir>/<name>.mdc` and optionally `<dir>/mcp.json`.
- **test** -- Emit the raw SkillBundle as structured markdown. Useful for debugging the parse stage.

## Extending

To add a new target platform:

1. Add a conversion function to `scripts/convert.sh` (pattern: `convert_<target>`)
2. Update the target table above
3. Add reference docs to `references/` if the target format needs documentation

## Examples

### Converting a single skill to Codex format

**User says:** `$converter skills/council codex`

**What happens:**
1. The converter parses `skills/council/SKILL.md` frontmatter, markdown body, and any `references/` and `scripts/` files into a SkillBundle.
2. The Codex adapter transforms the bundle into a `SKILL.md` (body + inlined references + scripts as code blocks) and a `prompt.md` (Codex prompt referencing the skill).
3. Output is written to `.agents/converter/codex/council/`.

**Result:** A Codex-compatible skill package ready to use with OpenAI Codex CLI.

### Batch-converting all skills to Cursor rules

**User says:** `$converter --all cursor`

**What happens:**
1. The converter scans every directory under `skills/` and parses each into a SkillBundle.
2. The Cursor adapter transforms each bundle into a `.mdc` rule file with YAML frontmatter and body content, budget-fitted to 100KB max. Skills referencing MCP servers also get a `mcp.json` stub.
3. Each skill's output is written to `.agents/converter/cursor/<skill-name>/`.

**Result:** All skills are available as Cursor rules, ready to drop into a `.cursor/rules/` directory.

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| `parse error: no frontmatter found` | SKILL.md is missing the `---` delimited YAML frontmatter block | Add frontmatter with at least `name:` and `description:` fields, or run `$heal-skill --fix` on the skill first |
| Cursor `.mdc` output is missing references | Total bundle size exceeded the 100KB budget limit | The converter omits references largest-first to fit the budget. Split large reference files or move non-essential content to external docs |
| Output directory already has old files | Previous conversion artifacts remain | This is expected -- the converter clean-writes by deleting the target directory before writing. If old files persist, manually delete `.agents/converter/<target>/<skill>/` |
| `--all` skips a skill directory | The directory has no `SKILL.md` file | Ensure each skill directory contains a valid `SKILL.md`. Run `$heal-skill` to detect empty directories |
| Codex `prompt.md` description is truncated | The skill description exceeds 1024 characters | This is by design. The converter truncates at a word boundary to fit Codex limits. Shorten the description in SKILL.md frontmatter if the truncation point is awkward |

## References

- `references/skill-bundle-schema.md` -- SkillBundle interchange format specification

## Reference Documents

- [references/skill-bundle-schema.md](references/skill-bundle-schema.md)

---

## References

### skill-bundle-schema.md

# SkillBundle Interchange Format

The SkillBundle is the universal intermediate representation produced by the converter's parse stage. Every target adapter consumes a SkillBundle and transforms it into platform-specific output.

## Schema

```yaml
SkillBundle:
  name: string          # from frontmatter 'name' field
  description: string   # from frontmatter 'description' field
  body: string          # markdown content after frontmatter (closing --- to EOF)
  references:           # files found in references/ directory
    - name: string      # filename (e.g. 'output-format.md')
      content: string   # full file content
  scripts:              # files found in scripts/ directory
    - name: string      # filename (e.g. 'validate.sh')
      content: string   # full file content
  frontmatter: object   # full parsed YAML frontmatter as key-value pairs
```

## Field Details

### name (string, required)

The skill's short name, extracted from the `name` field in SKILL.md YAML frontmatter.

Example: `council`, `vibe`, `crank`

### description (string, required)

The skill's description, extracted from the `description` field in SKILL.md frontmatter. May contain trigger lists and usage summaries.

### body (string, required)

The full markdown content of SKILL.md after the closing `---` of the frontmatter block. This is the skill's instructions, workflow documentation, and inline agent definitions.

### references (array of objects)

Each file in the skill's `references/` directory becomes one entry:

- **name**: The filename without path prefix (e.g. `output-format.md`)
- **content**: The complete file contents as a string

If no `references/` directory exists, this is an empty array.

### scripts (array of objects)

Each file in the skill's `scripts/` directory becomes one entry:

- **name**: The filename without path prefix (e.g. `validate.sh`)
- **content**: The complete file contents as a string

If no `scripts/` directory exists, this is an empty array.

### frontmatter (object)

The complete parsed YAML frontmatter as a flat or nested key-value structure. This includes all fields -- not just `name` and `description` -- so target adapters can access `metadata.tier`, `metadata.dependencies`, and any custom fields.

Example:

```yaml
frontmatter:
  name: council
  description: 'Multi-model consensus council...'
  metadata:
    tier: orchestration
    dependencies:
      - standards
    replaces: judge
```

## Usage in Target Adapters

Target adapters receive the SkillBundle and decide which fields to use:

| Adapter | Fields Used | Notes |
|---------|-------------|-------|
| codex | name, description, body, references | Flattens into single agents.md |
| cursor | name, description, body, references | Splits into .cursor/rules/ files |
| test | all | Dumps full bundle for inspection |

## Serialization

The SkillBundle is an in-memory structure passed between pipeline stages. When written to disk (e.g. by the `test` target), it is rendered as structured markdown with clear section headers for each field.


---

## Scripts

### convert.sh

```bash
#!/usr/bin/env bash
# convert.sh — Cross-platform skill converter pipeline
# Usage: bash skills/converter/scripts/convert.sh <skill-dir> <target> [output-dir]
#        bash skills/converter/scripts/convert.sh --all <target> [output-dir]
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
SKILL_PATTERN=""
NAMESPACE_FROM=()
NAMESPACE_TO=()

# ─── Helpers ───────────────────────────────────────────────────────────

die() { echo "ERROR: $*" >&2; exit 1; }

usage() {
  cat <<'EOF'
Usage:
  bash skills/converter/scripts/convert.sh <skill-dir> <target> [output-dir]
  bash skills/converter/scripts/convert.sh --all <target> [output-dir]

Targets: codex, cursor, test

Examples:
  bash skills/converter/scripts/convert.sh skills/council codex
  bash skills/converter/scripts/convert.sh --all codex
  bash skills/converter/scripts/convert.sh skills/vibe test /tmp/out
EOF
  exit 1
}

yaml_escape_single_quote() {
  printf '%s' "$1" | sed "s/'/''/g"
}

# Build an alternation regex for all known skill names.
load_skill_pattern() {
  local names=()
  local d
  for d in "$REPO_ROOT"/skills/*/; do
    [[ -f "$d/SKILL.md" ]] || continue
    names+=("$(basename "$d")")
  done

  if [[ ${#names[@]} -eq 0 ]]; then
    SKILL_PATTERN=""
    return
  fi

  local escaped=()
  local name
  for name in "${names[@]}"; do
    escaped+=("$(printf '%s' "$name" | sed -E 's/[][(){}.^$*+?|\\-]/\\&/g')")
  done
  SKILL_PATTERN="$(IFS='|'; printf '%s' "${escaped[*]}")"
}

# Rewrite Claude-style slash command references to Codex-style dollar references.
# Example: $plan -> $plan (for known skill names only).
codex_rewrite_text() {
  local input="$1"
  local output="$input"

  if [[ -n "$SKILL_PATTERN" ]]; then
    output="$(printf '%s' "$output" | SKILL_PATTERN="$SKILL_PATTERN" perl -0pe '
      my $pattern = qr/$ENV{SKILL_PATTERN}/;
      s{(?<![A-Za-z0-9_/])/($pattern)(?![A-Za-z0-9-])}{\$$1}g;
    ')"
  fi

  output="$(printf '%s' "$output" | perl -0pe '
    s/\bClaude Code\b/Codex/g;
    s{~/.codex/}{~/.codex/}g;
    s/\bTeamCreate\b/team-create/g;
    s/\bSendMessage\b/send-message/g;
    s/\bEnterPlanMode\b/enter-plan-mode/g;
    s/\bExitPlanMode\b/exit-plan-mode/g;
    s/\bEnterWorktree\b/enter-worktree/g;
  ')"

  # Rewrite flat ao command references to namespace-qualified forms expected by
  # codex-native lint rules (inverse of cli/cmd/ao/doctor.go deprecated map).
  if [[ ${#NAMESPACE_FROM[@]} -gt 0 ]]; then
    local i from_esc to_esc
    for i in "${!NAMESPACE_FROM[@]}"; do
      from_esc="$(printf '%s' "${NAMESPACE_FROM[$i]}" | sed -e 's/[\/&|]/\\&/g')"
      to_esc="$(printf '%s' "${NAMESPACE_TO[$i]}" | sed -e 's/[\/&|]/\\&/g')"
      output="$(printf '%s' "$output" | sed "s|${from_esc}|${to_esc}|g")"
    done
  fi
  printf '%s' "$output"
}

# Build replacements from doctor.go deprecated map.
# doctor.go stores old(namespace-qualified) -> new(flat) references.
# For codex-native skills we intentionally apply the inverse (new -> old).
load_namespace_rewrite_pairs() {
  local doctor_go="$REPO_ROOT/cli/cmd/ao/doctor.go"
  NAMESPACE_FROM=()
  NAMESPACE_TO=()

  [[ -f "$doctor_go" ]] || return

  local pairs
  pairs="$(
    sed -n '/var deprecatedCommands/,/^}/p' "$doctor_go" \
      | grep '"ao ' \
      | sed 's/.*"\(ao [^"]*\)".*:.*"\(ao [^"]*\)".*/\1|\2/' \
      | awk -F'|' '{print length($2), $1 "|" $2}' \
      | sort -rn \
      | cut -d' ' -f2-
  )"

  local old_cmd new_cmd
  while IFS='|' read -r old_cmd new_cmd; do
    [[ -n "$old_cmd" && -n "$new_cmd" ]] || continue
    NAMESPACE_FROM+=("$new_cmd")
    NAMESPACE_TO+=("$old_cmd")
  done <<< "$pairs"
}

# ─── Stage 1: Parse ───────────────────────────────────────────────────

# Parse SKILL.md frontmatter and body.
# Sets: BUNDLE_NAME, BUNDLE_DESC, BUNDLE_BODY, BUNDLE_FRONTMATTER
parse_skill_md() {
  local skill_md="$1"
  [[ -f "$skill_md" ]] || die "SKILL.md not found: $skill_md"

  local content
  content="$(<"$skill_md")"

  # Extract frontmatter (between first and second --- lines)
  local in_fm=0
  local fm_lines=()
  local body_lines=()
  local fm_ended=0
  local line_num=0

  while IFS= read -r line; do
    line_num=$((line_num + 1))
    if [[ $line_num -eq 1 && "$line" == "---" ]]; then
      in_fm=1
      continue
    fi
    if [[ $in_fm -eq 1 && "$line" == "---" ]]; then
      in_fm=0
      fm_ended=1
      continue
    fi
    if [[ $in_fm -eq 1 ]]; then
      fm_lines+=("$line")
    elif [[ $fm_ended -eq 1 ]]; then
      body_lines+=("$line")
    fi
  done <<< "$content"

  BUNDLE_FRONTMATTER="$(printf '%s\n' "${fm_lines[@]}")"

  # Extract name and description from frontmatter
  BUNDLE_NAME="$(echo "$BUNDLE_FRONTMATTER" | sed -n 's/^name: *//p' | tr -d "'" | tr -d '"')"
  BUNDLE_DESC="$(echo "$BUNDLE_FRONTMATTER" | sed -n 's/^description: *//p' | sed "s/^'//;s/'$//")"

  # Body: join with newlines
  BUNDLE_BODY="$(printf '%s\n' "${body_lines[@]}")"
}

# Collect files from a subdirectory into parallel arrays.
# Args: <dir> <array-name-names> <array-name-contents>
collect_files() {
  local dir="$1"
  local -n names_arr="$2"
  local -n contents_arr="$3"
  names_arr=()
  contents_arr=()

  if [[ -d "$dir" ]]; then
    local f
    for f in "$dir"/*; do
      [[ -f "$f" ]] || continue
      names_arr+=("$(basename "$f")")
      contents_arr+=("$(<"$f")")
    done
  fi
}

# Full parse: populate all BUNDLE_* variables and REF/SCRIPT arrays
parse_bundle() {
  local skill_dir="$1"
  parse_skill_md "$skill_dir/SKILL.md"
  collect_files "$skill_dir/references" REF_NAMES REF_CONTENTS
  collect_files "$skill_dir/scripts" SCRIPT_NAMES SCRIPT_CONTENTS
}

# ─── Stage 2: Convert ─────────────────────────────────────────────────

# Test target: emit SkillBundle as structured markdown
convert_test() {
  local out=""
  out+="# SkillBundle: ${BUNDLE_NAME}"$'\n\n'
  out+="## Name"$'\n\n'
  out+="${BUNDLE_NAME}"$'\n\n'
  out+="## Description"$'\n\n'
  out+="${BUNDLE_DESC}"$'\n\n'
  out+="## Frontmatter"$'\n\n'
  out+='```yaml'$'\n'
  out+="${BUNDLE_FRONTMATTER}"$'\n'
  out+='```'$'\n\n'
  out+="## Body"$'\n\n'
  out+="${BUNDLE_BODY}"$'\n\n'

  out+="## References (${#REF_NAMES[@]})"$'\n\n'
  local i
  for i in "${!REF_NAMES[@]}"; do
    out+="### ${REF_NAMES[$i]}"$'\n\n'
    out+='```'$'\n'
    out+="${REF_CONTENTS[$i]}"$'\n'
    out+='```'$'\n\n'
  done

  out+="## Scripts (${#SCRIPT_NAMES[@]})"$'\n\n'
  for i in "${!SCRIPT_NAMES[@]}"; do
    out+="### ${SCRIPT_NAMES[$i]}"$'\n\n'
    out+='```'$'\n'
    out+="${SCRIPT_CONTENTS[$i]}"$'\n'
    out+='```'$'\n\n'
  done

  CONVERTED_OUTPUT="$out"
  CONVERTED_FILENAME="bundle.md"
}

# Codex target: SKILL.md + prompt.md
# Codex skills live at ~/.codex/skills/<name>/SKILL.md
# Codex prompts live at ~/.codex/prompts/<name>.md
# Description max 1024 chars, no hooks support, tool names pass through
convert_codex() {
  local desc="$BUNDLE_DESC"
  local body
  body="$(codex_rewrite_text "$BUNDLE_BODY")"

  # Truncate description to 1024 chars at word boundary
  if [[ ${#desc} -gt 1024 ]]; then
    desc="${desc:0:1021}"
    # Trim to last word boundary (space)
    desc="${desc% *}..."
  fi
  desc="$(codex_rewrite_text "$desc")"
  local desc_escaped
  desc_escaped="$(yaml_escape_single_quote "$desc")"

  # ── Build SKILL.md ──
  local skill_md=""
  skill_md+="---"$'\n'
  skill_md+="name: ${BUNDLE_NAME}"$'\n'
  skill_md+="description: '${desc_escaped}'"$'\n'
  skill_md+="---"$'\n\n'
  skill_md+="${body}"$'\n'

  # Inline references as appended sections
  if [[ ${#REF_NAMES[@]} -gt 0 ]]; then
    skill_md+=$'\n'"---"$'\n\n'
    skill_md+="## References"$'\n\n'
    local i
    for i in "${!REF_NAMES[@]}"; do
      skill_md+="### ${REF_NAMES[$i]}"$'\n\n'
      skill_md+="$(codex_rewrite_text "${REF_CONTENTS[$i]}")"$'\n\n'
    done
  fi

  # Inline scripts as code blocks
  if [[ ${#SCRIPT_NAMES[@]} -gt 0 ]]; then
    skill_md+=$'\n'"---"$'\n\n'
    skill_md+="## Scripts"$'\n\n'
    local i
    for i in "${!SCRIPT_NAMES[@]}"; do
      # Detect language from extension
      local ext="${SCRIPT_NAMES[$i]##*.}"
      local lang=""
      case "$ext" in
        sh|bash) lang="bash" ;;
        py)      lang="python" ;;
        js)      lang="javascript" ;;
        ts)      lang="typescript" ;;
        *)       lang="$ext" ;;
      esac
      skill_md+="### ${SCRIPT_NAMES[$i]}"$'\n\n'
      skill_md+="\`\`\`${lang}"$'\n'
      skill_md+="$(codex_rewrite_text "${SCRIPT_CONTENTS[$i]}")"$'\n'
      skill_md+="\`\`\`"$'\n\n'
    done
  fi

  # ── Build prompt.md ──
  local prompt_md=""
  prompt_md+="# ${BUNDLE_NAME}"$'\n\n'
  prompt_md+="${desc}"$'\n\n'
  prompt_md+="## Instructions"$'\n\n'
  prompt_md+="Load and follow the skill instructions from \`~/.codex/skills/${BUNDLE_NAME}/SKILL.md\`."$'\n'

  # Set primary output (SKILL.md)
  CONVERTED_OUTPUT="$skill_md"
  CONVERTED_FILENAME="SKILL.md"

  # Set secondary output (prompt.md)
  CONVERTED_OUTPUT_2="$prompt_md"
  CONVERTED_FILENAME_2="prompt.md"
}

# Cursor target: .mdc rule file with YAML frontmatter + optional mcp.json
# Cursor rules format: .cursor/rules/<name>.mdc (Cursor 0.40+)
# Max output size: 100KB (102400 bytes). References are budget-fitted.
CURSOR_MAX_BYTES=102400

convert_cursor() {
  local out=""

  # ── YAML frontmatter ──
  out+="---"$'\n'
  out+="description: ${BUNDLE_DESC}"$'\n'
  out+="globs: "$'\n'
  out+="alwaysApply: false"$'\n'
  out+="---"$'\n\n'

  # ── Body content ──
  out+="${BUNDLE_BODY}"$'\n'

  # ── Scripts as code blocks (included before references — smaller, higher value) ──
  if [[ ${#SCRIPT_NAMES[@]} -gt 0 ]]; then
    out+=$'\n'"## Scripts"$'\n\n'
    local i
    for i in "${!SCRIPT_NAMES[@]}"; do
      local ext="${SCRIPT_NAMES[$i]##*.}"
      local lang=""
      case "$ext" in
        sh|bash) lang="bash" ;;
        py)      lang="python" ;;
        js)      lang="javascript" ;;
        ts)      lang="typescript" ;;
        *)       lang="$ext" ;;
      esac
      out+="### ${SCRIPT_NAMES[$i]}"$'\n\n'
      out+="\`\`\`${lang}"$'\n'
      out+="${SCRIPT_CONTENTS[$i]}"$'\n'
      out+="\`\`\`"$'\n\n'
    done
  fi

  # ── Inline references (budget-fitted to stay under CURSOR_MAX_BYTES) ──
  if [[ ${#REF_NAMES[@]} -gt 0 ]]; then
    local current_size=${#out}
    local budget=$(( CURSOR_MAX_BYTES - current_size - 200 ))  # 200 byte margin for section header + omission note
    local ref_section=""
    local omitted=0
    local i

    ref_section+=$'\n'"## References"$'\n\n'
    for i in "${!REF_NAMES[@]}"; do
      local entry=""
      entry+="### ${REF_NAMES[$i]}"$'\n\n'
      entry+="${REF_CONTENTS[$i]}"$'\n\n'
      local entry_size=${#entry}

      if [[ $budget -ge $entry_size ]]; then
        ref_section+="$entry"
        budget=$(( budget - entry_size ))
      else
        omitted=$(( omitted + 1 ))
      fi
    done

    if [[ $omitted -gt 0 ]]; then
      ref_section+="*${omitted} reference(s) omitted to stay under 100KB size limit.*"$'\n\n'
      echo "WARN: ${BUNDLE_NAME}: omitted $omitted reference(s) to stay under 100KB" >&2
    fi

    out+="$ref_section"
  fi

  CONVERTED_OUTPUT="$out"
  CONVERTED_FILENAME="${BUNDLE_NAME}.mdc"

  # ── MCP detection: scan body + references for MCP server references ──
  # If skill content references MCP servers, generate a stub mcp.json
  local all_content="${BUNDLE_BODY}"
  local i
  for i in "${!REF_CONTENTS[@]}"; do
    all_content+=$'\n'"${REF_CONTENTS[$i]}"
  done

  if echo "$all_content" | grep -qiE '(mcpServers|mcp_server|"mcp"|mcp\.json)'; then
    CONVERTED_OUTPUT_2='{
  "mcpServers": {}
}'
    CONVERTED_FILENAME_2="mcp.json"
  fi
}

run_convert() {
  local target="$1"
  case "$target" in
    test)   convert_test ;;
    codex)  convert_codex ;;
    cursor) convert_cursor ;;
    *)      die "Unknown target: $target. Supported: codex, cursor, test" ;;
  esac
}

# ─── Stage 3: Write ───────────────────────────────────────────────────

write_output() {
  local output_dir="$1"

  # Clean-write: delete target dir before writing
  if [[ -d "$output_dir" ]]; then
    rm -rf "$output_dir"
  fi
  mkdir -p "$output_dir"

  echo "$CONVERTED_OUTPUT" > "$output_dir/$CONVERTED_FILENAME"
  echo "OK: $output_dir/$CONVERTED_FILENAME"

  # Write secondary output if present (e.g., codex prompt.md)
  if [[ -n "${CONVERTED_OUTPUT_2:-}" && -n "${CONVERTED_FILENAME_2:-}" ]]; then
    echo "$CONVERTED_OUTPUT_2" > "$output_dir/$CONVERTED_FILENAME_2"
    echo "OK: $output_dir/$CONVERTED_FILENAME_2"
  fi
}

# ─── Main ─────────────────────────────────────────────────────────────

convert_one_skill() {
  local skill_dir="$1"
  local target="$2"
  local output_dir="$3"

  # Resolve skill_dir to absolute if relative
  if [[ "$skill_dir" != /* ]]; then
    skill_dir="$REPO_ROOT/$skill_dir"
  fi

  [[ -d "$skill_dir" ]] || die "Skill directory not found: $skill_dir"
  [[ -f "$skill_dir/SKILL.md" ]] || die "No SKILL.md in: $skill_dir"

  parse_bundle "$skill_dir"

  [[ -n "$BUNDLE_NAME" ]] || die "Failed to parse name from $skill_dir/SKILL.md"

  # Default output dir
  if [[ -z "$output_dir" ]]; then
    output_dir="$REPO_ROOT/.agents/converter/$target/$BUNDLE_NAME"
  elif [[ "$output_dir" != /* ]]; then
    output_dir="$REPO_ROOT/$output_dir"
  fi

  # Reset output variables
  CONVERTED_OUTPUT=""
  CONVERTED_FILENAME=""
  CONVERTED_OUTPUT_2=""
  CONVERTED_FILENAME_2=""

  run_convert "$target"
  write_output "$output_dir"
}

main() {
  [[ $# -ge 2 ]] || usage

  local skill_dir_or_flag="$1"
  local target="$2"
  local output_dir="${3:-}"

  load_skill_pattern
  load_namespace_rewrite_pairs

  if [[ "$skill_dir_or_flag" == "--all" ]]; then
    local skills_root="$REPO_ROOT/skills"
    local count=0
    for d in "$skills_root"/*/; do
      [[ -f "$d/SKILL.md" ]] || continue
      local sname
      sname="$(basename "$d")"
      local out="$output_dir"
      if [[ -n "$out" ]]; then
        # Per-skill subdir under the provided output dir
        if [[ "$out" != /* ]]; then
          out="$REPO_ROOT/$out/$sname"
        else
          out="$out/$sname"
        fi
      fi
      convert_one_skill "$d" "$target" "$out"
      count=$((count + 1))
    done
    echo "Converted $count skills to target '$target'"
  else
    convert_one_skill "$skill_dir_or_flag" "$target" "$output_dir"
  fi
}

main "$@"
```

### validate.sh

```bash
#!/usr/bin/env bash
set -euo pipefail
SKILL_DIR="$(cd "$(dirname "$0")/.." && pwd)"
PASS=0; FAIL=0

check() { if bash -c "$2"; then echo "PASS: $1"; PASS=$((PASS + 1)); else echo "FAIL: $1"; FAIL=$((FAIL + 1)); fi; }

check "SKILL.md exists" "[ -f '$SKILL_DIR/SKILL.md' ]"
check "SKILL.md has YAML frontmatter" "head -1 '$SKILL_DIR/SKILL.md' | grep -q '^---$'"

echo ""; echo "Results: $PASS passed, $FAIL failed"
[ $FAIL -eq 0 ] && exit 0 || exit 1
```


