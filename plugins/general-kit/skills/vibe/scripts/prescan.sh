#!/usr/bin/env bash
# Vibe Pre-Scan: Fast static detection for 9 failure patterns (6 language-agnostic + 3 Go-specific)
# Usage: prescan.sh <target>
#   target: recent | all | <directory> | <file>

set -euo pipefail

TARGET="${1:-recent}"

# File filtering (exclude generated code, build artifacts, test fixtures)
filter_files() {
  grep -v '__pycache__\|\.venv\|venv/\|node_modules\|\.git/\|test_fixtures\|/fixtures/\|\.eggs\|egg-info\|/dist/\|/build/\|\.tox\|\.mypy_cache\|\.pytest_cache' \
  | grep -v '\.gen\.go$\|zz_generated\|_generated\.go$\|\.pb\.go$\|mock_.*\.go$\|/generated/\|/gen/\|deepcopy'
}

# Resolve target to file lists (Python, Go, Bash)
case "$TARGET" in
  recent)
    PY_FILES=$(git diff --name-only HEAD~1 HEAD 2>/dev/null | grep '\.py$' | filter_files || true)
    GO_FILES=$(git diff --name-only HEAD~1 HEAD 2>/dev/null | grep '\.go$' | filter_files || true)
    SH_FILES=$(git diff --name-only HEAD~1 HEAD 2>/dev/null | grep '\.sh$' | filter_files || true)
    MODE="Recent"
    ;;
  all)
    PY_FILES=$(find . -name "*.py" -type f 2>/dev/null | filter_files | grep -v 'test_' || true)
    GO_FILES=$(find . -name "*.go" -type f 2>/dev/null | filter_files | grep -v '_test\.go$' || true)
    SH_FILES=$(find . -name "*.sh" -type f 2>/dev/null | filter_files || true)
    MODE="All"
    ;;
  *)
    if [ -d "$TARGET" ]; then
      PY_FILES=$(find "$TARGET" -name "*.py" -type f 2>/dev/null | filter_files || true)
      GO_FILES=$(find "$TARGET" -name "*.go" -type f 2>/dev/null | filter_files || true)
      SH_FILES=$(find "$TARGET" -name "*.sh" -type f 2>/dev/null | filter_files || true)
      MODE="Dir"
    elif [ -f "$TARGET" ]; then
      case "$TARGET" in
        *.py) PY_FILES="$TARGET"; GO_FILES=""; SH_FILES="" ;;
        *.go) GO_FILES="$TARGET"; PY_FILES=""; SH_FILES="" ;;
        *.sh) SH_FILES="$TARGET"; PY_FILES=""; GO_FILES="" ;;
        *) PY_FILES="$TARGET"; GO_FILES=""; SH_FILES="" ;;
      esac
      MODE="File"
    else
      echo "ERROR: Target not found: $TARGET" >&2
      exit 1
    fi
    ;;
esac

# Combine for backwards compatibility
FILES="$PY_FILES"
[ -n "$GO_FILES" ] && FILES=$(printf "%s\n%s" "$FILES" "$GO_FILES")
[ -n "$SH_FILES" ] && FILES=$(printf "%s\n%s" "$FILES" "$SH_FILES")

# Count files (handle empty strings properly)
count_lines() {
  local input="$1"
  [ -z "$input" ] && echo 0 && return
  echo "$input" | wc -l | tr -d ' '
}
PY_COUNT=$(count_lines "$PY_FILES")
GO_COUNT=$(count_lines "$GO_FILES")
SH_COUNT=$(count_lines "$SH_FILES")
FILE_COUNT=$((PY_COUNT + GO_COUNT + SH_COUNT))
if [ "$FILE_COUNT" -eq 0 ]; then
  echo "No files found for target: $TARGET"
  exit 0
fi

echo "Pre-Scan Target: $TARGET"
echo "Mode: $MODE | Files: $FILE_COUNT (py:$PY_COUNT go:$GO_COUNT sh:$SH_COUNT)"
echo ""

