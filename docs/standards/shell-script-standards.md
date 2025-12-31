# Shell Script Standards

> **Purpose:** Standardized shell scripting conventions for this repository.

## Quick Reference

| Standard | Value |
|----------|-------|
| **Shell** | Bash 4.0+ |
| **Shebang** | `#!/usr/bin/env bash` |
| **Flags** | `set -eEuo pipefail` |
| **Linter** | shellcheck |

## Required Patterns

Every script MUST start with:

```bash
#!/usr/bin/env bash
set -eEuo pipefail
```

**Flag meanings:**
- `-e` Exit on error
- `-E` ERR trap inherited by functions
- `-u` Exit on undefined variable
- `-o pipefail` Fail on pipe errors

## Variable Quoting

```bash
# ✅ GOOD
namespace="${NAMESPACE:-default}"
kubectl get pods -n "${namespace}"

# ❌ BAD - Unquoted (word splitting risk)
kubectl get pods -n $namespace
```

## Shellcheck

All scripts must pass: `shellcheck scripts/*.sh`

Create `.shellcheckrc` at repo root:
```ini
disable=SC1090
disable=SC1091
disable=SC2312
```

## Error Handling

```bash
on_error() {
    local exit_code=$?
    echo "ERROR: Failed on line $LINENO (exit $exit_code)" >&2
    exit "$exit_code"
}
trap on_error ERR
```

## Logging Functions

```bash
log()  { echo "[$(date '+%H:%M:%S')] $*"; }
warn() { echo "[$(date '+%H:%M:%S')] WARNING: $*" >&2; }
err()  { echo "[$(date '+%H:%M:%S')] ERROR: $*" >&2; }
die()  { err "$*"; exit 1; }
```

## Security

### Never Pass Secrets as CLI Arguments

```bash
# ❌ BAD - Visible in ps aux
kubectl create secret generic s --from-literal=token="$TOKEN"

# ✅ GOOD - Via stdin or file
echo "$TOKEN" | kubectl create secret generic s --from-file=token=/dev/stdin
```

### Use jq for JSON Construction

```bash
# ❌ BAD - Injection risk
json="{\"name\": \"$NAME\"}"

# ✅ GOOD - Proper escaping
json=$(jq -n --arg name "$NAME" '{name: $name}')
```

## Script Template

```bash
#!/usr/bin/env bash
# ═══════════════════════════════════════════════════════════════
# Script: <name>
# Purpose: <description>
# Exit Codes: 0=success, 1=arg error, 2=missing dep
# ═══════════════════════════════════════════════════════════════
set -eEuo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

log()  { echo "[$(date '+%H:%M:%S')] $*"; }
err()  { echo "[$(date '+%H:%M:%S')] ERROR: $*" >&2; }
die()  { err "$*"; exit 1; }

on_error() { err "Failed on line $LINENO"; exit 1; }
trap on_error ERR

check_deps() {
    for cmd in kubectl jq; do
        command -v "$cmd" &>/dev/null || die "Missing: $cmd"
    done
}

main() {
    check_deps
    log "Starting..."
    # Main logic
}

main "$@"
```

## Summary

1. Bash 4.0+ with `set -eEuo pipefail`
2. All scripts must pass shellcheck
3. Quote all variables: `"${var}"`
4. Use logging functions
5. Add ERR trap for debug context
6. Never pass secrets as CLI arguments
7. Use jq for JSON construction
