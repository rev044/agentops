#!/usr/bin/env bash
set -euo pipefail

# compile.sh — Pluggable LLM compilation engine for .agents/ → .agents/compiled/
# Usage: AGENTOPS_COMPILE_RUNTIME=ollama scripts/compile.sh [--sources DIR] [--output DIR] [--incremental] [--force] [--lint-only]

SOURCES_DIR=".agents"
OUTPUT_DIR=".agents/compiled"
INCREMENTAL=true
LINT_ONLY=false
HASH_FILE=""
RUNTIME="${AGENTOPS_COMPILE_RUNTIME:-}"
OLLAMA_MODEL="${AGENTOPS_COMPILE_MODEL:-gemma3:27b}"
CLAUDE_MODEL="${AGENTOPS_COMPILE_CLAUDE_MODEL:-claude-sonnet-4-20250514}"
OLLAMA_API="${OLLAMA_HOST:-http://localhost:11434}"

# --- Argument parsing ---
while [[ $# -gt 0 ]]; do
  case "$1" in
    --sources) SOURCES_DIR="$2"; shift 2 ;;
    --output) OUTPUT_DIR="$2"; shift 2 ;;
    --incremental) INCREMENTAL=true; shift ;;
    --force) INCREMENTAL=false; shift ;;
    --lint-only) LINT_ONLY=true; shift ;;
    --full) shift ;;  # default mode, accepted for SKILL.md parity
    *) echo "Unknown flag: $1" >&2; exit 1 ;;
  esac
done

HASH_FILE="$OUTPUT_DIR/.hashes.json"
mkdir -p "$OUTPUT_DIR"

# --- Utility functions ---

compute_hash() {
  if command -v md5sum &>/dev/null; then
    md5sum "$1" | cut -d' ' -f1
  else
    md5 -q "$1"
  fi
}

load_hashes() {
  if [[ -f "$HASH_FILE" ]]; then
    cat "$HASH_FILE"
  else
    echo "{}"
  fi
}

get_stored_hash() {
  local file="$1"
  local hashes="$2"
  echo "$hashes" | python3 -c "
import sys, json
data = json.load(sys.stdin)
print(data.get(sys.argv[1], ''))
" "$file" 2>/dev/null || echo ""
}

# --- LLM call abstraction ---

llm_call() {
  local system_prompt="$1"
  local user_prompt="$2"

  case "$RUNTIME" in
    ollama)
      local payload
      payload=$(python3 -c "
import json, sys
print(json.dumps({
    'model': '$OLLAMA_MODEL',
    'messages': [
        {'role': 'system', 'content': sys.argv[1]},
        {'role': 'user', 'content': sys.argv[2]}
    ],
    'stream': False
}))
" "$system_prompt" "$user_prompt")
      curl -sf "$OLLAMA_API/api/chat" \
        -H "Content-Type: application/json" \
        -d "$payload" | python3 -c "import sys,json; print(json.load(sys.stdin)['message']['content'])"
      ;;
    claude)
      local payload
      payload=$(python3 -c "
import json, sys
print(json.dumps({
    'model': '$CLAUDE_MODEL',
    'max_tokens': 4096,
    'system': sys.argv[1],
    'messages': [{'role': 'user', 'content': sys.argv[2]}]
}))
" "$system_prompt" "$user_prompt")
      curl -sf "https://api.anthropic.com/v1/messages" \
        -H "Content-Type: application/json" \
        -H "x-api-key: ${ANTHROPIC_API_KEY:?ANTHROPIC_API_KEY required for claude runtime}" \
        -H "anthropic-version: 2023-06-01" \
        -d "$payload" | python3 -c "import sys,json; print(json.load(sys.stdin)['content'][0]['text'])"
      ;;
    openai)
      local payload
      payload=$(python3 -c "
import json, sys
print(json.dumps({
    'model': '${AGENTOPS_COMPILE_OPENAI_MODEL:-gpt-4o}',
    'messages': [
        {'role': 'system', 'content': sys.argv[1]},
        {'role': 'user', 'content': sys.argv[2]}
    ]
}))
" "$system_prompt" "$user_prompt")
      curl -sf "${OPENAI_BASE_URL:-https://api.openai.com}/v1/chat/completions" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer ${OPENAI_API_KEY:?OPENAI_API_KEY required for openai runtime}" \
        -d "$payload" | python3 -c "import sys,json; print(json.load(sys.stdin)['choices'][0]['message']['content'])"
      ;;
    *)
      echo "ERROR: AGENTOPS_COMPILE_RUNTIME must be set to ollama, claude, or openai for headless compilation." >&2
      echo "For interactive compilation, invoke /compile directly in a Claude Code session." >&2
      exit 1
      ;;
  esac
}

