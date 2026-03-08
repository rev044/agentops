# Phase Data Contracts

How each consolidated phase passes data to the next. Artifacts are filesystem-based; no in-memory coupling between phases.

| Transition | Output | Extraction | Input to Next |
|------------|--------|------------|---------------|
| → Discovery | Research doc, plan doc, pre-mortem report, epic ID | Latest files in `.agents/research/`, `.agents/plans/`, `.agents/council/`; epic from `bd list --type epic --status open` | `epic_id`, `pre_mortem` verdict, and discovery summary are persisted in phased state |
| Discovery → Implementation | Epic execution context + discovery summary | `phased-state.json` + `.agents/rpi/phase-1-summary.md` | `/crank <epic-id>` with prior-phase context |
| Implementation → Validation | Completed/partial crank status + implementation summary | `bd children <epic-id>` + `.agents/rpi/phase-2-summary.md` | `/vibe` + `/post-mortem` with implementation context |
| Validation → Next Cycle (optional) | Vibe/post-mortem verdicts + harvested follow-up work + queue lifecycle fields (`claim_status`, `claimed_by`, `claimed_at`, `consumed`, `failed_at`) | Latest council reports + `.agents/rpi/next-work.jsonl` | Stop, loop (`--loop`), suggest next `/rpi` (`--spawn-next`), or hand work back to `/evolve` |

Queue lifecycle rule:
- post-mortem writes new entries as available: entry aggregate `consumed=false`, `claim_status="available"`
- consumers treat item lifecycle as authoritative inside `items[]`; omitted item `claim_status` means available
- `/evolve` and `/rpi loop` claim an item before starting a cycle: item `claim_status="in_progress"`
- successful `/rpi` + regression gate finalizes that item claim: item `consumed=true`, `claim_status="consumed"`, `consumed_by`, `consumed_at`
- failed or regressed cycles release the claim back to available state and may stamp item `failed_at` for retry ordering
- consumers may rewrite existing queue lines to claim, release, fail, or consume items after initial write
- the entry aggregate flips to `consumed=true` only after every child item is consumed

Canonical schema contract: [`.agents/rpi/next-work.schema.md`](../../../.agents/rpi/next-work.schema.md) (v1.3)
