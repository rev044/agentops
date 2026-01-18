# Wave Conflict Detection

Detect file overlaps before parallel execution.

## Step 1: Extract Affected Files

For each issue, extract from description, comments, and "Files affected:" annotations:
```bash
bd show <issue-id>  # Look for file paths
```

## Step 2: Build Conflict Graph

```python
file_map = {}
for issue in selected_issues:
    for file in issue.affected_files:
        file_map.setdefault(file, []).append(issue.id)
# Overlaps = files with len(issues) > 1
```

## Step 3: Partition Using Greedy Coloring

1. Create conflict graph (issues as nodes, edges where they share files)
2. Apply greedy coloring - each color becomes an execution group
3. Same-color groups run in parallel; different colors run sequentially

### Algorithm
1. Sort issues by conflict degree (most conflicts first)
2. Assign lowest available color
3. Same-color issues run in parallel

## Step 4: Present Execution Plan

```
FILE OVERLAPS DETECTED:
  services/gateway/routes.py -> ai-platform-101, ai-platform-102

Execution Plan: 2 sequential groups
  Group 1 (4 issues, parallel): ai-platform-101, 103, 104, 105
  Group 2 (2 issues, parallel): ai-platform-102, 106

Options:
  A) Execute all groups sequentially (recommended)
  B) Execute only Group 1 now
  C) Cancel
```

## Comment-Based Coordination

Sub-agents signal via beads comments:
- `READY_FOR_BATCH_CLOSE` - Success
- `BLOCKER: <reason>` - Failure

## Wave Commit Format

```bash
git commit -m "$(cat <<'EOF'
feat(wave): complete wave of N issues

Closes: ai-platform-xxx, ai-platform-yyy

Wave Summary:
- ai-platform-xxx: Brief description
- ai-platform-yyy: Brief description
EOF
)"
```
