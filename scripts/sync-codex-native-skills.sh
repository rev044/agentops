#!/usr/bin/env bash
#
# DEPRECATED: skills-codex/ is now manually maintained.
# This script overwrites manual edits and should NOT be used.
# Edit skills-codex/ directly or use skills-codex-overrides/.
# See CLAUDE.md "Codex maintenance flow" for the current workflow.
#
echo "ERROR: sync-codex-native-skills.sh is DEPRECATED." >&2
echo "skills-codex/ is manually maintained. Edit files directly." >&2
echo "To audit for drift: bash scripts/audit-codex-parity.sh" >&2
exit 1
#
# Original script below (preserved for reference, never executes):
#
# Build Codex-native skill artifacts into a dedicated area (default: ./skills-codex).
# Output is generated from ./skills via the converter's codex target and then
# overlaid with codex-specific overrides from ./skills-codex-overrides.
# Source of truth for Codex behavior is therefore:
#   1) canonical skill contract in ./skills
#   2) Codex-native tailoring in ./skills-codex-overrides
#   3) generated artifact in ./skills-codex
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CONVERTER="$REPO_ROOT/skills/converter/scripts/convert.sh"
MANIFEST_FILE_NAME=".agentops-manifest.json"
SKILL_MARKER_FILE_NAME=".agentops-generated.json"
CATALOG_FILE_NAME="catalog.json"

SRC="$REPO_ROOT/skills"
OUT="$REPO_ROOT/skills-codex"
OVERRIDES="$REPO_ROOT/skills-codex-overrides"
ONLY_CSV=""
SKIP_OVERRIDES="false"

usage() {
  cat <<'EOF'
sync-codex-native-skills.sh

Builds Codex-native skill folders from source skills.

Options:
  --src <dir>         Source skills root (default: ./skills)
  --out <dir>         Output codex skills root (default: ./skills-codex)
  --overrides <dir>   Codex-only override layer (default: ./skills-codex-overrides)
  --skip-overrides    Do not apply override layer
  --only <a,b,c>      Only build these skill names (comma-separated)
  --help              Show this help

Examples:
  ./scripts/sync-codex-native-skills.sh
  ./scripts/sync-codex-native-skills.sh --only research,vibe
  ./scripts/sync-codex-native-skills.sh --overrides ./skills-codex-overrides
  ./scripts/sync-codex-native-skills.sh --out /tmp/skills-codex
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --src)
      SRC="${2:-}"
      shift 2
      ;;
    --out)
      OUT="${2:-}"
      shift 2
      ;;
    --overrides)
      OVERRIDES="${2:-}"
      shift 2
      ;;
    --skip-overrides)
      SKIP_OVERRIDES="true"
      shift 1
      ;;
    --only)
      ONLY_CSV="${2:-}"
      shift 2
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      echo "Unknown arg: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

