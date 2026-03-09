# Finding Compiler Contract

This contract defines how normalized findings are compiled into preventive outputs. The compiler exists to make a finding learned once and consumed many times.

## Canonical Inputs

- Findings live under `.agents/findings/` as Markdown files with YAML frontmatter matching [`finding-artifact.schema.json`](finding-artifact.schema.json).
- Findings are repo-local, human-readable, and git-reviewable. Cross-repo reuse happens by searching files, not by moving them into a service.

## Compiler Targets

| Target | Output path | Purpose |
|--------|-------------|---------|
| `plan` | `.agents/planning-rules/<id>.md` | Prevent the planner from generating known-bad decomposition or sequencing |
| `pre-mortem` | `.agents/pre-mortem-checks/<id>.md` | Surface prior failure modes during spec/plan validation |
| `constraint` | `.agents/constraints/<id>.sh` plus `.agents/constraints/index.json` | Enforce mechanically detectable rules during validation |

## Compile Rules

1. The compiler reads one finding artifact.
2. For each target in `compiler_targets`, it emits the matching output.
3. Advisory findings may emit `plan` and `pre-mortem` outputs.
4. Mechanical findings may emit `constraint` outputs only when concrete detector metadata exists:
   - `enforcement_command`
   - `content_pattern`
   - or another compiler-supported detector field
5. Findings with `status: retired` or `status: superseded` must not leave active downstream outputs behind.

## Constraint Lifecycle

- Executable constraints remain tracked in `.agents/constraints/index.json`.
- Constraint entries must retain:
  - source finding id
  - source artifact path
  - compiled target metadata
  - status (`draft`, `active`, `retired`)
- `ao constraint activate` and `ao constraint retire` remain the lifecycle surface for executable rules.

## Promotion

- Findings are emitted initially as `draft`.
- Advisory outputs can be read immediately by planning and pre-mortem flows when the finding is `active`.
- Mechanical outputs must remain `draft` until explicitly reviewed and activated.

## Retirement and Supersession

- `retired` findings disable downstream outputs for future runs and should remove or skip active advisory artifacts.
- `superseded` findings point to a replacement via `superseded_by`.
- When a finding is superseded, the compiler should prefer the replacement artifact and retire stale downstream outputs from the replaced one.

## Enforcement Expectations

- The compiler is not complete until active compiled constraints are actually consumed by runtime validation.
- Narrative docs must not claim active enforcement unless `hooks/task-validation-gate.sh` loads active compiled rules and executes them.
