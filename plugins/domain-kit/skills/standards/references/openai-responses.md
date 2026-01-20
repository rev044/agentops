# OpenAI Responses API Standards

<!-- Last synced: 2026-01-19 -->

> **Purpose:** Production patterns for the Responses API, OpenAI's recommended interface for building agents.

## Scope

This document covers: API fundamentals, state management, built-in tools, MCP integration, streaming patterns, and production optimization.

**Related:**
- [OpenAI Standards](./openai.md) - Overview and quick reference
- [Function Calling](./openai-functions.md) - Tool definitions
- [Reasoning Models](./openai-reasoning.md) - o3/o4-mini specifics

---

## Quick Reference

| Feature | Value | Notes |
|---------|-------|-------|
| **Recommended for** | All new agent projects | Future-proof, active development |
| **State persistence** | `store: true` | Maintains reasoning between turns |
| **Context reuse** | `previous_response_id` | 40-80% better cache utilization |
| **Built-in tools** | web_search, file_search, code_interpreter, image_gen | No setup required |
| **MCP support** | Native | Connect to any MCP server |
| **Parallel execution** | Agentic loop | Multiple tools per request |

---

## Basic Usage

### Simple Request

```python
from openai import OpenAI

client = OpenAI()

response = client.responses.create(
    model="gpt-5",
    input="What's the weather like in San Francisco?",
    tools=["web_search"],
)

# Response contains multiple output items
for item in response.output:
    if item.type == "message":
        print(item.content)
    elif item.type == "web_search_call":
        print(f"Searched: {item.query}")
```

### Stateful Conversation

```python
# First turn
response1 = client.responses.create(
    model="gpt-5",
    input="Search for recent AI papers on code generation",
    tools=["web_search"],
    store=True,  # Persist state
)

# Second turn - reuses context
response2 = client.responses.create(
    model="gpt-5",
    input="Summarize the top 3 papers",
    previous_response_id=response1.id,  # Reuse context
    store=True,
)
```

---

## State Management

### Why State Matters

| Without State | With State |
|---------------|------------|
| Rebuilds context each turn | Preserves reasoning tokens |
| 73.9% TAUBench score | 78.2% TAUBench score |
| Higher latency | Lower latency via caching |
| No reasoning continuity | Reasoning persists across tools |

### State Persistence Pattern

```python
class ConversationManager:
    def __init__(self, client, model="gpt-5"):
        self.client = client
        self.model = model
        self.last_response_id = None

    def send(self, message, tools=None):
        response = self.client.responses.create(
            model=self.model,
            input=message,
            tools=tools or [],
            previous_response_id=self.last_response_id,
            store=True,
        )
        self.last_response_id = response.id
        return response

    def reset(self):
        """Start fresh conversation."""
        self.last_response_id = None
```

### Context Window Management

```python
# Check if context is getting large
if response.usage.total_tokens > 100000:
    # Summarize and start fresh
    summary = client.responses.create(
        model="gpt-5",
        input="Summarize our conversation so far in key points",
        previous_response_id=response.id,
    )

    # Start new context with summary
    manager.reset()
    manager.send(f"Context from previous conversation: {summary.output[0].content}")
```

---

## Built-in Tools

### Available Tools

| Tool | Purpose | Use Case |
|------|---------|----------|
| `web_search` | Real-time web search | Current events, recent info |
| `file_search` | Vector search over files | RAG, document Q&A |
| `code_interpreter` | Execute Python code | Data analysis, calculations |
| `image_generation` | Generate images | Visual content creation |

### Web Search

```python
response = client.responses.create(
    model="gpt-5",
    input="What are the latest developments in AI safety?",
    tools=["web_search"],
)

# Output includes search results
for item in response.output:
    if item.type == "web_search_call":
        print(f"Query: {item.query}")
        for result in item.results:
            print(f"  - {result.title}: {result.url}")
```

### File Search (RAG)

```python
# First, upload files to a vector store
vector_store = client.vector_stores.create(name="documentation")
client.vector_stores.file_batches.upload_and_poll(
    vector_store_id=vector_store.id,
    files=[open("docs/api.md", "rb"), open("docs/guide.md", "rb")]
)

# Use file_search tool
response = client.responses.create(
    model="gpt-5",
    input="How do I authenticate with the API?",
    tools=[{
        "type": "file_search",
        "vector_store_ids": [vector_store.id]
    }],
)
```

### Code Interpreter

```python
response = client.responses.create(
    model="gpt-5",
    input="Analyze this CSV data and create a visualization",
    tools=["code_interpreter"],
    attachments=[{
        "file_id": uploaded_file.id,
        "tools": ["code_interpreter"]
    }]
)

# Output may include generated images
for item in response.output:
    if item.type == "image":
        # Save generated visualization
        save_image(item.image_data)
```

