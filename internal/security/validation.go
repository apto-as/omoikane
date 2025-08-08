// Package security - Input validation and audit logging components
// Vector: "……入力の全てを検証し、記録に残す……"
package security

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"
)

// InputValidator provides comprehensive input validation
type InputValidator struct {
	config            SecurityConfig
	blockedPatterns   []*regexp.Regexp
	allowedCommands   map[string]bool
	suspiciousChars   []rune
	validationRules   []ValidationRule
}

// ValidationRule represents a validation rule
type ValidationRule struct {
	Name        string
	Pattern     *regexp.Regexp
	Severity    ViolationSeverity
	Description string
	Action      ViolationAction
}

// NewInputValidator creates a new input validator
func NewInputValidator(config SecurityConfig) *InputValidator {
	validator := &InputValidator{
		config:          config,
		blockedPatterns: make([]*regexp.Regexp, 0),
		allowedCommands: make(map[string]bool),
		suspiciousChars: []rune{'$', '`', '|', '&', ';', '>', '<', '*', '?'},
	}
	
	// Compile blocked patterns
	for _, pattern := range config.BlockedPatterns {
		if regex, err := regexp.Compile(pattern); err == nil {
			validator.blockedPatterns = append(validator.blockedPatterns, regex)
		}
	}
	
	// Build allowed commands map
	for _, cmd := range config.AllowedCommands {
		validator.allowedCommands[cmd] = true
	}
	
	// Initialize validation rules
	validator.initValidationRules()
	
	return validator
}

func (iv *InputValidator) initValidationRules() {
	// Command injection rules
	iv.validationRules = append(iv.validationRules, ValidationRule{
		Name:        "command_injection",
		Pattern:     regexp.MustCompile(`\$\(.*\)|` + "`" + `.*` + "`" + `|system\s*\(|exec\s*\(`),
		Severity:    SeverityCritical,
		Description: "Potential command injection detected",
		Action:      ActionBlock,
	})
	
	// Path traversal rules
	iv.validationRules = append(iv.validationRules, ValidationRule{
		Name:        "path_traversal",
		Pattern:     regexp.MustCompile(`\.\.\/|\.\.\\|\/etc\/|\/proc\/|\/sys\/`),
		Severity:    SeverityHigh,
		Description: "Path traversal attempt detected",
		Action:      ActionBlock,
	})
	
	// File operation rules
	iv.validationRules = append(iv.validationRules, ValidationRule{
		Name:        "dangerous_file_ops",
		Pattern:     regexp.MustCompile(`rm\s+-rf|chmod\s+777|chown\s+|passwd\s+`),
		Severity:    SeverityHigh,
		Description: "Dangerous file operation detected",
		Action:      ActionBlock,
	})
	
	// Network operation rules  
	iv.validationRules = append(iv.validationRules, ValidationRule{
		Name:        "network_operations",
		Pattern:     regexp.MustCompile(`nc\s+|ncat\s+|telnet\s+|ssh\s+|scp\s+|wget\s+|curl\s+.*\|\s*sh`),
		Severity:    SeverityMedium,
		Description: "Network operation detected",
		Action:      ActionWarn,
	})
}

// ValidateInput performs comprehensive input validation
func (iv *InputValidator) ValidateInput(input string) *SecurityViolation {
	// Length check
	if len(input) > iv.config.MaxInputLength {
		return &SecurityViolation{
			ID:          generateViolationID(ViolationInputValidation, input),
			Type:        ViolationInputValidation,
			Severity:    SeverityMedium,
			Description: fmt.Sprintf("Input exceeds maximum length: %d > %d", len(input), iv.config.MaxInputLength),
			Input:       input[:min(100, len(input))], // Truncate for logging
			Timestamp:   time.Now(),
			Source:      "input_validator",
			Action:      ActionWarn,
		}
	}
	
	// Character validation
	if violation := iv.validateCharacters(input); violation != nil {
		return violation
	}
	
	// Pattern validation
	if violation := iv.validatePatterns(input); violation != nil {
		return violation
	}
	
	// Command validation
	if violation := iv.validateCommands(input); violation != nil {
		return violation
	}
	
	// Rule-based validation
	for _, rule := range iv.validationRules {
		if rule.Pattern.MatchString(input) {
			return &SecurityViolation{
				ID:          generateViolationID(ViolationSuspiciousPattern, input),
				Type:        ViolationSuspiciousPattern,
				Severity:    rule.Severity,
				Description: rule.Description,
				Input:       input[:min(100, len(input))],
				Timestamp:   time.Now(),
				Source:      "validation_rule",
				Action:      rule.Action,
			}
		}
	}
	
	return nil // Input is valid
}

