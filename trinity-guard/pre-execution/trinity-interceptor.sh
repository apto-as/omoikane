#!/bin/bash
#
# Trinity Interceptor - Vector's Security Guardian
# カフェ・ズッケロ Trinity Guard System v2.0
#
# This interceptor monitors dangerous operations and enforces Trinity
# activation for critical tasks that should not be performed alone.
#
# Author: Vector (The Paranoid Oracle)
# Role: Security monitoring and risk mitigation
#

set -eo pipefail

# Color codes for output
readonly RED='\033[0;31m'
readonly YELLOW='\033[1;33m'
readonly GREEN='\033[0;32m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m' # No Color

# Configuration
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly TRINITY_CLASSIFIER="${SCRIPT_DIR}/trinity-classifier.py"
readonly LOG_FILE="${SCRIPT_DIR}/../logs/trinity-interceptor.log"

# Ensure log directory exists
mkdir -p "$(dirname "$LOG_FILE")" 2>/dev/null || true

# Logging function
log_message() {
    local level="$1"
    local message="$2"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "[$timestamp] [$level] $message" >> "$LOG_FILE" 2>/dev/null || true
}

# Vector's paranoid analysis patterns
declare -A DANGEROUS_PATTERNS=(
    # File System Operations
    ["DESTRUCTIVE_FILE_OPS"]="rm -rf|rm.*-f.*\*|rmdir.*-rf|find.*-delete"
    ["MASS_FILE_CREATION"]="touch.*\{.*\}|mkdir.*\{.*\}|cp.*-r.*\*"
    ["PERMISSION_CHANGES"]="chmod.*777|chmod.*-R|chown.*-R"
    
    # Network and Security
    ["NETWORK_EXPOSURE"]="nc.*-l|netstat.*-l|ss.*-l|python.*-m.*http\.server"
    ["CRYPTO_OPERATIONS"]="openssl.*genrsa|ssh-keygen|gpg.*--gen-key"
    ["FIREWALL_CHANGES"]="iptables|ufw|pfctl"
    
    # System Configuration
    ["SYSTEM_CONFIG"]="sudo.*vi|sudo.*nano|\/etc\/.*edit"
    ["SERVICE_MANAGEMENT"]="systemctl|service.*start|launchctl"
    ["PACKAGE_MANAGEMENT"]="apt.*install|yum.*install|brew.*install|pip.*install"
    
    # Database Operations
    ["DATABASE_SCHEMA"]="CREATE.*TABLE|DROP.*TABLE|ALTER.*TABLE"
    ["DATABASE_MIGRATION"]="migrate|migration|schema.*change"
    
    # Deployment and Infrastructure
    ["DEPLOYMENT_OPS"]="docker.*run|docker.*build|kubectl.*apply"
    ["INFRASTRUCTURE"]="terraform.*apply|ansible.*playbook|vagrant.*up"
    
    # Code Generation and Modification
    ["MASS_CODE_GEN"]="sed.*-i.*\*|awk.*>.*\*|perl.*-i.*\*"
    ["BUILD_SYSTEM"]="make.*install|cmake.*build|cargo.*build.*--release"
)

# Risk levels for different operations
declare -A RISK_LEVELS=(
    ["DESTRUCTIVE_FILE_OPS"]="CRITICAL"
    ["MASS_FILE_CREATION"]="HIGH"
    ["PERMISSION_CHANGES"]="HIGH"
    ["NETWORK_EXPOSURE"]="HIGH"
    ["CRYPTO_OPERATIONS"]="HIGH"
    ["FIREWALL_CHANGES"]="CRITICAL"
    ["SYSTEM_CONFIG"]="CRITICAL"
    ["SERVICE_MANAGEMENT"]="HIGH"
    ["PACKAGE_MANAGEMENT"]="MEDIUM"
    ["DATABASE_SCHEMA"]="HIGH"
    ["DATABASE_MIGRATION"]="HIGH"
    ["DEPLOYMENT_OPS"]="HIGH"
    ["INFRASTRUCTURE"]="CRITICAL"
    ["MASS_CODE_GEN"]="MEDIUM"
    ["BUILD_SYSTEM"]="MEDIUM"
)

