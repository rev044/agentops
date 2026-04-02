---
type: learning
maturity: candidate
confidence: high
utility: 0.7
---
# CI Pipeline Cache Invalidation

Cache invalidation in CI pipelines is the second most common cause of stale artifact failures after dependency drift. A reliable CI pipeline cache invalidation strategy keys the cache on a hash of the dependency manifest rather than branch name alone. Treating cache invalidation as an explicit CI pipeline step — with a force-bust flag — makes debugging reproducible.
