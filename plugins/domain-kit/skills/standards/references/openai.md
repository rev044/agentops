# OpenAI Standards & Style Guide

<!-- Last synced: 2026-01-19 -->

> **Purpose:** Comprehensive OpenAI API standards for building production-grade AI applications and agents.

## Scope

This document covers: API selection, prompt engineering, function calling, structured outputs, reasoning models, MCP integration, and agent development patterns.

**Related:**
- [Prompt Engineering](./openai-prompts.md) - Detailed prompting techniques
- [Function Calling](./openai-functions.md) - Tool definitions and patterns
- [Responses API](./openai-responses.md) - Agent workflows and state management
- [Reasoning Models](./openai-reasoning.md) - o3/o4-mini specific guidance

---

## Quick Reference

| Standard | Value | Notes |
|----------|-------|-------|
| **Recommended API** | Responses API | Chat Completions still supported |
| **Default Model** | gpt-4.1 or gpt-5 | Pin to specific snapshot in production |
| **Structured Outputs** | Always prefer over JSON mode | Guarantees schema adherence |
| **Function Calling** | `strict: true` always | Enforces schema compliance |
| **Reasoning Models** | o3, o4-mini | Native tool calling in chain-of-thought |
| **Tool Limit** | <100 tools, <20 args/tool | In-distribution performance |

---

## API Selection Matrix

| Use Case | Recommended API | Why |
|----------|-----------------|-----|
| New projects | Responses API | Future-proof, built-in tools, stateful |
| Agentic workflows | Responses API | Multi-tool orchestration, reasoning preservation |
| Simple completions | Chat Completions | Simpler, well-understood |
| Vision/code/retrieval | Responses API | Native multimodal support |
| Legacy migration | Gradual migration | Update incrementally |

### Responses API Benefits

```python
# Responses API: stateful, agentic by default
response = client.responses.create(
    model="gpt-5",
    input="Search for recent AI papers and summarize",
    tools=["web_search", "code_interpreter"],
    store=True,  # Persist state turn-to-turn
)

# Subsequent request preserves reasoning context
response = client.responses.create(
    model="gpt-5",
    input="Now analyze the top 3 papers in detail",
    previous_response_id=response.id,  # Reuse context
)
```

**Performance gains:** 40-80% improved cache utilization, 5% better TAUBench scores vs Chat Completions.

---

## Model Selection

### GPT Models (General Purpose)

| Model | Use Case | Notes |
|-------|----------|-------|
| **gpt-5** | Production default | Best balance of capability/cost |
| **gpt-5.2** | Latest features | Lower verbosity, stronger adherence |
| **gpt-4.1** | Cost-sensitive | Excellent for agentic workflows |
| **gpt-4.1-nano** | High-throughput | Disable parallel tool calls |

### Reasoning Models (o-series)

| Model | Use Case | Notes |
|-------|----------|-------|
| **o3** | Complex multi-step tasks | Native tool calling in CoT |
| **o4-mini** | Cost-efficient reasoning | Same capabilities, smaller |

**Key difference:** Reasoning models think internally before responding. Don't add chain-of-thought prompts.

---

## Structured Outputs vs JSON Mode

| Feature | Structured Outputs | JSON Mode |
|---------|-------------------|-----------|
| Schema adherence | Guaranteed | Best effort |
| Type safety | Full | None |
| Nested objects | Supported | May fail |
| Parallel tool calls | Incompatible | Compatible |
| Recommendation | **Always prefer** | Legacy only |

```python
from pydantic import BaseModel

class Analysis(BaseModel):
    summary: str
    confidence: float
    key_findings: list[str]

response = client.responses.create(
    model="gpt-5",
    input="Analyze this document",
    text={"format": {"type": "json_schema", "json_schema": Analysis.model_json_schema()}}
)
```

---

## Function Calling Quick Reference

### Always Enable Strict Mode

```python
tools = [{
    "type": "function",
    "function": {
        "name": "get_weather",
        "description": "Get current weather for a location. Use when user asks about weather conditions.",
        "strict": True,  # ALWAYS enable
        "parameters": {
            "type": "object",
            "properties": {
                "location": {
                    "type": "string",
                    "description": "City name, e.g., 'San Francisco, CA'"
                },
                "units": {
                    "type": "string",
                    "enum": ["celsius", "fahrenheit"],
                    "description": "Temperature units"
                }
            },
            "required": ["location", "units"],
            "additionalProperties": False  # Required for strict mode
        }
    }
}]
```

### Function Description Best Practices

| Element | Good | Bad |
|---------|------|-----|
| **Purpose** | "Get weather data. Use when user asks about current conditions." | "Weather function" |
| **When to use** | First sentence clarifies invocation criteria | Buried in description |
| **Arguments** | Clear format examples: "City name, e.g., 'San Francisco, CA'" | "The location" |
| **Edge cases** | "If location ambiguous, ask for clarification" | Not mentioned |

---

## Prompt Engineering Principles

### GPT Models (4.1, 5, 5.2)

| Principle | Implementation |
|-----------|----------------|
| **Be specific** | Context, outcome, format, style, length |
| **Use structure** | Markdown headings, XML tags for nesting |
| **Place instructions strategically** | Beginning AND end for long context |
| **Avoid contradictions** | Review prompts for conflicting rules |

### Reasoning Models (o3, o4-mini)

| Principle | Implementation |
|-----------|----------------|
| **Skip chain-of-thought** | Model reasons internally |
| **Use developer messages** | Explicit authority from developer |
| **Control reasoning_effort** | Higher = more exploration, lower = faster |
| **Persist reasoning** | Use Responses API with `store: true` |

