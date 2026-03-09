# Finding Registry Contract

The canonical reusable-finding registry lives at `.agents/findings/registry.jsonl`.

This file is the v1 prevention surface for the AgentOps flywheel. Planning and judgment should load it before rediscovering a known failure, and judgment should write back only the reusable findings that deserve to affect future work.

## Canonical Shape

- One JSON object per line.
- Each line must validate against [finding-registry.schema.json](finding-registry.schema.json).
- Canonical contract version is `version: 1`.
- Canonical path is repo-local. Do not require a service or database for the v1 slice.

## Required Fields

Each entry must include:

- `id`
- `version`
- `tier`
- `source.repo`
- `source.session`
- `source.file`
- `source.skill`
- `date`
- `severity`
- `category`
- `pattern`
- `detection_question`
- `checklist_item`
- `applicable_languages`
- `applicable_when`
- `status`
- `superseded_by`
- `dedup_key`
- `hit_count`
- `last_cited`
- `ttl_days`
- `confidence`

## Controlled Vocabulary

`applicable_when` is the only controlled vocabulary in v1. Allowed values:

- `plan-shape`
- `classifier`
- `enum-parser`
- `struct-change`
- `pattern-matcher`
- `validation-gap`
- `test-gap`
- `docs-drift`

Writers must choose at least one value. Readers should treat the first item as the primary work-shape hint.

## Canonical Normalization

`dedup_key = <category>|<pattern-slug>|<primary-applicable-when>`

Normalization rules:

- `pattern-slug` is lowercase
- strip punctuation
- collapse repeated whitespace
- join tokens with `-`
- `primary-applicable-when` is the first controlled-vocabulary item in `applicable_when`

Example:

- `category = validation-gap`
- `pattern = Prior finding injection was skipped in planning`
- `applicable_when = ["plan-shape", "validation-gap"]`
- `dedup_key = validation-gap|prior-finding-injection-was-skipped-in-planning|plan-shape`

## Reader Ranking Rules

Readers must:

- consider only `status=active`
- cap the injected set at 5 entries
- sort by severity first
- break ties by explicit scope match
- break remaining ties by literal text match

Recommended scope match order:

1. `applicable_when` overlap with the current work shape
2. language overlap
3. path or changed-file overlap when available
4. literal keyword overlap with the goal, plan, or review target

Reader failure behavior is fail-open:

- missing registry -> skip silently
- empty registry -> skip silently
- malformed line -> warn and ignore that line
- unreadable file -> warn once and continue without findings

## Write-Path Semantics

Writers must:

- persist only reusable findings, not every transient comment
- merge by `dedup_key`
- preserve the most actionable current wording for `pattern`, `detection_question`, and `checklist_item`
- keep lifecycle fields (`status`, `superseded_by`, `ttl_days`, `confidence`) explicit in the merged record

Lifecycle notes for v1:

- future citations should increment `hit_count`
- future citations should update `last_cited`
- automated CLI mutation of those fields is deferred in v1
- post-mortem scoring inputs such as confidence, citations, and recency may guide promotion decisions, but automated scoring is deferred in v1

## Retirement and Supersession

- `active` entries are eligible for injection
- `retired` entries stay in the registry for history but must not be injected
- `superseded` entries must set `superseded_by` to the replacement finding id
- `ttl_days` is advisory metadata in v1 and is not yet auto-enforced

## Atomic Update Rule

Writers must update `.agents/findings/registry.jsonl` with a temp file plus atomic rename:

1. Ensure `.agents/findings/` exists.
2. Read the current `registry.jsonl` if present.
3. Merge or append entries in memory by `dedup_key`.
4. Write the full result to a temp file under `.agents/findings/`.
5. Replace the live file with an atomic rename.

If a lock is used, the canonical lock path is `.agents/findings/registry.lock`, and writers must remove it on exit, including failure paths.

## Example Entry

```json
{
  "id": "f-2026-03-09-001",
  "version": 1,
  "tier": "local",
  "source": {
    "repo": "agentops/crew/nami",
    "session": "2026-03-09",
    "file": ".agents/council/2026-03-09-pre-mortem-finding-compiler-v1.md",
    "skill": "pre-mortem"
  },
  "date": "2026-03-09",
  "severity": "significant",
  "category": "validation-gap",
  "pattern": "Plans can omit prior-finding injection and rediscover the same failure mode.",
  "detection_question": "Did this plan load matching active findings before decomposition or review?",
  "checklist_item": "Verify the relevant skill reads registry.jsonl and cites applied finding IDs or known risks.",
  "applicable_languages": ["markdown", "shell"],
  "applicable_when": ["plan-shape", "validation-gap"],
  "status": "active",
  "superseded_by": null,
  "dedup_key": "validation-gap|prior-finding-injection-and-rediscovery|plan-shape",
  "hit_count": 0,
  "last_cited": null,
  "ttl_days": 30,
  "confidence": "high"
}
```

## Explicit Deferrals

The following are intentionally out of scope for the v1 contract:

- cross-repo seed, pull, and export transport
- a new top-level `ao findings` command surface
- automatic citation-count updates in CLI code
- automatic TTL retirement
- hook-enforced compiled constraints derived from findings

The current contract is registry-first and advisory. It exists so the same failure is discovered once, normalized once, and then checked earlier in future planning and review.
