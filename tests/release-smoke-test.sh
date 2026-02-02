#!/usr/bin/env bash
# Release smoke test - verify all agents and skills are loadable
# Usage: ./tests/release-smoke-test.sh [--full]
#
# Default: Fast verification (~30s) - checks components are registered
# --full:  Slow verification (~15min) - invokes each component individually

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

source "$SCRIPT_DIR/claude-code/test-helpers.sh"

# Logging (redefine to avoid conflict with macOS log command)
log() { echo -e "${BLUE}[TEST]${NC} $1"; }
pass() { echo -e "${GREEN}  ✓${NC} $1"; }
fail() { echo -e "${RED}  ✗${NC} $1"; }

# Expected counts
EXPECTED_AGENTS=20
EXPECTED_SKILLS=21

# Parse args
FULL_TEST=false
[[ "${1:-}" == "--full" ]] && FULL_TEST=true
[[ "${1:-}" == "--help" ]] && { echo "Usage: $0 [--full]"; echo "  --full  Run slow individual tests (~15min)"; exit 0; }

echo ""
echo "═══════════════════════════════════════════"
echo "     AgentOps Release Smoke Test"
echo "═══════════════════════════════════════════"
echo ""

if $FULL_TEST; then
    # =========================================================================
    # FULL TEST: Individual invocation of each component
    # =========================================================================
    log "Running FULL test (individual invocations)..."

    AGENTS=(assumption-challenger coverage-expert data-failure-expert depth-expert
            edge-case-hunter flywheel-feeder gap-identifier goal-achievement-expert
            integration-failure-expert ops-failure-expert plan-compliance-expert
            process-learnings-expert ratchet-validator technical-learnings-expert
            architecture-expert code-quality-expert code-reviewer security-expert
            security-reviewer ux-expert)

    SKILLS=(beads bug-hunt complexity crank doc extract flywheel forge implement
            inject knowledge plan post-mortem pre-mortem provenance ratchet
            research retro standards using-agentops vibe)

    passed=0
    failed=0

    for agent in "${AGENTS[@]}"; do
        if timeout 60 claude -p "Invoke agentops:$agent agent to analyze README.md briefly" \
            --plugin-dir "$REPO_ROOT" --dangerously-skip-permissions --max-turns 3 >/dev/null 2>&1; then
            pass "$agent"
            ((passed++))
        else
            fail "$agent"
            ((failed++))
        fi
    done

    for skill in "${SKILLS[@]}"; do
        if timeout 45 claude -p "Invoke agentops:$skill skill" \
            --plugin-dir "$REPO_ROOT" --dangerously-skip-permissions --max-turns 3 >/dev/null 2>&1; then
            pass "$skill"
            ((passed++))
        else
            fail "$skill"
            ((failed++))
        fi
    done

    print_summary "$passed" "$failed" 0
    exit $((failed > 0))
fi

# =============================================================================
# FAST TEST: Single prompt to verify all components are registered
# =============================================================================
log "Running FAST test (registration check)..."
echo ""

# Create a prompt that asks Claude to list available agentops agents and skills
PROMPT='List all available agentops agents and skills. Format your response as:

AGENTS: [comma-separated list]
SKILLS: [comma-separated list]
COUNTS: agents=N, skills=M

Only list agentops: prefixed items. Be thorough - check the Task tool for agents and Skill tool for skills.'

log "Querying Claude for registered components..."

output=$(timeout 120 claude -p "$PROMPT" \
    --plugin-dir "$REPO_ROOT" \
    --dangerously-skip-permissions \
    --max-turns 5 2>&1) || {
    fail "Claude query failed"
    exit 1
}

# Parse the output
echo "$output"
echo ""

# Extract counts from output
agent_count=$(echo "$output" | grep -oE 'agents?[=:] ?[0-9]+' | grep -oE '[0-9]+' | head -1 || echo "0")
skill_count=$(echo "$output" | grep -oE 'skills?[=:] ?[0-9]+' | grep -oE '[0-9]+' | head -1 || echo "0")

# Fallback: count comma-separated items if explicit count not found
if [[ -z "$agent_count" ]] || [[ "$agent_count" == "0" ]]; then
    agents_line=$(echo "$output" | grep -i "^AGENTS:" | head -1)
    if [[ -n "$agents_line" ]]; then
        agent_count=$(echo "$agents_line" | tr ',' '\n' | wc -l | tr -d ' ')
    fi
fi

if [[ -z "$skill_count" ]] || [[ "$skill_count" == "0" ]]; then
    skills_line=$(echo "$output" | grep -i "^SKILLS:" | head -1)
    if [[ -n "$skills_line" ]]; then
        skill_count=$(echo "$skills_line" | tr ',' '\n' | wc -l | tr -d ' ')
    fi
fi

echo ""
echo -e "${BLUE}═══════════════════════════════════════════${NC}"
echo "Release Smoke Test Results"
echo -e "${BLUE}───────────────────────────────────────────${NC}"

passed=0
failed=0

# Check agents
if [[ "$agent_count" -ge "$EXPECTED_AGENTS" ]]; then
    pass "Agents: $agent_count found (expected $EXPECTED_AGENTS)"
    ((passed++)) || true
else
    fail "Agents: $agent_count found (expected $EXPECTED_AGENTS)"
    ((failed++)) || true
fi

# Check skills
if [[ "$skill_count" -ge "$EXPECTED_SKILLS" ]]; then
    pass "Skills: $skill_count found (expected $EXPECTED_SKILLS)"
    ((passed++)) || true
else
    fail "Skills: $skill_count found (expected $EXPECTED_SKILLS)"
    ((failed++)) || true
fi

echo -e "${BLUE}───────────────────────────────────────────${NC}"
echo -e "  Total:  ${GREEN}$passed passed${NC}, ${RED}$failed failed${NC}"
echo -e "${BLUE}═══════════════════════════════════════════${NC}"

if [[ $failed -gt 0 ]]; then
    echo ""
    echo -e "${RED}RELEASE BLOCKED${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}RELEASE READY: All components registered${NC}"
exit 0
