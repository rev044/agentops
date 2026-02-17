#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
SKILL="$ROOT/skills/reverse-engineer-rpi"

if ! command -v go >/dev/null 2>&1; then
  echo "error: go is required for the demo fixture build" >&2
  exit 2
fi

TMP="$ROOT/.tmp/reverse-engineer-rpi-self-test"
OUT1="$TMP/out-core"
OUT2="$TMP/out-sec"
SRC="$TMP/fixture-src"
BIN="$TMP/demo_bin"
SITEMAP="$TMP/sitemap.xml"

rm -rf "$TMP"
mkdir -p "$SRC" "$OUT1" "$OUT2"

python3 - "$SRC" <<'PY'
import sys, zipfile
from pathlib import Path

src = Path(sys.argv[1])
(src / "payload.zip").parent.mkdir(parents=True, exist_ok=True)
with zipfile.ZipFile(src / "payload.zip", "w", compression=zipfile.ZIP_DEFLATED) as zf:
    zf.writestr("agent/main.py", "print('hello from demo agent')\n")
    zf.writestr("agent/README.md", "# Demo Agent\n")
    zf.writestr("agent/SYSTEM_PROMPT.txt", "DEMO PROMPT (do not dump in reports)\n")
PY

cat >"$SRC/main.go" <<'EOF'
package main

import _ "embed"
import "fmt"

//go:embed payload.zip
var payload []byte

func main() {
	// Ensure the bytes are referenced so the ZIP signature is present in the binary.
	fmt.Printf("demo binary; embedded payload bytes=%d\n", len(payload))
}
EOF

(cat >"$SRC/go.mod" <<'EOF'
module demo_embedded_zip

go 1.22
EOF
)

(cd "$SRC" && go build -o "$BIN" .)

cat >"$SITEMAP" <<'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url><loc>https://example.test/docs/features/alpha/overview</loc></url>
  <url><loc>https://example.test/docs/features/alpha/howto</loc></url>
  <url><loc>https://example.test/docs/features/beta/overview</loc></url>
</urlset>
EOF

python3 "$SKILL/scripts/reverse_engineer_rpi.py" demo \
  --authorized \
  --mode=binary \
  --binary-path="$BIN" \
  --docs-sitemap-url="file://$SITEMAP" \
  --output-dir="$OUT1"

python3 "$OUT1/validate-feature-registry.py"

python3 "$SKILL/scripts/reverse_engineer_rpi.py" demo \
  --authorized \
  --mode=binary \
  --binary-path="$BIN" \
  --docs-sitemap-url="file://$SITEMAP" \
  --output-dir="$OUT2" \
  --security-audit \
  --sbom

"$OUT2/security/validate-security-audit.sh" "$OUT2" --sbom

echo "OK: self-test passed"
