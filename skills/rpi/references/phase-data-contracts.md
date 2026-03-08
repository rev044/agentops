# Phase Data Contracts

How each consolidated phase passes data to the next. Artifacts are filesystem-based; no in-memory coupling between phases.

| Transition | Output | Extraction | Input to Next |
|------------|--------|------------|---------------|
| → Discovery | Goal string + repo execution profile contract | Goal from the `/rpi` invocation; repo policy from `docs/contracts/repo-execution-profile.md` and `repo-execution-profile.schema.json` | `repo_profile` state is loaded before research/planning begins |
| Discovery → Implementation | Epic execution context + discovery summary + `execution_packet` | `phased-state.json` + `.agents/rpi/phase-1-summary.md` + `.agents/rpi/execution-packet.json` | `/crank <epic-id>` with repo policy, contract surfaces, and validation bundle already normalized |
| Implementation → Validation | Completed/partial crank status + implementation summary + `execution_packet` | `bd children <epic-id>` + `.agents/rpi/phase-2-summary.md` + `.agents/rpi/execution-packet.json` | `/vibe` + `/post-mortem` with the same repo execution profile fields and done criteria |
| Validation → Next Cycle (optional) | Vibe/post-mortem verdicts + harvested follow-up work + queue lifecycle fields (`claim_status`, `claimed_by`, `claimed_at`, `consumed`, `failed_at`) | Latest council reports + `.agents/rpi/next-work.jsonl` | Stop, loop (`--loop`), suggest next `/rpi` (`--spawn-next`), or hand work back to `/evolve` |

Execution packet v1 should remain additive. Recommended fields:
- `objective`
- `contract_surfaces`
- `validation_commands`
- `tracker_mode`
- `done_criteria`

Queue lifecycle rule:
- post-mortem writes new entries as available: entry aggregate `consumed=false`, `claim_status="available"`
- consumers treat item lifecycle as authoritative inside `items[]`; omitted item `claim_status` means available
- `/evolve` and `/rpi loop` claim an item before starting a cycle: item `claim_status="in_progress"`
- successful `/rpi` + regression gate finalizes that item claim: item `consumed=true`, `claim_status="consumed"`, `consumed_by`, `consumed_at`
- failed or regressed cycles release the claim back to available state and may stamp item `failed_at` for retry ordering
- consumers may rewrite existing queue lines to claim, release, fail, or consume items after initial write
- the entry aggregate flips to `consumed=true` only after every child item is consumed

Canonical schema contract: [`.agents/rpi/next-work.schema.md`](../../../.agents/rpi/next-work.schema.md) (v1.3)
