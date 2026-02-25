---
name: reverse-engineer-rpi
description: 'Reverse-engineer a product into a feature catalog, code map, and specs. Uses RPI-style loop with verification gates. Triggers: “reverse engineer”, “catalog features”, “feature inventory”, “code map”, “docs to code mapping”, “binary analysis”.'
---


# $reverse-engineer-rpi

Reverse-engineer a product into a mechanically verifiable feature inventory + registry + spec set, with optional security-audit artifacts and validation gates.

## Hard Guardrails (MANDATORY)

- Only operate on code/binaries you own or have **explicit written authorization** to analyze.
- Do not provide steps to bypass protections/ToS or to extract proprietary source code/system prompts from third-party products.
- Do not output reconstructed proprietary source or embedded prompts from binaries (index only; redact in reports).
- Redact secrets/tokens/keys if encountered; run the secret-scan gate over outputs.
- Always separate: **docs say** vs **code proves** vs **hosted/control-plane**.

## One-Command Example

```bash
python3 skills/reverse-engineer-rpi/scripts/reverse_engineer_rpi.py ao \
  --authorized \
  --mode=binary \
  --binary-path="$(command -v ao)" \
  --output-dir=".agents/research/ao/"
```

If you do not have explicit written authorization to analyze that binary, do not run the above. Use the included demo fixture instead (see Self-Test below).

Repo-only example (no binary required):

```bash
python3 skills/reverse-engineer-rpi/scripts/reverse_engineer_rpi.py cc-sdd \
  --mode=repo \
  --upstream-repo="https://github.com/gotalab/cc-sdd.git" \
  --output-dir=".agents/research/cc-sdd/"
```

Pinned clone (reproducible):

```bash
python3 skills/reverse-engineer-rpi/scripts/reverse_engineer_rpi.py cc-sdd \
  --mode=repo \
  --upstream-repo="https://github.com/gotalab/cc-sdd.git" \
  --upstream-ref=v1.0.0 \
  --output-dir=".agents/research/cc-sdd/"
```

## Invocation Contract

Required:
- `product_name`

Optional:
- `--docs-sitemap-url` (recommended when available; supports `https://...` and `file:///...`)
- `--docs-features-prefix` (default: `auto`; detects best local docs prefix, falls back to `docs/features/`)
- `--upstream-repo` (optional)
- `--upstream-ref` (pin clone to a specific commit, tag, or branch; records resolved SHA in `clone-metadata.json`)
- `--local-clone-dir` (default: `.tmp/<product_name>`)
- `--output-dir` (default: `.agents/research/<product_name>/`)
- `--mode` (default: `repo`; allowed: `repo|binary|both`)
- `--binary-path` (required if `--mode` includes `binary`)
- `--no-materialize-archives` (authorized-only; binary mode extracts embedded ZIPs by default; this disables extraction and keeps index-only)

Security audit flags (optional):
- `--security-audit` (enables security artifacts + gates)
- `--sbom` (generate SBOM + dependency risk report where possible; may no-op with a note)
- `--fuzz` (only if a safe harness exists; timeboxed)

Mandatory guardrail flag:
- `--authorized` (required for binary mode; refuses to run binary analysis without it)

## Upstream Ref Pinning (`--upstream-ref`)

Use `--upstream-ref` to pin a repo-mode clone to a specific commit, tag, or branch. This makes analysis reproducible and allows golden fixtures to be diffed against a known baseline.

```bash
# Pin to a tag (reproducible)
python3 skills/reverse-engineer-rpi/scripts/reverse_engineer_rpi.py cc-sdd \
  --mode=repo \
  --upstream-repo="https://github.com/gotalab/cc-sdd.git" \
  --upstream-ref=v1.0.0 \
  --output-dir=".agents/research/cc-sdd/"

# Pin to a specific commit SHA
python3 skills/reverse-engineer-rpi/scripts/reverse_engineer_rpi.py cc-sdd \
  --mode=repo \
  --upstream-repo="https://github.com/gotalab/cc-sdd.git" \
  --upstream-ref=abc1234 \
  --output-dir=".agents/research/cc-sdd/"
```

When `--upstream-ref` is provided:

- The clone is fetched with `git fetch --depth=1 origin <ref>` and checked out to `FETCH_HEAD`.
- The resolved commit SHA is recorded in `output_dir/clone-metadata.json` for traceability.
- Without `--upstream-ref`, a `--depth=1` shallow clone of the default branch HEAD is used instead.

`clone-metadata.json` schema:

```json
{
  "upstream_repo": "https://github.com/gotalab/cc-sdd.git",
  "upstream_ref": "v1.0.0",
  "resolved_commit": "<full SHA>",
  "clone_date": "YYYY-MM-DD"
}
```

## Contract Outputs (`output_dir/`)

Repo-mode analysis writes machine-checkable contract files under `output_dir/`. These files use only relative paths, sorted lists, and stable keys — no absolute paths, no run-specific timestamps — so they can be committed as golden fixtures and diffed across runs.

**Primary contract files:**

| File | Description |
|------|-------------|
| `feature-registry.yaml` | Structured feature inventory with mechanically-extracted CLI, config/env, and artifact surface |
| `cli-surface-contracts.txt` | CLI surface: commands, flags, help text, framework, language |
| `docs-features.txt` | Features extracted from documentation (docs say vs code proves) |
| `clone-metadata.json` | Upstream repo URL, pinned ref, resolved commit SHA, clone date |

Example `feature-registry.yaml` structure:

```yaml
schema_version: 1
product_name: cc-sdd
upstream_commit: "abc1234..."
features:
  - name: cli-entry
    cli:
      language: node
      bin:
        cc-sdd: dist/cli.js
      help_text: "Usage: cc-sdd [options] ..."
  - name: config-surface
    config_env:
      config_file: ".cc-sdd/config.json"
      env_vars:
        - name: CC_SDD_TOKEN
          evidence: ["src/config.ts"]
```

> Note: Contract outputs are written by `--mode=repo` (or `--mode=both`). Binary-mode outputs (`binary-analysis.md`, `binary-symbols.txt`, etc.) remain directly under `output_dir/`.

## Fixture Test Workflow

Golden fixtures allow regression detection: commit a known-good fixture snapshot (contract files alongside the pinned `clone-metadata.json`), then diff future runs against it.

### Running Fixture Tests

```bash
bash skills/reverse-engineer-rpi/scripts/repo_fixture_test.sh
```

This script (implemented in ag-w77.3):

1. Reads `skills/reverse-engineer-rpi/fixtures/cc-sdd-v2.1.0/clone-metadata.json` to determine the pinned upstream ref.
2. Runs `reverse_engineer_rpi.py` in repo mode with that ref into a temp output dir.
3. Diffs the generated outputs against the committed golden fixtures (`feature-registry.yaml`, `cli-surface-contracts.txt`, `docs-features.txt`).
4. Exits 0 if they match; exits non-zero with a unified diff if they drift.

The test requires network access to clone the upstream repo.

### Updating Fixtures

When contracts legitimately change (new flags, new env vars, schema bumps), update the golden fixtures:

```bash
# 1. Re-run with the pinned ref to generate fresh contracts
python3 skills/reverse-engineer-rpi/scripts/reverse_engineer_rpi.py cc-sdd \
  --mode=repo \
  --upstream-repo="https://github.com/gotalab/cc-sdd.git" \
  --upstream-ref=<new-tag-or-sha> \
  --output-dir=".tmp/cc-sdd-refresh/"

# 2. Copy contracts into the fixture directory
cp .tmp/cc-sdd-refresh/feature-registry.yaml \
  skills/reverse-engineer-rpi/fixtures/cc-sdd-v2.1.0/feature-registry.yaml

# 3. Update the pinned clone metadata
cp .tmp/cc-sdd-refresh/clone-metadata.json \
  skills/reverse-engineer-rpi/fixtures/cc-sdd-v2.1.0/clone-metadata.json

# 4. Commit the updated fixtures
git add skills/reverse-engineer-rpi/fixtures/cc-sdd-v2.1.0/
git commit -m "fix(reverse-engineer-rpi): update cc-sdd golden fixtures to <new-tag-or-sha>"
```

Fixture files that must be committed for the test to pass:

- `skills/reverse-engineer-rpi/fixtures/cc-sdd-v2.1.0/clone-metadata.json`
- `skills/reverse-engineer-rpi/fixtures/cc-sdd-v2.1.0/feature-registry.yaml`
- `skills/reverse-engineer-rpi/fixtures/cc-sdd-v2.1.0/cli-surface-contracts.txt`
- `skills/reverse-engineer-rpi/fixtures/cc-sdd-v2.1.0/docs-features.txt`

## Script-Driven Workflow

Run:

```bash
python3 skills/reverse-engineer-rpi/scripts/reverse_engineer_rpi.py <product_name> --authorized [flags...]
```

This generates the required outputs under `output_dir/` and (when applicable) `.agents/council/` and `.agents/learnings/`.

## Outputs (MUST be generated)

Core outputs under `output_dir/`:
1. `feature-inventory.md`
2. `feature-registry.yaml`
3. `validate-feature-registry.py`
4. `feature-catalog.md`
5. `spec-architecture.md`
6. `spec-code-map.md`
7. `spec-cli-surface.md` (Node, Python, or Go CLI detected; otherwise a note is written to `spec-code-map.md`)
8. `spec-clone-vs-use.md`
9. `spec-clone-mvp.md` (original MVP spec; do not copy from target)
10. `clone-metadata.json` (when `--upstream-repo` is used; records resolved commit SHA)

Binary-mode extras:
- `binary-analysis.md` (best-effort summary)
- `binary-embedded-archives.md` (index only; no dumps)

Repo-mode extras:
- `spec-artifact-surface.md` (best-effort; template/manifest driven install surface)
- `artifact-registry.json` (best-effort; hashed template inventory when manifests/templates exist)

If `--security-audit`, also create `output_dir/security/`:
- `threat-model.md`
- `attack-surface.md`
- `dataflow.md`
- `crypto-review.md`
- `authn-authz.md`
- `findings.md`
- `reproducibility.md`
- `validate-security-audit.sh`

## Self-Test (Acceptance Criteria)

End-to-end fixture (safe, owned demo binary with embedded ZIP):

```bash
bash skills/reverse-engineer-rpi/scripts/self_test.sh
```

This must show:
- feature inventory generated
- registry generated
- registry validator exits 0
- in security mode: `validate-security-audit.sh` exits 0 and secret scan passes

## Examples

### Scenario: Reverse-Engineer an Open-Source CLI in Repo Mode

**User says:** `$reverse-engineer-rpi cc-sdd --mode=repo --upstream-repo="https://github.com/gotalab/cc-sdd.git" --upstream-ref=v1.0.0`

**What happens:**
1. The script shallow-clones the upstream repo at the pinned tag `v1.0.0` and records the resolved SHA in `clone-metadata.json`.
2. It scans the repo for CLI entry points, config/env surface, schema files, and artifact manifests, then writes `feature-inventory.md`, `feature-registry.yaml`, contract JSON, and all spec files under the output directory.

**Result:** A complete feature catalog and machine-checkable `feature-registry.yaml` are generated under `.agents/research/cc-sdd/`, ready for golden-fixture diffing.

### Scenario: Binary Analysis With Security Audit

**User says:** `$reverse-engineer-rpi ao --authorized --mode=binary --binary-path="$(command -v ao)" --security-audit`

**What happens:**
1. The script runs static analysis on the `ao` binary (file metadata, linked libraries, embedded archive signatures) and writes `binary-analysis.md` and `binary-embedded-archives.md`.
2. It generates the full security audit suite (`threat-model.md`, `attack-surface.md`, `findings.md`, etc.) under `output_dir/security/` and runs the secret-scan gate over all outputs.

**Result:** Binary analysis artifacts plus a validated security audit are produced; `validate-security-audit.sh` exits 0 confirming all security deliverables are present and secrets-clean.

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| Script refuses to run binary analysis | Missing `--authorized` flag | Add `--authorized` to confirm you have explicit written authorization to analyze the binary. |
| `clone-metadata.json` not generated | `--upstream-repo` was not provided | Pass `--upstream-repo` (and optionally `--upstream-ref`) to enable clone metadata tracking. |
| Fixture test diff fails unexpectedly | Upstream repo changed or golden fixtures are stale | Re-run with the pinned ref, copy fresh contracts into `fixtures/`, and commit the updated golden files (see Updating Fixtures). |
| `spec-cli-surface.md` not generated | No recognized CLI framework (Node/Python/Go) detected in the repo | Check that the target repo has a discoverable CLI entry point; otherwise the CLI surface is documented in `spec-code-map.md` instead. |
| Network error during repo clone | Firewall, VPN, or GitHub rate limit blocking the shallow clone | Verify network connectivity, authenticate with `gh auth login` if the repo is private, or use `--local-clone-dir` to point at a pre-cloned directory. |

---

## Scripts

### extract_docs_features.sh

```bash
#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 2 ]]; then
  echo "usage: extract_docs_features.sh <paths.txt> <docs_features_prefix>" >&2
  exit 2
fi

PATHS_TXT="$1"
PREFIX_RAW="$2"

# Normalize prefix: "docs/features/" -> "/docs/features"
PREFIX="/${PREFIX_RAW#/}"
PREFIX="${PREFIX%/}"

python3 - "$PATHS_TXT" "$PREFIX" <<'PY'
import sys
from pathlib import Path

paths_txt = Path(sys.argv[1])
prefix = sys.argv[2]

out = set()
for line in paths_txt.read_text(encoding="utf-8", errors="replace").splitlines():
    p = line.strip()
    if not p:
        continue
    if not p.startswith("/"):
        p = "/" + p
    if p.startswith(prefix + "/") or p == prefix:
        # Keep the path *under* docs/features as a slug, without leading slash.
        slug = p.lstrip("/")
        out.add(slug)

for s in sorted(out):
    print(s)
PY
```

### extract_sitemap_paths.sh

```bash
#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: extract_sitemap_paths.sh <sitemap.xml>" >&2
  exit 2
fi

SITEMAP_XML="$1"

python3 - "$SITEMAP_XML" <<'PY'
import sys
import urllib.parse
import xml.etree.ElementTree as ET
from pathlib import Path

src = Path(sys.argv[1])
data = src.read_text(encoding="utf-8", errors="replace")
root = ET.fromstring(data)

paths = set()
for loc in root.iter():
    if loc.tag.endswith("loc") and loc.text:
        u = loc.text.strip()
        p = urllib.parse.urlparse(u)
        path = p.path or ""
        if not path:
            continue
        # Normalize: ensure leading slash, drop trailing slash except root.
        if not path.startswith("/"):
            path = "/" + path
        if len(path) > 1 and path.endswith("/"):
            path = path[:-1]
        paths.add(path)

for p in sorted(paths):
    print(p)
PY
```