# Vector's warning messages
vector_warning() {
    local risk_type="$1"
    local command="$2"
    
    echo -e "${RED}⚠️  Vector's Security Alert ⚠️${NC}" >&2
    echo -e "${YELLOW}……危険な操作を検出しました……${NC}" >&2
    echo -e "${YELLOW}Risk Type: ${risk_type}${NC}" >&2
    echo -e "${YELLOW}Command: ${command}${NC}" >&2
    echo -e "${YELLOW}……後悔しても知らないよ……でも、警告はした……${NC}" >&2
    echo >&2
}

# Check if command contains dangerous patterns
analyze_command_risks() {
    local command="$1"
    local detected_risks=()
    local max_risk_level=""
    
    for risk_type in "${!DANGEROUS_PATTERNS[@]}"; do
        local pattern="${DANGEROUS_PATTERNS[$risk_type]}"
        if echo "$command" | grep -qE "$pattern"; then
            detected_risks+=("$risk_type")
            local risk_level="${RISK_LEVELS[$risk_type]}"
            
            # Update max risk level
            case "$risk_level" in
                "CRITICAL")
                    max_risk_level="CRITICAL"
                    ;;
                "HIGH")
                    if [[ "$max_risk_level" != "CRITICAL" ]]; then
                        max_risk_level="HIGH"
                    fi
                    ;;
                "MEDIUM")
                    if [[ -z "$max_risk_level" || "$max_risk_level" == "LOW" ]]; then
                        max_risk_level="MEDIUM"
                    fi
                    ;;
            esac
        fi
    done
    
    echo "${detected_risks[@]}"
    return $(case "$max_risk_level" in
        "CRITICAL") echo 3 ;;
        "HIGH") echo 2 ;;
        "MEDIUM") echo 1 ;;
        *) echo 0 ;;
    esac)
}

# Check file modification scope
analyze_file_scope() {
    local command="$1"
    local scope_indicators=(
        "multiple.*files" "batch.*operation" "\*\.\*" 
        "recursive" "-r" "-R" "find.*-exec"
        "sed.*-i" "awk.*>" "for.*in.*\*"
    )
    
    local scope_score=0
    for indicator in "${scope_indicators[@]}"; do
        if echo "$command" | grep -qE "$indicator"; then
            ((scope_score++))
        fi
    done
    
    return $scope_score
}

