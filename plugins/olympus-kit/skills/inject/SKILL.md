# Inject Skill

Inject relevant knowledge into the current session context.

## Triggers

- "inject knowledge"
- "recall context"
- "what do we know about"
- SessionStart hook (automatic)

## Usage

```bash
# Inject knowledge relevant to current directory
ol inject

# Inject with specific context filter
ol inject --context "authentication"

# Inject in markdown format
ol inject --format markdown --max-tokens 1000

# Inject for specific session
ol inject --session <session-id>
```

## How It Works

1. **Scans knowledge stores:**
   - `.agents/learnings/` - Lessons learned
   - `.agents/patterns/` - Reusable patterns
   - `.agents/olympus/index/sessions.jsonl` - Session history

2. **Ranks by relevance:**
   - Directory context
   - Recency
   - Category matching

3. **Formats output:**
   - Markdown (default for skills)
   - JSONL (for programmatic use)

## Output Formats

### Markdown (--format markdown)
```markdown
## Recent Learnings

### L1: DeepCopy required for K8s CRDs
Every CRD type needs make generate after types.go changes...

## Relevant Patterns

### Wave-Based Parallel Execution
When implementing parallel work...
```

### JSONL (--format jsonl)
```json
{"type":"learning","id":"L1","title":"DeepCopy required","content":"..."}
{"type":"pattern","id":"wave-parallel","title":"Wave-Based Parallel","content":"..."}
```

## Token Budget

The `--max-tokens` flag controls output size:
- Default: 1500 tokens (~6KB)
- SessionStart hook uses 1000 tokens
- Approximately 4 chars per token

## See Also

- `/forge` - Extract knowledge from transcripts
- `/provenance` - Trace knowledge lineage
