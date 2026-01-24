#!/usr/bin/env python3
"""
Root cause tracing for post-mortem session tracing.

Traces the full chain from research to implementation:
- Research artifacts
- Product briefs
- Pre-mortem results
- Spec/formula files
- Plan artifacts
- Implementation commits
- Key decisions made during session
"""
import re
from dataclasses import dataclass, field
from pathlib import Path
from typing import Optional

# Handle both package and standalone imports
try:
    from .transcript import SessionData
except ImportError:
    from transcript import SessionData


@dataclass
class Decision:
    """A key decision made during the session."""
    content: str
    source: str  # "beads_comment", "commit_message", "ask_user_question", "explicit"
    timestamp: str = ""
    context: str = ""


@dataclass
class KnowledgeInput:
    """Prior knowledge that informed this work."""
    path: Path
    source_type: str  # "learning", "retro", "pattern", "research"
    relevance: str  # Brief description of why it's relevant


@dataclass
class ProvenanceChain:
    """The full provenance chain for an epic."""
    epic_id: str
    # Forward chain: research -> implementation
    research_artifact: Optional[Path] = None
    product_brief: Optional[Path] = None
    pre_mortem_results: Optional[Path] = None
    spec_artifact: Optional[Path] = None
    plan_artifact: Optional[Path] = None
    implementation_commits: list[str] = field(default_factory=list)
    decisions_made: list[Decision] = field(default_factory=list)
    # Backward chain: prior knowledge -> this work
    knowledge_inputs: list[KnowledgeInput] = field(default_factory=list)


def trace_epic_provenance(
    epic_id: str,
    session: SessionData,
    project_path: Optional[str] = None,
) -> ProvenanceChain:
    """
    Trace how we got from research to implementation.

    Args:
        epic_id: The epic ID to trace
        session: Parsed session data
        project_path: Root of the project (for artifact discovery)

    Returns:
        ProvenanceChain with:
        - research_artifact: Path to .agents/research/*.md
        - product_brief: Path to .agents/products/*.md (if any)
        - pre_mortem_results: Path to pre-mortem findings
        - spec_artifact: Path to spec/formula used
        - plan_artifact: Path to plan file
        - implementation_commits: List of commits
        - decisions_made: Key decisions from session
    """
    chain = ProvenanceChain(epic_id=epic_id)

    # Determine project path
    if not project_path:
        project_path = _infer_project_path(session)

    if project_path:
        project_root = Path(project_path)

        # Find research artifact
        chain.research_artifact = find_research_for_epic(epic_id, project_root)

        # Find product brief
        chain.product_brief = _find_product_brief(epic_id, project_root)

        # Find pre-mortem results
        chain.pre_mortem_results = _find_pre_mortem(epic_id, project_root)

        # Find spec artifact
        chain.spec_artifact = _find_spec_artifact(epic_id, project_root)

        # Find plan artifact
        chain.plan_artifact = _find_plan_artifact(epic_id, project_root)

        # Backward trace: Find prior knowledge that informed this work
        chain.knowledge_inputs = trace_knowledge_inputs(epic_id, project_root)

    # Extract commits from session
    chain.implementation_commits = session.commits_made.copy()

    # Extract decisions
    chain.decisions_made = extract_decisions_from_session(session)

    return chain


def find_research_for_epic(epic_id: str, project_root: Path) -> Optional[Path]:
    """
    Find research artifact that led to this epic.

    Searches:
    1. .agents/research/*.md files mentioning the epic
    2. Files with similar naming (date-based)
    """
    # Search locations for research artifacts
    search_paths = [
        project_root / ".agents" / "research",
        project_root.parent / ".agents" / "research",  # Parent rig level
        Path.home() / "gt" / ".agents" / "research",  # Town level
    ]

    for search_path in search_paths:
        if not search_path.exists():
            continue

        for research_file in search_path.glob("*.md"):
            try:
                content = research_file.read_text()
                # Check if epic is mentioned
                if epic_id.lower() in content.lower():
                    return research_file

                # Check if title/filename matches epic description
                # (This is a heuristic - might need refinement)
                epic_parts = epic_id.split("-")
                if len(epic_parts) >= 2:
                    # Skip prefix, check rest
                    epic_desc = "-".join(epic_parts[1:])
                    if epic_desc.lower() in research_file.stem.lower():
                        return research_file
            except Exception:
                continue

    return None


