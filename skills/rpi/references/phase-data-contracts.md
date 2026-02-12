# Phase Data Contracts

How each phase passes data to the next. Artifacts are filesystem-based; no in-memory coupling between phases.

| Transition | Output | Extraction | Input to Next |
|------------|--------|------------|---------------|
| → Research | `.agents/research/YYYY-MM-DD-<slug>.md` | `ls -t .agents/research/ \| head -1` | /plan reads .agents/research/ automatically |
| Research → Plan | Plan doc + bd epic | Most recent epic from `bd list --type epic` | epic-id stored in session state |
| Plan → Pre-mortem | `.agents/plans/YYYY-MM-DD-<slug>.md` | /pre-mortem auto-discovers most recent plan | No args needed |
| Pre-mortem → Crank | Council report with verdict | Grep verdict from council report | epic-id passed to /crank |
| Crank → Vibe | Committed code + closed issues | Check `<promise>` tag | /vibe runs on recent changes |
| Vibe → Post-mortem | Council report with verdict | Grep verdict from council report | epic-id passed to /post-mortem |
