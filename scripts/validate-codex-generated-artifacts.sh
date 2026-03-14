#!/usr/bin/env bash
set -euo pipefail

ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
SCOPE="auto"
SKILLS_ROOT="$ROOT/skills-codex"
MANIFEST_FILE="$SKILLS_ROOT/.agentops-manifest.json"
MARKER_FILE_NAME=".agentops-generated.json"
MANIFEST_VALIDATOR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/validate-codex-generated-manifest.sh"
AUDIT_SCRIPT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/audit-codex-parity.sh"

usage() {
  cat <<'EOF'
Usage: bash scripts/validate-codex-generated-artifacts.sh [repo-root] [--scope auto|upstream|staged|worktree|head]
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --scope)
      SCOPE="${2:-}"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    --*)
      echo "Unknown arg: $1" >&2
      usage >&2
      exit 2
      ;;
    *)
      ROOT="$1"
      if [[ "$ROOT" != /* ]]; then
        ROOT="$(cd "$ROOT" && pwd)"
      fi
      SKILLS_ROOT="$ROOT/skills-codex"
      MANIFEST_FILE="$SKILLS_ROOT/.agentops-manifest.json"
      shift
      ;;
  esac
done

case "$SCOPE" in
  auto|upstream|staged|worktree|head) ;;
  *)
    echo "Invalid --scope: $SCOPE" >&2
    exit 2
    ;;
esac

failures=0

fail() {
  echo "FAIL: $1" >&2
  failures=$((failures + 1))
}

collect_changed_files() {
  local scope="$1"
  local ahead_files=""

  if ! git -C "$ROOT" rev-parse --git-dir >/dev/null 2>&1; then
    return 0
  fi

  case "$scope" in
    upstream)
      if git -C "$ROOT" rev-parse --abbrev-ref --symbolic-full-name '@{upstream}' >/dev/null 2>&1; then
        git -C "$ROOT" diff --name-only '@{upstream}...HEAD' 2>/dev/null || true
      fi
      ;;
    staged)
      git -C "$ROOT" diff --name-only --cached 2>/dev/null || true
      ;;
    worktree)
      git -C "$ROOT" diff --name-only --cached 2>/dev/null || true
      git -C "$ROOT" diff --name-only 2>/dev/null || true
      git -C "$ROOT" ls-files --others --exclude-standard 2>/dev/null || true
      ;;
    head)
      git -C "$ROOT" show --name-only --pretty=format: HEAD 2>/dev/null || true
      ;;
    auto)
      if git -C "$ROOT" rev-parse --abbrev-ref --symbolic-full-name '@{upstream}' >/dev/null 2>&1; then
        ahead_files="$(git -C "$ROOT" diff --name-only '@{upstream}...HEAD' 2>/dev/null || true)"
        if [[ -n "$ahead_files" ]]; then
          printf '%s\n' "$ahead_files"
          return 0
        fi
      fi
      git -C "$ROOT" diff --name-only --cached 2>/dev/null || true
      git -C "$ROOT" diff --name-only 2>/dev/null || true
      git -C "$ROOT" ls-files --others --exclude-standard 2>/dev/null || true
      git -C "$ROOT" show --name-only --pretty=format: HEAD 2>/dev/null || true
      ;;
  esac
}

echo "=== Codex generated artifact validation ==="

[[ -d "$SKILLS_ROOT" ]] || {
  echo "Missing skills-codex root: $SKILLS_ROOT" >&2
  exit 1
}
[[ -f "$MANIFEST_FILE" ]] || {
  echo "Missing generated manifest: $MANIFEST_FILE" >&2
  exit 1
}
if [[ -x "$MANIFEST_VALIDATOR" ]]; then
  bash "$MANIFEST_VALIDATOR" "$SKILLS_ROOT" >/dev/null
fi

while IFS= read -r skill_dir; do
  [[ -f "$skill_dir/SKILL.md" ]] || continue
  [[ -f "$skill_dir/$MARKER_FILE_NAME" ]] || fail "missing generated marker: ${skill_dir#"$ROOT"/}/$MARKER_FILE_NAME"
  if rg -q "^description:[[:space:]]*['\"]?[>|]['\"]?[[:space:]]*$" "$skill_dir/SKILL.md"; then
    fail "malformed generated description frontmatter: ${skill_dir#"$ROOT"/}/SKILL.md"
  fi
done < <(find "$SKILLS_ROOT" -mindepth 1 -maxdepth 1 -type d | LC_ALL=C sort)

mapfile -t changed_files < <(collect_changed_files "$SCOPE" | sed '/^[[:space:]]*$/d' | sort -u)

codex_changed=()
generator_changed=()
for file in "${changed_files[@]}"; do
  case "$file" in
    skills-codex/*)
      codex_changed+=("$file")
      ;;
    skills/*|skills-codex-overrides/*|scripts/sync-codex-native-skills.sh)
      generator_changed+=("$file")
      ;;
  esac
done

if [[ "${#codex_changed[@]}" -gt 0 && "${#generator_changed[@]}" -eq 0 ]]; then
  fail "skills-codex changed without matching source/generator edits; edit skills/ and regenerate instead"
fi

if [[ "${#generator_changed[@]}" -gt 0 && "${#codex_changed[@]}" -eq 0 ]]; then
  fail "source generator inputs changed without regenerating skills-codex; run scripts/sync-codex-native-skills.sh"
fi

audit_all_skills=0
declare -A audit_skills=()
for file in "${changed_files[@]}"; do
  case "$file" in
    skills/*/*)
      skill_name="${file#skills/}"
      skill_name="${skill_name%%/*}"
      audit_skills["$skill_name"]=1
      ;;
    skills-codex/*/*)
      skill_name="${file#skills-codex/}"
      skill_name="${skill_name%%/*}"
      audit_skills["$skill_name"]=1
      ;;
    skills-codex-overrides/*/*)
      skill_name="${file#skills-codex-overrides/}"
      skill_name="${skill_name%%/*}"
      audit_skills["$skill_name"]=1
      ;;
    skills-codex-overrides/catalog.json|skills/converter/*|scripts/sync-codex-native-skills.sh)
      audit_all_skills=1
      ;;
  esac
done

if [[ -x "$AUDIT_SCRIPT" && "${#generator_changed[@]}" -gt 0 ]]; then
  if [[ "$audit_all_skills" -eq 1 ]]; then
    if ! audit_output="$(bash "$AUDIT_SCRIPT" 2>&1)"; then
      fail "changed Codex generator inputs produced semantic parity drift"
      printf '%s\n' "$audit_output" >&2
    fi
  elif [[ "${#audit_skills[@]}" -gt 0 ]]; then
    audit_args=()
    for skill_name in "${!audit_skills[@]}"; do
      audit_args+=(--skill "$skill_name")
    done
    if ! audit_output="$(bash "$AUDIT_SCRIPT" "${audit_args[@]}" 2>&1)"; then
      fail "changed Codex skills produced semantic parity drift"
      printf '%s\n' "$audit_output" >&2
    fi
  fi
fi

if [[ "$failures" -gt 0 ]]; then
  echo "Codex generated artifact validation FAILED ($failures finding(s))." >&2
  exit 1
fi

echo "Codex generated artifact validation passed."
exit 0
