#!/usr/bin/env python3
"""
Transcript parser for post-mortem session tracing.

Parses Claude Code JSONL transcripts to extract:
- Beads operations (bd update, bd close, bd show)
- Epic IDs worked on during session
- File changes and commits
- Compaction markers
"""
import json
import re
from dataclasses import dataclass, field
from pathlib import Path
from typing import Optional


@dataclass
class BeadsOperation:
    """A beads operation from the session."""
    command: str
    issue_id: str
    operation: str  # update, close, show, etc.
    timestamp: str


@dataclass
class SessionData:
    """Parsed session data."""
    session_id: str
    project_path: str
    beads_operations: list[BeadsOperation] = field(default_factory=list)
    epic_ids: set[str] = field(default_factory=set)
    issue_ids: set[str] = field(default_factory=set)
    file_changes: list[str] = field(default_factory=list)
    commits_made: list[str] = field(default_factory=list)
    compaction_markers: list[int] = field(default_factory=list)  # Line numbers


# Pattern to extract issue IDs (prefix-xxxx format)
ISSUE_ID_PATTERN = re.compile(r'\b([a-z]{2,4}-[a-z0-9]{3,6})\b', re.IGNORECASE)

# Beads command patterns
BD_UPDATE_PATTERN = re.compile(r'bd\s+update\s+([a-z]{2,4}-[a-z0-9]{3,6})', re.IGNORECASE)
BD_CLOSE_PATTERN = re.compile(r'bd\s+close\s+([a-z]{2,4}-[a-z0-9]{3,6})', re.IGNORECASE)
BD_SHOW_PATTERN = re.compile(r'bd\s+show\s+([a-z]{2,4}-[a-z0-9]{3,6})', re.IGNORECASE)

# Epic detection patterns (usually have .1, .2 children or are type=epic)
EPIC_CHILD_PATTERN = re.compile(r'([a-z]{2,4}-[a-z0-9]{3,6})\.\d+', re.IGNORECASE)


def find_project_directory(project_path: str) -> Optional[Path]:
    """
    Find the Claude projects directory for a given project path.

    Claude stores transcripts in ~/.claude/projects/{project-path-with-dashes}/
    """
    claude_projects = Path.home() / ".claude" / "projects"
    if not claude_projects.exists():
        return None

    # Convert path to the format Claude uses (dashes for slashes)
    normalized = project_path.replace("/", "-")
    if normalized.startswith("-"):
        normalized = normalized[1:]  # Remove leading dash

    # Look for matching directory
    for d in claude_projects.iterdir():
        if d.is_dir() and normalized in d.name:
            return d

    # Try partial match
    path_parts = normalized.split("-")
    for d in claude_projects.iterdir():
        if d.is_dir():
            # Check if key parts match
            d_parts = d.name.split("-")
            if all(p in d_parts for p in path_parts[-3:]):  # Match last 3 parts
                return d

    return None


def find_session_transcripts(project_path: str) -> list[Path]:
    """
    Find all transcript files for current project.

    Location: ~/.claude/projects/{project-path-with-dashes}/*.jsonl
    Returns: Sorted by modification time (newest first)
    """
    project_dir = find_project_directory(project_path)
    if not project_dir:
        return []

    transcripts = list(project_dir.glob("*.jsonl"))
    # Sort by modification time, newest first
    transcripts.sort(key=lambda p: p.stat().st_mtime, reverse=True)
    return transcripts


def parse_transcript(transcript_path: Path) -> SessionData:
    """
    Parse JSONL transcript and extract structured data.

    Returns:
        SessionData with:
        - beads_operations: List of bd commands and results
        - file_changes: Files written/edited
        - commits_made: Git commits during session
        - epic_ids: All epic IDs referenced
        - issue_ids: All issue IDs referenced
        - compaction_markers: Where compactions occurred
    """
    session_data = SessionData(
        session_id=transcript_path.stem,
        project_path=str(transcript_path.parent),
    )

    try:
        with open(transcript_path, 'r') as f:
            for line_num, line in enumerate(f, 1):
                if not line.strip():
                    continue
                try:
                    entry = json.loads(line)
                    _process_entry(entry, session_data, line_num)
                except json.JSONDecodeError:
                    continue
    except Exception as e:
        print(f"Error parsing transcript: {e}")

    return session_data


def _process_entry(entry: dict, session_data: SessionData, line_num: int):
    """Process a single JSONL entry."""
    entry_type = entry.get("type", "")

    # Check for compaction markers
    if _is_compaction_boundary(entry):
        session_data.compaction_markers.append(line_num)
        return

    # Process assistant messages with tool calls
    if entry_type == "assistant":
        message = entry.get("message", {})
        content = message.get("content", [])
        timestamp = entry.get("timestamp", "")

        for item in content:
            if item.get("type") == "tool_use":
                tool_name = item.get("name", "")
                tool_input = item.get("input", {})

                if tool_name == "Bash":
                    command = tool_input.get("command", "")
                    _process_bash_command(command, session_data, timestamp)
                elif tool_name in ("Write", "Edit"):
                    file_path = tool_input.get("file_path", "")
                    if file_path:
                        session_data.file_changes.append(file_path)

    # Process user messages for epic context
    elif entry_type == "user":
        message = entry.get("message", {})
        content = message.get("content", "")
        if isinstance(content, str):
            _extract_issue_ids(content, session_data)


