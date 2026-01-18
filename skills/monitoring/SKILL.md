---
name: monitoring
description: >
  Use when: "monitoring", "observability", "metrics", "tracing", "OpenTelemetry",
  "alerts", "runbooks", "SLO", "SLI", "dashboards", "Prometheus", "Grafana",
  "performance", "latency", "throughput", "Core Web Vitals", "load testing".
version: 1.0.0
author: "AgentOps Team"
license: "MIT"
---

# Monitoring Skill

Observability, alerting, and performance engineering patterns.

## Quick Reference

| Area | Key Patterns | When to Use |
|------|--------------|-------------|
| **Alerts & Runbooks** | Actionable alerts, runbook linking | Alert setup |
| **Performance** | OpenTelemetry, tracing, load testing | Optimization |

---

## Alerting Best Practices

### Alert Design Principles

| Principle | Good | Bad |
|-----------|------|-----|
| **Actionable** | "Database connections > 90%" | "Something wrong" |
| **Specific** | "Order service latency p99 > 500ms" | "Slow" |
| **Documented** | Links to runbook | No context |
| **Appropriate** | Pages for real issues | Alert fatigue |

### Alert Severity Levels

| Level | Response | Examples |
|-------|----------|----------|
| **Critical** | Page immediately | Service down, data loss |
| **Warning** | Check within 1 hour | Degraded performance |
| **Info** | Review daily | Capacity planning |

### Alert Template

```yaml
# Prometheus alert example
groups:
  - name: service-alerts
    rules:
      - alert: HighLatency
        expr: histogram_quantile(0.99, http_request_duration_seconds_bucket) > 0.5
        for: 5m
        labels:
          severity: warning
          service: api
        annotations:
          summary: "High latency on {{ $labels.service }}"
          description: "p99 latency is {{ $value }}s (threshold: 0.5s)"
          runbook_url: "https://wiki/runbooks/high-latency"
          dashboard_url: "https://grafana/d/api-latency"
```

### Runbook-Linked Alerts

Every alert should link to:
1. **Runbook** - What to do
2. **Dashboard** - Where to look
3. **Escalation** - Who to contact

```markdown
## Alert: HighLatency

### Quick Check
- Dashboard: [link]
- Recent deploys: `git log --since="1 hour ago"`

### Common Causes
1. **High traffic** → Scale horizontally
2. **Database slow** → Check connection pool
3. **Upstream delay** → Check dependency

### Resolution
[Step-by-step instructions]

### Escalation
After 30 min: Page @backend-team
```

---

## Observability Stack

### Three Pillars

| Pillar | Purpose | Tools |
|--------|---------|-------|
| **Metrics** | What's happening | Prometheus, Datadog |
| **Logs** | Why it's happening | ELK, Loki |
| **Traces** | How it's happening | Jaeger, Tempo |

### OpenTelemetry Setup

```python
# Python OpenTelemetry example
from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter

# Setup
provider = TracerProvider()
processor = BatchSpanProcessor(OTLPSpanExporter(endpoint="otel-collector:4317"))
provider.add_span_processor(processor)
trace.set_tracer_provider(provider)

# Usage
tracer = trace.get_tracer(__name__)

@tracer.start_as_current_span("process_order")
def process_order(order_id):
    span = trace.get_current_span()
    span.set_attribute("order.id", order_id)
    # ... process
```

### Distributed Tracing

```
[Request] → [API Gateway] → [Order Service] → [Payment Service]
                                            → [Inventory Service]
                                            → [Notification Service]

Trace ID: abc123
├── Span: api-gateway (10ms)
├── Span: order-service (45ms)
│   ├── Span: validate-order (5ms)
│   ├── Span: payment-call (25ms)
│   └── Span: save-order (15ms)
└── Span: notification (8ms)
```

---

## Performance Engineering

### Key Metrics

| Category | Metrics |
|----------|---------|
| **Latency** | p50, p95, p99 response time |
| **Throughput** | Requests per second |
| **Errors** | Error rate, error types |
| **Saturation** | CPU, memory, connections |

### SLI/SLO Framework

```markdown
## SLO: Order Service

### Availability
- **SLI**: Successful requests / Total requests
- **SLO**: 99.9% over 30 days
- **Error budget**: 43.2 minutes/month

### Latency
- **SLI**: Requests < 500ms / Total requests
- **SLO**: 95% of requests < 500ms
- **Error budget**: 5% can be slow

### Dashboard
[Link to SLO dashboard]
```

### Load Testing

```javascript
// k6 load test example
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '2m', target: 100 },  // Ramp up
    { duration: '5m', target: 100 },  // Stay at 100
    { duration: '2m', target: 200 },  // Spike
    { duration: '2m', target: 0 },    // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],
    http_req_failed: ['rate<0.01'],
  },
};

export default function () {
  const res = http.get('https://api.example.com/orders');
  check(res, {
    'status is 200': (r) => r.status === 200,
    'latency < 500ms': (r) => r.timings.duration < 500,
  });
  sleep(1);
}
```

### Core Web Vitals

| Metric | Good | Needs Work | Poor |
|--------|------|------------|------|
| **LCP** | < 2.5s | 2.5-4s | > 4s |
| **FID** | < 100ms | 100-300ms | > 300ms |
| **CLS** | < 0.1 | 0.1-0.25 | > 0.25 |

### Performance Optimization Checklist

```markdown
## Performance Audit

### Backend
- [ ] Database queries optimized (no N+1)
- [ ] Appropriate caching in place
- [ ] Connection pooling configured
- [ ] Async where appropriate

### Frontend
- [ ] Assets minified and compressed
- [ ] Images optimized (WebP, lazy loading)
- [ ] Code splitting implemented
- [ ] CDN configured

### Infrastructure
- [ ] Horizontal scaling ready
- [ ] Load balancer configured
- [ ] Health checks in place
- [ ] Auto-scaling policies set
```

---

## Dashboard Design

### Dashboard Template

```markdown
## Dashboard: [Service Name]

### Row 1: Golden Signals
- Request Rate (RPS)
- Error Rate (%)
- Latency (p50, p95, p99)
- Saturation (CPU, Memory)

### Row 2: Business Metrics
- Orders/minute
- Revenue/hour
- Active users

### Row 3: Dependencies
- Database latency
- Cache hit rate
- External API status

### Row 4: Infrastructure
- Pod count
- Node health
- Disk usage
```

### Grafana Panel Examples

```yaml
# Request rate
rate(http_requests_total{service="api"}[5m])

# Error rate
sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m]))

# p99 latency
histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))

# Saturation
container_memory_usage_bytes / container_spec_memory_limit_bytes
```

---

## Capacity Planning

### Capacity Model

```markdown
## Capacity Planning: [Service]

### Current State
- Peak RPS: 1,000
- Average latency: 50ms
- Pod count: 5
- CPU per pod: 500m

### Growth Projection
| Timeline | RPS | Pods Needed |
|----------|-----|-------------|
| Current | 1,000 | 5 |
| +3 months | 1,500 | 8 |
| +6 months | 2,500 | 13 |

### Scaling Triggers
- CPU > 70% for 5 minutes → Scale up
- RPS > 800 per pod → Scale up
- Memory > 80% → Scale up

### Bottlenecks
- Database connections (max 100)
- External API rate limits (1000/min)
```