# Initialize counters
P1_COUNT=0
P4_COUNT=0
P5_COUNT=0
P8_COUNT=0
P9_COUNT=0
P12_COUNT=0
P13_COUNT=0
P14_COUNT=0
P15_COUNT=0
P16_COUNT=0
P17_COUNT=0
P18_COUNT=0
P19_COUNT=0
P20_COUNT=0
P11_COUNT=0

# P1: Phantom Modifications (CRITICAL)
# Committed lines not in current file
echo "[P1] Phantom Modifications"
if [ "$TARGET" = "recent" ]; then
  for file in $FILES; do
    [ -f "$file" ] || continue
    while IFS= read -r line; do
      clean=$(echo "$line" | sed 's/^+//' | xargs)
      if [ ${#clean} -gt 10 ] && ! grep -qF "$clean" "$file" 2>/dev/null; then
        echo "  - $file: Committed line missing: \"${clean:0:50}...\""
        P1_COUNT=$((P1_COUNT + 1))
      fi
    done < <(git show HEAD -- "$file" 2>/dev/null | grep '^+[^+]' || true)
  done
fi
echo "  $P1_COUNT findings"

# P4: Invisible Undone (HIGH)
# TODO markers, commented-out code
echo ""
echo "[P4] Invisible Undone"
for file in $FILES; do
  [ -f "$file" ] || continue
  # TODO/FIXME markers
  while IFS= read -r match; do
    line_num=$(echo "$match" | cut -d: -f1)
    echo "  - $file:$line_num: TODO marker"
    P4_COUNT=$((P4_COUNT + 1))
  done < <(grep -n "TODO\|FIXME\|XXX\|HACK" "$file" 2>/dev/null | head -3 || true)
  # Commented code
  while IFS= read -r match; do
    line_num=$(echo "$match" | cut -d: -f1)
    echo "  - $file:$line_num: Commented code"
    P4_COUNT=$((P4_COUNT + 1))
  done < <(grep -n "^\s*#\s*\(def \|class \|if \|for \)" "$file" 2>/dev/null | head -2 || true)
done
echo "  $P4_COUNT findings"

# P5: Eldritch Horror (HIGH)
# Complexity CC > 15 or function > 50 lines
echo ""
echo "[P5] Eldritch Horror"

# Python: radon for cyclomatic complexity
if [ -n "$PY_FILES" ]; then
  if command -v radon &>/dev/null; then
    for file in $PY_FILES; do
      [ -f "$file" ] || continue
      while IFS= read -r line; do
        cc=$(echo "$line" | grep -oE '\([0-9]+\)' | tr -d '()')
        if [ -n "$cc" ] && [ "$cc" -gt 15 ]; then
          func=$(echo "$line" | awk '{print $3}')
          echo "  - $file: $func CC=$cc (py)"
          P5_COUNT=$((P5_COUNT + 1))
        fi
      done < <(radon cc "$file" -s -n E 2>/dev/null | grep -E "^\s*[EF]\s+[0-9]+" || true)
    done
  else
    echo "  WARNING: radon not installed (Python CC skipped)"
  fi
fi

# Go: gocyclo for cyclomatic complexity
if [ -n "$GO_FILES" ]; then
  if command -v gocyclo &>/dev/null; then
    for file in $GO_FILES; do
      [ -f "$file" ] || continue
      while IFS= read -r line; do
        # gocyclo output: "15 pkg funcName file.go:42:1"
        cc=$(echo "$line" | awk '{print $1}')
        func=$(echo "$line" | awk '{print $3}')
        loc=$(echo "$line" | awk '{print $4}')
        if [ -n "$cc" ] && [ "$cc" -gt 15 ]; then
          echo "  - $loc: $func CC=$cc (go)"
          P5_COUNT=$((P5_COUNT + 1))
        fi
      done < <(gocyclo -over 15 "$file" 2>/dev/null || true)
    done
  else
    echo "  WARNING: gocyclo not installed (Go CC skipped)"
  fi
fi

# Python: Function length > 50 lines
for file in $PY_FILES; do
  [ -f "$file" ] || continue
  python3 -c "
import ast
try:
    with open('$file') as f: tree = ast.parse(f.read())
    for n in ast.walk(tree):
        if isinstance(n, (ast.FunctionDef, ast.AsyncFunctionDef)) and hasattr(n, 'end_lineno'):
            lines = n.end_lineno - n.lineno + 1
            if lines > 50: print(f'  - $file:{n.lineno}: {n.name}() is {lines} lines (py)')
except: pass
" 2>/dev/null || true
done

# Go: Function length > 50 lines (simple heuristic)
for file in $GO_FILES; do
  [ -f "$file" ] || continue
  awk '
    /^func / { fname=$0; start=NR; in_func=1 }
    in_func && /^}$/ {
      lines = NR - start + 1
      if (lines > 50) {
        # Extract function name
        match(fname, /func[[:space:]]+(\([^)]+\)[[:space:]]+)?([a-zA-Z_][a-zA-Z0-9_]*)/, arr)
        print "  - '"$file"':" start ": " arr[2] "() is " lines " lines (go)"
      }
      in_func=0
    }
  ' "$file" 2>/dev/null || true
