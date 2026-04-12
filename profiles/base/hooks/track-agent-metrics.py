#!/usr/bin/env python3
"""
Track agent execution metrics for showcase dashboard.
Runs on Stop hook - logs success/failure, tokens, duration.
"""
import json
import sys
import datetime
from pathlib import Path

try:
    # Read hook data from stdin
    data = json.load(sys.stdin)

    # Extract metrics
    metrics = {
        "timestamp": datetime.datetime.now().isoformat(),
        "session_id": data.get("session_id", "unknown"),
        "total_tokens": data.get("usage", {}).get("total_tokens", 0),
        "input_tokens": data.get("usage", {}).get("input_tokens", 0),
        "output_tokens": data.get("usage", {}).get("output_tokens", 0),
        "duration_ms": data.get("duration_ms", 0),
        "status": "success"  # If we got here, it completed
    }

    # Log to metrics file (JSONL format)
    metrics_file = Path.home() / ".claude/agent-metrics.jsonl"
    metrics_file.parent.mkdir(exist_ok=True)

    with open(metrics_file, "a") as f:
        f.write(json.dumps(metrics) + "\n")

    # Print summary (visible to user)
    print(f"📊 Logged: {metrics['total_tokens']:,} tokens, {metrics['duration_ms']/1000:.1f}s")

except Exception as e:
    # Don't fail the hook - just log error
    print(f"⚠️  Metrics tracking error: {e}", file=sys.stderr)
    sys.exit(0)  # Exit 0 so we don't block Claude
