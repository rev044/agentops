# Domain Kit

Reference knowledge across 21 domains. Loaded on-demand for specialized tasks.

## Install

```bash
/plugin install domain-kit@boshu2-agentops
```

## Domains

| Domain | Areas |
|--------|-------|
| **standards** | Python, Go, TypeScript, Shell, YAML, JSON, Markdown, OpenAI |
| **languages** | Python, Go, Rust, Java, TypeScript, JavaScript |
| **development** | API, backend, frontend, mobile, LLM, microservices |
| **documentation** | docs, README, OpenAPI, Diátaxis |
| **code-quality** | review, test, coverage |
| **validation** | validate, verify, tracer bullet |
| **operations** | incident, debug, postmortem |
| **monitoring** | metrics, alerts, SLO |
| **security** | pentest, SSL, secrets |
| **data** | ETL, Spark, ML, MLOps |
| **meta** | context, session, workflow |
| **specialized** | accessibility, UX, risk |
| **testing** | unit, integration, e2e |
| **git-workflow** | commit conventions, hooks |
| **brand-guidelines** | style, voice, assets |
| **skill-creator** | creating new skills |
| **test-gap-scanner** | finding untested code |
| **tekton** | Tekton pipeline builds, troubleshooting |
| **tekton-go-operator** | Go operator CI setup with Tekton |
| **container-build** | OpenShift container images, arbitrary UIDs |

### Standards Library

The **standards** skill is a library providing language-specific coding standards with AI-optimized structure:

| Reference | Use When |
|-----------|----------|
| `python.md` | `.py` files |
| `go.md` | `.go` files |
| `typescript.md` | `.ts`, `.tsx` files |
| `shell.md` | `.sh`, `.bash` files |
| `yaml.md` | `.yaml`, `.yml` files |
| `json.md` | `.json`, `.jsonl` files |
| `markdown.md` | `.md` files |
| `openai.md` | OpenAI API integration |

Each reference includes:
- **Common Errors** - symptom → cause → fix tables
- **Anti-Patterns** - named patterns to avoid
- **AI Agent Guidelines** - ALWAYS/NEVER rules

Other skills (like `bug-hunt` and `complexity` in vibe-kit) depend on standards for consistent code guidance.

## Usage

Domain skills are auto-triggered based on context. Examples:

```bash
# Python-specific guidance loads when working on .py files
# Security domain loads when reviewing auth code
# Testing domain loads when writing tests
```

## Philosophy

- **Load knowledge when needed** - JIT, not upfront
- **21 domains covering full development lifecycle**
- **Reference-style documentation** - consult, don't memorize

## Related Kits

- **core-kit** - Uses domain knowledge during research
- **vibe-kit** - Validation informed by domain expertise
