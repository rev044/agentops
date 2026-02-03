# Agent Mail (MCP) for Distributed Mode

AgentOps “distributed mode” (`--mode=distributed`) assumes you have an **Agent Mail**-style MCP server available for inter-agent coordination (messages + advisory file reservations).

This repo does **not** require a specific implementation. Any MCP server is fine as long as it provides equivalent capabilities.

## What Distributed Mode Needs

Your Agent Mail MCP server should support:
- **Messaging**: send messages to agents and fetch an inbox (used for `PROGRESS`, `HELP_REQUEST`, `DONE/FAILED`)
- **Agent registration/identity**: so workers can identify themselves to the orchestrator
- **File reservations (advisory)**: optional but recommended to reduce conflicts between parallel workers

Distributed mode also assumes:
- `tmux` is installed (workers run in separate tmux sessions)
- `claude` CLI is available (Claude Code installs this)

## How To Configure

1. Start your Agent Mail MCP server (implementation-specific).
2. Verify it’s reachable:
   - Either via MCP tools visible to the session, **or**
   - Via an HTTP health endpoint (commonly `http://localhost:8765/health`) if your implementation provides one.

Once available, you can use:
- `/swarm --mode=distributed --bead-ids ...`
- `/crank <epic-id> --mode=distributed`

If you don’t need persistence/coordination, use local mode (default) with zero extra dependencies.
