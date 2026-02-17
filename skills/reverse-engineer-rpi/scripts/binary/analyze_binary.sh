#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 2 ]]; then
  echo "usage: analyze_binary.sh <binary_path> <out_dir>" >&2
  exit 2
fi

BIN="$1"
OUT="$2"
mkdir -p "$OUT"

if [[ ! -f "$BIN" ]]; then
  echo "error: binary not found: $BIN" >&2
  exit 2
fi

{
  echo "# Binary Analysis (Best-Effort)"
  echo
  echo "- Target: \`$BIN\`"
  echo "- Generated: $(date +%F)"
  echo
  echo "## file(1)"
  echo
  if command -v file >/dev/null 2>&1; then
    file "$BIN" || true
  else
    echo "_file not available_"
  fi
  echo
  echo "## Linked Libraries (best-effort)"
  echo
  if command -v otool >/dev/null 2>&1; then
    otool -L "$BIN" 2>/dev/null || true
  elif command -v ldd >/dev/null 2>&1; then
    ldd "$BIN" 2>/dev/null || true
  else
    echo "_otool/ldd not available_"
  fi
  echo
  echo "## Language Heuristics (best-effort)"
  echo
  if command -v strings >/dev/null 2>&1; then
    # Go heuristics: runtime symbols + build id markers are common.
    if strings -a "$BIN" 2>/dev/null | rg -n -S -m 1 'runtime\\.morestack|go\\.buildid|Go build ID|type\\.\\*runtime\\.' >/dev/null 2>&1; then
      echo "- Likely language/runtime: Go (heuristic: Go runtime markers in strings)"
    else
      echo "- Likely language/runtime: unknown (no Go markers found)"
    fi
  else
    echo "- strings not available; cannot run heuristics"
  fi
  echo
  echo "## Embedded Archive Signatures (ZIP, best-effort)"
  echo
  if command -v python3 >/dev/null 2>&1; then
    python3 - "$BIN" <<'PY'
import sys
from pathlib import Path

p = Path(sys.argv[1])
data = p.read_bytes()

sig = b"PK\x03\x04"
hits = []
start = 0
while True:
    i = data.find(sig, start)
    if i < 0:
        break
    hits.append(i)
    start = i + 1

print(f"- ZIP local header occurrences: {len(hits)}")
for i in hits[:10]:
    print(f"  - offset: {i}")
if len(hits) > 10:
    print("  - ...")
PY
  else
    echo "_python3 not available_"
  fi
} >"$OUT/binary-analysis.md"

# Raw strings (kept under tmp out dir; do not copy into output_dir by default).
if command -v strings >/dev/null 2>&1; then
  strings -a "$BIN" 2>/dev/null | head -2000 >"$OUT/strings.head.txt" || true
  strings -a "$BIN" 2>/dev/null | rg -n -S 'mcp|prompt|system|tool|openai|anthropic|claude' >"$OUT/strings.ai-hits.txt" 2>/dev/null || true
fi

# Optional disassembly snippet (bounded). Keep under tmp out dir; do not paste into reports by default.
if command -v otool >/dev/null 2>&1; then
  otool -tvV "$BIN" 2>/dev/null | head -500 >"$OUT/disassembly.head.txt" || true
elif command -v objdump >/dev/null 2>&1; then
  objdump -d "$BIN" 2>/dev/null | head -500 >"$OUT/disassembly.head.txt" || true
fi
