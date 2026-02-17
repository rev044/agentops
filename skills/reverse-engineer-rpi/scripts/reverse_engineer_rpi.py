#!/usr/bin/env python3
from __future__ import annotations

import argparse
import datetime as _dt
import os
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


def _render_template(src: Path, dst: Path, vars: dict[str, str]) -> None:
    text = src.read_text(encoding="utf-8")
    for k, v in vars.items():
        text = text.replace("{{" + k + "}}", v)
    dst.write_text(text, encoding="utf-8")


def _write_wrapper_validate_feature_registry(output_dir: Path) -> None:
    wrapper = output_dir / "validate-feature-registry.py"
    wrapper.write_text(
        """#!/usr/bin/env python3
from __future__ import annotations

import os
import subprocess
import sys
from pathlib import Path

HERE = Path(__file__).resolve().parent
SKILL_VALIDATE = Path(__file__).resolve().parents[3] / "skills" / "reverse-engineer-rpi" / "scripts" / "validate_feature_registry.py"

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
    p = subprocess.run([sys.executable, str(SKILL_VALIDATE), *args])
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
    ap.add_argument("--docs-features-prefix", default="docs/features/")
    ap.add_argument("--upstream-repo", default=None)
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
        paths_txt.write_text(
            subprocess.check_output([str(SKILL_DIR / "scripts" / "extract_sitemap_paths.sh"), str(sitemap_xml)], text=True),
            encoding="utf-8",
        )

        docs_features = subprocess.check_output(
            [
                str(SKILL_DIR / "scripts" / "extract_docs_features.sh"),
                str(paths_txt),
                args.docs_features_prefix,
            ],
            text=True,
        )
        docs_features_txt.write_text(docs_features, encoding="utf-8")
    else:
        # No sitemap: for repo mode, inventory docs/features from the repo tree; otherwise empty.
        if args.mode in ("repo", "both") and analysis_root.exists():
            prefix_dir = args.docs_features_prefix.strip("/").rstrip("/")
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

    # 3) Acquire code (repo mode): shallow clone if requested.
    if args.mode in ("repo", "both"):
        if args.upstream_repo and not (local_clone_dir / ".git").exists():
            _run(["git", "clone", "--depth=1", args.upstream_repo, str(local_clone_dir)], check=True)
            analysis_root = local_clone_dir

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
            args.docs_features_prefix,
            "--docs-features",
            str(docs_features_txt),
            "--out",
            str(registry_yaml),
        ],
        check=True,
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

    # CLI surface is optional; generate skeleton only if it looks like a CLI exists in repo mode.
    cli_tmpl = TEMPLATES_DIR / "spec-cli-surface.md.tmpl"
    wrote_cli = False
    if args.mode in ("repo", "both") and local_clone_dir.exists():
        maybe_cli = any((local_clone_dir / p).exists() for p in ["cmd", "cli", "bin"]) or (local_clone_dir / "go.mod").exists()
        if maybe_cli:
            _render_template(cli_tmpl, output_dir / "spec-cli-surface.md", vars)
            wrote_cli = True
    if not wrote_cli:
        # Required behavior: omit the file, but leave an explicit note somewhere deterministic.
        (output_dir / "spec-code-map.md").write_text(
            (output_dir / "spec-code-map.md").read_text(encoding="utf-8") + "\n\n## CLI Surface\n\n_Omitted: no CLI surface detected (or mode did not include repo)._ \n",
            encoding="utf-8",
        )

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
