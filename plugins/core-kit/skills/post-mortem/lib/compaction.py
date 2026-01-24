#!/usr/bin/env python3
"""
Compaction chain traversal for post-mortem session tracing.

Claude Code compacts context by:
1. Creating a summary message
2. Starting fresh with the summary as context
3. The old transcript continues but is "sealed"

This module finds and merges all transcript files in a compaction chain.
"""
import json
from dataclasses import dataclass, field
from pathlib import Path
from typing import Optional

# Handle both package and standalone imports
try:
    from .transcript import SessionData, parse_transcript
except ImportError:
    from transcript import SessionData, parse_transcript


@dataclass
class CompactionChain:
    """A chain of transcript files across compactions."""
    files: list[Path] = field(default_factory=list)
    compaction_points: list[int] = field(default_factory=list)  # Line numbers
    total_lines: int = 0


def find_compaction_chain(transcript_path: Path) -> list[Path]:
    """
    Find all transcript files in the compaction chain.

    Claude Code creates continuation files when context is compacted:
    - session-abc.jsonl (compacted at line N)
    - A new session file starts with the summary

    However, Claude Code actually creates NEW session files for continuations,
    not numbered suffixes. We detect chains by:
    1. Looking at session metadata (parentUuid, slug)
    2. Finding files with matching slug in same directory

    Returns all files in order (oldest first).
    """
    if not transcript_path.exists():
        return []

    chain = [transcript_path]
    project_dir = transcript_path.parent

    # Extract session metadata to find related sessions
    session_slug = _extract_session_slug(transcript_path)

    if session_slug:
        # Find other sessions with same slug (indicates continuation)
        for other_file in project_dir.glob("*.jsonl"):
            if other_file == transcript_path:
                continue

            other_slug = _extract_session_slug(other_file)
            if other_slug == session_slug:
                chain.append(other_file)

    # Sort by modification time (oldest first for proper order)
    chain.sort(key=lambda p: p.stat().st_mtime)

    return chain


def _extract_session_slug(transcript_path: Path) -> Optional[str]:
    """Extract the session slug from a transcript file."""
    try:
        with open(transcript_path, 'r') as f:
            for line in f:
                if not line.strip():
                    continue
                try:
                    entry = json.loads(line)
                    slug = entry.get("slug")
                    if slug:
                        return slug
                except json.JSONDecodeError:
                    continue
                # Only check first 10 lines for performance
                break
    except Exception:
        pass
    return None


def build_composite_session(chain: list[Path]) -> SessionData:
    """
    Merge multiple transcript files into unified view.

    Handles:
    - Deduplication of operations
    - Summary message insertion points
    - Time continuity across compactions
    """
    if not chain:
        return SessionData(session_id="", project_path="")

    # Parse each transcript
    sessions = [parse_transcript(path) for path in chain]

    # Merge into composite
    composite = SessionData(
        session_id=sessions[0].session_id,
        project_path=sessions[0].project_path,
    )

    seen_operations: set[tuple[str, str, str]] = set()  # (issue_id, operation, command)

    for session in sessions:
        # Merge beads operations (deduplicate)
        for op in session.beads_operations:
            key = (op.issue_id, op.operation, op.command)
            if key not in seen_operations:
                seen_operations.add(key)
                composite.beads_operations.append(op)

        # Merge IDs (sets handle dedup)
        composite.epic_ids.update(session.epic_ids)
        composite.issue_ids.update(session.issue_ids)

        # Merge file changes (may have dups, but that's informative)
        composite.file_changes.extend(session.file_changes)

        # Merge commits
        composite.commits_made.extend(session.commits_made)

        # Track compaction points
        if session.compaction_markers:
            composite.compaction_markers.extend(session.compaction_markers)

    return composite


def is_compaction_boundary(message: dict) -> bool:
    """
    Detect compaction summary message.

    Compaction summaries have specific markers in their content.
    """
    content = message.get("content", "")

    # Check string content
    if isinstance(content, str):
        return _is_compaction_text(content)

    # Check array content
    if isinstance(content, list):
        for item in content:
            if item.get("type") == "text":
                if _is_compaction_text(item.get("text", "")):
                    return True

    return False


def _is_compaction_text(text: str) -> bool:
    """Check if text contains compaction markers."""
    markers = [
        "session is being continued",
        "summary below covers",
        "context has been compacted",
        "continuing from previous context",
        "this is a continuation",
    ]
    text_lower = text.lower()
    return any(marker in text_lower for marker in markers)


def find_all_project_sessions(project_path: str) -> dict[str, list[Path]]:
    """
    Find all session chains for a project.

    Returns dict mapping slug -> list of transcript files
    """
    from .transcript import find_project_directory

    project_dir = find_project_directory(project_path)
    if not project_dir:
        return {}

    chains: dict[str, list[Path]] = {}

    for transcript in project_dir.glob("*.jsonl"):
        slug = _extract_session_slug(transcript)
        if slug:
            if slug not in chains:
                chains[slug] = []
            chains[slug].append(transcript)

    # Sort each chain by time
    for slug in chains:
        chains[slug].sort(key=lambda p: p.stat().st_mtime)

    return chains


def main():
    """CLI entry point."""
    import argparse

    parser = argparse.ArgumentParser(description="Handle compaction chains")
    parser.add_argument("--transcript", help="Path to transcript file")
    parser.add_argument("--find-chain", action="store_true", help="Find compaction chain")
    parser.add_argument("--build-composite", action="store_true", help="Build composite session")
    parser.add_argument("--verbose", "-v", action="store_true", help="Verbose output")

    args = parser.parse_args()

    if args.transcript:
        transcript_path = Path(args.transcript)

        if args.find_chain:
            chain = find_compaction_chain(transcript_path)
            print(f"Chain has {len(chain)} file(s):")
            for f in chain:
                print(f"  {f}")

        if args.build_composite:
            chain = find_compaction_chain(transcript_path)
            composite = build_composite_session(chain)
            print(f"Composite session:")
            print(f"  Epic IDs: {composite.epic_ids}")
            print(f"  Issue IDs: {composite.issue_ids}")
            print(f"  Beads Operations: {len(composite.beads_operations)}")
            print(f"  File Changes: {len(composite.file_changes)}")
            print(f"  Compaction Points: {composite.compaction_markers}")


if __name__ == "__main__":
    main()
