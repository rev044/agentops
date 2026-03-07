#!/usr/bin/env bash
set -euo pipefail

# check-skill-flag-refs.sh
# Cross-references CLI flags mentioned in skills/*/SKILL.md (and references/*.md)
# against actual flag registrations in the Go CLI source (cli/cmd/ao/).
#
# Only checks flags in the context of `ao` commands. External tools (git, codex,
# go, radon, etc.) are excluded by design.
#
# Exit 0: clean (no unregistered flags found)
# Exit 1: findings (flags referenced but not registered)

ROOT="${1:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
SKILLS_DIR="$ROOT/skills"
CLI_DIR="$ROOT/cli/cmd/ao"

failures=0
warnings=0

fail() { echo "  FAIL: $1"; failures=$((failures + 1)); }
warn() { echo "  WARN: $1"; warnings=$((warnings + 1)); }
pass() { : ; }  # silent on pass

# ── Allowlist: flags that are valid but not found via simple grep ──
# Root persistent flags (available on every command)
# Universal flags (--help, --version are built-in by cobra)
# Flags used in prose/explanation contexts that don't map 1:1 to Go registrations
ALLOWLIST=(
  # Cobra built-ins
  "help"
  "version"
  # Root persistent flags (registered on rootCmd.PersistentFlags)
  "dry-run"
  "verbose"
  "output"
  "json"
  "config"
  # goals persistent flags (registered on goalsCmd.PersistentFlags)
  "file"
  "timeout"
  # Skill-only flags (parsed by skill logic, not CLI)
  "max-cycles"
  "beads-only"
  "skip-baseline"
  "quality"
  "auto"
  "interactive"
  "loop"
  "test-first"
  "quick"
  "deep"
  "mixed"
  "debate"
  "preset"
  "explorers"
  "perspectives-file"
  "no-beads"
  "markdown"
  # /evolve parse-level flags (not ao CLI flags)
  # /rpi skill-level flags (not ao CLI flags)
  "from"
  # /learn skill-level flags (global tier targeting)
  "global"
  "promote"
)

# Build a lookup set from the allowlist
declare -A ALLOWED
for flag in "${ALLOWLIST[@]}"; do
  ALLOWED["$flag"]=1
done

echo "=== Skill-to-CLI flag cross-reference check ==="
echo ""

# ── Step 1: Verify directories exist ──

if [[ ! -d "$SKILLS_DIR" ]]; then
  echo "FAIL: skills/ directory not found at $SKILLS_DIR"
  exit 1
fi

if [[ ! -d "$CLI_DIR" ]]; then
  echo "FAIL: cli/cmd/ao/ directory not found at $CLI_DIR"
  exit 1
fi

# ── Step 2: Collect all registered CLI flags from Go source ──
# Extract flag names from .Flags().XxxVar / .PersistentFlags().XxxVar patterns
# Match the quoted flag name in the registration call

declare -A REGISTERED_FLAGS

while IFS= read -r flag_name; do
  [[ -z "$flag_name" ]] && continue
  REGISTERED_FLAGS["$flag_name"]=1
done < <(
  grep -hroE 'Flags\(\)\.[A-Za-z]+\([^)]*"[^"]*"' "$CLI_DIR"/*.go 2>/dev/null \
    | grep -oE '"[a-z][a-z0-9-]*"' \
    | tr -d '"' \
    | sort -u
)

# Also pick up PersistentFlags registrations
while IFS= read -r flag_name; do
  [[ -z "$flag_name" ]] && continue
  REGISTERED_FLAGS["$flag_name"]=1
done < <(
  grep -hroE 'PersistentFlags\(\)\.[A-Za-z]+\([^)]*"[^"]*"' "$CLI_DIR"/*.go 2>/dev/null \
    | grep -oE '"[a-z][a-z0-9-]*"' \
    | tr -d '"' \
    | sort -u
)

flag_count=${#REGISTERED_FLAGS[@]}
echo "Found $flag_count registered CLI flags in Go source."
echo ""

# ── Step 3: Scan skill docs for ao command + flag patterns ──
# We look for lines containing `ao <subcommand> ... --<flag>`
# This captures patterns like:
#   ao goals measure --json
#   ao lookup --apply-decay --format markdown
#   ao ratchet record implement --output "<path>"
#   ao rpi cleanup --all --prune-worktrees
#
# We exclude lines that are clearly about external tools.

echo "--- Scanning skill docs for ao flag references ---"
echo ""

checked=0
skill_files_checked=0

# Process all SKILL.md and references/*.md files under skills/
while IFS= read -r skill_file; do
  rel_path="${skill_file#"$ROOT"/}"
  file_has_issues=false

  # Extract lines that reference `ao` commands with flags
  # Pattern: `ao` followed by subcommand(s) and --flag
  while IFS= read -r line; do
    [[ -z "$line" ]] && continue

    # Extract all --flag-name tokens from lines that start with or contain `ao `
    # but only from lines where `ao` appears as a command (not in prose like "also")
    # We require `ao ` to appear as a word boundary before subcommand + flags
    if ! echo "$line" | grep -qE '(^|[^a-z])ao [a-z]'; then
      continue
    fi

    # Skip lines that are clearly comments about ao being unavailable
    if echo "$line" | grep -qiE '(unavailable|not installed|if.*available)'; then
      continue
    fi

    # Extract all --flag-name patterns from this line
    while IFS= read -r flag; do
      [[ -z "$flag" ]] && continue
      checked=$((checked + 1))

      # Strip leading --
      flag_name="${flag#--}"

      # Strip =value suffix if present (e.g., --weight=5 -> weight)
      flag_name="${flag_name%%=*}"

      # Skip if in allowlist
      if [[ -n "${ALLOWED[$flag_name]+x}" ]]; then
        pass "$rel_path: --$flag_name (allowlisted)"
        continue
      fi

      # Check if registered in CLI
      if [[ -n "${REGISTERED_FLAGS[$flag_name]+x}" ]]; then
        pass "$rel_path: --$flag_name"
      else
        if [[ "$file_has_issues" == false ]]; then
          echo "$rel_path:"
          file_has_issues=true
        fi
        fail "--$flag_name not found in CLI registrations (line: $(echo "$line" | sed 's/^[[:space:]]*//' | head -c 120))"
      fi
    done < <(echo "$line" | grep -oE -- '--[a-z][a-z0-9-]*' | sort -u)

  done < <(
    # Extract lines containing `ao ` followed by what looks like a subcommand
    # Filter out code block markers and pure comment lines
    grep -E '(^|[^a-z])ao [a-z]' "$skill_file" 2>/dev/null \
      | grep -v '^```' \
      | grep -v '^#' \
      || true
  )

  skill_files_checked=$((skill_files_checked + 1))

done < <(
  find "$SKILLS_DIR" -type f \( -name 'SKILL.md' -o -path '*/references/*.md' \) 2>/dev/null | sort
)

echo ""
echo "--- Summary ---"
echo "Skill files checked: $skill_files_checked"
echo "Flag references checked: $checked"
echo "Findings: $failures"
echo "Warnings: $warnings"

if [[ $failures -gt 0 ]]; then
  echo ""
  echo "Flag cross-reference check FAILED ($failures finding(s))."
  echo ""
  echo "To fix: either register the flag in the CLI (cli/cmd/ao/), update the"
  echo "skill doc to remove the invalid reference, or add to the allowlist in"
  echo "this script if the flag is intentional (e.g., skill-only parse flags)."
  exit 1
fi

echo ""
echo "Flag cross-reference check passed."
exit 0