if [[ "$SRC" != /* ]]; then
  SRC="$REPO_ROOT/$SRC"
fi
if [[ "$OUT" != /* ]]; then
  OUT="$REPO_ROOT/$OUT"
fi
if [[ "$OVERRIDES" != /* ]]; then
  OVERRIDES="$REPO_ROOT/$OVERRIDES"
fi

[[ -x "$CONVERTER" ]] || {
  echo "Error: converter script not executable: $CONVERTER" >&2
  exit 1
}
[[ -d "$SRC" ]] || {
  echo "Error: source skills directory not found: $SRC" >&2
  exit 1
}

tmpdir="$(mktemp -d)"
cleanup() {
  rm -rf "$tmpdir"
}
trap cleanup EXIT

sha256_file() {
  local path="$1"

  if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$path" | awk '{print $1}'
    return
  fi
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$path" | awk '{print $1}'
    return
  fi
  if command -v openssl >/dev/null 2>&1; then
    openssl dgst -sha256 "$path" | awk '{print $NF}'
    return
  fi

  echo "Error: need shasum, sha256sum, or openssl to compute skill manifests." >&2
  exit 1
}

find_hashable_files() {
  local root="$1"

  find "$root" -type f \
    ! -path '*/__pycache__/*' \
    ! -name '*.pyc' \
    ! -name '.DS_Store' \
    ! -name "$MANIFEST_FILE_NAME" \
    ! -name "$SKILL_MARKER_FILE_NAME" \
    | LC_ALL=C sort
}

hash_tree() {
  local root="$1"
  local manifest_tmp
  local rel
  local file

  manifest_tmp="$(mktemp)"
  while IFS= read -r file; do
    rel="${file#"$root"/}"
    printf '%s\t%s\n' "$rel" "$(sha256_file "$file")"
  done < <(find_hashable_files "$root") > "$manifest_tmp"

  sha256_file "$manifest_tmp"
  rm -f "$manifest_tmp"
}

prune_transient_files() {
  local root="$1"

  find "$root" \
    \( -type d -name '__pycache__' -o -type f \( -name '*.pyc' -o -name '.DS_Store' \) \) \
    -exec rm -rf {} +
}

write_expected_skills() {
  local out_file="$1"

  if [[ -n "$ONLY_CSV" ]]; then
    IFS=',' read -r -a only_arr <<<"$ONLY_CSV"
    {
      for skill in "${only_arr[@]}"; do
        skill="$(echo "$skill" | xargs)"
        [[ -n "$skill" ]] || continue
        skill_dir="$SRC/$skill"
        [[ -d "$skill_dir" ]] || {
          echo "Error: requested skill not found under src: $skill" >&2
          exit 1
        }
        [[ -f "$skill_dir/SKILL.md" ]] || {
          echo "Error: requested skill missing SKILL.md under src: $skill" >&2
          exit 1
        }
        echo "$skill"
      done
    } | sort -u > "$out_file"
    return
  fi

  find "$SRC" -mindepth 1 -maxdepth 1 -type d \
    | while IFS= read -r d; do
        [[ -f "$d/SKILL.md" ]] || continue
        basename "$d"
      done \
    | sort -u > "$out_file"
}

write_actual_skills() {
  local built_root="$1"
  local out_file="$2"
  find "$built_root" -mindepth 1 -maxdepth 1 -type d \
    | while IFS= read -r d; do
        [[ -f "$d/SKILL.md" ]] || continue
        basename "$d"
      done \
    | sort -u > "$out_file"
}

catalog_contract_entries() {
  local catalog_path="$OVERRIDES/$CATALOG_FILE_NAME"

  [[ -f "$catalog_path" ]] || return 0

  command -v jq >/dev/null 2>&1 || {
    echo "Error: jq is required for Codex operator-contract synthesis." >&2
    exit 1
  }

  if [[ -n "$ONLY_CSV" ]]; then
    local only_json
    only_json="$(
      printf '%s\n' "$ONLY_CSV" \
        | tr ',' '\n' \
        | sed 's/^[[:space:]]*//; s/[[:space:]]*$//' \
        | sed '/^$/d' \
        | jq -R . \
        | jq -s .
    )"
    jq -c --argjson only "$only_json" '
      .skills[]
      | select(.operator_contract_required == true and .operator_contract != null)
      | select(.name as $name | $only | index($name))
    ' "$catalog_path"
    return
  fi

  jq -c '
    .skills[]
    | select(.operator_contract_required == true and .operator_contract != null)
  ' "$catalog_path"
}

