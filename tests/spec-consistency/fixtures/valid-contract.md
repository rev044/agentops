# Contract: Add Rate Limiting Middleware

---

```yaml
# --- Contract Frontmatter ---
issue:      ag-xyz.5
framework:  go
category:   feature
```

---

## Problem

API endpoints accept unlimited requests per client, enabling abuse and risking resource exhaustion.

## Inputs

- `request` (*http.Request) — incoming HTTP request with client IP in RemoteAddr
- `config.RateLimit` (int) — max requests per window per client (default: 100)
- `config.RateWindow` (time.Duration) — sliding window duration (default: 1 minute)

## Outputs

- **Pass-through** — request forwarded to next handler
- **429 response** — JSON error body with `Retry-After` header

## Invariants

1. A client sending <= `RateLimit` requests within `RateWindow` is never rejected.
2. A client exceeding `RateLimit` within `RateWindow` receives HTTP 429.
3. Rate limit state for one client never affects another client's quota.

## Failure Modes

1. **Rate store unreachable** → fail-open, log error.
2. **Malformed RemoteAddr** → treat as unknown client, apply default limit.

## Out of Scope

- Distributed rate limiting across multiple instances.

## Test Cases

| # | Input | Expected | Validates Invariant |
|---|-------|----------|---------------------|
| 1 | 100 requests from same IP in 60s | All 100 return 200 | #1 |
| 2 | 101st request from same IP in 60s | Returns 429 with Retry-After | #2 |
| 3 | 100 requests each from IP-A and IP-B | All 200 requests return 200 | #3 |