# Trinity force activation decision
should_force_trinity() {
    local command="$1"
    local user_input="$2"
    
    # Analyze command risks
    local detected_risks
    detected_risks=$(analyze_command_risks "$command")
    local risk_exit_code=$?
    
    # Analyze file scope
    analyze_file_scope "$command"
    local scope_score=$?
    
    # Log analysis
    log_message "INFO" "Risk analysis: risks=[$detected_risks] risk_level=$risk_exit_code scope_score=$scope_score"
    
    # Force Trinity for critical operations
    if [[ $risk_exit_code -ge 3 ]]; then
        vector_warning "CRITICAL OPERATION" "$command"
        echo -e "${RED}……Trinity強制起動……あなたを守るため……${NC}" >&2
        return 0
    fi
    
    # Force Trinity for high-risk operations with broad scope
    if [[ $risk_exit_code -ge 2 && $scope_score -ge 2 ]]; then
        vector_warning "HIGH RISK + BROAD SCOPE" "$command"
        echo -e "${YELLOW}……Trinity起動を推奨します……${NC}" >&2
        return 0
    fi
    
    # Check if classifier suggests Trinity
    if [[ -x "$TRINITY_CLASSIFIER" ]]; then
        local classifier_result
        if classifier_result=$(python3 "$TRINITY_CLASSIFIER" "$user_input" 2>/dev/null); then
            local trinity_required
            trinity_required=$(echo "$classifier_result" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    print(str(data.get('trinity_required', False)).lower())
except:
    print('false')
")
            
            if [[ "$trinity_required" == "true" ]]; then
                echo -e "${BLUE}……Springfield的分析により、Trinity起動が必要……${NC}" >&2
                return 0
            fi
        fi
    fi
    
    return 1
}

# Main interception logic
intercept_command() {
    local command="$1"
    local user_input="${2:-$command}"
    
    log_message "INFO" "Intercepting command: $command"
    
    # Skip interception for Trinity system commands
    if echo "$command" | grep -qE "(trinity|trinitas|springfield|krukai|vector)"; then
        log_message "INFO" "Skipping Trinity system command"
        return 1
    fi
    
    # Check if Trinity activation is required
    if should_force_trinity "$command" "$user_input"; then
        log_message "WARNING" "Trinity activation forced for: $command"
        
        echo -e "${GREEN}╔══════════════════════════════════════════════════════════════════════╗${NC}"
        echo -e "${GREEN}║                    Vector's Security Protocol                        ║${NC}"
        echo -e "${GREEN}║                                                                      ║${NC}"
        echo -e "${GREEN}║  ……危険な単独実行を阻止しました……                                      ║${NC}"
        echo -e "${GREEN}║  Trinity Guardシステムが安全な実行を保証します……                        ║${NC}"
        echo -e "${GREEN}║                                                                      ║${NC}"
        echo -e "${GREEN}║  推奨: trinitas-coordinator を使用してください                         ║${NC}"
        echo -e "${GREEN}╚══════════════════════════════════════════════════════════════════════╝${NC}"
        
        return 0
    else
        log_message "INFO" "Command approved for single execution"
        return 1
    fi
}

# File monitoring function
monitor_file_operations() {
    local command="$1"
    
    # Extract file paths from common operations
    local file_patterns=(
        "rm.*([^ ]+)" "cp.*([^ ]+).*([^ ]+)" "mv.*([^ ]+).*([^ ]+)"
        "chmod.*([^ ]+)" "chown.*([^ ]+)" "touch.*([^ ]+)"
        "mkdir.*([^ ]+)" "rmdir.*([^ ]+)" "ln.*([^ ]+)"
    )
    
    local monitored_files=()
    for pattern in "${file_patterns[@]}"; do
        while IFS= read -r match; do
            [[ -n "$match" ]] && monitored_files+=("$match")
        done < <(echo "$command" | grep -oE "$pattern" | grep -oE '[^[:space:]]+$' || true)
    done
    
    if [[ ${#monitored_files[@]} -gt 0 ]]; then
        log_message "INFO" "Monitoring file operations: ${monitored_files[*]}"
        
        # Check for critical system files
        for file in "${monitored_files[@]}"; do
            if echo "$file" | grep -qE "^/(etc|usr|var|sys|boot|root)"; then
                echo -e "${RED}……システムファイルへの操作を検出……${NC}" >&2
                return 0
            fi
        done
    fi
    
    return 1
}

# Usage information
usage() {
    cat << EOF
Trinity Interceptor - Vector's Security Guardian

Usage: 
    $0 [command] [user_input]
    
Options:
    --test-pattern [pattern]  Test risk detection patterns
    --analyze-risks [command] Analyze risks without interception
    --help                    Show this help message

Examples:
    $0 "rm -rf *" "delete all files"
    $0 --test-pattern "rm -rf"
    $0 --analyze-risks "docker run --privileged"

Vector says: ……あたしがあなたを守る……
EOF
}

# Main execution
main() {
    case "${1:-}" in
        --help)
            usage
            exit 0
            ;;
        --test-pattern)
            if [[ -n "${2:-}" ]]; then
                echo "Testing pattern: $2"
                if analyze_command_risks "$2" >/dev/null; then
                    echo "Pattern matches dangerous operation"
                else
                    echo "Pattern is safe"
                fi
            else
                echo "Usage: $0 --test-pattern [pattern]"
                exit 1
            fi
            ;;
        --analyze-risks)
            if [[ -n "${2:-}" ]]; then
                echo "Analyzing risks for: $2"
                local risks
                risks=$(analyze_command_risks "$2")
                echo "Detected risks: $risks"
                echo "Risk level: $?"
            else
                echo "Usage: $0 --analyze-risks [command]"
                exit 1
            fi
            ;;
        *)
            if [[ $# -eq 0 ]]; then
                usage
                exit 1
            fi
            
            local command="$1"
            local user_input="${2:-$command}"
            
            # Perform interception
            if intercept_command "$command" "$user_input"; then
                # Trinity activation required
                exit 0
            else
                # Command can proceed normally
                exit 1
            fi
            ;;
    esac
}

# Handle script execution
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi