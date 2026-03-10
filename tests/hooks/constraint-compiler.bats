#!/usr/bin/env bats

setup() {
  REPO_ROOT="$(cd "$BATS_TEST_DIRNAME/../.." && pwd)"
  FINDING_COMPILER="$REPO_ROOT/hooks/finding-compiler.sh"
  CONSTRAINT_WRAPPER="$REPO_ROOT/hooks/constraint-compiler.sh"
  TMP_REPO="$(mktemp -d)"
  cd "$TMP_REPO"
  git init -q
}

teardown() {
  rm -rf "$TMP_REPO"
}

@test "finding compiler promotes registry entries into advisory artifacts" {
  mkdir -p .agents/findings
  cat > .agents/findings/registry.jsonl <<'EOF'
{"id":"f-2026-03-09-777","version":1,"tier":"local","source":{"repo":"agentops/crew/nami","session":"2026-03-09","file":".agents/council/source.md","skill":"pre-mortem"},"date":"2026-03-09","severity":"significant","category":"validation-gap","pattern":"Prior findings should shape planning before decomposition.","detection_question":"Did planning load active findings first?","checklist_item":"Load active findings before decomposition.","applicable_languages":["markdown","shell"],"applicable_when":["plan-shape","validation-gap"],"status":"active","superseded_by":null,"dedup_key":"validation-gap|prior-findings-should-shape-planning-before-decomposition|plan-shape","hit_count":0,"last_cited":null,"ttl_days":30,"confidence":"high"}
EOF

  run bash "$FINDING_COMPILER"
  [ "$status" -eq 0 ]

  [ -f ".agents/findings/f-2026-03-09-777.md" ]
  [ -f ".agents/planning-rules/f-2026-03-09-777.md" ]
  [ -f ".agents/pre-mortem-checks/f-2026-03-09-777.md" ]
  [ ! -f ".agents/constraints/f-2026-03-09-777.sh" ]
}

@test "finding compiler compiles mechanical findings into declarative constraint entries" {
  mkdir -p .agents/findings
  cat > .agents/findings/f-mechanical.md <<'EOF'
---
id: "f-mechanical"
type: "finding"
version: 1
date: "2026-03-09"
source_skill: "post-mortem"
source_artifact: ".agents/council/post-mortem.md"
title: "Need \"safe\" parsing with \\slashes"
summary: "Mechanical findings should preserve JSON-safe titles."
pattern: "Keep issue_type propagation checks active."
detection_question: "Did the changed files preserve issue_type?"
checklist_item: "Verify issue_type survives all task metadata paths."
severity: "significant"
detectability: "mechanical"
status: "draft"
compiler_targets: ["plan", "pre-mortem", "constraint"]
scope_tags: ["validation-gap", "task-sync"]
dedup_key: "validation-gap|keep-issue-type-propagation-checks-active|validation-gap"
applicable_when: ["validation-gap"]
applicable_languages: ["markdown", "shell", "go"]
compiler: {"review_file":".agents/constraints/f-mechanical.sh","applies_to":{"scope":"files","issue_types":["feature","bug","task"],"path_globs":["skills/*.md","hooks/*.sh"]},"detector":{"kind":"content_pattern","mode":"must_contain","pattern":"issue_type","message":"issue_type must remain present"}} 
---
# Finding: mechanical
EOF

  run bash "$FINDING_COMPILER" "$TMP_REPO/.agents/findings/f-mechanical.md"
  [ "$status" -eq 0 ]

  [ -x ".agents/constraints/f-mechanical.sh" ]
  [ -f ".agents/constraints/index.json" ]
  run jq -r '.constraints[0].title' ".agents/constraints/index.json"
  [ "$status" -eq 0 ]
  [[ "$output" == *safe* ]]
  run jq -r '.constraints[0].detector.kind' ".agents/constraints/index.json"
  [ "$status" -eq 0 ]
  [ "$output" = "content_pattern" ]
}

@test "constraint compiler wrapper promotes tagged legacy learnings through finding compiler" {
  cat > learning.md <<'EOF'
---
id: "learn-constraint"
title: "Constraint rule"
date: "2026-02-24"
tags: ["constraint", "reliability"]
---
This learning describes a guardrail to prevent direct bypass of safety checks.
EOF

  run bash "$CONSTRAINT_WRAPPER" "$TMP_REPO/learning.md"
  [ "$status" -eq 0 ]
  [[ "$output" == *"Promoted legacy learning"* || "$output" == *"finding-compiler.sh"* ]]

  [ -f ".agents/findings/learn-constraint.md" ]
  [ -f ".agents/planning-rules/learn-constraint.md" ]
  [ -f ".agents/pre-mortem-checks/learn-constraint.md" ]
  [ -x ".agents/constraints/learn-constraint.sh" ]
  [ -f ".agents/constraints/index.json" ]
}
