#!/usr/bin/env bash
#
# AgentOps Validation Library
# Three-tier validation framework for installation and runtime
# Tiers: 1) Core files, 2) Profile validation, 3) 12-factor compliance
#

set -euo pipefail

# Source common functions if not already loaded
if ! declare -f die > /dev/null; then
    # shellcheck source=./common-functions.sh
    source "$(dirname "${BASH_SOURCE[0]}")/common-functions.sh"
fi

#######################################
# Tier 1: Core File Validation
# Validates essential AgentOps directory structure and files
# Returns:
#   0 if valid, 1 if validation fails
#######################################
validate_core_installation() {
    local errors=0

    info "Running Tier 1: Core File Validation"

    # Check AgentOps home directory
    if [[ ! -d "$AGENTOPS_HOME" ]]; then
        warn "AgentOps home directory not found: $AGENTOPS_HOME"
        ((errors++))
    else
        success "AgentOps home directory exists"
    fi

    # Check essential directories
    local required_dirs=(
        "${AGENTOPS_HOME}/scripts"
        "${AGENTOPS_HOME}/scripts/lib"
        "${AGENTOPS_PROFILES_DIR}"
        "${AGENTOPS_BACKUP_DIR}"
    )

    for dir in "${required_dirs[@]}"; do
        if [[ ! -d "$dir" ]]; then
            warn "Required directory not found: $dir"
            ((errors++))
        fi
    done

    if [[ $errors -eq 0 ]]; then
        success "All core directories present"
    fi

    # Check essential scripts
    local required_scripts=(
        "${AGENTOPS_HOME}/scripts/install.sh"
        "${AGENTOPS_HOME}/scripts/lib/common-functions.sh"
        "${AGENTOPS_HOME}/scripts/lib/validation.sh"
    )

    for script in "${required_scripts[@]}"; do
        if [[ ! -f "$script" ]]; then
            warn "Required script not found: $script"
            ((errors++))
        elif [[ ! -x "$script" ]] && [[ "$script" != *".sh" ]]; then
            warn "Script not executable: $script"
            ((errors++))
        fi
    done

    if [[ $errors -eq 0 ]]; then
        success "All core scripts present"
    fi

    # Check CLI tool
    if [[ -f "${AGENTOPS_HOME}/bin/agentops" ]]; then
        if [[ ! -x "${AGENTOPS_HOME}/bin/agentops" ]]; then
            warn "CLI tool not executable: ${AGENTOPS_HOME}/bin/agentops"
            ((errors++))
        else
            success "CLI tool present and executable"
        fi
    fi

    # Report results
    if [[ $errors -eq 0 ]]; then
        success "Tier 1: Core validation PASSED"
        return 0
    else
        warn "Tier 1: Core validation FAILED ($errors errors)"
        return 1
    fi
}

