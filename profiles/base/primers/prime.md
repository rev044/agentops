---
description: Interactive JIT context router - guides you to the right context for any task
---

# /prime - Interactive Context Router

**Purpose:** Guide you to the right context for your task using interactive JIT loading.

**Philosophy:** Load constitutional baseline, understand your needs, JIT load relevant patterns.

**Token budget:** 3-5k total (1.5-2.5% of context window)

---

## Step 1: Constitutional Baseline (Always Enforced)

{{cat .claude/CONSTITUTION.md}}

**Status:** ✅ Constitutional foundation loaded (2k tokens)

You are now operating under AgentOps constitution:
- **Laws 1-5:** Enforced by git hooks
- **Three Rules:** kubic-cm read-only, config.env source of truth, semantic commits
- **40% Rule:** Stay under 80k tokens per phase
- **Validation gates:** make quick, make ci-all, human review, CI/CD

---

## Step 2: Understand Your Task

**What are you working on?**

Please describe your task, or choose from common categories:

### Creating Something New
- **New application** - Deploy a new service/app to the platform
- **New documentation** - Write guides, tutorials, or reference docs
- **New site config** - Add configuration for a new site
- **New infrastructure** - Add services, policies, databases

### Modifying Something Existing
- **Modify application** - Change existing app configuration
- **Modify site config** - Update config.env values
- **Modify documentation** - Update or improve existing docs
- **Refactor code** - Restructure without changing behavior

### Debugging/Troubleshooting
- **ArgoCD sync failure** - App won't sync to cluster
- **YAML syntax error** - Validation failing
- **Performance issue** - Slow response, high resource usage
- **Other debug** - General troubleshooting

### Operating/Deploying
- **Deploy changes** - Sync to cluster
- **Harmonize config** - Render config.env to values.yaml
- **Run validation** - Check before committing
- **Other operations** - Routine maintenance

### Other
- Describe your task in your own words

---

## Step 3: JIT Load Relevant Context

**[After user responds, I will:]**

1. Analyze your task type
2. Load ONLY relevant pattern from `docs/reference/workflows/`
3. Suggest 5-6 workflows from `docs/reference/workflows/COMMON_WORKFLOWS.md`
4. Guide you to the right agent or workflow

**Examples of what gets loaded:**

→ User: "Create new application"
  Load: `docs/reference/workflows/application-creation.md` (1k)
  Suggest: applications-create-app, applications-create-app-jren
  Total: 3k tokens (1.5%)

→ User: "Modify site config"
  Load: `docs/reference/workflows/config-env-pattern.md` (0.5k)
  Suggest: sites-site-config, harmonize-guide
  Total: 2.5k tokens (1.25%)

→ User: "ArgoCD sync failing"
  Load: `docs/reference/workflows/argocd-troubleshooting.md` (0.8k)
  Suggest: applications-debug-sync, argocd-debug-sync
  Total: 2.8k tokens (1.4%)

→ User: "Write documentation"
  Load: `docs/reference/workflows/diataxis-format.md` (0.5k)
  Suggest: documentation-create-docs, documentation-optimize-docs
  Total: 2.5k tokens (1.25%)

---

## Step 4: Continue Loading as Needed

As we work together, I can load additional context if needed:

- Specific agent definitions (`.claude/agents/[agent-name].md`)
- Related patterns (`docs/reference/workflows/`)
- Troubleshooting guides (`docs/how-to/troubleshooting/`)
- Configuration reference (`docs/reference/configuration-schema.md`)

**Always staying under 40% token budget (80k of 200k)**

---

## Token Budget Tracking

```text
Context Window: 200,000 tokens
Target: <40% (80k) for entire session

Current allocation:
  Constitution:    ████░░░░░░  2k/200k   (1%)
  Task pattern:    ██░░░░░░░░  1k/200k   (0.5%)
  Workflow guide:  ██░░░░░░░░  1k/200k   (0.5%)
  Reserved:        ░░░░░░░░░░  196k/200k (98%)

Status: 🟢 GREEN - Plenty of headroom for execution
```

---

## What Happens Next?

**I will:**
1. ✅ Understand your task (from your description above)
2. ✅ JIT load the relevant pattern (300-1k tokens)
3. ✅ Suggest 5-6 workflows that match your task
4. ✅ Guide you to the right agent or workflow
5. ✅ Continue to load context as needed during execution

**You do:**
1. Tell me what you're working on (see Step 2 above)
2. Review the suggested workflows
3. Pick the one that fits best (or ask for recommendations)
4. Execute the workflow with my guidance

---

## Why This Approach Works

**Unix Philosophy:**
- Do one thing well (each pattern is focused)
- Composable (load multiple patterns if needed)
- Text streams (patterns are text files)

**Context Engineering:**
- JIT loading (not upfront loading)
- Stay under 40% rule (efficient token usage)
- Progressive disclosure (load more when needed)

**Learning Science:**
- Interactive (ask questions, understand needs)
- Guided (suggest workflows, don't prescribe)
- Contextual (load what's relevant now)

---

**Ready! What are you working on?** (Describe your task or choose from categories above)