done
echo "  $P5_COUNT findings"

# P8: Cargo Cult Error Handling (HIGH)
# Empty except, pass-only handlers, bare except
echo ""
echo "[P8] Cargo Cult Error Handling"

# Python: except:pass, bare except
for file in $PY_FILES; do
  [ -f "$file" ] || continue
  python3 -c "
import ast
try:
    with open('$file') as f: tree = ast.parse(f.read())
    for n in ast.walk(tree):
        if isinstance(n, ast.Try):
            for h in n.handlers:
                if len(h.body) == 1 and isinstance(h.body[0], ast.Pass):
                    print(f'  - $file:{h.lineno}: except: pass (swallowed) (py)')
                if h.type is None:
                    print(f'  - $file:{h.lineno}: bare except (catches SystemExit) (py)')
except: pass
" 2>/dev/null || true
done

# Bash: shellcheck for error handling issues
if [ -n "$SH_FILES" ]; then
  if command -v shellcheck &>/dev/null; then
    for file in $SH_FILES; do
      [ -f "$file" ] || continue
      # SC2181: Check exit code directly, not via $?
      # SC2086: Double quote to prevent globbing/splitting
      # SC2046: Quote to prevent word splitting
      # SC2155: Declare and assign separately to avoid masking return values
      while IFS= read -r line; do
        echo "  - $line (sh)"
        P8_COUNT=$((P8_COUNT + 1))
      done < <(shellcheck -f gcc -S warning "$file" 2>/dev/null | head -5 || true)
    done
  else
    echo "  WARNING: shellcheck not installed (Bash checks skipped)"
  fi
fi
echo "  $P8_COUNT findings"

# P9: Documentation Phantom (MEDIUM)
# Docstrings claiming behavior not implemented (Python only)
echo ""
echo "[P9] Documentation Phantom"
for file in $PY_FILES; do
  [ -f "$file" ] || continue
  python3 -c "
import ast, re
try:
    with open('$file') as f: src = f.read()
    tree = ast.parse(src)
    PATTERNS = [
        (r'\bvalidates?\b', ['raise', 'ValueError', 'return False']),
        (r'\bensures?\b', ['assert', 'raise']),
        (r'\bencrypts?\b', ['crypto', 'cipher']),
        (r'\bauthenticat', ['token', 'password']),
        (r'\bsanitiz', ['escape', 'strip'])
    ]
    for n in ast.walk(tree):
        if isinstance(n, (ast.FunctionDef, ast.AsyncFunctionDef)):
            if n.body and isinstance(n.body[0], ast.Expr) and isinstance(getattr(n.body[0], 'value', None), ast.Constant):
                doc = str(n.body[0].value.value).lower()
                fsrc = (ast.get_source_segment(src, n) or '').lower()
                for pat, impl in PATTERNS:
                    if re.search(pat, doc) and not any(i in fsrc for i in impl):
                        print(f'  - $file:{n.lineno}: {n.name}() docstring mismatch')
                        break
except: pass
" 2>/dev/null || true
done
echo "  $P9_COUNT findings"

# P12: Zombie Code (MEDIUM)
# Unused functions, unreachable code after return (Python only)
echo ""
echo "[P12] Zombie Code"
for file in $PY_FILES; do
  [ -f "$file" ] || continue
  python3 -c "
