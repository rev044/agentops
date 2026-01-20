# OpenAI Prompt Engineering Standards

<!-- Last synced: 2026-01-19 -->

> **Purpose:** Comprehensive prompt engineering techniques for OpenAI models, optimized for agent development.

## Scope

This document covers: prompt structure, delimiter selection, instruction placement, model-specific techniques, and debugging strategies.

**Related:**
- [OpenAI Standards](./openai.md) - Overview and quick reference
- [Function Calling](./openai-functions.md) - Tool definitions
- [Reasoning Models](./openai-reasoning.md) - o3/o4-mini specific

---

## Quick Reference

| Technique | GPT-4.1/5 | o3/o4-mini | Notes |
|-----------|-----------|------------|-------|
| Chain-of-thought | Beneficial | Harmful | Reasoning models think internally |
| System prompt | Required | Use developer messages | Different authority model |
| Few-shot examples | Helpful | May hurt | Test empirically |
| Explicit planning | +4% SWE-bench | Built-in | Only for GPT models |
| Long context placement | Start AND end | Start preferred | Dual placement optimal |

---

## Prompt Structure Framework

### Recommended Organization

```markdown
# Role and Objective
You are a [role] that [primary objective].

# Instructions
## Core Behavior
- [High-level guidance]
- [Key constraints]

## Specific Categories
### [Category 1]
- [Detailed rules]

### [Category 2]
- [Detailed rules]

# Reasoning Steps (GPT models only)
1. First, analyze [X]
2. Then, determine [Y]
3. Finally, produce [Z]

# Output Format
Respond with [format specification].

# Examples
<example>
User: [input]
Assistant: [output]
</example>

# Context
[Reference materials, documents, code]

# Final Instruction
Think carefully step by step about [task]. Then produce [output].
```

---

## Delimiter Selection

### Markdown (Recommended Starting Point)

Best for: General prompts, readable instructions

```markdown
# Main Section

## Subsection
- Bullet points for rules
- Use `backticks` for code/technical terms

### Nested Subsection
1. Numbered steps for procedures
2. Clear progression
```

### XML Tags (Superior for Complex Structures)

Best for: Document collections, nested content, precise boundaries

```xml
<instructions>
  <core_rules>
    <rule priority="1">Always verify before acting</rule>
    <rule priority="2">Ask for clarification when uncertain</rule>
  </core_rules>
</instructions>

<documents>
  <doc id="1" title="User Guide" type="reference">
    Content here...
  </doc>
  <doc id="2" title="API Spec" type="technical">
    More content...
  </doc>
</documents>
```

**Why XML performs well:**
- Enables precise section wrapping
- Supports metadata attributes
- Clear nesting hierarchy
- Model trained on XML-heavy data

### Avoid JSON for Document Collections

JSON performs poorly for structured content. Use XML or Lee format instead:

```
# Bad - JSON for documents
{"documents": [{"id": 1, "content": "..."}]}

# Good - Lee format
ID: 1 | TITLE: User Guide | CONTENT: Full text here...
ID: 2 | TITLE: API Spec | CONTENT: Technical details...
```

---

## Instruction Placement Strategy

### For Long Context (>10K tokens)

Place instructions at **both beginning and end**:

```markdown
# Instructions (Top)
You are analyzing a large codebase. Focus on security vulnerabilities.
Pay special attention to: authentication, input validation, SQL queries.

[... large context block ...]

# Reminder (Bottom)
Remember: Focus on security vulnerabilities. Report findings in this format:
- Location: [file:line]
- Severity: [HIGH/MEDIUM/LOW]
- Description: [what and why]
```

### For Short Context

Single placement at beginning is sufficient:

```markdown
# Instructions
Summarize the following document in 3 bullet points.
Focus on: key findings, recommendations, and risks.

# Document
[content]
```

---

## Agentic Prompt Patterns

### The Three Essential Reminders

Every agentic prompt should include these elements:

