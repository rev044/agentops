---
name: knowledge
description: 'Query knowledge artifacts across all locations. Triggers: "find learnings", "search patterns", "query knowledge", "what do we know about", "where is the plan".'
---


# Knowledge Skill

**YOU MUST EXECUTE THIS WORKFLOW. Do not just describe it.**

Find and retrieve knowledge from past work.

## Execution Steps

Given `$knowledge <query>`:

### Step 1: Search with ao CLI (if available)

```bash
ao know search "<query>" --limit 10 2>/dev/null
```

If results found, read the relevant files.

### Step 2: Search .agents/ Directory

```bash
# Search learnings
grep -r "<query>" .agents/learnings/ 2>/dev/null | head -10

# Search patterns
grep -r "<query>" .agents/patterns/ 2>/dev/null | head -10

# Search research
grep -r "<query>" .agents/research/ 2>/dev/null | head -10

# Search retros
grep -r "<query>" .agents/retros/ 2>/dev/null | head -10
```

### Step 3: Search Plans

```bash
# Local plans
grep -r "<query>" .agents/plans/ 2>/dev/null | head -10

# Global plans
grep -r "<query>" ~/.claude/plans/ 2>/dev/null | head -10
```

### Step 3.5: Search Global Patterns

```bash
# Global patterns (cross-repo knowledge)
grep -r "<query>" ~/.claude/patterns/ 2>/dev/null | head -10
```

Global patterns contain knowledge promoted from any repository via `$learn --global`. These are high-confidence, cross-project learnings.

### Step 3.6: Search Global Learnings

```bash
# Global learnings (cross-repo abstracted knowledge)
grep -r "<query>" ~/.agents/learnings/ 2>/dev/null | head -10
```

Global learnings are abstracted, transferable insights promoted from repo-specific learnings via `$learn --promote` or classified as cross-cutting by `$retro`.

### Step 3.7: Search Global Patterns (new location)

```bash
# Global patterns (new location, cross-repo)
grep -r "<query>" ~/.agents/patterns/ 2>/dev/null | head -10
```

### Step 4: Use Semantic Search (if MCP available)

```
Tool: mcp__smart-connections-work__lookup
Parameters:
  query: "<query>"
  limit: 10
```

### Step 5: Read Relevant Files

For each match found, use the Read tool to get full content.

### Step 6: Synthesize Results

Combine findings into a coherent response:
- What do we know about this topic?
- What learnings are relevant?
- What patterns apply?
- What past decisions were made?

### Step 7: Report to User

Present the knowledge found:
1. Summary of findings
2. Key learnings (with IDs)
3. Relevant patterns
4. Links to source files
5. Confidence level (how much we know)

## Knowledge Locations

| Type | Location | Format |
|------|----------|--------|
| Learnings | `.agents/learnings/` | Markdown |
| Patterns | `.agents/patterns/` | Markdown |
| Research | `.agents/research/` | Markdown |
| Retros | `.agents/retros/` | Markdown |
| Plans | `.agents/plans/` | Markdown |
| Global Plans | `~/.claude/plans/` | Markdown |
| Global Learnings | `~/.agents/learnings/` | Cross-repo abstracted learnings |
| Global Patterns | `~/.agents/patterns/` | Cross-repo reusable patterns |
| Legacy Patterns | `~/.claude/patterns/` | Read-only fallback (deprecated for writes) |

## Key Rules

- **Search multiple locations** - knowledge may be scattered
- **Use ao CLI first** - semantic search is better
- **Fall back to grep** - if ao not available
- **Read full files** - don't just report matches
- **Synthesize** - combine findings into useful answer

## Example Queries

```bash
$knowledge authentication    # Find auth-related learnings
$knowledge "rate limiting"   # Find rate limit patterns
$knowledge kubernetes        # Find K8s knowledge
$knowledge "what do we know about caching"
```

## Examples

### Finding Past Learnings

**User says:** `$knowledge "error handling patterns"`

**What happens:**
1. Agent tries `ao know search "error handling patterns"`, finds 3 matches
2. Agent searches `.agents/learnings/` with grep, finds 5 additional matches
3. Agent searches `.agents/patterns/` for related patterns, finds 2 matches
4. Agent reads all matched files using Read tool
5. Agent synthesizes findings into coherent response
6. Agent reports: "We have 5 learnings about error handling: L1 (always wrap errors), L3 (use typed errors), L12 (log before returning), L15 (context propagation), L22 (retry with backoff)"
7. Agent provides links to source files and confidence level: high (multiple confirmations)

**Result:** Complete knowledge synthesis with 5 specific learnings and 2 related patterns, all with source citations.

### Querying Without ao CLI

**User says:** `$knowledge "database migrations"`

**What happens:**
1. Agent tries `ao know search`, command not found
2. Agent falls back to grep search across `.agents/` directories
3. Agent finds 2 matches in learnings, 1 in research, 0 in patterns
4. Agent reads matched files
5. Agent synthesizes: "Limited knowledge found. L8 recommends using transaction-wrapped migrations. Research doc from 2026-01-20 analyzed migration tools."
6. Agent reports medium confidence (only 2 sources)

**Result:** Knowledge found despite missing ao CLI, with appropriate confidence level based on source count.

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| No results found | Query too specific or knowledge not yet captured | Broaden search terms. Try synonyms. Check if topic was covered in recent work but retro not yet run. Suggest running `$retro` to extract recent learnings. |
| Too many results (overwhelming) | Very broad query term | Narrow query with more specific terms. Filter by date: search only recent learnings. Use semantic search (ao CLI) for better ranking if available. |
| Results lack context | Grep matches found but files don't address query | Read full files, not just matching lines. Synthesize from surrounding context. May need to trace back to original research with `$trace`. |
| Confidence level unclear | Mixed or contradictory sources | Report conflicting information explicitly. Note which sources agree/disagree. Suggest running `$research` to investigate further if critical. |

---

## Scripts

### validate.sh

```bash
#!/usr/bin/env bash
set -euo pipefail
SKILL_DIR="$(cd "$(dirname "$0")/.." && pwd)"
PASS=0; FAIL=0
check() { if bash -c "$2"; then echo "PASS: $1"; PASS=$((PASS + 1)); else echo "FAIL: $1"; FAIL=$((FAIL + 1)); fi; }

check "SKILL.md exists" "[ -f '$SKILL_DIR/SKILL.md' ]"
check "SKILL.md has YAML frontmatter" "head -1 '$SKILL_DIR/SKILL.md' | grep -q '^---$'"
check "name is knowledge" "grep -q '^name: knowledge' '$SKILL_DIR/SKILL.md'"
check "mentions .agents/ directory" "grep -q '\.agents/' '$SKILL_DIR/SKILL.md'"
check "mentions search" "grep -qi 'search' '$SKILL_DIR/SKILL.md'"
check "mentions query" "grep -qi 'query' '$SKILL_DIR/SKILL.md'"
check "mentions knowledge locations" "grep -q 'Knowledge Locations' '$SKILL_DIR/SKILL.md'"

echo ""; echo "Results: $PASS passed, $FAIL failed"
[ $FAIL -eq 0 ] && exit 0 || exit 1
```


