---
name: tekton
version: 1.0.0
context: fork
description: >
  Tekton pipeline builds for ai-platform services. Triggers on "tekton",
  "pipeline build", "build image", "tkn pipeline", "pipelinerun".
triggers:
  - "tekton"
  - "pipeline build"
  - "container build"
  - "build image"
  - "tkn pipeline"
  - "pipelinerun"
skills:
  - standards
---

# Tekton Build Skill

This skill provides verified patterns for running Tekton builds in the ai-platform namespace on ocppoc.

---

## Critical Facts (Memorize)

### Git URL

**WRONG**: `https://git.deepskylab.io/openshift/admin/ai-platform`
**CORRECT**: `https://git.deepskylab.io/openshift/admin/ai-platform`

### Required Resources

| Resource | Name | Purpose |
|----------|------|---------|
| PVC | `pipeline-workspace` | Shared workspace (RWO - one pipeline at a time) |
| Secret | `registry-credentials` | Push to internal registry |
| Secret | `git-credentials` | Clone private repos (needs `tekton.dev/git-0` annotation) |

### Pipelines

| Pipeline | Use For |
|----------|---------|
| `jren-standard-build` | Most services |
| `jren-python-build` | Python services |

---

## Golden Command Template

```bash
tkn pipeline start jren-standard-build \
  -n ai-platform \
  -p git-url=https://git.deepskylab.io/openshift/admin/ai-platform \
  -p git-revision=main \
  -p image-name=dprusocplvjmp01.deepsky.lab:5000/ai-platform/<SERVICE>:<TAG> \
  -p context-dir=services/<PATH> \
  -p dockerfile-path=<FILE> \
  -w name=source,claimName=pipeline-workspace \
  -w name=dockerconfig,secret=registry-credentials \
  --use-param-defaults
```

---

## Common Failures

### "repository not found" / 404

Git URL is wrong. Use `fullerbt/ai-platform`, not `openshift/admin/ai-platform`.

### "permission denied" cleaning workspace

Shared PVC corrupted. Clean it:
```bash
oc run pvc-cleanup --image=registry.access.redhat.com/ubi9/ubi-minimal:latest \
  --rm -it --restart=Never --overrides='{
  "spec": {
    "containers": [{"name": "pvc-cleanup", "image": "registry.access.redhat.com/ubi9/ubi-minimal:latest",
      "command": ["sh", "-c", "rm -rf /workspace/* /workspace/.[!.]* 2>/dev/null; echo done"],
      "securityContext": {"runAsUser": 0},
      "volumeMounts": [{"name": "workspace", "mountPath": "/workspace"}]}],
    "volumes": [{"name": "workspace", "persistentVolumeClaim": {"claimName": "pipeline-workspace"}}]
  }}' -n ai-platform
```

### Git credentials expired

```bash
oc delete secret git-credentials -n ai-platform
oc create secret generic git-credentials \
  -n ai-platform \
  --from-literal=username=gitlab-ci-token \
  --from-literal=password="$GITLAB_TOKEN" \
  --type=kubernetes.io/basic-auth
oc annotate secret git-credentials -n ai-platform tekton.dev/git-0=https://git.deepskylab.io
```

### Docker Hub rate limit

If GitLab CI jobs fail with "toomanyrequests: You have reached your unauthenticated pull rate limit":

**Root cause:** Using Docker Hub images directly in .gitlab-ci.yml

**Fix:** All CI images must come from DPR (DeepSky Private Registry):

```yaml
# WRONG:
image: golang:1.24
image: golangci/golangci-lint:v2.7.2

# CORRECT:
variables:
  DPR_REGISTRY: "dprusocplvjmp01.deepsky.lab:5000"
  GO_IMAGE: "${DPR_REGISTRY}/ci-images/golang:1.24"
  GO_LINT_IMAGE: "${DPR_REGISTRY}/ci-images/golangci-lint:v2.7.2"
```

**Mirror missing images:**
```bash
cd ~/gt/release_engineering/crew/boden
./scripts/mirror-ci-images.sh --check   # See what's missing
./scripts/mirror-ci-images.sh           # Mirror all
```

---

## Go Operators

For Go operators (kubebuilder/controller-runtime), use the specialized skill:

**`tekton-go-operator`** - Interactive guidance for:
- envtest version matching
- .dockerignore artifact patterns
- Pre-build artifact pattern for private deps
- ClusterTasks: `jren-go-test`, `jren-build-artifact`

Triggers: "tekton go operator", "set up tekton ci for my operator"

**Quick setup**: `/formulate go-operator-tekton-ci`

---

## JIT Load

For full documentation: Read `docs/standards/tekton-builds.md`
For troubleshooting: Read `charts/tekton-build/docs/troubleshooting.md`
For pipeline stages: Read `charts/tekton-build/docs/pipeline-guide.md`
For Go operators: Read `domain-kit/skills/tekton-go-operator/SKILL.md`

---

## Monitoring

```bash
# Watch logs
tkn pipelinerun logs -f -n ai-platform

# Check status
oc get pipelinerun -n ai-platform --sort-by='.metadata.creationTimestamp' | tail -5
```

---

## Standards Loading

When configuring pipelines, reference these standards:

| File Type | Load Reference |
|-----------|----------------|
| Pipeline YAML | `domain-kit/skills/standards/references/yaml.md` |
