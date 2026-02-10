# Troubleshooting

Common issues and quick fixes for AgentOps.

---

## Hooks aren't running

Hooks require configuration in Claude Code's settings file.

**Diagnosis:**

```bash
ao doctor
```

Look for the "Hooks installed" check. If it shows `✗`, hooks are not configured.

**Fixes:**

1. Verify hooks are configured in `~/.claude/settings.json`:
   ```json
   {
     "hooks": {
       "PostToolUse": [...],
       "UserPromptSubmit": [...]
     }
   }
   ```
   The `ao doctor` check counts all hooks across event types. If it reports "no hooks configured", hooks are missing from settings.json entirely.

2. Check that hooks are not disabled via environment variable:
   ```bash
   echo $AGENTOPS_HOOKS_DISABLED
   ```
   If set to `1`, all hooks are bypassed. Unset it:
   ```bash
   unset AGENTOPS_HOOKS_DISABLED
   ```

3. Verify hook scripts exist and are executable:
   ```bash
   ls -la hooks/
   ```
   All `.sh` files in the hooks directory should have execute permissions.

---

## Skills not showing up

Skills must be installed as a Claude Code plugin.

**Diagnosis:**

```bash
npx skills@latest list
ao doctor
```

The `ao doctor` "Plugin" check scans the `skills/` directory for subdirectories containing a `SKILL.md` file. If it reports "no skills found" or "skills directory not found", the plugin is not installed correctly.

**Fixes:**

1. Install or reinstall the AgentOps skills:
   ```bash
   npx skills@latest add boshu2/agentops --all -g
   ```

2. Update existing skills:
   ```bash
   npx skills@latest update
   ```

3. If updates seem stale, clear the cache and reinstall:
   ```bash
   # The skills cache lives here:
   ls ~/.claude/plugins/marketplaces/agentops-marketplace/
   # Pull latest directly if npx update lags:
   cd ~/.claude/plugins/marketplaces/agentops-marketplace/ && git pull
   ```

4. Verify the plugin loads:
   ```bash
   claude --plugin ./
   ```

---

## Push blocked by vibe gate

The push gate hook blocks `git push` unless a recent `/vibe` check has passed. This enforces quality validation before code reaches the remote.

**Why it exists:** The vibe gate prevents untested or unreviewed code from being pushed. It is part of the AgentOps quality enforcement workflow.

**Quick bypass (use sparingly):**

```bash
AGENTOPS_HOOKS_DISABLED=1 git push
```

**Proper resolution:**

1. Run `/vibe` on your changes:
   ```
   /vibe
   ```

2. Address any findings until you get a PASS verdict.

3. Push normally:
   ```bash
   git push
   ```

---

## Worker tried to commit

This is expected behavior in the **lead-only commit** pattern used by `/crank` and `/swarm`.

**How it works:**

- Workers write files but NEVER run `git add`, `git commit`, or `git push`.
- The team lead validates all worker output, then commits once per wave.
- This prevents merge conflicts when multiple workers run in parallel.

**If a worker accidentally committed:**

1. The lead should review the commit before pushing.
2. Amend or squash if needed to maintain clean history.

**For workers:** If you are a worker agent, your only job is to write files. The lead handles all git operations.

---

## Phantom command error

If you see errors for commands like `bd mol`, `gt convoy`, or `bd cook`, these are **planned future features** that do not exist yet.

**How to identify:** Look for `FUTURE` markers in skill documentation. These indicate commands or features that are designed but not yet implemented.

**What to do:**

- Do not retry the command. It will not work.
- Check the skill's `SKILL.md` for current supported commands.
- Use `bd --help` or `gt --help` to see available subcommands.

---

## ao doctor shows failures

`ao doctor` runs 7 health checks. Here is how to fix each one.

### Required checks (failures make the result UNHEALTHY)

| Check | What it verifies | How to fix |
|-------|-----------------|------------|
| **ao CLI** | The `ao` binary is running and reports its version. | Reinstall: `brew install boshu2/tap/ao` or build from `cli/`. |
| **Hooks installed** | `~/.claude/settings.json` contains a `hooks` key with at least one hook configured. | See [Hooks aren't running](#hooks-arent-running) above. |
| **Knowledge base** | The `.agents/ao/` directory exists in the current working directory. | Run `ao init` from your project root, or verify you are in the correct directory. |
| **Plugin** | The `skills/` directory exists and contains at least one subdirectory with a `SKILL.md` file. | See [Skills not showing up](#skills-not-showing-up) above. |

### Optional checks (warnings, result stays HEALTHY)

| Check | What it verifies | How to fix |
|-------|-----------------|------------|
| **Codex CLI** | The `codex` binary is on your PATH. Needed for `--mixed` council mode. | Install Codex CLI and ensure it is in PATH. Requires an API account (not ChatGPT). |
| **Bd CLI** | The `bd` binary is on your PATH. Needed for issue tracking. | Install: `brew install boshu2/tap/bd` |
| **Knowledge pool** | At least one file exists in `.agents/learnings/` or `.agents/patterns/`. | Run `/retro` or `/forge` to extract learnings from your sessions. An empty pool is normal for new projects. |

### Reading the output

```
AgentOps Health Check
=====================
✓ ao CLI: v2.0.1
✓ Hooks installed: 12 hooks configured
✓ Knowledge base: .agents/ao initialized
✓ Plugin: 33 skills found
⚠ Codex CLI: not found (optional — needed for --mixed council)
✓ Bd CLI: available
⚠ Knowledge pool: empty (no learnings or patterns yet)

Result: HEALTHY (2 optional warnings)
```

- `✓` = pass
- `⚠` = warning (optional component missing or degraded)
- `✗` = failure (required component missing or broken)

Use `ao doctor --json` for machine-readable output.

---

## Getting help

- **New to AgentOps?** Run `/quickstart` for an interactive onboarding walkthrough.
- **Run diagnostics:** `ao doctor` checks your installation health.
- **Report issues:** [github.com/boshu2/agentops/issues](https://github.com/boshu2/agentops/issues)
- **Full workflow guide:** Run `/using-agentops` for the complete RPI workflow reference.
