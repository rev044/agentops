---
type: learning
maturity: provisional
confidence: high
utility: 0.6
---
# Database Migration Rollback Strategy

Safe database migration rollback procedures require that every forward migration has a tested reverse migration written and validated before the forward migration is merged. A database migration rollback strategy that relies on restoring from backup is too slow for production incidents — schema-level reversibility must be built into each database migration step. Testing the rollback path in CI as part of the database migration pipeline prevents discovering rollback failures during an outage.