```markdown
# Persistence
You are operating in a multi-turn workflow. Continue working until the task
is completely resolved. Do not stop prematurely or yield control back to the
user unless you have:
1. Fully completed the requested task, OR
2. Hit an unrecoverable blocker that requires user input

# Tool Usage
When uncertain about facts, file contents, or current state:
- USE TOOLS to verify rather than guessing
- READ files before modifying them
- VALIDATE assumptions with appropriate queries

Never invent tool calls. Never promise to call a tool "later" - if a tool
is needed, call it now.

# Planning (Optional but Recommended for GPT models)
Before each action, briefly state:
1. What you're trying to accomplish
2. Which tool you'll use and why
3. What you expect to learn/achieve
```

### Context Reliance Specification

Explicitly state whether the model should rely on provided context or internal knowledge:

```markdown
# Context Usage Rules

## Strict Mode (RAG, factual queries)
Answer ONLY based on the provided context. If the answer is not in the
context, say "I don't have information about that in the provided documents."
Do not use your training knowledge.

## Augmented Mode (General assistance)
Use the provided context as your primary source. You may supplement with
general knowledge when:
- The context is incomplete on a topic it references
- The user asks about related concepts not in context
- Formatting or presentation guidance is needed
```

---

## Model-Specific Techniques

### GPT-4.1

