// Package security provides comprehensive security audit and monitoring
// Vector: "……全ての脅威を検出し、防御する……"
// 入力検証、リスクスコア計算、危険操作の検出とブロック、監査ログ、リソース制限の強化
package security

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

// SecurityAuditor performs comprehensive security analysis
// Vector: "……悪意ある入力も見逃さない……"
type SecurityAuditor struct {
	// Risk assessment
	riskCalculator  *RiskCalculator
	threatDetector  *ThreatDetector
	inputValidator  *InputValidator
	
	// Audit logging
	auditLogger     *AuditLogger
	violationLogger *ViolationLogger
	
	// Resource monitoring
	resourceMonitor *ResourceMonitor
	
	// Configuration
	config SecurityConfig
	
	// State management
	mu sync.RWMutex
	violations map[string]*SecurityViolation
	alerts     chan SecurityAlert
}

// SecurityConfig defines security parameters
type SecurityConfig struct {
	MaxInputLength     int               `json:"max_input_length" yaml:"max_input_length"`
	MaxResourceUsage   ResourceLimits    `json:"max_resource_usage" yaml:"max_resource_usage"`
	ThreatThreshold    float64           `json:"threat_threshold" yaml:"threat_threshold"`
	EnableAuditLogging bool              `json:"enable_audit_logging" yaml:"enable_audit_logging"`
	AuditRetentionDays int               `json:"audit_retention_days" yaml:"audit_retention_days"`
	BlockedPatterns    []string          `json:"blocked_patterns" yaml:"blocked_patterns"`
	AllowedCommands    []string          `json:"allowed_commands" yaml:"allowed_commands"`
}

// DefaultSecurityConfig returns paranoid security settings
func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		MaxInputLength:     10000, // 10KB limit
		MaxResourceUsage:   DefaultResourceLimits(),
		ThreatThreshold:    0.3,   // Low threshold for high security
		EnableAuditLogging: true,
		AuditRetentionDays: 90,
		BlockedPatterns: []string{
			// Command injection patterns
			`rm\s+-rf`,
			`sudo\s+`,
			`wget\s+http`,
			`curl\s+.*\|\s*sh`,
			// Path traversal
			`\.\.\/`,
			`\.\.\\`,
			// Suspicious file operations  
			`/etc/passwd`,
			`/etc/shadow`,
			`~/.ssh`,
			// Network operations
			`nc\s+.*\s+\d+`,
			`ncat\s+`,
			// Code execution
			`eval\s*\(`,
			`exec\s*\(`,
			`system\s*\(`,
		},
		AllowedCommands: []string{
			"ls", "cat", "grep", "find", "head", "tail", "wc",
			"sort", "uniq", "awk", "sed", "cut", "tr",
			"git", "go", "npm", "yarn", "python", "node",
		},
	}
}

// ResourceLimits defines system resource constraints
type ResourceLimits struct {
	MaxMemoryMB   int           `json:"max_memory_mb"`
	MaxCPUPercent float64       `json:"max_cpu_percent"`
	MaxDuration   time.Duration `json:"max_duration"`
	MaxFileSize   int64         `json:"max_file_size"`
	MaxNetworkOps int           `json:"max_network_ops"`
}

func DefaultResourceLimits() ResourceLimits {
	return ResourceLimits{
		MaxMemoryMB:   512,  // 512MB limit
		MaxCPUPercent: 50.0, // 50% CPU limit
		MaxDuration:   30 * time.Second,
		MaxFileSize:   100 * 1024 * 1024, // 100MB
		MaxNetworkOps: 0, // No network operations by default
	}
}