import ast
try:
    with open('$file') as f: src = f.read()
    tree = ast.parse(src)
    defined, called = set(), set()
    for n in ast.walk(tree):
        if isinstance(n, (ast.FunctionDef, ast.AsyncFunctionDef)) and not n.name.startswith('_'):
            defined.add(n.name)
        if isinstance(n, ast.Call):
            if isinstance(n.func, ast.Name): called.add(n.func.id)
            elif isinstance(n.func, ast.Attribute): called.add(n.func.attr)
    for f in (defined - called):
        if f not in ('main', 'setup', 'teardown') and not f.startswith('test_'):
            print(f'  - $file: {f}() may be unused')
    # Unreachable code
    for n in ast.walk(tree):
        if isinstance(n, (ast.FunctionDef, ast.AsyncFunctionDef)):
            for i, s in enumerate(n.body[:-1]):
                if isinstance(s, (ast.Return, ast.Raise)) and n.body[i+1:]:
                    nxt = n.body[i+1]
                    if not (isinstance(nxt, ast.Expr) and isinstance(getattr(nxt, 'value', None), ast.Constant)):
                        print(f'  - $file:{nxt.lineno}: Unreachable after return/raise')
except: pass
" 2>/dev/null || true
done
echo "  $P12_COUNT findings"

# P13: Undocumented Error Ignores (HIGH)
# Go: _ = without nolint:errcheck comment (intentional ignores must be documented)
echo ""
echo "[P13] Undocumented Error Ignores (Go)"
for file in $GO_FILES; do
  [ -f "$file" ] || continue
  while IFS=: read -r linenum content; do
    # Check if there's a nolint:errcheck comment on the same line
    if ! echo "$content" | grep -q "nolint:errcheck"; then
      # Check line before
      if [ "$linenum" -gt 1 ]; then
        prev_line=$((linenum - 1))
        prev_content=$(sed -n "${prev_line}p" "$file" 2>/dev/null)
        if ! echo "$prev_content" | grep -q "nolint:errcheck"; then
          echo "  - $file:$linenum: Error ignored without documentation"
          P13_COUNT=$((P13_COUNT + 1))
        fi
      else
        echo "  - $file:$linenum: Error ignored without documentation"
        P13_COUNT=$((P13_COUNT + 1))
      fi
    fi
  done < <(grep -n "_ =" "$file" 2>/dev/null)
done
echo "  $P13_COUNT findings"

# P14: Error Wrapping with %v (MEDIUM)
# Go: fmt.Errorf with %v instead of %w (breaks error chains)
echo ""
echo "[P14] Error Wrapping with %v (Go)"
for file in $GO_FILES; do
  [ -f "$file" ] || continue
  while IFS=: read -r linenum content; do
    # Only flag if wrapping an error (common pattern: err variable in format args)
    if echo "$content" | grep -qE 'fmt\.Errorf\([^,]*%v[^,]*,.*\berr\b'; then
      echo "  - $file:$linenum: Use %w instead of %v for error wrapping"
      P14_COUNT=$((P14_COUNT + 1))
    fi
  done < <(grep -n 'fmt\.Errorf.*%v' "$file" 2>/dev/null)
done
echo "  $P14_COUNT findings"

# P15: golangci-lint Violations (HIGH)
# Go: Static analysis via golangci-lint (if available)
echo ""
echo "[P15] golangci-lint Violations (Go)"
if command -v golangci-lint >/dev/null 2>&1 && [ -n "$GO_FILES" ]; then
  # Run golangci-lint with JSON output
  LINT_OUTPUT=$(golangci-lint run --out-format=json --issues-exit-code=0 2>/dev/null || true)
  if [ -n "$LINT_OUTPUT" ] && [ "$LINT_OUTPUT" != "null" ]; then
    # Parse and display issues (limit to 10 most severe)
    echo "$LINT_OUTPUT" | jq -r '.Issues[]? | "  - \(.Pos.Filename):\(.Pos.Line): [\(.FromLinter)] \(.Text)"' 2>/dev/null | head -10 || true
    P15_COUNT=$(echo "$LINT_OUTPUT" | jq '.Issues | length' 2>/dev/null || echo 0)
  fi
