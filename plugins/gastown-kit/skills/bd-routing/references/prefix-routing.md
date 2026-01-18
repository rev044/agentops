# Prefix-Based Routing

**For:** AI agents needing to route `bd` commands to the correct database

## Overview

Every bead has an ID with a prefix (e.g., `hq-abc123`, `gt-xyz789`). The `bd` CLI uses this prefix to route commands to the correct beads database.

## How Routing Works

1. You run `bd show gt-abc123`
2. `bd` extracts the prefix: `gt-`
3. Looks up `gt-` in `~/gt/.beads/routes.jsonl`
4. Finds: `{"prefix":"gt-","path":"daedalus/mayor/rig"}`
5. Routes command to `~/gt/daedalus/mayor/rig/.beads/`

## Route Registration

Routes are stored in `~/gt/.beads/routes.jsonl`:

```json
{"prefix":"hq-","path":""}
{"prefix":"gt-","path":"daedalus/mayor/rig"}
{"prefix":"ap-","path":"athena/mayor/rig"}
{"prefix":"ho-","path":"argus/mayor/rig"}
{"prefix":"be-","path":"chronicle/mayor/rig"}
```

**Registration:** Routes are auto-registered by `gt rig add`.

## Common Prefixes

| Prefix | Rig | Description |
|--------|-----|-------------|
| `hq-` | (town) | HQ coordination, mail |
| `gt-` | daedalus | Gas Town multi-agent system |
| `ap-` | athena | AI Platform services |
| `ho-` | argus | Houston orchestration |
| `be-` | chronicle | Beads issue tracker |
| `fr-` | cyclopes | Fractal framework |
| `gitops-` | gitops | GitOps automation |

## Using Prefix Routing

### Reading Beads

```bash
# These automatically route to the correct database
bd show hq-abc123     # Routes to town beads
bd show gt-xyz789     # Routes to daedalus beads
bd list               # Uses current directory's database
```

### Creating Beads

When you create a bead, the prefix is determined by which database you're in:

```bash
cd ~/gt
bd create --title="Task"        # Creates hq-xxxxx

cd ~/gt/daedalus/crew/boden
bd create --title="Task"        # Creates gt-xxxxx
```

### Cross-Database Creation

Use `BEADS_DIR` to target a specific database:

```bash
# From anywhere, create in daedalus's database
BEADS_DIR=~/gt/daedalus/mayor/rig/.beads bd create --title="Task"
# Creates: gt-xxxxx
```

## Debugging Routing

Enable routing debug output:

```bash
BD_DEBUG_ROUTING=1 bd show hq-abc123
```

This shows:
- Which prefix was extracted
- Which route matched
- Final target database path

## Adding New Routes

Routes are normally added via `gt rig add`. For manual registration:

```bash
# Append to routes.jsonl
echo '{"prefix":"new-","path":"new_rig/mayor/rig"}' >> ~/gt/.beads/routes.jsonl
```

## Resolving Conflicts

If two rigs accidentally share a prefix:

```bash
# Check current routes
cat ~/gt/.beads/routes.jsonl | grep "conflicting-"

# Rename one rig's prefix
cd ~/gt/affected_rig/mayor/rig
bd rename-prefix new-prefix-
```

## Edge Cases

### No Route Found

If a prefix isn't registered, `bd` falls back to the current directory's database.

### Empty Path (Town Level)

The town-level route uses an empty path:
```json
{"prefix":"hq-","path":""}
```

This routes to `~/gt/.beads/` directly.

### Multiple Matches

First match wins. Routes are processed in file order.