strip_generated_operator_contract_block() {
  local prompt_path="$1"
  local first_section="$2"
  local stripped_tmp trimmed_tmp

  stripped_tmp="$(mktemp "$tmpdir/operator-contract-strip.XXXXXX")"
  trimmed_tmp="$(mktemp "$tmpdir/operator-contract-trim.XXXXXX")"

  awk '
    $0 == "<!-- BEGIN AGENTOPS OPERATOR CONTRACT -->" { skip = 1; next }
    $0 == "<!-- END AGENTOPS OPERATOR CONTRACT -->" { skip = 0; next }
    !skip { print }
  ' "$prompt_path" > "$stripped_tmp"

  if grep -Fxq "$first_section" "$stripped_tmp"; then
    awk -v first_section="$first_section" '
      $0 == first_section { exit }
      { print }
    ' "$stripped_tmp" > "$trimmed_tmp"
    mv "$trimmed_tmp" "$stripped_tmp"
  fi

  awk '
    { lines[NR] = $0 }
    END {
      last = NR
      while (last > 0 && lines[last] == "") {
        last--
      }
      for (i = 1; i <= last; i++) {
        print lines[i]
      }
    }
  ' "$stripped_tmp" > "$prompt_path"

  rm -f "$stripped_tmp" "$trimmed_tmp"
}

render_operator_contract_block() {
  local entry="$1"
  local skill
  local marker_index=0
  local remaining_markers=0
  local section_count=0
  local section_index
  local sections=()
  local markers=()

  skill="$(jq -r '.name' <<<"$entry")"
  mapfile -t sections < <(jq -r '.operator_contract.required_sections[]' <<<"$entry")
  mapfile -t markers < <(jq -r '.operator_contract.required_markers[]' <<<"$entry")
  section_count="${#sections[@]}"
  remaining_markers="${#markers[@]}"

  printf '<!-- BEGIN AGENTOPS OPERATOR CONTRACT -->\n'
  printf '<!-- Generated from skills-codex-overrides/catalog.json for %s. -->\n\n' "$skill"

  for ((section_index = 0; section_index < section_count; section_index++)); do
    local sections_left count bullet_index

    printf '%s\n\n' "${sections[$section_index]}"
    sections_left=$((section_count - section_index))

    if (( remaining_markers == 0 )); then
      count=0
    elif (( sections_left == 1 )); then
      count=$remaining_markers
    else
      count=$((remaining_markers - (sections_left - 1)))
      if (( count < 1 )); then
        count=1
      fi
    fi

    for ((bullet_index = 0; bullet_index < count; bullet_index++)); do
      printf '%d. %s\n' "$((bullet_index + 1))" "${markers[$marker_index]}"
      marker_index=$((marker_index + 1))
      remaining_markers=$((remaining_markers - 1))
    done

    if (( section_index < section_count - 1 )); then
      printf '\n'
    fi
  done

  printf '\n<!-- END AGENTOPS OPERATOR CONTRACT -->\n'
}

apply_operator_contract_synthesis() {
  local built_root="$1"
  local synthesized=0

  while IFS= read -r entry; do
    local skill prompt_path first_section rendered_tmp final_tmp

    [[ -n "$entry" ]] || continue
    skill="$(jq -r '.name' <<<"$entry")"
    prompt_path="$built_root/$skill/prompt.md"
    first_section="$(jq -r '.operator_contract.required_sections[0]' <<<"$entry")"

    [[ -f "$prompt_path" ]] || {
      echo "Error: cannot synthesize operator contract for missing prompt: $prompt_path" >&2
      exit 1
    }

    strip_generated_operator_contract_block "$prompt_path" "$first_section"

    rendered_tmp="$(mktemp "$tmpdir/operator-contract-render.XXXXXX")"
    final_tmp="$(mktemp "$tmpdir/operator-contract-final.XXXXXX")"
    render_operator_contract_block "$entry" > "$rendered_tmp"

    if [[ -s "$prompt_path" ]]; then
      cat "$prompt_path" > "$final_tmp"
      printf '\n\n' >> "$final_tmp"
    fi
    cat "$rendered_tmp" >> "$final_tmp"
    mv "$final_tmp" "$prompt_path"
    rm -f "$rendered_tmp"

    synthesized=$((synthesized + 1))
  done < <(catalog_contract_entries)

  echo "Codex operator-contract synthesis applied: $synthesized"
}

