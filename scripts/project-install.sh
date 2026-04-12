#!/usr/bin/env bash
#
# AgentOps Project Installer
# Installs AgentOps into a project with profile-specific configuration
# Supports both Claude Code projects and standard projects
#

set -euo pipefail

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Source libraries
if [[ -f "${SCRIPT_DIR}/lib/common-functions.sh" ]]; then
    # shellcheck source=./lib/common-functions.sh
    source "${SCRIPT_DIR}/lib/common-functions.sh"
elif [[ -f "${HOME}/.agentops/scripts/lib/common-functions.sh" ]]; then
    # shellcheck source=./lib/common-functions.sh
    source "${HOME}/.agentops/scripts/lib/common-functions.sh"
else
    echo "ERROR: Cannot find common-functions.sh library" >&2
    exit 1
fi

# Project root
PROJECT_ROOT="$(pwd)"

#######################################
# Print usage
#######################################
usage() {
    cat <<EOF
AgentOps Project Installer

Usage:
  $0 [OPTIONS] [PROFILE]

Arguments:
  PROFILE              Profile to install (default: auto-detect or prompt)

Options:
  --force              Overwrite existing installation
  --claude-code        Force Claude Code integration
  --no-claude          Skip Claude Code integration
  --help               Show this help message

Profiles:
  product-dev          Application development workflows
  infrastructure-ops   Operations and monitoring workflows
  devops               Complete GitOps ecosystem
  life                 Personal development and career planning

Examples:
  $0                          # Interactive profile selection
  $0 devops                   # Install devops profile
  $0 --force product-dev      # Force reinstall product-dev profile
  $0 --no-claude devops       # Install devops without Claude Code

EOF
}

#######################################
# Detect if this is a Claude Code project
# Returns:
#   0 if Claude Code project, 1 if not
#######################################
detect_claude_code() {
    [[ -d ".claude" ]] || [[ -f "claude.md" ]] || [[ -f "CLAUDE.md" ]]
}