def _find_product_brief(epic_id: str, project_root: Path) -> Optional[Path]:
    """Find product brief for this epic."""
    search_paths = [
        project_root / ".agents" / "products",
        project_root.parent / ".agents" / "products",
        Path.home() / "gt" / ".agents" / "products",
    ]

    for search_path in search_paths:
        if not search_path.exists():
            continue

        for brief_file in search_path.glob("*.md"):
            try:
                content = brief_file.read_text()
                if epic_id.lower() in content.lower():
                    return brief_file
            except Exception:
                continue

    return None


def _find_pre_mortem(epic_id: str, _project_root: Path) -> Optional[Path]:
    """Find pre-mortem results for this epic."""
    # Pre-mortem results are typically in .claude/plans/ with -agent suffix
    # _project_root kept for interface consistency
    plans_dir = Path.home() / ".claude" / "plans"

    if not plans_dir.exists():
        return None

    # Look for files with epic ID and -agent suffix
    for plan_file in plans_dir.glob("*-agent*.md"):
        try:
            content = plan_file.read_text()
            if epic_id.lower() in content.lower():
                return plan_file
        except Exception:
            continue

    return None


def _find_spec_artifact(epic_id: str, project_root: Path) -> Optional[Path]:
    """Find spec or formula artifact."""
    # Check for .formula.toml
    formula_path = project_root / ".formula.toml"
    if formula_path.exists():
        return formula_path

    # Check .agents/specs/
    specs_dir = project_root / ".agents" / "specs"
    if specs_dir.exists():
        for spec_file in specs_dir.glob("*.md"):
            try:
                content = spec_file.read_text()
                if epic_id.lower() in content.lower():
                    return spec_file
            except Exception:
                continue

    return None


def _find_plan_artifact(epic_id: str, _project_root: Path) -> Optional[Path]:
    """Find plan artifact (not pre-mortem)."""
    # _project_root kept for interface consistency
    plans_dir = Path.home() / ".claude" / "plans"

    if not plans_dir.exists():
        return None

    # Look for regular plan files (not -agent suffix)
    for plan_file in plans_dir.glob("*.md"):
        if "-agent" in plan_file.name:
            continue  # Skip pre-mortem files

        try:
            content = plan_file.read_text()
            if epic_id.lower() in content.lower():
                return plan_file
        except Exception:
            continue

    return None


def extract_decisions_from_session(session: SessionData) -> list[Decision]:
    """
    Extract key decisions made during implementation.

    Sources:
    - AskUserQuestion tool calls
    - Comments in beads
    - Commit messages
    - Explicit decision markers in conversation
    """
    decisions = []

    # Extract decisions from beads operations (comments)
    for op in session.beads_operations:
        if "comment" in op.command.lower():
            # Try to extract comment content
            match = re.search(r'-m\s+["\']([^"\']+)["\']', op.command)
            if match:
                decisions.append(Decision(
                    content=match.group(1),
                    source="beads_comment",
                    timestamp=op.timestamp,
                    context=f"Issue: {op.issue_id}",
                ))

    # Extract decisions from commit messages
    for commit_cmd in session.commits_made:
        match = re.search(r'-m\s+["\']([^"\']+)["\']', commit_cmd)
        if match:
            msg = match.group(1)
            # Look for decision indicators
            if any(word in msg.lower() for word in ["decide", "chose", "selected", "opt"]):
                decisions.append(Decision(
                    content=msg,
                    source="commit_message",
                    context="git commit",
                ))

    return decisions


