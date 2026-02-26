---
name: heal-skill
description: 'Automated skill maintenance. Detects and fixes common skill issues: missing frontmatter, name mismatches, unlinked references, empty directories, dead references. Triggers: "heal-skill", "heal skill", "fix skills", "skill maintenance", "repair skills".'
---


# $heal-skill — Automated Skill Maintenance

> **Purpose:** Detect and auto-fix common skill hygiene issues across the skills/ directory.

**YOU MUST EXECUTE THIS WORKFLOW. Do not just describe it.**

---

## Quick Start

```bash
$heal-skill                    # Check all skills (report only)
$heal-skill --fix              # Auto-repair all fixable issues
$heal-skill --strict           # Check all skills, exit 1 on findings (CI mode)
$heal-skill skills/council     # Check a specific skill
$heal-skill --fix skills/vibe  # Fix a specific skill
```

---

## What It Detects

Ten checks, run in order:

| Code | Issue | Auto-fixable? |
|------|-------|---------------|
| `MISSING_NAME` | No `name:` field in SKILL.md frontmatter | Yes -- adds name from directory |
| `MISSING_DESC` | No `description:` field in SKILL.md frontmatter | Yes -- adds placeholder |
| `NAME_MISMATCH` | Frontmatter `name` differs from directory name | Yes -- updates to match directory |
| `UNLINKED_REF` | File in references/ not linked in SKILL.md | Yes -- converts bare backtick refs to markdown links |
| `EMPTY_DIR` | Skill directory exists but has no SKILL.md | Yes -- removes empty directory |
| `DEAD_REF` | SKILL.md references a non-existent references/ file | No -- warn only |
| `SCRIPT_REF_MISSING` | SKILL.md references a scripts/ file that does not exist | No -- warn only |
| `INVALID_AO_CMD` | SKILL.md references an `ao` subcommand that does not exist (only runs if `ao` is on PATH) | No -- warn only |
| `DEAD_XREF` | SKILL.md references a `/skill-name` that has no matching skill directory | No -- warn only |
| `CATALOG_MISSING` | A user-invocable skill is missing from the using-agentops catalog | No -- warn only |

---

## Execution Steps

### Step 1: Run the heal script

```bash
# Check mode (default) -- report only, no changes
bash skills/heal-skill/scripts/heal.sh --check

# Fix mode -- auto-repair what it can
bash skills/heal-skill/scripts/heal.sh --fix

# Target a specific skill
bash skills/heal-skill/scripts/heal.sh --check skills/council
bash skills/heal-skill/scripts/heal.sh --fix skills/council
```

### Step 2: Interpret results

- **Exit 0:** All clean, no findings. Also exit 0 for `--check` mode with findings (report-only).
- **Exit 1:** Findings reported with `--strict` or `--fix` flag. In `--fix` mode, fixable issues were repaired; re-run `--check` to confirm.

### Step 3: Report to user

Show the output. If `--fix` was used, summarize what changed. If `DEAD_REF` findings remain, advise the user to remove or update the broken references manually.

---

## Output Format

One line per finding:

```
[MISSING_NAME] skills/foo: No name field in frontmatter
[MISSING_DESC] skills/foo: No description field in frontmatter
[NAME_MISMATCH] skills/foo: Frontmatter name 'bar' != directory 'foo'
[UNLINKED_REF] skills/foo: refs/bar.md not linked in SKILL.md
[EMPTY_DIR] skills/foo: Directory exists but no SKILL.md
[DEAD_REF] skills/foo: SKILL.md links to non-existent refs/bar.md
[SCRIPT_REF_MISSING] skills/foo: references scripts/bar.sh but file not found
[INVALID_AO_CMD] skills/foo: references 'ao badcmd' which is not a valid subcommand
[DEAD_XREF] skills/foo: references /nonexistent but skill directory not found
[CATALOG_MISSING] using-agentops: bar is user-invocable but missing from catalog
```

---

## Notes

