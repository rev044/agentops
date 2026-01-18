# Gas Town Setup Guide

## Overview

This guide walks through setting up Gas Town for use with sk-gastown.

## Prerequisites Check

```bash
# Check if Gas Town installed
ls ~/gt/mayor/town.json 2>/dev/null && echo "✓ Gas Town installed" || echo "✗ Gas Town not installed"

# Check if daemon running
pushd ~/gt && ./gt daemon status && popd 2>/dev/null && echo "✓ Daemon running" || echo "✗ Daemon not running"

# Check if rig configured
pushd ~/gt && ./gt rig list && popd 2>/dev/null | grep -q "gastown" && echo "✓ Rig configured" || echo "✗ No rig"
```

## Installation

**Full guide:** `~/gt/daedalus/mayor/rig/INSTALLING.md`

### Quick Install

```bash
# Clone Gas Town
git clone https://github.com/boshu2/gastown.git ~/gt

# Initialize
cd ~/gt
./gt init

# Start daemon
./gt daemon start
```

### Add a Rig

```bash
# Add rig pointing to your project repo
cd ~/gt
./gt rig add myproject https://github.com/user/repo.git

# Verify
./gt rig list
```

## Verification

```bash
# Full health check
cd ~/gt
./gt doctor

# Should show:
# ✓ Town initialized
# ✓ Daemon running
# ✓ Rigs configured
# ✓ Hooks installed
```

## Troubleshooting

### Gas Town Not Installed

```bash
# Check if ~/gt exists
ls ~/gt

# If not, clone it
git clone https://github.com/boshu2/gastown.git ~/gt
cd ~/gt && ./gt init
```

### Daemon Not Running

```bash
# Check status
cd ~/gt && ./gt daemon status

# Start daemon
cd ~/gt && ./gt daemon start

# If fails, check logs
cat ~/gt/deacon/logs/daemon.log
```

### No Rig Configured

```bash
# List rigs
cd ~/gt && ./gt rig list

# Add rig
cd ~/gt && ./gt rig add <name> <git-url>
```

### Polecat Won't Spawn

```bash
# Check tmux
tmux list-sessions

# Check daemon
cd ~/gt && ./gt daemon status

# Force restart
cd ~/gt && ./gt daemon stop && ./gt daemon start
```

## Deep Dive Documentation

For detailed architecture and troubleshooting:

| Topic | Document |
|-------|----------|
| **Installation** | `~/gt/daedalus/mayor/rig/INSTALLING.md` |
| **Architecture** | `~/gt/daedalus/mayor/rig/architecture.md` |
| **Understanding roles** | `~/gt/daedalus/mayor/rig/understanding-gas-town.md` |
| **Why these features** | `~/gt/daedalus/mayor/rig/why-these-features.md` |
| **Operational state** | `~/gt/daedalus/mayor/rig/operational-state.md` |

## Next Steps

Once setup is complete:

1. Test with: `/gastown status`
2. Run a simple task: `/gastown "Add hello world endpoint"`
3. Monitor: `cd ~/gt && ./gt convoy list`