# --- Phase: Inventory ---

inventory_sources() {
  local stored_hashes
  stored_hashes=$(load_hashes)
  local changed_files=()
  local unchanged_count=0

  local source_dirs=(
    "$SOURCES_DIR/learnings"
    "$SOURCES_DIR/patterns"
    "$SOURCES_DIR/research"
    "$SOURCES_DIR/retros"
    "$SOURCES_DIR/forge"
    "$SOURCES_DIR/knowledge"
  )

  for dir in "${source_dirs[@]}"; do
    [[ -d "$dir" ]] || continue
    while IFS= read -r -d '' file; do
      local current_hash
      current_hash=$(compute_hash "$file")
      local stored_hash
      stored_hash=$(get_stored_hash "$file" "$stored_hashes")

      if [[ "$INCREMENTAL" == "true" ]] && [[ "$current_hash" == "$stored_hash" ]]; then
        unchanged_count=$((unchanged_count + 1))
      else
        changed_files+=("$file")
      fi
    done < <(find "$dir" -type f -name "*.md" -print0)
  done

  echo "Changed: ${#changed_files[@]}, Unchanged: $unchanged_count" >&2

  # Output changed files, one per line
  for f in "${changed_files[@]}"; do
    echo "$f"
  done
}

# --- Phase: Compile ---

