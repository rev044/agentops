#!/usr/bin/env bash
#
# Build Codex-native skills and install them into ~/.codex/skills.
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SYNC_SCRIPT="$REPO_ROOT/scripts/sync-codex-native-skills.sh"
EXPORT_SCRIPT="$REPO_ROOT/scripts/export-claude-skills-to-codex.sh"

SRC="$REPO_ROOT/skills-codex"
DST="$HOME/.codex/skills"
INSTALL_META="$HOME/.codex/.agentops-codex-install.json"
BACKUP=""
ONLY_CSV=""
SKIP_SYNC="false"
DRY_RUN="false"

usage() {
  cat <<'EOF'
install-codex-native-skills.sh

Builds Codex-native skills into ./skills-codex and installs them to ~/.codex/skills.

Options:
  --source <dir>      Codex-native source skills root (default: ./skills-codex)
  --dest <dir>        Destination Codex skills root (default: ~/.codex/skills)
  --backup <dir>      Backup directory (default: ~/.codex/skills.backup.<timestamp>)
  --only <a,b,c>      Only install these skill names (comma-separated)
  --skip-sync         Skip build step and install from existing --source
  --dry-run           Preview install copy operations
  --help              Show this help

Examples:
  ./scripts/install-codex-native-skills.sh
  ./scripts/install-codex-native-skills.sh --only research,vibe
  ./scripts/install-codex-native-skills.sh --dest /tmp/codex-skills --dry-run
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --source)
      SRC="${2:-}"
      shift 2
      ;;
    --dest)
      DST="${2:-}"
      shift 2
      ;;
    --backup)
      BACKUP="${2:-}"
      shift 2
      ;;
    --only)
      ONLY_CSV="${2:-}"
      shift 2
      ;;
    --skip-sync)
      SKIP_SYNC="true"
      shift 1
      ;;
    --dry-run)
      DRY_RUN="true"
      shift 1
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
if [[ "$DST" != /* ]]; then
  DST="$REPO_ROOT/$DST"
fi

if [[ "$SKIP_SYNC" != "true" ]]; then
  sync_args=(--src "$REPO_ROOT/skills" --out "$SRC")
  if [[ -n "$ONLY_CSV" ]]; then
    sync_args+=(--only "$ONLY_CSV")
  fi
  bash "$SYNC_SCRIPT" "${sync_args[@]}"
fi

[[ -d "$SRC" ]] || {
  echo "Error: source codex skills directory not found: $SRC" >&2
  echo "Run without --skip-sync or build first via scripts/sync-codex-native-skills.sh." >&2
  exit 1
}

export_args=(--src "$SRC" --dst "$DST")
if [[ -n "$BACKUP" ]]; then
  export_args+=(--backup "$BACKUP")
fi
if [[ -n "$ONLY_CSV" ]]; then
  export_args+=(--only "$ONLY_CSV")
fi
if [[ "$DRY_RUN" == "true" ]]; then
  export_args+=(--dry-run)
fi

bash "$EXPORT_SCRIPT" "${export_args[@]}"

# Write local install metadata for stale-skill reminders (no telemetry)
INSTALLED_AT="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
VERSION="unknown"
if git -C "$REPO_ROOT" rev-parse --short HEAD >/dev/null 2>&1; then
  VERSION="$(git -C "$REPO_ROOT" rev-parse --short HEAD)"
fi
mkdir -p "$(dirname "$INSTALL_META")"
cat > "$INSTALL_META" <<EOF
{
  "installed_at": "$INSTALLED_AT",
  "source": "install-codex-native-skills.sh",
  "version": "$VERSION",
  "update_command": "curl -fsSL https://raw.githubusercontent.com/boshu2/agentops/main/scripts/install-codex.sh | bash"
}
EOF
echo "Install metadata written: $INSTALL_META"

echo "Restart Codex to pick up installed skills."
