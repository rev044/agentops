# OL-AO Bridge Contracts

> Interchange formats, exit codes, and version negotiation for the Olympus (ol) ↔ AgentOps (ao) CLI bridge.

## 1. Learning Interchange Format

### OL → AO (harvest → forge)

OL `harvestCandidate` maps to AO `Candidate` as follows:

| OL Field (`harvestCandidate`) | AO Field (`Candidate`) | Transform |
|-------------------------------|------------------------|-----------|
| `id` | `id` | Direct copy |
| `type` ("LEARNING", "ANTI_PATTERN") | `type` | Map: `LEARNING` → `learning`, `ANTI_PATTERN` → `failure` |
| `title` | `content` | Direct copy |
| `summary` | `context` | Direct copy |
| `category` | `metadata["category"]` | Store in metadata |
| `quest_id` | `metadata["quest_id"]` | Store in metadata |
| `source_artifacts` | `source.transcript_path` | Join with `;` |
| `tags` | `metadata["tags"]` | Store in metadata |
| `created_at` | `extracted_at` | Parse RFC3339 |
| _(missing)_ | `utility` | Default: `0.5` |
| _(missing)_ | `is_current` | Default: `true` |
| _(missing)_ | `maturity` | Default: `provisional` |
| _(missing)_ | `tier` | Default: `bronze` |
| _(missing)_ | `source.session_id` | Set to `ol-harvest-<quest_id>` |

### File Format

OL harvest outputs markdown with YAML frontmatter to `.agents/learnings/`. AO `inject` discovers files at this same path. The interchange format IS the file system — both read/write `.agents/learnings/*.md`.

**Required frontmatter fields for AO compatibility:**

```yaml
---
id: "learn-2026-02-09-abc12345"
type: learning           # learning | failure | decision | solution | reference
source: "ol-harvest"     # provenance marker
quest_id: "ol-572"       # OL quest traceability
created_at: "2026-02-09T12:00:00Z"
tags: ["go", "testing"]
---
```

**AO reads:** `id`, `type`, `created_at` from frontmatter. Title from first `# ` heading. Summary from body text.

### AO → OL (anti-patterns → constraints)

AO learnings with `type: failure` or `maturity: anti-pattern` can be converted to OL constraints:

| AO Field (`Candidate`) | OL Field (`Constraint`) | Transform |
|-------------------------|-------------------------|-----------|
| `content` | `pattern` | First sentence or title |
| `context` | `detection` | Full description |
| _(needs manual)_ | `test_template` | Derive from content or leave empty |
| `id` | `source` | `ao-learning-<id>` |
| `confidence` | `confidence` | Direct copy (default 0.5 if missing) |
| _(missing)_ | `status` | Default: `proposed` |

### File Location

AO writes anti-patterns/failures to `.agents/learnings/`. OL reads constraints from `.ol/constraints/quarantine.json`. The bridge must translate between formats.

**Bridge command:** `ao export-constraints --format=ol` writes to `.ol/constraints/quarantine.json`.

## 2. Exit Code Contract

### OL Exit Codes

| Code | Meaning | When |
|------|---------|------|
| 0 | Success | Normal completion |
| 1 | Error | Invalid input, test failure, merge conflict |
| 2 | Escalation | Human/agent intervention needed (iteration limit, hard blocker) |
| 42 | Complete | Zeus step: all phases done |

### AO Exit Codes

AO skills run inside Claude Code sessions and don't use exit codes directly. Instead, they signal via:
- File artifacts (`.agents/council/*` with verdicts)
- Promise tags (`<promise>DONE</promise>`, `<promise>BLOCKED</promise>`)

### Bridge Mapping

When AO invokes `ol` commands:

| ol exit code | AO interpretation |
|--------------|-------------------|
| 0 | Proceed to next phase |
| 1 | Log error, retry or escalate |
| 2 | Present to user or invoke fallback skill |
| 42 | Epic/quest complete |

When OL dispatches to Claude Code (which runs AO skills):

| AO outcome | Claude Code exit | ol interpretation |
|------------|-----------------|-------------------|
| Skill succeeds, artifacts written | 0 | Phase complete, advance |
| Skill fails, no artifacts | 1 | Retry phase |
| User cancels / context exhausted | 1 | Escalation (treat as exit 2) |

## 3. Version Negotiation

### Capability Query

Before bridge commands call across CLIs, verify the other CLI exists and supports the expected interface:

```bash
# AO checking OL
ol_version=$(ol --version 2>/dev/null) || { echo "ol CLI not found"; exit 1; }

# OL checking AO
ao_version=$(ao version 2>/dev/null) || { echo "ao CLI not found"; exit 1; }
```

### Feature Detection

Rather than version parsing, use feature detection:

```bash
# Check if ol harvest supports --format flag
ol harvest --help 2>&1 | grep -q "\-\-format" && OL_HAS_FORMAT=true

# Check if ao inject supports --ol-constraints flag
ao inject --help 2>&1 | grep -q "ol-constraints" && AO_HAS_OL=true

# Check if ol validate stage1 exists
ol validate stage1 --help 2>/dev/null && OL_HAS_STAGE1=true
```

### Graceful Degradation

| Missing capability | Fallback |
|-------------------|----------|
| `ol` not on PATH | Skip OL integration, pure AO mode |
| `ao` not on PATH | Skip AO integration, pure OL mode |
| `ol harvest --format=ao` not supported | Manual file copy from `.agents/learnings/` |
| `ao inject --ol-constraints` not supported | Skip constraint injection |
| `.ol/` directory missing | Not an Olympus project, skip all OL features |

## 4. Validation Result Format

### OL Stage1Result → AO Vibe Input

When `/vibe` invokes `ol validate stage1`, it reads the JSON output:

```json
{
  "quest_id": "ol-572",
  "bead_id": "ol-572.3",
  "worktree": "/path/to/worktree",
  "passed": true,
  "steps": [
    {"name": "go build", "passed": true, "duration": "1.2s"},
    {"name": "go vet", "passed": true, "duration": "0.8s"},
    {"name": "go test", "passed": true, "duration": "3.4s"}
  ],
  "summary": "all steps passed"
}
```

**AO vibe integration:** Include OL stage1 results in the vibe report as a "deterministic validation" section. If `passed: false`, auto-FAIL the vibe without running council.

## 5. Storage Path Conventions

| Artifact | OL Location | AO Location | Shared? |
|----------|-------------|-------------|---------|
| Learnings | `.agents/learnings/` | `.agents/learnings/` | **Yes** — both read/write |
| Patterns | `.agents/patterns/` | `.agents/patterns/` | **Yes** — both read/write |
| Sessions | `.ol/quests/*/` | `.agents/ao/sessions/` | No |
| Constraints | `.ol/constraints/` | N/A | OL only |
| Council reports | N/A | `.agents/council/` | AO only |
| Validation runs | `.ol/runs/` | N/A | OL only |
| Ratchet chain | N/A | `.agents/ao/chain.jsonl` | AO only |

**Key insight:** `.agents/learnings/` and `.agents/patterns/` are already shared. No unification needed for the primary knowledge flow. The bridge work is about format compatibility within these shared directories.
