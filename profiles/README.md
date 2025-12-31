# Role-Based Profiles

**Purpose**: Organize 55 agents into 3 discoverable profiles for different work contexts.

**Version**: 3.0.0 (Updated for AgentOps marketplace agents)

---

## Quick Start

| You're doing... | Load this profile | Key Agents |
|-----------------|-------------------|------------|
| Building apps (APIs, frontends, features) | **software-dev** | backend-architect, python-pro, code-reviewer |
| Operations (incidents, monitoring, deploys) | **platform-ops** | incident-responder, performance-engineer, error-detective |
| Writing (docs, research, patterns) | **content-creation** | documentation-create-docs, meta-retro-analyzer, doc-explorer |

---

## The 3 Profiles

### 1. Software Development (`software-dev`)

**What you do**:
- Build applications (backend APIs, frontends, full-stack)
- Write code in Python, Go, Rust, TypeScript, Java
- Review code and generate tests
- Deploy with CI/CD pipelines

**Agents included** (26 agents):

| Domain | Agents |
|--------|--------|
| **Languages** | python-pro, golang-pro, rust-pro, java-pro, typescript-pro, shell-scripting-pro |
| **Development** | backend-architect, frontend-developer, fullstack-developer, mobile-developer, ios-developer, deployment-engineer, ai-engineer, prompt-engineer |
| **Code Quality** | code-reviewer, code-review-improve, test-generator |
| **Validation** | assumption-validator, continuous-validator, validation-planner, tracer-bullet-deployer |
| **Data** | data-engineer, data-scientist, ml-engineer, mlops-engineer |

**Example workflow**:
```
1. /research "API design for user auth"
2. Load backend-architect for architecture
3. Load python-pro for implementation
4. Load code-reviewer before commit
5. Load test-generator for coverage
```

---

### 2. Platform Operations (`platform-ops`)

**What you do**:
- Respond to production incidents
- Monitor system health and performance
- Debug errors and analyze logs
- Manage security and network

**Agents included** (12 agents):

| Domain | Agents |
|--------|--------|
| **Operations** | incident-responder, incidents-response, incidents-postmortems, error-detective |
| **Monitoring** | performance-engineer, monitoring-alerts-runbooks |
| **Security** | penetration-tester, network-engineer |
| **Validation** | assumption-validator, tracer-bullet-deployer |
| **Meta** | change-executor, autonomous-worker |

**Example workflow**:
```
1. Alert fires → Load incident-responder
2. Analyze logs → Load error-detective
3. Fix applied → Load incidents-postmortems
4. Improve monitoring → Load monitoring-alerts-runbooks
```

---

### 3. Content Creation (`content-creation`)

**What you do**:
- Write documentation and tutorials
- Extract patterns from completed work
- Conduct research and analysis
- Synthesize learnings

**Agents included** (17 agents):

| Domain | Agents |
|--------|--------|
| **Documentation** | documentation-create-docs, documentation-optimize-docs, documentation-diataxis-auditor, api-documenter |
| **Research** | code-explorer, doc-explorer, history-explorer, archive-researcher, document-structure-analyzer, spec-architect |
| **Meta** | context-manager, meta-observer, meta-memory-manager, meta-retro-analyzer |
| **Specialized** | accessibility-specialist, customer-support, ui-ux-designer |

**Example workflow**:
```
1. /research "existing auth patterns"
2. Load code-explorer for codebase analysis
3. Load documentation-create-docs for writing
4. Load meta-retro-analyzer for insights
```

---

## Profile Structure

```
profiles/
├── README.md                 ← This file
├── COMPARISON.md             ← Profile comparison
├── META_PATTERNS.md          ← Extracted patterns
├── schema/
│   └── role-profile.yaml     ← Profile schema
├── shared/
│   ├── foundational.yaml     ← Core standards
│   ├── orchestration.yaml    ← Workflow commands
│   └── context.yaml          ← Bundle/memory
├── roles/
│   ├── software-dev.yaml     ← Development (26 agents)
│   ├── platform-ops.yaml     ← Operations (12 agents)
│   └── content-creation.yaml ← Writing (17 agents)
└── examples/
    ├── software-dev-session.md
    ├── platform-ops-session.md
    └── content-creation-session.md
```

---

## Agent Domain Reference

All 55 agents organized by domain:

| Domain | Count | Agents |
|--------|-------|--------|
| **languages** | 6 | python-pro, golang-pro, rust-pro, java-pro, typescript-pro, shell-scripting-pro |
| **development** | 8 | backend-architect, frontend-developer, fullstack-developer, mobile-developer, ios-developer, deployment-engineer, ai-engineer, prompt-engineer |
| **documentation** | 4 | documentation-create-docs, documentation-optimize-docs, documentation-diataxis-auditor, api-documenter |
| **code_quality** | 3 | code-reviewer, code-review-improve, test-generator |
| **research** | 6 | code-explorer, doc-explorer, history-explorer, archive-researcher, document-structure-analyzer, spec-architect |
| **validation** | 4 | assumption-validator, continuous-validator, validation-planner, tracer-bullet-deployer |
| **operations** | 4 | incident-responder, incidents-response, incidents-postmortems, error-detective |
| **monitoring** | 2 | performance-engineer, monitoring-alerts-runbooks |
| **security** | 2 | penetration-tester, network-engineer |
| **data** | 4 | data-engineer, data-scientist, ml-engineer, mlops-engineer |
| **meta** | 6 | context-manager, change-executor, autonomous-worker, meta-observer, meta-memory-manager, meta-retro-analyzer |
| **specialized** | 6 | accessibility-specialist, customer-support, connection-agent, ui-ux-designer, task-decomposition-expert, risk-assessor |

---

## Profile Selection Guide

### By Task Keywords

| Keywords | Profile |
|----------|---------|
| "build", "implement", "feature", "api", "code" | software-dev |
| "incident", "outage", "monitor", "debug", "logs" | platform-ops |
| "document", "write", "research", "pattern", "tutorial" | content-creation |

### By Agent Need

| You need... | Profile |
|-------------|---------|
| Language experts (python-pro, rust-pro) | software-dev |
| Code review and testing | software-dev |
| Incident response | platform-ops |
| Performance analysis | platform-ops |
| Documentation writing | content-creation |
| Pattern extraction | content-creation |

---

## Integration with Commands

Profiles work with the core RPI workflow:

```
/research → Explore before planning (any profile)
/plan     → Create implementation spec (any profile)
/implement → Execute with validation (software-dev)
/retro    → Extract learnings (content-creation)
```

**Session management**:
```
/session-start → Initialize with profile context
/session-end   → Save progress
/bundle-save   → Persist for multi-session work
```

---

## Version History

**v3.0.0 (2025-12-30)**:
- Updated for AgentOps marketplace (55 agents, 12 domains)
- Removed references to external workspace agents
- Simplified to 3 core profiles
- Added agent domain reference table

**v2.0.0 (2025-11-09)**:
- Consolidated from 5 roles → 3 profiles

**v1.0.0 (2025-11-09)**:
- Initial taxonomy creation