compile_articles() {
  local changed_files=()
  while IFS= read -r line; do
    [[ -n "$line" ]] && changed_files+=("$line")
  done

  if [[ ${#changed_files[@]} -eq 0 ]]; then
    echo "No changed files to compile." >&2
    return 0
  fi

  # Read all changed files into a single context
  local context=""
  for f in "${changed_files[@]}"; do
    context+="
--- FILE: $f ---
$(cat "$f")
"
  done

  local system_prompt="You are a knowledge compiler. You read raw knowledge artifacts (learnings, research, patterns, retros) and compile them into encyclopedia-style wiki articles.

Rules:
- Each article covers ONE topic/theme
- Use [[wikilinks]] for cross-references between articles
- Include a Sources section listing source files
- Include a Related section with [[links]] to related topics
- Write synthesis, not summaries — connect insights across sources
- Use YAML frontmatter with title, compiled date, sources list, and tags
- Article filenames should be kebab-case topic slugs (e.g., testing-strategy.md)"

  local user_prompt="Compile the following raw knowledge artifacts into wiki articles. Output each article separated by '=== ARTICLE: <filename> ===' markers.

$context

For each article output:
=== ARTICLE: <topic-slug>.md ===
<full article content with frontmatter>

After all articles, output:
=== INDEX ===
<index.md content cataloging all articles by category>"

  local result
  result=$(llm_call "$system_prompt" "$user_prompt")

  # Parse result into individual files
  local current_file=""
  local current_content=""

  while IFS= read -r line; do
    if [[ "$line" =~ ^===\ ARTICLE:\ (.+)\ ===$ ]]; then
      # Save previous article if exists
      if [[ -n "$current_file" ]] && [[ -n "$current_content" ]]; then
        echo "$current_content" > "$OUTPUT_DIR/$current_file"
        echo "Compiled: $current_file" >&2
      fi
      current_file=$(basename "${BASH_REMATCH[1]}")
      current_content=""
    elif [[ "$line" =~ ^===\ INDEX\ ===$ ]]; then
      # Save previous article
      if [[ -n "$current_file" ]] && [[ -n "$current_content" ]]; then
        echo "$current_content" > "$OUTPUT_DIR/$current_file"
        echo "Compiled: $current_file" >&2
      fi
      current_file="index.md"
      current_content=""
    else
      current_content+="$line
"
    fi
  done <<< "$result"

  # Save last article
  if [[ -n "$current_file" ]] && [[ -n "$current_content" ]]; then
    echo "$current_content" > "$OUTPUT_DIR/$current_file"
    echo "Compiled: $current_file" >&2
  fi

  # Update hashes
  save_hashes "${changed_files[@]}"

  # Append to log
  local article_count
  article_count=$(find "$OUTPUT_DIR" -name "*.md" ! -name "index.md" ! -name "log.md" ! -name "lint-report.md" | wc -l | tr -d ' ')
  local log_entry
  log_entry="## [$(date +%Y-%m-%d)] compile | $article_count articles from ${#changed_files[@]} sources
- Compiled from: ${changed_files[*]}
"
  echo "$log_entry" >> "$OUTPUT_DIR/log.md"
}

save_hashes() {
  local files=("$@")
  local json="{"
  local first=true

  for f in "${files[@]}"; do
    local h
    h=$(compute_hash "$f")
    if [[ "$first" == "true" ]]; then
      first=false
    else
      json+=","
    fi
    json+="\"$f\":\"$h\""
  done

  json+="}"

  # Merge with existing hashes
  if [[ -f "$HASH_FILE" ]]; then
    python3 -c "
import json, sys
hash_file = sys.argv[1]
new = json.loads(sys.argv[2])
with open(hash_file) as f:
    existing = json.load(f)
existing.update(new)
with open(hash_file, 'w') as f:
    json.dump(existing, f, indent=2)
" "$HASH_FILE" "$json"
  else
    echo "$json" | python3 -m json.tool > "$HASH_FILE"
  fi
}

# --- Phase: Lint ---

lint_wiki() {
  local articles=()
  while IFS= read -r -d '' f; do
    articles+=("$f")
  done < <(find "$OUTPUT_DIR" -name "*.md" ! -name "index.md" ! -name "log.md" ! -name "lint-report.md" ! -name ".hashes.json" -print0)

  if [[ ${#articles[@]} -eq 0 ]]; then
    echo "No articles to lint." >&2
    return 0
  fi

  local orphans=()
  local stale_claims=()

  for article in "${articles[@]}"; do
    local basename_article
    basename_article=$(basename "$article" .md)

    # Check for orphans: articles with no inbound [[wikilinks]]
    local inbound=0
    for other in "${articles[@]}"; do
      [[ "$other" == "$article" ]] && continue
      if grep -q "\[\[$basename_article\]\]" "$other" 2>/dev/null; then
        inbound=$((inbound + 1))
        break
      fi
    done
    if [[ $inbound -eq 0 ]]; then
      orphans+=("$basename_article")
    fi

    # Check for stale code references
    while IFS= read -r ref; do
      if [[ -n "$ref" ]] && [[ ! -e "$ref" ]]; then
        stale_claims+=("$basename_article references $ref")
      fi
    done < <(grep -oE '`[a-zA-Z0-9_./-]+\.(go|py|sh|ts|js|yaml|json|md)`' "$article" 2>/dev/null | tr -d '`' || true)
  done

  # Write lint report
  cat > "$OUTPUT_DIR/lint-report.md" << REPORT
# Wiki Lint Report — $(date +%Y-%m-%d)

## Orphan Pages: ${#orphans[@]}
$(for o in "${orphans[@]}"; do echo "- [[$o]]"; done)

## Stale Claims: ${#stale_claims[@]}
$(for s in "${stale_claims[@]}"; do echo "- $s"; done)

## Articles Scanned: ${#articles[@]}
REPORT

  echo "Lint complete: ${#orphans[@]} orphans, ${#stale_claims[@]} stale claims" >&2
}

# --- Main ---

main() {
  echo "=== Knowledge Compiler ===" >&2
  echo "Runtime: ${RUNTIME:-inline}" >&2
  echo "Sources: $SOURCES_DIR" >&2
  echo "Output: $OUTPUT_DIR" >&2
  echo "Incremental: $INCREMENTAL" >&2

  if [[ "$LINT_ONLY" == "true" ]]; then
    lint_wiki
    return 0
  fi

  # Inventory → Compile → Lint
  local changed_files
  changed_files=$(inventory_sources)

  if [[ -n "$changed_files" ]]; then
    echo "$changed_files" | compile_articles
  else
    echo "All source files unchanged — skipping compilation." >&2
  fi

  lint_wiki
  echo "=== Compilation complete ===" >&2
}

main
