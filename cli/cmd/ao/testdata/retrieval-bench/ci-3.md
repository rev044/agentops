---
type: learning
maturity: provisional
confidence: high
utility: 0.6
---
# CI Pipeline Flaky Test Detection

Detecting flaky tests in CI pipeline runs requires tracking per-test pass/fail history across multiple CI pipeline executions rather than a single run. A flaky test quarantine label applied automatically after three non-deterministic failures in the CI pipeline prevents blocking merges while keeping signal. Re-admitting quarantined tests back to the main CI pipeline gate requires a clean streak of ten consecutive passes.