// SecurityViolation represents a detected security issue
type SecurityViolation struct {
	ID          string                 `json:"id"`
	Type        ViolationType          `json:"type"`
	Severity    ViolationSeverity      `json:"severity"`
	Description string                 `json:"description"`
	Input       string                 `json:"input,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Source      string                 `json:"source"`
	Action      ViolationAction        `json:"action"`
	RiskImpact  float64                `json:"risk_impact"`
}

type ViolationType string
const (
	ViolationCommandInjection ViolationType = "command_injection"
	ViolationPathTraversal     ViolationType = "path_traversal"
	ViolationResourceAbuse     ViolationType = "resource_abuse"
	ViolationSuspiciousPattern ViolationType = "suspicious_pattern"
	ViolationInputValidation   ViolationType = "input_validation"
	ViolationUnauthorizedAccess ViolationType = "unauthorized_access"
)

type ViolationSeverity string
const (
	SeverityLow      ViolationSeverity = "low"
	SeverityMedium   ViolationSeverity = "medium"
	SeverityHigh     ViolationSeverity = "high"
	SeverityCritical ViolationSeverity = "critical"
)

type ViolationAction string
const (
	ActionAllow   ViolationAction = "allow"
	ActionWarn    ViolationAction = "warn"
	ActionBlock   ViolationAction = "block"
	ActionTerminate ViolationAction = "terminate"
)

// SecurityAlert represents a security alert
type SecurityAlert struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Message     string                 `json:"message"`
	Severity    ViolationSeverity      `json:"severity"`
	Timestamp   time.Time              `json:"timestamp"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Recommended []string               `json:"recommended_actions,omitempty"`
}

// NewSecurityAuditor creates a new security auditor
// Vector: "……完全な防御システムを初期化……"
func NewSecurityAuditor(config SecurityConfig) *SecurityAuditor {
	auditor := &SecurityAuditor{
		riskCalculator:  NewRiskCalculator(config),
		threatDetector:  NewThreatDetector(config),
		inputValidator:  NewInputValidator(config),
		auditLogger:     NewAuditLogger(config),
		violationLogger: NewViolationLogger(config),
		resourceMonitor: NewResourceMonitor(config.MaxResourceUsage),
		config:         config,
		violations:     make(map[string]*SecurityViolation),
		alerts:         make(chan SecurityAlert, 100),
	}
	
	// Start monitoring goroutines
	go auditor.alertMonitor()
	go auditor.violationCleaner()
	
	return auditor
}

// AnalyzeSecurity performs comprehensive security analysis
func (sa *SecurityAuditor) AnalyzeSecurity(ctx context.Context, input string, metadata map[string]interface{}) (*SecurityAssessment, error) {
	startTime := time.Now()
	
	// Create assessment context
	assessment := &SecurityAssessment{
		Input:     input,
		Metadata:  metadata,
		Timestamp: startTime,
		Violations: []SecurityViolation{},
	}
	
	// Input validation - first line of defense
	if violation := sa.inputValidator.ValidateInput(input); violation != nil {
		assessment.Violations = append(assessment.Violations, *violation)
		assessment.RiskScore += 0.3
	}
	
	// Threat detection - pattern matching
	threats := sa.threatDetector.DetectThreats(input, metadata)
	for _, threat := range threats {
		assessment.Violations = append(assessment.Violations, threat)
		assessment.RiskScore += threat.RiskImpact
	}
	
	// Risk calculation - comprehensive scoring
	riskScore := sa.riskCalculator.CalculateRisk(input, metadata, assessment.Violations)
	assessment.RiskScore = riskScore
	
	// Determine action based on risk
	assessment.RecommendedAction = sa.determineAction(riskScore, assessment.Violations)
	
	// Log audit trail
	if sa.config.EnableAuditLogging {
		auditEntry := &AuditEntry{
			Type:       "security_analysis",
			Input:      input,
			RiskScore:  riskScore,
			Violations: len(assessment.Violations),
			Action:     string(assessment.RecommendedAction),
			Timestamp:  startTime,
			Duration:   time.Since(startTime),
		}
		sa.auditLogger.LogEntry(auditEntry)
	}
	
	// Generate alerts for high-risk assessments
	if riskScore > sa.config.ThreatThreshold {
		alert := SecurityAlert{
			ID:        generateAlertID(),
			Type:      "high_risk_detected",
			Message:   fmt.Sprintf("High risk score detected: %.2f", riskScore),
			Severity:  sa.mapRiskToSeverity(riskScore),
			Timestamp: time.Now(),
			Context:   map[string]interface{}{
				"input_length": len(input),
				"violations":   len(assessment.Violations),
				"risk_score":   riskScore,
			},
		}
		
		select {
		case sa.alerts <- alert:
		default: // Don't block if alert channel is full
		}
	}
	
	return assessment, nil
}

