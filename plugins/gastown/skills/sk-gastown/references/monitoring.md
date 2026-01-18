# Convoy Monitoring

## Overview

Monitor batch progress via `gt convoy status`. Output is small and fixed-size regardless of work complexity.

## Check Convoy Status

```bash
# Get convoy status
gt convoy status hq-cv-xyz

# Output (~10-20 lines, always):
# 🚚 hq-cv-xyz: Wave 1: Feature X
# Progress: 2/3 complete
#
# Tracked:
#   ✓ gt-abc: Implemented OAuth config
#   ✓ gt-def: Added token refresh
#   ● gt-ghi: In progress...
```

## Polling Loop

```python
# Pseudocode for monitoring
def monitor_convoy(convoy_id: str, timeout_minutes: int = 60):
    start = time.now()
    poll_interval = 30  # seconds

    while True:
        # Get status (small output)
        status = bash(f"gt convoy status {convoy_id}")

        # Parse progress
        match = re.search(r"Progress: (\d+)/(\d+)", status)
        if match:
            completed = int(match.group(1))
            total = int(match.group(2))

            if completed == total:
                return "complete"

        # Check timeout
        if time.now() - start > timeout_minutes * 60:
            return "timeout"

        # Check for blockers in status
        if "BLOCKER" in status:
            return "blocked"

        sleep(poll_interval)
```

## Bash Polling

```bash
# Polling loop in bash
convoy_id="hq-cv-xyz"
timeout_seconds=$((60 * 30))  # 30 minutes
start_time=$(date +%s)

while true; do
    status=$(gt convoy status $convoy_id 2>&1)

    # Check if all complete
    if echo "$status" | grep -qE "Progress: ([0-9]+)/\1 complete"; then
        echo "Convoy landed"
        break
    fi

    # Check timeout
    now=$(date +%s)
    if ((now - start_time > timeout_seconds)); then
        echo "TIMEOUT: Convoy not landing"
        break
    fi

    sleep 30
done
```

## Dashboard View

```bash
# List all active convoys
gt convoy list

# Output:
# 🚚 Active Convoys:
#
# hq-cv-xyz: Wave 1: Feature X  [2/3] ●
# hq-cv-abc: Bug fixes          [5/5] ✓ landed
```

## Detecting Completion

Two methods:

```bash
# Method 1: Progress string
gt convoy status hq-cv-xyz | grep "Progress:"
# → "Progress: 3/3 complete"

# Method 2: Count checkmarks
completed=$(gt convoy status hq-cv-xyz | grep -c "✓")
total=$(gt convoy status hq-cv-xyz | grep -c "gt-")
if [ "$completed" -eq "$total" ]; then
    echo "Convoy landed"
fi
```

## Detecting Issues

```bash
# Check for stuck issues (in progress too long)
status=$(gt convoy status hq-cv-xyz)

# Issues marked with ● are in progress
in_progress=$(echo "$status" | grep "●" | awk '{print $2}')

for issue in $in_progress; do
    # Check last update time
    updated=$(bd show $issue --json 2>/dev/null | jq -r '.updated_at')
    # If stale, investigate
done
```

## Context Cost

| Operation | Tokens |
|-----------|--------|
| `gt convoy status` | ~50-100 |
| `gt convoy list` | ~20-50 |
| Poll every 30s for 30min | ~3000 total |

Compare to Task() agents returning 80K tokens. **Massive savings.**

## Best Practices

1. **Poll at 30s intervals** - Balance between responsiveness and overhead
2. **Set reasonable timeouts** - 30-60 minutes typical
3. **Watch for blockers** - Escalate BLOCKER comments immediately
4. **Use convoy list for overview** - Before diving into specific convoy
