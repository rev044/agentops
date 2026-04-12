#!/usr/bin/env bash
#
# AgentOps v1.0.0 Production Installer
# Multi-profile installation with layered command resolution
#
# Usage:
#   ./install.sh                    # Interactive mode
#   ./install.sh --profile devops   # Install specific profile
#   ./install.sh --all              # Install all profiles
#   ./install.sh --uninstall        # Remove installation
#

set -euo pipefail

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Source libraries
# shellcheck source=./lib/common-functions.sh
source "${SCRIPT_DIR}/lib/common-functions.sh"
# shellcheck source=./lib/logging.sh
source "${SCRIPT_DIR}/lib/logging.sh"
# shellcheck source=./lib/validation.sh
source "${SCRIPT_DIR}/lib/validation.sh"

# Installation source (where this script lives)
INSTALL_SOURCE="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Available profiles
# Note: "example" is a template, not meant for actual use
readonly AVAILABLE_PROFILES=("example" "devops" "base")

#######################################
# Print usage information
#######################################
usage() {
    cat <<EOF
AgentOps v1.0.0 Installer

Usage:
  $0 [OPTIONS]

Options:
  --profile NAME    Install specific profile (example, devops, base, or custom)
  --all             Install all available profiles
  --uninstall       Remove AgentOps installation
  --upgrade         Upgrade existing installation
  --validate-only   Validate installation without making changes
  --help            Show this help message

Core + Extensibility:
  Core is always installed (universal commands, agents, workflows, skills)
  Profiles extend core with domain-specific capabilities
  Create custom profiles using: docs/CREATE_PROFILE.md

Interactive Mode:
  Run without options for interactive profile selection menu

Examples:
  $0                           # Interactive mode
  $0 --profile devops          # Install devops profile
  $0 --all                     # Install all profiles
  $0 --uninstall               # Remove installation
  $0 --validate-only           # Check existing installation

EOF
}

#######################################
# Interactive profile selection menu
# Outputs:
#   Space-separated list of selected profiles
#######################################
select_profiles() {
    echo ""
    info "AgentOps Profile Selection"
    echo ""
    echo "Available profiles:"
    echo ""
    echo "  1) example              (Template)   - Copy this to create custom profiles"
    echo "  2) devops               (Agents TBD) - DevOps workflows (K8s, containers, CI/CD)"
    echo "  3) base                 (Minimal)    - Base profile with core only"
    echo "  4) All profiles         (All)        - Install everything"
    echo "  5) Cancel installation"
    echo ""
    echo "Note: Core is always installed (commands, agents, workflows, skills)"
    echo "      Profiles extend core with domain-specific capabilities"
    echo ""

    local choice
    read -r -p "Select profile (1-5): " choice

    case "$choice" in
        1) echo "example" ;;
        2) echo "devops" ;;
        3) echo "base" ;;
        4) echo "${AVAILABLE_PROFILES[*]}" ;;
        5) echo ""; return 1 ;;
        *) warn "Invalid choice"; select_profiles ;;
    esac
}

#######################################
# Check prerequisites
# Returns:
#   0 if all prerequisites met, 1 otherwise
#######################################
check_prerequisites() {
    local errors=0

    info "Checking prerequisites..."

    # Check for required commands
    local required_commands=("git" "bash" "grep" "sed" "find")
    for cmd in "${required_commands[@]}"; do
        if ! command -v "$cmd" &> /dev/null; then
            warn "Required command not found: $cmd"
            ((errors++))
        fi
    done

    # Check bash version (need 4.0+)
    if [[ "${BASH_VERSINFO[0]}" -lt 4 ]]; then
        warn "Bash 4.0+ required (found ${BASH_VERSION})"
        ((errors++))
    fi

    if [[ $errors -eq 0 ]]; then
        success "All prerequisites met"
        return 0
    else
        die "Prerequisites check failed ($errors errors)"
    fi
}