- The script is **idempotent** -- running `--fix` twice produces the same result.
- `DEAD_REF`, `SCRIPT_REF_MISSING`, `INVALID_AO_CMD`, `DEAD_XREF`, and `CATALOG_MISSING` are warn-only because the correct resolution requires human judgment.
- `INVALID_AO_CMD` only runs if the `ao` CLI is available on PATH. Skipped silently otherwise.
- `CATALOG_MISSING` is a global check (not per-skill) and only runs when `using-agentops/SKILL.md` exists.
- When run without a path argument, scans all directories under `skills/`.
- Use `--strict` for CI gates: exits 1 on any finding. Without `--strict`, check mode exits 0 even with findings.

## Examples

### Running a health check across all skills

**User says:** `$heal-skill`

**What happens:**
1. The heal script scans every directory under `skills/`, checking each for the ten issue types (missing name, missing description, name mismatch, unlinked references, empty directories, dead references, script reference integrity, CLI command validation, cross-reference validation, catalog completeness).
2. Findings are printed one per line with issue codes (e.g., `[NAME_MISMATCH] skills/foo: Frontmatter name 'bar' != directory 'foo'`).
3. The script exits with code 0 in check mode (even with findings), or code 1 with `--strict` or `--fix` flags.

**Result:** A diagnostic report showing all skill hygiene issues across the repository, with no files modified.

### Auto-fixing a specific skill

**User says:** `$heal-skill --fix skills/vibe`

**What happens:**
1. The heal script inspects only `skills/vibe/`, running all per-skill checks against that skill.
2. For each fixable issue found (e.g., `MISSING_NAME`, `UNLINKED_REF`), the script applies the repair automatically -- adding the name from the directory, converting bare backtick references to markdown links, etc.
3. Any `DEAD_REF` findings are reported as warnings since they require human judgment to resolve.

**Result:** The `skills/vibe/SKILL.md` is repaired in place, with a summary of changes applied and any remaining warnings.

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| `DEAD_REF` findings persist after `--fix` | Dead references are warn-only because the correct fix (delete, create, or update) requires human judgment | Manually inspect each dead reference and either create the missing file, remove the link from SKILL.md, or update the path |
| Script reports `EMPTY_DIR` for a skill in progress | The skill directory was created but SKILL.md has not been written yet | Either add a SKILL.md to the directory or remove the empty directory. Running `--fix` will remove it automatically |
| `NAME_MISMATCH` fix changed the wrong name | The script always updates the frontmatter `name` to match the directory name, not the other way around | If the directory name is wrong, rename the directory first, then re-run `--fix` |
| Script exits 0 but a skill still has issues | The issue type is not one of the ten checks the heal script detects | The heal script covers structural hygiene only. Content quality issues require manual review or `$council` validation |
| Running `--fix` twice produces different output | This should not happen -- the script is idempotent | File a bug. Check if another process modified the skill files between runs |

---

## Scripts

### heal.sh

