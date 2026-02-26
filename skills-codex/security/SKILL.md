---
name: security
description: 'Continuous repository security scanning and release gating. Triggers: "security scan", "security audit", "pre-release security", "run scanners", "check vulnerabilities".'
---


# Security Skill

> **Purpose:** Run repeatable security checks across code, scripts, hooks, and release gates.

Use this skill when you need deterministic security validation before merge/release, or recurring scheduled checks.

## Quick Start

```bash
$security                      # quick security gate
$security --full               # full gate with test-inclusive toolchain checks
$security --release            # full gate for release readiness
$security --json               # machine-readable report output
```

## Execution Contract

### 1) Pre-PR (fast)

Run quick gate:

```bash
scripts/security-gate.sh --mode quick
```

Expected behavior:
- Fails on high/critical findings from available scanners.
- Writes artifacts under `$TMPDIR/agentops-security/<run-id>/`.

### 2) Pre-Release (strict)

Run full gate:

```bash
scripts/security-gate.sh --mode full
```

Expected behavior:
- Full scanner pass before release workflow can continue.
- Artifacts retained for audit and incident response.

### 3) Nightly (continuous)

Nightly workflow should run:

```bash
scripts/security-gate.sh --mode full
```

Expected behavior:
- Detects drift/regressions outside active PR windows.
- Failing run creates actionable signal in workflow summary/issues.

## Triage Guidance

When gate fails:
1. Open latest artifact in `$TMPDIR/agentops-security/` and identify scanner + file.
2. Classify severity (critical/high/medium).
3. Fix immediately for critical/high or create tracked follow-up issue with owner.
4. Re-run `scripts/security-gate.sh` until gate passes.

## Reporting Template

```markdown
Security gate run: <run-id>
Mode: <quick|full>
Result: <pass|blocked>
Top findings:
- <scanner> <severity> <file> <summary>
Actions:
- <fix or issue id>
```

## Notes

- Use this as the canonical security runbook instead of ad-hoc scanner commands.
- Keep workflow wiring aligned with this contract in:
  - `.github/workflows/validate.yml`
  - `.github/workflows/nightly.yml`
  - `.github/workflows/release.yml`
- For binary/internal black-box assurance (static + dynamic + baseline + policy), use:
  - `skills/security-suite/SKILL.md` (includes `security_suite.py` in its scripts dir)

## Examples

### Scenario: Quick Security Gate Before Opening a PR

**User says:** `$security`

**What happens:**
1. The skill runs `scripts/security-gate.sh --mode quick`, which executes available scanners (semgrep, gosec, gitleaks) against the current working tree and flags high/critical findings.
2. Scan artifacts are written to `$TMPDIR/agentops-security/<run-id>/` for review, and the gate reports a pass/blocked verdict.

**Result:** The gate passes with no high/critical findings, confirming the branch is safe to open a PR.

### Scenario: Full Security Gate for a Release

**User says:** `$security --release`

**What happens:**
1. The skill runs `scripts/security-gate.sh --mode full`, which performs a comprehensive scan including all scanner passes, test-inclusive toolchain checks, and stricter severity thresholds.
2. Artifacts are retained under `$TMPDIR/agentops-security/<run-id>/` for audit trail and incident response, and a structured report is generated.

**Result:** The full gate blocks the release on two medium-severity findings in `cli/internal/config.go`; the operator triages and fixes them before re-running the gate to get a clean pass.

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| Gate reports "scanner not found" and skips checks | Required scanner (semgrep, gosec, or gitleaks) is not installed | Install the missing scanner: `brew install semgrep`, `go install github.com/securego/gosec/v2/cmd/gosec@latest`, or `brew install gitleaks`. |
| Gate passes locally but fails in CI | CI environment has additional scanners or stricter config | Compare `$TMPDIR/agentops-security/` artifacts from both environments; align scanner versions and config files across local and CI. |
| False positive blocking the gate | Scanner flags a non-issue as high/critical severity | Add a scanner-specific inline suppression comment (e.g., `# nosemgrep: rule-id`) or update the scanner config to exclude the pattern, then document the suppression reason. |
| Artifacts directory `$TMPDIR/agentops-security/` not created | Script lacks write permissions or `$TMPDIR` is not writable | Verify `$TMPDIR` is set and writable; the script auto-creates subdirectories on each run. |
| Nightly scan not detecting regressions | Nightly workflow is not configured or is pointing at stale branch | Verify `.github/workflows/nightly.yml` runs `scripts/security-gate.sh --mode full` against the correct branch (typically `main`). |

---

## Scripts

### security-gate.sh

