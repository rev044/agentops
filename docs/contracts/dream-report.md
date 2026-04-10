# Dream Report Contract

`summary.json` and `summary.md` are the canonical Dream report artifacts.

All future Dream layers must consume this contract instead of inventing parallel formats.

## Required Files

- `summary.json`
- `summary.md`

## Required JSON Fields

| Field | Type | Meaning |
|------|------|---------|
| `schema_version` | integer | Report contract version |
| `mode` | string | Run mode identifier |
| `run_id` | string | Stable run identifier |
| `goal` | string | Optional goal carried into the run |
| `repo_root` | string | Absolute repo path |
| `output_dir` | string | Absolute artifact directory |
| `status` | string | `planned`, `dry-run`, `done`, or `failed` |
| `dry_run` | boolean | Whether execution mutated state |
| `started_at` | string | RFC3339 UTC start time |
| `finished_at` | string | RFC3339 UTC finish time |
| `duration` | string | Human-readable duration |
| `runtime` | object | Runtime/process metadata |
| `steps` | array | Ordered step outcomes |
| `artifacts` | object | Named artifact paths |
| `recommended` | array | Recommended follow-up commands |
| `next_action` | string | Single highest-signal first move |

## Runtime Object

Required fields:

- `keep_awake`
- `keep_awake_mode`
- `requested_timeout`
- `effective_timeout`
- `lock_path`
- `log_path`
- `process_contract_doc`
- `report_contract_doc`

Optional fields:

- `keep_awake_note`

## Step Object

Each step entry must include:

- `name`
- `status`

Optional fields:

- `command`
- `artifact`
- `note`

## Optional Embedded Data

These may be embedded when available:

- `close_loop`
- `metrics_health`
- `retrieval_live`
- `briefing`
- `council`
- `dreamscape`
- `degraded`

The JSON contract must tolerate their absence when the corresponding step soft-fails.

## Markdown Contract

`summary.md` is the human-readable rendering of the same run.

It must answer four questions fast:

1. What state did I wake up to?
2. What ran overnight?
3. What degraded or failed?
4. What should I do first?

Minimum sections:

- status header
- DreamScape, when available
- terrain or health section
- what ran
- Dream Council, when available
- degraded, when applicable
- first move
- recommended commands

## Downstream Consumers

Future consumers of this contract:

- `ao overnight report`
- Dream Setup validation
- Dream Council synthesis
- DreamScape terrain visualization

Those layers must extend the report additively instead of replacing these fields.
