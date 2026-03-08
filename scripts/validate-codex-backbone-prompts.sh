#!/usr/bin/env bash
set -euo pipefail

ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
CATALOG_PATH="$ROOT/skills-codex-overrides/catalog.json"
GENERATED_DIR="$ROOT/skills-codex"

usage() {
  cat <<'EOF'
Usage: bash scripts/validate-codex-backbone-prompts.sh [--repo-root <path>]
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --repo-root)
      ROOT="${2:-}"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown arg: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

if [[ "$ROOT" != /* ]]; then
  ROOT="$(cd "$ROOT" && pwd)"
fi

CATALOG_PATH="$ROOT/skills-codex-overrides/catalog.json"
GENERATED_DIR="$ROOT/skills-codex"

failures=0

fail() {
  echo "FAIL: $1" >&2
  failures=$((failures + 1))
}

require_file() {
  local path="$1"
  [[ -f "$path" ]] || {
    echo "Missing required file: ${path#$ROOT/}" >&2
    exit 1
  }
}

find_fixed_line() {
  local needle="$1"
  local path="$2"
  grep -nF -m1 -- "$needle" "$path" | cut -d: -f1
}

tmpdir="$(mktemp -d)"
cleanup() {
  rm -rf "$tmpdir"
}
trap cleanup EXIT

selected_entries_file="$tmpdir/backbone-entries.jsonl"

require_file "$CATALOG_PATH"
[[ -d "$GENERATED_DIR" ]] || {
  echo "Missing generated Codex skills root: $GENERATED_DIR" >&2
  exit 1
}
command -v jq >/dev/null 2>&1 || {
  echo "jq is required for Codex backbone prompt validation." >&2
  exit 1
}

if ! jq -e '
  (.skills | type) == "array" and
  all(.skills[];
    (.name | type) == "string" and
    ((.operator_contract? | not) or
      (
        (.operator_contract | type) == "object" and
        (.operator_contract.required_sections | type) == "array" and
        (.operator_contract.required_markers | type) == "array"
      )
    )
  )
' "$CATALOG_PATH" >/dev/null; then
  echo "Invalid Codex override catalog schema for backbone prompt validation: $CATALOG_PATH" >&2
  exit 1
fi

jq -c '
  .skills[]
  | select(.operator_contract != null)
  | {
      name,
      required_sections: .operator_contract.required_sections,
      required_markers: .operator_contract.required_markers
    }
' "$CATALOG_PATH" > "$selected_entries_file"

selected_count="$(wc -l < "$selected_entries_file" | tr -d ' ')"
if [[ "$selected_count" == "0" ]]; then
  fail "no backbone Codex operator contracts found in catalog"
fi

while IFS= read -r entry; do
  [[ -n "$entry" ]] || continue
  skill="$(jq -r '.name' <<<"$entry")"
  prompt_path="$GENERATED_DIR/$skill/prompt.md"

  if [[ ! -f "$prompt_path" ]]; then
    fail "missing generated Codex backbone prompt: skills-codex/$skill/prompt.md"
    continue
  fi

  previous_line=0
  while IFS= read -r section; do
    [[ -n "$section" ]] || continue
    line="$(find_fixed_line "$section" "$prompt_path" || true)"
    if [[ -z "$line" ]]; then
      fail "generated prompt for $skill is missing required section: $section"
      continue
    fi
    if (( line <= previous_line )); then
      fail "generated prompt for $skill has out-of-order sections around: $section"
    fi
    previous_line="$line"
  done < <(jq -r '.required_sections[]' <<<"$entry")

  while IFS= read -r marker; do
    [[ -n "$marker" ]] || continue
    if ! grep -Fq -- "$marker" "$prompt_path"; then
      fail "generated prompt for $skill is missing required behavior marker: $marker"
    fi
  done < <(jq -r '.required_markers[]' <<<"$entry")
done < "$selected_entries_file"

if [[ "$failures" -gt 0 ]]; then
  echo "Codex backbone prompt validation FAILED ($failures finding(s))." >&2
  exit 1
fi

echo "Codex backbone prompt validation passed for $selected_count skill(s)."
