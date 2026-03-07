---
name: update
description: 'Reinstall all AgentOps skills globally from the latest source. Triggers: "update skills", "reinstall skills", "sync skills".'
skill_api_version: 1
user-invocable: true
context:
  window: isolated
  intent:
    mode: none
  sections:
    exclude: [HISTORY, INTEL, TASK]
  intel_scope: none
metadata:
  tier: meta
  dependencies: []
---

# /update — Reinstall AgentOps Skills

> **Purpose:** One command to pull the latest skills from the repo and install them globally across all agents.

**YOU MUST EXECUTE THIS WORKFLOW. Do not just describe it.**

---

## Execution

### Step 1: Install

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/boshu2/agentops/main/scripts/install.sh)
```

Run this command. Wait for it to complete.

### Step 2: Verify

Confirm the output shows all skills installed with no failures.

If any skills failed to install, report which ones failed and suggest re-running or manual sync:
```bash
# Manual sync for a failed skill (replace <skill-name>):
/bin/cp -r ~/.agents/skills/<skill-name>/ ~/.claude/skills/<skill-name>/
```

### Step 3: Report

Tell the user:
1. How many skills installed successfully
2. Any failures and how to fix them

## Examples

### Routine skill update

**User says:** `/update`

**What happens:**
1. Runs the install script to pull the latest skills from the repository and install them globally.
2. Verifies the output confirms all skills installed with no failures.
3. Reports the total count of successfully installed skills.

**Result:** All AgentOps skills are updated to the latest version and available globally across all agent sessions.

### Recovering from a partial failure

**User says:** `/update` (after a previous run failed for some skills)

**What happens:**
1. Re-runs the install script which re-downloads and overwrites all skills from the latest source.
2. Detects that 2 of 50 skills failed to install and identifies them by name.
3. Reports the failures and provides manual sync commands as a fallback.

**Result:** 48 skills installed successfully, with clear instructions to manually sync the 2 that failed.

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| `curl: command not found` | curl is not installed | Install curl via your package manager |
| Download fails | Network or GitHub unreachable | Check connectivity; retry |
| Individual skills fail | Permissions issue in `~/.claude/skills/` | `chmod -R u+rwX ~/.claude/skills/` then re-run `/update` |
| Skills not available after install | Agent session not restarted | Restart your agent session |
| `EACCES: permission denied` | Restrictive permissions on skills dir | `chmod -R u+rwX ~/.claude/skills/` and re-run `/update` |
