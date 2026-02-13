# The Complete Guide to Building Skills for Claude

> Source: Anthropic official guide, February 2026. Saved verbatim for reference.

## Contents

- Introduction
- Chapter 1: Fundamentals
- Chapter 2: Planning and Design
- Chapter 3: Testing and Iteration
- Chapter 4: Distribution and Sharing
- Chapter 5: Patterns and Troubleshooting
- Chapter 6: Resources and References
- Reference A: Quick Checklist
- Reference B: YAML Frontmatter
- Reference C: Complete Skill Examples

---

## Introduction

A skill is a set of instructions - packaged as a simple folder - that teaches Claude how to handle specific tasks or workflows. Skills are one of the most powerful ways to customize Claude for your specific needs. Instead of re-explaining your preferences, processes, and domain expertise in every conversation, skills let you teach Claude once and benefit every time.

Skills are powerful when you have repeatable workflows: generating frontend designs from specs, conducting research with consistent methodology, creating documents that follow your team's style guide, or orchestrating multi-step processes. They work well with Claude's built-in capabilities like code execution and document creation. For those building MCP integrations, skills add another powerful layer helping turn raw tool access into reliable, optimized workflows.

This guide covers everything you need to know to build effective skills - from planning and structure to testing and distribution. Whether you're building a skill for yourself, your team, or for the community, you'll find practical patterns and real-world examples throughout.

**What you'll learn:**

- Technical requirements and best practices for skill structure
- Patterns for standalone skills and MCP-enhanced workflows
- Patterns we've seen work well across different use cases
- How to test, iterate, and distribute your skills

**Who this is for:**

- Developers who want Claude to follow specific workflows consistently
- Power users who want Claude to follow specific workflows
- Teams looking to standardize how Claude works across their organization

**Two Paths Through This Guide:**

Building standalone skills? Focus on Fundamentals, Planning and Design, and category 1-2. Enhancing an MCP integration? The "Skills + MCP" section and category 3 are for you. Both paths share the same technical requirements, but you choose what's relevant to your use case.

**What you'll get out of this guide:** By the end, you'll be able to build a functional skill in a single sitting. Expect about 15-30 minutes to build and test your first working skill using the skill-creator.

---

## Chapter 1: Fundamentals

### What is a skill?

A skill is a folder containing:

