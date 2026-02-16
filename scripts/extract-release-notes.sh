#!/usr/bin/env bash
# Extract release notes for a given version from CHANGELOG.md.
# Wraps with install/verification header and full-changelog footer.
#
# Usage: scripts/extract-release-notes.sh v2.9.1 [v2.9.0]
#   $1 = current tag (required)
#   $2 = previous tag (optional, for footer link)
#
# Output: writes release-notes.md to repo root

set -euo pipefail

TAG="${1:?Usage: extract-release-notes.sh TAG [PREV_TAG]}"
PREV_TAG="${2:-}"
VERSION="${TAG#v}"
REPO="boshu2/agentops"

CHANGELOG="CHANGELOG.md"
if [[ ! -f "$CHANGELOG" ]]; then
  echo "ERROR: $CHANGELOG not found" >&2
  exit 1
fi

# Extract the section for this version from CHANGELOG.md.
# Matches from "## [VERSION]" to the next "## [" line (exclusive).
NOTES=$(awk -v ver="$VERSION" '
  /^## \[/ {
    if (found) exit
    if (index($0, "[" ver "]")) { found=1; next }
  }
  found { print }
' "$CHANGELOG")

if [[ -z "$NOTES" ]]; then
  echo "WARN: No CHANGELOG entry for $VERSION — falling back to commit summary" >&2
  NOTES="*No curated release notes for this version. See commit history below.*"
  # Append abbreviated commit log as fallback
  if [[ -n "$PREV_TAG" ]]; then
    NOTES="$NOTES"$'\n\n'"$(git log --oneline "${PREV_TAG}..${TAG}" 2>/dev/null || echo "(git log unavailable)")"
  fi
fi

# Build the release notes file
cat > release-notes.md <<EOF
\`brew upgrade agentops\` · [checksums](https://github.com/${REPO}/releases/download/${TAG}/checksums.txt) · [verify provenance](https://docs.github.com/en/actions/security-for-github-actions/using-artifact-attestations/using-artifact-attestations-to-establish-provenance-for-builds)

---

${NOTES}

---

**Full Changelog**: https://github.com/${REPO}/compare/${PREV_TAG:-v0.0.0}...${TAG}
EOF

echo "Release notes written to release-notes.md ($(wc -l < release-notes.md) lines)"
