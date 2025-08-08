package subagent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoaderBasic(t *testing.T) {
	// Springfield: "YAMLローダーの基本機能を検証"
	loader := NewLoader()
	
	// Test loading from string
	yamlContent := `
name: test-agent
description: Test agent for loader
version: 1.0.0
tools:
  - read
  - write
model: gpt-4
color: blue
prompt: |
  You are a test agent.
  Help with testing.
metadata:
  author: Trinity
  category: testing
`
	
	def, err := loader.LoadFromString(yamlContent)
	if err != nil {
		t.Fatalf("Failed to load from string: %v", err)
	}
	
	// Verify basic fields
	if def.Name != "test-agent" {
		t.Errorf("Expected name 'test-agent', got '%s'", def.Name)
	}
	
	if def.Description != "Test agent for loader" {
		t.Errorf("Expected description mismatch")
	}
	
	if def.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", def.Version)
	}
	
	if len(def.Tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(def.Tools))
	}
	
	if def.Model != "gpt-4" {
		t.Errorf("Expected model 'gpt-4', got '%s'", def.Model)
	}
	
	if def.Color != "blue" {
		t.Errorf("Expected color 'blue', got '%s'", def.Color)
	}
	
	if !strings.Contains(def.Prompt, "test agent") {
		t.Errorf("Prompt not loaded correctly")
	}
	
	if def.Metadata["author"] != "Trinity" {
		t.Errorf("Metadata not loaded correctly")
	}
}

func TestLoaderTrinityConfig(t *testing.T) {
	// Krukai: "Trinity拡張設定を検証"
	loader := LoaderWithTrinity()
	
	yamlContent := `
name: trinity-agent
description: Trinity-enhanced agent
version: 2.0.0
tools:
  - analyze
  - optimize
prompt: Trinity agent prompt
trinity:
  role: optimizer
  security_level: confidential
  max_concurrency: 5
  timeout: 1m30s
  requirements:
    - high_performance
    - security_clearance
  consensus:
    required: true
    min_agreement: 0.8
    required_roles:
      - strategist
      - auditor
`
	
	def, err := loader.LoadFromString(yamlContent)
	if err != nil {
		t.Fatalf("Failed to load Trinity config: %v", err)
	}
	
	// Verify Trinity configuration
	if def.TrinityRole != TrinityRoleOptimizer {
		t.Errorf("Expected optimizer role, got %s", def.TrinityRole)
	}
	
	if def.SecurityLevel != SecurityLevelConfidential {
		t.Errorf("Expected confidential security level, got %v", def.SecurityLevel)
	}
	
	if def.MaxConcurrency != 5 {
		t.Errorf("Expected max concurrency 5, got %d", def.MaxConcurrency)
	}
	
	expectedTimeout := 90 * time.Second
	if def.Timeout != expectedTimeout {
		t.Errorf("Expected timeout %v, got %v", expectedTimeout, def.Timeout)
	}
}

func TestLoaderValidation(t *testing.T) {
	// Vector: "……設定の妥当性検証を確認……"
	loader := NewLoader()
	
	// Test missing name
	yamlContent := `
description: No name agent
prompt: Test prompt
`
	
	_, err := loader.LoadFromString(yamlContent)
	if err == nil {
		t.Error("Expected validation error for missing name")
	}
	
	// Test missing description
	yamlContent = `
name: test-agent
prompt: Test prompt
`
	
	_, err = loader.LoadFromString(yamlContent)
	if err == nil {
		t.Error("Expected validation error for missing description")
	}
	
	// Test missing prompt
	yamlContent = `
name: test-agent
description: Test agent
`
	
	_, err = loader.LoadFromString(yamlContent)
	if err == nil {
		t.Error("Expected validation error for missing prompt")
	}
	
	// Test invalid Trinity configuration
	loader = LoaderWithTrinity()
	
	yamlContent = `
name: invalid-trinity
description: Invalid Trinity config
prompt: Test
trinity:
  role: invalid_role
`
	
	_, err = loader.LoadFromString(yamlContent)
	if err == nil {
		t.Error("Expected validation error for invalid Trinity role")
	}
	
	yamlContent = `
name: invalid-security
description: Invalid security level
prompt: Test
trinity:
  security_level: top_secret
`
	
	_, err = loader.LoadFromString(yamlContent)
	if err == nil {
		t.Error("Expected validation error for invalid security level")
	}
	
	yamlContent = `
name: invalid-consensus
description: Invalid consensus config
prompt: Test
trinity:
  consensus:
    min_agreement: 1.5
`
	
	_, err = loader.LoadFromString(yamlContent)
	if err == nil {
		t.Error("Expected validation error for invalid consensus agreement")
	}
}