- **SKILL.md** (required): Instructions in Markdown with YAML frontmatter
- **scripts/** (optional): Executable code (Python, Bash, etc.)
- **references/** (optional): Documentation loaded as needed
- **assets/** (optional): Templates, fonts, icons used in output

### Core design principles

#### Progressive Disclosure

Skills use a three-level system:

- **First level (YAML frontmatter):** Always loaded in Claude's system prompt. Provides just enough information for Claude to know when each skill should be used without loading all of it into context.
- **Second level (SKILL.md body):** Loaded when Claude thinks the skill is relevant to the current task. Contains the full instructions and guidance.
- **Third level (Linked files):** Additional files bundled within the skill directory that Claude can choose to navigate and discover only as needed.

This progressive disclosure minimizes token usage while maintaining specialized expertise.

#### Composability

Claude can load multiple skills simultaneously. Your skill should work well alongside others, not assume it's the only capability available.

#### Portability

Skills work identically across Claude.ai, Claude Code, and API. Create a skill once and it works across all surfaces without modification, provided the environment supports any dependencies the skill requires.

### For MCP Builders: Skills + Connectors

If you already have a working MCP server, you've done the hard part. Skills are the knowledge layer on top - capturing the workflows and best practices you already know, so Claude can apply them consistently.

#### The kitchen analogy

MCP provides the professional kitchen: access to tools, ingredients, and equipment.
Skills provide the recipes: step-by-step instructions on how to create something valuable.

Together, they enable users to accomplish complex tasks without needing to figure out every step themselves.

#### How they work together:

| MCP (Connectivity) | Skills (Knowledge) |
|---|---|
| Connects Claude to your service (Notion, Asana, Linear, etc.) | Teaches Claude how to use your service effectively |
| Provides real-time data access and tool invocation | Captures workflows and best practices |
| What Claude can do | How Claude should do it |

#### Why this matters for your MCP users

**Without skills:**

- Users connect your MCP but don't know what to do next
- Support tickets asking "how do I do X with your integration"
- Each conversation starts from scratch
- Inconsistent results because users prompt differently each time
- Users blame your connector when the real issue is workflow guidance

**With skills:**

- Pre-built workflows activate automatically when needed
- Consistent, reliable tool usage
- Best practices embedded in every interaction
- Lower learning curve for your integration

---

## Chapter 2: Planning and Design

### Start with use cases

Before writing any code, identify 2-3 concrete use cases your skill should enable.

**Good use case definition:**

```
Use Case: Project Sprint Planning
Trigger: User says "help me plan this sprint" or "create sprint tasks"
Steps:
1. Fetch current project status from Linear (via MCP)
2. Analyze team velocity and capacity
3. Suggest task prioritization
4. Create tasks in Linear with proper labels and estimates
Result: Fully planned sprint with tasks created
```

**Ask yourself:**

- What does a user want to accomplish?
- What multi-step workflows does this require?
- Which tools are needed (built-in or MCP?)
- What domain knowledge or best practices should be embedded?

### Common skill use case categories

At Anthropic, we've observed three common use cases:

#### Category 1: Document & Asset Creation

**Used for:** Creating consistent, high-quality output including documents, presentations, apps, designs, code, etc.

**Key techniques:**

- Embedded style guides and brand standards
- Template structures for consistent output
- Quality checklists before finalizing
- No external tools required - uses Claude's built-in capabilities

#### Category 2: Workflow Automation

**Used for:** Multi-step processes that benefit from consistent methodology, including coordination across multiple MCP servers.

**Key techniques:**

- Step-by-step workflow with validation gates
- Templates for common structures
- Built-in review and improvement suggestions
- Iterative refinement loops

#### Category 3: MCP Enhancement

**Used for:** Workflow guidance to enhance the tool access an MCP server provides.

**Key techniques:**

- Coordinates multiple MCP calls in sequence
- Embeds domain expertise
- Provides context users would otherwise need to specify
- Error handling for common MCP issues

### Define success criteria

How will you know your skill is working?

**Quantitative metrics:**

- **Skill triggers on 90% of relevant queries** — Run 10-20 test queries that should trigger your skill. Track how many times it loads automatically vs. requires explicit invocation.
- **Completes workflow in X tool calls** — Compare the same task with and without the skill enabled. Count tool calls and total tokens consumed.
- **0 failed API calls per workflow** — Monitor MCP server logs during test runs. Track retry rates and error codes.

**Qualitative metrics:**

- **Users don't need to prompt Claude about next steps** — During testing, note how often you need to redirect or clarify. Ask beta users for feedback.
- **Workflows complete without user correction** — Run the same request 3-5 times. Compare outputs for structural consistency and quality.
- **Consistent results across sessions** — Can a new user accomplish the task on first try with minimal guidance?

### Technical requirements

#### File structure

```
your-skill-name/
├── SKILL.md              # Required - main skill file
├── scripts/              # Optional - executable code
│   ├── process_data.py   # Example
│   └── validate.sh       # Example
├── references/           # Optional - documentation
│   ├── api-guide.md      # Example
│   └── examples/         # Example
└── assets/               # Optional - templates, etc.
    └── report-template.md # Example
```

#### Critical rules

**SKILL.md naming:**

- Must be exactly `SKILL.md` (case-sensitive)
- No variations accepted (`SKILL.MD`, `skill.md`, etc.)

**Skill folder naming:**

- Use kebab-case: `notion-project-setup`
- No spaces, underscores, or capitals

**No README.md:**

- Don't include README.md inside your skill folder
- All documentation goes in SKILL.md or references/

### YAML frontmatter: The most important part

The YAML frontmatter is how Claude decides whether to load your skill. Get this right.

#### Minimal required format

```yaml
---
name: your-skill-name
description: What it does. Use when user asks to [specific phrases].
---
```

#### Field requirements

**name** (required):

- kebab-case only
- No spaces or capitals
- Should match folder name

**description** (required):

- MUST include BOTH: What the skill does AND When to use it (trigger conditions)
- Under 1024 characters
- No XML tags (`<` or `>`)
- Include specific tasks users might say
- Mention file types if relevant

**license** (optional):

- Use if making skill open source
- Common: MIT, Apache-2.0

**allowed-tools** (optional):

- Restrict tool access

**compatibility** (optional):

- 1-500 characters
- Indicates environment requirements

**metadata** (optional):

- Any custom key-value pairs
- Suggested: author, version, mcp-server

#### Security restrictions

**Forbidden in frontmatter:**

- XML angle brackets (`<` `>`)
- Skills with "claude" or "anthropic" in name (reserved)

### Writing effective skills

#### The description field

**Structure:** `[What it does] + [When to use it] + [Key capabilities]`

**Examples of good descriptions:**

```yaml
# Good - specific and actionable
description: Analyzes Figma design files and generates developer handoff documentation. Use when user uploads .fig files, asks for "design specs", "component documentation", or "design-to-code handoff".

# Good - includes trigger phrases
description: Manages Linear project workflows including sprint planning, task creation, and status tracking. Use when user mentions "sprint", "Linear tasks", "project planning", or asks to "create tickets".
```

**Examples of bad descriptions:**

```yaml
# Too vague
description: Helps with projects.

# Missing triggers
description: Creates sophisticated multi-page documentation systems.
```

#### Best Practices for Instructions

**Be Specific and Actionable:**

```
# Good
Run `python scripts/validate.py --input {filename}` to check data format.

# Bad
Validate the data before proceeding.
```

**Include error handling:**

```markdown
# Common Issues

# MCP Connection Failed
If you see "Connection refused":
1. Verify MCP server is running
2. Confirm API key is valid
3. Try reconnecting
```

**Reference bundled resources clearly:**

```
Before writing queries, consult `references/api-patterns.md` for:
- Rate limiting guidance
- Pagination patterns
- Error codes and handling
```

**Use progressive disclosure:**

Keep SKILL.md focused on core instructions. Move detailed documentation to `references/` and link to it.

---

## Chapter 3: Testing and Iteration

### Recommended Testing Approach

#### 1. Triggering tests

**Goal:** Ensure your skill loads at the right times.

```
Should trigger:
- "Help me set up a new ProjectHub workspace"
- "I need to create a project in ProjectHub"
- "Initialize a ProjectHub project for Q4 planning"

Should NOT trigger:
- "What's the weather in San Francisco?"
- "Help me write Python code"
```

#### 2. Functional tests

**Goal:** Verify the skill produces correct outputs.

#### 3. Performance comparison

**Goal:** Prove the skill improves results vs. baseline.

### Iteration based on feedback

**Undertriggering signals:**

- Skill doesn't load when it should
- Users manually enabling it

Solution: Add more detail and nuance to the description.

**Overtriggering signals:**

- Skill loads for irrelevant queries
- Users disabling it

Solution: Add negative triggers, be more specific.

**Execution issues:**

- Inconsistent results
- API call failures

Solution: Improve instructions, add error handling.

---

## Chapter 4: Distribution and Sharing

### Current distribution model (January 2026)

**How individual users get skills:**

1. Download the skill folder
2. Zip the folder (if needed)
3. Upload to Claude.ai via Settings > Capabilities > Skills
4. Or place in Claude Code skills directory

**Organization-level skills:**

- Admins can deploy skills workspace-wide
- Automatic updates
- Centralized management

### An open standard

Skills are published as an open standard. Portable across tools and platforms.

### Using skills via API

- `/v1/skills` endpoint for listing and managing skills
- Add skills to Messages API requests via the `container.skills` parameter
- Version control and management through the Claude Console
- Works with the Claude Agent SDK for building custom agents

---

## Chapter 5: Patterns and Troubleshooting

### Pattern 1: Sequential workflow orchestration

Use when users need multi-step processes in a specific order. Explicit step ordering, dependencies between steps, validation at each stage, rollback instructions for failures.

### Pattern 2: Multi-MCP coordination

Use when workflows span multiple services. Clear phase separation, data passing between MCPs, validation before moving to next phase.

### Pattern 3: Iterative refinement

Use when output quality improves with iteration. Explicit quality criteria, iterative improvement, validation scripts.

### Pattern 4: Context-aware tool selection

Use when same outcome requires different tools depending on context. Clear decision criteria, fallback options, transparency about choices.

### Pattern 5: Domain-specific intelligence

Use when your skill adds specialized knowledge beyond tool access. Domain expertise embedded in logic, compliance before action, comprehensive documentation.

### Troubleshooting

#### Skill won't upload

- **"Could not find SKILL.md"** — File not named exactly SKILL.md (case-sensitive)
- **"Invalid frontmatter"** — YAML formatting issue (missing `---` delimiters, unclosed quotes)
- **"Invalid skill name"** — Name has spaces or capitals

#### Skill doesn't trigger

Revise description field. Check: Is it too generic? Does it include trigger phrases? Does it mention relevant file types?

#### Skill triggers too often

Add negative triggers. Be more specific. Clarify scope.

#### MCP connection issues

1. Verify MCP server is connected
2. Check authentication
3. Test MCP independently
4. Verify tool names

#### Instructions not followed

1. Instructions too verbose — use bullet points, move detail to references/
2. Instructions buried — put critical instructions at top
3. Ambiguous language — be specific about validation criteria
4. Use scripts for critical validations — code is deterministic

#### Large context issues

1. Optimize SKILL.md size — keep under 5,000 words
2. Reduce enabled skills if >20-50 simultaneously

---

## Chapter 6: Resources and References

### Official Documentation

- Best Practices Guide
- Skills Documentation
- API Reference
- MCP Documentation

### Tools and Utilities

- **skill-creator skill** — Generate skills from descriptions, review and improve existing skills

---

## Reference A: Quick Checklist

### Before you start

- [ ] Identified 2-3 concrete use cases
- [ ] Tools identified (built-in or MCP)
- [ ] Planned folder structure

### During development

- [ ] Folder named in kebab-case
- [ ] SKILL.md file exists (exact spelling)
- [ ] YAML frontmatter has `---` delimiters
- [ ] name field: kebab-case, no spaces, no capitals
- [ ] description includes WHAT and WHEN
- [ ] No XML tags (`< >`) anywhere
- [ ] Instructions are clear and actionable
- [ ] Error handling included
- [ ] Examples provided
- [ ] References clearly linked

### Before upload

- [ ] Tested triggering on obvious tasks
- [ ] Tested triggering on paraphrased requests
- [ ] Verified doesn't trigger on unrelated topics
- [ ] Functional tests pass

### After upload

- [ ] Test in real conversations
- [ ] Monitor for under/over-triggering
- [ ] Collect user feedback
- [ ] Iterate on description and instructions

---

## Reference B: YAML Frontmatter

### Required fields

```yaml
---
name: skill-name-in-kebab-case
description: What it does and when to use it. Include specific trigger phrases.
---
```

### All optional fields

```yaml
name: skill-name
description: [required description]
license: MIT
allowed-tools: "Bash(python:*) Bash(npm:*) WebFetch"
compatibility: "Requires Claude Code"
metadata:
  author: Company Name
  version: 1.0.0
  mcp-server: server-name
  category: productivity
  tags: [project-management, automation]
```

### Security notes

**Allowed:**

- Any standard YAML types (strings, numbers, booleans, lists, objects)
- Custom metadata fields
- Long descriptions (up to 1024 characters)

**Forbidden:**

- XML angle brackets (`<` `>`) — security restriction
- Code execution in YAML (uses safe YAML parsing)
- Skills named with "claude" or "anthropic" prefix (reserved)
