# Script Contracts

## Builder Boundary

The current product slice still expects packet refresh builders under `WORKSPACE/.agents/scripts/` when `ao knowledge activate` needs to refresh the evidence substrate.

### Packet Builders

- `source_manifest_build.py`
- `topic_packet_build.py`
- `corpus_packet_promote.py`
- `knowledge_chunk_build.py`

### Native Activation Surfaces

These product surfaces are implemented inside the `ao` binary and no longer require workspace-local Python builders:

- `ao knowledge beliefs`
- `ao knowledge playbooks`
- `ao knowledge brief --goal "<goal>"`
- `ao knowledge gaps`

## Command Ownership

### `ao knowledge activate`

Runs the full outer loop:

1. source manifests
2. topic packets
3. promoted packets
4. chunk bundles
5. native belief book build
6. native playbook build
7. optional native briefing build for `--goal`

### `ao knowledge beliefs`

Refreshes only the belief book.

### `ao knowledge playbooks`

Refreshes candidate playbooks from healthy topics.

### `ao knowledge brief --goal "<goal>"`

Compiles a goal-time briefing.

### `ao knowledge gaps`

Reads generated artifacts and reports thin topics, promotion gaps, weak claims, and next recommended work.

## Roadmap Boundary

This slice now splits responsibility:

- packet refresh remains workspace-local while the corpus contracts keep moving
- belief/playbook/brief/gap surfaces are durable `ao`-native product surfaces

The skill contract stays stable across that boundary.