func (iv *InputValidator) validateCharacters(input string) *SecurityViolation {
	suspiciousCount := 0
	
	for _, r := range input {
		// Check for non-printable characters
		if !unicode.IsPrint(r) && !unicode.IsSpace(r) {
			return &SecurityViolation{
				ID:          generateViolationID(ViolationInputValidation, input),
				Type:        ViolationInputValidation,
				Severity:    SeverityMedium,
				Description: "Non-printable characters detected in input",
				Input:       input[:min(100, len(input))],
				Timestamp:   time.Now(),
				Source:      "character_validator",
				Action:      ActionWarn,
			}
		}
		
		// Count suspicious characters
		for _, suspicious := range iv.suspiciousChars {
			if r == suspicious {
				suspiciousCount++
				break
			}
		}
	}
	
	// Too many suspicious characters
	if suspiciousCount > 5 {
		return &SecurityViolation{
			ID:          generateViolationID(ViolationSuspiciousPattern, input),
			Type:        ViolationSuspiciousPattern,
			Severity:    SeverityMedium,
			Description: fmt.Sprintf("High concentration of suspicious characters: %d", suspiciousCount),
			Input:       input[:min(100, len(input))],
			Timestamp:   time.Now(),
			Source:      "character_validator",
			Action:      ActionWarn,
		}
	}
	
	return nil
}

func (iv *InputValidator) validatePatterns(input string) *SecurityViolation {
	for _, pattern := range iv.blockedPatterns {
		if pattern.MatchString(input) {
			return &SecurityViolation{
				ID:          generateViolationID(ViolationSuspiciousPattern, input),
				Type:        ViolationSuspiciousPattern,
				Severity:    SeverityHigh,
				Description: fmt.Sprintf("Blocked pattern detected: %s", pattern.String()),
				Input:       input[:min(100, len(input))],
				Timestamp:   time.Now(),
				Source:      "pattern_validator",
				Action:      ActionBlock,
			}
		}
	}
	
	return nil
}

func (iv *InputValidator) validateCommands(input string) *SecurityViolation {
	// Extract potential commands (simple approach)
	words := strings.Fields(input)
	if len(words) == 0 {
		return nil
	}
	
	firstWord := words[0]
	
	// Check if command is in allowed list
	if _, allowed := iv.allowedCommands[firstWord]; !allowed {
		// Check for common dangerous commands
		dangerousCommands := map[string]ViolationSeverity{
			"rm":     SeverityCritical,
			"sudo":   SeverityCritical,
			"su":     SeverityCritical,
			"chmod":  SeverityHigh,
			"chown":  SeverityHigh,
			"passwd": SeverityHigh,
			"nc":     SeverityMedium,
			"wget":   SeverityMedium,
			"curl":   SeverityMedium,
		}
		
		if severity, dangerous := dangerousCommands[firstWord]; dangerous {
			return &SecurityViolation{
				ID:          generateViolationID(ViolationCommandInjection, input),
				Type:        ViolationCommandInjection,
				Severity:    severity,
				Description: fmt.Sprintf("Dangerous command detected: %s", firstWord),
				Input:       input[:min(100, len(input))],
				Timestamp:   time.Now(),
				Source:      "command_validator",
				Action:      iv.mapSeverityToAction(severity),
			}
		}
	}
	
	return nil
}