def trace_knowledge_inputs(epic_id: str, project_root: Path) -> list[KnowledgeInput]:
    """
    Find prior knowledge that informed this work (backward tracing).

    This closes the knowledge loop by identifying what learnings, retros,
    and patterns from previous work were relevant to this epic.

    Searches:
    - .agents/learnings/ - Extracted lessons from previous post-mortems
    - .agents/retros/ - Retrospective artifacts
    - .agents/patterns/ - Reusable solution patterns
    - .agents/research/ - Prior research (older than this epic's research)
    """
    inputs: list[KnowledgeInput] = []

    # Extract keywords from epic ID for relevance matching
    # e.g., "ol-r1-closure" -> ["r1", "closure"]
    epic_keywords = _extract_keywords_from_epic(epic_id)

    # Search paths (rig level, then town level)
    search_bases = [
        project_root / ".agents",
        project_root.parent / ".agents",  # Rig level
        Path.home() / "gt" / ".agents",  # Town level
    ]

    for base in search_bases:
        if not base.exists():
            continue

        # Check learnings
        learnings_dir = base / "learnings"
        if learnings_dir.exists():
            for f in learnings_dir.glob("*.md"):
                relevance = _check_relevance(f, epic_keywords)
                if relevance:
                    inputs.append(KnowledgeInput(
                        path=f,
                        source_type="learning",
                        relevance=relevance,
                    ))

        # Check retros
        retros_dir = base / "retros"
        if retros_dir.exists():
            for f in retros_dir.glob("*.md"):
                relevance = _check_relevance(f, epic_keywords)
                if relevance:
                    inputs.append(KnowledgeInput(
                        path=f,
                        source_type="retro",
                        relevance=relevance,
                    ))

        # Check patterns
        patterns_dir = base / "patterns"
        if patterns_dir.exists():
            for f in patterns_dir.glob("*.md"):
                relevance = _check_relevance(f, epic_keywords)
                if relevance:
                    inputs.append(KnowledgeInput(
                        path=f,
                        source_type="pattern",
                        relevance=relevance,
                    ))

    # Deduplicate by path
    seen_paths: set[str] = set()
    unique_inputs: list[KnowledgeInput] = []
    for inp in inputs:
        path_str = str(inp.path)
        if path_str not in seen_paths:
            seen_paths.add(path_str)
            unique_inputs.append(inp)

    return unique_inputs


def _extract_keywords_from_epic(epic_id: str) -> list[str]:
    """Extract searchable keywords from epic ID."""
    # Split on hyphens, remove prefix (first part if 2-4 chars)
    parts = epic_id.lower().split("-")
    if parts and len(parts[0]) <= 4:
        parts = parts[1:]  # Remove prefix like "ol-", "at-"

    # Filter out very short or numeric parts
    keywords = [p for p in parts if len(p) > 2 and not p.isdigit()]

    return keywords


def _check_relevance(file_path: Path, keywords: list[str]) -> Optional[str]:
    """
    Check if a file is relevant to the epic based on keywords.

    Returns relevance description if relevant, None otherwise.
    """
    if not keywords:
        return None

    try:
        content = file_path.read_text().lower()
        filename = file_path.stem.lower()

        # Check filename first (faster)
        filename_matches = [k for k in keywords if k in filename]
        if filename_matches:
            return f"Filename matches: {', '.join(filename_matches)}"

        # Check content
        content_matches = [k for k in keywords if k in content]
        if content_matches:
            return f"Content mentions: {', '.join(content_matches)}"

    except Exception:
        pass

    return None


def _infer_project_path(session: SessionData) -> Optional[str]:
    """Infer project path from session data."""
    # Look for cwd in session operations
    if session.file_changes:
        # Get common path prefix
        paths = [Path(f) for f in session.file_changes]
        if paths:
            # Find common ancestor
            common = paths[0].parent
            for p in paths[1:]:
                while common not in p.parents and common != p.parent:
                    common = common.parent
                    if common == Path("/"):
                        return None
            return str(common)

    return None