def _process_bash_command(command: str, session_data: SessionData, timestamp: str):
    """Extract beads operations and git commits from bash commands."""
    # Beads update operations
    match = BD_UPDATE_PATTERN.search(command)
    if match:
        issue_id = match.group(1).lower()
        session_data.beads_operations.append(BeadsOperation(
            command=command,
            issue_id=issue_id,
            operation="update",
            timestamp=timestamp,
        ))
        session_data.issue_ids.add(issue_id)
        _check_for_epic(issue_id, session_data)

    # Beads close operations
    match = BD_CLOSE_PATTERN.search(command)
    if match:
        issue_id = match.group(1).lower()
        session_data.beads_operations.append(BeadsOperation(
            command=command,
            issue_id=issue_id,
            operation="close",
            timestamp=timestamp,
        ))
        session_data.issue_ids.add(issue_id)
        _check_for_epic(issue_id, session_data)

    # Beads show operations
    match = BD_SHOW_PATTERN.search(command)
    if match:
        issue_id = match.group(1).lower()
        session_data.beads_operations.append(BeadsOperation(
            command=command,
            issue_id=issue_id,
            operation="show",
            timestamp=timestamp,
        ))
        session_data.issue_ids.add(issue_id)
        _check_for_epic(issue_id, session_data)

    # Git commits
    if "git commit" in command:
        session_data.commits_made.append(command)

    # Extract any other issue IDs mentioned
    _extract_issue_ids(command, session_data)


def _extract_issue_ids(text: str, session_data: SessionData):
    """Extract issue IDs from text."""
    for match in ISSUE_ID_PATTERN.finditer(text):
        issue_id = match.group(1).lower()
        session_data.issue_ids.add(issue_id)
        _check_for_epic(issue_id, session_data)


def _check_for_epic(issue_id: str, session_data: SessionData):
    """Check if an issue ID is an epic (has children like .1, .2)."""
    # Check if we've seen children of this issue
    base_id = issue_id.split('.')[0]

    # Look for child patterns in existing issues
    for existing in session_data.issue_ids:
        if existing.startswith(f"{base_id}."):
            session_data.epic_ids.add(base_id)
            return

    # Check if this is a child pattern (parent is likely epic)
    match = EPIC_CHILD_PATTERN.match(issue_id)
    if match:
        parent_id = match.group(1).lower()
        session_data.epic_ids.add(parent_id)


def _is_compaction_boundary(entry: dict) -> bool:
    """Detect compaction summary message."""
    message = entry.get("message", {})
    content = message.get("content", "")

    # String content
    if isinstance(content, str):
        return (
            "session is being continued" in content.lower() or
            "summary below covers" in content.lower() or
            "context has been compacted" in content.lower()
        )

    # Array content
    if isinstance(content, list):
        for item in content:
            if item.get("type") == "text":
                text = item.get("text", "").lower()
                if any(marker in text for marker in [
                    "session is being continued",
                    "summary below covers",
                    "context has been compacted"
                ]):
                    return True

    return False


def extract_epic_from_session(session: SessionData) -> Optional[str]:
    """
    Determine which epic was worked on in this session.

    Priority:
    1. Epic explicitly passed to /implement or /crank
    2. Epic mentioned in bd update --status in_progress
    3. Epic with most bd close operations
    4. Epic referenced most frequently
    """
    if not session.epic_ids and not session.issue_ids:
        return None

    # If we have identified epics, find the one most worked on
    if session.epic_ids:
        # Count operations per epic
        epic_scores: dict[str, int] = {}
        for epic_id in session.epic_ids:
            score = 0
            for op in session.beads_operations:
                if op.issue_id == epic_id or op.issue_id.startswith(f"{epic_id}."):
                    if op.operation == "close":
                        score += 3  # Closes are strong signals
                    elif op.operation == "update":
                        score += 2
                    else:
                        score += 1
            epic_scores[epic_id] = score

        if epic_scores:
            best_epic = max(epic_scores.items(), key=lambda x: x[1])
            return best_epic[0]

    # Fallback: look for issue with most operations
    if session.issue_ids:
        issue_scores: dict[str, int] = {}
        for issue_id in session.issue_ids:
            issue_scores[issue_id] = sum(
                1 for op in session.beads_operations if op.issue_id == issue_id
            )
        if issue_scores:
            best_issue = max(issue_scores.items(), key=lambda x: x[1])
            return best_issue[0]

    return None


def main():
    """CLI entry point."""
    import argparse

    parser = argparse.ArgumentParser(description="Parse Claude Code session transcripts")
    parser.add_argument("--transcript", help="Path to transcript file")
    parser.add_argument("--project", help="Project path to find transcripts for")
    parser.add_argument("--extract-epic", action="store_true", help="Extract and print epic ID")
    parser.add_argument("--list-transcripts", action="store_true", help="List transcripts")
    parser.add_argument("--verbose", "-v", action="store_true", help="Verbose output")

    args = parser.parse_args()

    if args.list_transcripts and args.project:
        transcripts = find_session_transcripts(args.project)
        for t in transcripts:
            print(t)
        return

    if args.transcript:
        transcript_path = Path(args.transcript)
        if not transcript_path.exists():
            print(f"Error: Transcript not found: {transcript_path}")
            return

        session = parse_transcript(transcript_path)

        if args.extract_epic:
            epic = extract_epic_from_session(session)
            if epic:
                print(epic)
            return

        if args.verbose:
            print(f"Session ID: {session.session_id}")
            print(f"Epic IDs: {session.epic_ids}")
            print(f"Issue IDs: {session.issue_ids}")
            print(f"Beads Operations: {len(session.beads_operations)}")
            for op in session.beads_operations[:10]:
                print(f"  {op.operation} {op.issue_id}")
            print(f"File Changes: {len(session.file_changes)}")
            print(f"Commits: {len(session.commits_made)}")
            print(f"Compaction Markers: {session.compaction_markers}")


if __name__ == "__main__":
    main()