---

## Agentic Workflow Patterns

### Three Essential System Prompt Reminders

```markdown
## Persistence
You are in a multi-message workflow. Continue until the task is fully resolved.
Do not yield control prematurely.

## Tool Usage
Use tools when uncertain. If unsure about file content, read it rather than guessing.
Never invent tool calls or promise future calls.

## Planning (Optional)
Before each action, briefly explain your reasoning and next step.
```

### Agent Loop Pattern

```python
while not task_complete:
    response = client.responses.create(
        model="gpt-5",
        input=messages,
        tools=available_tools,
        previous_response_id=last_response_id,
        store=True,
    )

    for item in response.output:
        if item.type == "function_call":
            result = execute_function(item)
            messages.append({"role": "tool", "content": result})
        elif item.type == "message":
            task_complete = check_completion(item.content)

    last_response_id = response.id
```

---

## MCP Integration

### Best Practices

| Practice | Implementation |
|----------|----------------|
| **Limit tool exposure** | Use `allowed_tools` to filter |
| **Cache tool lists** | Pass `previous_response_id` |
| **Trust verification** | Only use trusted MCP servers |
| **Transport selection** | Prefer Streamable HTTP over SSE |

```python
response = client.responses.create(
    model="gpt-5",
    input="Search my documents",
    tools=[{
        "type": "mcp",
        "server_url": "https://my-server.com/mcp",
        "allowed_tools": ["search_docs", "read_file"],  # Filter tools
    }],
)
```

### Security Considerations

- MCP servers can access sensitive data
- Use `require_approval` for sensitive actions
- Implement input validation for user-provided content
- Monitor for prompt injection attempts

---

## Common Errors

| Symptom | Cause | Fix |
|---------|-------|-----|
| Schema validation fails | Missing `additionalProperties: false` | Add to all object types |
| Hallucinated tool calls | Ambiguous tool descriptions | Clarify invocation criteria |
| Inconsistent JSON | Using JSON mode | Switch to Structured Outputs |
| Slow reasoning model | High `reasoning_effort` | Lower for simple tasks |
| Tool list bloat | MCP imports all tools | Use `allowed_tools` filter |
| Lost context between turns | Not persisting state | Use `store: true` and `previous_response_id` |
| Parallel call errors | Structured outputs + parallel | Set `parallel_tool_calls: false` |
| Contradictory behavior | Conflicting instructions | Audit prompt for conflicts |

---

## Anti-Patterns

| Name | Pattern | Why Bad | Instead |
|------|---------|---------|---------|
| JSON Mode Default | Using JSON mode for outputs | No schema guarantees | Use Structured Outputs |
| Loose Functions | `strict: false` in function schemas | Schema violations | Always `strict: true` |
| Prompt Contradiction | "Never X" and "Always X" in same prompt | Wastes reasoning tokens | Explicit priority rules |
| CoT for Reasoning Models | Adding "think step by step" to o3/o4 | Degrades performance | Let model reason internally |
| Tool Hallucination | Vague tool descriptions | Model invents tool calls | Explicit invocation criteria |
| Kitchen Sink Tools | 100+ tools exposed | Latency, confusion | Filter to needed subset |
| State Amnesia | Not using `previous_response_id` | Rebuilds context each turn | Persist state |
| Promise Future Calls | "I'll call X next turn" | Tool never executes | Call immediately or don't mention |

---

## AI Agent Guidelines

| Guideline | Rationale |
|-----------|-----------|
| ALWAYS use Responses API for agents | Better state management, tool orchestration |
| ALWAYS enable `strict: true` for functions | Prevents schema violations |
| ALWAYS prefer Structured Outputs | Guarantees valid output schema |
| ALWAYS persist state with `store: true` | Preserves reasoning between turns |
| NEVER use chain-of-thought with o3/o4-mini | Degrades reasoning performance |
| NEVER expose untrusted MCP servers | Security risk - data exfiltration |
| PREFER `allowed_tools` over full tool lists | Reduces latency and confusion |
| PREFER developer messages for authority | Clear instruction hierarchy |
| PIN model versions in production | Consistent behavior across deploys |
| TEST with evaluations before deploying | LLMs are nondeterministic |

---

## Model-Specific Quick Reference

### GPT-4.1

- Literal instruction following
- Use `tools` parameter, not inline descriptions
- Place instructions at start AND end for long context
- Inducing planning improves SWE-bench by 4%

### GPT-5

- Default `verbosity` parameter available
- Self-optimization: use GPT-5 to refine prompts
- Responses API shows 5% better TAUBench scores
- Internal reasoning with configurable effort

### GPT-5.2

- Lower verbosity by default
- Stronger instruction adherence
- Less drift from user intent
- Explicit parallelism encouragement for scanning operations

### o3/o4-mini

- Native tool calling in chain-of-thought
- Persist reasoning items via Responses API
- <100 tools, <20 args per tool stays in-distribution
- Use developer messages for explicit authority

---

## Summary Checklist

| Category | Requirement |
|----------|-------------|
| **API** | Use Responses API for new agent projects |
| **Models** | Pin to specific snapshot in production |
| **Functions** | `strict: true`, `additionalProperties: false` |
| **Outputs** | Structured Outputs over JSON mode |
| **Prompts** | No contradictions, explicit priorities |
| **Reasoning** | No CoT prompts for o-series models |
| **State** | `store: true` + `previous_response_id` |
| **MCP** | Filter with `allowed_tools`, trust verification |
| **Security** | `require_approval` for sensitive actions |
| **Testing** | Build evals, measure prompt performance |
