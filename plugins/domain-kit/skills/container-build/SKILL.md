---
name: container-build
version: 1.0.0
description: >
  This skill should be used when the user asks to "build container", "create Dockerfile",
  "configure podman", or needs guidance on OpenShift container images, arbitrary UIDs,
  or image tagging.
context: fork
allowed-tools: "Read,Bash,Grep,Glob"
skills:
  - standards
---

# Container Build Skill for OpenShift

Build container images that work on OpenShift with arbitrary UIDs.

## Critical Requirements

### 1. Platform Targeting (MANDATORY)

**Always build for AMD64** - the cluster runs AMD64, Mac builds ARM64 by default:

```bash
podman build --platform linux/amd64 -t registry/image:tag .
```

### 2. OpenShift File Permissions (MANDATORY)

OpenShift runs containers with **arbitrary UIDs** in the root group (GID 0). Files must be group-readable/executable:

```dockerfile
# Bad - only owner can read
COPY main.py .

# Good - group can read
COPY --chown=1001:0 main.py .
RUN chmod g+r main.py requirements.txt

# For executables
RUN chmod g+rx entrypoint.sh

# For directories that need writing
RUN chmod -R g+rwX /app/data
```

### 3. Versioned Tags (MANDATORY)

**Never use `latest` tag** - Kubernetes caches images. Use semantic versions:

```bash
# Bad - cached images won't update
podman build -t registry/image:latest .

# Good - forces fresh pull
podman build -t registry/image:v1.0.0 .
```

Update `values.yaml` with the new tag AND use `imagePullPolicy: Always` or versioned tags.

### 4. Non-Root User Setup

```dockerfile
# Create non-root user with GID 0 for OpenShift
RUN useradd -u 1001 -g 0 -m appuser

# Set ownership to user:root-group
COPY --chown=1001:0 . /app

# Switch to non-root
USER 1001
```

### 5. Health Check Endpoints

**Match health probes to actual endpoints:**

| Framework | Default Endpoint | Probe Type |
|-----------|------------------|------------|
| FastAPI | `/health` or `/` | HTTP GET |
| FastMCP 2.0 | `/mcp` (returns 406 for plain GET) | **TCP Socket** |
| Express | `/health` | HTTP GET |

For MCP servers using FastMCP 2.0, use TCP probes:
```yaml
livenessProbe:
  tcpSocket:
    port: http
  initialDelaySeconds: 10
readinessProbe:
  tcpSocket:
    port: http
  initialDelaySeconds: 5
```

## Standard Containerfile Template

```dockerfile
# Multi-stage build for Python services
FROM registry.access.redhat.com/ubi9/python-312:latest AS builder

WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Production stage
FROM registry.access.redhat.com/ubi9/python-312:latest

WORKDIR /app

# Copy dependencies from builder
COPY --from=builder /opt/app-root /opt/app-root

# Copy application with correct ownership
COPY --chown=1001:0 main.py .
COPY --chown=1001:0 requirements.txt .

# Ensure group-readable for OpenShift arbitrary UID
RUN chmod g+r main.py requirements.txt

# Non-root user (already 1001 in UBI images)
USER 1001

EXPOSE 8080
CMD ["python", "main.py"]
```

## Build and Push Workflow

```bash
# 1. Build for correct platform
podman build --platform linux/amd64 \
  -t dprusocplvjmp01.deepsky.lab:5000/ai-platform/SERVICE:vX.Y.Z \
  -f services/SERVICE/Containerfile \
  services/SERVICE/

# 2. Push to registry
podman push dprusocplvjmp01.deepsky.lab:5000/ai-platform/SERVICE:vX.Y.Z

# 3. Update values.yaml with new tag
# 4. Helm upgrade
helm upgrade ai-platform ./charts/ai-platform -n ai-platform -f deploy/ocppoc/values.yaml

# 5. Verify rollout
oc rollout status deploy/ai-platform-SERVICE -n ai-platform
```

## Debugging Failed Containers

```bash
# Check pod status
oc get pods -n ai-platform -l app.kubernetes.io/component=SERVICE

# Check logs
oc logs -n ai-platform deploy/ai-platform-SERVICE --tail=50

# Common errors:
# - "exec format error" → Wrong platform (ARM64 on AMD64)
# - "Permission denied" → File permissions not group-readable
# - "404 Not Found" on health → Wrong probe endpoint
# - "406 Not Acceptable" → Use TCP probe instead of HTTP
```

## Checklist Before Building

- [ ] Containerfile uses `--platform linux/amd64` or builder does
- [ ] Files are `COPY --chown=1001:0`
- [ ] `chmod g+r` on all copied files
- [ ] Using versioned tag (vX.Y.Z), not `latest`
- [ ] Health probe matches actual endpoint
- [ ] values.yaml updated with new tag

---

## Standards Loading

When configuring containers, reference these standards:

| File Type | Load Reference |
|-----------|----------------|
| `values.yaml`, `*.yml` | `domain-kit/skills/standards/references/yaml.md` |
| `*.json` | `domain-kit/skills/standards/references/json.md` |
| Shell scripts | `domain-kit/skills/standards/references/shell.md` |
