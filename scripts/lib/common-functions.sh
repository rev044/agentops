#!/usr/bin/env bash
#
# AgentOps Common Functions Library
# Provides profile resolution, command routing, and utility functions
# Used by: install.sh, project-install.sh, agentops CLI
#

set -euo pipefail

# Colors for output
readonly COLOR_RED='\033[0;31m'
readonly COLOR_GREEN='\033[0;32m'
readonly COLOR_YELLOW='\033[1;33m'
readonly COLOR_BLUE='\033[0;34m'
readonly COLOR_RESET='\033[0m'

# AgentOps installation paths
readonly AGENTOPS_HOME="${HOME}/.agentops"
readonly AGENTOPS_PROFILES_DIR="${AGENTOPS_HOME}/profiles"
readonly AGENTOPS_CONFIG_FILE="${AGENTOPS_HOME}/.profile"
readonly AGENTOPS_BACKUP_DIR="${AGENTOPS_HOME}/backups"

# Project-level paths
readonly PROJECT_AGENTOPS_DIR=".agentops"
readonly PROJECT_CONFIG_FILE="${PROJECT_AGENTOPS_DIR}/config.yml"

#######################################
# Print colored message
# Arguments:
#   $1: Color (red, green, yellow, blue)
#   $2+: Message
#######################################
print_color() {
    local color="$1"
    shift
    local message="$*"

    case "$color" in
        red)    echo -e "${COLOR_RED}${message}${COLOR_RESET}" ;;
        green)  echo -e "${COLOR_GREEN}${message}${COLOR_RESET}" ;;
        yellow) echo -e "${COLOR_YELLOW}${message}${COLOR_RESET}" ;;
        blue)   echo -e "${COLOR_BLUE}${message}${COLOR_RESET}" ;;
        *)      echo "$message" ;;
    esac
}

#######################################
# Print error message and exit
# Arguments:
#   $1+: Error message
#######################################
die() {
    print_color red "ERROR: $*" >&2
    exit 1
}

#######################################
# Print success message
# Arguments:
#   $1+: Success message
#######################################
success() {
    print_color green "✓ $*"
}

#######################################
# Print warning message
# Arguments:
#   $1+: Warning message
#######################################
warn() {
    print_color yellow "⚠ $*"
}

#######################################
# Print info message
# Arguments:
#   $1+: Info message
#######################################
info() {
    print_color blue "ℹ $*"
}

#######################################
# Get active profile using resolution chain
# Resolution order (highest to lowest priority):
#   1. Explicit --profile flag (passed as $1)
#   2. Environment variable AGENTOPS_PROFILE
#   3. Project config (.agentops/config.yml)
#   4. User default (~/.agentops/.profile)
#   5. Base fallback (no profile)
# Arguments:
#   $1: Optional explicit profile from --profile flag
# Outputs:
#   Profile name or empty string for base
#######################################
get_active_profile() {
    local explicit_profile="${1:-}"

    # 1. Explicit --profile flag (highest priority)
    if [[ -n "$explicit_profile" ]]; then
        echo "$explicit_profile"
        return 0
    fi

    # 2. Environment variable
    if [[ -n "${AGENTOPS_PROFILE:-}" ]]; then
        echo "$AGENTOPS_PROFILE"
        return 0
    fi

    # 3. Project config
    if [[ -f "$PROJECT_CONFIG_FILE" ]]; then
        local project_profile
        project_profile=$(grep -E "^profile:" "$PROJECT_CONFIG_FILE" 2>/dev/null | cut -d: -f2 | tr -d ' ' || echo "")
        if [[ -n "$project_profile" ]]; then
            echo "$project_profile"
            return 0
        fi
    fi

    # 4. User default
    if [[ -f "$AGENTOPS_CONFIG_FILE" ]]; then
        cat "$AGENTOPS_CONFIG_FILE"
        return 0
    fi

    # 5. Base fallback (empty string = base profile)
    echo ""
}

#######################################
# Set user default profile
# Arguments:
#   $1: Profile name
#######################################
set_default_profile() {
    local profile="$1"

    if [[ ! -d "$AGENTOPS_HOME" ]]; then
        mkdir -p "$AGENTOPS_HOME"
    fi

    echo "$profile" > "$AGENTOPS_CONFIG_FILE"
    success "Default profile set to: $profile"
}

#######################################
# List installed profiles
# Outputs:
#   List of profile names, one per line
#######################################
list_installed_profiles() {
    if [[ ! -d "$AGENTOPS_PROFILES_DIR" ]]; then
        return 0
    fi

    find "$AGENTOPS_PROFILES_DIR" -mindepth 1 -maxdepth 1 -type d -exec basename {} \;
}

#######################################
# Check if profile is installed
# Arguments:
#   $1: Profile name
# Returns:
#   0 if installed, 1 if not
#######################################
is_profile_installed() {
    local profile="$1"
    local profile_dir="${AGENTOPS_PROFILES_DIR}/${profile}"

    [[ -d "$profile_dir" ]]
}

#######################################
# Resolve command path using layered resolution
# Arguments:
#   $1: Command name (e.g., "research")
#   $2: Active profile (optional, will auto-detect if not provided)
# Outputs:
#   Path to command file
# Returns:
#   0 if found, 1 if not found
#######################################
resolve_command() {
    local command_name="$1"
    local profile="${2:-$(get_active_profile)}"

    # If profile specified, check profile-specific command first
    if [[ -n "$profile" ]]; then
        local profile_command="${AGENTOPS_PROFILES_DIR}/${profile}/commands/${command_name}.md"
        if [[ -f "$profile_command" ]]; then
            echo "$profile_command"
            return 0
        fi
    fi

    # Fallback to base command
    local base_command="${AGENTOPS_HOME}/commands/${command_name}.md"
    if [[ -f "$base_command" ]]; then
        echo "$base_command"
        return 0
    fi

    # Command not found
    return 1
}

