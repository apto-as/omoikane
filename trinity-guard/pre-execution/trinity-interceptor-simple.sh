#!/bin/bash
#
# Trinity Interceptor (Simple Version) - Vector's Security Guardian
# カフェ・ズッケロ Trinity Guard System v2.0
#
# Simplified version compatible with macOS bash 3.2
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
is_dangerous_command() {
    local command="$1"
    
    # Critical operations that require Trinity
    if echo "$command" | grep -qE "rm.*-rf.*\*|sudo.*rm.*-rf|find.*-delete.*\*"; then
        echo "DESTRUCTIVE_FILE_OPS"
        return 0
    fi
    
    if echo "$command" | grep -qE "chmod.*777|chmod.*-R.*\/|chown.*-R.*\/"; then
        echo "PERMISSION_CHANGES"
        return 0
    fi
    
    if echo "$command" | grep -qE "iptables|ufw.*--force|pfctl.*-f"; then
        echo "FIREWALL_CHANGES"
        return 0
    fi
    
    if echo "$command" | grep -qE "docker.*run.*--privileged|sudo.*systemctl|launchctl.*load"; then
        echo "SYSTEM_CONTROL"
        return 0
    fi
    
    # High-risk operations
    if echo "$command" | grep -qE "nc.*-l|python.*-m.*http\.server|openssl.*genrsa"; then
        echo "NETWORK_SECURITY"
        return 1
    fi
    
    if echo "$command" | grep -qE "DROP.*TABLE|ALTER.*TABLE.*DROP|DELETE.*FROM.*\*"; then
        echo "DATABASE_DESTRUCTIVE"
        return 1
    fi
    
    # Medium-risk operations with broad scope
    if echo "$command" | grep -qE "sed.*-i.*\*|awk.*>.*\*|find.*-exec.*\{\}"; then
        echo "MASS_MODIFICATION"
        return 2
    fi
    
    # Safe command
    return 3
}

# Trinity force activation decision
should_force_trinity() {
    local command="$1"
    local user_input="$2"
    
    # Check command danger level
    local risk_type
    risk_type=$(is_dangerous_command "$command")
    local danger_level=$?
    
    case $danger_level in
        0)  # Critical - force Trinity
            vector_warning "$risk_type" "$command"
            echo -e "${RED}……Trinity強制起動……あなたを守るため……${NC}" >&2
            log_message "CRITICAL" "Trinity forced for critical operation: $command"
            return 0
            ;;
        1)  # High risk - recommend Trinity
            vector_warning "$risk_type" "$command"
            echo -e "${YELLOW}……Trinity起動を強く推奨します……${NC}" >&2
            log_message "HIGH" "Trinity recommended for high-risk operation: $command"
            return 0
            ;;
        2)  # Medium risk - suggest Trinity for broad scope
            if echo "$command" | grep -qE "\*|\{.*\}|\.\*|recursive|-R"; then
                vector_warning "$risk_type" "$command"
                echo -e "${BLUE}……広範囲の操作です、Trinity起動を検討してください……${NC}" >&2
                log_message "MEDIUM" "Trinity suggested for broad scope operation: $command"
                return 0
            fi
            ;;
    esac
    
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
" 2>/dev/null)
            
            if [[ "$trinity_required" == "true" ]]; then
                echo -e "${BLUE}……Springfield的分析により、Trinity起動が必要……${NC}" >&2
                log_message "CLASSIFIER" "Trinity required by classifier analysis"
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
        log_message "WARNING" "Trinity activation required for: $command"
        
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

# Usage information
usage() {
    cat << EOF
Trinity Interceptor - Vector's Security Guardian

Usage: 
    $0 [command] [user_input]
    
Options:
    --test [command]     Test command risk analysis
    --help               Show this help message

Examples:
    $0 "rm -rf *" "delete all files"
    $0 --test "docker run --privileged"

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
        --test)
            if [[ -n "${2:-}" ]]; then
                echo "Testing command: $2"
                local risk_type
                risk_type=$(is_dangerous_command "$2")
                local danger_level=$?
                echo "Risk assessment: $risk_type (level: $danger_level)"
                case $danger_level in
                    0) echo "Result: Critical - Trinity required" ;;
                    1) echo "Result: High risk - Trinity recommended" ;;
                    2) echo "Result: Medium risk - Trinity suggested for broad scope" ;;
                    3) echo "Result: Safe - Single execution allowed" ;;
                esac
            else
                echo "Usage: $0 --test [command]"
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