# Shell Kit

> Shell scripting standards and tooling for AgentOps.

## Install

```bash
/plugin install shell-kit@agentops
```

Requires: `solo-kit`

## What's Included

### Standards

Comprehensive shell scripting standards in `skills/standards/references/shell.md`:
- POSIX compliance for portability
- Error handling with `set -euo pipefail`
- Shellcheck compliance
- Process management
- CI/CD script patterns
- Common anti-patterns to avoid

### Hooks

| Hook | Trigger | What It Does |
|------|---------|--------------|
| `shellcheck` | Edit *.{sh,bash} | Lint with shellcheck |

### Patterns

**Script Template**
```bash
#!/usr/bin/env bash
set -euo pipefail

# Constants
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Functions
log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $*" >&2
}

die() {
    log "ERROR: $*"
    exit 1
}

main() {
    # Your logic here
    log "Starting..."
}

main "$@"
```

**Safe Variable Expansion**
```bash
# Always quote variables
echo "$variable"

# Use defaults
echo "${var:-default}"

# Check if set
if [[ -n "${var:-}" ]]; then
    echo "var is set"
fi
```

**Error Handling**
```bash
# Trap for cleanup
cleanup() {
    rm -f "$tmp_file"
}
trap cleanup EXIT

# Check command success
if ! command -v jq &> /dev/null; then
    die "jq is required but not installed"
fi
```

## Requirements

- Bash 4.0+
- Optional: shellcheck (for hooks)

## License

MIT
