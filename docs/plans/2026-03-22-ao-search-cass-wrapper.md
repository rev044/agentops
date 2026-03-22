# AO Search as an Upstream CASS Wrapper

Date: 2026-03-22
Issue: `ag-6qk`
Status: approved implementation plan

## Problem

`ao search` currently presents itself as CASS-backed session-history search, but the shipped implementation is a repo-local AO search path over forged sessions and nearby `.agents/` artifacts. That creates three product problems:

1. AO is reimplementing session-history search instead of delegating to the upstream tool that already owns it.
2. Known chat history can be indexed and searchable in `cass` while remaining undiscoverable through `ao search`.
3. Help text and docs blur two distinct concerns:
   - session-history search
   - curated repo-local AO memory retrieval

## Product Decision

`ao search` becomes a broker command.

- Upstream `cass` is the source of truth for session-history indexing and search.
- AO keeps a repo-local search path for `.agents/` artifacts because those are AO-native memory surfaces, not raw session history.
- `ao lookup` remains the primary command for curated knowledge retrieval by relevance.

In short:

- `cass` answers: "What happened in chat history for this workspace?"
- `ao lookup` answers: "What durable knowledge artifacts should I load?"
- `ao search` brokers across those surfaces without reimplementing raw session-history search.

## Backend Contract

`ao search` supports three backends:

1. `cass`
   - Upstream session-history search via `cass search`
   - Workspace-scoped with `--workspace <cwd>`
   - AO normalizes output into the existing `searchResult` schema
2. `local`
   - Repo-local AO search over forged sessions and adjacent `.agents/` surfaces
   - This remains in-process and AO-owned
3. `smart-connections`
   - Optional semantic-first search when `--use-sc` is requested
   - Falls back to the selected non-SC backend chain

## Default Behavior

Default mode is `auto`.

`auto` means:

1. If `--use-sc` is set and Smart Connections is available, try it first.
2. If upstream `cass` is available on `PATH`, query it for workspace-scoped session history.
3. If repo-local AO search data exists, query AO local memory surfaces too.
4. Normalize, merge, deduplicate, sort, and return one unified result set.

This keeps session-history discovery tied to the real upstream tool while preserving AO-native memory retrieval in the same command surface.

## Explicit Modes

Add one explicit mode and redefine one existing flag:

- `--cass`
  - Force upstream `cass` only
  - Error if `cass` is unavailable
- `--local`
  - Force repo-local AO search only

`--cass` and `--local` are mutually exclusive.

## Output Contract

AO preserves the existing outward result shape:

```json
[
  {
    "path": "/abs/path",
    "score": 12.34,
    "context": "snippet",
    "type": "session"
  }
]
```

Normalization rules:

- Upstream `cass` hits map from `source_path`, `score`, and `snippet`/`content`.
- AO local hits keep their current path/context/type behavior.
- Multiple `cass` hits from the same `source_path` collapse to the best-scoring hit for that file.
- Type classification remains AO-owned:
  - `session` for upstream `cass` transcript hits
  - path-derived AO types for local artifacts (`learning`, `pattern`, `finding`, `research`, `decision`, `session`, `knowledge`)

## Help and Docs Contract

Help text and docs must say exactly this:

- `ao search` is a wrapper over upstream `cass` for session history when available.
- `ao search` also searches repo-local AO memory surfaces in auto mode.
- Raw global/session history is owned by `cass`, not AO's local grep implementation.
- `ao lookup` is still the better fit when the user specifically wants curated learnings/patterns/findings.

## Non-Goals

This change does not:

- vendor or embed the `cass` implementation into AO
- recreate `cass` ranking/indexing logic inside AO
- add global reindexing logic to AO
- replace `ao lookup`

## TDD Plan

Write failing tests first for:

1. auto mode calls upstream `cass` when available
2. auto mode falls back to repo-local AO search when `cass` is unavailable
3. `--cass` requires upstream `cass`
4. `--local` skips upstream `cass`
5. upstream `cass` hits are normalized and deduplicated by source path
6. help text/CLI docs reflect wrapper behavior instead of claiming built-in CASS-only search

## Acceptance Criteria

1. `ao search` uses upstream `cass search` for workspace-scoped session-history results when `cass` is installed.
2. `ao search` still returns useful repo-local results when `cass` is unavailable.
3. `ao search --help`, `README.md`, and `docs/reference.md` match the shipped behavior.
4. AO no longer claims to own raw session-history search semantics that belong to upstream `cass`.
5. Known workspace chat history discoverable through `cass` becomes discoverable through the supported `ao search` path.
