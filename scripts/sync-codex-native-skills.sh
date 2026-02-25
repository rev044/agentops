#!/usr/bin/env bash
#
# Build Codex-native skill artifacts into a dedicated area (default: ./skills-codex).
# Output is generated from ./skills via the converter's codex target.
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CONVERTER="$REPO_ROOT/skills/converter/scripts/convert.sh"

SRC="$REPO_ROOT/skills"
OUT="$REPO_ROOT/skills-codex"
ONLY_CSV=""

usage() {
  cat <<'EOF'
sync-codex-native-skills.sh

Builds Codex-native skill folders from source skills.

Options:
  --src <dir>         Source skills root (default: ./skills)
  --out <dir>         Output codex skills root (default: ./skills-codex)
  --only <a,b,c>      Only build these skill names (comma-separated)
  --help              Show this help

Examples:
  ./scripts/sync-codex-native-skills.sh
  ./scripts/sync-codex-native-skills.sh --only research,vibe
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
    bash "$CONVERTER" "$skill_dir" codex "$tmpdir/$skill"
  done
else
  bash "$CONVERTER" --all codex "$tmpdir"
fi

rm -rf "$OUT"
mkdir -p "$OUT"
rsync -a --delete --copy-links "$tmpdir"/ "$OUT"/

count="$(find "$OUT" -mindepth 1 -maxdepth 1 -type d | wc -l | tr -d ' ')"
echo "Codex-native skills synced: $count"
echo "Output: $OUT"