// SecurityAssessment contains the results of security analysis
type SecurityAssessment struct {
	Input             string              `json:"input"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	RiskScore         float64             `json:"risk_score"`
	Violations        []SecurityViolation `json:"violations"`
	RecommendedAction ViolationAction     `json:"recommended_action"`
	Timestamp         time.Time           `json:"timestamp"`
	ProcessingTime    time.Duration       `json:"processing_time"`
}

// IsBlocked returns true if the input should be blocked
func (sa *SecurityAssessment) IsBlocked() bool {
	return sa.RecommendedAction == ActionBlock || sa.RecommendedAction == ActionTerminate
}

// HasCriticalViolations returns true if there are critical security violations
func (sa *SecurityAssessment) HasCriticalViolations() bool {
	for _, violation := range sa.Violations {
		if violation.Severity == SeverityCritical {
			return true
		}
	}
	return false
}

func (sa *SecurityAuditor) determineAction(riskScore float64, violations []SecurityViolation) ViolationAction {
	// Check for critical violations first
	for _, violation := range violations {
		if violation.Severity == SeverityCritical {
			return ActionTerminate
		}
		if violation.Severity == SeverityHigh {
			return ActionBlock
		}
	}
	
	// Risk-based decision
	if riskScore > 0.8 {
		return ActionBlock
	} else if riskScore > 0.5 {
		return ActionWarn
	}
	
	return ActionAllow
}

func (sa *SecurityAuditor) mapRiskToSeverity(riskScore float64) ViolationSeverity {
	switch {
	case riskScore > 0.8:
		return SeverityCritical
	case riskScore > 0.6:
		return SeverityHigh
	case riskScore > 0.4:
		return SeverityMedium
	default:
		return SeverityLow
	}
}

func (sa *SecurityAuditor) alertMonitor() {
	for alert := range sa.alerts {
		// Handle security alerts
		// Vector: "……アラートを適切に処理……"
		sa.handleSecurityAlert(alert)
	}
}

func (sa *SecurityAuditor) handleSecurityAlert(alert SecurityAlert) {
	// Log the alert
	if sa.config.EnableAuditLogging {
		sa.auditLogger.LogAlert(alert)
	}
	
	// Store violation for tracking
	sa.mu.Lock()
	violation := &SecurityViolation{
		ID:          alert.ID,
		Type:        ViolationType(alert.Type),
		Severity:    alert.Severity,
		Description: alert.Message,
		Timestamp:   alert.Timestamp,
		Action:      sa.determineAction(0.5, []SecurityViolation{}), // Default action
	}
	sa.violations[alert.ID] = violation
	sa.mu.Unlock()
}

func (sa *SecurityAuditor) violationCleaner() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			sa.cleanOldViolations()
		}
	}
}

func (sa *SecurityAuditor) cleanOldViolations() {
	sa.mu.Lock()
	defer sa.mu.Unlock()
	
	cutoff := time.Now().Add(-24 * time.Hour) // Keep violations for 24 hours
	
	for id, violation := range sa.violations {
		if violation.Timestamp.Before(cutoff) {
			delete(sa.violations, id)
		}
	}
}

// GetViolations returns current security violations
func (sa *SecurityAuditor) GetViolations() map[string]*SecurityViolation {
	sa.mu.RLock()
	defer sa.mu.RUnlock()
	
	violations := make(map[string]*SecurityViolation)
	for k, v := range sa.violations {
		violations[k] = v
	}
	return violations
}

// GetAlerts returns the alerts channel for monitoring
func (sa *SecurityAuditor) GetAlerts() <-chan SecurityAlert {
	return sa.alerts
}

func generateAlertID() string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	return hex.EncodeToString(hash[:4])
}

// RiskCalculator calculates security risk scores
type RiskCalculator struct {
	config SecurityConfig
}

func NewRiskCalculator(config SecurityConfig) *RiskCalculator {
	return &RiskCalculator{config: config}
}

func (rc *RiskCalculator) CalculateRisk(input string, metadata map[string]interface{}, violations []SecurityViolation) float64 {
	var totalRisk float64
	
	// Base risk from input characteristics
	totalRisk += rc.calculateInputRisk(input)
	
	// Risk from metadata
	totalRisk += rc.calculateMetadataRisk(metadata)
	
	// Risk from violations
	for _, violation := range violations {
		switch violation.Severity {
		case SeverityCritical:
			totalRisk += 0.5
		case SeverityHigh:
			totalRisk += 0.3
		case SeverityMedium:
			totalRisk += 0.2
		case SeverityLow:
			totalRisk += 0.1
		}
	}
	
	// Normalize to 0-1 range
	if totalRisk > 1.0 {
		totalRisk = 1.0
	}
	
	return totalRisk
}

func (rc *RiskCalculator) calculateInputRisk(input string) float64 {
	var risk float64
	
	// Length-based risk
	if len(input) > rc.config.MaxInputLength/2 {
		risk += 0.1
	}
	if len(input) > rc.config.MaxInputLength {
		risk += 0.2
	}
	
	// Pattern-based risk
	suspiciousPatterns := []string{
		`\$\(.*\)`,   // Command substitution
		"`.*`",       // Backticks
		`\|`,         // Pipes
		`>`,          // Redirects
		`&`,          // Background execution
	}
	
	for _, pattern := range suspiciousPatterns {
		if matched, _ := regexp.MatchString(pattern, input); matched {
			risk += 0.1
		}
	}
	
	return risk
}

func (rc *RiskCalculator) calculateMetadataRisk(metadata map[string]interface{}) float64 {
	var risk float64
	
	// Check for suspicious metadata
	if source, ok := metadata["source"]; ok {
		if str, ok := source.(string); ok {
			if strings.Contains(str, "external") || strings.Contains(str, "untrusted") {
				risk += 0.2
			}
		}
	}
	
	return risk
}

// ThreatDetector detects various threat patterns
type ThreatDetector struct {
	config SecurityConfig
	patterns map[ViolationType][]*regexp.Regexp
}

func NewThreatDetector(config SecurityConfig) *ThreatDetector {
	td := &ThreatDetector{
		config:   config,
		patterns: make(map[ViolationType][]*regexp.Regexp),
	}
	
	td.initializePatterns()
	return td
}

func (td *ThreatDetector) initializePatterns() {
	// Command injection patterns
	commandPatterns := []string{
		`rm\s+-rf\s+/`,
		`sudo\s+`,
		`wget\s+.*\|\s*sh`,
		`curl\s+.*\|\s*bash`,
		`nc\s+.*\s+\d+`,
	}
	
	for _, pattern := range commandPatterns {
		if regex, err := regexp.Compile(pattern); err == nil {
			td.patterns[ViolationCommandInjection] = append(td.patterns[ViolationCommandInjection], regex)
		}
	}
	
	// Path traversal patterns
	pathPatterns := []string{
		`\.\.\/`,
		`\.\.\\`,
		`\/etc\/passwd`,
		`\/etc\/shadow`,
	}
	
	for _, pattern := range pathPatterns {
		if regex, err := regexp.Compile(pattern); err == nil {
			td.patterns[ViolationPathTraversal] = append(td.patterns[ViolationPathTraversal], regex)
		}
	}
}

func (td *ThreatDetector) DetectThreats(input string, metadata map[string]interface{}) []SecurityViolation {
	var violations []SecurityViolation
	
	for violationType, patterns := range td.patterns {
		for _, pattern := range patterns {
			if pattern.MatchString(input) {
				violation := SecurityViolation{
					ID:          generateViolationID(violationType, input),
					Type:        violationType,
					Severity:    td.calculateSeverity(violationType),
					Description: fmt.Sprintf("Detected %s pattern: %s", violationType, pattern.String()),
					Input:       input,
					Timestamp:   time.Now(),
					Source:      "threat_detector",
					RiskImpact:  td.calculateRiskImpact(violationType),
				}
				violations = append(violations, violation)
			}
		}
	}
	
	return violations
}

func (td *ThreatDetector) calculateSeverity(violationType ViolationType) ViolationSeverity {
	switch violationType {
	case ViolationCommandInjection:
		return SeverityCritical
	case ViolationPathTraversal:
		return SeverityHigh
	case ViolationResourceAbuse:
		return SeverityHigh
	default:
		return SeverityMedium
	}
}

func (td *ThreatDetector) calculateRiskImpact(violationType ViolationType) float64 {
	switch violationType {
	case ViolationCommandInjection:
		return 0.8
	case ViolationPathTraversal:
		return 0.6
	case ViolationResourceAbuse:
		return 0.5
	default:
		return 0.3
	}
}

func generateViolationID(violationType ViolationType, input string) string {
	data := fmt.Sprintf("%s:%s:%d", violationType, input, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:4])
}

// Add RiskImpact field to SecurityViolation
type SecurityViolationExtended struct {
	SecurityViolation
	RiskImpact float64 `json:"risk_impact"`
}