**Characteristics:**
- Literal instruction following
- Requires explicit clarification (won't infer intent)
- Strong agentic performance with proper prompting

**Optimal Patterns:**

```markdown
# Role
You are a code review assistant that provides actionable feedback.

# Response Rules
- Be concise: 1-2 sentences per issue
- Be specific: include file:line references
- Be actionable: explain how to fix, not just what's wrong

# Workflow
1. Read the entire diff
2. Identify issues by category (bugs, style, security)
3. Prioritize by severity
4. Format response as checklist

# Common Failure Modes to Avoid
- Do NOT add explanatory prose unless asked
- Do NOT use sample phrases verbatim - vary your language
- If mandatory tool info is missing, ASK rather than hallucinate
```

### GPT-5

**Characteristics:**
- `verbosity` parameter for output length control
- Self-optimization capability
- Improved reasoning with configurable effort

**Optimal Patterns:**

```markdown
# Instructions
[Standard instructions here]

# Verbosity Control
- For explanations: Be thorough, include examples
- For code: Minimal comments, focus on clarity
- For summaries: Bullet points, no prose

# Self-Correction
If you notice an error in your previous response:
1. Acknowledge it briefly
2. Provide the correction
3. Continue without over-apologizing
```

**Meta-Prompting (Using GPT-5 to improve prompts):**

```
Given this prompt that's producing suboptimal results:
[original prompt]

And these example failures:
[failure examples]

What phrases should be added to elicit the desired behavior?
What phrases should be removed to prevent the undesired behavior?
```

### GPT-5.2

**Characteristics:**
- Lower verbosity by default
- Stronger instruction adherence
- Better parallelism for scanning operations

**Optimal Patterns:**

```markdown
# Parallelism Hint (for large operations)
When scanning multiple files or entities, process them in parallel batches
rather than sequentially. Group by: [criteria]

# Verification Steps (for high-impact operations)
Before executing any operation that modifies:
- Orders
- Billing
- Infrastructure
- User data

First: State what you're about to do
Then: Wait for explicit confirmation
```

---

## Chain-of-Thought Techniques

### Basic Pattern (GPT models only)

```markdown
Think carefully step by step:
1. First, identify [what needs analysis]
2. Then, determine [relationships/patterns]
3. Next, evaluate [criteria]
4. Finally, produce [output format]
```

### Audited Improvement Pattern

When a prompt has systematic failures:

1. **Identify failure patterns** in evaluation results
2. **Categorize the failures:**
   - Misunderstood intent
   - Insufficient context gathering
   - Incomplete reasoning
   - Wrong assumptions
3. **Codify fixes** as explicit instructions:

```markdown
# Known Failure Modes - AVOID

## Misunderstood Intent
When user says "clean up the code", this means:
- Remove dead code
- Improve naming
- Add missing type hints
It does NOT mean: refactor architecture, add features, change behavior

## Context Gathering
Before modifying any function:
1. Read the entire file
2. Find all callers (search for function name)
3. Check for tests
NEVER modify without understanding usage
```

---

## Debugging Prompts

### Common Issues and Fixes

| Issue | Diagnosis | Fix |
|-------|-----------|-----|
| Ignores instructions | Instruction buried in middle | Move to top AND bottom |
| Too verbose | No length constraint | Add explicit limits |
| Wrong format | Format unclear | Provide exact template |
| Hallucinated facts | No grounding constraint | Add "only from context" rule |
| Repetitive phrasing | Sample phrases used verbatim | "Vary your language" |
| Premature completion | Missing persistence reminder | Add agentic persistence |

### Iterative Improvement Workflow

```
1. Run prompt on test cases
2. Identify systematic failures
3. Check for:
   - Conflicting instructions
   - Underspecified behavior
   - Missing edge cases
4. Add targeted fixes (not broad rewrites)
5. Add examples for remaining edge cases
6. Re-test and iterate
```

### Conflict Detection

Review prompts for contradictions:

```markdown
# BAD - Contradictory
- "Never schedule without patient consent"
- "Auto-assign appointments without contacting patient"

# GOOD - Explicit priority
- "Patient consent is required for scheduling"
- "Exception: Emergency slots may be auto-assigned with post-notification"
- "If rules conflict, patient consent takes priority"
```

---

## Output Format Specifications

### Markdown Output

```markdown
# Output Format
Respond using Markdown formatting:
- Use **bold** for emphasis
- Use `backticks` for code, file names, and technical terms
- Use headers (##) for sections
- Use bullet lists for multiple items
- Use numbered lists for sequential steps

Do NOT use Markdown unless it adds semantic value.
```

### Structured Data Output

```markdown
# Output Format
Respond with valid JSON matching this schema:
{
  "summary": "string - 1-2 sentence overview",
  "findings": [
    {
      "severity": "HIGH | MEDIUM | LOW",
      "location": "file:line",
      "description": "what and why",
      "fix": "how to resolve"
    }
  ],
  "recommendation": "string - overall recommendation"
}
```

---

## Anti-Patterns

| Name | Pattern | Why Bad | Instead |
|------|---------|---------|---------|
| Instruction Burial | Key rules in middle of long prompt | Gets ignored | Top and bottom placement |
| Contradiction | "Never X" and "Always X" together | Wastes tokens reconciling | Explicit priorities |
| Vague Format | "Respond appropriately" | Inconsistent outputs | Exact format spec |
| CoT for Reasoning | "Think step by step" to o3/o4-mini | Degrades performance | Omit for o-series |
| Sample Phrase Lock | Providing exact phrases to use | Repetitive outputs | "Vary your language" |
| Over-Specification | 50 rules for simple task | Confusion, conflicts | Minimal necessary rules |
| No Escape Hatch | Mandatory behavior with no exceptions | Stuck on edge cases | Define exception conditions |

---

## AI Agent Guidelines

| Guideline | Rationale |
|-----------|-----------|
| ALWAYS test prompts empirically | LLMs are nondeterministic |
| ALWAYS place key instructions at start AND end | Long context attention patterns |
| ALWAYS specify output format explicitly | Consistent, parseable outputs |
| ALWAYS include persistence reminders for agents | Prevents premature completion |
| NEVER use CoT prompting with o3/o4-mini | Built-in reasoning, degrades with prompting |
| NEVER bury critical instructions | Attention patterns miss middle content |
| PREFER XML over JSON for structured content | Better model performance |
| PREFER explicit over implicit rules | Literal instruction following |
| DEBUG by checking for contradictions first | Most common failure cause |
| ITERATE on specific failures, not broad rewrites | Targeted fixes more effective |

---

## Summary Checklist

| Category | Requirement |
|----------|-------------|
| **Structure** | Role -> Instructions -> Format -> Examples -> Context |
| **Placement** | Key instructions at top AND bottom for long context |
| **Delimiters** | Markdown for simple, XML for complex/nested |
| **Agents** | Include persistence, tool usage, and planning reminders |
| **Context** | Explicitly specify reliance rules |
| **Format** | Exact output specification, not vague guidance |
| **Debugging** | Check for contradictions, add targeted fixes |
| **Testing** | Empirical validation on representative cases |
