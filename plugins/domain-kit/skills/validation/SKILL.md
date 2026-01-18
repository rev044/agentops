---
name: validation
description: >
  Use when: "validate", "verify", "assumption", "test strategy", "tracer bullet",
  "smoke test", "continuous validation", "environment check", "reality check",
  "deployment validation", "pre-flight".
version: 1.0.0
author: "AgentOps Team"
license: "MIT"
---

# Validation Skill

Assumption validation, test strategy design, and deployment verification patterns.

## Quick Reference

| Area | Key Patterns | When to Use |
|------|--------------|-------------|
| **Assumption** | Environment validation | Before implementation |
| **Continuous** | Lifecycle validation | During development |
| **Planning** | Test strategy design | Before testing |
| **Tracer Bullet** | Minimal deployment | Early validation |

---

## Assumption Validation

### Purpose
Validate that research assumptions match reality before implementation.

### Validation Categories

| Category | What to Check |
|----------|---------------|
| **API** | CRDs exist, versions match |
| **Images** | Pullable, signatures valid |
| **Operators** | Installed, status Succeeded |
| **Config** | Parameters supported |
| **Permissions** | RBAC allows operations |

### Validation Commands

```bash
# API exists
kubectl api-resources | grep -i "<resource>"

# CRD versions
kubectl get crd <name> -o jsonpath='{.spec.versions[*].name}'

# Image pullable (Kubernetes)
kubectl run test --image=<image> --restart=Never --dry-run=server -o yaml

# Operator status
kubectl get csv -A | grep -i "<operator>"

# Permission check
kubectl auth can-i create <resource> -n <namespace>

# Admission test (dry run)
kubectl apply --dry-run=server -f <manifest>
```

### Output Template

```markdown
# Assumption Validation Report

## Summary
| Category | Tested | Passed | Failed |
|----------|--------|--------|--------|
| API | 5 | 5 | 0 |
| Images | 3 | 2 | 1 |
| Operators | 2 | 2 | 0 |

## Detailed Results

### Passed ✅
- [Assumption 1]: [Evidence]
- [Assumption 2]: [Evidence]

### Failed ❌
- [Assumption]: Expected [X], Found [Y]
  - **Impact**: [What this breaks]
  - **Fix**: [How to resolve]

## Gate Decision
**Status**: [PASS | FAIL]
**Action**: [Proceed | Return to research]
```

---

## Continuous Validation

### Lifecycle Stages

| Stage | Validation Type |
|-------|-----------------|
| Pre-implementation | Assumption validation |
| During development | Unit tests, linting |
| Pre-commit | Full test suite |
| Pre-deploy | Integration tests |
| Post-deploy | Smoke tests, monitoring |

### Continuous Checks

```yaml
# CI pipeline validation
validate:
  pre-commit:
    - lint
    - type-check
    - unit-tests

  pre-merge:
    - integration-tests
    - security-scan
    - coverage-check

  pre-deploy:
    - smoke-tests
    - config-validation
    - dependency-audit

  post-deploy:
    - health-checks
    - synthetic-monitoring
    - alert-verification
```

### Validation Hooks

```bash
# Pre-commit hook
#!/bin/bash
set -euo pipefail

echo "Running pre-commit validation..."
npm run lint
npm run test:unit
npm run typecheck
```

---

## Test Strategy Planning

### Test Pyramid

```
        /\
       /  \  E2E (few)
      /----\
     /      \  Integration (some)
    /--------\
   /          \  Unit (many)
  /______________\
```

### Strategy Template

```markdown
# Test Strategy: [Feature]

## Scope
- [What's being tested]
- [What's out of scope]

## Test Types

### Unit Tests
| Component | Tests | Coverage Target |
|-----------|-------|-----------------|
| [Core logic] | 20 | 95% |
| [Utilities] | 10 | 90% |

### Integration Tests
| Integration | Tests | Focus |
|-------------|-------|-------|
| [API + DB] | 5 | Data flow |
| [Service A + B] | 3 | Communication |

### E2E Tests
| Flow | Priority | Automation |
|------|----------|------------|
| [Happy path] | P0 | Yes |
| [Error recovery] | P1 | Yes |

## Test Data
- [Data source 1]
- [Data source 2]

## Environment
- Unit: Local
- Integration: Docker Compose
- E2E: Staging

## Success Criteria
- [ ] Unit coverage > 90%
- [ ] All integration tests pass
- [ ] E2E happy path works
```

---

## Tracer Bullet Deployment

### Purpose
Deploy minimal resources to validate critical assumptions before full implementation.

### Approach
1. Identify critical path
2. Deploy minimum viable component
3. Validate end-to-end flow
4. Document findings
5. Iterate or proceed

### Tracer Bullet Template

```markdown
# Tracer Bullet: [Component]

## Objective
Validate that [critical assumption] works in [environment].

## Minimal Deployment
```yaml
# Smallest possible manifest
apiVersion: v1
kind: Pod
metadata:
  name: tracer-test
spec:
  containers:
  - name: test
    image: [image]
    command: ["echo", "success"]
```

## Validation Steps
1. Deploy minimal resource
2. Check status: `kubectl get pod tracer-test`
3. Verify logs: `kubectl logs tracer-test`
4. Clean up: `kubectl delete pod tracer-test`

## Results
| Check | Status | Notes |
|-------|--------|-------|
| Image pulls | ✅ | |
| Pod runs | ✅ | |
| Network works | ❌ | Missing network policy |

## Findings
- [Finding 1]: [Impact]
- [Finding 2]: [Impact]

## Next Steps
- [ ] Fix network policy
- [ ] Retry tracer bullet
- [ ] Proceed to full deployment
```

---

## Pre-Flight Checks

### Standard Checklist

```markdown
# Pre-Flight Checklist

## Environment
- [ ] Target namespace exists
- [ ] RBAC permissions granted
- [ ] Resource quotas sufficient
- [ ] Network policies allow traffic

## Dependencies
- [ ] Required operators installed
- [ ] External services accessible
- [ ] Secrets configured
- [ ] ConfigMaps present

## Code
- [ ] Tests passing
- [ ] Lint passing
- [ ] No security vulnerabilities
- [ ] Documentation updated

## Deployment
- [ ] Manifests validated (dry-run)
- [ ] Rollback plan documented
- [ ] Monitoring configured
- [ ] Alerts set up
```

### Automated Pre-Flight

```bash
#!/bin/bash
set -euo pipefail

echo "Running pre-flight checks..."

# Environment
kubectl auth can-i create deployment -n $NAMESPACE || exit 1
kubectl get ns $NAMESPACE || exit 1

# Dependencies
kubectl get secret $SECRET_NAME -n $NAMESPACE || exit 1
kubectl get cm $CONFIG_NAME -n $NAMESPACE || exit 1

# Dry run
kubectl apply --dry-run=server -f manifests/ || exit 1

echo "Pre-flight checks passed ✅"
```

---

## Failure Pattern Prevention

### Patterns This Skill Prevents

| Pattern | Prevention |
|---------|------------|
| Tests Passing Lie | Validate against real environment |
| Copy-Pasta Blindspot | Check assumptions before using examples |
| External Dependency | Verify integrations work |
| Environment Drift | Continuous validation |

### Validation Gates

```yaml
gates:
  research_complete:
    - assumptions_documented
    - evidence_collected

  assumptions_validated:
    - all_apis_verified
    - all_images_pullable
    - permissions_granted

  ready_to_implement:
    - validation_passed
    - tracer_bullet_succeeded
    - test_strategy_approved
```
