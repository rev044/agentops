---
name: shared
description: Shared reference documents for distributed mode skills (not directly invocable)
metadata:
  internal: true
---

# Shared References

This directory contains shared reference documents used by multiple skills:

- `agent-mail-protocol.md` - Message protocol for distributed mode coordination
- `validation-contract.md` - Verification requirements for accepting spawned work

These are **not directly invocable skills**. They are loaded by other skills (crank, swarm, inbox, implement) when needed for distributed mode operation.

---

## CLI Availability Pattern

All skills that reference external CLIs MUST degrade gracefully when those CLIs are absent.

### Check Pattern

```bash
# Before using any external CLI, check availability
if command -v bd &>/dev/null; then
  # Full behavior with bd
else
  echo "Note: bd CLI not installed. Using plain text tracking."
  # Fallback: use TaskList, plain markdown, or skip
fi
```

### Fallback Table

| Capability | When Missing | Fallback Behavior |
|------------|-------------|-------------------|
| `bd` | Issue tracking unavailable | Use TaskList for tracking. Note "install bd for persistent issue tracking" |
| `ao` | Knowledge flywheel unavailable | Write learnings to `.agents/knowledge/` directly. Skip flywheel metrics |
| `gt` | Workspace management unavailable | Work in current directory. Skip convoy/sling operations |
| `codex` | CLI missing or model unavailable | Fall back to Claude-only. Council pre-flight checks both CLI presence (`which codex`) and model availability (single test call). Account-type restrictions (e.g. gpt-5.3-codex on ChatGPT accounts) are caught before spawning agents. |
| `cass` | Session search unavailable | Skip transcript search. Note "install cass for session history" |
| Native teams (`TeamCreate`) | API unavailable | Fall back to `Task(run_in_background=true)` fire-and-forget pattern. Council loses debate-via-message (reverts to R2 re-spawn). Swarm loses retry-via-message (reverts to re-spawn). |

### Rules

1. **Never crash** — missing CLI = skip or fallback, not error
2. **Always inform** — tell the user what was skipped and how to enable it
3. **Preserve core function** — the skill's primary purpose must still work without optional CLIs
4. **Progressive enhancement** — CLIs add capabilities, their absence removes them cleanly
