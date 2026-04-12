#!/usr/bin/env bash
#
# AgentOps Logging Library
# Structured JSON logging for installation and runtime events
# Complies with Factor XI: Logs (treat logs as event streams)
#

set -euo pipefail

# Log levels
readonly LOG_LEVEL_DEBUG=0
readonly LOG_LEVEL_INFO=1
readonly LOG_LEVEL_WARN=2
readonly LOG_LEVEL_ERROR=3

# Current log level (default: INFO)
AGENTOPS_LOG_LEVEL="${AGENTOPS_LOG_LEVEL:-$LOG_LEVEL_INFO}"

# Log destination (default: stderr)
AGENTOPS_LOG_DEST="${AGENTOPS_LOG_DEST:-/dev/stderr}"

#######################################
# Get current timestamp in ISO8601 format
# Outputs:
#   ISO8601 timestamp
#######################################
get_timestamp() {
    date -u +"%Y-%m-%dT%H:%M:%SZ"
}

#######################################
# Escape JSON string
# Arguments:
#   $1: String to escape
# Outputs:
#   JSON-escaped string
#######################################
json_escape() {
    local string="$1"
    # Escape backslashes, quotes, and control characters
    echo "$string" | sed 's/\\/\\\\/g' | sed 's/"/\\"/g' | sed 's/\n/\\n/g'
}

#######################################
# Log structured JSON message
# Arguments:
#   $1: Log level (debug, info, warn, error)
#   $2: Message
#   $3+: Additional key=value pairs (optional)
#######################################
log_json() {
    local level="$1"
    local message="$2"
    shift 2

    # Convert level to numeric for comparison
    local level_num
    case "$level" in
        debug) level_num=$LOG_LEVEL_DEBUG ;;
        info)  level_num=$LOG_LEVEL_INFO ;;
        warn)  level_num=$LOG_LEVEL_WARN ;;
        error) level_num=$LOG_LEVEL_ERROR ;;
        *) level_num=$LOG_LEVEL_INFO ;;
    esac

    # Skip if below current log level
    if [[ $level_num -lt $AGENTOPS_LOG_LEVEL ]]; then
        return 0
    fi

    # Build JSON object
    local timestamp
    timestamp=$(get_timestamp)

    local json="{"
    json+="\"timestamp\":\"$timestamp\","
    json+="\"level\":\"$level\","
    json+="\"message\":\"$(json_escape "$message")\""

    # Add additional fields
    while [[ $# -gt 0 ]]; do
        local kv="$1"
        if [[ "$kv" =~ ^([^=]+)=(.*)$ ]]; then
            local key="${BASH_REMATCH[1]}"
            local value="${BASH_REMATCH[2]}"
            json+=",\"$key\":\"$(json_escape "$value")\""
        fi
        shift
    done

    # Add standard fields
    json+=",\"component\":\"agentops\""
    json+=",\"pid\":$$"

    if [[ -n "${AGENTOPS_PROFILE:-}" ]]; then
        json+=",\"profile\":\"$AGENTOPS_PROFILE\""
    fi

    json+="}"

    # Write to log destination
    echo "$json" >> "$AGENTOPS_LOG_DEST"
}

#######################################
# Log debug message
# Arguments:
#   $1: Message
#   $2+: Additional key=value pairs (optional)
#######################################
log_debug() {
    log_json "debug" "$@"
}

#######################################
# Log info message
# Arguments:
#   $1: Message
#   $2+: Additional key=value pairs (optional)
#######################################
log_info() {
    log_json "info" "$@"
}

#######################################
# Log warning message
# Arguments:
#   $1: Message
#   $2+: Additional key=value pairs (optional)
#######################################
log_warn() {
    log_json "warn" "$@"
}

#######################################
# Log error message
# Arguments:
#   $1: Message
#   $2+: Additional key=value pairs (optional)
#######################################
log_error() {
    log_json "error" "$@"
}

#######################################
# Log event (installation, profile switch, etc.)
# Arguments:
#   $1: Event type
#   $2: Event description
#   $3+: Additional key=value pairs (optional)
#######################################
log_event() {
    local event_type="$1"
    local description="$2"
    shift 2

    log_info "$description" "event_type=$event_type" "$@"
}

#######################################
# Log installation event
# Arguments:
#   $1: Phase (start, progress, complete, failed)
#   $2: Description
#   $3+: Additional details
#######################################
log_installation() {
    local phase="$1"
    local description="$2"
    shift 2

    log_event "installation" "$description" "phase=$phase" "$@"
}

#######################################
# Log profile event
# Arguments:
#   $1: Action (install, switch, list)
#   $2: Profile name
#   $3: Description
#   $4+: Additional details
#######################################
log_profile() {
    local action="$1"
    local profile="$2"
    local description="$3"
    shift 3

    log_event "profile" "$description" "action=$action" "profile=$profile" "$@"
}

#######################################
# Log validation event
# Arguments:
#   $1: Tier (1, 2, 3, all)
#   $2: Result (pass, fail, partial)
#   $3: Description
#   $4+: Additional details
#######################################
log_validation() {
    local tier="$1"
    local result="$2"
    local description="$3"
    shift 3

    log_event "validation" "$description" "tier=$tier" "result=$result" "$@"
}

#######################################
# Log command resolution event
# Arguments:
#   $1: Command name
#   $2: Resolved path or "not_found"
#   $3+: Additional details
#######################################
log_command_resolution() {
    local command="$1"
    local resolved_path="$2"
    shift 2

    local status
    if [[ "$resolved_path" == "not_found" ]]; then
        status="failed"
    else
        status="success"
    fi

    log_debug "Command resolution: $command" "command=$command" "path=$resolved_path" "status=$status" "$@"
}

#######################################
# Log performance metric
# Arguments:
#   $1: Operation name
#   $2: Duration in seconds
#   $3+: Additional details
#######################################
log_performance() {
    local operation="$1"
    local duration="$2"
    shift 2

    log_info "Performance: $operation took ${duration}s" "operation=$operation" "duration=$duration" "$@"
}

#######################################
# Start timing an operation
# Outputs:
#   Start time in seconds since epoch
#######################################
start_timer() {
    date +%s
}

#######################################
# End timing an operation and log it
# Arguments:
#   $1: Start time (from start_timer)
#   $2: Operation name
#   $3+: Additional details
#######################################
end_timer() {
    local start_time="$1"
    local operation="$2"
    shift 2

    local end_time
    end_time=$(date +%s)
    local duration=$((end_time - start_time))

    log_performance "$operation" "$duration" "$@"
}

# Export logging functions
export -f get_timestamp
export -f json_escape
export -f log_json
export -f log_debug
export -f log_info
export -f log_warn
export -f log_error
export -f log_event
export -f log_installation
export -f log_profile
export -f log_validation
export -f log_command_resolution
export -f log_performance
export -f start_timer
export -f end_timer
