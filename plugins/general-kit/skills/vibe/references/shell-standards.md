# Shell Script Standards Catalog - Vibe Canonical Reference

**Version:** 1.0.0
**Last Updated:** 2026-01-21
**Purpose:** Canonical shell scripting standards for vibe skill validation

---

## Table of Contents

1. [Required Patterns](#required-patterns)
2. [Shellcheck Integration](#shellcheck-integration)
3. [Error Handling](#error-handling)
4. [Logging Functions](#logging-functions)
5. [Script Organization](#script-organization)
6. [Security](#security)
7. [Common Patterns](#common-patterns)
8. [Anti-Patterns Avoided](#anti-patterns-avoided)
9. [Compliance Assessment](#compliance-assessment)

---

## Required Patterns

### Shebang and Flags

Every shell script MUST start with:

```bash
#!/usr/bin/env bash
set -eEuo pipefail
```

**Flag explanation:**

| Flag | Effect | Failure without |
|------|--------|-----------------|
| `-e` | Exit on error | Silent failures, continued execution |
| `-E` | ERR trap inherited | Traps don't fire in functions |
| `-u` | Exit on undefined | Empty variables cause silent bugs |
| `-o pipefail` | Pipe fails propagate | `false \| true` returns 0 |

### Variable Quoting

Always quote variables to prevent word splitting and globbing:

```bash
# GOOD - Quoted variables, safe defaults
namespace="${NAMESPACE:-default}"
kubectl get pods -n "${namespace}"

# BAD - Unquoted variables (word splitting, globbing risks)
kubectl get pods -n $namespace
```

### Safe Defaults

```bash
# Pattern: ${VAR:-default}
namespace="${NAMESPACE:-default}"
timeout="${TIMEOUT:-300}"

# Pattern: ${VAR:?error message}
api_key="${API_KEY:?API_KEY must be set}"
```

---

## Shellcheck Integration

### Repository Configuration

Create `.shellcheckrc` at repo root:

```ini
# .shellcheckrc
shell=bash
disable=SC1090
disable=SC1091
disable=SC2312
```

### Common Shellcheck Fixes

| Code | Issue | Fix |
|------|-------|-----|
| SC2086 | Word splitting | Quote: `"$var"` |
| SC2164 | cd can fail | `cd /path \|\| exit 1` |
| SC2046 | Word splitting in $() | Quote: `"$(command)"` |
| SC2181 | Checking $? | Use `if command; then` |
| SC2155 | declare/local hides exit | Split: `local x; x=$(cmd)` |

---

## Error Handling

### ERR Trap for Debug Context

```bash
#!/usr/bin/env bash
set -eEuo pipefail

on_error() {
    local exit_code=$?
    local line_no=$1
    echo "ERROR: Script failed on line $line_no with exit code $exit_code" >&2
    echo "Command: ${BASH_COMMAND}" >&2
    exit "$exit_code"
}
trap 'on_error $LINENO' ERR
```

### Cleanup Pattern

```bash
TMPDIR=$(mktemp -d)

cleanup() {
    local exit_code=$?
    rm -rf "$TMPDIR" 2>/dev/null || true
    exit "$exit_code"
}
trap cleanup EXIT
```

---

## Logging Functions

### Standard Logging

```bash
log()  { echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*"; }
warn() { echo "[$(date '+%Y-%m-%d %H:%M:%S')] WARNING: $*" >&2; }
err()  { echo "[$(date '+%Y-%m-%d %H:%M:%S')] ERROR: $*" >&2; }
die()  { err "$*"; exit 1; }

# Debug logging (controlled by variable)
debug() {
    [[ "${DEBUG:-false}" == "true" ]] && echo "[DEBUG] $*" >&2
}
```

---

## Script Organization

### Full Template

```bash
#!/usr/bin/env bash
# ===================================================================
# Script: <name>.sh
# Purpose: <one-line description>
# Usage: ./<script>.sh [args]
#
# Exit Codes:
#   0 - Success
#   1 - Argument error
#   2 - Missing dependency
# ===================================================================

set -eEuo pipefail

# Configuration
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_NAME="$(basename "${BASH_SOURCE[0]}")"

NAMESPACE="${NAMESPACE:-default}"
DRY_RUN="${DRY_RUN:-false}"

# Functions
log()  { echo "[$(date '+%H:%M:%S')] $*"; }
die()  { echo "[ERROR] $*" >&2; exit 1; }

on_error() {
    err "Script failed on line $1"
    exit 1
}
trap 'on_error $LINENO' ERR

usage() {
    cat <<EOF
Usage: $SCRIPT_NAME [options] <required-arg>

Options:
    -h, --help      Show this help message
    -n, --namespace Kubernetes namespace
EOF
}

main() {
    # Parse arguments and main logic
    log "Starting with namespace: $NAMESPACE"
}

main "$@"
```

---

## Security

### Secret Handling

**Never pass secrets as CLI arguments** - they're visible in `ps aux`:

```bash
# BAD - Secrets visible in process list
kubectl create secret generic my-secret --from-literal=token="$TOKEN"

# GOOD - Pass via stdin
kubectl create secret generic my-secret --from-literal=token=- <<< "$TOKEN"

# GOOD - Use file-based approach
echo "$SECRET" > "$TMPDIR/secret"
chmod 600 "$TMPDIR/secret"
kubectl create secret generic my-secret --from-file=token="$TMPDIR/secret"
```

### JSON Construction

```bash
# BAD - String interpolation (injection risk)
json="{\"name\": \"$NAME\", \"value\": \"$VALUE\"}"

# GOOD - Use jq for proper escaping
json=$(jq -n --arg name "$NAME" --arg value "$VALUE" \
    '{name: $name, value: $value}')
```

---

## Common Patterns

### Retry Pattern

```bash
retry() {
    local max_attempts=${1:-3}
    local delay=${2:-5}
    shift 2
    local cmd=("$@")

    local attempt=1
    while [[ $attempt -le $max_attempts ]]; do
        if "${cmd[@]}"; then
            return 0
        fi
        warn "Attempt $attempt/$max_attempts failed, retrying in ${delay}s..."
        sleep "$delay"
        attempt=$((attempt + 1))
    done

    err "All $max_attempts attempts failed"
    return 1
}

# Usage
retry 3 5 kubectl apply -f manifest.yaml
```

### Checking Command Success

```bash
# GOOD - Direct conditional
if kubectl get namespace "$ns" &>/dev/null; then
    echo "Namespace exists"
else
    kubectl create namespace "$ns"
fi
```

---

## Anti-Patterns Avoided

### No Parsing ls Output

```bash
# Bad
for f in $(ls); do echo "$f"; done

# Good
for f in *; do
    [[ -e "$f" ]] || continue
    echo "$f"
done
```

### No Useless Cat

```bash
# Bad
cat file.txt | grep pattern

# Good
grep pattern file.txt
```

### No Backticks

```bash
# Bad
result=`command`

# Good
result=$(command)
```

---

## Compliance Assessment

**Use letter grades + evidence, NOT numeric scores.**

### Grading Scale

| Grade | Criteria |
|-------|----------|
| A+ | 0 shellcheck errors, set flags, ERR trap, 0 security issues |
| A | <5 shellcheck warnings, set flags, ERR trap, quoted vars |
| A- | <15 shellcheck warnings, set flags, mostly quoted |
| B+ | <30 shellcheck warnings, set flags present |
| B | <50 shellcheck warnings, some flags |
| C | Significant safety issues |
| D | Not production-ready |
| F | Critical issues |

---

## Additional Resources

- [Bash Manual](https://www.gnu.org/software/bash/manual/)
- [ShellCheck Wiki](https://www.shellcheck.net/wiki/)
- [Google Shell Style Guide](https://google.github.io/styleguide/shellguide.html)
