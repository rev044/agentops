#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
ARTIFACT_ROOT="$REPO_ROOT/.agents/releases/local-ci"

usage() {
    cat <<'USAGE'
Usage: scripts/resolve-release-artifacts.sh <version>

Find the newest full local-CI artifact set for a release version and print its
manifest JSON. The version may be passed as X.Y.Z or vX.Y.Z.
USAGE
}

if [[ $# -ne 1 || "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
    usage
    [[ $# -eq 1 ]] && exit 0
    exit 1
fi

VERSION="${1#v}"

if [[ ! -d "$ARTIFACT_ROOT" ]]; then
    echo "ERROR: local CI artifact root not found: $ARTIFACT_ROOT" >&2
    exit 1
fi

while IFS= read -r manifest; do
    if ! jq -e --arg version "$VERSION" '
        .schema_version == 1 and
        .release_version == $version and
        .fast_mode == false and
        (.artifact_dir | type == "string" and length > 0) and
        (.sbom_cyclonedx | type == "string" and length > 0) and
        (.sbom_spdx | type == "string" and length > 0) and
        (.security_report | type == "string" and length > 0)
    ' "$manifest" >/dev/null 2>&1; then
        continue
    fi

    artifact_dir="$(jq -r '.artifact_dir' "$manifest")"
    sbom_cyclonedx="$(jq -r '.sbom_cyclonedx' "$manifest")"
    sbom_spdx="$(jq -r '.sbom_spdx' "$manifest")"
    security_report="$(jq -r '.security_report' "$manifest")"

    if [[ -f "$REPO_ROOT/$artifact_dir/$sbom_cyclonedx" && \
          -f "$REPO_ROOT/$artifact_dir/$sbom_spdx" && \
          -f "$REPO_ROOT/$artifact_dir/$security_report" ]]; then
        cat "$manifest"
        exit 0
    fi
done < <(find "$ARTIFACT_ROOT" -mindepth 2 -maxdepth 2 -type f -name 'release-artifacts.json' | sort -r)

echo "ERROR: no full local CI artifacts found for release version $VERSION" >&2
echo "Run: ./scripts/ci-local-release.sh --release-version $VERSION" >&2
exit 1
