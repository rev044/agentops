#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SCRIPT="$ROOT/scripts/install-dev-hooks.sh"

PASS=0
FAIL=0

pass() { echo "PASS: $1"; PASS=$((PASS + 1)); }
fail() { echo "FAIL: $1"; FAIL=$((FAIL + 1)); }

if [[ ! -f "$SCRIPT" ]]; then
  echo "FAIL: missing script: $SCRIPT" >&2
  exit 1
fi

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

setup_fixture() {
  local fixture="$1"
  mkdir -p "$fixture/scripts" "$fixture/.githooks"
  cp "$SCRIPT" "$fixture/scripts/install-dev-hooks.sh"
  chmod +x "$fixture/scripts/install-dev-hooks.sh"

  cat > "$fixture/.githooks/pre-commit" <<'EOF'
#!/usr/bin/env bash
exit 0
EOF
  cat > "$fixture/.githooks/pre-push" <<'EOF'
#!/usr/bin/env bash
exit 0
EOF
  chmod +x "$fixture/.githooks/pre-commit" "$fixture/.githooks/pre-push"

  git -C "$fixture" init -q
}

test_installs_repo_hooks_path() {
  local fixture="$TMP_DIR/install"

  setup_fixture "$fixture"

  if ! (cd "$fixture" && bash scripts/install-dev-hooks.sh >/dev/null); then
    fail "install should succeed"
    return
  fi

  if [[ "$(git -C "$fixture" config --local --get core.hooksPath)" == ".githooks" ]]; then
    pass "installs repo-managed hooks path"
  else
    fail "core.hooksPath should be .githooks after install"
  fi
}

test_check_mode_passes_after_install() {
  local fixture="$TMP_DIR/check-pass"

  setup_fixture "$fixture"
  (cd "$fixture" && bash scripts/install-dev-hooks.sh >/dev/null)

  if (cd "$fixture" && bash scripts/install-dev-hooks.sh --check >/dev/null); then
    pass "check mode passes when hooks path is configured"
  else
    fail "check mode should pass after install"
  fi
}

test_check_mode_fails_when_hooks_path_missing() {
  local fixture="$TMP_DIR/check-fail"

  setup_fixture "$fixture"

  if (cd "$fixture" && bash scripts/install-dev-hooks.sh --check >/dev/null 2>&1); then
    fail "check mode should fail when hooks path is unset"
  else
    pass "check mode fails when hooks path is unset"
  fi
}

echo "== test-install-dev-hooks =="
test_installs_repo_hooks_path
test_check_mode_passes_after_install
test_check_mode_fails_when_hooks_path_missing

echo ""
echo "Results: $PASS PASS, $FAIL FAIL"
if [[ "$FAIL" -gt 0 ]]; then
  exit 1
fi
exit 0
