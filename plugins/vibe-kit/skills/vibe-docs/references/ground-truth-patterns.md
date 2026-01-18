# Ground Truth Patterns

Patterns for establishing and maintaining authoritative documentation sources.

## Core Principle

**One Source of Truth Per Domain.**

When information appears in multiple places, ONE must be canonical and others MUST reference it.

---

## Ground Truth Registry

| Domain | Ground Truth File | Update Frequency | Validator |
|--------|-------------------|------------------|-----------|
| Agents | `docs/agents/catalog.md` | On deployment | `oc get agents` |
| Container Images | `charts/*/IMAGE-LIST.md` | On build | Registry query |
| API Endpoints | OpenAPI spec or code | On code change | Automated |
| Config Options | `values.yaml` | On chart change | Helm lint |
| Dependencies | `pyproject.toml` / `package.json` | On update | Lock file |

---

## Reference Pattern

### Wrong: Duplicate Data

```markdown
<!-- faq.md -->
The following agents are deployed:
- Knowledge Assistant
- MR Reviewer
- Slack Assistant

<!-- prd.md -->
Deployed agents include:
- Knowledge Assistant (chat)
- MR Reviewer (webhook)
- Slack Assistant (mention)

<!-- code-map/README.md -->
| Agent | Status |
|-------|--------|
| Knowledge Assistant | Deployed |
| MR Reviewer | Deployed |
```

**Problem:** Three files, three potential points of drift.

### Right: Reference Ground Truth

```markdown
<!-- faq.md -->
See [Agent Catalog](../agents/catalog.md) for deployed agents.

<!-- prd.md -->
Agent deployment status is tracked in [Agent Catalog](docs/agents/catalog.md).

<!-- code-map/README.md -->
> **Runtime Status:** See [Agent Catalog](../agents/catalog.md) for actual deployment status.
```

**Benefit:** One source to update, all docs stay current.

---

## Validation Commands

### Agents

```bash
# Get ground truth from cluster
oc get agents.kagent.dev -n ai-platform -o json | \
  jq -r '.items[] | "\(.metadata.name): \(.status.conditions[0].type)=\(.status.conditions[0].status)"'

# Compare with catalog
cat docs/agents/catalog.md | grep -E "^\| \*\*" | awk -F'|' '{print $2}'
```

### Images

```bash
# Get deployed images
oc get pods -n ai-platform -o json | \
  jq -r '.items[].spec.containers[].image' | sort -u

# Compare with IMAGE-LIST
grep -E "^\|.*\|.*\|" charts/ai-platform/IMAGE-LIST.md | awk -F'|' '{print $3}'
```

### Configs

```bash
# Get deployed configmap values
oc get configmap ai-platform-config -n ai-platform -o yaml

# Compare with values.yaml defaults
helm template ai-platform charts/ai-platform/ | grep -A10 "kind: ConfigMap"
```

---

## Update Protocol

When ground truth changes:

1. **Update ground truth file first**
2. **Run validation** - `grep -r "Status:" docs/ | grep -v "catalog.md"`
3. **Fix references** - Update any docs that duplicate rather than reference
4. **Update validation dates** - Add/update "Validated: DATE against SOURCE"

---

## Staleness Detection

```bash
# Find docs with old validation dates
grep -r "Validated:" docs/ | while read line; do
  date=$(echo "$line" | grep -oE "[0-9]{4}-[0-9]{2}-[0-9]{2}")
  if [[ -n "$date" ]]; then
    age=$((( $(date +%s) - $(date -d "$date" +%s) ) / 86400))
    if [[ $age -gt 30 ]]; then
      echo "STALE ($age days): $line"
    fi
  fi
done
```

---

## Anti-Patterns

| Pattern | Problem | Fix |
|---------|---------|-----|
| Inline agent lists | Drift independently | Reference catalog |
| Hardcoded versions | Outdated quickly | Reference Chart.yaml |
| Copy-paste configs | Diverge over time | Reference values.yaml |
| "As of DATE" claims | Become stale | Use validation dates with source |
