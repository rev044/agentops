# openai-docs

Use OpenAI docs MCP tools from Codex with a single, unambiguous install flow.

## Codex Execution Profile

1. Treat `skills/openai-docs/SKILL.md` as canonical content policy.
2. Prefer `mcp__openaiDeveloperDocs__search_openai_docs` and `mcp__openaiDeveloperDocs__fetch_openai_doc`.
3. Keep fallback web browsing restricted to OpenAI domains.

## Guardrails

1. Present one Codex install path for MCP server setup; avoid duplicated setup sections.
2. Keep citations explicit in final answers.
3. If MCP is unavailable, provide a clear next command and retry step.
