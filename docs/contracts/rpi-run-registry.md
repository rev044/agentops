# RPI Run Registry

The RPI phased orchestrator (`ao rpi phased`) writes structured artifacts to a well-known directory layout. This document defines the registry layout, file naming conventions, required fields per phase transition, and contract guarantees.

## Directory Layout

All RPI artifacts live under `.agents/rpi/` relative to the working directory (which may be a worktree):

```
.agents/rpi/
  phased-state.json           # Orchestrator state (survives across phases)
  phased-orchestration.log    # Append-only log of phase transitions
  phase-1-result.json         # Phase result artifact (discovery)
  phase-2-result.json         # Phase result artifact (implementation)
  phase-3-result.json         # Phase result artifact (validation)
  phase-1-summary.md          # Phase summary (written by Claude or fallback)
  phase-2-summary.md          # Phase summary
  phase-3-summary.md          # Phase summary
  phase-1-handoff.md          # Handoff file (context degradation signal)
  phase-2-handoff.md          # Handoff file
  phase-3-handoff.md          # Handoff file
  live-status.md              # Live status file (optional, --live-status flag)
```

## File Naming Conventions

| File | Purpose | Lifecycle |
|------|---------|-----------|
| `phased-state.json` | Orchestrator state: goal, epic ID, phase, verdicts, attempts | Written after each phase; read on resume (`--from`) |
| `phased-orchestration.log` | Append-only transition log for debugging | Appended at every transition point |
| `phase-{N}-result.json` | Structured phase outcome matching `rpi-phase-result.schema.json` | Written atomically after each phase completes or fails |
| `phase-{N}-summary.md` | Human-readable summary for cross-phase context | Written by Claude (preferred) or orchestrator fallback |
| `phase-{N}-handoff.md` | Context degradation signal from Claude | Written by Claude when it detects context degradation |
| `live-status.md` | Real-time progress for external watchers | Continuously updated when `--live-status` is enabled |

Where `{N}` is the phase number: 1 (discovery), 2 (implementation), 3 (validation).

## Required Fields Per Phase Transition

Each `phase-{N}-result.json` must contain at minimum:

| Field | Type | Description |
|-------|------|-------------|
| `schema_version` | integer | Always `1` (current version) |
| `run_id` | string | Hex run identifier |
| `phase` | integer | Phase number (1-3) |
| `phase_name` | string | `discovery`, `implementation`, or `validation` |
| `status` | string | `started`, `completed`, `failed`, or `retrying` |
| `started_at` | string | ISO 8601 timestamp |

Optional fields populated when available:

| Field | Type | Description |
|-------|------|-------------|
| `retries` | integer | Number of retry attempts (default 0) |
| `error` | string | Error message on failure |
| `backend` | string | Execution backend: `direct`, `ntm`, or `stream` |
| `artifacts` | object | Map of artifact names to paths |
| `verdicts` | object | Map of gate names to verdict strings |
| `completed_at` | string | ISO 8601 timestamp |
| `duration_seconds` | number | Wall-clock duration |

## Phase Transition Validation

Before starting phase N (for N > 1), the orchestrator validates that `phase-{N-1}-result.json` exists and has `status: "completed"`. This ensures phases execute in order and that prior phases completed successfully.

Validation failures produce a clear error message indicating which prior phase result is missing or incomplete.

## Contract Guarantees

### Atomic Writes

Phase result files are written atomically using a write-to-temp-then-rename pattern:

1. Marshal JSON to `phase-{N}-result.json.tmp`
2. Rename `.tmp` to final path

This ensures readers never see a partial write. The orchestration log (`phased-orchestration.log`) uses append-only writes which are atomic at the OS level for reasonable line lengths.

### Schema Version

Every result file includes `schema_version: 1`. Consumers must check this field and handle unknown versions gracefully (fail open or warn). When the schema evolves, the version will increment and the orchestrator will maintain backward compatibility for at least one major version.

### Idempotent Resume

The orchestrator can resume from any phase using `--from=<phase>`. On resume, it reads `phased-state.json` to recover epic ID, verdicts, and attempt counts. Phase result files from prior phases are preserved and not overwritten on resume.

### Clean Start

When starting from phase 1 (fresh run), the orchestrator removes stale phase summaries and handoff files from prior runs. Phase result files from the current run are written fresh.