#######################################
# Resolve agent path using layered resolution
# Arguments:
#   $1: Agent name
#   $2: Active profile (optional)
# Outputs:
#   Path to agent file
# Returns:
#   0 if found, 1 if not found
#######################################
resolve_agent() {
    local agent_name="$1"
    local profile="${2:-$(get_active_profile)}"

    # If profile specified, check profile-specific agent first
    if [[ -n "$profile" ]]; then
        local profile_agent="${AGENTOPS_PROFILES_DIR}/${profile}/agents/${agent_name}.md"
        if [[ -f "$profile_agent" ]]; then
            echo "$profile_agent"
            return 0
        fi
    fi

    # Fallback to base agents
    local base_agent="${AGENTOPS_HOME}/agents/${agent_name}.md"
    if [[ -f "$base_agent" ]]; then
        echo "$base_agent"
        return 0
    fi

    # Agent not found
    return 1
}

#######################################
# Create backup of directory
# Arguments:
#   $1: Source directory
#   $2: Backup name (optional, defaults to timestamp)
# Returns:
#   0 on success, 1 on failure
#######################################
create_backup() {
    local source_dir="$1"
    local backup_name="${2:-$(date +%Y%m%d_%H%M%S)}"

    if [[ ! -d "$source_dir" ]]; then
        warn "Source directory does not exist: $source_dir"
        return 1
    fi

    mkdir -p "$AGENTOPS_BACKUP_DIR"

    local backup_path="${AGENTOPS_BACKUP_DIR}/${backup_name}"

    if cp -r "$source_dir" "$backup_path"; then
        success "Backup created: $backup_path"
        return 0
    else
        die "Failed to create backup"
    fi
}

#######################################
# Restore from backup
# Arguments:
#   $1: Backup name or path
#   $2: Destination directory
# Returns:
#   0 on success, 1 on failure
#######################################
restore_backup() {
    local backup_name="$1"
    local dest_dir="$2"

    # Check if it's a full path or just a name
    local backup_path
    if [[ -d "$backup_name" ]]; then
        backup_path="$backup_name"
    else
        backup_path="${AGENTOPS_BACKUP_DIR}/${backup_name}"
    fi

    if [[ ! -d "$backup_path" ]]; then
        die "Backup not found: $backup_path"
    fi

    # Remove existing destination if it exists
    if [[ -d "$dest_dir" ]]; then
        rm -rf "$dest_dir"
    fi

    if cp -r "$backup_path" "$dest_dir"; then
        success "Restored from backup: $backup_path"
        return 0
    else
        die "Failed to restore from backup"
    fi
}

#######################################
# Check if running in Claude Code project
# Returns:
#   0 if Claude Code project, 1 if not
#######################################
is_claude_code_project() {
    [[ -d ".claude" ]] || [[ -f "claude.md" ]] || [[ -f "CLAUDE.md" ]]
}

#######################################
# Get profile metadata from manifest
# Arguments:
#   $1: Profile name
#   $2: Field name (name, version, agent_count, etc.)
# Outputs:
#   Field value
#######################################
get_profile_metadata() {
    local profile="$1"
    local field="$2"
    local manifest="${AGENTOPS_PROFILES_DIR}/${profile}/profile.yaml"

    if [[ ! -f "$manifest" ]]; then
        return 1
    fi

    # Simple YAML parsing (assumes well-formed YAML)
    grep -E "^  ${field}:" "$manifest" | cut -d: -f2- | sed 's/^ *//' | sed 's/ *$//'
}

#######################################
# Print profile information
# Arguments:
#   $1: Profile name
#######################################
print_profile_info() {
    local profile="$1"

    if ! is_profile_installed "$profile"; then
        warn "Profile not installed: $profile"
        return 1
    fi

    local version agent_count description
    version=$(get_profile_metadata "$profile" "version" || echo "unknown")
    agent_count=$(get_profile_metadata "$profile" "agent_count" || echo "unknown")
    description=$(get_profile_metadata "$profile" "description" || echo "No description")

    info "Profile: $profile"
    echo "  Version: $version"
    echo "  Agents: $agent_count"
    echo "  Description: $description"
}

#######################################
# Initialize AgentOps directory structure
# Creates necessary directories if they don't exist
#######################################
init_agentops_dirs() {
    mkdir -p "$AGENTOPS_HOME"
    mkdir -p "$AGENTOPS_PROFILES_DIR"
    mkdir -p "$AGENTOPS_BACKUP_DIR"
    mkdir -p "${AGENTOPS_HOME}/commands"
    mkdir -p "${AGENTOPS_HOME}/agents"
}

# Export functions for use in other scripts
export -f print_color
export -f die
export -f success
export -f warn
export -f info
export -f get_active_profile
export -f set_default_profile
export -f list_installed_profiles
export -f is_profile_installed
export -f resolve_command
export -f resolve_agent
export -f create_backup
export -f restore_backup
export -f is_claude_code_project
export -f get_profile_metadata
export -f print_profile_info
export -f init_agentops_dirs
