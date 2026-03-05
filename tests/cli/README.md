# CLI Tests

Validates that the `ao` CLI's `--json` flag produces valid JSON output across all subcommands that support it. Ensures flag consistency between documented behavior and actual output, checks that JSON responses parse without errors, and verifies expected fields are present. Requires the `ao` binary (auto-builds if missing).

## Running

```bash
bash tests/cli/test-json-flag-consistency.sh
```
