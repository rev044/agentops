# OpenAI Reasoning Models Standards

<!-- Last synced: 2026-01-19 -->

> **Purpose:** Production patterns for o3 and o4-mini reasoning models with native tool calling.

## Scope

This document covers: reasoning model selection, developer prompts, tool calling patterns, reasoning effort tuning, and performance optimization.

**Related:**
- [OpenAI Standards](./openai.md) - Overview and quick reference
- [Function Calling](./openai-functions.md) - Tool definitions
- [Responses API](./openai-responses.md) - Agent orchestration

---

## Quick Reference

| Feature | o3 | o4-mini |
|---------|-----|---------|
| **Use Case** | Complex multi-step tasks | Cost-efficient reasoning |
| **Tool Calling** | Native in chain-of-thought | Native in chain-of-thought |
| **Reasoning Effort** | Configurable | Configurable |
| **Tool Limit** | <100 tools, <20 args | <100 tools, <20 args |
| **Best API** | Responses API | Responses API |
| **CoT Prompting** | AVOID | AVOID |

---

## Model Selection

### When to Use Reasoning Models

| Scenario | Recommended | Why |
|----------|-------------|-----|
| Multi-step problem solving | o3 | Deep reasoning with tool access |
| Complex code generation | o3 | Systematic planning and execution |
| Cost-sensitive reasoning | o4-mini | Same capabilities, lower cost |
| Simple queries | GPT-4.1/5 | Reasoning overhead unnecessary |
| Latency-critical | GPT-4.1-nano | Fastest responses |

### o3 vs o4-mini

| Aspect | o3 | o4-mini |
|--------|-----|---------|
| Reasoning depth | Highest | Good |
| Cost | Higher | Lower |
| Latency | Higher | Lower |
| Complex tasks | Best | Good |
| Simple tasks | Overkill | Efficient |

---

## Developer Messages

### Why Developer Messages Matter

Reasoning models distinguish between instruction sources:

| Role | Purpose | Authority |
|------|---------|-----------|
| `developer` | Instructions from application builder | Highest |
| `system` | Legacy, use developer instead | Medium |
| `user` | End user input | Standard |

### Basic Pattern

```python
response = client.responses.create(
    model="o3",
    input=[
        {
            "role": "developer",
            "content": """You are a code analysis assistant.

Rules:
1. Always read files before modifying them
2. Explain your reasoning before making changes
3. Never modify files without explicit user request

When analyzing code:
- Check for security vulnerabilities
- Identify performance issues
- Suggest improvements with rationale
"""
        },
        {
            "role": "user",
            "content": "Analyze the authentication module"
        }
    ],
    tools=[file_tools],
)
```

### Authority Hierarchy

```python
# Clear authority hierarchy in developer message
developer_prompt = """
# Authority Levels

## Absolute Rules (never override)
- Never expose credentials or secrets
- Never delete production data
- Always validate before destructive operations

## Default Behavior (user can override)
- Prefer verbose explanations
- Show intermediate steps
- Ask for confirmation on ambiguous requests

## User Preferences (follow when specified)
- Output format preferences
- Language/tone preferences
- Level of detail
"""
```

---

## Tool Calling Patterns

### Native Tool Calling in Chain-of-Thought

Reasoning models call tools **within** their thinking process:

```python
# The model reasons and calls tools as part of thinking
response = client.responses.create(
    model="o3",
    input="Analyze this codebase for security issues",
    tools=[
        {
            "type": "function",
            "function": {
                "name": "read_file",
                "description": "Read file contents. Use when you need to examine code.",
                "strict": True,
                "parameters": {...}
            }
        },
        {
            "type": "function",
            "function": {
                "name": "search_codebase",
                "description": "Search for patterns in code. Use before modifying files.",
                "strict": True,
                "parameters": {...}
            }
        }
    ],
    store=True,  # Persist reasoning between tool calls
)
```

### Persisting Reasoning Items

Critical for performance - preserve reasoning context:

```python
# First request with tools
response1 = client.responses.create(
    model="o3",
    input="Find all SQL queries in the codebase",
    tools=code_tools,
    store=True,  # Persist state including reasoning
)

# Second request reuses reasoning context
response2 = client.responses.create(
    model="o3",
    input="Now check those queries for injection vulnerabilities",
    previous_response_id=response1.id,  # Reuse reasoning context
    store=True,
)
```

**Without persistence:** Model rebuilds reasoning each turn
**With persistence:** Model continues from previous reasoning state

