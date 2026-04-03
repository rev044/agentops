# next-work.jsonl Schema

schema_version: 1.3

Contract for `.agents/rpi/next-work.jsonl` — the carry-forward queue that feeds harvested findings from `/post-mortem` into `/evolve`, `/rpi loop`, and related pre-flight checks.

---

## File Format

Newline-delimited JSON (JSONL). Each line is one **Entry** object. Consumers MUST handle any number of lines, including zero. Lines MUST be valid JSON; parsers MUST skip malformed lines with a warning.

The queue is **append-on-write, rewrite-on-lifecycle**:
- producers append new entries
- consumers may rewrite existing lines to claim, release, fail, or consume individual items
- readers MUST tolerate unknown fields for forward compatibility

---

## Entry Object

One entry per producer event, usually one `/post-mortem` run.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `source_epic` | string | yes | ID or slug of the source epic or session |
| `timestamp` | string (ISO-8601) | yes | When the entry was written |
| `items` | array of Item | yes | Harvested follow-up work; may be empty when a post-mortem finds nothing actionable |
| `consumed` | boolean | yes | Aggregate entry status; `true` only when every child item is consumed |
| `claim_status` | enum | yes | Aggregate entry status: `available`, `in_progress`, or `consumed` |
| `claimed_by` | string or null | yes | Aggregate claimant identifier, usually copied from the currently claimed item |
| `claimed_at` | string (ISO-8601) or null | yes | Aggregate claim timestamp |
| `consumed_by` | string or null | yes | Consumer that finalized the batch aggregate |
| `consumed_at` | string (ISO-8601) or null | yes | When the batch aggregate became fully consumed |
| `failed_at` | string (ISO-8601) or null | no | Latest failure timestamp across child items, used for retry ordering |

Entry lifecycle fields are aggregates for dashboards and legacy readers. Item lifecycle inside `items[]` is authoritative.

---

## Item Object

One actionable follow-up item.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `title` | string | yes | Short imperative title |
| `type` | enum | yes | Work category |
| `severity` | enum | yes | Priority signal |
| `source` | enum | yes | Origin of the item |
| `description` | string | yes | Full problem statement and recommended action |
| `evidence` | string | no | Concrete supporting evidence |
| `target_repo` | string | no | Repo slug this applies to, or `*` for cross-repo/process work |
| `consumed` | boolean | no | Item lifecycle flag; omitted or `false` means not yet consumed |
| `claim_status` | enum | no | Item lifecycle state; omitted means `available` |
| `claimed_by` | string or null | no | Item claimant identifier |
| `claimed_at` | string (ISO-8601) or null | no | Item claim timestamp |
| `consumed_by` | string or null | no | Consumer that finalized this item |
| `consumed_at` | string (ISO-8601) or null | no | When this item was consumed |
| `failed_at` | string (ISO-8601) or null | no | Last failure timestamp for retry ordering |

Compatibility notes:
- omitted item `claim_status` means `available`
- producers may attach extra metadata fields (for example `id`, `file`, or `func`); readers MUST ignore unknown fields

---

## Enums

### Type

- `tech-debt`
- `improvement`
- `pattern-fix`
- `process-improvement`
- `feature`
- `bug`
- `task`

### Severity

- `high`
- `medium`
- `low`

### Source

- `council-finding`
- `retro-learning`
- `retro-pattern`
- `evolve-generator`
- `feature-suggestion`
- `backlog-processing`

### Claim Status

- `available`
- `in_progress`
- `consumed`

## Lifecycle Rules

1. Writers create new entries in available state:
   - entry `consumed=false`
   - entry `claim_status="available"`
   - entry `claimed_by=null`
   - entry `claimed_at=null`
   - entry `consumed_by=null`
   - entry `consumed_at=null`
2. Consumers claim one item before execution:
   - item `claim_status="in_progress"`
   - item `claimed_by=<consumer>`
   - item `claimed_at=<timestamp>`
3. Consumers finalize only after the full cycle succeeds:
   - item `consumed=true`
   - item `claim_status="consumed"`
   - item `consumed_by=<consumer>`
   - item `consumed_at=<timestamp>`
4. Failed or regressed cycles release the claim:
   - item `consumed=false`
   - item `claim_status="available"`
   - item `claimed_by=null`
   - item `claimed_at=null`
   - optionally stamp item `failed_at`
5. The entry aggregate flips to `consumed=true` only when every child item is consumed.
6. Never mark an item consumed at pick-time.

## Legacy Compatibility

Runtime readers still tolerate older flat rows with top-level `title`, `type`, `severity`, `source`, `description`, `evidence`, `target_repo`, `id`, and `created_at`. New writers must emit the batch entry shape above.

## Canonical Example

```jsonl
{"source_epic":"na-fr0","timestamp":"2026-03-08T17:30:00Z","items":[{"title":"Publish next-work schema v1.3 and add contract parity checks","type":"tech-debt","severity":"high","source":"council-finding","description":"Collapse next-work queue docs to one tracked v1.3 contract and validate drift against runtime behavior.","evidence":"March 8 audit found the local schema file at v1.2 while runtime and skill docs had already moved to per-item lifecycle semantics.","target_repo":"agentops","consumed":false,"claim_status":"available"}],"consumed":false,"claim_status":"available","claimed_by":null,"claimed_at":null,"consumed_by":null,"consumed_at":null}
```