```bash
#!/usr/bin/env bash
# heal.sh — Detect and fix common skill hygiene issues.
# Usage: heal.sh [--check|--fix] [--strict] [skills/path ...]
# Exit 0 = clean (or findings in non-strict mode).
# Exit 1 = findings reported in --strict mode (or --fix with findings).

set -euo pipefail

MODE="check"
STRICT=0
TARGETS=()

# Parse args
while [[ $# -gt 0 ]]; do
  case "$1" in
    --check)  MODE="check";  shift ;;
    --fix)    MODE="fix";    shift ;;
    --strict) STRICT=1;      shift ;;
    *)        TARGETS+=("$1"); shift ;;
  esac
done

# Find repo root (location of skills/ directory)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
SKILLS_ROOT="$REPO_ROOT/skills"

# If no targets, scan all skill dirs
if [[ ${#TARGETS[@]} -eq 0 ]]; then
  for d in "$REPO_ROOT"/skills/*/; do
    [[ -d "$d" ]] && TARGETS+=("${d%/}")
  done
else
  # Normalize targets to absolute paths
  normalized=()
  for t in "${TARGETS[@]}"; do
    if [[ "$t" = /* ]]; then
      normalized+=("$t")
    else
      normalized+=("$REPO_ROOT/$t")
    fi
  done
  TARGETS=("${normalized[@]}")
fi

FINDINGS=0

report() {
  local code="$1" path="$2" msg="$3"
  # Show relative path from repo root
  local rel="${path#"$REPO_ROOT"/}"
  echo "[$code] $rel: $msg"
  FINDINGS=$((FINDINGS + 1))
}

# Extract YAML frontmatter value. Handles quoted and unquoted values.
get_frontmatter() {
  local file="$1" key="$2"
  # Read between first --- pair
  local in_fm=0 value=""
  while IFS= read -r line; do
    if [[ "$line" == "---" ]]; then
      if [[ $in_fm -eq 1 ]]; then break; fi
      in_fm=1
      continue
    fi
    if [[ $in_fm -eq 1 ]]; then
      # Match key at start of line (not indented = top-level)
      if [[ "$line" =~ ^${key}:\ *(.*) ]]; then
        value="${BASH_REMATCH[1]}"
        # Strip surrounding quotes
        value="${value#\"}"
        value="${value%\"}"
        value="${value#\'}"
        value="${value%\'}"
        echo "$value"
        return 0
      fi
    fi
  done < "$file"
  return 1
}

# Check if a references/ file is linked in SKILL.md (as a proper markdown link or Read instruction)
is_linked() {
  local skill_md="$1" ref_file="$2"
  # Check for markdown link pattern [text](references/file) or Read tool pattern referencing it
  # Also accept any non-backtick reference to the file path
  local ref_basename
  ref_basename="$(basename "$ref_file")"
  local ref_rel="references/$ref_basename"
  # Linked = appears in a markdown link or Read instruction (not just bare backtick)
  if grep -qE "\]\(.*${ref_rel}.*\)" "$skill_md" 2>/dev/null; then
    return 0
  fi
  if grep -qE "Read.*${ref_rel}" "$skill_md" 2>/dev/null; then
    return 0
  fi
  # Also accept if referenced via a relative path in some other link form
  if grep -qE "\(${ref_rel}\)" "$skill_md" 2>/dev/null; then
    return 0
  fi
  return 1
}

# Fix: add missing name field to frontmatter
fix_missing_name() {
  local file="$1" dirname="$2"
  # Insert name: after first ---
  local tmp
  tmp="$(mktemp)"
  local first_fence=0
  while IFS= read -r line; do
    echo "$line" >> "$tmp"
    if [[ "$line" == "---" && $first_fence -eq 0 ]]; then
      first_fence=1
      echo "name: $dirname" >> "$tmp"
    fi
  done < "$file"
  /bin/cp "$tmp" "$file"
  rm -f "$tmp"
}

# Fix: add missing description field to frontmatter
fix_missing_desc() {
  local file="$1" dirname="$2"
  # Insert description after name line, or after first ---
  local tmp
  tmp="$(mktemp)"
  local inserted=0 first_fence=0
  while IFS= read -r line; do
    echo "$line" >> "$tmp"
    if [[ $inserted -eq 0 ]]; then
      if [[ "$line" =~ ^name: ]]; then
        echo "description: '$dirname skill'" >> "$tmp"
        inserted=1
      elif [[ "$line" == "---" && $first_fence -eq 0 ]]; then
        first_fence=1
      elif [[ "$line" == "---" && $first_fence -eq 1 && $inserted -eq 0 ]]; then
        # Closing fence without finding name — shouldn't happen but handle it
        :
      fi
    fi
  done < "$file"
  if [[ $inserted -eq 0 ]]; then
    # Fallback: insert after first ---
    tmp2="$(mktemp)"
    first_fence=0
    while IFS= read -r line; do
      echo "$line" >> "$tmp2"
      if [[ "$line" == "---" && $first_fence -eq 0 ]]; then
        first_fence=1
        echo "description: '$dirname skill'" >> "$tmp2"
      fi
    done < "$file"
    /bin/cp "$tmp2" "$file"
    rm -f "$tmp2"
  else
    /bin/cp "$tmp" "$file"
  fi
  rm -f "$tmp"
}

# Fix: correct name mismatch
fix_name_mismatch() {
  local file="$1" dirname="$2"
  local tmp
  tmp="$(mktemp)"
  local in_fm=0
  while IFS= read -r line; do
    if [[ "$line" == "---" ]]; then
      in_fm=$((1 - in_fm))
      echo "$line" >> "$tmp"
      continue
    fi
    if [[ $in_fm -eq 1 && "$line" =~ ^name:\ * ]]; then
      echo "name: $dirname" >> "$tmp"
    else
      echo "$line" >> "$tmp"
    fi
  done < "$file"
  /bin/cp "$tmp" "$file"
  rm -f "$tmp"
}

# Fix: convert bare backtick ref to markdown link
fix_unlinked_ref() {
  local file="$1" ref_rel="$2"
  local ref_basename
  ref_basename="$(basename "$ref_rel")"
  # Replace bare `references/foo.md` with [references/foo.md](references/foo.md)
  local tmp
  tmp="$(mktemp)"
  sed "s|\`${ref_rel}\`|[${ref_rel}](${ref_rel})|g" "$file" > "$tmp"
  /bin/cp "$tmp" "$file"
  rm -f "$tmp"
}

# Process each skill directory
for skill_dir in "${TARGETS[@]}"; do
  dirname="$(basename "$skill_dir")"
  skill_md="$skill_dir/SKILL.md"

  # Check 5: Empty directory (no SKILL.md)
  if [[ ! -f "$skill_md" ]]; then
    # Only report if directory is truly empty (no files at all) or just missing SKILL.md
    if [[ -z "$(ls -A "$skill_dir" 2>/dev/null)" ]]; then
      report "EMPTY_DIR" "$skill_dir" "Directory exists but no SKILL.md"
      if [[ "$MODE" == "fix" ]]; then
        rmdir "$skill_dir" 2>/dev/null || true
      fi
    fi
    continue
  fi

  # Check 1: Missing name
  if ! name="$(get_frontmatter "$skill_md" "name")"; then
    report "MISSING_NAME" "$skill_dir" "No name field in frontmatter"
    if [[ "$MODE" == "fix" ]]; then
      fix_missing_name "$skill_md" "$dirname"
    fi
    name=""
  fi

  # Check 2: Missing description
  if ! get_frontmatter "$skill_md" "description" >/dev/null 2>&1; then
    report "MISSING_DESC" "$skill_dir" "No description field in frontmatter"
    if [[ "$MODE" == "fix" ]]; then
      fix_missing_desc "$skill_md" "$dirname"
    fi
  fi

  # Check 3: Name mismatch
  if [[ -n "$name" && "$name" != "$dirname" ]]; then
    report "NAME_MISMATCH" "$skill_dir" "Frontmatter name '$name' != directory '$dirname'"
    if [[ "$MODE" == "fix" ]]; then
      fix_name_mismatch "$skill_md" "$dirname"
    fi
  fi

  # Check 4: Unlinked references
  if [[ -d "$skill_dir/references" ]]; then
    for ref_file in "$skill_dir"/references/*.md; do
      [[ -f "$ref_file" ]] || continue
      ref_basename="$(basename "$ref_file")"
      ref_rel="references/$ref_basename"
      if ! is_linked "$skill_md" "$ref_file"; then
        report "UNLINKED_REF" "$skill_dir" "$ref_rel not linked in SKILL.md"
        if [[ "$MODE" == "fix" ]]; then
          fix_unlinked_ref "$skill_md" "$ref_rel"
        fi
      fi
    done
  fi

  # Check 6: Dead references (SKILL.md mentions references/ files that don't exist)
  # Strip fenced code blocks before scanning to avoid false positives from examples
  while IFS= read -r ref_path; do
    [[ -z "$ref_path" ]] && continue
    if [[ ! -f "$skill_dir/$ref_path" ]]; then
      report "DEAD_REF" "$skill_dir" "SKILL.md references non-existent $ref_path"
      if [[ "$MODE" == "fix" ]]; then
        echo "  [WARN] Cannot auto-fix DEAD_REF -- manually remove or create $ref_path"
      fi
    fi
  done < <(awk 'BEGIN{skip=0} /^```/{skip=1-skip; next} skip==0{print}' "$skill_md" | grep -oE 'references/[A-Za-z0-9_.-]+\.md' 2>/dev/null | sort -u || true)

  # Check 7: Script reference integrity
  # Strip fenced code blocks before scanning to avoid false positives from examples
  while IFS= read -r ref; do
    [[ -z "$ref" ]] && continue
    if [[ ! -f "$skill_dir/$ref" ]]; then
      report "SCRIPT_REF_MISSING" "$skill_dir" "references $ref but file not found"
    fi
  done < <(awk 'BEGIN{skip=0} /^```/{skip=1-skip; next} skip==0{print}' "$skill_md" | grep -oE '\bscripts/[a-zA-Z0-9_-]+\.[a-z]+' 2>/dev/null | sort -u || true)

  # Check 8: CLI command validation (only if ao is on PATH)
  if command -v ao >/dev/null 2>&1; then
    ao_cmds="$(ao help 2>&1 | grep -oE '^\s+[a-z][-a-z]*' | tr -d ' ' | sort -u || true)"
    while IFS= read -r subcmd; do
      [[ -z "$subcmd" ]] && continue
      if ! echo "$ao_cmds" | grep -qx "$subcmd"; then
        report "INVALID_AO_CMD" "$skill_dir" "references 'ao $subcmd' which is not a valid subcommand"
      fi
    done < <(grep -oE '`ao [a-z][-a-z]*`' "$skill_md" 2>/dev/null | sed 's/`//g; s/^ao //' | sort -u || true)
  fi

  # Check 9: Cross-reference validation (skill invocation references)
  # Strip fenced code blocks before scanning to avoid false positives from examples
  while IFS= read -r ref; do
    [[ -z "$ref" ]] && continue
    # Skip common filesystem path false positives
    case "$ref" in
      dev|tmp|usr|bin|etc|opt|var|home|proc|sys|path|null|dev/null|skill-name) continue ;;
    esac
    if [[ ! -d "$SKILLS_ROOT/$ref" ]]; then
      report "DEAD_XREF" "$skill_dir" "references /$ref but skill directory not found"
    fi
  done < <(awk 'BEGIN{skip=0} /^```/{skip=1-skip; next} skip==0{print}' "$skill_md" | grep -oE '`/[a-z][-a-z]*`' 2>/dev/null | sed 's/`//g; s|^/||' | sort -u || true)

