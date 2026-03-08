#!/usr/bin/env bash
set -euo pipefail

ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
OVERRIDES_DIR="$ROOT/skills-codex-overrides"
GENERATED_DIR="$ROOT/skills-codex"

required_skills=(
  research
  plan
  crank
  vibe
  rpi
  evolve
  post-mortem
)

failures=0

fail() {
  echo "FAIL: $1" >&2
  failures=$((failures + 1))
}

for skill in "${required_skills[@]}"; do
  override_prompt="$OVERRIDES_DIR/$skill/prompt.md"
  generated_prompt="$GENERATED_DIR/$skill/prompt.md"

  [[ -f "$generated_prompt" ]] || fail "missing generated Codex prompt for $skill"
  [[ -f "$override_prompt" ]] || fail "missing Codex override prompt for $skill"

  if [[ -f "$override_prompt" ]] && ! rg -q '^## Codex Execution Profile$' "$override_prompt"; then
    fail "override prompt for $skill lacks '## Codex Execution Profile'"
  fi
done

if [[ "$failures" -gt 0 ]]; then
  echo "Codex override coverage validation FAILED ($failures finding(s))." >&2
  exit 1
fi

echo "Codex override coverage validation passed for ${#required_skills[@]} core skill(s)."
