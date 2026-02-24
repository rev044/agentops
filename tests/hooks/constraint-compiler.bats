#!/usr/bin/env bats

setup() {
  REPO_ROOT="$(cd "$BATS_TEST_DIRNAME/../.." && pwd)"
  SCRIPT="$REPO_ROOT/hooks/constraint-compiler.sh"
  TMP_REPO="$(mktemp -d)"
  cd "$TMP_REPO"
  git init -q
}

teardown() {
  rm -rf "$TMP_REPO"
}

@test "constraint compiler escapes JSON fields with quotes and backslashes" {
  cat > learning.md <<'EOF'
---
id: learn-special
title: "Need \"safe\" parsing with \\slashes"
date: 2026-02-24
tags: [constraint]
---
This summary line includes "quotes" and \slashes for JSON escaping.
EOF

  run bash "$SCRIPT" "$TMP_REPO/learning.md"
  [ "$status" -eq 0 ]

  [ -f ".agents/constraints/index.json" ]
  run jq -e '.constraints | length == 1' ".agents/constraints/index.json"
  [ "$status" -eq 0 ]
  run jq -r '.constraints[0].title' ".agents/constraints/index.json"
  [ "$status" -eq 0 ]
  [[ "$output" == *safe* ]]
}

@test "jq-less fallback preserves existing index and queues pending update" {
  cat > .agents-constraints-initial.json <<'EOF'
{"schema_version":1,"constraints":[{"id":"existing","title":"Existing","source":"old.md","status":"draft","compiled_at":"2026-01-01T00:00:00Z","file":".agents/constraints/existing.sh"}]}
EOF
  mkdir -p .agents/constraints
  cp .agents-constraints-initial.json .agents/constraints/index.json

  cat > learning.md <<'EOF'
---
id: learn-nojq
title: "No jq fallback"
date: 2026-02-24
tags: [constraint]
---
Fallback should queue updates without rewriting index.
EOF

  if PATH="/usr/bin:/bin" command -v jq >/dev/null 2>&1; then
    skip "Cannot simulate jq absence because jq is on /usr/bin:/bin"
  fi

  run env PATH="/usr/bin:/bin" bash "$SCRIPT" "$TMP_REPO/learning.md"
  [ "$status" -eq 0 ]

  run jq -r '.constraints[0].id' ".agents/constraints/index.json"
  [ "$status" -eq 0 ]
  [ "$output" = "existing" ]

  [ -f ".agents/constraints/index.pending.jsonl" ]
  run grep -q '"id":"learn-nojq"' ".agents/constraints/index.pending.jsonl"
  [ "$status" -eq 0 ]
}