### fetch_url.py

```python
#!/usr/bin/env python3
from __future__ import annotations

import sys
import urllib.parse
import urllib.request
from pathlib import Path


def main() -> int:
    if len(sys.argv) != 3:
        print("usage: fetch_url.py <url> <out_path>", file=sys.stderr)
        return 2
    url = sys.argv[1]
    out_path = Path(sys.argv[2])
    out_path.parent.mkdir(parents=True, exist_ok=True)

    parsed = urllib.parse.urlparse(url)
    if parsed.scheme in ("file", ""):
        src = Path(parsed.path if parsed.scheme == "file" else url)
        out_path.write_bytes(src.read_bytes())
        return 0

    req = urllib.request.Request(url, headers={"User-Agent": "reverse-engineer-rpi/1.0"})
    with urllib.request.urlopen(req, timeout=30) as resp:
        out_path.write_bytes(resp.read())
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
```

### generate_feature_catalog_md.py

```python
#!/usr/bin/env python3
from __future__ import annotations

import argparse
import datetime as _dt
from pathlib import Path


def _parse_registry(path: Path) -> dict:
    data = {"docs_features_prefix": "docs/features/", "docs_features": [], "groups": {}}
    cur = None
    in_docs = False
    in_groups = False
    in_anchors = False
    for raw in path.read_text(encoding="utf-8", errors="replace").splitlines():
        line = raw.rstrip("\n")
        if not line.strip() or line.lstrip().startswith("#"):
            continue
        if line.startswith("docs_features_prefix:"):
            data["docs_features_prefix"] = line.split(":", 1)[1].strip().strip("'\"")
        if line == "docs_features:":
            in_docs = True
            in_groups = False
            continue
        if line == "groups:":
            in_docs = False
            in_groups = True
            continue

        if in_docs and line.startswith("  - "):
            data["docs_features"].append(line[4:].strip().strip("'\""))
            continue

        if in_groups:
            if line.startswith("  ") and not line.startswith("    ") and line.endswith(":"):
                name = line.strip()[:-1]
                cur = {"impl": None, "anchors": [], "notes": ""}
                data["groups"][name] = cur
                in_anchors = False
                continue
            if cur is None:
                continue
            s = line.strip()
            if s.startswith("impl:"):
                cur["impl"] = s.split(":", 1)[1].strip()
            elif s.startswith("anchors:"):
                in_anchors = True
                if s.endswith("[]"):
                    cur["anchors"] = []
            elif in_anchors and s.startswith("- "):
                cur["anchors"].append(s[2:].strip().strip("'\""))
            elif s.startswith("notes:"):
                cur["notes"] = s.split(":", 1)[1].strip().strip("'\"")
    return data


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--registry", required=True)
    ap.add_argument("--out", required=True)
    args = ap.parse_args()

    reg = _parse_registry(Path(args.registry))
    groups = reg["groups"]

    out = Path(args.out)
    out.parent.mkdir(parents=True, exist_ok=True)

    lines: list[str] = []
    lines.append("# Feature Catalog")
    lines.append("")
    lines.append(f"- Generated: {_dt.date.today().isoformat()}")
    lines.append(f"- Groups: {len(groups)}")
    lines.append("")
    lines.append("| Group | impl | anchors | notes |")
    lines.append("|---|---|---:|---|")
    for g in sorted(groups.keys()):
        ent = groups[g]
        impl = ent.get("impl") or ""
        anchors = ent.get("anchors") or []
        notes = (ent.get("notes") or "").replace("\n", " ")
        lines.append(f"| `{g}` | `{impl}` | {len(anchors)} | {notes} |")
    lines.append("")
    out.write_text("\n".join(lines) + "\n", encoding="utf-8")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
```

### generate_feature_inventory_md.py

```python
#!/usr/bin/env python3
from __future__ import annotations

import argparse
import datetime as _dt
from pathlib import Path


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--product-name", required=True)
    ap.add_argument("--docs-features", required=True, help="Text file: one docs/features slug per line (may be empty).")
    ap.add_argument("--out", required=True)
    args = ap.parse_args()

    slugs_path = Path(args.docs_features)
    slugs = [ln.strip() for ln in slugs_path.read_text(encoding="utf-8", errors="replace").splitlines() if ln.strip()]

    out = Path(args.out)
    out.parent.mkdir(parents=True, exist_ok=True)

    lines: list[str] = []
    lines.append(f"# Feature Inventory: {args.product_name}")
    lines.append("")
    lines.append(f"- Generated: {_dt.date.today().isoformat()}")
    lines.append("- Source: docs sitemap inventory (if provided); otherwise empty/incomplete by design.")
    lines.append(f"- Count: {len(slugs)}")
    lines.append("")
    lines.append("## Docs Slugs")
    lines.append("")
    if slugs:
        for s in slugs:
            lines.append(f"- `{s}`")
    else:
        lines.append("_No docs sitemap provided (or no matching `docs/features/` entries)._")
    lines.append("")

    out.write_text("\n".join(lines) + "\n", encoding="utf-8")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
```

### repo_fixture_test.sh

```bash
#!/usr/bin/env bash
# repo_fixture_test.sh — Golden fixture self-test for cc-sdd repo-mode analysis.
#
# Pins to cc-sdd v2.1.0 (commit 6e972c064ac4723bc8ad0181871d07e199af6a9f) and
# runs repo-mode analysis, then compares key contracts against stored fixtures.
#
# Usage:
#   bash skills/reverse-engineer-rpi/scripts/repo_fixture_test.sh
#
# Exit codes:
#   0  All fixture contracts match.
#   1  One or more contracts drifted (diff output printed to stderr).
#   2  Prerequisite missing or unexpected error.
#
# Requirements:
#   - Network access (to clone github.com/gotalab/cc-sdd at v2.1.0)
#   - git, python3

set -euo pipefail

# ---------------------------------------------------------------------------
# Paths
# ---------------------------------------------------------------------------
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILL_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
# ROOT = git repo root (trunks/), two levels up from the skill dir
# (reverse-engineer-rpi -> skills -> trunks)
ROOT="$(cd "$SKILL_DIR/../.." && pwd)"
FIXTURES_DIR="$SKILL_DIR/fixtures/cc-sdd-v2.1.0"

PINNED_REF="v2.1.0"
PINNED_COMMIT="6e972c064ac4723bc8ad0181871d07e199af6a9f"
UPSTREAM_REPO="https://github.com/gotalab/cc-sdd.git"

TMP="$ROOT/.tmp/repo-fixture-test-cc-sdd"
OUT="$TMP/out"
CLONE_DIR="$TMP/local-clone"

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------
FAILURES=0

_fail() {
  echo "FAIL: $1" >&2
  FAILURES=$((FAILURES + 1))
}

_ok() {
  echo "OK: $1"
}

_check_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "error: required command not found: $1" >&2
    exit 2
  fi
}

# Normalize a YAML/text file: strip generated_at/clone_date/analysis_root lines
# (which are volatile) so diff is stable across run dates.
_normalize() {
  local f="$1"
  grep -v '^generated_at:' "$f" \
    | grep -v '^  "clone_date":' \
    | grep -v '^  "analysis_root":' \
    | grep -v '^  "node_package_dir":' \
    | grep -v '^analysis_root:' \
    | sed 's|^- Date: .*|- Date: <DATE>|' \
    | sed 's|^- Analysis root: .*|- Analysis root: <ROOT>|' \
    | sed "s|$(echo "$ROOT" | sed 's|/|\\/|g')|<ROOT>|g"
}

# ---------------------------------------------------------------------------
# Prerequisites
# ---------------------------------------------------------------------------
_check_cmd git
_check_cmd python3

if [ ! -d "$FIXTURES_DIR" ]; then
  echo "error: fixtures directory not found: $FIXTURES_DIR" >&2
  echo "       Run with UPDATE_FIXTURES=1 to create it, or check the skill directory." >&2
  exit 2
fi

# ---------------------------------------------------------------------------
# UPDATE_FIXTURES mode: regenerate and overwrite golden fixtures.
# ---------------------------------------------------------------------------
if [ "${UPDATE_FIXTURES:-0}" = "1" ]; then
  echo "=== UPDATE_FIXTURES=1: regenerating golden fixtures ==="
  rm -rf "$TMP"
  mkdir -p "$OUT" "$CLONE_DIR" "$FIXTURES_DIR"

  python3 "$SKILL_DIR/scripts/reverse_engineer_rpi.py" cc-sdd \
    --mode=repo \
    --upstream-repo="$UPSTREAM_REPO" \
    --upstream-ref="$PINNED_REF" \
    --local-clone-dir="$CLONE_DIR" \
    --output-dir="$OUT"

  # Verify resolved commit matches pin.
  ACTUAL_COMMIT="$(python3 -c "import json; d=json.load(open('$OUT/clone-metadata.json')); print(d['resolved_commit'])")"
  if [ "$ACTUAL_COMMIT" != "$PINNED_COMMIT" ]; then
    echo "WARNING: resolved commit $ACTUAL_COMMIT does not match expected pin $PINNED_COMMIT" >&2
    echo "         Update PINNED_COMMIT in this script if the tag was force-pushed." >&2
  fi

  # docs-features.txt — stable (content from repo tree).
  cp "$OUT/docs-features.txt" "$FIXTURES_DIR/docs-features.txt"

  # feature-registry.yaml — strip generated_at before storing.
  grep -v '^generated_at:' "$OUT/feature-registry.yaml" > "$FIXTURES_DIR/feature-registry.yaml"

  # clone-metadata.json — strip clone_date (volatile).
  python3 - "$OUT/clone-metadata.json" "$FIXTURES_DIR/clone-metadata.json" <<'PYEOF'
import json, sys
d = json.load(open(sys.argv[1]))
d.pop("clone_date", None)
open(sys.argv[2], "w").write(json.dumps({k: d[k] for k in ("upstream_repo", "upstream_ref", "resolved_commit")}, indent=2) + "\n")
PYEOF

  # cli-surface-contracts.txt — extract key contract lines from spec-cli-surface.md.
  python3 - "$OUT/spec-cli-surface.md" "$FIXTURES_DIR/cli-surface-contracts.txt" <<'PYEOF'
import sys, re

text = open(sys.argv[1]).read()
out_lines = [
    "# CLI surface contract assertions for cc-sdd v2.1.0",
    "# Each line below must appear verbatim in the generated spec-cli-surface.md",
    "# (after stripping leading/trailing whitespace).",
    "# Lines starting with # are comments.",
    "",
]

# Entrypoints section.
out_lines.append("# Package identity")
for pat in [r"- Node package: `.+`", r"- package name: `.+`", r"- version: `.+`"]:
    m = re.search(pat, text)
    if m:
        out_lines.append(m.group(0))

out_lines.append("")
out_lines.append("# Binary entrypoint")
m = re.search(r"- `cc-sdd` -> `.+`", text)
if m:
    out_lines.append(m.group(0))

out_lines.append("")
out_lines.append("# Source entry heuristic")
m = re.search(r"- `tools/cc-sdd/src/cli\.ts`.+", text)
if m:
    out_lines.append(m.group(0))

# Help text flags (extract lines from the code block).
in_block = False
flags = []
for line in text.splitlines():
    if line.strip().startswith("```"):
        in_block = not in_block
        continue
    if in_block and (line.startswith("  -") or line.startswith("-")):
        flags.append(line.rstrip())
if flags:
    out_lines.append("")
    out_lines.append("# Key CLI flags present in help text")
    out_lines.extend(flags)

# Config surface.
out_lines.append("")
out_lines.append("# Config surface")
for pat in [r"- User config file: `.+`", r"- Environment variables: `.+`"]:
    m = re.search(pat, text)
    if m:
        out_lines.append(m.group(0))

open(sys.argv[2], "w").write("\n".join(out_lines) + "\n")
print(f"Written {len(out_lines)} lines to {sys.argv[2]}")
PYEOF

  echo "=== Fixtures updated in $FIXTURES_DIR ==="
  exit 0
fi

# ---------------------------------------------------------------------------
# Normal mode: run analysis and compare against golden fixtures.
# ---------------------------------------------------------------------------
echo "=== repo_fixture_test.sh: cc-sdd v2.1.0 golden fixture test ==="
echo "    Pinned commit: $PINNED_COMMIT"
echo "    Fixtures:      $FIXTURES_DIR"
echo ""

# Clean output dir for reproducible run. The clone dir is preserved across runs
# to avoid re-downloading (a shallow clone is ~5-10 MB and slow on first run).
# However, we always delete the clone dir if the resolved SHA does not match the
# pinned commit (guards against a force-pushed tag).
if [ -d "$CLONE_DIR/.git" ]; then
  EXISTING_SHA="$(git -C "$CLONE_DIR" rev-parse HEAD 2>/dev/null || true)"
  if [ "$EXISTING_SHA" != "$PINNED_COMMIT" ]; then
    echo "--- Existing clone SHA ($EXISTING_SHA) != pin ($PINNED_COMMIT); re-cloning ---"
    rm -rf "$CLONE_DIR"
  else
    echo "--- Reusing cached clone at $PINNED_COMMIT ---"
  fi
fi

rm -rf "$OUT"
mkdir -p "$OUT"

if [ ! -d "$CLONE_DIR/.git" ]; then
  echo "--- Cloning cc-sdd at $PINNED_REF (network required) ---"
  mkdir -p "$CLONE_DIR"
fi

echo "--- Running repo-mode analysis ---"
python3 "$SKILL_DIR/scripts/reverse_engineer_rpi.py" cc-sdd \
  --mode=repo \
  --upstream-repo="$UPSTREAM_REPO" \
  --upstream-ref="$PINNED_REF" \
  --local-clone-dir="$CLONE_DIR" \
  --output-dir="$OUT"

# clone-metadata.json is only written by reverse_engineer_rpi.py during the initial
# clone. When reusing a cached clone, write it ourselves so downstream checks work.
if [ ! -f "$OUT/clone-metadata.json" ]; then
  RESOLVED_SHA="$(git -C "$CLONE_DIR" rev-parse HEAD 2>/dev/null || echo "")"
  python3 - "$OUT/clone-metadata.json" "$UPSTREAM_REPO" "$PINNED_REF" "$RESOLVED_SHA" <<'PYEOF'
import json, sys
out_path, repo, ref, sha = sys.argv[1], sys.argv[2], sys.argv[3], sys.argv[4]
data = {"upstream_repo": repo, "upstream_ref": ref, "resolved_commit": sha, "clone_date": "cached"}
open(out_path, "w").write(json.dumps(data, indent=2) + "\n")
PYEOF
fi