done

# Check 10: Catalog completeness (global, not per-skill)
if [[ -f "$SKILLS_ROOT/using-agentops/SKILL.md" ]]; then
  for skill_check in "$SKILLS_ROOT"/*/SKILL.md; do
    [[ -f "$skill_check" ]] || continue
    check_dir="$(dirname "$skill_check")"
    check_name="$(basename "$check_dir")"
    # Skip internal/non-invocable skills
    if grep -q 'user-invocable: false' "$skill_check" 2>/dev/null; then continue; fi
    if grep -q 'internal: true' "$skill_check" 2>/dev/null; then continue; fi
    # Check if skill appears in catalog
    if ! grep -q "$check_name" "$SKILLS_ROOT/using-agentops/SKILL.md" 2>/dev/null; then
      report "CATALOG_MISSING" "$SKILLS_ROOT/using-agentops" "$check_name is user-invocable but missing from catalog"
    fi
  done
fi

# Check 11: skill_api_version presence (global, not per-skill)
for skill_check in "$SKILLS_ROOT"/*/SKILL.md; do
  [[ -f "$skill_check" ]] || continue
  check_dir="$(dirname "$skill_check")"
  check_name="$(basename "$check_dir")"
  if ! get_frontmatter "$skill_check" "skill_api_version" >/dev/null 2>&1; then
    report "MISSING_API_VERSION" "$check_dir" "No skill_api_version field in frontmatter"
    if [[ "$MODE" == "fix" ]]; then
      # Insert skill_api_version: 1 after description: line
      tmp="$(mktemp)"
      inserted=0
      while IFS= read -r line; do
        echo "$line" >> "$tmp"
        if [[ $inserted -eq 0 && "$line" =~ ^description: ]]; then
          echo "skill_api_version: 1" >> "$tmp"
          inserted=1
        fi
      done < "$skill_check"
      /bin/cp "$tmp" "$skill_check"
      rm -f "$tmp"
    fi
  fi
done

if [[ $FINDINGS -gt 0 ]]; then
  echo ""
  echo "$FINDINGS finding(s) detected."
  if [[ $STRICT -eq 1 || "$MODE" == "fix" ]]; then
    exit 1
  fi
  exit 0
else
  echo "All clean. No findings."
  exit 0
fi
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


