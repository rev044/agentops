# Explicit Skill Request Tests

Validates that natural-language trigger phrases correctly resolve to the expected skill. Each `.txt` file in `prompts/` contains a phrase that should match the skill named by the filename. The test runner feeds each prompt through the skill-matching logic and asserts the correct skill is selected, catching regressions in trigger phrase routing.

## Running

```bash
bash tests/explicit-skill-requests/run-all.sh
```