---

## MCP Integration

### Basic MCP Connection

```python
response = client.responses.create(
    model="gpt-5",
    input="Search my documents for project updates",
    tools=[{
        "type": "mcp",
        "server_url": "https://my-mcp-server.com/mcp",
    }],
)
```

### Filtering MCP Tools

```python
# Only import specific tools (reduces latency and confusion)
response = client.responses.create(
    model="gpt-5",
    input="Check my calendar for tomorrow",
    tools=[{
        "type": "mcp",
        "server_url": "https://calendar-server.com/mcp",
        "allowed_tools": ["get_events", "check_availability"],  # Filter
    }],
)
```

### Caching MCP Tool Lists

```python
# First request imports tool list
response1 = client.responses.create(
    model="gpt-5",
    input="What tools are available?",
    tools=[{"type": "mcp", "server_url": "..."}],
    store=True,
)

# Subsequent requests reuse cached tool list
response2 = client.responses.create(
    model="gpt-5",
    input="Use the search tool",
    previous_response_id=response1.id,  # Tool list cached
)
```

### Security Considerations

```python
# Require approval for sensitive operations
response = client.responses.create(
    model="gpt-5",
    input="Update my billing settings",
    tools=[{
        "type": "mcp",
        "server_url": "https://billing-server.com/mcp",
        "require_approval": ["update_billing", "delete_account"],  # Approval required
    }],
)

# Check if approval needed
for item in response.output:
    if item.type == "mcp_approval_request":
        # Present to user for confirmation
        if user_confirms(item):
            # Continue with approved action
            response = client.responses.create(
                model="gpt-5",
                input="Proceed with the update",
                previous_response_id=response.id,
                approved_actions=[item.action_id],
            )
```

---

## Mixing Tools

### Custom Functions + Built-in Tools

```python
response = client.responses.create(
    model="gpt-5",
    input="Check inventory and current market prices",
    tools=[
        # Built-in tool
        "web_search",
        # Custom function
        {
            "type": "function",
            "function": {
                "name": "check_inventory",
                "description": "Check product inventory. Use for internal stock levels.",
                "strict": True,
                "parameters": {
                    "type": "object",
                    "properties": {
                        "product_id": {"type": "string"}
                    },
                    "required": ["product_id"],
                    "additionalProperties": False
                }
            }
        }
    ],
)
```

### Tool Decision Boundaries

```markdown
# System Prompt for Mixed Tools

## Tool Selection Rules

### Internal Data (use custom functions)
- check_inventory: Product stock levels
- get_order: Order details
- customer_lookup: Customer information

### External Data (use built-in tools)
- web_search: Market prices, competitor info, news
- file_search: Documentation, policies, procedures

### Decision Priority
1. Use internal tools for company data
2. Use web_search for real-time external info
3. Use file_search for static documentation
4. If unclear, ask user which source to use
```

---

## Streaming Patterns

### Basic Streaming

```python
stream = client.responses.create(
    model="gpt-5",
    input="Write a detailed analysis",
    stream=True,
)

for event in stream:
    if event.type == "content_block_delta":
        print(event.delta.text, end="", flush=True)
    elif event.type == "tool_call":
        print(f"\n[Calling tool: {event.name}]")
```

### Streaming with Tool Calls

```python
async def process_stream(stream):
    current_tool_call = None

    async for event in stream:
        if event.type == "tool_call_start":
            current_tool_call = {"name": event.name, "args": ""}
        elif event.type == "tool_call_delta":
            current_tool_call["args"] += event.delta
        elif event.type == "tool_call_end":
            # Execute tool and continue
            result = await execute_tool(current_tool_call)
            yield {"type": "tool_result", "result": result}
        elif event.type == "content_block_delta":
            yield {"type": "text", "content": event.delta.text}
```

---

## Agent Loop Patterns

### Complete Agent Loop

```python
async def agent_loop(client, initial_message, tools, max_iterations=10):
    messages = [initial_message]
    last_response_id = None

    for iteration in range(max_iterations):
        response = client.responses.create(
            model="gpt-5",
            input=messages,
            tools=tools,
            previous_response_id=last_response_id,
            store=True,
        )
        last_response_id = response.id

        tool_calls = []
        final_message = None

        for item in response.output:
            if item.type == "function_call":
                tool_calls.append(item)
            elif item.type == "message":
                final_message = item.content

        # If no tool calls, we're done
        if not tool_calls:
            return final_message

        # Execute tools and add results
        for call in tool_calls:
            result = await execute_function(call.name, call.arguments)
            messages.append({
                "role": "tool",
                "tool_call_id": call.id,
                "content": json.dumps(result)
            })

    raise Exception("Max iterations reached")
```

