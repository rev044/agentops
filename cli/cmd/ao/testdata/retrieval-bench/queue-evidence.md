---
type: learning
maturity: candidate
confidence: high
utility: 0.78
---
# Queue Execution Notes

Queue execution stays trustworthy when retry metadata is separated from
completion proof and legacy rows remain actionable until evidence exists.

## Retry Ordering Metadata

Top-level `failed_at` is retry ordering metadata. It can change scheduling
priority, but it should not suppress a queue row when no completion proof is
present for that work item.

## Proof-Backed Completion Evidence

Proof-backed completion evidence means a completed run or an evidence-only
closure packet exists for the queue item. Auto-consume should happen only after
that proof is present and verifiable.

## Legacy Queue Compatibility

Legacy queue rows should stay selectable through ingestion and let preflight
decide whether proof-backed completion already satisfied the task.
