# AgentOps vs Compound Engineer

> **Compound Engineer** is Every's coding-agent plugin and workflow built around a `Plan -> Work -> Review -> Compound` loop, with install/sync tooling across Claude Code and other agent runtimes.
>
> *Comparison as of March 2026. See the [Compound Engineer repo](https://github.com/EveryInc/compound-engineering-plugin) for current features.*

---

## At a Glance

| Aspect | Compound Engineer | AgentOps |
|--------|-------------------|----------|
| **Philosophy** | "Each unit of engineering work should make subsequent units easier" | "Knowledge compounds over time" |
| **Core strength** | Structured plan/work/review loop, cross-tool sync | Git-tracked memory, validation gates, autonomous workflow |
| **GitHub** | EveryInc/compound-engineering-plugin | boshu2/agentops |
| **Primary use** | Standardized engineering workflow across tools | Ongoing codebase work with persistent memory and validation |

---

## What Compound Engineer Does Well

### 1. The Workflow Is Very Clear

Compound Engineer names the core loop directly:

```text
Plan -> Work -> Review -> Compound -> Repeat
```

That is a strong mental model. It is easy to teach, easy to adopt, and easy to standardize across a team.

### 2. Cross-Runtime Reach Is Excellent

The upstream repo now supports Claude Code plus experimental conversion or sync for Codex, OpenCode, Cursor-adjacent flows, Copilot, Gemini, Windsurf, Qwen, and more. If your main problem is "make one workflow portable across tools," Compound Engineer is strong here.

### 3. Planning and Review Are First-Class

The repo philosophy is explicit: spend most of the effort on planning and review, not just execution. That is close in spirit to AgentOps' emphasis on failure prevention and quality gates.

### 4. Work Execution Is Structured

Compound Engineer's `work` step is oriented around plans, worktrees, and task tracking rather than free-form prompting. That is a real advantage over looser prompt collections.

---

## The Real Difference

Both projects care about compounding. The distinction is **where the system puts the weight**.

```text
Compound Engineer:
  workflow discipline
  + cross-tool portability
  + explicit plan/work/review loop

AgentOps:
  repo-native memory
  + validation gates
  + tracked execution and replayable artifacts
```

Compound Engineer is strongest when you want a consistent engineering loop that can travel across runtimes.

AgentOps is strongest when you want the repository itself to accumulate memory, issue state, validation history, and learnings that future sessions can mechanically retrieve.

---

## Where AgentOps Goes Further

### Automatic Local Memory, Not Just Workflow Compounding

Compound Engineer explicitly cares about compounding, but AgentOps pushes harder on the persistence mechanism itself:

- learnings are written into `.agents/`
- the `ao` CLI scores and retrieves prior artifacts
- relevant knowledge is injected into later sessions automatically

The difference is not "compound vs no compound." The difference is **workflow compounding** versus **git-native artifact compounding with retrieval and injection**.

### More Explicit Validation Gates

AgentOps adds named gates around the implementation loop:

- `/pre-mortem` before building
- `/vibe` or `/council` after implementation
- `/post-mortem` to extract and score learnings

Compound Engineer has a strong review step. AgentOps turns validation into more explicit named gates with stronger mechanical enforcement.

### Issue Graph and Wave Execution

AgentOps is more opinionated about tracked work:

- `/plan` creates dependency-aware issues through beads
- `/crank` executes unblocked waves
- `/evolve` measures goals and repeats the loop automatically

Compound Engineer is structured, but AgentOps has more repo-native machinery for multi-issue orchestration and repeated operation.

---

## Feature Comparison

| Feature | Compound Engineer | AgentOps | Winner |
|---------|:-----------------:|:--------:|:------:|
| Workflow clarity | ✅ Explicit loop | ✅ Explicit loop | Tie |
| Cross-runtime sync | ✅ Strong | ⚠️ Strong, but less sync-centric | Compound Engineer |
| Planning emphasis | ✅ Core strength | ✅ Core strength | Tie |
| Worktree-oriented execution | ✅ Built in | ✅ Built in | Tie |
| **Cross-session memory** | ⚠️ Compounding is documented, but retrieval is less central | ✅ Git-persisted and injected | **AgentOps** |
| **Knowledge compounding** | ✅ Core concept | ✅ Core concept | Tie |
| **Pre-mortem simulation** | ❌ Not a named workflow primitive | ✅ Built in | **AgentOps** |
| **Validation gates** | ⚠️ Strong review step | ✅ Multi-stage validation | **AgentOps** |
| **Issue graph orchestration** | ⚠️ Task tracking | ✅ Beads + waves + goals | **AgentOps** |

---

## Workflow Comparison

### Compound Engineer Workflow

```text
/ce:plan      -> turn ideas into an implementation plan
     ↓
/ce:work      -> execute with worktrees and task tracking
     ↓
/ce:review    -> review before merge
     ↓
/ce:compound  -> document learnings for later reuse
```

### AgentOps Workflow

```text
/research     -> explore codebase + inject prior knowledge
     ↓
/plan         -> break work into tracked issues
     ↓
/pre-mortem   -> simulate likely failure modes
     ↓
/crank        -> implement in dependency-ordered waves
     ↓
/vibe         -> validate across multiple dimensions
     ↓
/post-mortem  -> extract and score learnings for next time
```

**Key difference:** Compound Engineer gives you a portable operating rhythm. AgentOps gives you a repo-native operating system for memory, validation, and tracked execution.

---

## When to Choose Compound Engineer

- You want a **clean, teachable loop** for engineering work.
- You care about **cross-tool portability** and config sync.
- You want **strong planning/review discipline** without as much local machinery.
- Your team already has its own issue tracking and validation systems.

## When to Choose AgentOps

- You want **automatic retrieval of prior learnings** from the repo itself.
- You want **explicit failure prevention** before implementation.
- You want **issue-graph execution** for multi-step work.
- You want the system to behave more like a **knowledge flywheel** than a workflow shell.

---

## Can They Work Together?

**Yes.** This is one of the better pairings.

- Use Compound Engineer for its workflow ergonomics and cross-runtime sync.
- Use AgentOps for memory, validation, and tracked execution.

If you already like Compound Engineer's loop, AgentOps fits best as the layer that makes that loop accumulate durable repo knowledge and enforce stronger gates.

---

## The Bottom Line

| Dimension | Compound Engineer | AgentOps |
|-----------|-------------------|----------|
| **Optimizes** | Workflow consistency across tools | Repo learning and validation over time |
| **Where the leverage lives** | Process shape and portability | Stored artifacts, retrieval, and gates |
| **Compounding model** | Plan, review, document, reuse | Extract, score, inject, repeat |
| **Best fit** | Portable engineering workflow | Long-running codebase memory system |

**Compound Engineer is the closest philosophical neighbor in this comparison set.**
**AgentOps differentiates by making memory, validation, and execution state more mechanical and repo-native.**

---

<div align="center">

[← Back to Comparisons](README.md) · [vs. GSD →](vs-gsd.md)

</div>