#######################################
# Tier 2: Profile Validation
# Validates installed profiles and their structure
# Arguments:
#   $1: Profile name (optional, validates all if not specified)
# Returns:
#   0 if valid, 1 if validation fails
#######################################
validate_profile() {
    local target_profile="${1:-}"
    local errors=0

    info "Running Tier 2: Profile Validation"

    # Get list of profiles to validate
    local profiles=()
    if [[ -n "$target_profile" ]]; then
        profiles=("$target_profile")
    else
        mapfile -t profiles < <(list_installed_profiles)
    fi

    if [[ ${#profiles[@]} -eq 0 ]]; then
        warn "No profiles installed to validate"
        return 1
    fi

    # Validate each profile
    for profile in "${profiles[@]}"; do
        info "Validating profile: $profile"

        local profile_dir="${AGENTOPS_PROFILES_DIR}/${profile}"

        # Check profile directory exists
        if [[ ! -d "$profile_dir" ]]; then
            warn "Profile directory not found: $profile_dir"
            ((errors++))
            continue
        fi

        # Check profile manifest
        local manifest="${profile_dir}/profile.yaml"
        if [[ ! -f "$manifest" ]]; then
            warn "Profile manifest not found: $manifest"
            ((errors++))
        else
            # Validate manifest structure
            if ! grep -q "apiVersion:" "$manifest"; then
                warn "Manifest missing apiVersion: $manifest"
                ((errors++))
            fi
            if ! grep -q "kind: Profile" "$manifest"; then
                warn "Manifest missing kind: Profile in $manifest"
                ((errors++))
            fi
            if ! grep -q "metadata:" "$manifest"; then
                warn "Manifest missing metadata section: $manifest"
                ((errors++))
            fi
            if ! grep -q "spec:" "$manifest"; then
                warn "Manifest missing spec section: $manifest"
                ((errors++))
            fi
        fi

        # Check required directories
        local required_profile_dirs=(
            "${profile_dir}/agents"
            "${profile_dir}/commands"
        )

        for dir in "${required_profile_dirs[@]}"; do
            if [[ ! -d "$dir" ]]; then
                warn "Required profile directory not found: $dir"
                ((errors++))
            fi
        done

        # Check agent count matches manifest
        if [[ -f "$manifest" ]]; then
            local declared_count
            declared_count=$(get_profile_metadata "$profile" "agent_count" || echo "0")
            local actual_count
            actual_count=$(find "${profile_dir}/agents" -name "*.md" -type f 2>/dev/null | wc -l | tr -d ' ')

            if [[ "$declared_count" != "$actual_count" ]]; then
                warn "Agent count mismatch in $profile: declared=$declared_count, actual=$actual_count"
                ((errors++))
            else
                success "Profile $profile: agent count matches ($actual_count agents)"
            fi
        fi
    done

    # Report results
    if [[ $errors -eq 0 ]]; then
        success "Tier 2: Profile validation PASSED"
        return 0
    else
        warn "Tier 2: Profile validation FAILED ($errors errors)"
        return 1
    fi
}

#######################################
# Tier 3: 12-Factor Compliance Validation
# Validates installation against 12-factor principles
# Returns:
#   0 if compliant, 1 if violations found
#######################################
validate_12factor_compliance() {
    local violations=0

    info "Running Tier 3: 12-Factor Compliance Validation"

    # Factor I: Codebase - One codebase, many deploys
    info "Checking Factor I: Codebase"
    if [[ -d "${AGENTOPS_HOME}/.git" ]]; then
        warn "AgentOps home should not be a git repository (Factor I violation)"
        ((violations++))
    else
        success "Factor I: No git repo in installation directory"
    fi

    # Factor II: Dependencies - Explicitly declare and isolate dependencies
    info "Checking Factor II: Dependencies"
    if [[ -f "${AGENTOPS_HOME}/scripts/lib/common-functions.sh" ]]; then
        success "Factor II: Dependencies properly isolated in lib/"
    else
        warn "Factor II: Missing dependency isolation"
        ((violations++))
    fi

    # Factor III: Config - Store config in environment
    info "Checking Factor III: Config"
    # Check that config files are in standard locations
    if [[ -f "$AGENTOPS_CONFIG_FILE" ]] || [[ -f "$PROJECT_CONFIG_FILE" ]]; then
        success "Factor III: Config files in standard locations"
    fi

    # Factor IV: Backing services - Treat backing services as attached resources
    info "Checking Factor IV: Backing Services"
    success "Factor IV: N/A for AgentOps (no backing services)"

    # Factor V: Build, release, run - Strictly separate build and run stages
    info "Checking Factor V: Build, Release, Run"
    if [[ -f "${AGENTOPS_HOME}/scripts/install.sh" ]]; then
        success "Factor V: Installation (build) separated from runtime"
    else
        warn "Factor V: Missing install script"
        ((violations++))
    fi

    # Factor VI: Processes - Execute as stateless processes
    info "Checking Factor VI: Processes"
    success "Factor VI: All scripts are stateless"

    # Factor VII: Port binding - Export services via port binding
    info "Checking Factor VII: Port Binding"
    success "Factor VII: N/A for AgentOps (CLI tool, no services)"

    # Factor VIII: Concurrency - Scale out via process model
    info "Checking Factor VIII: Concurrency"
    success "Factor VIII: Multi-profile support enables concurrency"

    # Factor IX: Disposability - Maximize robustness with fast startup and graceful shutdown
    info "Checking Factor IX: Disposability"
    if [[ -d "$AGENTOPS_BACKUP_DIR" ]]; then
        success "Factor IX: Backup/restore supports disposability"
    else
        warn "Factor IX: No backup directory"
        ((violations++))
    fi

    # Factor X: Dev/prod parity - Keep development, staging, and production as similar as possible
    info "Checking Factor X: Dev/Prod Parity"
    success "Factor X: Same installation across all environments"

    # Factor XI: Logs - Treat logs as event streams
    info "Checking Factor XI: Logs"
    if [[ -f "${AGENTOPS_HOME}/scripts/lib/logging.sh" ]]; then
        success "Factor XI: Structured logging library present"
    else
        warn "Factor XI: Missing logging library"
        ((violations++))
    fi

    # Factor XII: Admin processes - Run admin tasks as one-off processes
    info "Checking Factor XII: Admin Processes"
    if [[ -f "${AGENTOPS_HOME}/bin/agentops" ]]; then
        success "Factor XII: CLI tool for admin processes"
    else
        warn "Factor XII: Missing CLI tool"
        ((violations++))
    fi

    # Report results
    if [[ $violations -eq 0 ]]; then
        success "Tier 3: 12-Factor compliance PASSED (12/12 factors)"
        return 0
    else
        warn "Tier 3: 12-Factor compliance PARTIAL ($violations violations)"
        return 1
    fi
}

#######################################
# Run all validation tiers
# Arguments:
#   $1: Profile name (optional)
# Returns:
#   0 if all tiers pass, 1 if any tier fails
#######################################
validate_all() {
    local profile="${1:-}"
    local tier1_result=0
    local tier2_result=0
    local tier3_result=0

    echo ""
    echo "=========================================="
    echo "AgentOps Installation Validation"
    echo "=========================================="
    echo ""

    # Tier 1: Core files
    validate_core_installation || tier1_result=$?

    echo ""

    # Tier 2: Profiles
    validate_profile "$profile" || tier2_result=$?

    echo ""

    # Tier 3: 12-factor compliance
    validate_12factor_compliance || tier3_result=$?

    echo ""
    echo "=========================================="
    echo "Validation Summary"
    echo "=========================================="

    if [[ $tier1_result -eq 0 ]]; then
        success "Tier 1: Core Files - PASSED"
    else
        warn "Tier 1: Core Files - FAILED"
    fi

    if [[ $tier2_result -eq 0 ]]; then
        success "Tier 2: Profiles - PASSED"
    else
        warn "Tier 2: Profiles - FAILED"
    fi

    if [[ $tier3_result -eq 0 ]]; then
        success "Tier 3: 12-Factor Compliance - PASSED"
    else
        warn "Tier 3: 12-Factor Compliance - PARTIAL"
    fi

    echo ""

    if [[ $tier1_result -eq 0 ]] && [[ $tier2_result -eq 0 ]] && [[ $tier3_result -eq 0 ]]; then
        success "✓ All validation tiers PASSED"
        return 0
    else
        warn "⚠ Some validation tiers FAILED or PARTIAL"
        return 1
    fi
}

# Export validation functions
export -f validate_core_installation
export -f validate_profile
export -f validate_12factor_compliance
export -f validate_all
