---
name: base
description: Utility skills for cleanup and auditing - deprecated code removal, doc cleanup, plan cleanup, repo cleanup, and diataxis/onboarding/workflow audits
version: 1.0.0
---

# Base Utilities

Collection of cleanup and audit skills. See subdirectories for individual skills.

## Available Skills

| Skill | Purpose |
|-------|---------|
| `cleanup-deprecated` | Remove deprecated code patterns |
| `cleanup-docs` | Clean up documentation files |
| `cleanup-plans` | Archive or remove stale plans |
| `cleanup-repo` | General repository cleanup |
| `audit-diataxis` | Audit docs against Diataxis framework |
| `audit-onboarding` | Audit onboarding documentation |
| `audit-workflow` | Audit workflow documentation |

## Usage

These are utility skills, typically invoked for maintenance tasks:

```bash
/cleanup-deprecated   # Find and remove deprecated patterns
/cleanup-docs         # Clean up documentation
/audit-diataxis       # Check docs follow Diataxis
```