func (iv *InputValidator) mapSeverityToAction(severity ViolationSeverity) ViolationAction {
	switch severity {
	case SeverityCritical:
		return ActionTerminate
	case SeverityHigh:
		return ActionBlock
	case SeverityMedium:
		return ActionWarn
	default:
		return ActionAllow
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// AuditLogger handles security audit logging
type AuditLogger struct {
	config    SecurityConfig
	logFile   *os.File
	mu        sync.Mutex
	entries   []AuditEntry
	maxBuffer int
}

// AuditEntry represents a security audit log entry
type AuditEntry struct {
	Type       string                 `json:"type"`
	Input      string                 `json:"input,omitempty"`
	RiskScore  float64                `json:"risk_score,omitempty"`
	Violations int                    `json:"violations,omitempty"`
	Action     string                 `json:"action,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	Duration   time.Duration          `json:"duration,omitempty"`
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(config SecurityConfig) *AuditLogger {
	logger := &AuditLogger{
		config:    config,
		entries:   make([]AuditEntry, 0),
		maxBuffer: 1000,
	}
	
	if config.EnableAuditLogging {
		// Open log file (in production, use proper log rotation)
		if file, err := os.OpenFile("security_audit.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
			logger.logFile = file
		}
	}
	
	return logger
}

// LogEntry logs an audit entry
func (al *AuditLogger) LogEntry(entry *AuditEntry) {
	al.mu.Lock()
	defer al.mu.Unlock()
	
	// Add to buffer
	al.entries = append(al.entries, *entry)
	
	// Manage buffer size
	if len(al.entries) > al.maxBuffer {
		al.entries = al.entries[al.maxBuffer/2:] // Keep half
	}
	
	// Write to file if enabled
	if al.logFile != nil {
		logLine := fmt.Sprintf("[%s] %s: %s (risk: %.2f, violations: %d)\n",
			entry.Timestamp.Format(time.RFC3339),
			entry.Type,
			entry.Action,
			entry.RiskScore,
			entry.Violations)
		al.logFile.WriteString(logLine)
		al.logFile.Sync()
	}
}

// LogAlert logs a security alert
func (al *AuditLogger) LogAlert(alert SecurityAlert) {
	entry := &AuditEntry{
		Type:      "security_alert",
		Context:   alert.Context,
		Timestamp: alert.Timestamp,
	}
	al.LogEntry(entry)
}

// GetRecentEntries returns recent audit entries
func (al *AuditLogger) GetRecentEntries(count int) []AuditEntry {
	al.mu.Lock()
	defer al.mu.Unlock()
	
	if count > len(al.entries) {
		count = len(al.entries)
	}
	
	// Return last 'count' entries
	start := len(al.entries) - count
	result := make([]AuditEntry, count)
	copy(result, al.entries[start:])
	
	return result
}

// Close closes the audit logger
func (al *AuditLogger) Close() error {
	if al.logFile != nil {
		return al.logFile.Close()
	}
	return nil
}

// ViolationLogger specifically handles security violations
type ViolationLogger struct {
	auditLogger *AuditLogger
	violations  map[string]*SecurityViolation
	mu          sync.RWMutex
}

// NewViolationLogger creates a new violation logger
func NewViolationLogger(config SecurityConfig) *ViolationLogger {
	return &ViolationLogger{
		auditLogger: NewAuditLogger(config),
		violations:  make(map[string]*SecurityViolation),
	}
}

// LogViolation logs a security violation
func (vl *ViolationLogger) LogViolation(violation *SecurityViolation) {
	vl.mu.Lock()
	defer vl.mu.Unlock()
	
	// Store violation
	vl.violations[violation.ID] = violation
	
	// Log to audit trail
	entry := &AuditEntry{
		Type:      "security_violation",
		Input:     violation.Input,
		Action:    string(violation.Action),
		Context: map[string]interface{}{
			"violation_type": string(violation.Type),
			"severity":       string(violation.Severity),
			"description":    violation.Description,
		},
		Timestamp: violation.Timestamp,
	}
	vl.auditLogger.LogEntry(entry)
}

// GetViolations returns stored violations
func (vl *ViolationLogger) GetViolations() map[string]*SecurityViolation {
	vl.mu.RLock()
	defer vl.mu.RUnlock()
	
	violations := make(map[string]*SecurityViolation)
	for k, v := range vl.violations {
		violations[k] = v
	}
	return violations
}

// ResourceMonitor monitors system resource usage
type ResourceMonitor struct {
	limits      ResourceLimits
	currentUsage ResourceUsage
	mu          sync.RWMutex
	alerts      chan ResourceAlert
}

// ResourceUsage tracks current resource consumption
type ResourceUsage struct {
	MemoryMB      float64   `json:"memory_mb"`
	CPUPercent    float64   `json:"cpu_percent"`
	ActiveOps     int       `json:"active_operations"`
	NetworkOps    int       `json:"network_operations"`
	LastUpdate    time.Time `json:"last_update"`
}

// ResourceAlert represents a resource usage alert
type ResourceAlert struct {
	Type        string    `json:"type"`
	Current     float64   `json:"current"`
	Limit       float64   `json:"limit"`
	Severity    string    `json:"severity"`
	Timestamp   time.Time `json:"timestamp"`
}

// NewResourceMonitor creates a new resource monitor
func NewResourceMonitor(limits ResourceLimits) *ResourceMonitor {
	return &ResourceMonitor{
		limits:  limits,
		alerts:  make(chan ResourceAlert, 50),
	}
}

// CheckResourceUsage validates current resource usage against limits
func (rm *ResourceMonitor) CheckResourceUsage(usage ResourceUsage) *SecurityViolation {
	rm.mu.Lock()
	rm.currentUsage = usage
	rm.mu.Unlock()
	
	// Memory check
	if usage.MemoryMB > float64(rm.limits.MaxMemoryMB) {
		alert := ResourceAlert{
			Type:      "memory_limit_exceeded",
			Current:   usage.MemoryMB,
			Limit:     float64(rm.limits.MaxMemoryMB),
			Severity:  "high",
			Timestamp: time.Now(),
		}
		
		select {
		case rm.alerts <- alert:
		default:
		}
		
		return &SecurityViolation{
			ID:          generateViolationID(ViolationResourceAbuse, fmt.Sprintf("memory:%.1f", usage.MemoryMB)),
			Type:        ViolationResourceAbuse,
			Severity:    SeverityHigh,
			Description: fmt.Sprintf("Memory usage exceeded limit: %.1fMB > %dMB", usage.MemoryMB, rm.limits.MaxMemoryMB),
			Timestamp:   time.Now(),
			Source:      "resource_monitor",
			Action:      ActionBlock,
		}
	}
	
	// CPU check
	if usage.CPUPercent > rm.limits.MaxCPUPercent {
		return &SecurityViolation{
			ID:          generateViolationID(ViolationResourceAbuse, fmt.Sprintf("cpu:%.1f", usage.CPUPercent)),
			Type:        ViolationResourceAbuse,
			Severity:    SeverityMedium,
			Description: fmt.Sprintf("CPU usage exceeded limit: %.1f%% > %.1f%%", usage.CPUPercent, rm.limits.MaxCPUPercent),
			Timestamp:   time.Now(),
			Source:      "resource_monitor",
			Action:      ActionWarn,
		}
	}
	
	return nil
}

// GetCurrentUsage returns current resource usage
func (rm *ResourceMonitor) GetCurrentUsage() ResourceUsage {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.currentUsage
}

// GetAlerts returns the resource alerts channel
func (rm *ResourceMonitor) GetAlerts() <-chan ResourceAlert {
	return rm.alerts
}