else
  [ -z "$GO_FILES" ] && echo "  No Go files to check" || echo "  WARNING: golangci-lint not installed (Go linting skipped)"
fi
echo "  $P15_COUNT findings"

# P11: Shellcheck Error/Warning Violations (HIGH)
# Shell scripts with shellcheck errors or warnings (full violations, not just samples)
echo ""
echo "[P11] Shellcheck Violations (Shell)"
if [ -n "$SH_FILES" ]; then
  if command -v shellcheck &>/dev/null; then
    for file in $SH_FILES; do
      [ -f "$file" ] || continue
      # Get error-level violations (more severe than P8's warning check)
      while IFS= read -r line; do
        code=$(echo "$line" | jq -r '.code' 2>/dev/null)
        msg=$(echo "$line" | jq -r '.message' 2>/dev/null | head -c 60)
        line_num=$(echo "$line" | jq -r '.line' 2>/dev/null)
        level=$(echo "$line" | jq -r '.level' 2>/dev/null)
        if [ "$level" = "error" ]; then
          echo "  - $file:$line_num: SC$code - $msg (error)"
          P11_COUNT=$((P11_COUNT + 1))
        fi
      done < <(shellcheck -f json "$file" 2>/dev/null | jq -c '.[]' 2>/dev/null || true)
    done
  else
    echo "  WARNING: shellcheck not installed"
  fi
else
  echo "  No shell files to check"
fi
echo "  $P11_COUNT findings"

# P16: Catch-All Pattern Precedence (HIGH)
# Catch-all patterns (.*) evaluated before specific patterns in dicts/maps
echo ""
echo "[P16] Catch-All Pattern Precedence"
# Check Python files for pattern dicts where catch-all comes early
for file in $PY_FILES; do
  [ -f "$file" ] || continue
  # Look for dicts with regex patterns where ".*" appears before specific patterns
  python3 -c "
import ast
import re
try:
    with open('$file') as f: src = f.read()
    tree = ast.parse(src)
    for node in ast.walk(tree):
        if isinstance(node, ast.Dict):
            keys = []
            for k in node.keys:
                if isinstance(k, ast.Constant) and isinstance(k.value, str):
                    keys.append((k.value, k.lineno))
            # Check if a catch-all pattern appears before specific patterns
            for i, (key, lineno) in enumerate(keys):
                if key in ('.*', '*', r'.*', 'default', '_'):
                    # Check if there are more specific patterns after this
                    remaining = [k for k, _ in keys[i+1:] if k not in ('.*', '*', r'.*', 'default', '_')]
                    if remaining:
                        print(f'  - \$file:{lineno}: Catch-all pattern \"{key}\" before specific patterns')
except: pass
" 2>/dev/null || true
done
# Check shell scripts for associative arrays with catch-all patterns
for file in $SH_FILES; do
  [ -f "$file" ] || continue
  # Look for case statements where *) appears before specific patterns
  while IFS=: read -r line_num content; do
    if echo "$content" | grep -qE '^\s*\*\)'; then
      # Check if this is not the last pattern in the case block
      next_pattern=$(sed -n "$((line_num + 1)),\$p" "$file" 2>/dev/null | grep -n -m1 ')$' | cut -d: -f1)
      if [ -n "$next_pattern" ] && [ "$next_pattern" -lt 10 ]; then
        echo "  - $file:$line_num: Catch-all *) may execute before specific patterns"
        P16_COUNT=$((P16_COUNT + 1))
      fi
    fi
  done < <(grep -n ')$' "$file" 2>/dev/null | head -20 || true)
done
echo "  $P16_COUNT findings"

# P17: String Comparison on Version Strings (HIGH)
# Using string comparison operators on semantic versions
echo ""
echo "[P17] Version String Comparison"
for file in $PY_FILES; do
  [ -f "$file" ] || continue
  # Pattern: version >= "4.18" or similar (string comparison on semver)
  while IFS=: read -r line_num content; do
    # Skip if using proper version comparison libraries
    if ! grep -q "packaging\|semver\|version\|LooseVersion\|StrictVersion\|parse_version" "$file" 2>/dev/null; then
      echo "  - $file:$line_num: Possible string comparison on semantic version"
      P17_COUNT=$((P17_COUNT + 1))
    fi
  done < <(grep -nE '[><=!]=?\s*["\x27][0-9]+\.[0-9]+' "$file" 2>/dev/null | \
           grep -v '#.*[><=]' | head -5 || true)
