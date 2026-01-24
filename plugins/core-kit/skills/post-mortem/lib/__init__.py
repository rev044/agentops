"""
Post-mortem session tracing library.

Provides tools for:
- Parsing Claude Code session transcripts
- Traversing compaction chains
- Tracing root causes from research to implementation
"""

# Make imports work both as package and standalone
try:
    from .transcript import (
        SessionData,
        BeadsOperation,
        find_session_transcripts,
        parse_transcript,
        extract_epic_from_session,
    )
    from .compaction import (
        CompactionChain,
        find_compaction_chain,
        build_composite_session,
        is_compaction_boundary,
    )
    from .trace import (
        Decision,
        KnowledgeInput,
        ProvenanceChain,
        trace_epic_provenance,
        find_research_for_epic,
        extract_decisions_from_session,
        trace_knowledge_inputs,
        format_provenance_report,
    )
except ImportError:
    # Standalone execution
    from transcript import (  # type: ignore[import-not-found]
        SessionData,
        BeadsOperation,
        find_session_transcripts,
        parse_transcript,
        extract_epic_from_session,
    )
    from compaction import (  # type: ignore[import-not-found]
        CompactionChain,
        find_compaction_chain,
        build_composite_session,
        is_compaction_boundary,
    )
    from trace import (  # type: ignore[import-not-found]
        Decision,
        KnowledgeInput,
        ProvenanceChain,
        trace_epic_provenance,
        find_research_for_epic,
        extract_decisions_from_session,
        trace_knowledge_inputs,
        format_provenance_report,
    )


__all__ = [
    # transcript
    "SessionData",
    "BeadsOperation",
    "find_session_transcripts",
    "parse_transcript",
    "extract_epic_from_session",
    # compaction
    "CompactionChain",
    "find_compaction_chain",
    "build_composite_session",
    "is_compaction_boundary",
    # trace
    "Decision",
    "KnowledgeInput",
    "ProvenanceChain",
    "trace_epic_provenance",
    "find_research_for_epic",
    "extract_decisions_from_session",
    "trace_knowledge_inputs",
    "format_provenance_report",
]
