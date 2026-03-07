#!/usr/bin/env bash
# validate-codex-install-bundle.sh — ensure release archive ships current Codex skills
#
# Builds a git archive for the selected ref, regenerates skills-codex from the
# archived sources, and fails if the archived skills-codex tree differs from the
# regenerated output. This protects curl-based Codex installs from shipping a
# stale prebuilt bundle.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

REF=""
KEEP_TMP=false

usage() {
    cat <<'EOF'
Usage: bash scripts/validate-codex-install-bundle.sh [--ref <git-ref>] [--keep-tmp]

Options:
  --ref <git-ref>   Git ref to archive and validate (default: current worktree)
  --keep-tmp        Keep temporary files on failure for inspection
  -h, --help        Show this help
EOF
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --ref)
            REF="${2:-}"
            shift 2
            ;;
        --keep-tmp)
            KEEP_TMP=true
            shift
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            echo "Unknown option: $1" >&2
            usage >&2
            exit 2
            ;;
    esac
done

for cmd in git tar diff mktemp; do
    if ! command -v "$cmd" >/dev/null 2>&1; then
        echo "Missing required command: $cmd" >&2
        exit 1
    fi
done

git -C "$REPO_ROOT" rev-parse --show-toplevel >/dev/null 2>&1 || {
    echo "Not a git repository: $REPO_ROOT" >&2
    exit 1
}

TMP_DIR="$(mktemp -d)"
# shellcheck disable=SC2329
cleanup() {
    local rc=$?
    if [[ "$KEEP_TMP" == "true" && "$rc" -ne 0 ]]; then
        echo "Keeping temp dir for inspection: $TMP_DIR" >&2
        return
    fi
    rm -rf "$TMP_DIR"
}
trap cleanup EXIT

BUNDLE_DIR="$TMP_DIR/bundle"
GENERATED_DIR="$TMP_DIR/generated"
ARCHIVE_FILE="$TMP_DIR/release-bundle.tar"
DIFF_FILE="$TMP_DIR/codex-bundle.diff"

mkdir -p "$BUNDLE_DIR" "$GENERATED_DIR"

archive_label="working tree"
if [[ -n "$REF" ]]; then
    git -C "$REPO_ROOT" rev-parse --verify "${REF}^{commit}" >/dev/null 2>&1 || {
        echo "Unknown git ref: $REF" >&2
        exit 1
    }
    git -C "$REPO_ROOT" archive --format=tar --output "$ARCHIVE_FILE" "$REF"
    archive_label="$REF"
else
    tmp_index="$TMP_DIR/index"
    base_tree="$(git -C "$REPO_ROOT" rev-parse "HEAD^{tree}" 2>/dev/null || git -C "$REPO_ROOT" hash-object -t tree /dev/null)"

    GIT_INDEX_FILE="$tmp_index" git -C "$REPO_ROOT" read-tree "$base_tree"
    GIT_INDEX_FILE="$tmp_index" git -C "$REPO_ROOT" add -A -- .
    tree_id="$(GIT_INDEX_FILE="$tmp_index" git -C "$REPO_ROOT" write-tree)"
    GIT_INDEX_FILE="$tmp_index" git -C "$REPO_ROOT" archive --format=tar --output "$ARCHIVE_FILE" "$tree_id"
fi
tar -xf "$ARCHIVE_FILE" -C "$BUNDLE_DIR"

for required_path in \
    "$BUNDLE_DIR/.codex-plugin/plugin.json" \
    "$BUNDLE_DIR/.agents/plugins/marketplace.json" \
    "$BUNDLE_DIR/skills" \
    "$BUNDLE_DIR/skills-codex" \
    "$BUNDLE_DIR/scripts/sync-codex-native-skills.sh"
do
    if [[ ! -e "$required_path" ]]; then
        echo "Release bundle missing required path: ${required_path#"$BUNDLE_DIR"/}" >&2
        exit 1
    fi
done

(
    cd "$BUNDLE_DIR"
    bash scripts/sync-codex-native-skills.sh --out "$GENERATED_DIR" >/dev/null
)

if diff -ruN "$BUNDLE_DIR/skills-codex" "$GENERATED_DIR" >"$DIFF_FILE"; then
    skill_count="$(find "$BUNDLE_DIR/skills-codex" -mindepth 2 -maxdepth 2 -name SKILL.md | wc -l | tr -d ' ')"
    echo "Codex install bundle parity OK for $archive_label ($skill_count skill package(s))."
    exit 0
fi

echo "Codex install bundle drift detected for $archive_label." >&2
echo "Archived skills-codex does not match regenerated output from archived skills." >&2
echo "Diff preview:" >&2
sed -n '1,80p' "$DIFF_FILE" >&2
exit 1
