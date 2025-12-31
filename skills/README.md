# Skills Catalog

Domain-specific knowledge modules that activate based on trigger patterns.

## Quick Reference

| Skill | Triggers | Purpose |
|-------|----------|---------|
| [languages](languages/) | Python, Go, Rust, Java, TypeScript, shell | Language-specific patterns |
| [development](development/) | API, backend, frontend, mobile, deploy, LLM | Software development |
| [documentation](documentation/) | docs, README, OpenAPI, Diátaxis | Documentation creation |
| [code-quality](code-quality/) | review, test, coverage | Code review and testing |
| [research](research/) | explore, find, analyze, git history | Codebase exploration |
| [validation](validation/) | validate, verify, assumption | Testing and verification |
| [operations](operations/) | incident, outage, debug, logs | Incident response |
| [monitoring](monitoring/) | metrics, alerts, SLO, performance | Observability |
| [security](security/) | pentest, SSL, firewall, secrets | Security and network |
| [data](data/) | pipeline, ETL, ML, MLOps | Data and ML engineering |
| [meta](meta/) | context, session, memory, retro | Workflow coordination |
| [specialized](specialized/) | accessibility, UX, Obsidian, risk | Domain specialists |

## Skill Structure

Each skill directory contains:

```
skills/<domain>/
└── SKILL.md          # Main skill definition with triggers and knowledge
```

## SKILL.md Format

```yaml
---
name: <skill-name>
description: >
  Use when: "trigger1", "trigger2", "trigger3"...
version: 1.0.0
---

# Skill Title

[Quick reference tables]
[Detailed knowledge sections]
[Code examples and patterns]
```

## How Skills Work

1. **Trigger** - Keywords in description activate the skill
2. **Load** - Skill content loaded into main context
3. **Apply** - Knowledge used for current task
4. **Execute** - All tools available (no sub-agent limitations)

## Skills vs Agents

| Aspect | Skills | Agents |
|--------|--------|--------|
| Execution | Main context | Sub-process |
| Tools | Full access | Limited |
| Context | Preserved | Isolated |
| Chaining | Yes | No |
| MCP | Available | Unavailable |

Skills are the preferred pattern for domain knowledge.

## Adding New Skills

1. Create directory: `skills/<domain>/`
2. Create `SKILL.md` with frontmatter
3. Add trigger patterns in description
4. Include reference tables and examples
5. Update this README

## Skill Domains

### Languages (6 areas)
Python, Go, Rust, Java, TypeScript, Shell - language-specific idioms and patterns.

### Development (8 areas)
Backend, Frontend, Fullstack, Mobile, iOS, Deployment, AI, Prompts - software development.

### Documentation (4 areas)
Create, Optimize, Audit, API - documentation patterns using Diátaxis framework.

### Code Quality (3 areas)
Review, Improve, Test - code review and test generation.

### Research (6 areas)
Code, Docs, History, Archive, Structure, Specs - exploration and analysis.

### Validation (4 areas)
Assumptions, Continuous, Planning, Tracer Bullets - verification patterns.

### Operations (4 areas)
Incident Response, Triage, Postmortems, Error Detection - production support.

### Monitoring (2 areas)
Alerts/Runbooks, Performance - observability and optimization.

### Security (2 areas)
Penetration Testing, Network Engineering - security and infrastructure.

### Data (4 areas)
Engineering, Science, ML, MLOps - data and machine learning pipelines.

### Meta (6 areas)
Context, Execution, Autonomous, Observer, Memory, Retros - workflow coordination.

### Specialized (6 areas)
Accessibility, Support, UI/UX, Knowledge Graphs, Decomposition, Risk - domain specialists.

---

**Total**: 12 domains, 55 knowledge areas consolidated into 12 skills