```bash
#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

MODE="quick"
JSON_OUTPUT=false
REQUIRE_TOOLS=false

usage() {
  cat <<'USAGE'
Usage: scripts/security-gate.sh [--mode quick|full] [--json] [--require-tools]

Runs the unified security gate using scripts/toolchain-validate.sh.

Options:
  --mode quick|full   quick = skip slow tests (default), full = full suite
  --json              output machine-readable summary JSON
  --require-tools     fail if any scanner reports not_installed/error
  -h, --help          show this help
USAGE
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --mode)
      MODE="${2:-}"
      shift 2
      ;;
    --json)
      JSON_OUTPUT=true
      shift
      ;;
    --require-tools)
      REQUIRE_TOOLS=true
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

if [[ "$MODE" != "quick" && "$MODE" != "full" ]]; then
  echo "Invalid mode: $MODE (expected quick or full)" >&2
  exit 1
fi

# Canonical scanner invocation contract: scripts/toolchain-validate.sh --gate
TOOLCHAIN_SCRIPT="${SECURITY_GATE_TOOLCHAIN_SCRIPT:-scripts/toolchain-validate.sh}"
if [[ ! -x "$TOOLCHAIN_SCRIPT" ]]; then
  echo "Missing executable: $TOOLCHAIN_SCRIPT" >&2
  exit 1
fi

RUN_ID="$(date -u +%Y%m%dT%H%M%SZ)-${MODE}"
SECURITY_BASE="${SECURITY_GATE_OUTPUT_DIR:-${TMPDIR:-/tmp}/agentops-security}"
SECURITY_DIR="$SECURITY_BASE/$RUN_ID"
mkdir -p "$SECURITY_DIR"

TOOLCHAIN_ARGS=(--gate --json)
if [[ "$MODE" == "quick" ]]; then
  TOOLCHAIN_ARGS=(--quick --gate --json)
fi

set +e
TOOLCHAIN_OUTPUT="$($TOOLCHAIN_SCRIPT "${TOOLCHAIN_ARGS[@]}" 2>&1)"
TOOLCHAIN_EXIT=$?
set -e

SUMMARY_JSON="$SECURITY_DIR/summary.json"
printf '%s\n' "$TOOLCHAIN_OUTPUT" > "$SUMMARY_JSON"

TOOLING_SRC="${TOOLCHAIN_OUTPUT_DIR:-${TMPDIR:-/tmp}/agentops-tooling}"
if [[ -d "$TOOLING_SRC" ]]; then
  cp -a "$TOOLING_SRC/." "$SECURITY_DIR/" 2>/dev/null || true
fi

if command -v jq >/dev/null 2>&1 && jq empty "$SUMMARY_JSON" >/dev/null 2>&1; then
  GATE_STATUS="$(jq -r '.gate_status // "UNKNOWN"' "$SUMMARY_JSON")"
  MISSING_TOOLS="$(jq -r '[.tools[] | select(. == "not_installed" or . == "error")] | length' "$SUMMARY_JSON")"

  EXTENDED_JSON="$SECURITY_DIR/security-gate-summary.json"
  jq -n \
    --arg mode "$MODE" \
    --arg run_id "$RUN_ID" \
    --arg output_dir "$SECURITY_DIR" \
    --argjson toolchain "$(cat "$SUMMARY_JSON")" \
    --arg gate_status "$GATE_STATUS" \
    --argjson missing_tools "$MISSING_TOOLS" \
    --arg require_tools "$REQUIRE_TOOLS" \
    '{
      mode: $mode,
      run_id: $run_id,
      output_dir: $output_dir,
      gate_status: $gate_status,
      missing_tool_count: $missing_tools,
      require_tools: ($require_tools == "true"),
      toolchain: $toolchain
    }' > "$EXTENDED_JSON"

  if [[ "$REQUIRE_TOOLS" == "true" && "$MISSING_TOOLS" -gt 0 ]]; then
    if [[ "$JSON_OUTPUT" == "true" ]]; then
      cat "$EXTENDED_JSON"
    else
      echo "Security gate FAILED: missing/error tools detected ($MISSING_TOOLS)"
      echo "Report: $EXTENDED_JSON"
    fi
    exit 4
  fi

  if [[ "$JSON_OUTPUT" == "true" ]]; then
    cat "$EXTENDED_JSON"
  else
    echo "Security gate mode: $MODE"
    echo "Gate status: $GATE_STATUS"
    echo "Missing/error tools: $MISSING_TOOLS"
    echo "Report: $EXTENDED_JSON"
  fi
else
  if [[ "$JSON_OUTPUT" == "true" ]]; then
    jq -n \
      --arg mode "$MODE" \
      --arg run_id "$RUN_ID" \
      --arg output_dir "$SECURITY_DIR" \
      --arg raw "$TOOLCHAIN_OUTPUT" \
      '{mode: $mode, run_id: $run_id, output_dir: $output_dir, parse_error: true, raw_output: $raw}'
  else
    echo "Security gate warning: toolchain output was not valid JSON"
    echo "Raw output saved to: $SUMMARY_JSON"
  fi
  exit 1
fi

# Preserve toolchain gate semantics for findings.
if [[ "$TOOLCHAIN_EXIT" -ne 0 ]]; then
  exit "$TOOLCHAIN_EXIT"
fi

exit 0
```

### validate.sh

```bash
#!/usr/bin/env bash
set -euo pipefail
SKILL_DIR="$(cd "$(dirname "$0")/.." && pwd)"
PASS=0; FAIL=0

check() { if bash -c "$2"; then echo "PASS: $1"; PASS=$((PASS + 1)); else echo "FAIL: $1"; FAIL=$((FAIL + 1)); fi; }

check "SKILL.md exists" "[ -f '$SKILL_DIR/SKILL.md' ]"
check "SKILL.md has YAML frontmatter" "head -1 '$SKILL_DIR/SKILL.md' | grep -q '^---$'"

echo ""; echo "Results: $PASS passed, $FAIL failed"
[ $FAIL -eq 0 ] && exit 0 || exit 1
```


