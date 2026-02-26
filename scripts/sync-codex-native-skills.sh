#!/usr/bin/env bash
#
# Build Codex-native skill artifacts into a dedicated area (default: ./skills-codex).
# Output is generated from ./skills via the converter's codex target and then
# optionally overlaid with codex-specific overrides.
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CONVERTER="$REPO_ROOT/skills/converter/scripts/convert.sh"

SRC="$REPO_ROOT/skills"
OUT="$REPO_ROOT/skills-codex"
OVERRIDES="$REPO_ROOT/skills-codex-overrides"
ONLY_CSV=""
SKIP_OVERRIDES="false"

usage() {
  cat <<'EOF'
sync-codex-native-skills.sh

Builds Codex-native skill folders from source skills.

Options:
  --src <dir>         Source skills root (default: ./skills)
  --out <dir>         Output codex skills root (default: ./skills-codex)
  --overrides <dir>   Codex-only override layer (default: ./skills-codex-overrides)
  --skip-overrides    Do not apply override layer
  --only <a,b,c>      Only build these skill names (comma-separated)
  --help              Show this help

Examples:
  ./scripts/sync-codex-native-skills.sh
  ./scripts/sync-codex-native-skills.sh --only research,vibe
  ./scripts/sync-codex-native-skills.sh --overrides ./skills-codex-overrides
  ./scripts/sync-codex-native-skills.sh --out /tmp/skills-codex
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --src)
      SRC="${2:-}"
      shift 2
      ;;
    --out)
      OUT="${2:-}"
      shift 2
      ;;
    --overrides)
      OVERRIDES="${2:-}"
      shift 2
      ;;
    --skip-overrides)
      SKIP_OVERRIDES="true"
      shift 1
      ;;
    --only)
      ONLY_CSV="${2:-}"
      shift 2
      ;;
    --help|-h)
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

if [[ "$SRC" != /* ]]; then
  SRC="$REPO_ROOT/$SRC"
fi
if [[ "$OUT" != /* ]]; then
  OUT="$REPO_ROOT/$OUT"
fi
if [[ "$OVERRIDES" != /* ]]; then
  OVERRIDES="$REPO_ROOT/$OVERRIDES"
fi

[[ -x "$CONVERTER" ]] || {
  echo "Error: converter script not executable: $CONVERTER" >&2
  exit 1
}
[[ -d "$SRC" ]] || {
  echo "Error: source skills directory not found: $SRC" >&2
  exit 1
}

tmpdir="$(mktemp -d)"
cleanup() {
  rm -rf "$tmpdir"
}
trap cleanup EXIT

write_expected_skills() {
  local out_file="$1"

  if [[ -n "$ONLY_CSV" ]]; then
    IFS=',' read -r -a only_arr <<<"$ONLY_CSV"
    {
      for skill in "${only_arr[@]}"; do
        skill="$(echo "$skill" | xargs)"
        [[ -n "$skill" ]] || continue
        skill_dir="$SRC/$skill"
        [[ -d "$skill_dir" ]] || {
          echo "Error: requested skill not found under src: $skill" >&2
          exit 1
        }
        [[ -f "$skill_dir/SKILL.md" ]] || {
          echo "Error: requested skill missing SKILL.md under src: $skill" >&2
          exit 1
        }
        echo "$skill"
      done
    } | sort -u > "$out_file"
    return
  fi

  find "$SRC" -mindepth 1 -maxdepth 1 -type d \
    | while IFS= read -r d; do
        [[ -f "$d/SKILL.md" ]] || continue
        basename "$d"
      done \
    | sort -u > "$out_file"
}

write_actual_skills() {
  local built_root="$1"
  local out_file="$2"
  find "$built_root" -mindepth 1 -maxdepth 1 -type d \
    | while IFS= read -r d; do
        [[ -f "$d/SKILL.md" ]] || continue
        basename "$d"
      done \
    | sort -u > "$out_file"
}

apply_overrides() {
  local built_root="$1"
  local applied=0

  if [[ "$SKIP_OVERRIDES" == "true" ]]; then
    echo "Codex overrides skipped (--skip-overrides)."
    return
  fi

  if [[ ! -d "$OVERRIDES" ]]; then
    echo "Codex overrides directory not found; continuing without overrides: $OVERRIDES"
    return
  fi

  while IFS= read -r -d '' override_dir; do
    local skill
    skill="$(basename "$override_dir")"
    if [[ ! -d "$built_root/$skill" ]]; then
      if [[ -n "$ONLY_CSV" ]]; then
        continue
      fi
      echo "Error: override exists for unknown skill '$skill' (no generated output)." >&2
      exit 1
    fi
    rsync -a --copy-links "$override_dir"/ "$built_root/$skill"/
    applied=$((applied + 1))
  done < <(find "$OVERRIDES" -mindepth 1 -maxdepth 1 -type d -print0)

  echo "Codex overrides applied: $applied"
}

validate_parity() {
  local built_root="$1"
  local expected_file="$tmpdir/.expected-skills.txt"
  local actual_file="$tmpdir/.actual-skills.txt"
  local missing extra

  write_expected_skills "$expected_file"
  write_actual_skills "$built_root" "$actual_file"

  missing="$(comm -23 "$expected_file" "$actual_file" || true)"
  extra="$(comm -13 "$expected_file" "$actual_file" || true)"

  if [[ -n "$missing" || -n "$extra" ]]; then
    echo "Error: codex skill parity check failed." >&2
    if [[ -n "$missing" ]]; then
      echo "Missing in codex output:" >&2
      echo "$missing" >&2
    fi
    if [[ -n "$extra" ]]; then
      echo "Unexpected extras in codex output:" >&2
      echo "$extra" >&2
    fi
    exit 1
  fi

  local count
  count="$(wc -l < "$expected_file" | tr -d ' ')"
  echo "Codex skill parity check passed: $count skill(s)."
}

if [[ -n "$ONLY_CSV" ]]; then
  IFS=',' read -r -a only_arr <<<"$ONLY_CSV"
  for skill in "${only_arr[@]}"; do
    skill="$(echo "$skill" | xargs)"
    [[ -n "$skill" ]] || continue
    skill_dir="$SRC/$skill"
    [[ -d "$skill_dir" ]] || {
      echo "Error: requested skill not found under src: $skill" >&2
      exit 1
    }
    bash "$CONVERTER" --codex-layout modular "$skill_dir" codex "$tmpdir/$skill"
  done
else
  bash "$CONVERTER" --codex-layout modular --all codex "$tmpdir"
fi

# Pre-validate: check for unallowlisted residual markers
if [[ -x "${SCRIPT_DIR}/lint/generate-allowlist-candidates.sh" ]]; then
  echo "Checking for unallowlisted residual markers..."
  if ! bash "${SCRIPT_DIR}/lint/generate-allowlist-candidates.sh" "$tmpdir"; then
    echo "WARNING: Unallowlisted markers found. Add to allowlist or fix converter rules."
    # Non-blocking warning — validation gate will catch in CI
  fi
fi

apply_overrides "$tmpdir"
validate_parity "$tmpdir"

rm -rf "$OUT"
mkdir -p "$OUT"
rsync -a --delete --copy-links "$tmpdir"/ "$OUT"/

count="$(find "$OUT" -mindepth 1 -maxdepth 1 -type d | wc -l | tr -d ' ')"
echo "Codex-native skills synced: $count"
echo "Output: $OUT"
