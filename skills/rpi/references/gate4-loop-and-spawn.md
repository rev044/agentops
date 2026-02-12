# Gate 4 Loop and Spawn Next Work

## Phase 6.5: Gate 4 Loop (Optional) -- Post-mortem to Spawn Another /rpi

**Default behavior:** /rpi ends after Phase 6.

**Enable loop:** pass `--loop` (and optionally `--max-cycles=<n>`).

**Gate 4 goal:** make the "ITERATE vs TEMPER" decision explicit, and if iteration is required, run another full /rpi cycle with tighter context.

**Loop decision input:** the most recent post-mortem council verdict.

1. Find the most recent post-mortem report:
   ```bash
   REPORT=$(ls -t .agents/council/*post-mortem*.md 2>/dev/null | head -1)
   ```
2. Read `REPORT` and extract the verdict line (`## Council Verdict: PASS / WARN / FAIL`).
3. Apply gate logic (only when `--loop` is set). If verdict is PASS or WARN, stop (TEMPER path). If verdict is FAIL, iterate (spawn another /rpi cycle), up to `--max-cycles`.
4. Iterate behavior (spawn). Read the post-mortem report and extract 3 concrete fixes, then re-invoke /rpi from Phase 1 with a tightened goal that includes the fixes:
   ```
   /rpi "<original goal> (Iteration <n>): Fix <item1>; <item2>; <item3>" --test-first   # if --test-first set
   /rpi "<original goal> (Iteration <n>): Fix <item1>; <item2>; <item3>"                 # otherwise
   ```
   If still FAIL after `--max-cycles` total cycles, stop and require manual intervention (file follow-up bd issues).

## Phase 6.6: Spawn Next Work (Optional) -- Post-mortem to Queue Next RPI

**Enable:** pass `--spawn-next` flag.

**Complementary to Gate 4:** Gate 4 (`--loop`) handles FAIL->iterate (same goal, tighter). `--spawn-next` handles PASS/WARN->new-goal (different work harvested from post-mortem).

1. Read `.agents/rpi/next-work.jsonl` for unconsumed entries (schema: `.agents/rpi/next-work.schema.md`)
2. If unconsumed entries exist:
   - If `--dry-run` is set: report items but do NOT mutate next-work.jsonl (skip consumption). Log: "Dry run -- items not marked consumed."
   - Otherwise: mark the current cycle's entry as consumed (set `consumed: true`, `consumed_by: <epic-id>`, `consumed_at: <now>`)
   - Report harvested items to user with suggested next command:
     ```
     ## Next Work Available

     Post-mortem harvested N follow-up items from <source_epic>:
     1. <title> (severity: <severity>, type: <type>)
     ...

     To start the next RPI cycle:
       /rpi "<highest-severity item title>"
     ```
   - Do NOT auto-invoke `/rpi` -- the user decides when to start the next cycle
3. If no unconsumed entries: report "No follow-up work harvested. Flywheel stable."

**Note:** Only `--spawn-next` mutates next-work.jsonl (marks consumed). Phase 0 read is read-only.
