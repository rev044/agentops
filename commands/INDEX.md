# Commands Index

> **⚠️ DEPRECATED**: Commands are being replaced by **skills**. Skills are directly invokable with `/skill-name` and provide better context loading. See [plugins/vibe-kit/skills/](../plugins/vibe-kit/skills/) for the modern approach.
>
> **Migration guide:**
> - `/research` → Use skill: `/research`
> - `/plan` → Use skill: `/formulate` (or Claude's built-in plan mode)
> - `/implement` → Use skill: `/implement`
> - `/retro` → Use skill: `/retro`
>
> Commands here are maintained for legacy compatibility only.

## By Category

### RPI Workflow (Core)
| Command | Description | Skill Replacement |
|---------|-------------|-------------------|
| [research](research.md) | Phase 1 - Deep exploration before planning | `research` |
| ~~plan~~ | **REMOVED** - Use Claude's native plan mode or `/formulate` | Native plan mode |
| [implement](implement.md) | Phase 3 - Execute approved plan with validation | `implement` |

### Bundles (Context Persistence)

> **Note:** Bundle commands are legacy. Consider using `.agents/` directory structure instead.

| Command | Description |
|---------|-------------|
| [bundle-save](bundle-save.md) | Save compressed context to filesystem |
| [bundle-load](bundle-load.md) | Restore context from saved bundle |
| [bundle-search](bundle-search.md) | Semantic search across bundles |
| [bundle-list](bundle-list.md) | List all available bundles |
| [bundle-prune](bundle-prune.md) | Garbage collection with knowledge extraction |
| [bundle-load-multi](bundle-load-multi.md) | Load multiple bundles at once |

### Session Management
| Command | Description |
|---------|-------------|
| [session-start](session-start.md) | Initialize session with progress check |
| [session-end](session-end.md) | Save state before ending |
| [session-resume](session-resume.md) | Single-command resume with auto-detection |

### Metrics & Calibration
| Command | Description | Skill Replacement |
|---------|-------------|-------------------|
| [vibe-check](vibe-check.md) | Analyze git history for session metrics | `vibe` |
| [vibe-level](vibe-level.md) | Classify task trust level (0-5) | `vibe` |

### Learning & Improvement
| Command | Description | Skill Replacement |
|---------|-------------|-------------------|
| [learn](learn.md) | Extract patterns from completed work | `retro` |
| [retro](retro.md) | Post-work retrospective with analysis | `retro` |

### Project Management
| Command | Description |
|---------|-------------|
| [project-init](project-init.md) | Initialize 2-Agent Harness for multi-day projects |
| [progress-update](progress-update.md) | Update progress files interactively |

### Quality & Review
| Command | Description | Skill Replacement |
|---------|-------------|-------------------|
| [code-review](code-review.md) | Comprehensive code quality review | `vibe` |
| [architecture-review](architecture-review.md) | Architecture analysis and recommendations | `vibe` |
| [generate-tests](generate-tests.md) | Generate comprehensive test suites | - |

### Documentation
| Command | Description | Skill Replacement |
|---------|-------------|-------------------|
| [update-docs](update-docs.md) | Systematically update project documentation | `doc` |
| [create-architecture-documentation](create-architecture-documentation.md) | Generate architecture docs with diagrams | `doc` |
| [create-onboarding-guide](create-onboarding-guide.md) | Create developer onboarding guides | `doc` |

### Utilities
| Command | Description |
|---------|-------------|
| [ultra-think](ultra-think.md) | Deep analysis with multi-dimensional thinking |
| [maintain](maintain.md) | Weekly knowledge maintenance |
| [containerize-application](containerize-application.md) | Containerize with optimized Docker config |

### Multi-Agent
| Command | Description | Skill Replacement |
|---------|-------------|-------------------|
| [research-multi](research-multi.md) | Launch 3 parallel agents for 3x faster research | `implement-wave` |

## Installation (Legacy)

```bash
# Prefer installing plugins instead:
/plugin install vibe-kit@boshu2-agentops

# Legacy command installation (deprecated):
cp *.md /path/to/project/.claude/commands/
```