done
# Check shell scripts too
for file in $SH_FILES; do
  [ -f "$file" ] || continue
  # Look for -gt/-lt/-ge/-le on version-like strings or == with version
  while IFS=: read -r line_num content; do
    if echo "$content" | grep -qE '\[\[.*[0-9]+\.[0-9]+.*(-gt|-lt|-ge|-le|==|!=)'; then
      echo "  - $file:$line_num: Version comparison may need numeric parsing"
      P17_COUNT=$((P17_COUNT + 1))
    fi
  done < <(grep -nE '\[\[.*[0-9]+\.[0-9]+' "$file" 2>/dev/null | head -5 || true)
done
echo "  $P17_COUNT findings"

# P18: Unused CLI Flags/Variables (MEDIUM)
# Flags defined but never used in the script
echo ""
echo "[P18] Unused CLI Flags"
for file in $SH_FILES; do
  [ -f "$file" ] || continue
  # Find variables that look like flags (uppercase, ending in _FLAG, _MODE, _ENABLED)
  while IFS= read -r var; do
    [ -z "$var" ] && continue
    # Count usages (excluding the assignment line)
    usage_count=$(grep -c "\$$var\|\${$var}" "$file" 2>/dev/null || echo 0)
    # If only defined once (the assignment), it's unused
    if [ "$usage_count" -le 1 ]; then
      echo "  - $file: Flag $var defined but may be unused"
      P18_COUNT=$((P18_COUNT + 1))
    fi
  done < <(grep -oE '^[A-Z][A-Z0-9_]*(_FLAG|_MODE|_ENABLED|_OUTPUT|_DEBUG|_VERBOSE)=' "$file" 2>/dev/null | \
           cut -d= -f1 | sort -u || true)
done
# Check Python argparse flags
for file in $PY_FILES; do
  [ -f "$file" ] || continue
  python3 -c "