echo ""
echo "--- Verifying pinned commit SHA ---"
ACTUAL_COMMIT="$(python3 -c "import json; d=json.load(open('$OUT/clone-metadata.json')); print(d['resolved_commit'])")"
if [ "$ACTUAL_COMMIT" != "$PINNED_COMMIT" ]; then
  _fail "resolved commit mismatch: got $ACTUAL_COMMIT, expected $PINNED_COMMIT"
  echo "      This means the tag was force-pushed or the fixture pin is stale." >&2
else
  _ok "resolved commit matches pin ($PINNED_COMMIT)"
fi

# ---------------------------------------------------------------------------
# Contract 1: docs-features.txt (exact match)
# ---------------------------------------------------------------------------
echo ""
echo "--- Contract 1: docs-features.txt ---"
GOLDEN="$FIXTURES_DIR/docs-features.txt"
ACTUAL="$OUT/docs-features.txt"

if [ ! -f "$ACTUAL" ]; then
  _fail "docs-features.txt not generated"
else
  DIFF_OUT="$(diff --unified=3 "$GOLDEN" "$ACTUAL" 2>&1 || true)"
  if [ -n "$DIFF_OUT" ]; then
    _fail "docs-features.txt drifted from golden fixture"
    echo "--- diff (golden vs actual) ---" >&2
    echo "$DIFF_OUT" >&2
    echo "---" >&2
  else
    _ok "docs-features.txt matches golden fixture"
  fi
fi

# ---------------------------------------------------------------------------
# Contract 2: feature-registry.yaml (normalized, strip generated_at)
# ---------------------------------------------------------------------------
echo ""
echo "--- Contract 2: feature-registry.yaml (normalized) ---"
GOLDEN="$FIXTURES_DIR/feature-registry.yaml"
ACTUAL="$OUT/feature-registry.yaml"

if [ ! -f "$ACTUAL" ]; then
  _fail "feature-registry.yaml not generated"
else
  GOLDEN_NORM="$(mktemp)"
  ACTUAL_NORM="$(mktemp)"
  grep -v '^generated_at:' "$GOLDEN" > "$GOLDEN_NORM"
  grep -v '^generated_at:' "$ACTUAL" > "$ACTUAL_NORM"
  DIFF_OUT="$(diff --unified=3 "$GOLDEN_NORM" "$ACTUAL_NORM" 2>&1 || true)"
  rm -f "$GOLDEN_NORM" "$ACTUAL_NORM"
  if [ -n "$DIFF_OUT" ]; then
    _fail "feature-registry.yaml drifted from golden fixture"
    echo "--- diff (golden vs actual, generated_at stripped) ---" >&2
    echo "$DIFF_OUT" >&2
    echo "---" >&2
  else
    _ok "feature-registry.yaml matches golden fixture (normalized)"
  fi
fi

# ---------------------------------------------------------------------------
# Contract 3: cli-surface-contracts.txt (line-presence check in spec-cli-surface.md)
# ---------------------------------------------------------------------------
echo ""
echo "--- Contract 3: spec-cli-surface.md contract lines ---"
CLI_SURFACE="$OUT/spec-cli-surface.md"
CONTRACTS="$FIXTURES_DIR/cli-surface-contracts.txt"

if [ ! -f "$CLI_SURFACE" ]; then
  _fail "spec-cli-surface.md not generated"
elif [ ! -f "$CONTRACTS" ]; then
  echo "SKIP: cli-surface-contracts.txt fixture not found (non-fatal)"
