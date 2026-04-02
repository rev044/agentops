---
type: learning
maturity: established
confidence: high
utility: 0.8
---
# Session Intelligence Trust Policies

Trust policy enforcement in session intelligence prevents low-provenance artifacts from contaminating ranked context. Session intelligence trust policies should be evaluated at injection time, not at retrieval time, so that a policy change takes effect immediately without re-indexing. Attaching a trust tier to every session intelligence artifact at write time makes policy enforcement in session intelligence fast and auditable.
