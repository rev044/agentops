#!/usr/bin/env bash
# post-merge-check.sh — validate integration after merging parallel worktree results
# Run from repo root after copying worktree changes into the main tree.
#
# Checks: duplicate function declarations, go build, go vet, go test.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ERRORS=0

# 1. Check for duplicate function declarations in Go test files
echo "=== Checking for duplicate test functions ==="
DUPES=$(grep -rn "^func " "$REPO_ROOT"/cli/cmd/ao/*_test.go 2>/dev/null \
  | sed 's/.*:func //' | sed 's/(.*//' | sort | uniq -d) || true
if [ -n "$DUPES" ]; then
    echo "FAIL: Duplicate function declarations found:"
    echo "$DUPES"
    ERRORS=$((ERRORS + 1))
else
    echo "PASS: No duplicate functions"
fi

# 2. go build
echo "=== Building ==="
if (cd "$REPO_ROOT/cli" && go build ./cmd/ao/...); then
    echo "PASS: Build succeeded"
else
    echo "FAIL: Build failed"
    ERRORS=$((ERRORS + 1))
fi

# 3. go vet
echo "=== Running go vet ==="
if (cd "$REPO_ROOT/cli" && go vet ./cmd/ao/...); then
    echo "PASS: go vet clean"
else
    echo "FAIL: go vet found issues"
    ERRORS=$((ERRORS + 1))
fi

# 4. Run tests
echo "=== Running tests ==="
if (cd "$REPO_ROOT/cli" && go test ./cmd/ao/... -count=1 -short); then
    echo "PASS: Tests passed"
else
    echo "FAIL: Tests failed"
    ERRORS=$((ERRORS + 1))
fi

# 5. go mod tidy (non-blocking WARN)
echo "=== Checking go mod tidy ==="
(cd "$REPO_ROOT/cli" && go mod tidy) 2>/dev/null || true
TIDY_DIRTY=$(git -C "$REPO_ROOT" diff --name-only cli/go.mod cli/go.sum 2>/dev/null || true)
if [ -n "$TIDY_DIRTY" ]; then
    echo "WARN: go mod tidy changed go.mod/go.sum after merge — commit the changes"
else
    echo "PASS: go.mod/go.sum are tidy"
fi

# 6. Symlink check (blocking ERROR)
echo "=== Checking for symlinks ==="
SYMLINKS=$(find "$REPO_ROOT" -type l -not -path '*/.git/*' 2>/dev/null || true)
if [ -n "$SYMLINKS" ]; then
    echo "ERROR: symlinks found after merge (CI plugin-load-test will reject): $SYMLINKS"
    exit 1
fi
echo "PASS: No symlinks found"

# 7. Summary
echo "=== Integration Check Summary ==="
if [ "$ERRORS" -eq 0 ]; then
    echo "All checks passed."
else
    echo "Errors: $ERRORS"
fi
exit "$ERRORS"