### Tool Description Best Practices

```python
# Good - Clear invocation criteria at start
{
    "name": "read_file",
    "description": "Read file contents. Use when you need to examine existing code before analysis or modification. Returns file content as string.",
    ...
}

# Bad - Criteria buried in description
{
    "name": "read_file",
    "description": "This function reads a file from the filesystem and returns its contents as a string which can then be analyzed by the model when needed.",
    ...
}
```

---

## Reasoning Effort Parameter

### Controlling Depth vs Speed

```python
# Low effort - fast responses, less exploration
response = client.responses.create(
    model="o3",
    input="Simple factual question",
    reasoning_effort="low",  # Faster, less exploration
)

# Medium effort - balanced (default)
response = client.responses.create(
    model="o3",
    input="Moderate complexity task",
    reasoning_effort="medium",  # Default
)

# High effort - thorough, more exploration
response = client.responses.create(
    model="o3",
    input="Complex multi-step problem",
    reasoning_effort="high",  # More thorough
)
```

### Effort Selection Guidelines

| Task Complexity | Recommended Effort | Notes |
|-----------------|-------------------|-------|
| Simple Q&A | `low` | Minimal reasoning needed |
| Standard tasks | `medium` | Default, balanced |
| Complex analysis | `high` | Multi-step, exploratory |
| Debugging | `high` | Thorough investigation |
| Quick checks | `low` | Speed matters more |

### Dynamic Effort Adjustment

```python
def select_effort(task_description):
    """Select reasoning effort based on task complexity."""
    complexity_indicators = {
        "high": ["analyze", "debug", "investigate", "complex", "multi-step"],
        "low": ["quick", "simple", "check", "verify", "status"]
    }

    task_lower = task_description.lower()

    for indicator in complexity_indicators["high"]:
        if indicator in task_lower:
            return "high"

    for indicator in complexity_indicators["low"]:
        if indicator in task_lower:
            return "low"

    return "medium"  # Default
```

---

## Agentic Workflow Patterns

### Controlling Exploration

```python
# Reduce exploration for focused tasks
developer_prompt = """
# Exploration Guidelines

## When to explore broadly:
- Initial codebase analysis
- Debugging with unknown root cause
- Feature discovery

## When to stay focused:
- Specific file modifications
- Known issue fixes
- User provided exact location

Current mode: FOCUSED
- Only examine files directly related to the task
- Ask for clarification rather than exploring tangentially
- Set a mental budget of 5 tool calls before reassessing
"""
```

### Escape Hatches

```python
# Allow model to proceed despite uncertainty
developer_prompt = """
# Handling Uncertainty

When you encounter situations where:
- Information is incomplete but sufficient to proceed
- Multiple valid approaches exist
- Minor clarification could help but isn't blocking

You may:
1. Document your assumptions clearly
2. Proceed with the most reasonable interpretation
3. Note what additional information would be helpful

DO NOT get stuck asking for clarification on minor details.
Only stop for clarification when:
- Action could cause data loss
- Multiple interpretations lead to very different outcomes
- User explicitly asked for confirmation workflow
"""
```

### Differentiating Safe vs Risky Actions

```python
developer_prompt = """
# Action Safety Classification

## Safe (proceed without confirmation)
- Reading files
- Searching codebase
- Running tests
- Generating reports

## Requires Confirmation
- Modifying files
- Deleting resources
- Sending external requests
- Database writes

## Prohibited
- Credential exposure
- Production data deletion
- Bypassing authentication
- External API calls without explicit request
"""
```

---

## Prompt Anti-Patterns

### AVOID: Chain-of-Thought Prompting

```python
# BAD - Reasoning models already think internally
developer_prompt = """
Think step by step:
1. First, analyze the problem
2. Then, consider alternatives
3. Finally, provide solution
"""

# GOOD - Just state what you want
developer_prompt = """
Analyze the code for security vulnerabilities.
Report findings with severity, location, and fix.
"""
```

### AVOID: Instruction Contradictions

```python
# BAD - Contradictory rules
developer_prompt = """
- Always ask before modifying files
- Automatically fix any issues you find
"""

# GOOD - Clear priority
developer_prompt = """
When you find issues:
1. Read-only analysis: Report without asking
2. Simple fixes: Ask once, then apply to all similar
3. Complex changes: Ask for each modification

Priority: Safety > Efficiency > Automation
"""
```