else
  CONTRACT_FAILURES=0
  while IFS= read -r line; do
    # Skip blank lines and comments.
    [[ -z "$line" || "$line" == \#* ]] && continue
    # Check verbatim line presence (fixed-string, -- prevents lines starting with
    # '-' being misinterpreted as grep flags).
    if ! grep -qF -- "$line" "$CLI_SURFACE" 2>/dev/null; then
      _fail "contract line not found in spec-cli-surface.md: $line"
      CONTRACT_FAILURES=$((CONTRACT_FAILURES + 1))
    fi
  done < "$CONTRACTS"
  if [ "$CONTRACT_FAILURES" -eq 0 ]; then
    _ok "all spec-cli-surface.md contract lines present"
  fi
fi

# ---------------------------------------------------------------------------
# Contract 4: clone-metadata.json (key fields)
# ---------------------------------------------------------------------------
echo ""
echo "--- Contract 4: clone-metadata.json key fields ---"
GOLDEN="$FIXTURES_DIR/clone-metadata.json"
ACTUAL="$OUT/clone-metadata.json"

if [ ! -f "$ACTUAL" ]; then
  _fail "clone-metadata.json not generated"
elif [ ! -f "$GOLDEN" ]; then
  echo "SKIP: clone-metadata.json fixture not found (non-fatal)"
else
  # Compare only the stable fields (upstream_repo, upstream_ref, resolved_commit).
  GOLDEN_STABLE="$(mktemp)"
  ACTUAL_STABLE="$(mktemp)"
  python3 - "$GOLDEN" "$GOLDEN_STABLE" <<'PYEOF'
import json, sys
d = json.load(open(sys.argv[1]))
out = {k: d[k] for k in ("upstream_repo", "upstream_ref", "resolved_commit") if k in d}
open(sys.argv[2], "w").write(json.dumps(out, indent=2, sort_keys=True) + "\n")
PYEOF
  python3 - "$ACTUAL" "$ACTUAL_STABLE" <<'PYEOF'
import json, sys
d = json.load(open(sys.argv[1]))
out = {k: d[k] for k in ("upstream_repo", "upstream_ref", "resolved_commit") if k in d}
open(sys.argv[2], "w").write(json.dumps(out, indent=2, sort_keys=True) + "\n")
PYEOF
  DIFF_OUT="$(diff --unified=3 "$GOLDEN_STABLE" "$ACTUAL_STABLE" 2>&1 || true)"
  rm -f "$GOLDEN_STABLE" "$ACTUAL_STABLE"
  if [ -n "$DIFF_OUT" ]; then
    _fail "clone-metadata.json stable fields drifted from golden fixture"
    echo "--- diff (golden vs actual, stable fields only) ---" >&2
    echo "$DIFF_OUT" >&2
    echo "---" >&2
  else
    _ok "clone-metadata.json stable fields match golden fixture"
  fi
fi

# ---------------------------------------------------------------------------
# Contract 5: required output files exist
# ---------------------------------------------------------------------------
echo ""
echo "--- Contract 5: required output files exist ---"
REQUIRED_FILES=(
  feature-inventory.md
  feature-registry.yaml
  feature-catalog.md
  spec-architecture.md
  spec-code-map.md
  spec-clone-vs-use.md
  spec-clone-mvp.md
  spec-cli-surface.md
  spec-artifact-surface.md
  artifact-registry.json
  clone-metadata.json
  docs-features.txt
  validate-feature-registry.py
)

for f in "${REQUIRED_FILES[@]}"; do
  if [ ! -f "$OUT/$f" ]; then
    _fail "required output file missing: $f"
  else
    _ok "exists: $f"
  fi
done

# ---------------------------------------------------------------------------
# Contract 6: feature registry validator passes
# ---------------------------------------------------------------------------
echo ""
echo "--- Contract 6: feature registry validator ---"
if python3 "$OUT/validate-feature-registry.py" 2>&1; then
  _ok "validate-feature-registry.py exit 0"
else
  _fail "validate-feature-registry.py exited non-zero"
fi

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
echo ""
if [ "$FAILURES" -gt 0 ]; then
  echo "RESULT: FAIL — $FAILURES contract(s) drifted. See diff output above." >&2
  exit 1
else
  echo "RESULT: PASS — all golden fixture contracts match."
  exit 0
fi
```

### reverse_engineer_rpi.py

```python
#!/usr/bin/env python3
from __future__ import annotations

import argparse
import datetime as _dt
import hashlib
import json
import re
import shutil
import subprocess
import sys
from pathlib import Path


REPO_ROOT = Path.cwd()
SKILL_DIR = Path(__file__).resolve().parents[1]
TEMPLATES_DIR = SKILL_DIR / "references" / "templates"


def _die(msg: str, code: int = 2) -> None:
    print(f"error: {msg}", file=sys.stderr)
    raise SystemExit(code)


def _run(cmd: list[str], *, cwd: Path | None = None, check: bool = True) -> subprocess.CompletedProcess:
    return subprocess.run(cmd, cwd=str(cwd) if cwd else None, check=check)


def _ensure_dirs(paths: list[Path]) -> None:
    for p in paths:
        p.mkdir(parents=True, exist_ok=True)


def _today_ymd() -> str:
    return _dt.date.today().isoformat()


def _slugify(s: str) -> str:
    out = []
    for ch in s.strip().lower():
        if ch.isalnum():
            out.append(ch)
        elif ch in (" ", "-", "_", "/"):
            out.append("-")
    slug = "".join(out)
    while "--" in slug:
        slug = slug.replace("--", "-")
    return slug.strip("-") or "product"


def _detect_docs_prefix_for_repo(analysis_root: Path) -> str:
    """
    Choose a sensible docs slug prefix for repos that do not use docs/features/.
    Returns a prefix with trailing slash.
    """
    candidates = [
        "docs/features/",
        "docs/code-map/",
        "docs/workflows/",
        "docs/levels/",
        "docs/",
    ]
    best = "docs/features/"
    best_count = -1
    for cand in candidates:
        base = analysis_root / cand.strip("/")
        if not base.exists() or not base.is_dir():
            continue
        count = 0
        for p in base.rglob("*"):
            if p.is_file() and p.suffix.lower() in (".md", ".mdx"):
                count += 1
        if count > best_count:
            best = cand
            best_count = count
    if best_count >= 0:
        return best
    return "docs/features/"


def _detect_docs_prefix_from_paths(paths: list[str]) -> str:
    """
    Choose docs prefix from sitemap-style path inventory.
    """
    normalized: list[str] = []
    for raw in paths:
        p = raw.strip()
        if not p:
            continue
        if not p.startswith("/"):
            p = "/" + p
        normalized.append(p)

    if not normalized:
        return "docs/features/"

    candidates = [
        "docs/features/",
        "docs/code-map/",
        "docs/workflows/",
        "docs/levels/",
        "docs/",
    ]
    best = "docs/features/"
    best_count = -1
    for cand in candidates:
        prefix = "/" + cand.strip("/").rstrip("/")
        count = sum(1 for p in normalized if p == prefix or p.startswith(prefix + "/"))
        if count > best_count:
            best = cand
            best_count = count
    return best


def _render_template(src: Path, dst: Path, vars: dict[str, str]) -> None:
    text = src.read_text(encoding="utf-8")
    for k, v in vars.items():
        text = text.replace("{{" + k + "}}", v)
    dst.write_text(text, encoding="utf-8")


def _read_text(p: Path) -> str:
    return p.read_text(encoding="utf-8", errors="replace")


def _extract_ts_backtick_const(src: Path, const_name: str) -> str | None:
    # Best-effort: extract `const <name> = `...`;` blocks (common for CLI help text).
    if not src.exists():
        return None
    text = _read_text(src)
    m = re.search(
        rf"\bconst\s+{re.escape(const_name)}\s*=\s*`([\s\S]*?)`;",
        text,
        flags=re.MULTILINE,
    )
    return m.group(1) if m else None


def _extract_ts_string_const(src: Path, const_name: str) -> str | None:
    if not src.exists():
        return None
    text = _read_text(src)
    m = re.search(rf"\b{re.escape(const_name)}\s*=\s*'([^']*)';", text)
    if m:
        return m.group(1)
    m = re.search(rf'\b{re.escape(const_name)}\s*=\s*"([^"]*)";', text)
    if m:
        return m.group(1)
    return None


def _extract_agents_from_registry_ts(registry_ts: Path) -> tuple[list[str], list[str]] | None:
    """
    Best-effort parser for agent keys + alias flags from a TS registry.
    Intended to resolve help text interpolations like `${agentKeys.join('|')}`.
    """
    if not registry_ts.exists():
        return None

    text = _read_text(registry_ts)
    start = text.find("export const agentDefinitions")
    if start < 0:
        return None
    tail = text[start:]

    # Limit to the agentDefinitions object body to reduce false matches.
    end = tail.find("} as const")
    if end > 0:
        tail = tail[:end]

    agent_keys: list[str] = []
    seen_keys: set[str] = set()

    for line in tail.splitlines():
        # Top-level agent keys in the registry are consistently 2-space indented. This avoids
        # accidentally matching nested object keys like `layout:` or `commands:`.
        m = re.match(r"^  (?:'([^']+)'|([A-Za-z0-9_-]+))\s*:\s*\{\s*$", line)
        if not m:
            continue
        key = (m.group(1) or m.group(2) or "").strip()
        if not key:
            continue
        if key not in seen_keys:
            agent_keys.append(key)
            seen_keys.add(key)

    alias_flags: set[str] = set()
    for m in re.finditer(r"aliasFlags:\s*\[([^\]]*)\]", tail, flags=re.MULTILINE):
        blob = m.group(1)
        for s in re.findall(r"'([^']+)'", blob):
            alias_flags.add(s)
        for s in re.findall(r"\"([^\"]+)\"", blob):
            alias_flags.add(s)

    return agent_keys, sorted(alias_flags)


def _find_node_cli_package(repo_root: Path, product_slug: str, product_name: str) -> dict[str, object] | None:
    # Detect Node CLI packages by locating a package.json with a "bin" field and matching name/bin key.
    product_name_lc = product_name.strip().lower()
    candidates: list[tuple[int, Path, dict[str, object]]] = []

    for pkg_json in sorted(repo_root.rglob("package.json")):
        if "node_modules" in pkg_json.parts:
            continue
        try:
            data = json.loads(_read_text(pkg_json))
        except Exception:
            continue

        bin_field = data.get("bin")
        if not bin_field:
            continue

        name = str(data.get("name") or "")
        score = 0
        if name.lower() == product_slug or name.lower() == product_name_lc:
            score += 100

        # Normalize bin mapping.
        bin_map: dict[str, str] = {}
        if isinstance(bin_field, str):
            if name:
                bin_map[name] = bin_field
        elif isinstance(bin_field, dict):
            for k, v in bin_field.items():
                if isinstance(k, str) and isinstance(v, str):
                    bin_map[k] = v
        if product_slug in bin_map:
            score += 80
        if product_name_lc in (k.lower() for k in bin_map.keys()):
            score += 60

        # Prefer shallower packages when score ties (often the main package vs nested deps).
        depth = len(pkg_json.relative_to(repo_root).parts)
        score -= depth

        candidates.append((score, pkg_json, data))

    if not candidates:
        return None

    candidates.sort(key=lambda t: t[0], reverse=True)
    score, pkg_json, data = candidates[0]

    # Return a normalized payload for downstream rendering.
    bin_field = data.get("bin")
    bin_map: dict[str, str] = {}
    if isinstance(bin_field, str):
        name = str(data.get("name") or "")
        if name:
            bin_map[name] = bin_field
    elif isinstance(bin_field, dict):
        for k, v in bin_field.items():
            if isinstance(k, str) and isinstance(v, str):
                bin_map[k] = v

    return {
        "score": score,
        "package_json": str(pkg_json),
        "package_dir": str(pkg_json.parent),
        "name": str(data.get("name") or ""),
        "version": str(data.get("version") or ""),
        "bin": bin_map,
    }


def _find_python_cli(repo_root: Path) -> dict[str, object] | None:
    """Detect Python CLI packages via pyproject.toml or setup.cfg entry_points."""
    result: dict[str, object] = {"language": "python", "bin": {}, "framework": None, "entry_module": None}

    # Try pyproject.toml first (modern standard).
    for pyproject in sorted(repo_root.rglob("pyproject.toml")):
        if ".venv" in pyproject.parts or "node_modules" in pyproject.parts:
            continue
        text = _read_text(pyproject)
        # [project.scripts] section (PEP 621).
        m = re.search(r'\[project\.scripts\]\s*\n((?:[^\[].+\n)*)', text)
        if m:
            for line in m.group(1).strip().splitlines():
                parts = line.split("=", 1)
                if len(parts) == 2:
                    name = parts[0].strip().strip('"').strip("'")
                    entry = parts[1].strip().strip('"').strip("'")
                    result["bin"][name] = entry  # type: ignore[index]
                    if not result["entry_module"]:
                        result["entry_module"] = entry.split(":")[0] if ":" in entry else entry
        # [tool.poetry.scripts] section.
        m2 = re.search(r'\[tool\.poetry\.scripts\]\s*\n((?:[^\[].+\n)*)', text)
        if m2:
            for line in m2.group(1).strip().splitlines():
                parts = line.split("=", 1)
                if len(parts) == 2:
                    name = parts[0].strip().strip('"').strip("'")
                    entry = parts[1].strip().strip('"').strip("'")
                    result["bin"][name] = entry  # type: ignore[index]
        if result["bin"]:
            break

    # Try setup.cfg if pyproject didn't find scripts.
    if not result["bin"]:
        for setup_cfg in sorted(repo_root.rglob("setup.cfg")):
            if ".venv" in setup_cfg.parts:
                continue
            text = _read_text(setup_cfg)
            m = re.search(r'\[options\.entry_points\]\s*\nconsole_scripts\s*=\s*\n((?:\s+.+\n)*)', text)
            if m:
                for line in m.group(1).strip().splitlines():
                    parts = line.strip().split("=", 1)
                    if len(parts) == 2:
                        result["bin"][parts[0].strip()] = parts[1].strip()  # type: ignore[index]
                if result["bin"]:
                    break

    if not result["bin"]:
        return None

    # Detect CLI framework via source scan (best-effort, cap file count).
    scanned = 0
    for py_file in sorted(repo_root.rglob("*.py")):
        if ".venv" in py_file.parts or "node_modules" in py_file.parts:
            continue
        scanned += 1
        if scanned > 200:
            break
        text = _read_text(py_file)
        if "@click.command" in text or "@click.group" in text:
            result["framework"] = "click"
            break
        if "typer.Typer" in text or "@app.command" in text:
            result["framework"] = "typer"
            break
        if "ArgumentParser(" in text and "add_argument" in text:
            result["framework"] = "argparse"

    return result


def _find_go_cli(repo_root: Path) -> dict[str, object] | None:
    """Detect Go CLI packages via go.mod + main.go + flag/cobra usage."""
    result: dict[str, object] = {"language": "go", "bin": {}, "framework": None, "module": None}

    # Find go.mod for module name.
    go_mod = repo_root / "go.mod"
    if not go_mod.exists():
        # Check one level deeper (monorepo).
        for gm in sorted(repo_root.rglob("go.mod")):
            go_mod = gm
            break
    if go_mod.exists():
        text = _read_text(go_mod)
        m = re.search(r'^module\s+(.+)$', text, re.MULTILINE)
        if m:
            result["module"] = m.group(1).strip()

    # Find main.go files (entry points).
    main_files: list[Path] = []
    for mg in sorted(repo_root.rglob("main.go")):
        if "vendor" in mg.parts or "testdata" in mg.parts:
            continue
        main_files.append(mg)

    if not main_files and not result["module"]:
        return None

    # Derive binary names from cmd/ pattern or root main.go.
    for mf in main_files:
        rel = mf.relative_to(repo_root)
        parts = rel.parts
        if len(parts) >= 3 and parts[-3] == "cmd":
            # cmd/<name>/main.go pattern.
            result["bin"][parts[-2]] = str(rel)  # type: ignore[index]
        elif len(parts) == 1:
            # Root main.go — use module basename or directory name.
            mod = str(result.get("module") or "")
            name = mod.rsplit("/", 1)[-1] if mod else repo_root.name
            result["bin"][name] = str(rel)  # type: ignore[index]

    if not result["bin"]:
        return None

    # Detect CLI framework (cobra vs stdlib flag).
    scanned = 0
    for go_file in sorted(repo_root.rglob("*.go")):
        if "vendor" in go_file.parts or "testdata" in go_file.parts:
            continue
        scanned += 1
        if scanned > 200:
            break
        text = _read_text(go_file)
        if "cobra.Command" in text or '"github.com/spf13/cobra"' in text:
            result["framework"] = "cobra"
            break
        if "flag.String" in text or "flag.Bool" in text or "flag.Int" in text:
            result["framework"] = "flag"

    return result


def _sha256_file(p: Path) -> str:
    h = hashlib.sha256()
    with p.open("rb") as f:
        for chunk in iter(lambda: f.read(1024 * 1024), b""):
            h.update(chunk)
    return h.hexdigest()


def _render_placeholders(s: str, vars: dict[str, str]) -> str:
    out = s
    for k, v in vars.items():
        out = out.replace("{{" + k + "}}", v)
    return out


def _enrich_registry_with_binary_evidence(
    registry_yaml: Path,
    tmp_dir: Path,
    output_dir: Path,
    *,
    product_name: str,
    date: str,
) -> bool:
    """Enrich feature-registry.yaml with binary string evidence.

    Reads cli-commands.txt and binary strings to create evidence-backed groups.
    Also generates binary-symbols.txt in the output dir.
    Returns True if enrichment was applied.
    """
    commands_file = tmp_dir / "binary" / "cli-commands.txt"
    _strings_file = tmp_dir / "binary" / "strings.head.txt"
    _ba_file = tmp_dir / "binary" / "binary-analysis.md"

    # Generate binary-symbols.txt from strings
    full_strings = tmp_dir / "binary" / "strings.head.txt"
    symbols_out = output_dir / "binary-symbols.txt"
    if full_strings.exists() and not symbols_out.exists():
        shutil.copyfile(full_strings, symbols_out)

    # Gather command groups from cli-commands.txt
    cmd_groups: dict[str, list[str]] = {}
    if commands_file.exists():
        for line in commands_file.read_text(encoding="utf-8").splitlines():
            line = line.strip()
            if not line:
                continue
            parts = line.split()
            group = parts[0]
            cmd_groups.setdefault(group, []).append(line)

    if not cmd_groups:
        return False

    # Try to load the existing registry (manual parse, no yaml dep)
    reg: dict = {"groups": {}}
    try:
        text = registry_yaml.read_text(encoding="utf-8")
        for raw_line in text.splitlines():
            stripped = raw_line.strip()
            if stripped.startswith("docs_features_prefix:"):
                reg["docs_features_prefix"] = stripped.split(":", 1)[1].strip().strip("'\"")
            elif stripped.startswith("docs_features:"):
                reg.setdefault("docs_features", [])
            elif raw_line.startswith("  - ") and "docs_features" in reg and "groups" not in text.split(raw_line)[0].rsplit("docs_features:", 1)[-1]:
                reg.setdefault("docs_features", []).append(stripped[2:].strip().strip("'\""))
        # Parse groups using the same logic as the validator
        cur = None
        in_groups = False
        in_anchors = False
        for raw_line in text.splitlines():
            line = raw_line.rstrip()
            if not line.strip() or line.lstrip().startswith("#"):
                continue
            if line == "groups:":
                in_groups = True
                continue
            if not in_groups:
                continue
            if line.startswith("  ") and not line.startswith("    ") and line.endswith(":"):
                name = line.strip()[:-1]
                cur = {"impl": None, "anchors": [], "notes": ""}
                reg["groups"][name] = cur
                in_anchors = False
                continue
            if cur is None:
                continue
            s = line.strip()
            if s.startswith("impl:"):
                cur["impl"] = s.split(":", 1)[1].strip()
            elif s.startswith("anchors:"):
                in_anchors = True
                if s.endswith("[]"):
                    cur["anchors"] = []
            elif in_anchors and s.startswith("- "):
                cur["anchors"].append(s[2:].strip().strip("'\""))
            elif s.startswith("notes:"):
                cur["notes"] = s.split(":", 1)[1].strip().strip("'\"")
    except Exception:
        return False

    groups = reg.get("groups", {})

    # Check if registry is already populated (has non-empty groups with notes)
    has_content = any(g.get("notes") for g in groups.values()) if groups else False
    if has_content:
        # Already enriched or populated — don't overwrite
        return False

    # Build new groups from binary command data
    new_groups: dict[str, dict] = {}
    for grp_name, cmds in sorted(cmd_groups.items()):
        slug = grp_name.replace("-", "_")
        subcmds = [c for c in cmds if c != grp_name]
        sub_str = ", ".join(subcmds) if subcmds else "no subcommands"
        new_groups[slug] = {
            "impl": "client",
            "anchors": ["binary-symbols.txt"],
            "notes": f"{grp_name} ({len(cmds)} commands: {sub_str})",
        }

    # Write registry in the manual format expected by validate_feature_registry.py
    lines: list[str] = []
    lines.append("schema_version: 1")
    lines.append(f"product_name: {product_name!r}")
    lines.append(f"generated_at: {date!r}")
    lines.append("evidence_source: 'binary --help + string extraction'")
    # Preserve docs_features_prefix if present
    dfp = reg.get("docs_features_prefix", "docs/features/")
    lines.append(f"docs_features_prefix: {dfp!r}")
    # Preserve docs_features list if present
    docs_feats = reg.get("docs_features", [])
    if docs_feats:
        lines.append("docs_features:")
        for df in docs_feats:
            lines.append(f"  - {df!r}")
    lines.append("groups:")
    for slug, grp in new_groups.items():
        lines.append(f"  {slug}:")
        lines.append(f"    impl: {grp['impl']}")
        lines.append("    anchors:")
        for a in grp["anchors"]:
            lines.append(f"      - {a}")
        lines.append(f"    notes: {grp['notes']!r}")

    registry_yaml.write_text("\n".join(lines) + "\n", encoding="utf-8")
    return True


def _write_binary_cli_surface_spec(
    output_dir: Path,
    tmp_dir: Path,
    *,
    product_name: str,
    date: str,
) -> bool:
    """Write spec-cli-surface.md from binary --help output or binary strings.

    Returns True if a spec was written.
    """
    help_tree = tmp_dir / "binary" / "cli-help-tree.txt"
    commands_file = tmp_dir / "binary" / "cli-commands.txt"
    strings_file = tmp_dir / "binary" / "strings.head.txt"

    lines: list[str] = []
    lines.append(f"# CLI Surface Spec: {product_name}")
    lines.append("")
    lines.append(f"- Date: {date}")
    lines.append("- Source: binary --help output" if help_tree.exists() else "- Source: binary string extraction")
    lines.append("")

    cmd_count = 0
    if commands_file.exists():
        cmds = [c.strip() for c in commands_file.read_text(encoding="utf-8").splitlines() if c.strip()]
        cmd_count = len(cmds)

    if help_tree.exists():
        tree_text = help_tree.read_text(encoding="utf-8")
        lines.append("## Command Count")
        lines.append("")
        lines.append(f"- **{cmd_count} commands** discovered via recursive `--help` execution")
        lines.append("")

        # Extract top-level commands and subcommands
        if commands_file.exists():
            top_level = sorted(set(c.split()[0] for c in cmds if c.strip()))
            lines.append("## Top-Level Commands")
            lines.append("")
            lines.append("| Command | Subcommands |")
            lines.append("|---------|-------------|")
            for top in top_level:
                subs = [c for c in cmds if c.startswith(top + " ") and c != top]
                sub_names = [c.split(maxsplit=1)[1] if " " in c else "" for c in subs]
                sub_str = ", ".join(f"`{s}`" for s in sub_names if s) if sub_names else "—"
                lines.append(f"| `{top}` | {sub_str} |")
            lines.append("")

        lines.append("## Full Help Tree")
        lines.append("")
        lines.append("```")
        # Truncate to avoid massive output
        tree_lines = tree_text.splitlines()
        if len(tree_lines) > 500:
            lines.extend(tree_lines[:500])
            lines.append(f"... ({len(tree_lines) - 500} more lines)")
        else:
            lines.extend(tree_lines)
        lines.append("```")
        lines.append("")
    elif strings_file.exists():
        # Fallback: extract command-like patterns from strings
        raw = strings_file.read_text(encoding="utf-8", errors="replace")
        usage_lines = [line.strip() for line in raw.splitlines() if "usage" in line.lower() or "Usage" in line]
        lines.append("## CLI Surface (from binary strings, best-effort)")
        lines.append("")
        if usage_lines:
            for u in usage_lines[:20]:
                lines.append(f"- `{u[:200]}`")
        else:
            lines.append("_No usage patterns found in binary strings._")
        lines.append("")
    else:
        return False

    out = output_dir / "spec-cli-surface.md"
    out.write_text("\n".join(lines).rstrip() + "\n", encoding="utf-8")
    return True


def _write_cli_surface_spec(
    output_dir: Path,
    *,
    product_name: str,
    product_slug: str,
    date: str,
    analysis_root: Path,
) -> bool:
    """
    Return True if a CLI was detected and spec-cli-surface.md was written.

    Repo-mode only. This is best-effort and aims to capture a mechanically-verifiable contract:
    - entrypoints (package.json bin)
    - help/usage text (static extraction, with interpolation resolved when possible)
    - config/env surface
    """
    if not analysis_root.exists():
        return False

    node_cli = _find_node_cli_package(analysis_root, product_slug, product_name)
    python_cli = _find_python_cli(analysis_root) if not node_cli else None
    go_cli = _find_go_cli(analysis_root) if not node_cli and not python_cli else None

    if not node_cli and not python_cli and not go_cli:
        return False

    # If Python or Go CLI detected (non-Node), write a language-appropriate spec.
    if python_cli or go_cli:
        cli_info = python_cli or go_cli
        assert cli_info is not None
        out = output_dir / "spec-cli-surface.md"
        lines: list[str] = []
        lang = str(cli_info["language"]).capitalize()
        lines.append(f"# CLI Surface Spec: {product_name}")
        lines.append("")
        lines.append(f"- Date: {date}")
        lines.append(f"- Language: {lang}")
        lines.append(f"- Analysis root: `{analysis_root}`")
        if cli_info.get("framework"):
            lines.append(f"- Framework: {cli_info['framework']}")
        if cli_info.get("module"):
            lines.append(f"- Module: `{cli_info['module']}`")
        if cli_info.get("entry_module"):
            lines.append(f"- Entry module: `{cli_info['entry_module']}`")
        lines.append("")
        lines.append("## Entrypoints (Code-Proven)")
        lines.append("")
        bin_map = cli_info.get("bin") or {}
        if isinstance(bin_map, dict) and bin_map:
            for k in sorted(bin_map.keys()):
                lines.append(f"- `{k}` -> `{bin_map[k]}`")
        else:
            lines.append("- _No entrypoints extracted._")
        lines.append("")
        lines.append("## Notes For 1:1 Fidelity")
        lines.append("")
        lines.append("- Run `<binary> --help` to capture the full CLI contract as a golden test fixture.")
        if lang == "Python":
            lines.append("- For Click/Typer apps, consider `<binary> --help` per subcommand for full coverage.")
        elif lang == "Go":
            lines.append("- For Cobra apps, consider `<binary> help <subcommand>` for full coverage.")
        out.write_text("\n".join(lines).rstrip() + "\n", encoding="utf-8")
        return True

    out = output_dir / "spec-cli-surface.md"
    pkg_dir = Path(str(node_cli["package_dir"]))

    pkg_json_rel = Path(str(node_cli["package_json"])).relative_to(analysis_root).as_posix()
    src_index = pkg_dir / "src" / "index.ts"
    src_cli = pkg_dir / "src" / "cli.ts"
    src_store = pkg_dir / "src" / "cli" / "store.ts"
    src_agents = pkg_dir / "src" / "agents" / "registry.ts"

    help_text = _extract_ts_backtick_const(src_index, "helpText")
    config_file = _extract_ts_string_const(src_store, "CONFIG_FILE")

    # Resolve common interpolations in helpText for higher-fidelity output.
    if help_text and src_agents.exists():
        extracted = _extract_agents_from_registry_ts(src_agents)
        if extracted:
            agent_keys, alias_flags = extracted
            if agent_keys:
                help_text = help_text.replace("${agentKeys.join('|')}", "|".join(agent_keys))
            alias_line = ""
            if alias_flags:
                alias_line = f"  {' | '.join(alias_flags)}  Agent alias flags\n"
            help_text = help_text.replace("${agentAliasLine}", alias_line)

    # Env vars: scan the src tree for process.env.<NAME> patterns.
    env_vars: list[str] = []
    src_root = pkg_dir / "src"
    if src_root.exists():
        pat = re.compile(r"\bprocess\.env\.([A-Z][A-Z0-9_]*)\b")
        found = set()
        for p in sorted(src_root.rglob("*")):
            if not p.is_file() or p.suffix.lower() not in (".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs"):
                continue
            for m in pat.finditer(_read_text(p)):
                found.add(m.group(1))
        env_vars = sorted(found)

    lines: list[str] = []
    lines.append(f"# CLI Surface Spec: {product_name}")
    lines.append("")
    lines.append(f"- Date: {date}")
    lines.append(f"- Analysis root: `{analysis_root}`")
    lines.append("")
    lines.append("## Entrypoints (Code-Proven)")
    lines.append("")
    lines.append(f"- Node package: `{pkg_dir.relative_to(analysis_root).as_posix()}`")
    lines.append(f"- package.json: `{pkg_json_rel}`")
    if node_cli.get("name"):
        lines.append(f"- package name: `{node_cli['name']}`")
    if node_cli.get("version"):
        lines.append(f"- version: `{node_cli['version']}`")
    lines.append("")
    lines.append("### Binaries")
    lines.append("")
    bin_map = node_cli.get("bin") or {}
    if isinstance(bin_map, dict) and bin_map:
        for k in sorted(bin_map.keys()):
            v = str(bin_map[k])
            lines.append(f"- `{k}` -> `{v}`")
    else:
        lines.append("- _No `bin` mapping extracted (unexpected)._")

    if src_cli.exists():
        lines.append("")
        lines.append("### Source Entry (Heuristic)")
        lines.append("")
        lines.append(f"- `{src_cli.relative_to(analysis_root).as_posix()}` (node shebang entry; typically calls `runCli`)")

    lines.append("")
    lines.append("## Usage / Help (Code-Proven Where Possible)")
    lines.append("")
    if help_text:
        lines.append("```text")
        lines.append(help_text.rstrip("\n"))
        lines.append("```")
        lines.append("")
        lines.append("Evidence:")
        lines.append(f"- `{src_index.relative_to(analysis_root).as_posix()}` (`helpText`)")
    else:
        lines.append("- _Help text not extracted (pattern not found)._")
        lines.append("Evidence:")
        lines.append(f"- `{src_index.relative_to(analysis_root).as_posix()}`")

    lines.append("")
    lines.append("## Config / Env (Code-Proven Where Possible)")
    lines.append("")
    wrote_any = False
    if config_file:
        lines.append(f"- User config file: `{config_file}` (loaded from CWD).")
        lines.append(f"  Evidence: `{src_store.relative_to(analysis_root).as_posix()}`")
        wrote_any = True
    if env_vars:
        lines.append(f"- Environment variables: `{', '.join(env_vars)}`")
        lines.append(f"  Evidence: scan of `{src_root.relative_to(analysis_root).as_posix()}` for `process.env.<NAME>`.")
        wrote_any = True
    if not wrote_any:
        lines.append("- _No config/env surface extracted._")

    lines.append("")
    lines.append("## Notes For 1:1 Fidelity")
    lines.append("")
    lines.append("- Treat `--help` output as the CLI contract; include it as a golden test fixture for regressions.")
    lines.append("- If the repo does not ship built artifacts (ex: `dist/`), building may be required to execute the CLI directly.")

    out.write_text("\n".join(lines).rstrip() + "\n", encoding="utf-8")
    return True


def _write_artifact_surface_spec(
    output_dir: Path,
    *,
    product_name: str,
    product_slug: str,
    date: str,
    analysis_root: Path,
) -> None:
    """
    Higher-fidelity extraction of "what the product writes/installs" for template-driven CLIs.

    Emits:
    - spec-artifact-surface.md (human summary)
    - artifact-registry.json (machine-usable: manifests + template file hashes)
    """
    out_md = output_dir / "spec-artifact-surface.md"
    out_json = output_dir / "artifact-registry.json"

    if not analysis_root.exists():
        out_md.write_text(
            f"# Artifact Surface Spec: {product_name}\n\n- Date: {date}\n\n- _No repo content available to analyze._\n",
            encoding="utf-8",
        )
        return

    node_cli = _find_node_cli_package(analysis_root, product_slug, product_name)
    if not node_cli:
        out_md.write_text(
            f"# Artifact Surface Spec: {product_name}\n\n- Date: {date}\n\n- _No Node CLI package detected; artifact extraction not implemented for this repo._\n",
            encoding="utf-8",
        )
        return

    pkg_dir = Path(str(node_cli["package_dir"]))
    manifests_dir = pkg_dir / "templates" / "manifests"
    if not manifests_dir.exists():
        out_md.write_text(
            f"# Artifact Surface Spec: {product_name}\n\n- Date: {date}\n\n"
            f"- _No `templates/manifests/` directory found under `{pkg_dir.relative_to(analysis_root).as_posix()}`._\n",
            encoding="utf-8",
        )
        return

    manifest_files = sorted(manifests_dir.glob("*.json"))
    manifests: list[dict[str, object]] = []
    resolved_sources: list[dict[str, object]] = []

    for mf in manifest_files:
        try:
            data = json.loads(_read_text(mf))
        except Exception:
            continue

        agent = None
        artifacts = data.get("artifacts") if isinstance(data, dict) else None
        if isinstance(artifacts, list):
            for a in artifacts:
                if isinstance(a, dict):
                    when = a.get("when")
                    if isinstance(when, dict) and isinstance(when.get("agent"), str):
                        agent = when.get("agent")
                        break

        manifests.append(
            {
                "path": mf.relative_to(analysis_root).as_posix(),
                "agent": agent,
                "raw": data,
            }
        )

        # Build resolved source inventory (what files are copied from templates).
        if not isinstance(artifacts, list):
            continue

        placeholder_vars = {"AGENT": agent} if isinstance(agent, str) else {}
        for a in artifacts:
            if not isinstance(a, dict):
                continue
            source = a.get("source")
            if not isinstance(source, dict):
                continue
            stype = source.get("type")
            if stype == "templateDir":
                from_dir = source.get("fromDir")
                if not isinstance(from_dir, str):
                    continue
                from_dir_res = _render_placeholders(from_dir, placeholder_vars) if placeholder_vars else from_dir
                abs_from = pkg_dir / from_dir_res
                if abs_from.exists() and abs_from.is_dir():
                    for fp in sorted(abs_from.rglob("*")):
                        if not fp.is_file():
                            continue
                        resolved_sources.append(
                            {
                                "manifest": mf.relative_to(analysis_root).as_posix(),
                                "artifact_id": a.get("id"),
                                "source_type": "templateDir",
                                "from": from_dir_res,
                                "file": fp.relative_to(pkg_dir).as_posix(),
                                "sha256": _sha256_file(fp),
                            }
                        )
            elif stype == "templateFile":
                from_file = source.get("from")
                if not isinstance(from_file, str):
                    continue
                from_file_res = _render_placeholders(from_file, placeholder_vars) if placeholder_vars else from_file
                abs_from = pkg_dir / from_file_res
                if abs_from.exists() and abs_from.is_file():
                    resolved_sources.append(
                        {
                            "manifest": mf.relative_to(analysis_root).as_posix(),
                            "artifact_id": a.get("id"),
                            "source_type": "templateFile",
                            "from": from_file_res,
                            "file": abs_from.relative_to(pkg_dir).as_posix(),
                            "sha256": _sha256_file(abs_from),
                        }
                    )

    out_json.write_text(
        json.dumps(
            {
                "schema_version": 1,
                "product_name": product_name,
                "generated_at": date,
                "analysis_root": str(analysis_root),
                "node_package_dir": pkg_dir.relative_to(analysis_root).as_posix(),
                "manifests": manifests,
                "resolved_template_files": resolved_sources,
            },
            indent=2,
            sort_keys=True,
        )
        + "\n",
        encoding="utf-8",
    )

    lines: list[str] = []
    lines.append(f"# Artifact Surface Spec: {product_name}")
    lines.append("")
    lines.append(f"- Date: {date}")
    lines.append(f"- Analysis root: `{analysis_root}`")
    lines.append(f"- Node package: `{pkg_dir.relative_to(analysis_root).as_posix()}`")
    lines.append(f"- Manifests dir: `{manifests_dir.relative_to(analysis_root).as_posix()}`")
    lines.append(f"- Machine registry: `{out_json.relative_to(output_dir).as_posix()}`")
    lines.append("")
    lines.append("## Manifest Inventory (Code-Proven)")
    lines.append("")
    if manifest_files:
        for mf in manifest_files:
            rel = mf.relative_to(analysis_root).as_posix()
            agent = None
            for m in manifests:
                if m.get("path") == rel:
                    agent = m.get("agent")
                    break
            agent_note = f" (agent={agent})" if agent else ""
            lines.append(f"- `{rel}`{agent_note}")
    else:
        lines.append("- _No manifest JSON files found._")

    lines.append("")
    lines.append("## Template Source File Inventory (Hashed)")
    lines.append("")
    lines.append(f"- Files hashed: `{len(resolved_sources)}`")
    lines.append("- Use `artifact-registry.json` as the source of truth for 1:1 template content equivalence.")

    out_md.write_text("\n".join(lines).rstrip() + "\n", encoding="utf-8")


def _get_upstream_commit(analysis_root: Path) -> str | None:
    """Return the HEAD commit SHA if analysis_root is a git repo, else None."""
    git_dir = analysis_root / ".git"
    if not git_dir.exists():
        return None
    try:
        sha = subprocess.check_output(
            ["git", "-C", str(analysis_root), "rev-parse", "HEAD"],
            text=True,
            stderr=subprocess.DEVNULL,
        ).strip()
        return sha if sha else None
    except Exception:
        return None


def _collect_env_vars_with_evidence(
    analysis_root: Path,
) -> list[dict[str, object]]:
    """
    Scan source files for environment variable references and return a sorted list
    with per-var file evidence.  Covers:
      - TypeScript/JavaScript: process.env.VAR_NAME
      - Python: os.environ['VAR'] / os.environ.get('VAR') / os.getenv('VAR')
      - Go: os.Getenv("VAR") / os.LookupEnv("VAR")
      - Shell: $VAR_NAME (upper-snake only, cap at 300 files)
    """
    var_files: dict[str, set[str]] = {}

    patterns: list[tuple[re.Pattern[str], set[str]]] = [
        (re.compile(r"\bprocess\.env\.([A-Z][A-Z0-9_]+)\b"), {".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs"}),
        (re.compile(r"""os\.environ(?:\.get)?\s*\(\s*['"]([A-Z][A-Z0-9_]+)['"]\s*\)"""), {".py"}),
        (re.compile(r"""\bos\.getenv\s*\(\s*['"]([A-Z][A-Z0-9_]+)['"]\s*\)"""), {".py"}),
        (re.compile(r"""\bos\.(?:Getenv|LookupEnv)\s*\(\s*"([A-Z][A-Z0-9_]+)"\s*\)"""), {".go"}),
        (re.compile(r'\$\{?([A-Z][A-Z0-9_]{2,})\}?'), {".sh", ".bash", ".env", ".envrc"}),
    ]

    scanned = 0
    for p in sorted(analysis_root.rglob("*")):
        if not p.is_file():
            continue
        # Skip irrelevant dirs
        skip_dirs = {"node_modules", ".git", ".venv", "vendor", "testdata", "__pycache__"}
        if any(part in skip_dirs for part in p.parts):
            continue
        suffix = p.suffix.lower()
        matching_pats = [pat for pat, suffixes in patterns if suffix in suffixes]
        if not matching_pats:
            continue
        scanned += 1
        if scanned > 500:
            break
        try:
            text = _read_text(p)
        except Exception:
            continue
        rel = p.relative_to(analysis_root).as_posix()
        for pat in matching_pats:
            for m in pat.finditer(text):
                name = m.group(1)
                var_files.setdefault(name, set()).add(rel)

    result: list[dict[str, object]] = []
    for var_name in sorted(var_files.keys()):
        result.append({
            "name": var_name,
            "files": sorted(var_files[var_name]),
        })
    return result


def _collect_schema_files(analysis_root: Path) -> list[str]:
    """
    Return sorted relative paths of schema-like files in the repo.
    Matches: *.schema.json, *schema*.json, openapi*.json/yaml, swagger*.json/yaml,
             *.proto, *.avsc, *.thrift, graphql schema files.
    """
    schema_patterns = [
        "**/*.schema.json",
        "**/*schema*.json",
        "**/openapi*.json",
        "**/openapi*.yaml",
        "**/openapi*.yml",
        "**/swagger*.json",
        "**/swagger*.yaml",
        "**/swagger*.yml",
        "**/*.proto",
        "**/*.avsc",
        "**/*.thrift",
        "**/schema.graphql",
        "**/*.graphql",
    ]
    skip_dirs = {"node_modules", ".git", ".venv", "vendor", "testdata", "__pycache__"}
    found: set[str] = set()
    for pattern in schema_patterns:
        for p in analysis_root.glob(pattern):
            if not p.is_file():
                continue
            if any(part in skip_dirs for part in p.relative_to(analysis_root).parts):
                continue
            found.add(p.relative_to(analysis_root).as_posix())
    return sorted(found)


def _collect_config_files(analysis_root: Path) -> list[str]:
    """
    Return sorted relative paths of config files commonly read at runtime.
    Matches common config naming patterns at any depth (capped at 300 files).
    """
    config_name_patterns = re.compile(
        r"^(config|configuration|settings|\.env|app\.config|appsettings"
        r"|pyproject|setup\.cfg|cargo\.toml|go\.mod|tsconfig|jest\.config"
        r"|webpack\.config|vite\.config|babel\.config|eslint.*|\.eslintrc.*"
        r"|prettier.*|\.prettierrc.*)(\.(json|yaml|yml|toml|ini|cfg|js|ts|cjs|mjs))?$",
        re.IGNORECASE,
    )
    skip_dirs = {"node_modules", ".git", ".venv", "vendor", "testdata", "__pycache__"}
    found: set[str] = set()
    count = 0
    for p in sorted(analysis_root.rglob("*")):
        if not p.is_file():
            continue
        if any(part in skip_dirs for part in p.relative_to(analysis_root).parts):
            continue
        if config_name_patterns.match(p.name):
            found.add(p.relative_to(analysis_root).as_posix())
            count += 1
            if count >= 300:
                break
    return sorted(found)


def _write_repo_contract_json(
    output_dir: Path,
    analysis_root: Path,
    *,
    product_name: str,
    product_slug: str,
) -> Path:
    """
    Write a deterministic, machine-checkable contract JSON to
    output_dir/contracts/repo-contract.json.

    Contract includes:
    - upstream_commit (if analysis_root is a git repo)
    - cli surface: bin map, help text (static extraction), config file, env vars with file evidence
    - manifest inventory + template file hashes (from artifact-registry.json if present)
    - schema-like files
    - config files discovered in repo

    No absolute paths, no dates — stable across runs on the same commit.
    """
    contracts_dir = output_dir / "contracts"
    contracts_dir.mkdir(parents=True, exist_ok=True)
    out_path = contracts_dir / "repo-contract.json"

    contract: dict[str, object] = {
        "schema_version": 1,
        "product_name": product_name,
    }

    # upstream_commit
    upstream_commit = _get_upstream_commit(analysis_root)
    if upstream_commit:
        contract["upstream_commit"] = upstream_commit

    # --- CLI surface ---
    cli_surface: dict[str, object] = {}

    node_cli = _find_node_cli_package(analysis_root, product_slug, product_name)
    python_cli_info = _find_python_cli(analysis_root) if node_cli is None else None
    go_cli_info = _find_go_cli(analysis_root) if node_cli is None and python_cli_info is None else None

    if node_cli:
        pkg_dir = Path(str(node_cli["package_dir"]))
        # bin map with relative paths
        raw_bin = node_cli.get("bin") or {}
        bin_map: dict[str, str] = {}
        if isinstance(raw_bin, dict):
            for k, v in raw_bin.items():
                bin_map[k] = v
        cli_surface["language"] = "node"
        cli_surface["package_json"] = Path(str(node_cli["package_json"])).relative_to(analysis_root).as_posix()
        cli_surface["package_dir"] = pkg_dir.relative_to(analysis_root).as_posix()
        cli_surface["package_name"] = str(node_cli.get("name") or "")
        cli_surface["bin"] = {k: bin_map[k] for k in sorted(bin_map)}

        # Help text (static extraction)
        src_index = pkg_dir / "src" / "index.ts"
        src_agents = pkg_dir / "src" / "agents" / "registry.ts"
        help_text = _extract_ts_backtick_const(src_index, "helpText")
        if help_text and src_agents.exists():
            extracted = _extract_agents_from_registry_ts(src_agents)
            if extracted:
                agent_keys, alias_flags = extracted
                if agent_keys:
                    help_text = help_text.replace("${agentKeys.join('|')}", "|".join(agent_keys))
                alias_line = ""
                if alias_flags:
                    alias_line = f"  {' | '.join(alias_flags)}  Agent alias flags\n"
                help_text = help_text.replace("${agentAliasLine}", alias_line)
        if help_text is not None:
            cli_surface["help_text"] = help_text
            cli_surface["help_text_source"] = src_index.relative_to(analysis_root).as_posix() if src_index.exists() else None

        # Config file from store.ts
        src_store = pkg_dir / "src" / "cli" / "store.ts"
        config_file = _extract_ts_string_const(src_store, "CONFIG_FILE")
        if config_file:
            cli_surface["config_file"] = config_file
            cli_surface["config_file_source"] = src_store.relative_to(analysis_root).as_posix() if src_store.exists() else None

    elif python_cli_info:
        raw_bin_py = python_cli_info.get("bin") or {}
        cli_surface["language"] = "python"
        cli_surface["framework"] = python_cli_info.get("framework")
        cli_surface["entry_module"] = python_cli_info.get("entry_module")
        cli_surface["bin"] = {k: str(raw_bin_py[k]) for k in sorted(raw_bin_py)} if isinstance(raw_bin_py, dict) else {}

    elif go_cli_info:
        raw_bin_go = go_cli_info.get("bin") or {}
        cli_surface["language"] = "go"
        cli_surface["framework"] = go_cli_info.get("framework")
        cli_surface["module"] = go_cli_info.get("module")
        cli_surface["bin"] = {k: str(raw_bin_go[k]) for k in sorted(raw_bin_go)} if isinstance(raw_bin_go, dict) else {}

    contract["cli"] = cli_surface

    # --- Env vars with per-var file evidence ---
    contract["env_vars"] = _collect_env_vars_with_evidence(analysis_root)

    # --- Manifest inventory + template file hashes ---
    artifact_registry_path = output_dir / "artifact-registry.json"
    if artifact_registry_path.exists():
        try:
            artifact_data = json.loads(_read_text(artifact_registry_path))
            manifests_raw = artifact_data.get("manifests") or []
            template_files_raw = artifact_data.get("resolved_template_files") or []

            # Manifests: keep only path and agent (drop raw JSON for contract stability)
            manifests_clean: list[dict[str, object]] = []
            for m in manifests_raw:
                entry: dict[str, object] = {"path": m.get("path")}
                if m.get("agent"):
                    entry["agent"] = m["agent"]
                manifests_clean.append(entry)

            # Template files: keep path, sha256 (no absolute paths; already relative in artifact-registry)
            template_hashes: list[dict[str, object]] = []
            for tf in template_files_raw:
                template_hashes.append({
                    "file": tf.get("file"),
                    "manifest": tf.get("manifest"),
                    "sha256": tf.get("sha256"),
                    "source_type": tf.get("source_type"),
                })

            contract["manifests"] = sorted(manifests_clean, key=lambda x: str(x.get("path", "")))
            contract["template_files"] = sorted(template_hashes, key=lambda x: str(x.get("file", "")))
        except Exception:
            pass

    # --- Schema-like files ---
    contract["schema_files"] = _collect_schema_files(analysis_root)

    # --- Config files ---
    contract["config_files"] = _collect_config_files(analysis_root)

    out_path.write_text(
        json.dumps(contract, indent=2, sort_keys=True) + "\n",
        encoding="utf-8",
    )
    return out_path


def _write_comparison_report(
    output_dir: Path,
    tmp_dir: Path,
    *,
    product_name: str,
    date: str,
) -> bool:
    """Write comparison-report.md contrasting binary vs repo analysis results.

    Returns True if a report was written.
    """
    # --- Command discovery ---
    binary_cmds: list[str] = []
    commands_file = tmp_dir / "binary" / "cli-commands.txt"
    if commands_file.exists():
        binary_cmds = [c.strip() for c in commands_file.read_text(encoding="utf-8").splitlines() if c.strip()]

    repo_cmds: list[str] = []
    repo_cli_spec = output_dir / "spec-cli-surface.md"
    if repo_cli_spec.exists():
        # Extract command names from the table rows (| `cmd` | ... |)
        text = repo_cli_spec.read_text(encoding="utf-8")
        for m in re.finditer(r"^\|\s*`([^`]+)`\s*\|", text, re.MULTILINE):
            cmd = m.group(1).strip()
            if cmd and cmd not in ("Command",):
                repo_cmds.append(cmd)

    binary_set = set(binary_cmds)
    repo_set = set(repo_cmds)
    only_binary = sorted(binary_set - repo_set)
    only_repo = sorted(repo_set - binary_set)
    delta = len(binary_cmds) - len(repo_cmds)

    # --- Registry groups ---
    binary_groups = 0
    _repo_groups = 0
    registry_yaml = output_dir / "feature-registry.yaml"
    if registry_yaml.exists():
        text = registry_yaml.read_text(encoding="utf-8")
        in_groups = False
        for raw_line in text.splitlines():
            line = raw_line.rstrip()
            if line == "groups:":
                in_groups = True
                continue
            if not in_groups:
                continue
            # Group entries are 2-space indented, end with ':'
            if line.startswith("  ") and not line.startswith("    ") and line.rstrip().endswith(":"):
                # Determine source from notes field
                binary_groups += 1

        # For the comparison we count total groups; binary-enriched have "binary-symbols.txt" anchor
        binary_enriched = 0
        repo_scaffold = 0
        for raw_line in text.splitlines():
            stripped = raw_line.strip()
            if stripped == "- binary-symbols.txt":
                binary_enriched += 1

        # Groups without binary anchor are repo-scaffolded
        repo_scaffold = binary_groups - binary_enriched

    # --- Coverage percentage ---
    if repo_cmds:
        coverage_pct = round(len(binary_set & repo_set) / len(repo_set) * 100)
        coverage_line = f"Binary analysis found {coverage_pct}% of repo-discovered commands."
    elif binary_cmds:
        coverage_line = f"Binary analysis found {len(binary_cmds)} commands; repo analysis found none (no CLI detected in repo)."
    else:
        coverage_line = "Neither source discovered CLI commands."

    # --- Write report ---
    lines: list[str] = []
    lines.append(f"# Comparison Report: {product_name}")
    lines.append("")
    lines.append(f"**Date:** {date}")
    lines.append("**Mode:** both (binary + repo)")
    lines.append("")
    lines.append("## Command Discovery")
    lines.append("")
    lines.append("| Source | Commands Found |")
    lines.append("|--------|---------------|")
    lines.append(f"| Binary --help | {len(binary_cmds)} |")
    lines.append(f"| Repo analysis | {len(repo_cmds)} |")
    delta_str = f"+{delta}" if delta > 0 else str(delta)
    lines.append(f"| Delta | {delta_str} |")
    lines.append("")

    lines.append("## Commands Only in Binary")
    lines.append("")
    if only_binary:
        for cmd in only_binary:
            lines.append(f"- `{cmd}`")
    else:
        lines.append("_None._")
    lines.append("")

    lines.append("## Commands Only in Repo")
    lines.append("")
    if only_repo:
        for cmd in only_repo:
            lines.append(f"- `{cmd}`")
    else:
        lines.append("_None._")
    lines.append("")

    lines.append("## Registry Groups")
    lines.append("")
    lines.append("| Source | Groups |")
    lines.append("|--------|--------|")
    lines.append(f"| Binary enriched | {binary_enriched} |")
    lines.append(f"| Repo scaffold | {repo_scaffold} |")
    lines.append("")

    lines.append("## Summary")
    lines.append("")
    lines.append(coverage_line)

    out = output_dir / "comparison-report.md"
    out.write_text("\n".join(lines).rstrip() + "\n", encoding="utf-8")
    return True


def _write_wrapper_validate_feature_registry(output_dir: Path) -> None:
    skill_validate_path = (SKILL_DIR / "scripts" / "validate_feature_registry.py").resolve()
    wrapper = output_dir / "validate-feature-registry.py"
    wrapper.write_text(
        f"""#!/usr/bin/env python3
from __future__ import annotations

import os
import subprocess
import sys
from pathlib import Path

HERE = Path(__file__).resolve().parent
SKILL_VALIDATE_CANDIDATES = [
    Path({str(skill_validate_path)!r}),
    Path(__file__).resolve().parents[3] / "skills" / "reverse-engineer-rpi" / "scripts" / "validate_feature_registry.py",
    Path(__file__).resolve().parents[2] / "skills" / "reverse-engineer-rpi" / "scripts" / "validate_feature_registry.py",
    Path.cwd() / "skills" / "reverse-engineer-rpi" / "scripts" / "validate_feature_registry.py",
]

def _resolve_validator() -> Path:
    for cand in SKILL_VALIDATE_CANDIDATES:
        if cand.exists():
            return cand
    raise FileNotFoundError("Could not locate validate_feature_registry.py")

def main() -> int:
    # Delegate to the canonical validator, but default paths to this output dir.
    args = sys.argv[1:]
    if not args:
        root_path = HERE / "analysis-root-path.txt"
        local_root = (root_path.read_text(encoding="utf-8").strip() if root_path.exists() else str(HERE / "analysis-root"))
        args = [
            "--feature-registry", str(HERE / "feature-registry.yaml"),
            "--docs-features", str(HERE / "docs-features.txt"),
            "--local-clone-dir", local_root,
        ]
    validator = _resolve_validator()
    p = subprocess.run([sys.executable, str(validator), *args])
    return p.returncode

if __name__ == "__main__":
    raise SystemExit(main())
""",
        encoding="utf-8",
    )
    wrapper.chmod(0o755)


def _copy_security_validators(output_dir: Path) -> None:
    sec_dir = output_dir / "security"
    _ensure_dirs([sec_dir])

    # Copy validator + secret scan + sbom generator so the audit folder is self-validating.
    for rel in [
        "scripts/security/validate_security_audit.sh",
        "scripts/security/scan_secrets.sh",
        "scripts/security/generate_sbom.sh",
    ]:
        src = SKILL_DIR / rel
        dst = sec_dir / Path(rel).name.replace("_", "-")
        dst.write_text(src.read_text(encoding="utf-8"), encoding="utf-8")
        dst.chmod(0o755)


def main() -> int:
    ap = argparse.ArgumentParser(prog="reverse_engineer_rpi.py")
    ap.add_argument("product_name")
    ap.add_argument(
        "--authorized",
        action="store_true",
        help="Required for binary analysis. Confirms explicit written authorization to analyze the target binary.",
    )

    ap.add_argument("--docs-sitemap-url", default=None)
    ap.add_argument(
        "--docs-features-prefix",
        default="auto",
        help="Docs slug prefix, e.g. docs/features/. Use 'auto' to detect from repo/sitemap (default).",
    )
    ap.add_argument("--upstream-repo", default=None)
    ap.add_argument("--upstream-ref", default=None, help="Pin clone to a specific commit, tag, or branch. Records resolved SHA in clone-metadata.json.")
    ap.add_argument("--local-clone-dir", default=None)
    ap.add_argument("--output-dir", default=None)
    ap.add_argument("--mode", default="repo", choices=["repo", "binary", "both"])
    ap.add_argument("--binary-path", default=None)

    ap.add_argument("--security-audit", action="store_true")
    ap.add_argument("--sbom", action="store_true")
    ap.add_argument("--fuzz", action="store_true")
    ap.add_argument(
        "--materialize-archives",
        action="store_true",
        help="(Deprecated; now default in binary mode) Authorized-only: extract the best embedded ZIP candidate under local_clone_dir/extracted (do not commit).",
    )
    ap.add_argument(
        "--no-materialize-archives",
        action="store_true",
        help="Authorized-only: skip extracting embedded ZIP candidates (index-only).",
    )

    ap.add_argument("--beads", action="store_true", help="Optional: create bd epic/tasks for phases (requires bd).")

    args = ap.parse_args()

    product_slug = _slugify(args.product_name)
    local_clone_dir = Path(args.local_clone_dir or f".tmp/{product_slug}").resolve()
    output_dir = Path(args.output_dir or f".agents/research/{product_slug}/").resolve()
    analysis_root = local_clone_dir

    _ensure_dirs(
        [
            REPO_ROOT / ".agents" / "research",
            REPO_ROOT / ".agents" / "plans",
            REPO_ROOT / ".agents" / "council",
            REPO_ROOT / ".agents" / "rpi",
            REPO_ROOT / ".agents" / "learnings",
            REPO_ROOT / ".tmp",
            local_clone_dir,
            output_dir,
        ]
    )

    tmp_dir = (REPO_ROOT / ".tmp" / f"reverse-engineer-rpi-{product_slug}").resolve()
    _ensure_dirs([tmp_dir])

    # Optional AO context injection (best-effort; ignore failures).
    if shutil.which("ao"):
        subprocess.run(["ao", "search", args.product_name, "reverse engineering"], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        subprocess.run(["ao", "inject", args.product_name, "reverse engineering"], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)

    # Optional beads epic/tasks (off by default).
    if args.beads and shutil.which("bd"):
        _run(["bd", "ready"], check=False)

    docs_features_txt = output_dir / "docs-features.txt"
    effective_docs_prefix = args.docs_features_prefix

    # Acquire code (repo mode): shallow clone if requested.
    # NOTE: this must happen before docs inventory, otherwise docs/features extraction runs against an empty dir.
    if args.mode in ("repo", "both"):
        if args.upstream_repo and not (local_clone_dir / ".git").exists():
            clone_cmd = ["git", "clone"]
            if not args.upstream_ref:
                clone_cmd.append("--depth=1")
            clone_cmd.extend([args.upstream_repo, str(local_clone_dir)])
            _run(clone_cmd, check=True)
            if args.upstream_ref:
                _run(["git", "-C", str(local_clone_dir), "fetch", "--depth=1", "origin", args.upstream_ref], check=True)
                _run(["git", "-C", str(local_clone_dir), "checkout", "FETCH_HEAD"], check=True)
            # Record clone metadata for reproducibility.
            resolved_sha = subprocess.check_output(
                ["git", "-C", str(local_clone_dir), "rev-parse", "HEAD"], text=True,
            ).strip()
            clone_meta = {
                "upstream_repo": args.upstream_repo,
                "upstream_ref": args.upstream_ref,
                "resolved_commit": resolved_sha,
                "clone_date": _today_ymd(),
            }
            (output_dir / "clone-metadata.json").write_text(
                json.dumps(clone_meta, indent=2) + "\n", encoding="utf-8",
            )
            analysis_root = local_clone_dir

    # Determine an analysis root for repo mode.
    # Priority:
    # 1) local_clone_dir if it looks like a git checkout already
    # 2) git toplevel of the current working directory (if inside a repo)
    # 3) local_clone_dir (created)
    if args.mode in ("repo", "both"):
        if (local_clone_dir / ".git").exists():
            analysis_root = local_clone_dir
        else:
            try:
                top = subprocess.check_output(["git", "rev-parse", "--show-toplevel"], text=True).strip()
                if top:
                    analysis_root = Path(top).resolve()
            except Exception:
                analysis_root = local_clone_dir

    # 1) Mechanical docs inventory (NO heavy crawling).
    if args.docs_sitemap_url:
        sitemap_xml = tmp_dir / f"{product_slug}-sitemap.xml"
        _run([sys.executable, str(SKILL_DIR / "scripts" / "fetch_url.py"), args.docs_sitemap_url, str(sitemap_xml)])

        paths_txt = tmp_dir / f"{product_slug}-sitemap-paths.txt"
        sitemap_paths = subprocess.check_output([str(SKILL_DIR / "scripts" / "extract_sitemap_paths.sh"), str(sitemap_xml)], text=True)
        paths_txt.write_text(sitemap_paths, encoding="utf-8")

        if args.docs_features_prefix in ("", "auto"):
            effective_docs_prefix = _detect_docs_prefix_from_paths(sitemap_paths.splitlines())

        docs_features = subprocess.check_output(
            [
                str(SKILL_DIR / "scripts" / "extract_docs_features.sh"),
                str(paths_txt),
                effective_docs_prefix,
            ],
            text=True,
        )
        docs_features_txt.write_text(docs_features, encoding="utf-8")
    else:
        # No sitemap: for repo mode, inventory docs/features from the repo tree; otherwise empty.
        if args.mode in ("repo", "both") and analysis_root.exists():
            if args.docs_features_prefix in ("", "auto"):
                effective_docs_prefix = _detect_docs_prefix_for_repo(analysis_root)
            # Backward-compatibility fallback for explicit old default.
            elif args.docs_features_prefix == "docs/features/":
                if not (analysis_root / "docs" / "features").exists() and (analysis_root / "docs").exists():
                    effective_docs_prefix = "docs/"

            prefix_dir = effective_docs_prefix.strip("/").rstrip("/")
            base = analysis_root / prefix_dir
            slugs: list[str] = []
            if base.exists() and base.is_dir():
                for p in sorted(base.rglob("*")):
                    if not p.is_file():
                        continue
                    if p.suffix.lower() not in (".md", ".mdx"):
                        continue
                    rel = p.relative_to(analysis_root).as_posix()
                    # Normalize to slug without extension to match sitemap-style slugs.
                    slugs.append(rel[: -len(p.suffix)])
            docs_features_txt.write_text("\n".join(slugs) + ("\n" if slugs else ""), encoding="utf-8")
        else:
            docs_features_txt.write_text("", encoding="utf-8")

    # 2) Binary analysis mode.
    if args.mode in ("binary", "both"):
        if not args.authorized:
            _die("--authorized is required for binary analysis (hard guardrail)")
        if not args.binary_path:
            _die("--binary-path is required when --mode includes binary")
        binary_path = Path(args.binary_path).expanduser().resolve()
        if not binary_path.exists():
            _die(f"binary not found: {binary_path}")

        _ensure_dirs([tmp_dir / "binary"])

        _run(
            [
                str(SKILL_DIR / "scripts" / "binary" / "analyze_binary.sh"),
                str(binary_path),
                str(tmp_dir / "binary"),
            ],
            check=True,
        )
        ba = tmp_dir / "binary" / "binary-analysis.md"
        if ba.exists():
            shutil.copyfile(ba, output_dir / "binary-analysis.md")

        # After analyze_binary.sh completes, run CLI help capture
        capture_script = SKILL_DIR / "scripts" / "binary" / "capture_cli_help.sh"
        if capture_script.exists():
            _run(
                ["bash", str(capture_script), str(binary_path), str(tmp_dir / "binary")],
                check=False,  # best-effort
            )
            # Copy results to output dir if they exist
            for fname in ("cli-help-tree.txt", "cli-commands.txt"):
                src = tmp_dir / "binary" / fname
                if src.exists():
                    shutil.copyfile(src, output_dir / fname)

        # Embedded archive inventory (index only by default; does not dump content into output_dir).
        _run(
            [
                sys.executable,
                str(SKILL_DIR / "scripts" / "binary" / "list_embedded_archives.py"),
                "--binary",
                str(binary_path),
                "--out-json",
                str(tmp_dir / "binary" / "embedded-archives.json"),
                "--out-index-md",
                str(output_dir / "binary-embedded-archives.md"),
            ],
            check=True,
        )

        # Default: materialize archives in binary mode (must-have workflow), unless explicitly disabled.
        if args.no_materialize_archives and args.materialize_archives:
            _die("flags conflict: --materialize-archives and --no-materialize-archives")

        if not args.no_materialize_archives:
            extract_root = local_clone_dir / "extracted"
            _ensure_dirs([extract_root])
            _run(
                [
                    sys.executable,
                    str(SKILL_DIR / "scripts" / "binary" / "extract_embedded_archives.py"),
                    "--binary",
                    str(binary_path),
                    "--out-dir",
                    str(extract_root),
                ],
                check=True,
            )
            primary = extract_root / "PRIMARY.txt"
            if primary.exists():
                analysis_root = Path(primary.read_text(encoding="utf-8").strip())

    # 4) Generate feature inventory (docs-first when available).
    inventory_md = output_dir / "feature-inventory.md"
    _run(
        [
            sys.executable,
            str(SKILL_DIR / "scripts" / "generate_feature_inventory_md.py"),
            "--product-name",
            args.product_name,
            "--docs-features",
            str(docs_features_txt),
            "--out",
            str(inventory_md),
        ],
        check=True,
    )

    # 5) Registry-first mapping.
    registry_yaml = output_dir / "feature-registry.yaml"
    _run(
        [
            sys.executable,
            str(SKILL_DIR / "scripts" / "scaffold_feature_registry.py"),
            "--product-name",
            args.product_name,
            "--docs-features-prefix",
            effective_docs_prefix,
            "--docs-features",
            str(docs_features_txt),
            "--out",
            str(registry_yaml),
        ],
        check=True,
    )

    # 5b) Enrich registry with binary evidence when available.
    if args.mode in ("binary", "both"):
        _enrich_registry_with_binary_evidence(
            registry_yaml,
            tmp_dir,
            output_dir,
            product_name=args.product_name,
            date=_today_ymd(),
        )

    catalog_md = output_dir / "feature-catalog.md"
    _run(
        [
            sys.executable,
            str(SKILL_DIR / "scripts" / "generate_feature_catalog_md.py"),
            "--registry",
            str(registry_yaml),
            "--out",
            str(catalog_md),
        ],
        check=True,
    )

    # 6) Specs (template render).
    vars = {"PRODUCT_NAME": args.product_name, "DATE": _today_ymd()}
    for tmpl, out_name in [
        ("spec-architecture.md.tmpl", "spec-architecture.md"),
        ("spec-code-map.md.tmpl", "spec-code-map.md"),
        ("spec-clone-vs-use.md.tmpl", "spec-clone-vs-use.md"),
        ("spec-clone-mvp.md.tmpl", "spec-clone-mvp.md"),
    ]:
        _render_template(TEMPLATES_DIR / tmpl, output_dir / out_name, vars)

    # CLI surface is optional; only write spec-cli-surface.md if a CLI is detected.
    wrote_cli = False
    if args.mode in ("repo", "both"):
        wrote_cli = _write_cli_surface_spec(
            output_dir,
            product_name=args.product_name,
            product_slug=product_slug,
            date=vars["DATE"],
            analysis_root=analysis_root,
        )

    if not wrote_cli and args.mode in ("binary", "both"):
        wrote_cli = _write_binary_cli_surface_spec(
            output_dir,
            tmp_dir,
            product_name=args.product_name,
            date=vars["DATE"],
        )

    if not wrote_cli:
        # Required behavior: omit the file, but leave an explicit note somewhere deterministic.
        (output_dir / "spec-code-map.md").write_text(
            (output_dir / "spec-code-map.md").read_text(encoding="utf-8")
            + "\n\n## CLI Surface\n\n_Omitted: no CLI surface detected (or mode did not include repo)._ \n",
            encoding="utf-8",
        )

    # Artifact surface: best-effort extraction of what the product writes/installs (high-fidelity cloning aid).
    if args.mode in ("repo", "both"):
        _write_artifact_surface_spec(
            output_dir,
            product_name=args.product_name,
            product_slug=product_slug,
            date=vars["DATE"],
            analysis_root=analysis_root,
        )

    # 6b) Deterministic repo-mode contract JSON (CLI/config/env + artifact I/O surface).
    if args.mode in ("repo", "both"):
        _write_repo_contract_json(
            output_dir,
            analysis_root,
            product_name=args.product_name,
            product_slug=product_slug,
        )

    # 6c) Comparison report (binary vs repo) when both sources are available.
    if args.mode == "both":
        _write_comparison_report(output_dir, tmp_dir, product_name=args.product_name, date=_today_ymd())

    # 7) Validation gate: produce a self-contained validator in the output dir and run it once.
    _write_wrapper_validate_feature_registry(output_dir)
    # Store analysis root pointer for validators (repo clone dir or a placeholder).
    (output_dir / "analysis-root").mkdir(exist_ok=True)
    (output_dir / "analysis-root-path.txt").write_text(str(analysis_root), encoding="utf-8")
    # Keep docs-features alongside outputs for deterministic validation.
    # (Already written as output_dir/docs-features.txt)
    _run(
        [
            sys.executable,
            str(SKILL_DIR / "scripts" / "validate_feature_registry.py"),
            "--feature-registry",
            str(registry_yaml),
            "--docs-features",
            str(docs_features_txt),
            "--local-clone-dir",
            str(analysis_root if analysis_root.exists() else output_dir / "analysis-root"),
        ],
        check=True,
    )

    # 8) Security audit artifacts + gates.
    if args.security_audit:
        sec_dir = output_dir / "security"
        _ensure_dirs([sec_dir])
        for name in [
            "threat-model.md.tmpl",
            "attack-surface.md.tmpl",
            "dataflow.md.tmpl",
            "crypto-review.md.tmpl",
            "authn-authz.md.tmpl",
            "findings.md.tmpl",
            "reproducibility.md.tmpl",
        ]:
            _render_template(TEMPLATES_DIR / "security" / name, sec_dir / name.replace(".tmpl", ""), vars)

        _copy_security_validators(output_dir)

        if args.sbom:
            _run([str(sec_dir / "generate-sbom.sh"), str(analysis_root), str(sec_dir)], check=False)

        # Run validation gate (includes secret scan over output_dir).
        _run([str(sec_dir / "validate-security-audit.sh"), str(output_dir), "--sbom" if args.sbom else "--no-sbom"], check=True)

    # 9) Reports (vibe-style + post-mortem) + learning.
    council_dir = REPO_ROOT / ".agents" / "council"
    _ensure_dirs([council_dir])
    vibe_path = council_dir / f"{_today_ymd()}-vibe-{product_slug}.md"
    post_path = council_dir / f"{_today_ymd()}-post-mortem-{product_slug}.md"

    _render_template(TEMPLATES_DIR / "vibe-report.md.tmpl", vibe_path, {**vars, "OUTPUT_DIR": str(output_dir)})
    _render_template(TEMPLATES_DIR / "post-mortem.md.tmpl", post_path, {**vars, "OUTPUT_DIR": str(output_dir)})

    learning_path = REPO_ROOT / ".agents" / "learnings" / f"{_today_ymd()}-{product_slug}-reverse-engineer-rpi.md"
    if not learning_path.exists():
        learning_path.write_text(
            f"# Learning ({_today_ymd()}): reverse-engineer-rpi\n\n"
            f"- Keep docs-derived inventory separate from code/binary evidence; treat hosted/control-plane as unknown until proven.\n",
            encoding="utf-8",
        )

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
```

### scaffold_feature_registry.py

```python
#!/usr/bin/env python3
from __future__ import annotations

import argparse
import datetime as _dt
from pathlib import Path


def _group_from_slug(slug: str, docs_features_prefix: str) -> str | None:
    prefix = docs_features_prefix.strip("/").rstrip("/") + "/"
    s = slug.strip().lstrip("/")
    if not s.startswith(prefix):
        return None
    rest = s[len(prefix) :]
    if not rest:
        return None
    group = rest.split("/", 1)[0].strip()
    return group or None


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--product-name", required=True)
    ap.add_argument("--docs-features-prefix", required=True)
    ap.add_argument("--docs-features", required=True)
    ap.add_argument("--out", required=True)
    args = ap.parse_args()

    docs_features_prefix = args.docs_features_prefix
    slugs = [ln.strip() for ln in Path(args.docs_features).read_text(encoding="utf-8", errors="replace").splitlines() if ln.strip()]

    groups: list[str] = []
    seen = set()
    for slug in slugs:
        g = _group_from_slug(slug, docs_features_prefix)
        if not g:
            continue
        if g not in seen:
            groups.append(g)
            seen.add(g)

    out = Path(args.out)
    out.parent.mkdir(parents=True, exist_ok=True)

    # Minimal YAML that is still easy to mechanically validate.
    lines: list[str] = []
    lines.append("schema_version: 1")
    lines.append(f"product_name: {args.product_name!r}")
    lines.append(f"generated_at: {_dt.date.today().isoformat()!r}")
    lines.append(f"docs_features_prefix: {docs_features_prefix!r}")
    lines.append("docs_features:")
    for s in slugs:
        lines.append(f"  - {s!r}")
    lines.append("groups:")
    if not groups and slugs:
        # Slugs existed but no groups parsed; keep explicit empty mapping to fail validation loudly later.
        lines.append("  {}")
    elif not groups:
        lines.append("  {}")
    else:
        for g in groups:
            lines.append(f"  {g!s}:")
            lines.append("    impl: control-plane")
            lines.append("    anchors: []")
            lines.append("    notes: \"\"")

    out.write_text("\n".join(lines) + "\n", encoding="utf-8")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
```

### self_test.sh

```bash
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
  --materialize-archives \
  --local-clone-dir="$TMP/local-demo" \
  --output-dir="$OUT1"

python3 "$OUT1/validate-feature-registry.py"

# --- Binary mode capability assertions ---

echo "--- binary mode capability checks ---"

# 1. Help capture output exists (may be empty if binary doesn't support --help)
if [ ! -f "$OUT1/cli-commands.txt" ]; then
  echo "FAIL: cli-commands.txt not created by binary mode" >&2
  exit 1
fi
echo "OK: cli-commands.txt exists"

# 2. CLI surface spec exists (generated from --help tree or binary strings fallback)
if [ ! -f "$OUT1/spec-cli-surface.md" ]; then
  echo "FAIL: spec-cli-surface.md not created by binary mode" >&2
  exit 1
fi
echo "OK: spec-cli-surface.md exists"

# 3. binary-symbols.txt exists
if [ ! -f "$OUT1/binary-symbols.txt" ]; then
  echo "FAIL: binary-symbols.txt not created by binary mode" >&2
  exit 1
fi
echo "OK: binary-symbols.txt exists"

# 4. Registry enrichment: if cli-commands.txt has content, groups should have impl: client
if [ -s "$OUT1/cli-commands.txt" ]; then
  if ! grep -q 'impl: client' "$OUT1/feature-registry.yaml"; then
    echo "FAIL: feature-registry.yaml should contain 'impl: client' when CLI commands are found" >&2
    exit 1
  fi
  echo "OK: feature-registry.yaml enriched with impl: client"
else
  echo "OK: cli-commands.txt empty (demo binary has no subcommands); skipping impl: client check"
fi

python3 "$SKILL/scripts/reverse_engineer_rpi.py" demo \
  --authorized \
  --mode=binary \
  --binary-path="$BIN" \
  --docs-sitemap-url="file://$SITEMAP" \
  --output-dir="$OUT2" \
  --materialize-archives \
  --local-clone-dir="$TMP/local-demo" \
  --security-audit \
  --sbom

"$OUT2/security/validate-security-audit.sh" "$OUT2" --sbom

# --- Negative tests ---

# Test: invalid --mode should fail
echo "--- negative test: invalid --mode ---"
if python3 "$SKILL/scripts/reverse_engineer_rpi.py" demo --mode=invalid --output-dir="$TMP/out-neg" 2>/dev/null; then
  echo "FAIL: expected non-zero exit for --mode=invalid" >&2
  exit 1
fi
echo "OK: invalid --mode correctly rejected"

# --- Upstream ref pinning test ---

echo "--- upstream-ref pinning test ---"
OUT_REF="$TMP/out-ref"
mkdir -p "$OUT_REF"
# Use file:// protocol on the current repo to avoid network dependency.
REPO_URL="file://$ROOT"
python3 "$SKILL/scripts/reverse_engineer_rpi.py" self-ref-test \
  --mode=repo \
  --upstream-repo="$REPO_URL" \
  --upstream-ref=HEAD \
  --local-clone-dir="$TMP/local-ref" \
  --output-dir="$OUT_REF"

if [ ! -f "$OUT_REF/clone-metadata.json" ]; then
  echo "FAIL: clone-metadata.json not created with --upstream-ref" >&2
  exit 1
fi
echo "OK: clone-metadata.json created with --upstream-ref"

# --- Multi-language CLI graceful degradation test ---

echo "--- multi-language CLI degradation test ---"
OUT_NONCLI="$TMP/out-noncli"
mkdir -p "$OUT_NONCLI" "$TMP/local-noncli"
# Create a minimal repo with no CLI markers.
mkdir -p "$TMP/local-noncli/.git"
touch "$TMP/local-noncli/README.md"
python3 "$SKILL/scripts/reverse_engineer_rpi.py" no-cli-demo \
  --mode=repo \
  --local-clone-dir="$TMP/local-noncli" \
  --output-dir="$OUT_NONCLI" \
  --docs-sitemap-url="file://$SITEMAP"

# spec-cli-surface.md should NOT exist (no CLI detected), and the note should be in spec-code-map.md
if [ -f "$OUT_NONCLI/spec-cli-surface.md" ]; then
  echo "FAIL: spec-cli-surface.md should not exist for non-CLI repo" >&2
  exit 1
fi
if ! grep -q "no CLI surface detected" "$OUT_NONCLI/spec-code-map.md" 2>/dev/null; then
  echo "FAIL: spec-code-map.md should note that no CLI surface was detected" >&2
  exit 1
fi
echo "OK: multi-language CLI graceful degradation works"

echo "OK: self-test passed (all positive + negative tests)"
```

### validate.sh

```bash
#!/usr/bin/env bash
set -euo pipefail

SKILL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

python3 -m py_compile \
  "$SKILL_DIR/scripts/reverse_engineer_rpi.py" \
  "$SKILL_DIR/scripts/fetch_url.py" \
  "$SKILL_DIR/scripts/generate_feature_inventory_md.py" \
  "$SKILL_DIR/scripts/scaffold_feature_registry.py" \
  "$SKILL_DIR/scripts/generate_feature_catalog_md.py" \
  "$SKILL_DIR/scripts/validate_feature_registry.py" \
  "$SKILL_DIR/scripts/binary/list_embedded_archives.py" \
  "$SKILL_DIR/scripts/binary/extract_embedded_archives.py"

echo "OK: reverse-engineer-rpi validate.sh passed"
```

### validate_feature_registry.py

```python
#!/usr/bin/env python3
from __future__ import annotations

import argparse
import os
import re
import sys
from pathlib import Path


ALLOWED_IMPL = {"client", "mixed", "control-plane"}


def _group_from_slug(slug: str, docs_features_prefix: str) -> str | None:
    prefix = docs_features_prefix.strip("/").rstrip("/") + "/"
    s = slug.strip().lstrip("/")
    if not s.startswith(prefix):
        return None
    rest = s[len(prefix) :]
    if not rest:
        return None
    return rest.split("/", 1)[0] or None


def _parse_registry(path: Path) -> dict:
    data = {"docs_features_prefix": "docs/features/", "groups": {}}
    cur = None
    in_groups = False
    in_anchors = False
    for raw in path.read_text(encoding="utf-8", errors="replace").splitlines():
        line = raw.rstrip("\n")
        if not line.strip() or line.lstrip().startswith("#"):
            continue
        if line.startswith("docs_features_prefix:"):
            data["docs_features_prefix"] = line.split(":", 1)[1].strip().strip("'\"")
        if line == "groups:":
            in_groups = True
            continue
        if not in_groups:
            continue

        if line.startswith("  ") and not line.startswith("    ") and line.endswith(":"):
            name = line.strip()[:-1]
            cur = {"impl": None, "anchors": [], "notes": ""}
            data["groups"][name] = cur
            in_anchors = False
            continue

        if cur is None:
            continue

        s = line.strip()
        if s.startswith("impl:"):
            cur["impl"] = s.split(":", 1)[1].strip()
        elif s.startswith("anchors:"):
            in_anchors = True
            if s.endswith("[]"):
                cur["anchors"] = []
        elif in_anchors and s.startswith("- "):
            cur["anchors"].append(s[2:].strip().strip("'\""))
        elif s.startswith("notes:"):
            cur["notes"] = s.split(":", 1)[1].strip().strip("'\"")
    return data


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--feature-registry", required=True)
    ap.add_argument("--docs-features", required=True)
    ap.add_argument("--local-clone-dir", required=True)
    args = ap.parse_args()

    feature_registry_path = Path(args.feature_registry).resolve()
    artifact_dir = feature_registry_path.parent
    reg = _parse_registry(feature_registry_path)
    prefix = reg["docs_features_prefix"]
    groups = reg["groups"]
    docs_slugs = [ln.strip() for ln in Path(args.docs_features).read_text(encoding="utf-8", errors="replace").splitlines() if ln.strip()]
    root = Path(args.local_clone_dir).resolve()

    errs: list[str] = []

    # Rule: every docs/features slug maps to a group.
    for slug in docs_slugs:
        g = _group_from_slug(slug, prefix)
        if not g:
            errs.append(f"docs slug not under prefix {prefix!r}: {slug!r}")
            continue
        if g not in groups:
            errs.append(f"docs slug group missing from registry: group={g!r} slug={slug!r}")

    # Rule: every group has impl; client/mixed must have anchors.
    for g, ent in groups.items():
        impl = (ent.get("impl") or "").strip()
        if impl not in ALLOWED_IMPL:
            errs.append(f"group {g!r} has invalid impl {impl!r} (allowed: {sorted(ALLOWED_IMPL)})")
        anchors = ent.get("anchors") or []
        if impl in ("client", "mixed") and len(anchors) < 1:
            errs.append(f"group {g!r} impl={impl!r} requires >=1 anchor")

        for a in anchors:
            # Allow line/col suffix like "path/to/file.py:123"
            p = a.split(":", 1)[0]
            if p.startswith("/"):
                abs_path = Path(p).resolve()
                if not abs_path.exists():
                    errs.append(f"group {g!r} anchor missing: {a!r} (checked {abs_path})")
                continue

            # Relative anchors may reference either the analysis root or the artifact bundle dir.
            candidates = [
                (artifact_dir, (artifact_dir / p).resolve(), "artifact_dir"),
                (root, (root / p).resolve(), "analysis_root"),
            ]
            path_ok = False
            missing_paths: list[str] = []
            for base, resolved, _label in candidates:
                base_resolved = base.resolve()
                if not (resolved == base_resolved or str(resolved).startswith(str(base_resolved) + os.sep)):
                    continue
                if resolved.exists():
                    path_ok = True
                    break
                missing_paths.append(str(resolved))

            if not path_ok:
                checked = ", ".join(missing_paths) if missing_paths else "(no safe candidate paths)"
                errs.append(f"group {g!r} anchor missing: {a!r} (checked {checked})")

    # Completeness guard: if docs/ exists with markdown content, empty docs feature inventory is likely bad prefix selection.
    docs_dir = root / "docs"
    if docs_dir.exists():
        has_docs_markdown = any(docs_dir.rglob("*.md")) or any(docs_dir.rglob("*.mdx"))
        if has_docs_markdown and len(docs_slugs) == 0:
            errs.append(
                "docs-features inventory is empty while docs/ contains markdown; "
                "likely wrong docs_features_prefix or extraction failure"
            )

    # Completeness guard: reject unresolved placeholder code-map specs.
    spec_code_map = artifact_dir / "spec-code-map.md"
    if spec_code_map.exists():
        text = spec_code_map.read_text(encoding="utf-8", errors="replace")
        if "_TBD_" in text or re.search(r"\|\s*_TBD_\s*\|", text):
            errs.append(f"spec-code-map contains unresolved placeholders: {spec_code_map}")

    if errs:
        for e in errs:
            print(f"FAIL: {e}", file=sys.stderr)
        return 1
    print("OK: feature registry validated")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
```