import ast
try:
    with open('$file') as f: src = f.read()
    tree = ast.parse(src)
    args_vars = set()
    for node in ast.walk(tree):
        # Find add_argument calls
        if isinstance(node, ast.Call):
            if isinstance(node.func, ast.Attribute) and node.func.attr == 'add_argument':
                for arg in node.args:
                    if isinstance(arg, ast.Constant) and isinstance(arg.value, str):
                        if arg.value.startswith('--'):
                            dest = arg.value[2:].replace('-', '_')
                            args_vars.add(dest)
    # Check if these args are used in the file
    for var in args_vars:
        if src.count(f'args.{var}') + src.count(f\"'{var}'\") <= 1:
            pass  # Could report but too many false positives
except: pass
" 2>/dev/null || true
done
echo "  $P18_COUNT findings"

# P19: YAML Schema Coverage (HIGH)
# YAML fields that may not be read by companion code
echo ""
echo "[P19] YAML Schema Coverage"
if command -v yq &>/dev/null; then
  # Find YAML files and check if their keys are accessed in code
  yaml_files=$(find "${TARGET:-.}" -name "*.yaml" -o -name "*.yml" 2>/dev/null | filter_files | head -15)
  for yaml in $yaml_files; do
    [ -f "$yaml" ] || continue
    dir=$(dirname "$yaml")
    base=$(basename "$yaml" | sed 's/\.ya\?ml$//')

    # Extract top-level keys
    keys=$(yq eval 'keys | .[]' "$yaml" 2>/dev/null | head -10 || true)
    for key in $keys; do
      [ -z "$key" ] && continue
      # Check if this key is accessed in any Python/Shell code in same directory tree
      found=$(grep -r "\.get(['\"]$key\|getattr.*['\"]$key\|\['$key'\]\|\[\"$key\"\]\|\$$key\|\${$key}" "$dir" \
              --include="*.py" --include="*.sh" 2>/dev/null | head -1 || true)
      if [ -z "$found" ]; then
        echo "  - $yaml: Key '$key' may not be read by code"
        P19_COUNT=$((P19_COUNT + 1))
      fi
    done
  done
else
  echo "  WARNING: yq not installed (YAML coverage skipped)"
fi
echo "  $P19_COUNT findings"

# P20: Missing Cluster Connectivity Gate (HIGH)
# Scripts using cluster commands without connectivity check
echo ""
echo "[P20] Missing Cluster Connectivity Gate"
for file in $SH_FILES; do
  [ -f "$file" ] || continue
  # Check if file uses oc/kubectl commands
  if grep -qE '\b(oc|kubectl)\s+(get|patch|apply|delete|create|replace|rollout|scale)\b' "$file" 2>/dev/null; then
    # Check if there's a connectivity check
    if ! grep -qE 'oc whoami|kubectl cluster-info|oc cluster-info|oc auth can-i|kubectl auth can-i' "$file" 2>/dev/null; then
      echo "  - $file: Uses cluster commands without connectivity check"
      P20_COUNT=$((P20_COUNT + 1))
    fi
  fi
done
# Check Python files for kubernetes client usage without connectivity check
for file in $PY_FILES; do
  [ -f "$file" ] || continue
  if grep -qE 'kubernetes\.|from kubernetes' "$file" 2>/dev/null; then
    if ! grep -qE 'list_namespace|get_api_versions|version_api|CoreV1Api\(\)\.list_namespace' "$file" 2>/dev/null; then
      echo "  - $file: Uses kubernetes client without connectivity check"
      P20_COUNT=$((P20_COUNT + 1))
    fi
  fi
done
echo "  $P20_COUNT findings"

# Summary
echo ""
echo "=============================================="
echo "Pre-Scan Results:"
CRITICAL=$P1_COUNT
HIGH=$((P4_COUNT + P5_COUNT + P8_COUNT + P11_COUNT + P13_COUNT + P15_COUNT + P16_COUNT + P17_COUNT + P19_COUNT + P20_COUNT))
MEDIUM=$((P9_COUNT + P12_COUNT + P14_COUNT + P18_COUNT))
TOTAL=$((CRITICAL + HIGH + MEDIUM))

echo "[P1] Phantom Modifications: $P1_COUNT findings"
echo "[P4] Invisible Undone: $P4_COUNT findings"
echo "[P5] Eldritch Horror: $P5_COUNT findings"
echo "[P8] Cargo Cult Error Handling: $P8_COUNT findings"
echo "[P9] Documentation Phantom: $P9_COUNT findings"
echo "[P11] Shellcheck Violations (Shell): $P11_COUNT findings"
echo "[P12] Zombie Code: $P12_COUNT findings"
echo "[P13] Undocumented Error Ignores (Go): $P13_COUNT findings"
echo "[P14] Error Wrapping with %v (Go): $P14_COUNT findings"
echo "[P15] golangci-lint Violations (Go): $P15_COUNT findings"
echo "[P16] Catch-All Pattern Precedence: $P16_COUNT findings"
echo "[P17] Version String Comparison: $P17_COUNT findings"
echo "[P18] Unused CLI Flags: $P18_COUNT findings"
echo "[P19] YAML Schema Coverage: $P19_COUNT findings"
echo "[P20] Missing Cluster Connectivity Gate: $P20_COUNT findings"
echo "----------------------------------------------"
echo "Summary: $TOTAL findings ($CRITICAL CRITICAL, $HIGH HIGH, $MEDIUM MEDIUM)"
echo ""

[ "$CRITICAL" -gt 0 ] && echo "CRITICAL: Fix P1 immediately"
[ "$HIGH" -gt 0 ] && echo "HIGH: Review P4, P5, P8, P11, P13, P15, P16, P17, P19, P20"
[ "$MEDIUM" -gt 0 ] && echo "MEDIUM: Consider P9, P12, P14, P18"
[ "$TOTAL" -eq 0 ] && echo "All clear - no violations"
echo "=============================================="

# Exit code based on findings
[ "$CRITICAL" -gt 0 ] && exit 2
[ "$HIGH" -gt 0 ] && exit 3
exit 0
