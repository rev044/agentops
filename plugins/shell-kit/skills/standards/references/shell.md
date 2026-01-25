# Shell Script Standards - Tier 1 Quick Reference

<!-- Tier 1: Generic standards (~5KB), always loaded -->
<!-- Tier 2: Deep standards in vibe/references/shell-standards.md (~20KB), loaded on --deep -->
<!-- Last synced: 2026-01-21 -->

> **Purpose:** Quick reference for shell scripting standards. For comprehensive patterns, load Tier 2.

---

## Quick Reference

| Standard | Value | Validation |
|----------|-------|------------|
| **Shell** | Bash 4.0+ | `bash --version` |
| **Shebang** | `#!/usr/bin/env bash` | First line |
| **Flags** | `set -eEuo pipefail` | Line 2 or 3 |
| **Linter** | shellcheck | `.shellcheckrc` at root |

---

## Required Preamble

Every shell script MUST start with:

```bash
#!/usr/bin/env bash
set -eEuo pipefail
```

| Flag | Effect |
|------|--------|
| `-e` | Exit on error |
| `-E` | ERR trap inherited by functions |
| `-u` | Exit on undefined variable |
| `-o pipefail` | Fail if any pipe command fails |

---

## Common Errors

| Symptom | Cause | Fix |
|---------|-------|-----|
| `unbound variable` | Unset variable | `${VAR:-default}` |
| `command not found` | Missing dep | Add to check_dependencies() |
| `syntax error` | Missing quote/semicolon | Check quoting |
| `permission denied` | Not executable | `chmod +x script.sh` |
| `bad substitution` | bash-only in sh | Use `#!/usr/bin/env bash` |
| Word splitting | Unquoted variable | Quote: `"${var}"` |

---

## Anti-Patterns

| Name | Pattern | Instead |
|------|---------|---------|
| Parsing ls | `for f in $(ls)` | `for f in *` |
| Cat Abuse | `cat file \| grep` | `grep pattern file` |
| Backticks | `` `cmd` `` | `$(cmd)` |
| No Set Flags | Missing `set -e` | Add `set -eEuo pipefail` |
| Secrets in Args | `--password="x"` | Use stdin or file |
| Hardcoded Paths | `/home/user/file` | Use `$HOME`, variables |
| eval Abuse | `eval "$input"` | Avoid eval |

---

## Essential Patterns

### Logging Functions

```bash
log()  { echo "[$(date '+%H:%M:%S')] $*"; }
warn() { echo "[$(date '+%H:%M:%S')] WARNING: $*" >&2; }
err()  { echo "[$(date '+%H:%M:%S')] ERROR: $*" >&2; }
die()  { err "$*"; exit 1; }
```

### ERR Trap

```bash
on_error() {
    local exit_code=$?
    echo "ERROR: Failed on line $LINENO with exit $exit_code" >&2
    exit "$exit_code"
}
trap on_error ERR
```

### Cleanup Pattern

```bash
TMPDIR=$(mktemp -d)
cleanup() { rm -rf "$TMPDIR"; }
trap cleanup EXIT
```

### Check Dependencies

```bash
check_dependencies() {
    local missing=()
    for cmd in kubectl jq; do
        command -v "$cmd" &>/dev/null || missing+=("$cmd")
    done
    [[ ${#missing[@]} -gt 0 ]] && die "Missing: ${missing[*]}"
}
```

---

## Variable Quoting

```bash
# GOOD - Quoted, safe defaults
namespace="${NAMESPACE:-default}"
kubectl get pods -n "${namespace}"

# BAD - Unquoted
kubectl get pods -n $namespace
```

---

## Security Quick Rules

| Rule | Bad | Good |
|------|-----|------|
| Secrets | `--token="$TOKEN"` | `echo "$TOKEN" \| cmd --token=-` |
| Path traversal | Accept `../` | Validate: `[[ "$path" != *..* ]]` |
| JSON construction | `"{\"k\": \"$v\"}"` | `jq -n --arg v "$v" '{k: $v}'` |
| Sed injection | `sed "s/X/$USER/"` | Escape special chars |

---

## Summary Checklist

| Category | Requirement |
|----------|-------------|
| **Shebang** | `#!/usr/bin/env bash` |
| **Set Flags** | `set -eEuo pipefail` |
| **Linting** | shellcheck passes |
| **Variables** | All quoted |
| **Error Handling** | ERR trap present |
| **Exit Codes** | Documented in header |
| **Secrets** | Never in CLI args |
| **Logging** | Use log/warn/err/die |

---

## Talos Prescan Checks

| Check | Pattern | Rationale |
|-------|---------|-----------|
| PRE-002 | TODO/FIXME markers | Track technical debt |
| PRE-005 | Missing set flags | Script safety |
| PRE-006 | Unquoted variables | Word splitting risk |
| PRE-009 | Secrets in CLI | Security violation |

---

## JIT Loading

**Tier 2 (Deep Standards):** For comprehensive patterns including:
- Full script template structure
- Shellcheck configuration and common fixes
- Security patterns (secrets, injection, validation)
- Common patterns (polling, parallel execution)
- BATS testing framework
- Validation & evidence requirements

Load: `vibe/references/shell-standards.md`
