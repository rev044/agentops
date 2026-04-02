#!/usr/bin/env bash
set -euo pipefail

# release-cadence-check.sh
# Compatibility wrapper retained after removing the enforced release cadence policy.
#
# Usage:
#   ./scripts/release-cadence-check.sh
#
# Exit codes:
#   0 = informational pass

if [[ $# -gt 0 ]]; then
    echo "INFO: release cadence flags are ignored; no minimum release spacing is enforced."
fi

echo "PASS: Release cadence policy removed; releases may ship whenever maintainers decide they are ready."
exit 0