apply_overrides() {
  local built_root="$1"
  local applied=0

  if [[ "$SKIP_OVERRIDES" == "true" ]]; then
    echo "Codex overrides skipped (--skip-overrides)."
    return
  fi

  if [[ ! -d "$OVERRIDES" ]]; then
    echo "Codex overrides directory not found; continuing without overrides: $OVERRIDES"
    return
  fi

  while IFS= read -r -d '' override_dir; do
    local skill
    skill="$(basename "$override_dir")"
    if [[ ! -d "$built_root/$skill" ]]; then
      if [[ -n "$ONLY_CSV" ]]; then
        continue
      fi
      echo "Error: override exists for unknown skill '$skill' (no generated output)." >&2
      exit 1
    fi
    rsync -a --copy-links "$override_dir"/ "$built_root/$skill"/
    applied=$((applied + 1))
  done < <(find "$OVERRIDES" -mindepth 1 -maxdepth 1 -type d -print0)

  echo "Codex overrides applied: $applied"
}

validate_parity() {
  local built_root="$1"
  local parity_dir="$tmpdir/.parity"
  local expected_file="$parity_dir/.expected-skills.txt"
  local actual_file="$parity_dir/.actual-skills.txt"
  local missing extra

  mkdir -p "$parity_dir"
  write_expected_skills "$expected_file"
  write_actual_skills "$built_root" "$actual_file"

  missing="$(comm -23 "$expected_file" "$actual_file" || true)"
  extra="$(comm -13 "$expected_file" "$actual_file" || true)"

  if [[ -n "$missing" || -n "$extra" ]]; then
    echo "Error: codex skill parity check failed." >&2
    if [[ -n "$missing" ]]; then
      echo "Missing in codex output:" >&2
      echo "$missing" >&2
    fi
    if [[ -n "$extra" ]]; then
      echo "Unexpected extras in codex output:" >&2
      echo "$extra" >&2
    fi
    exit 1
  fi

  local count
  count="$(wc -l < "$expected_file" | tr -d ' ')"
  echo "Codex skill parity check passed: $count skill(s)."
}

sync_output() {
  local built_root="$1"

  mkdir -p "$OUT"

  if [[ -n "$ONLY_CSV" ]]; then
    local updated=0
    while IFS= read -r -d '' skill_dir; do
      [[ -f "$skill_dir/SKILL.md" ]] || continue
      local skill
      skill="$(basename "$skill_dir")"
      mkdir -p "$OUT/$skill"
      rsync -a --delete --copy-links "$skill_dir"/ "$OUT/$skill"/
      updated=$((updated + 1))
    done < <(find "$built_root" -mindepth 1 -maxdepth 1 -type d -print0)

    # Remove legacy parity artifacts left by earlier script versions.
    rm -rf "$OUT/.parity"
    rm -f "$OUT/.expected-skills.txt" "$OUT/.actual-skills.txt"
    rm -f "$OUT/.agentops-generated.json"
    write_generated_manifest "$OUT"

    echo "Codex-native skills synced (partial): $updated"
    echo "Output: $OUT"
    return
  fi

  rsync -a --delete --copy-links \
    --exclude='.parity/' \
    --exclude='.expected-skills.txt' \
    --exclude='.actual-skills.txt' \
    "$built_root"/ "$OUT"/

  # Remove legacy parity artifacts left by earlier script versions.
  rm -rf "$OUT/.parity"
  rm -f "$OUT/.expected-skills.txt" "$OUT/.actual-skills.txt"
  rm -f "$OUT/.agentops-generated.json"

  local count
  count="$(find "$OUT" -mindepth 1 -maxdepth 1 -type d | wc -l | tr -d ' ')"
  echo "Codex-native skills synced: $count"
  echo "Output: $OUT"
}

