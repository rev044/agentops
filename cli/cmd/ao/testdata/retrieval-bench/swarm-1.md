---
type: learning
maturity: candidate
confidence: high
utility: 0.75
---
# Swarm Parallel Execution File Conflicts

Detecting file conflicts in swarm parallel execution must happen before workers are dispatched, not after they collide. Swarm parallel execution file conflict detection requires building a file ownership map across all planned worker tasks and rejecting the wave if any file appears in more than one worker's scope. Post-hoc merge resolution in swarm parallel execution is expensive and often introduces subtle bugs that neither conflicting worker's tests would catch independently.
