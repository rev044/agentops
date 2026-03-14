#!/usr/bin/env bash
# validate-codex-api-conformance.sh — Check generated codex skills against Codex API contract.
# Exit 0 = pass, exit 1 = failures found.
# Contract: docs/contracts/codex-skill-api.md
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SKILLS_ROOT="$REPO_ROOT/skills-codex"

failures=0
warnings=0

if [[ ! -d "$SKILLS_ROOT" ]]; then
  echo "Error: skills-codex directory not found: $SKILLS_ROOT" >&2
  exit 1
fi

# --- Check 1: Frontmatter must only contain name + description ---
echo "=== Check 1: Frontmatter fields ==="
while IFS= read -r skill_md; do
  skill_name="$(basename "$(dirname "$skill_md")")"

  # Extract frontmatter between first and second ---
  fm=$(awk 'NR==1 && /^---$/{in_fm=1; next} in_fm && /^---$/{exit} in_fm{print}' "$skill_md")

  # Check for non-Codex fields
  bad_fields=$(echo "$fm" | grep -oE '^[a-z_-]+:' | sed 's/:$//' | grep -vE '^(name|description)$' || true)
  if [[ -n "$bad_fields" ]]; then
    echo "  FAIL [$skill_name] Non-Codex frontmatter fields: $(echo "$bad_fields" | tr '\n' ', ')"
    failures=$((failures + 1))
  fi
done < <(find "$SKILLS_ROOT" -mindepth 2 -maxdepth 2 -name 'SKILL.md' -type f | sort)

# --- Check 2: No Claude-only primitive names ---
echo "=== Check 2: Claude primitive references ==="
# Claude-only primitives that have NO Codex equivalent (not even mapped ones)
# Note: TaskCreate→todo_write, TaskList→update_plan etc. are valid Codex mappings
CLAUDE_PRIMITIVES='TeamCreate|TeamDelete|SendMessage|EnterPlanMode|ExitPlanMode|EnterWorktree|team-create|team-delete|send-message|enter-plan-mode|exit-plan-mode|enter-worktree'

while IFS= read -r skill_md; do
  skill_name="$(basename "$(dirname "$skill_md")")"

  # Search body (after frontmatter) for Claude primitives
  body=$(awk 'BEGIN{skip=0} NR==1 && /^---$/{skip=1; next} skip && /^---$/{skip=0; next} !skip{print}' "$skill_md")
  matches=$(echo "$body" | grep -onE "\b($CLAUDE_PRIMITIVES)\b" 2>/dev/null || true)
  if [[ -n "$matches" ]]; then
    count=$(echo "$matches" | wc -l | tr -d ' ')
    echo "  FAIL [$skill_name] $count Claude primitive reference(s)"
    failures=$((failures + 1))
  fi
done < <(find "$SKILLS_ROOT" -mindepth 2 -maxdepth 2 -name 'SKILL.md' -type f | sort)

# --- Check 3: No Claude-specific paths ---
echo "=== Check 3: Claude-specific paths ==="
while IFS= read -r skill_md; do
  skill_name="$(basename "$(dirname "$skill_md")")"

  matches=$(grep -n '~/\.claude/' "$skill_md" 2>/dev/null || true)
  if [[ -n "$matches" ]]; then
    count=$(echo "$matches" | wc -l | tr -d ' ')
    echo "  FAIL [$skill_name] $count ~/.claude/ path reference(s)"
    failures=$((failures + 1))
  fi
done < <(find "$SKILLS_ROOT" -mindepth 2 -maxdepth 2 -name 'SKILL.md' -type f | sort)

# --- Check 4: agents/openai.yaml validity (if present) ---
echo "=== Check 4: agents/openai.yaml validity ==="
while IFS= read -r yaml_file; do
  skill_name="$(basename "$(dirname "$(dirname "$yaml_file")")")"
  # Basic YAML syntax check
  if ! python3 -c "import yaml; yaml.safe_load(open('$yaml_file'))" 2>/dev/null; then
    echo "  FAIL [$skill_name] Invalid YAML: $yaml_file"
    failures=$((failures + 1))
  fi
done < <(find "$SKILLS_ROOT" -path '*/agents/openai.yaml' -type f 2>/dev/null | sort)

# --- Summary ---
echo ""
echo "=== Summary ==="
echo "Failures: $failures"
echo "Warnings: $warnings"

if [[ $failures -gt 0 ]]; then
  echo ""
  echo "Codex API conformance check FAILED with $failures failure(s)."
  exit 1
else
  echo ""
  echo "Codex API conformance check passed."
  exit 0
fi
