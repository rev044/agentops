# Installing AgentOps for Codex

AgentOps Codex skills install directly into Codex's native skills directory.

## Installation

Run:

```bash
curl -fsSL https://raw.githubusercontent.com/boshu2/agentops/main/scripts/install-codex.sh | bash
```

This installs skills to:

```bash
~/.codex/skills
```

## Verification

You should see AgentOps skills as normal native Codex skills in your next session.

## Update policy

AgentOps updates frequently. Codex does not currently provide a universal auto-update channel for this style of skill install.

Re-run the installer regularly, especially after new releases:

```bash
curl -fsSL https://raw.githubusercontent.com/boshu2/agentops/main/scripts/install-codex.sh | bash
```