def format_provenance_report(chain: ProvenanceChain) -> str:
    """Format provenance chain as markdown report section."""
    lines = ["## Provenance Chain\n"]

    if chain.research_artifact:
        lines.append("### Research Phase")
        lines.append(f"- **Artifact:** `{chain.research_artifact}`")
        lines.append("")

    if chain.product_brief:
        lines.append("### Product Phase")
        lines.append(f"- **Brief:** `{chain.product_brief}`")
        lines.append("")

    if chain.pre_mortem_results:
        lines.append("### Pre-Mortem Phase")
        lines.append(f"- **Results:** `{chain.pre_mortem_results}`")
        lines.append("")

    if chain.spec_artifact:
        lines.append("### Spec Phase")
        lines.append(f"- **Spec/Formula:** `{chain.spec_artifact}`")
        lines.append("")

    if chain.plan_artifact:
        lines.append("### Plan Phase")
        lines.append(f"- **Plan:** `{chain.plan_artifact}`")
        lines.append("")

    if chain.implementation_commits:
        lines.append("### Implementation Phase")
        lines.append(f"- **Commits:** {len(chain.implementation_commits)}")
        for commit in chain.implementation_commits[:5]:
            # Truncate long commit commands
            if len(commit) > 80:
                commit = commit[:77] + "..."
            lines.append(f"  - `{commit}`")
        lines.append("")

    if chain.decisions_made:
        lines.append("### Key Decisions")
        for decision in chain.decisions_made[:5]:
            lines.append(f"- **[{decision.source}]** {decision.content}")
            if decision.context:
                lines.append(f"  - Context: {decision.context}")
        lines.append("")

    # Knowledge loop: prior knowledge that informed this work
    if chain.knowledge_inputs:
        lines.append("### Knowledge Loop Inputs (Backward Trace)")
        lines.append("*Prior knowledge that may have informed this work:*\n")

        # Group by source type
        by_type: dict[str, list[KnowledgeInput]] = {}
        for inp in chain.knowledge_inputs:
            if inp.source_type not in by_type:
                by_type[inp.source_type] = []
            by_type[inp.source_type].append(inp)

        for source_type, inputs in sorted(by_type.items()):
            lines.append(f"**{source_type.title()}s:**")
            for inp in inputs[:3]:  # Limit to 3 per type
                lines.append(f"- `{inp.path.name}` - {inp.relevance}")
            if len(inputs) > 3:
                lines.append(f"- ... and {len(inputs) - 3} more")
            lines.append("")

    return "\n".join(lines)


def main():
    """CLI entry point."""
    import argparse

    parser = argparse.ArgumentParser(description="Trace epic provenance")
    parser.add_argument("--epic", required=True, help="Epic ID to trace")
    parser.add_argument("--project", help="Project root path")
    parser.add_argument("--transcript", help="Session transcript to analyze")
    parser.add_argument("--report", action="store_true", help="Output formatted report")

    args = parser.parse_args()

    # Create minimal session if no transcript
    session = SessionData(session_id="manual", project_path=args.project or "")

    if args.transcript:
        try:
            from .transcript import parse_transcript
        except ImportError:
            from transcript import parse_transcript
        session = parse_transcript(Path(args.transcript))

    chain = trace_epic_provenance(
        epic_id=args.epic,
        session=session,
        project_path=args.project,
    )

    if args.report:
        print(format_provenance_report(chain))
    else:
        print(f"Epic: {chain.epic_id}")
        print(f"Research: {chain.research_artifact}")
        print(f"Product: {chain.product_brief}")
        print(f"Pre-mortem: {chain.pre_mortem_results}")
        print(f"Spec: {chain.spec_artifact}")
        print(f"Plan: {chain.plan_artifact}")
        print(f"Commits: {len(chain.implementation_commits)}")
        print(f"Decisions: {len(chain.decisions_made)}")


if __name__ == "__main__":
    main()
