# OpenAI Function Calling Standards

<!-- Last synced: 2026-01-19 -->

> **Purpose:** Comprehensive function calling and tool definition standards for OpenAI API integrations.

## Scope

This document covers: function schema design, parameter patterns, error handling, structured outputs integration, and debugging tool calls.

**Related:**
- [OpenAI Standards](./openai.md) - Overview and quick reference
- [Prompt Engineering](./openai-prompts.md) - System prompt patterns
- [Responses API](./openai-responses.md) - Agent orchestration

---

## Quick Reference

| Standard | Value | Notes |
|----------|-------|-------|
| **Strict Mode** | Always `strict: true` | Guarantees schema compliance |
| **additionalProperties** | Always `false` | Required for strict mode |
| **All fields required** | Use `null` type for optional | `"type": ["string", "null"]` |
| **Tool limit** | <100 tools | In-distribution performance |
| **Args per tool** | <20 parameters | In-distribution performance |
| **Description placement** | When-to-use FIRST | Critical invocation criteria upfront |

---

## Function Schema Template

### Complete Example

```python
tools = [{
    "type": "function",
    "function": {
        "name": "search_database",
        "description": "Search the product database. Use when user asks about product availability, pricing, or specifications. Returns matching products with details.",
        "strict": True,
        "parameters": {
            "type": "object",
            "properties": {
                "query": {
                    "type": "string",
                    "description": "Search query. Examples: 'red shoes size 10', 'laptop under $500'"
                },
                "category": {
                    "type": ["string", "null"],
                    "enum": ["electronics", "clothing", "home", "sports", None],
                    "description": "Product category to filter. Null for all categories."
                },
                "max_results": {
                    "type": "integer",
                    "description": "Maximum results to return (1-50)",
                    "minimum": 1,
                    "maximum": 50
                },
                "sort_by": {
                    "type": "string",
                    "enum": ["relevance", "price_asc", "price_desc", "rating"],
                    "description": "Sort order for results"
                }
            },
            "required": ["query", "category", "max_results", "sort_by"],
            "additionalProperties": False
        }
    }
}]
```

### Schema Requirements for Strict Mode

```python
# REQUIRED elements for strict: true
{
    "strict": True,
    "parameters": {
        "type": "object",
        "properties": { ... },
        "required": [...],  # ALL properties must be listed
        "additionalProperties": False  # REQUIRED
    }
}
```

---

## Function Description Best Practices

### Structure Template

```
[1-2 sentence purpose]. [When to use]. [What it returns].
```

### Good vs Bad Descriptions

| Aspect | Bad | Good |
|--------|-----|------|
| **Purpose** | "Get weather" | "Get current weather conditions for a location" |
| **When to use** | (missing) | "Use when user asks about current weather, temperature, or conditions" |
| **What returns** | (missing) | "Returns temperature, conditions, humidity, and wind speed" |
| **Examples** | (missing) | "Examples: 'weather in Tokyo', 'is it raining in London'" |

### Complete Good Description

```python
"description": """Get current weather conditions for a location.

USE WHEN: User asks about current weather, temperature, precipitation, or
atmospheric conditions for a specific place.

DO NOT USE WHEN: User asks about weather forecasts (use get_forecast instead)
or historical weather data (use get_historical_weather).

RETURNS: Current temperature, conditions (sunny/cloudy/rain/etc), humidity
percentage, wind speed and direction.

EXAMPLES:
- "What's the weather in Tokyo?" -> location: "Tokyo, Japan"
- "Is it raining in London?" -> location: "London, UK"
- "Temperature in NYC" -> location: "New York City, NY"

NOTES: For ambiguous locations, prefer major cities. If user specifies
a country, use the capital city unless context suggests otherwise."""
```

---

## Parameter Design Patterns

### Optional Parameters (Nullable Types)

```python
# Correct - nullable type for optional
"category": {
    "type": ["string", "null"],
    "description": "Filter by category. Null to search all categories."
}

# Wrong - not in required array
# This will fail strict mode validation
"category": {
    "type": "string",
    "description": "Filter by category"
}
# And category not in "required" array
```

### Enums for Constrained Values

```python
"status": {
    "type": "string",
    "enum": ["pending", "approved", "rejected", "cancelled"],
    "description": "Order status filter"
}
```

### Nested Objects

```python
"address": {
    "type": "object",
    "properties": {
        "street": {"type": "string"},
        "city": {"type": "string"},
        "country": {"type": "string"},
        "postal_code": {"type": ["string", "null"]}
    },
    "required": ["street", "city", "country", "postal_code"],
    "additionalProperties": False  # Required at every level
}
```