### AVOID: Over-Specification for Reasoning Models

```python
# BAD - Micromanaging reasoning (works for GPT, not o-series)
developer_prompt = """
Step 1: Read the file
Step 2: Parse the contents
Step 3: Identify functions
Step 4: Analyze each function
Step 5: Score complexity
Step 6: Generate report
"""

# GOOD - State goal, let model reason
developer_prompt = """
Analyze the code complexity in this file.
Output a report with function-level complexity scores
and recommendations for refactoring high-complexity areas.
"""
```

---

## Tool Limits and Performance

### In-Distribution Guidance

| Limit | Value | Notes |
|-------|-------|-------|
| Max tools | ~100 | Performance degrades beyond |
| Max args per tool | ~20 | Schema complexity limit |
| Description length | Keep concise | 1-2 sentences preferred |

### Handling Large Tool Sets

```python
# Filter tools based on task context
def select_tools(task_category):
    """Return relevant tool subset."""
    tool_sets = {
        "code_analysis": [read_file, search_code, get_symbols],
        "code_modification": [read_file, write_file, apply_patch],
        "documentation": [read_file, search_docs, web_search],
        "testing": [run_tests, read_file, search_code],
    }
    return tool_sets.get(task_category, [])

# In developer prompt, clarify tool boundaries
developer_prompt = """
# Available Tool Categories

For this session, you have access to:
- Code analysis tools (read, search)

NOT available:
- File modification tools
- External API tools

If you need capabilities outside current tools, ask the user.
"""
```

---

## Common Errors

| Symptom | Cause | Fix |
|---------|-------|-----|
| Slow responses | High reasoning effort | Lower effort for simple tasks |
| Verbose thinking | CoT prompts | Remove "think step by step" |
| Tool hallucinations | Vague descriptions | Explicit invocation criteria |
| Lost context | Not persisting state | Use `store: true` + `previous_response_id` |
| Contradictory behavior | Conflicting rules | Review prompt for conflicts |
| Over-exploration | No scope limits | Set tool call budgets |

---

## Anti-Patterns

| Name | Pattern | Why Bad | Instead |
|------|---------|---------|---------|
| CoT Prompting | "Think step by step" | Degrades reasoning | State goal directly |
| Micromanaging | Step-by-step instructions | Model reasons better alone | High-level goals |
| Rule Conflicts | Contradictory instructions | Wastes reasoning tokens | Clear priorities |
| No Escape Hatch | Must follow rules exactly | Gets stuck on edge cases | Allow flexibility |
| All Tools | Exposing 100+ tools | Latency, confusion | Subset per context |
| Stateless | Not persisting reasoning | Rebuilds each turn | Use store + previous_id |

---

## AI Agent Guidelines

| Guideline | Rationale |
|-----------|-----------|
| NEVER add chain-of-thought prompts | Model reasons internally |
| NEVER micromanage reasoning steps | Let model plan |
| ALWAYS use developer messages | Clear authority hierarchy |
| ALWAYS persist reasoning state | Context continuity |
| ALWAYS set clear tool boundaries | Reduces confusion |
| PREFER Responses API | Preserves reasoning between tools |
| PREFER medium reasoning effort default | Balance speed/quality |
| PROVIDE escape hatches | Prevents getting stuck |
| FILTER tools by context | Stay within limits |
| TEST with various complexity levels | Validate effort settings |

---

## Comparison with GPT Models

| Aspect | GPT-4.1/5 | o3/o4-mini |
|--------|-----------|------------|
| Reasoning | External (prompt) | Internal (built-in) |
| CoT prompting | Beneficial | Harmful |
| System prompt | Standard | Use developer message |
| Few-shot examples | Helpful | May hurt |
| Planning prompts | +4% SWE-bench | Not needed |
| Tool calling | Separate step | In chain-of-thought |
| State management | Manual | Automatic with Responses API |

---

## Summary Checklist

| Category | Requirement |
|----------|-------------|
| **Messages** | Use `developer` role, not `system` |
| **Prompts** | NO chain-of-thought instructions |
| **Prompts** | State goals, not steps |
| **Prompts** | No contradictory rules |
| **Tools** | <100 tools, <20 args per tool |
| **Tools** | Clear invocation criteria first |
| **State** | `store: true` for all requests |
| **Context** | `previous_response_id` for continuity |
| **Effort** | Match to task complexity |
| **Safety** | Classify actions by risk level |
| **Flexibility** | Provide escape hatches |
