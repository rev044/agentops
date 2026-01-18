# Crew vs Polecat - When to Use Which

Understanding the two worker types in Gas Town.

## Core Distinction

| Aspect | Crew | Polecat |
|--------|------|---------|
| **Management** | Human-guided | Witness-managed |
| **Autonomy** | Wait for confirmation | Auto-execute |
| **Persistence** | Long-lived | Ephemeral |
| **Scope** | Flexible, multi-issue | Single issue |
| **Identity** | Named (dave, emma) | Auto-generated |
| **Permission** | `default` | `auto` |

---

## When to Use Crew

**Crew is ideal for:**

1. **Interactive Development**
   - You want to discuss approaches
   - Requirements need clarification
   - Architecture decisions required

2. **Long-Running Projects**
   - Multi-day or multi-week work
   - Complex features with evolving requirements
   - Exploratory or research work

3. **Human Oversight Required**
   - Sensitive code changes
   - Security-critical implementations
   - Breaking changes needing review

4. **Learning and Exploration**
   - Understanding new codebases
   - Experimenting with approaches
   - Iterative design

**Example**: Building a new authentication system with iterative feedback.

---

## When to Use Polecats

**Polecats are ideal for:**

1. **Batch Processing**
   - Multiple independent issues
   - Well-defined, single-issue tasks
   - Routine bug fixes

2. **Parallel Execution**
   - Epic wave execution via `/crank`
   - Multiple issues without dependencies
   - Maximizing throughput

3. **Autonomous Work**
   - Clear requirements, no ambiguity
   - Standard patterns
   - Test additions, doc updates

4. **Context Isolation**
   - Keep orchestrator context clean
   - 100x context reduction vs Task()
   - Long-running background work

**Example**: Implementing 10 independent test files from an epic.

---

## Hook Behavior Comparison

### Crew (Human-Guided)

```
Session Start:
1. Check hook (gt hook)
2. If work hooked:
   - Show the hooked work to human
   - Explain the task
   - WAIT for human confirmation
3. If not hooked:
   - Wait for human instructions
```

### Polecat (Autonomous)

```
Session Start:
1. Check hook (gt hook)
2. If work hooked:
   - IMMEDIATELY begin execution
   - No confirmation needed
   - Push before saying done
3. If not hooked:
   - Error state (polecats always have work)
```

---

## Communication Patterns

| | Crew | Polecat |
|-|------|---------|
| **Human interaction** | Direct, interactive | None |
| **Progress reporting** | Real-time conversation | Beads comments |
| **Questions** | Ask human directly | File as blocking issue |
| **Completion** | Tell human | Close bead, push branch |

---

## Decision Matrix

| Scenario | Use | Why |
|----------|-----|-----|
| "Help me understand this code" | Crew | Interactive exploration |
| "Implement these 5 features" | Polecat | Parallel batch execution |
| "Design a new API" | Crew | Needs discussion |
| "Add tests for existing code" | Polecat | Well-defined, repeatable |
| "Debug this tricky issue" | Crew | Needs investigation |
| "Run this epic overnight" | Polecat | `/crank` autonomous |
| "Refactor with my guidance" | Crew | Iterative feedback |
| "Fix these 10 linter warnings" | Polecat | Independent, trivial |

---

## Hybrid Workflows

### Crew Plans, Polecats Execute

1. **Crew creates the plan**: Research, design, create beads
2. **Mayor dispatches**: `gt sling` issues to polecats
3. **Polecats execute**: Parallel autonomous work
4. **Crew reviews**: Human oversight on merged work

### Context Handoff

When crew context gets heavy, spawn polecats for remaining work:

```bash
# Crew identifies remaining issues
bd ready --parent=epic-123

# Dispatch to polecats
gt sling issue-a gastown
gt sling issue-b gastown
gt sling issue-c gastown

# Monitor from crew (low context cost)
gt convoy list
```

---

## Summary

**Crew**: Interactive, human-guided, persistent, flexible scope.

**Polecat**: Autonomous, witness-managed, ephemeral, single issue.

**Rule of Thumb**: Use crew when you need human judgment. Use polecats when you need parallel throughput.