#######################################
# Interactive profile selection
# Outputs:
#   Selected profile name
#######################################
select_profile_interactive() {
    echo ""
    info "AgentOps Profile Selection"
    echo ""

    # Get available profiles
    local profiles=()
    mapfile -t profiles < <(list_installed_profiles)

    if [[ ${#profiles[@]} -eq 0 ]]; then
        die "No profiles installed. Run: agentops install"
    fi

    echo "Available profiles:"
    echo ""

    local i=1
    for profile in "${profiles[@]}"; do
        local desc
        desc=$(get_profile_metadata "$profile" "description" || echo "No description")
        echo "  $i) $profile"
        echo "     $desc"
        echo ""
        ((i++))
    done

    echo "  $i) Cancel"
    echo ""

    local choice
    read -r -p "Select profile (1-$i): " choice

    if [[ "$choice" -ge 1 ]] && [[ "$choice" -lt "$i" ]]; then
        echo "${profiles[$((choice-1))]}"
        return 0
    elif [[ "$choice" -eq "$i" ]]; then
        echo ""
        return 1
    else
        warn "Invalid choice"
        select_profile_interactive
    fi
}

#######################################
# Install AgentOps into project
# Arguments:
#   $1: Profile name
#   $2: Force overwrite (optional)
#   $3: Claude Code mode (auto|force|skip)
#######################################
install_project() {
    local profile="$1"
    local force="${2:-false}"
    local claude_mode="${3:-auto}"

    echo ""
    echo "=========================================="
    echo "AgentOps Project Installation"
    echo "=========================================="
    echo ""

    info "Project: $PROJECT_ROOT"
    info "Profile: $profile"
    echo ""

    # Check if profile installed
    if ! is_profile_installed "$profile"; then
        die "Profile not installed: $profile. Run: agentops install --profile $profile"
    fi

    # Check for existing installation
    if [[ -d "${PROJECT_ROOT}/.agentops" ]] && [[ "$force" != "true" ]]; then
        warn "AgentOps already installed in this project"
        echo ""
        local overwrite
        read -r -p "Overwrite existing installation? (yes/no): " overwrite
        if [[ "$overwrite" != "yes" ]]; then
            info "Installation cancelled"
            return 0
        fi
        force="true"
    fi

    # Detect Claude Code project
    local is_claude=false
    if [[ "$claude_mode" == "force" ]]; then
        is_claude=true
    elif [[ "$claude_mode" == "skip" ]]; then
        is_claude=false
    elif detect_claude_code; then
        is_claude=true
        info "Claude Code project detected"
    fi

    # Create project AgentOps directory
    info "Creating project structure..."
    mkdir -p "${PROJECT_ROOT}/.agentops"

    # Create project config
    cat > "${PROJECT_ROOT}/.agentops/config.yml" <<EOF
# AgentOps Project Configuration
profile: $profile
version: v1.0.0

# Profile-specific settings
settings:
  # Add project-specific overrides here

# Command resolution
resolution:
  # Priority: explicit > env > project > user > base
  check_project_first: true
EOF

    success "Project config created: .agentops/config.yml"

    # Install for Claude Code if applicable
    if [[ "$is_claude" == "true" ]]; then
        info "Installing Claude Code integration..."

        # Create .claude directory if needed
        mkdir -p "${PROJECT_ROOT}/.claude"

        # Install layered commands
        mkdir -p "${PROJECT_ROOT}/.claude/commands"

        # Copy base commands
        if [[ -d "${AGENTOPS_HOME}/commands" ]]; then
            cp -r "${AGENTOPS_HOME}/commands/"* "${PROJECT_ROOT}/.claude/commands/" 2>/dev/null || true
        fi

        # Copy profile-specific commands (these override base)
        local profile_commands="${AGENTOPS_PROFILES_DIR}/${profile}/commands"
        if [[ -d "$profile_commands" ]]; then
            cp -r "${profile_commands}/"* "${PROJECT_ROOT}/.claude/commands/" 2>/dev/null || true
        fi

        success "Claude Code commands installed"

        # Install agents if profile has them
        mkdir -p "${PROJECT_ROOT}/.claude/agents"
        local profile_agents="${AGENTOPS_PROFILES_DIR}/${profile}/agents"
        if [[ -d "$profile_agents" ]]; then
            cp -r "${profile_agents}/"* "${PROJECT_ROOT}/.claude/agents/" 2>/dev/null || true
            success "Profile agents installed"
        fi

        # Create or update Claude settings
        if [[ ! -f "${PROJECT_ROOT}/.claude/settings.json" ]]; then
            cat > "${PROJECT_ROOT}/.claude/settings.json" <<EOFSETTINGS
{
  "agentops": {
    "profile": "$profile",
    "version": "v1.0.0"
  }
}
EOFSETTINGS
            success "Claude Code settings created"
        fi
    fi

    # Install git hooks if git repository
    if [[ -d "${PROJECT_ROOT}/.git" ]]; then
        info "Installing git hooks..."

        mkdir -p "${PROJECT_ROOT}/.git/hooks"

        # Copy core hooks if they exist
        if [[ -d "${INSTALL_SOURCE}/core/hooks" ]]; then
            cp -r "${INSTALL_SOURCE}/core/hooks/"* "${PROJECT_ROOT}/.git/hooks/" 2>/dev/null || true
            chmod +x "${PROJECT_ROOT}/.git/hooks/"* 2>/dev/null || true
            success "Git hooks installed"
        fi
    fi

    # Create README
    cat > "${PROJECT_ROOT}/.agentops/README.md" <<'EOFREADME'
# AgentOps Project Installation

This project uses AgentOps for AI-assisted development and operations.

## Configuration

**Profile:** See `.agentops/config.yml` for active profile

**Commands:** Available via `/command-name` (if using Claude Code)

**Agents:** Profile-specific agents in `.claude/agents/`

## Usage

### With Claude Code

```bash
# List available commands
ls .claude/commands/

# Use a command
/research
/plan
/implement
```

### Profile Information

```bash
# Check current profile
agentops current-profile

# Switch profile
agentops use-profile <profile-name>

# Profile info
agentops profile-info <profile-name>
```

## Resolution Chain

Commands and agents resolve in this order:

1. **Explicit:** `--profile` flag
2. **Environment:** `AGENTOPS_PROFILE` variable
3. **Project:** `.agentops/config.yml`
4. **User:** `~/.agentops/.profile`
5. **Base:** Default fallback

## Validation

```bash
# Validate installation
agentops validate

# Project-specific validation
make validate  # (if project has Makefile)
```

## Documentation

- **Profile docs:** `~/.agentops/profiles/<profile>/README.md`
- **AgentOps:** `~/.agentops/README.md`
- **Commands:** `.claude/commands/` (each command is documented)

EOFREADME

    success "Project README created"

    echo ""
    echo "=========================================="
    echo "Installation Complete"
    echo "=========================================="
    echo ""

    success "AgentOps installed with profile: $profile"

    if [[ "$is_claude" == "true" ]]; then
        echo ""
        info "Claude Code Integration:"
        echo "  Commands: .claude/commands/"
        echo "  Agents: .claude/agents/"
        echo "  Settings: .claude/settings.json"
    fi

    echo ""
    info "Next steps:"
    echo "  1. Review: .agentops/README.md"
    echo "  2. Configure: .agentops/config.yml (if needed)"

    if [[ "$is_claude" == "true" ]]; then
        echo "  3. Use commands: /research, /plan, /implement"
    else
        echo "  3. Check profile: agentops profile-info $profile"
    fi

    echo ""
}

#######################################
# Main entry point
#######################################
main() {
    local profile=""
    local force=false
    local claude_mode="auto"

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --force)
                force=true
                shift
                ;;
            --claude-code)
                claude_mode="force"
                shift
                ;;
            --no-claude)
                claude_mode="skip"
                shift
                ;;
            --help|-h)
                usage
                exit 0
                ;;
            -*)
                warn "Unknown option: $1"
                usage
                exit 1
                ;;
            *)
                profile="$1"
                shift
                ;;
        esac
    done

    # Interactive profile selection if not specified
    if [[ -z "$profile" ]]; then
        if profile=$(select_profile_interactive); then
            # Profile selected
            :
        else
            info "Installation cancelled"
            exit 0
        fi
    fi

    # Install
    install_project "$profile" "$force" "$claude_mode"
}

# Run main
main "$@"
