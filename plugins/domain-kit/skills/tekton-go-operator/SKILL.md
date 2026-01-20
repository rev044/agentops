---
name: tekton-go-operator
description: >
  Interactive guidance for setting up Tekton CI on Go operators (kubebuilder).
  Warns about common friction points: .dockerignore artifacts, private deps, envtest.
version: 1.0.0
context: memory
triggers:
  - "set up tekton ci for my operator"
  - "tekton go operator"
  - "migrate to tekton ci"
  - "operator ci pipeline"
  - "tekton for kubebuilder"
  - "go operator ci"
allowed-tools: Bash, Read, Glob, Grep
skills:
  - beads
  - tekton
  - standards
---

# tekton-go-operator: Tekton CI for Go Operators

> **Interactive guidance for setting up Tekton CI on kubebuilder/controller-runtime operators.**

## Pre-flight Checks

Before setting up Tekton CI, verify your cluster has the prerequisites:

```bash
# 1. ClusterTasks exist
oc get clustertask -l app.kubernetes.io/part-of=jren-tekton

# 2. CI namespace exists
oc get ns olympus-ci

# 3. Registry credentials configured
oc get secret registry-credentials -n olympus-ci
```

---

## Friction Point Warnings

### 1. .dockerignore Blocking Artifacts

**Problem**: If your `.dockerignore` uses deny-all pattern (`**`), it blocks artifacts produced by upstream Tasks.

**Symptom**: "file not found" errors in Kaniko build step.

**Solution**: Add explicit exceptions for Task-produced artifacts:

```
# .dockerignore
**

# Allow source files
!go.mod
!go.sum
!**/*.go
**/*_test.go

# CRITICAL: Allow pre-built binaries from Tasks
!bin/*
```

**When this applies**: Any pipeline where one Task builds artifacts for another (e.g., `build-artifact` → `kaniko-build`).

---

### 2. Private Dependency Strategy

**Problem**: Building from private repos during Kaniko exposes credentials in layers.

**Solutions** (choose one):

| Strategy | Pros | Cons | When to Use |
|----------|------|------|-------------|
| **Public fork** | Zero credential management | Must keep synced | Build-time deps that rarely change |
| **Pre-build Task** | Artifacts only, no creds in image | Extra Task | Private deps with frequent changes |
| **Git credentials Secret** | Full access | Complex, credentials in cluster | Last resort |

**Recommended**: Use pre-build Task pattern (see `jren-build-artifact` ClusterTask).

```yaml
# Example: Build artifact in separate Task
- name: build-dependency
  taskRef:
    name: jren-build-artifact
    kind: ClusterTask
  params:
    - name: artifact-url
      value: "https://github.com/your-org/private-dep.git"
```

---

### 3. envtest K8s Version Mismatch

**Problem**: envtest K8s version must match controller-runtime's supported version.

**Symptom**: API version errors, "unknown resource" in tests.

**Version Matrix**:

| controller-runtime | Supported K8s | envtest Version |
|--------------------|---------------|-----------------|
| v0.22.x | 1.31.x | 1.31.x |
| v0.21.x | 1.30.x | 1.30.x |
| v0.20.x | 1.29.x | 1.29.x |
| v0.19.x | 1.28.x | 1.28.x |

**How to check your version**:

```bash
grep 'sigs.k8s.io/controller-runtime' go.mod
# Then check controller-runtime README for K8s compatibility
```

**In your go-test Task**:

```yaml
params:
  - name: envtest-k8s-version
    value: "1.31.x"  # Match your controller-runtime
```

---

## Quick Setup

### Option A: Use Formula (Recommended)

```bash
# 1. Generate all files from template
/formulate go-operator-tekton-ci

# 2. Review and apply
oc apply -f deploy/tekton/

# 3. Run pipeline
oc create -f deploy/tekton/pipelinerun.yaml
```

### Option B: Manual Setup