### Guardrails Pattern

```python
def apply_output_guardrails(response):
    """Check response against safety rules."""
    content = extract_content(response)

    # Check for prohibited content
    if contains_pii(content):
        return {"blocked": True, "reason": "PII detected"}

    if contains_harmful_content(content):
        return {"blocked": True, "reason": "Content policy violation"}

    # Check for action safety
    for item in response.output:
        if item.type == "function_call":
            if item.name in HIGH_RISK_FUNCTIONS:
                if not has_user_confirmation(item):
                    return {"blocked": True, "reason": "Requires confirmation"}

    return {"blocked": False, "content": content}
```

---

## Production Optimization

### Latency Reduction

```python
# 1. Persist state to enable caching
response = client.responses.create(
    store=True,
    previous_response_id=last_id,  # Reuse context
)

# 2. Filter MCP tools
tools=[{
    "type": "mcp",
    "allowed_tools": ["only", "needed", "tools"],  # Reduce tool list
}]

# 3. Use appropriate model
model="gpt-4.1"  # Faster for simple tasks
model="gpt-5"    # Better for complex reasoning

# 4. Adjust reasoning effort (for o-series)
reasoning_effort="low"  # Faster responses
```

### Cost Optimization

```python
# 1. Use state persistence (40-80% better cache hit rate)
store=True

# 2. Choose appropriate model tier
model="gpt-4.1-nano"  # Cheapest for simple tasks
model="gpt-4.1"       # Balance of cost/capability
model="gpt-5"         # When quality matters most

# 3. Limit tool scope
allowed_tools=["essential", "tools", "only"]

# 4. Monitor usage
print(f"Tokens used: {response.usage.total_tokens}")
```

### Error Handling

```python
from openai import RateLimitError, APIError, Timeout

def robust_request(client, **kwargs):
    max_retries = 3
    base_delay = 1

    for attempt in range(max_retries):
        try:
            return client.responses.create(**kwargs)
        except RateLimitError:
            delay = base_delay * (2 ** attempt)
            time.sleep(delay)
        except Timeout:
            # Retry with longer timeout
            kwargs["timeout"] = kwargs.get("timeout", 60) * 2
        except APIError as e:
            if e.status_code >= 500:
                time.sleep(base_delay)
            else:
                raise

    raise Exception("Max retries exceeded")
```

---

## Common Errors

| Symptom | Cause | Fix |
|---------|-------|-----|
| Lost context between turns | Missing `store: true` | Enable state persistence |
| High latency | Not reusing context | Pass `previous_response_id` |
| MCP tool bloat | Importing all tools | Use `allowed_tools` filter |
| Tool call failures | Missing error handling | Add try/catch around execution |
| Infinite loops | No iteration limit | Set `max_iterations` in agent loop |

---

## Anti-Patterns

| Name | Pattern | Why Bad | Instead |
|------|---------|---------|---------|
| Stateless Requests | Omitting `store: true` | Loses reasoning context | Always persist state |
| Fresh Context | Not using `previous_response_id` | Rebuilds every turn | Chain responses |
| MCP All-Import | No `allowed_tools` filter | Token bloat, latency | Filter to needed tools |
| Blind Trust MCP | No `require_approval` | Security risk | Approval for sensitive ops |
| No Guardrails | Direct output to user | Safety/quality issues | Apply output checks |
| Unbounded Loops | No iteration limit | Runaway agents | Set max iterations |

---

## AI Agent Guidelines

| Guideline | Rationale |
|-----------|-----------|
| ALWAYS use `store: true` for agents | Preserves reasoning between turns |
| ALWAYS chain with `previous_response_id` | 40-80% better cache utilization |
| ALWAYS filter MCP tools | Reduces latency and confusion |
| ALWAYS set iteration limits | Prevents runaway agents |
| ALWAYS implement guardrails | Safety and quality control |
| NEVER trust MCP servers blindly | Security risk |
| NEVER omit error handling | Crashes degrade UX |
| PREFER built-in tools when applicable | No setup, well-tested |
| MONITOR token usage | Cost and context management |
| TEST agent loops thoroughly | Complex failure modes |

---

## Summary Checklist

| Category | Requirement |
|----------|-------------|
| **State** | Enable `store: true` for all agent requests |
| **Context** | Pass `previous_response_id` for continuity |
| **Tools** | Filter with `allowed_tools` for MCP |
| **Security** | Use `require_approval` for sensitive actions |
| **Safety** | Implement output guardrails |
| **Limits** | Set max iterations for agent loops |
| **Errors** | Handle rate limits, timeouts, API errors |
| **Monitoring** | Track token usage and costs |
| **Testing** | Validate agent behavior with evals |
