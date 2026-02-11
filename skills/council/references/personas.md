# Perspectives

## Default: Independent Judges (No Perspectives)

When no `--preset` or `--perspectives` flag is provided, all judges get the **same prompt** with no perspective label. Diversity comes from independent sampling, not personality labels.

| Judge | Prompt | Assigned To |
|-------|--------|-------------|
| **Judge 1** | Independent judge — same prompt as all others | Agent 1 |
| **Judge 2** | Independent judge — same prompt as all others | Agent 2 |
| **Judge 3** | Independent judge — same prompt as all others | Agent 3 (--deep/--mixed) |

The default judge prompt (no perspective labels):

```
You are Council Judge {N}. You are one of {TOTAL} independent judges evaluating the same target.

{JSON_PACKET}

Instructions:
1. Analyze the target thoroughly
2. Write your analysis to: .agents/council/{OUTPUT_FILENAME}
   - Start with a JSON code block matching the output_schema
   - Follow with Markdown explanation
3. Send verdict to team lead

Your job is to find problems. A PASS with caveats is less valuable than a specific FAIL.
```

When `--preset` or `--perspectives` is used, judges receive the perspective-labeled prompt instead (see Agent Prompts section).

## Custom Perspectives

Simple name-based:
```bash
/council --perspectives="security,performance,ux" validate the API
```

## Built-in Presets

Use `--preset=<name>` for common persona configurations:

| Preset | Perspectives | Best For |
|--------|-------------|----------|
| `default` | (none — independent judges) | General validation |
| `security-audit` | attacker, defender, compliance | Security review |
| `architecture` | scalability, maintainability, simplicity | System design |
| `research` | breadth, depth, contrarian | Deep investigation |
| `ops` | reliability, observability, incident-response | Operations review |
| `code-review` | error-paths, api-surface, spec-compliance | Code validation (used by /vibe) |
| `plan-review` | missing-requirements, feasibility, scope, spec-completeness | Plan validation (used by /pre-mortem) |
| `retrospective` | plan-compliance, tech-debt, learnings | Post-implementation review (used by /post-mortem) |

```bash
/council --preset=security-audit validate the auth system
/council --preset=research --explorers=3 research upgrade automation
/council --preset=architecture research microservices boundaries
```

**Preset definitions** are built-in perspective configurations.

**Preset perspective details:**

```
security-audit:
  attacker:   "How would I exploit this? What's the weakest link?"
  defender:   "How do we detect and prevent attacks? What's our blast radius?"
  compliance: "Does this meet regulatory requirements? What's our audit trail?"

architecture:
  scalability:     "Will this handle 10x load? Where are the bottlenecks?"
  maintainability: "Can a new engineer understand this in a week? Where's the complexity?"
  simplicity:      "What can we remove? Is this the simplest solution?"

research:
  breadth:     "What's the full landscape? What options exist? What's adjacent?"
  depth:       "What are the deep technical details? What's under the surface?"
  contrarian:  "What's the conventional wisdom wrong about? What's overlooked?"

ops:
  reliability:       "What fails first? What's our recovery time? Where are SPOFs?"
  observability:     "Can we see what's happening? What metrics/logs/traces do we need?"
  incident-response: "When this breaks at 3am, what do we need? What's our runbook?"

code-review:
  error-paths:      "Trace every error handling path. What's uncaught? What fails silently?"
  api-surface:      "Review every public interface. Is the contract clear? Breaking changes?"
  spec-compliance:  "Compare implementation against the spec/bead. What's missing? What diverges?"
  # Note: spec-compliance gracefully degrades to general correctness review when no spec
  # is present in context.spec. The judge reviews code on its own merits.

plan-review:
  missing-requirements: "What's not in the spec that should be? What questions haven't been asked?"
  feasibility:          "What's technically hard or impossible here? What will take 3x longer than estimated?"
  scope:                "What's unnecessary? What's missing? Where will scope creep?"
  spec-completeness:    "Are boundaries defined (Always/Ask First/Never)? Do conformance checks cover all acceptance criteria? Can every acceptance criterion be mechanically verified? Are schema enum values and field names domain-neutral (meaningful in ANY codebase, not just this repo)? Plans without boundaries get WARN, plans with zero conformance checks get FAIL, self-referential schema terms get WARN."

retrospective:
  plan-compliance: "What was planned vs what was delivered? What's missing? What was added?"
  tech-debt:       "What shortcuts were taken? What will bite us later? What needs cleanup?"
  learnings:       "What patterns emerged? What should be extracted as reusable knowledge?"
```