### Arrays with Item Schema

```python
"tags": {
    "type": "array",
    "items": {
        "type": "string",
        "description": "A single tag"
    },
    "description": "List of tags to filter by",
    "minItems": 1,
    "maxItems": 10
}
```

---

## Parallel Tool Calls

### When to Enable

```python
response = client.chat.completions.create(
    model="gpt-4.1",
    messages=messages,
    tools=tools,
    parallel_tool_calls=True  # Default: True
)
```

**Enable when:**
- Independent operations (multiple searches)
- Batch processing (get info for multiple items)
- No order dependency between calls

### When to Disable

```python
response = client.chat.completions.create(
    model="gpt-4.1",
    messages=messages,
    tools=tools,
    parallel_tool_calls=False  # Disable for structured outputs
)
```

**Disable when:**
- Using Structured Outputs (incompatible)
- Operations must be sequential
- Using gpt-4.1-nano (can duplicate calls)
- Previous call result needed for next call

---

## Structured Outputs with Functions

### Response Format Integration

```python
from pydantic import BaseModel

class FunctionResult(BaseModel):
    success: bool
    data: dict | None
    error: str | None

# Use structured output for the final response
response = client.chat.completions.create(
    model="gpt-4.1",
    messages=messages,
    tools=tools,
    parallel_tool_calls=False,  # Required for structured outputs
    response_format={
        "type": "json_schema",
        "json_schema": {
            "name": "response",
            "schema": FunctionResult.model_json_schema()
        }
    }
)
```

### Native SDK Support (Recommended)

```python
from pydantic import BaseModel
from openai import OpenAI

class SearchParams(BaseModel):
    query: str
    max_results: int = 10

# SDK handles schema generation
client = OpenAI()
response = client.beta.chat.completions.parse(
    model="gpt-4.1",
    messages=[{"role": "user", "content": "Find laptops under $500"}],
    tools=[{
        "type": "function",
        "function": {
            "name": "search",
            "parameters": SearchParams.model_json_schema()
        }
    }]
)
```

---

## Error Handling Patterns

### Graceful Degradation

```python
def handle_tool_call(tool_call):
    try:
        args = json.loads(tool_call.function.arguments)
        result = execute_function(tool_call.function.name, args)
        return {"success": True, "result": result}
    except json.JSONDecodeError:
        return {"success": False, "error": "Invalid arguments format"}
    except KeyError as e:
        return {"success": False, "error": f"Missing required field: {e}"}
    except Exception as e:
        return {"success": False, "error": str(e)}
```

### Informative Error Messages

```python
# Return structured errors the model can understand
tool_result = {
    "status": "error",
    "error_type": "not_found",
    "message": "No product found with ID 'ABC123'",
    "suggestion": "Try searching by product name instead"
}
```

### Retry Logic

```python
MAX_RETRIES = 3

for attempt in range(MAX_RETRIES):
    response = client.chat.completions.create(...)

    if response.choices[0].finish_reason == "tool_calls":
        tool_calls = response.choices[0].message.tool_calls
        results = [execute_tool(tc) for tc in tool_calls]

        # Check for recoverable errors
        if all(r.get("success") for r in results):
            break
        elif attempt < MAX_RETRIES - 1:
            # Add error context for retry
            messages.append({
                "role": "assistant",
                "content": None,
                "tool_calls": tool_calls
            })
            for tc, result in zip(tool_calls, results):
                messages.append({
                    "role": "tool",
                    "tool_call_id": tc.id,
                    "content": json.dumps(result)
                })
```

---

## Debugging Tool Calls

### Common Issues

| Symptom | Cause | Fix |
|---------|-------|-----|
| Schema validation fails | Missing `additionalProperties: false` | Add to all object levels |
| Optional param errors | Not using nullable type | Use `["string", "null"]` |
| Wrong tool selected | Ambiguous descriptions | Clarify "use when" criteria |
| Hallucinated tool calls | Vague boundaries | Explicit "do not use when" |
| Missing tool calls | Description doesn't match query | Add trigger phrase examples |
| Malformed arguments | Complex nested structure | Flatten schema if possible |

### Validation Checklist

