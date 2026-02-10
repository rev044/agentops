#!/bin/bash
# Test: Codex review integration (codex review --uncommitted)
# Proves codex review runs and produces output in a controlled repo
# ag-3b7.4
set -euo pipefail

OUTPUT="/tmp/codex-review-output-$RANDOM.md"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m'

passed=0
failed=0
skipped=0

pass() { echo -e "${GREEN}  ✓${NC} $1"; ((passed++)) || true; }
fail() { echo -e "${RED}  ✗${NC} $1"; ((failed++)) || true; }
skip() { echo -e "${YELLOW}  ⊘${NC} $1"; ((skipped++)) || true; }

TEST_DIR=""
cleanup() {
    rm -f "$OUTPUT"
    [[ -n "$TEST_DIR" && -d "$TEST_DIR" ]] && rm -rf "$TEST_DIR"
}
trap cleanup EXIT

echo -e "${BLUE}[TEST]${NC} Codex review integration (codex review --uncommitted)"

# Pre-flight: Codex CLI available?
if ! command -v codex > /dev/null 2>&1; then
    skip "Codex CLI not found — skipping all tests"
    echo -e "${YELLOW}SKIPPED${NC} - Codex CLI not available"
    exit 0
fi
pass "Codex CLI found"

# Setup: Create temp git repo with uncommitted changes
TEST_DIR=$(mktemp -d)
cd "$TEST_DIR"
git init -q
git config user.email "test@test.com"
git config user.name "Test"

# Create initial commit
cat > hello.py << 'PYEOF'
def greet(name):
    print(f"Hello, {name}!")

if __name__ == "__main__":
    greet("world")
PYEOF
git add hello.py
git commit -q -m "Initial commit"

# Stage a change (uncommitted)
cat > hello.py << 'PYEOF'
def greet(name):
    if not name:
        raise ValueError("Name cannot be empty")
    print(f"Hello, {name}!")

def farewell(name):
    print(f"Goodbye, {name}!")

if __name__ == "__main__":
    greet("world")
    farewell("world")
PYEOF
git add hello.py
pass "Test repo created with staged changes"

# Test 1: codex review --uncommitted runs and produces output
# Note: codex review outputs review to stderr (interactive tool), capture both channels
echo -e "${BLUE}  Running codex review --uncommitted (up to 120s)...${NC}"
if timeout 120 codex review --uncommitted > "$OUTPUT" 2>&1; then
    pass "codex review --uncommitted succeeded (exit 0)"
else
    EXIT_CODE=$?
    if [[ $EXIT_CODE -eq 124 ]]; then
        fail "codex review timed out after 120s"
    else
        fail "codex review --uncommitted failed (exit $EXIT_CODE)"
    fi
fi

# Test 2: Output has content
if [[ -s "$OUTPUT" ]]; then
    SIZE=$(wc -c < "$OUTPUT" | tr -d ' ')
    pass "Output has content (${SIZE} bytes)"
else
    fail "Output empty or missing"
fi

# Summary
echo ""
echo -e "${BLUE}═══════════════════════════════════════════${NC}"
if [[ $failed -gt 0 ]]; then
    echo -e "${RED}FAILED${NC} - $passed passed, $failed failed, $skipped skipped"
    exit 1
else
    echo -e "${GREEN}PASSED${NC} - $passed passed, $failed failed, $skipped skipped"
    exit 0
fi
