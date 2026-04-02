---
type: learning
maturity: candidate
confidence: high
utility: 0.5
---
# Database Schema Version Tracking

Version tracking for database schema changes must be stored in the database itself, not only in migration files on disk. A database schema version tracking table that records migration hash, author, and timestamp enables auditing of which database schema changes were applied in which environment and when. Drift between expected and actual database schema version state is the leading cause of environment-specific query failures.