write_generated_manifest() {
  local built_root="$1"
  local manifest_path="$built_root/$MANIFEST_FILE_NAME"
  local catalog_path="$OVERRIDES/$CATALOG_FILE_NAME"
  local skill_dirs=()
  local skill_dir
  local skill
  local source_hash
  local generated_hash
  local catalog_hash=""
  local catalog_json=""
  local first=1

  while IFS= read -r skill_dir; do
    [[ -f "$skill_dir/SKILL.md" ]] || continue
    skill_dirs+=("$skill_dir")
  done < <(find "$built_root" -mindepth 1 -maxdepth 1 -type d | LC_ALL=C sort)

  if [[ -f "$catalog_path" ]]; then
    command -v jq >/dev/null 2>&1 || {
      echo "Error: jq is required to embed $catalog_path into the generated Codex manifest." >&2
      exit 1
    }
    catalog_hash="$(sha256_file "$catalog_path")"
    catalog_json="$(jq -c '.' "$catalog_path")"
  fi

  {
    printf '{\n'
    printf '  "generator": "scripts/sync-codex-native-skills.sh",\n'
    printf '  "source_root": "skills",\n'
    printf '  "layout": "modular",\n'
    if [[ -n "$catalog_hash" ]]; then
      printf '  "codex_override_catalog_hash": "%s",\n' "$catalog_hash"
    fi
    if [[ -n "$catalog_json" ]]; then
      printf '  "codex_override_catalog": %s,\n' "$catalog_json"
    fi
    printf '  "skills": [\n'

    for skill_dir in "${skill_dirs[@]}"; do
      skill="$(basename "$skill_dir")"
      source_hash="$(hash_tree "$SRC/$skill")"
      generated_hash="$(hash_tree "$skill_dir")"

      cat > "$skill_dir/$SKILL_MARKER_FILE_NAME" <<EOF
{
  "generator": "scripts/sync-codex-native-skills.sh",
  "source_skill": "skills/$skill",
  "layout": "modular",
  "source_hash": "$source_hash",
  "generated_hash": "$generated_hash"
}
EOF

      if [[ "$first" -eq 0 ]]; then
        printf ',\n'
      fi
      first=0
      printf '    {\n'
      printf '      "name": "%s",\n' "$skill"
      printf '      "source_skill": "skills/%s",\n' "$skill"
      printf '      "source_hash": "%s",\n' "$source_hash"
      printf '      "generated_hash": "%s"\n' "$generated_hash"
      printf '    }'
    done

    printf '\n'
    printf '  ]\n'
    printf '}\n'
  } > "$manifest_path"
}

if [[ -n "$ONLY_CSV" ]]; then
  IFS=',' read -r -a only_arr <<<"$ONLY_CSV"
  for skill in "${only_arr[@]}"; do
    skill="$(echo "$skill" | xargs)"
    [[ -n "$skill" ]] || continue
    skill_dir="$SRC/$skill"
    [[ -d "$skill_dir" ]] || {
      echo "Error: requested skill not found under src: $skill" >&2
      exit 1
    }
    bash "$CONVERTER" --codex-layout modular "$skill_dir" codex "$tmpdir/$skill"
  done
else
  bash "$CONVERTER" --codex-layout modular --all codex "$tmpdir"
fi

# Pre-validate: check for unallowlisted residual markers
if [[ -x "${SCRIPT_DIR}/lint/generate-allowlist-candidates.sh" ]]; then
  echo "Checking for unallowlisted residual markers..."
  if ! bash "${SCRIPT_DIR}/lint/generate-allowlist-candidates.sh" "$tmpdir"; then
    echo "WARNING: Unallowlisted markers found. Add to allowlist or fix converter rules."
    # Non-blocking warning — validation gate will catch in CI
  fi
fi

apply_overrides "$tmpdir"
apply_operator_contract_synthesis "$tmpdir"
prune_transient_files "$tmpdir"
write_generated_manifest "$tmpdir"
validate_parity "$tmpdir"
sync_output "$tmpdir"