```bash
# 1. Copy Tasks from existing operator (e.g., gastown-operator)
mkdir -p deploy/tekton/tasks

# 2. Create Pipeline referencing ClusterTasks
cat > deploy/tekton/pipeline.yaml << 'EOF'
apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: my-operator-ci
spec:
  workspaces:
    - name: source
    - name: cache
  tasks:
    - name: clone
      taskRef:
        name: git-clone
        kind: ClusterTask
    # ... see gastown-operator for full example
EOF

# 3. Create PipelineRun template
# 4. Update .dockerignore if needed
# 5. Run pipeline
```

---

## Pipeline Structure

Standard Go operator pipeline stages:

```
┌────────────┐
│ git-clone  │
└─────┬──────┘
      │
      ▼
┌─────────────────────────────────────────┐
│              PARALLEL                    │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ │
│  │ go-test  │ │ trivy-fs │ │ hadolint │ │
│  └──────────┘ └──────────┘ └──────────┘ │
└──────────────────┬──────────────────────┘
                   │
                   ▼
          ┌────────────────┐
          │ build-artifact │  (if external deps)
          └───────┬────────┘
                  │
                  ▼
          ┌────────────────┐
          │  kaniko-build  │
          └───────┬────────┘
                  │
                  ▼
┌─────────────────────────────────────────┐
│              PARALLEL                    │
│  ┌─────────────┐ ┌─────────────────────┐│
│  │ trivy-image │ │ generate-sbom       ││
│  └─────────────┘ └─────────────────────┘│
└─────────────────────────────────────────┘
```

---

## Available ClusterTasks

| ClusterTask | Purpose |
|-------------|---------|
| `git-clone` | Clone source repo |
| `jren-go-test` | Run Go tests with envtest |
| `jren-build-artifact` | Build binary from external repo |
| `jren-kaniko-build` | Build container image |
| `jren-trivy-fs` | Scan filesystem for vulnerabilities |
| `jren-trivy-image` | Scan container image |
| `jren-hadolint-scan` | Lint Dockerfile |
| `jren-generate-sbom` | Generate CycloneDX SBOM |

---

## GitLab CI Images - DPR Registry

If your operator also has GitLab CI (not just Tekton), all CI images **must** come from DPR to avoid Docker Hub rate limits.

```yaml
# .gitlab-ci.yml
variables:
  DPR_REGISTRY: "dprusocplvjmp01.deepsky.lab:5000"
  GO_IMAGE: "${DPR_REGISTRY}/ci-images/golang:1.24"
  GO_LINT_IMAGE: "${DPR_REGISTRY}/ci-images/golangci-lint:v2.7.2"
  KUBECTL_IMAGE: "${DPR_REGISTRY}/ci-images/kubectl:latest"
```

**Mirror missing images:**
```bash
cd ~/gt/release_engineering/crew/<user>
./scripts/mirror-ci-images.sh --check
./scripts/mirror-ci-images.sh
```

---

## Troubleshooting

### "file not found" in Kaniko

Check `.dockerignore` - likely blocking artifacts from upstream Tasks.

### envtest "unknown API" errors

Verify envtest K8s version matches controller-runtime. See version matrix above.

### Private repo auth failures

Use pre-build Task pattern instead of cloning during Dockerfile build.

### Pipeline hangs on tests

Check envtest binary download - may need to pre-cache in base image or allow network access.

---

## Related

- [gastown-operator Tekton CI](https://github.com/olympus/gastown-operator/tree/main/deploy/tekton) - Working example
- [Tekton Pre-build Artifact Pattern](~/.agents/patterns/tekton-prebuild-artifact.md) - Detailed pattern
- [tekton](tekton/) - General Tekton guidance
- [go-operator-tekton-ci.formula.toml](~/.claude/molecules/) - Automated setup

---

## Standards Loading

When configuring Go operator CI, reference these standards:

| File Type | Load Reference |
|-----------|----------------|
| `*.go` | `standards/references/go.md` |
| Pipeline YAML | `standards/references/yaml.md` |