func TestLoaderFileOperations(t *testing.T) {
	// Springfield: "ファイル操作の動作を検証"
	
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "loader-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create test agent files
	testAgents := map[string]string{
		"agent1.yaml": `
name: agent-one
description: First test agent
version: 1.0.0
tools: [read]
prompt: Agent one prompt
trinity:
  role: strategist
`,
		"agent2.yml": `
name: agent-two
description: Second test agent
version: 1.0.0
tools: [write]
prompt: Agent two prompt
trinity:
  role: optimizer
`,
		"not-yaml.txt": "This is not a YAML file",
	}
	
	for filename, content := range testAgents {
		path := filepath.Join(tempDir, filename)
		err := os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file %s: %v", filename, err)
		}
	}
	
	// Test loading from directory
	loader := NewLoader(tempDir)
	agents, err := loader.LoadDirectory(tempDir)
	if err != nil {
		t.Fatalf("Failed to load directory: %v", err)
	}
	
	// Should load 2 YAML files, ignore the .txt file
	if len(agents) != 2 {
		t.Errorf("Expected 2 agents, got %d", len(agents))
	}
	
	// Verify agents loaded
	agentNames := make(map[string]bool)
	for _, agent := range agents {
		agentNames[agent.Name] = true
	}
	
	if !agentNames["agent-one"] {
		t.Error("agent-one not loaded")
	}
	
	if !agentNames["agent-two"] {
		t.Error("agent-two not loaded")
	}
	
	// Test LoadAll with custom search path
	allAgents, err := loader.LoadAll()
	if err != nil {
		t.Fatalf("Failed to load all agents: %v", err)
	}
	
	if len(allAgents) < 2 {
		t.Errorf("Expected at least 2 agents from LoadAll, got %d", len(allAgents))
	}
	
	// Test loading specific file
	agentPath := filepath.Join(tempDir, "agent1.yaml")
	def, err := loader.LoadFromFile(agentPath)
	if err != nil {
		t.Fatalf("Failed to load from file: %v", err)
	}
	
	if def.Name != "agent-one" {
		t.Errorf("Expected agent-one, got %s", def.Name)
	}
	
	if def.TrinityRole != TrinityRoleStrategist {
		t.Errorf("Expected strategist role, got %s", def.TrinityRole)
	}
}

func TestLoaderCache(t *testing.T) {
	// Krukai: "キャッシュ機能の効率性を検証"
	
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "cache-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create test agent file
	agentContent := `
name: cached-agent
description: Agent for cache testing
version: 1.0.0
tools: [cache]
prompt: Cache test
`
	
	agentPath := filepath.Join(tempDir, "cached-agent.yaml")
	err = os.WriteFile(agentPath, []byte(agentContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write agent file: %v", err)
	}
	
	loader := NewLoader(tempDir)
	
	// First load - should read from file
	def1, err := loader.LoadAgent("cached-agent")
	if err != nil {
		t.Fatalf("Failed to load agent: %v", err)
	}
	
	// Second load - should use cache
	def2, err := loader.LoadAgent("cached-agent")
	if err != nil {
		t.Fatalf("Failed to load cached agent: %v", err)
	}
	
	// Verify same instance (cached)
	if def1 != def2 {
		t.Error("Expected cached instance to be returned")
	}
	
	// Clear cache
	loader.ClearCache()
	
	// Load again - should read from file
	def3, err := loader.LoadAgent("cached-agent")
	if err != nil {
		t.Fatalf("Failed to load agent after cache clear: %v", err)
	}
	
	// Should be different instance after cache clear
	if def1 == def3 {
		t.Error("Expected new instance after cache clear")
	}
}

func TestLoaderRoleParsing(t *testing.T) {
	// Vector: "……役割の正確な解析を検証……"
	tests := []struct {
		input    string
		expected TrinityRole
	}{
		{"strategist", TrinityRoleStrategist},
		{"springfield", TrinityRoleStrategist},
		{"optimizer", TrinityRoleOptimizer},
		{"krukai", TrinityRoleOptimizer},
		{"auditor", TrinityRoleAuditor},
		{"vector", TrinityRoleAuditor},
		{"coordinator", TrinityRoleCoordinator},
		{"trinity", TrinityRoleCoordinator},
		{"invalid", ""}, // Should use default
		{"", ""},        // Should use default
	}
	
	for _, test := range tests {
		result := parseTrinityRole(test.input)
		if result != test.expected {
			t.Errorf("parseTrinityRole(%s) = %s, expected %s", 
				test.input, result, test.expected)
		}
	}
}

func TestLoaderSecurityLevelParsing(t *testing.T) {
	// Vector: "……セキュリティレベルの解析を確認……"
	tests := []struct {
		input    string
		expected SecurityLevel
	}{
		{"public", SecurityLevelPublic},
		{"internal", SecurityLevelInternal},
		{"confidential", SecurityLevelConfidential},
		{"restricted", SecurityLevelRestricted},
		{"invalid", 0}, // Should use default
		{"", 0},        // Should use default
	}
	
	for _, test := range tests {
		result := parseSecurityLevel(test.input)
		if result != test.expected {
			t.Errorf("parseSecurityLevel(%s) = %v, expected %v",
				test.input, result, test.expected)
		}
	}
}

func TestLoaderDefaults(t *testing.T) {
	// Springfield: "デフォルト値の適用を検証"
	loader := NewLoader()
	
	// Minimal configuration
	yamlContent := `
name: minimal-agent
description: Minimal configuration
prompt: Basic prompt
`
	
	def, err := loader.LoadFromString(yamlContent)
	if err != nil {
		t.Fatalf("Failed to load minimal config: %v", err)
	}
	
	// Check defaults were applied
	if def.TrinityRole != TrinityRoleOptimizer {
		t.Errorf("Expected default role optimizer, got %s", def.TrinityRole)
	}
	
	if def.SecurityLevel != SecurityLevelInternal {
		t.Errorf("Expected default security level internal, got %v", def.SecurityLevel)
	}
	
	if def.Timeout != 30*time.Second {
		t.Errorf("Expected default timeout 30s, got %v", def.Timeout)
	}
	
	if def.MaxConcurrency != 1 {
		t.Errorf("Expected default max concurrency 1, got %d", def.MaxConcurrency)
	}
}