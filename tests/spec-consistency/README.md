# Spec Consistency Tests

Tests for the `scripts/spec-consistency-gate.sh` validation gate. Exercises each failure mode using temporary fixture directories to ensure that skill specs, schemas, and contracts remain internally consistent across the repository. The test harness uses a pass/fail accumulator pattern and the `fixtures/` directory for controlled test scenarios.

## Running

```bash
bash tests/spec-consistency/test-gate.sh
```
