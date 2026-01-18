# Gas Town Kit

Gas Town multi-agent orchestration. 4 skills for crew, polecats, and routing.

## Install

```bash
/plugin install gastown-kit@boshu2-agentops
```

## Skills

| Skill | Invoke | Purpose |
|-------|--------|---------|
| `/crew` | auto-triggered | Crew workspace management |
| `/polecat-lifecycle` | auto-triggered | Polecat spawn/nuke/gc |
| `/gastown` | `/gastown` | Gas Town status and utilities |
| `/bd-routing` | auto-triggered | Beads prefix routing |

## Gas Town Architecture

```
Town (~/gt)
├── mayor/          ← Global coordinator
├── <rig>/          ← Your rig (e.g., daedalus)
│   ├── .beads/     ← Issue tracking
│   ├── crew/       ← Human-guided workspaces
│   ├── polecats/   ← Transient workers
│   ├── refinery/   ← Merge queue
│   └── witness/    ← Lifecycle monitor
```

## Key Commands

### Crew management

```bash
gt crew list <rig>           # List crew members
gt crew add <name> --rig <rig>  # Add crew member
```

### Polecat lifecycle

```bash
gt polecat list <rig>        # List polecats
gt polecat status <rig>/<name>  # Detailed status
gt polecat nuke <rig>/<name> --force  # Destroy
gt polecat gc <rig>          # Clean merged branches
gt polecat reset <rig>/<name>  # Reset to idle
```

### Beads routing

```bash
bd show gt-1234              # Routes to daedalus
bd show hq-5678              # Routes to town beads
BD_DEBUG_ROUTING=1 bd show <id>  # Debug routing
```

## Philosophy

- **Gas Town is a steam engine** - agents are pistons
- **Crew for human-guided work** - persistent, named workspaces
- **Polecats for autonomous parallel execution** - transient, disposable
- **Beads routing enables multi-rig coordination** - prefixes route work

## Related Kits

- **dispatch-kit** - Work assignment primitives
- **beads-kit** - Issue tracking
- **core-kit** - What polecats execute