```python
def validate_tool_schema(tool):
    """Validate tool follows best practices."""
    issues = []

    func = tool.get("function", {})

    # Check strict mode
    if not func.get("strict"):
        issues.append("WARN: strict mode not enabled")

    params = func.get("parameters", {})

    # Check additionalProperties
    if params.get("additionalProperties") != False:
        issues.append("ERROR: additionalProperties must be False")

    # Check all properties in required
    props = set(params.get("properties", {}).keys())
    required = set(params.get("required", []))
    if props != required:
        missing = props - required
        issues.append(f"ERROR: Properties not in required: {missing}")

    # Check description quality
    desc = func.get("description", "")
    if len(desc) < 50:
        issues.append("WARN: Description too short")
    if "use when" not in desc.lower():
        issues.append("WARN: Description missing 'use when' criteria")

    return issues
```

---

## Tool Organization Patterns

### Hierarchical Tool Sets

```python
# Group related tools with clear boundaries
SEARCH_TOOLS = [
    {
        "function": {
            "name": "search_products",
            "description": "Search product catalog. Use for finding products by name, category, or attributes."
        }
    },
    {
        "function": {
            "name": "search_orders",
            "description": "Search order history. Use for finding past orders by date, status, or product."
        }
    }
]

ACTION_TOOLS = [
    {
        "function": {
            "name": "place_order",
            "description": "Create new order. Use ONLY after user explicitly confirms purchase."
        }
    },
    {
        "function": {
            "name": "cancel_order",
            "description": "Cancel existing order. Use ONLY with explicit user request and order ID."
        }
    }
]

# Load appropriate subset based on context
tools = SEARCH_TOOLS  # Start with read-only
if user_authenticated and user_confirmed_action:
    tools = SEARCH_TOOLS + ACTION_TOOLS
```

### Decision Boundaries in System Prompt

```markdown
# Tool Usage Guidelines

## Search Tools (use freely)
- search_products: For any product-related queries
- search_orders: For order history questions

## Action Tools (require confirmation)
- place_order: ONLY after user says "yes, place the order" or similar
- cancel_order: ONLY with explicit "cancel order [ID]" request

## Decision Priority
When query could match multiple tools:
1. Prefer search over action
2. Prefer specific over general
3. Ask for clarification if ambiguous
```

---

## Common Errors

| Symptom | Cause | Fix |
|---------|-------|-----|
| `additionalProperties` error | Missing from nested objects | Add to every object level |
| Null handling fails | Using default values | Use nullable types instead |
| Tool not called | Description doesn't match intent | Add example trigger phrases |
| Wrong tool called | Overlapping descriptions | Add explicit "do not use when" |
| Arguments malformed | Deep nesting | Flatten structure |
| Parallel call issues | Using structured outputs | Set `parallel_tool_calls: false` |

---

## Anti-Patterns

| Name | Pattern | Why Bad | Instead |
|------|---------|---------|---------|
| Loose Schema | `strict: false` | Schema violations | Always `strict: true` |
| Missing additionalProperties | Omitting from objects | Strict mode fails | Add to all object levels |
| Vague Description | "Handles user requests" | Model can't decide when to use | Specific "use when" criteria |
| Deep Nesting | 4+ levels of objects | Complex argument construction | Flatten to 2 levels max |
| No Error Handling | Assume success | Crashes on failures | Return structured errors |
| Kitchen Sink | 100+ tools exposed | Latency, confusion | Subset per context |
| Implicit Optional | Not in required, no null type | Validation fails | Explicit nullable types |

---

## AI Agent Guidelines

| Guideline | Rationale |
|-----------|-----------|
| ALWAYS use `strict: true` | Guarantees schema compliance |
| ALWAYS include `additionalProperties: false` | Required for strict mode |
| ALWAYS list all properties in `required` | Strict mode requirement |
| ALWAYS start description with "when to use" | Critical info first |
| ALWAYS use nullable types for optional params | Strict mode compatible |
| NEVER exceed 100 tools | Performance degradation |
| NEVER nest deeper than 2-3 levels | Argument construction errors |
| PREFER enums over free strings | Constrain valid values |
| PREFER flat schemas | Easier for model to construct |
| TEST tool selection empirically | Descriptions affect behavior |

---

## Summary Checklist

| Category | Requirement |
|----------|-------------|
| **Schema** | `strict: true` on all functions |
| **Schema** | `additionalProperties: false` at every object level |
| **Schema** | All properties in `required` array |
| **Optional** | Use `["type", "null"]` for nullable |
| **Description** | "When to use" in first sentence |
| **Description** | Include example trigger phrases |
| **Description** | Add "do not use when" for disambiguation |
| **Parameters** | <20 parameters per function |
| **Tools** | <100 total tools |
| **Parallel** | Disable for structured outputs |
| **Errors** | Return structured, informative errors |
