# End-to-End Tests

Flywheel proof-run harness that starts from a raw transcript fixture, forges pending learnings, ingests and promotes them through the pool, retrieves the promoted artifact, records applied citation evidence, and closes feedback with a deterministic success outcome. It is fully automated, local-only, CI-runnable, and uses a temporary work directory with fixtures.

## Running

```bash
bash tests/e2e/proof-run.sh
```
