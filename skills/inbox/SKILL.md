---
name: inbox
tier: solo
description: 'Agent Mail inbox monitoring. Check pending messages, HELP_REQUESTs, and recent completions. Triggers: "inbox", "check mail", "any messages", "show inbox", "pending messages", "who needs help".'
dependencies: []
---

# Inbox Skill

> **Quick Ref:** Monitor Agent Mail from any session. View pending messages, help requests, completions.

**YOU MUST EXECUTE THIS WORKFLOW. Do not just describe it.**

Monitor Agent Mail messages for coordination across agents.

**Requires:** MCP Agent Mail tools OR HTTP endpoint at localhost:8765.

## Invocation

```bash
/inbox              # Show current inbox state
/inbox --watch      # Continuous polling mode
```

## Execution Steps

Given `/inbox [--watch]`:

### Step 1: Check Agent Mail Availability

```bash
# Check if Agent Mail MCP tools are available
# Look for tools starting with mcp__mcp-agent-mail__

# Alternatively, check HTTP endpoint
curl -s http://localhost:8765/health 2>/dev/null && echo "Agent Mail HTTP available" || echo "Agent Mail not running"
```

### Step 2: Determine Agent Identity

```bash
# Check environment for agent identity
if [ -n "$OLYMPUS_DEMIGOD_ID" ]; then
    AGENT_NAME="$OLYMPUS_DEMIGOD_ID"
elif [ -n "$AGENT_NAME" ]; then
    AGENT_NAME="$AGENT_NAME"
else
    # Default to asking or using hostname
    AGENT_NAME="${USER:-unknown}-$(hostname -s 2>/dev/null || echo local)"
fi

# Get project key (current repo path)
PROJECT_KEY=$(git rev-parse --show-toplevel 2>/dev/null || pwd)
echo "Agent: $AGENT_NAME"
echo "Project: $PROJECT_KEY"
```

### Step 3: Fetch Inbox (MCP Method)

**Use MCP tool if available:**

```
Tool: mcp__mcp-agent-mail__fetch_inbox
Parameters:
  project_key: "<project-key>"
  agent_name: "<agent-name>"
```

**Parse results into categories:**
- **Pending:** Messages without acknowledgement
- **HELP_REQUEST:** Messages with subject containing "HELP_REQUEST"
- **Completions:** Messages with subject "OFFERING_READY" or "DONE"

### Step 4: Search for HELP_REQUESTs Needing Response

**Search for unresolved help requests:**

```
Tool: mcp__mcp-agent-mail__search_messages
Parameters:
  project_key: "<project-key>"
  query: "HELP_REQUEST"
```

**Filter to those without HELP_RESPONSE in same thread.**

### Step 5: Get Recent Completions

**Search for done messages:**

```
Tool: mcp__mcp-agent-mail__search_messages
Parameters:
  project_key: "<project-key>"
  query: "OFFERING_READY OR DONE OR COMPLETED"
```

### Step 6: Summarize Threads (Optional)

For active threads with multiple messages:

```
Tool: mcp__mcp-agent-mail__summarize_thread
Parameters:
  project_key: "<project-key>"
  thread_id: "<thread-id>"
```

### Step 7: Display Results

Read `references/output-format.md` for the display template and example session.

## Watch Mode

Read `references/watch-mode.md` for watch mode polling loop, alerting, and message summaries.

## Transport Details

Read `references/transport-reference.md` for MCP tool reference, HTTP fallback, and setup instructions.

## Key Rules

- **Check regularly** - Agents may be waiting for help
- **Prioritize HELP_REQUESTs** - Blocked agents waste resources
- **Acknowledge completions** - Closes the coordination loop
- **Use watch mode** - For active orchestration sessions

---

## References

- **Agent Mail Protocol:** See `skills/shared/agent-mail-protocol.md` for message format specifications
- **Parser (Go):** `cli/internal/agentmail/` - shared parser for all message types
