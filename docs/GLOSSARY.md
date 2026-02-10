# Glossary

Alphabetical reference of AgentOps terms and concepts.

---

### AgentOps
A skills plugin that turns coding agents into autonomous software engineering systems. Provides the RPI workflow, knowledge flywheel, multi-model validation, and parallel execution — all with local-only state. [Full documentation →](../README.md)

### Beads
Git-native issue tracking system accessed via the `bd` CLI. Issues are called beads. Supports dependencies, epics, and parent/child relationships. [Full documentation →](../skills/beads/SKILL.md)

### Brownian Ratchet
The core reliability pattern: spawn parallel agents (chaos), validate with multi-model council (filter), merge to main (ratchet). Progress only moves forward — once a phase passes validation, it stays locked. [Full documentation →](../README.md#the-brownian-ratchet)

### Codex Team
A skill (`/codex-team`) that spawns parallel Codex (OpenAI) execution agents orchestrated by Claude, enabling cross-vendor parallel task execution. [Full documentation →](../skills/codex-team/SKILL.md)

### Council
The core validation primitive. Spawns 2-6 parallel judge agents (Claude and/or Codex) with different perspectives, then consolidates into a single verdict: PASS, WARN, or FAIL. Supports default, `--deep`, and `--mixed` modes. [Full documentation →](../skills/council/SKILL.md)

### Crank
Autonomous epic execution loop. Runs `/swarm` waves repeatedly until all issues in an epic are closed, with retry loops on failure. [Full documentation →](../skills/crank/SKILL.md)

### Epic
A collection of related beads issues that together accomplish a larger goal. Decomposed into waves by `/plan` for parallel execution. [Full documentation →](SKILLS.md#plan)

### Extract
An internal skill that pulls learnings, patterns, and decisions from session transcripts and artifacts into structured knowledge files. [Full documentation →](../skills/extract/SKILL.md)

### Flywheel (Knowledge)
The system for accumulating validated learnings across sessions. Each session forges learnings into `.agents/`; the next session injects them. Intelligence compounds over time. [Full documentation →](ARCHITECTURE.md#knowledge-flywheel)

### Forge
An internal skill that mines session transcripts for knowledge artifacts — decisions, patterns, failures, and fixes — and stores them in `.agents/`. [Full documentation →](../skills/forge/SKILL.md)

### Handoff
A skill (`/handoff`) that creates structured session handoff documents so another agent or future session can continue work with full context. [Full documentation →](../skills/handoff/SKILL.md)

### Hook
A shell script that fires automatically on Claude Code events (session start, git push, task completion, etc.). AgentOps includes 11 hooks that auto-enforce the workflow — blocking pushes without validation, preventing workers from committing, and nudging agents through the lifecycle. [Full documentation →](../hooks/hooks.json)

### Inject
An internal skill triggered at session start that loads relevant prior knowledge from `.agents/` into the current session context. [Full documentation →](../skills/inject/SKILL.md)

### Issue
A discrete unit of trackable work, stored as a bead. Created by `/plan`, executed by `/implement` or `/crank`. Has status, dependencies, and parent/child relationships. [Full documentation →](SKILLS.md#beads)

### Judge
An agent in a council that evaluates work from a specific perspective (security, architecture, correctness, etc.). Judges deliberate asynchronously, then the lead consolidates verdicts. [Full documentation →](../skills/council/SKILL.md)

### Level
A learning progression stage (L1-L5) that indicates the maturity of a knowledge artifact, from raw observation to validated organizational knowledge. [Full documentation →](ARCHITECTURE.md#knowledge-artifacts)

### Pool
A knowledge quality tier — pending, tempered, or promoted. Artifacts start in pending, get tempered through repeated validation and use, and can be promoted to the permanent knowledge base. [Full documentation →](ARCHITECTURE.md#knowledge-artifacts)

### Pre-mortem
A skill (`/pre-mortem`) that simulates failures before implementation begins. Spawns judges focused on integration risks, operational risks, data integrity, and edge cases. Produces a verdict on the plan. [Full documentation →](../skills/pre-mortem/SKILL.md)

### Profile
A documentation grouping for domain-specific workflows and standards. Profiles organize coding standards and validation rules by language or domain. [Full documentation →](../skills/standards/SKILL.md)

### Provenance
An internal skill that traces the lineage and sources of knowledge artifacts — where a learning came from, which sessions produced it, and how it was validated. [Full documentation →](../skills/provenance/SKILL.md)

### Ralph Loop
The execution pattern where every wave gets a new team and every worker gets clean context. Named after the "Ralph Wiggum" fresh-context pattern. No bleed-through between waves; the lead is the only one who commits. [Full documentation →](../README.md#ralph-loops)

### Ratchet
A progress gate checkpoint in the RPI lifecycle. Once a phase passes validation, the ratchet locks it — progress cannot go backward. Implemented by the internal `ratchet` skill and the `ao ratchet` CLI. [Full documentation →](../skills/ratchet/SKILL.md)

### Research
The first phase of the RPI lifecycle. Deep codebase exploration using Explore agents that produce structured findings in `.agents/research/`. [Full documentation →](../skills/research/SKILL.md)

### Retro
A skill (`/retro`) that extracts learnings from completed work — decisions made, patterns discovered, and failures encountered — feeding them back into the knowledge flywheel. [Full documentation →](../skills/retro/SKILL.md)

### RPI
Research, Plan, Implement — the three-phase lifecycle that AgentOps orchestrates. The full `/rpi` skill runs six sub-phases: research, plan, pre-mortem, crank, vibe, and post-mortem. [Full documentation →](ARCHITECTURE.md#the-rpi-workflow)

### Skill
A structured capability loaded by Claude Code (or any Skills-protocol-compatible agent). Each skill has a `SKILL.md` entry point with YAML frontmatter, optional references for progressive disclosure, and optional validation scripts. [Full documentation →](SKILLS.md)

### Swarm
A skill (`/swarm`) that spawns parallel worker agents with fresh context. Each wave gets a new team; the lead validates and commits. Workers never commit directly. [Full documentation →](../skills/swarm/SKILL.md)

### Tempered
A knowledge quality state indicating an artifact has been validated through multiple uses across sessions. Tempered knowledge has higher confidence than pending and can be promoted to the permanent knowledge base. [Full documentation →](ARCHITECTURE.md#knowledge-artifacts)

### Vibe
A skill (`/vibe`) that performs code validation combining complexity analysis with multi-model council review. Checks security, quality, architecture, complexity, testing, accessibility, performance, and documentation. [Full documentation →](../skills/vibe/SKILL.md)

### Wave
A group of parallelizable issues within an epic. Issues in the same wave have no dependencies on each other and can be executed concurrently by a swarm. Waves are ordered by dependency depth. [Full documentation →](../skills/crank/SKILL.md)

### Worker
An agent executing a single task in a swarm. Each worker gets fresh context (no bleed-through from other workers), writes files but never commits — the team lead validates and commits. [Full documentation →](../skills/swarm/SKILL.md)