#######################################
# Install a single profile
# Arguments:
#   $1: Profile name
# Returns:
#   0 on success, 1 on failure
#######################################
install_profile() {
    local profile="$1"
    local profile_source="${INSTALL_SOURCE}/profiles/${profile}"
    local profile_dest="${AGENTOPS_PROFILES_DIR}/${profile}"

    info "Installing profile: $profile"
    log_profile "install" "$profile" "Starting profile installation"

    # Validate profile source exists
    if [[ ! -d "$profile_source" ]]; then
        warn "Profile source not found: $profile_source"
        log_profile "install" "$profile" "Profile source not found" "status=failed"
        return 1
    fi

    # Create profile destination
    mkdir -p "$profile_dest"

    # Copy profile files
    local start_time
    start_time=$(start_timer)

    cp -r "${profile_source}"/* "${profile_dest}/"

    end_timer "$start_time" "install_profile_${profile}"

    # Verify installation
    if [[ -d "${profile_dest}/agents" ]] && [[ -d "${profile_dest}/commands" ]]; then
        success "Profile installed: $profile"
        log_profile "install" "$profile" "Profile installed successfully" "status=success"
        return 0
    else
        warn "Profile installation incomplete: $profile"
        log_profile "install" "$profile" "Profile installation incomplete" "status=failed"
        return 1
    fi
}

#######################################
# Install base commands and agents
# Returns:
#   0 on success, 1 on failure
#######################################
install_base() {
    info "Installing base commands and agents..."
    log_installation "progress" "Installing base components"

    # Create base directories
    mkdir -p "${AGENTOPS_HOME}/commands"
    mkdir -p "${AGENTOPS_HOME}/agents"

    # Copy core directory if it exists (base commands/agents)
    if [[ -d "${INSTALL_SOURCE}/core" ]]; then
        if [[ -d "${INSTALL_SOURCE}/core/commands" ]]; then
            cp -r "${INSTALL_SOURCE}/core/commands/"* "${AGENTOPS_HOME}/commands/" 2>/dev/null || true
        fi
        if [[ -d "${INSTALL_SOURCE}/core/agents" ]]; then
            cp -r "${INSTALL_SOURCE}/core/agents/"* "${AGENTOPS_HOME}/agents/" 2>/dev/null || true
        fi
    fi

    success "Base components installed"
    return 0
}

#######################################
# Install scripts and libraries
# Returns:
#   0 on success, 1 on failure
#######################################
install_scripts() {
    info "Installing scripts and libraries..."
    log_installation "progress" "Installing scripts"

    # Create scripts directory
    mkdir -p "${AGENTOPS_HOME}/scripts"
    mkdir -p "${AGENTOPS_HOME}/scripts/lib"

    # Copy scripts
    cp "${SCRIPT_DIR}/install.sh" "${AGENTOPS_HOME}/scripts/"
    cp "${SCRIPT_DIR}/project-install.sh" "${AGENTOPS_HOME}/scripts/"
    cp "${SCRIPT_DIR}/lib/"* "${AGENTOPS_HOME}/scripts/lib/"

    # Make scripts executable
    chmod +x "${AGENTOPS_HOME}/scripts/install.sh"
    chmod +x "${AGENTOPS_HOME}/scripts/project-install.sh"

    success "Scripts installed"
    return 0
}

#######################################
# Install CLI tool
# Returns:
#   0 on success, 1 on failure
#######################################
install_cli() {
    info "Installing CLI tool..."

    # Create bin directory
    mkdir -p "${AGENTOPS_HOME}/bin"

    # Create agentops CLI wrapper
    cat > "${AGENTOPS_HOME}/bin/agentops" <<'EOFCLI'
#!/usr/bin/env bash
# AgentOps CLI Tool
set -euo pipefail

AGENTOPS_HOME="${HOME}/.agentops"
source "${AGENTOPS_HOME}/scripts/lib/common-functions.sh"

case "${1:-help}" in
    use-profile|switch)
        set_default_profile "${2:-}"
        ;;
    current-profile|current)
        get_active_profile
        ;;
    list-profiles|list)
        list_installed_profiles
        ;;
    profile-info|info)
        print_profile_info "${2:-}"
        ;;
    validate)
        source "${AGENTOPS_HOME}/scripts/lib/validation.sh"
        validate_all "${2:-}"
        ;;
    version|--version|-v)
        echo "AgentOps v1.0.0"
        ;;
    help|--help|-h|*)
        cat <<EOF
AgentOps CLI Tool v1.0.0

Usage:
  agentops <command> [arguments]

Commands:
  use-profile <name>    Set default profile
  current-profile       Show active profile
  list-profiles         List installed profiles
  profile-info <name>   Show profile information
  validate [profile]    Validate installation
  version               Show version
  help                  Show this help

Examples:
  agentops use-profile devops
  agentops current-profile
  agentops list-profiles
  agentops validate
EOF
        ;;
esac
EOFCLI

    chmod +x "${AGENTOPS_HOME}/bin/agentops"

    success "CLI tool installed"

    # Check if bin is in PATH
    if [[ ":$PATH:" != *":${AGENTOPS_HOME}/bin:"* ]]; then
        echo ""
        warn "Add ${AGENTOPS_HOME}/bin to your PATH to use 'agentops' command:"
        echo ""
        echo "  export PATH=\"\${HOME}/.agentops/bin:\$PATH\""
        echo ""
        echo "Add this to your ~/.bashrc, ~/.zshrc, or ~/.profile"
        echo ""
    fi

    return 0
}

#######################################
# Main installation function
# Arguments:
#   $@: List of profiles to install
# Returns:
#   0 on success, 1 on failure
#######################################
main_install() {
    local profiles=("$@")

    echo ""
    echo "=========================================="
    echo "AgentOps v1.0.0 Installation"
    echo "=========================================="
    echo ""

    log_installation "start" "Beginning AgentOps installation"

    local overall_start
    overall_start=$(start_timer)

    # Check prerequisites
    check_prerequisites

    # Initialize directory structure
    info "Initializing AgentOps directories..."
    init_agentops_dirs

    # Create backup if existing installation
    if [[ -d "$AGENTOPS_HOME" ]] && [[ -n "$(ls -A "$AGENTOPS_HOME" 2>/dev/null)" ]]; then
        info "Existing installation detected"
        local backup_name="backup_$(date +%Y%m%d_%H%M%S)"
        create_backup "$AGENTOPS_HOME" "$backup_name"
    fi

    # Install base components
    install_base

    # Install scripts
    install_scripts

    # Install CLI tool
    install_cli

    # Install selected profiles
    local profile_count=0
    local failed_profiles=()

    for profile in "${profiles[@]}"; do
        if install_profile "$profile"; then
            ((profile_count++))
        else
            failed_profiles+=("$profile")
        fi
    done

    echo ""
    echo "=========================================="
    echo "Installation Summary"
    echo "=========================================="
    echo ""

    success "Profiles installed: $profile_count"

    if [[ ${#failed_profiles[@]} -gt 0 ]]; then
        warn "Failed profiles: ${failed_profiles[*]}"
    fi

    # Set default profile to first successfully installed
    if [[ $profile_count -gt 0 ]]; then
        local first_profile="${profiles[0]}"
        set_default_profile "$first_profile"
    fi

    end_timer "$overall_start" "full_installation"

    echo ""
    info "Running post-installation validation..."
    echo ""

    # Run validation
    if validate_all; then
        echo ""
        success "✓ Installation completed successfully!"
        log_installation "complete" "Installation completed successfully" "profiles=$profile_count"
        echo ""
        info "Next steps:"
        echo "  1. Add ${AGENTOPS_HOME}/bin to your PATH"
        echo "  2. Run: agentops validate"
        echo "  3. Try: agentops current-profile"
        echo "  4. See: ${AGENTOPS_HOME}/README.md"
        echo ""
        return 0
    else
        warn "Installation completed with validation warnings"
        log_installation "complete" "Installation completed with warnings" "profiles=$profile_count"
        return 1
    fi
}

#######################################
# Uninstall AgentOps
#######################################
uninstall() {
    echo ""
    warn "This will remove all AgentOps files from ${AGENTOPS_HOME}"
    echo ""

    local confirmation
    read -r -p "Are you sure? (yes/no): " confirmation

    if [[ "$confirmation" != "yes" ]]; then
        info "Uninstall cancelled"
        return 0
    fi

    log_installation "start" "Uninstalling AgentOps"

    # Create final backup
    if [[ -d "$AGENTOPS_HOME" ]]; then
        create_backup "$AGENTOPS_HOME" "final_backup_$(date +%Y%m%d_%H%M%S)"

        # Remove installation
        rm -rf "$AGENTOPS_HOME"

        success "AgentOps uninstalled"
        log_installation "complete" "Uninstallation complete"

        info "Backup saved to: ${AGENTOPS_BACKUP_DIR}"
    else
        info "No installation found"
    fi

    return 0
}

#######################################
# Main script entry point
#######################################
main() {
    local profiles_to_install=()
    local mode="interactive"

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --profile)
                mode="specific"
                profiles_to_install+=("$2")
                shift 2
                ;;
            --all)
                mode="all"
                profiles_to_install=("${AVAILABLE_PROFILES[@]}")
                shift
                ;;
            --uninstall)
                uninstall
                exit 0
                ;;
            --validate-only)
                source "${SCRIPT_DIR}/lib/validation.sh"
                validate_all
                exit $?
                ;;
            --help|-h)
                usage
                exit 0
                ;;
            *)
                warn "Unknown option: $1"
                usage
                exit 1
                ;;
        esac
    done

    # Interactive mode if no profiles specified
    if [[ $mode == "interactive" ]]; then
        if profiles_selected=$(select_profiles); then
            read -ra profiles_to_install <<< "$profiles_selected"
        else
            info "Installation cancelled"
            exit 0
        fi
    fi

    # Run installation
    if [[ ${#profiles_to_install[@]} -gt 0 ]]; then
        main_install "${profiles_to_install[@]}"
    else
        die "No profiles selected for installation"
    fi
}

# Run main function
main "$@"
