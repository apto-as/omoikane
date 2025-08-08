// Package subagent implements the YAML-based agent configuration loader
// Springfield: "設定ファイルから柔軟にagentを読み込みます"
// Krukai: "効率的なパースと検証処理を実装"
// Vector: "……設定の妥当性を厳密にチェック……"
package subagent

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// YAMLConfig represents the raw YAML configuration structure
// Compatible with Claude Code agent.yaml format
type YAMLConfig struct {
	// Claude Code compatible fields
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Version     string            `yaml:"version"`
	Tools       []string          `yaml:"tools"`
	Model       string            `yaml:"model,omitempty"`
	Color       string            `yaml:"color,omitempty"`
	Prompt      string            `yaml:"prompt"`
	Metadata    map[string]string `yaml:"metadata,omitempty"`
	
	// Trinity-specific extensions
	Trinity *TrinityConfig `yaml:"trinity,omitempty"`
}

// TrinityConfig contains Trinity-specific configuration
type TrinityConfig struct {
	Role           string        `yaml:"role"`           // strategist, optimizer, auditor
	SecurityLevel  string        `yaml:"security_level"` // public, internal, confidential, restricted
	MaxConcurrency int           `yaml:"max_concurrency"`
	Timeout        string        `yaml:"timeout"`        // Duration string
	Requirements   []string      `yaml:"requirements,omitempty"`
	Consensus      *ConsensusConfig `yaml:"consensus,omitempty"`
}

// ConsensusConfig defines Trinity consensus requirements
type ConsensusConfig struct {
	Required      bool     `yaml:"required"`
	MinAgreement  float64  `yaml:"min_agreement"`  // 0.0 to 1.0
	RequiredRoles []string `yaml:"required_roles"`
}

// Loader handles loading and parsing of agent configurations
type Loader struct {
	searchPaths []string
	validators  []ConfigValidator
	cache       map[string]*SubagentDefinition
}

// ConfigValidator validates loaded configurations
type ConfigValidator interface {
	ValidateConfig(config *YAMLConfig) error
}

// NewLoader creates a new configuration loader
func NewLoader(searchPaths ...string) *Loader {
	defaultPaths := []string{
		"agents",
		".agents",
		"config/agents",
		filepath.Join(os.Getenv("HOME"), ".omoikane", "agents"),
	}
	
	if len(searchPaths) > 0 {
		defaultPaths = append(searchPaths, defaultPaths...)
	}
	
	return &Loader{
		searchPaths: defaultPaths,
		validators:  []ConfigValidator{&defaultConfigValidator{}},
		cache:       make(map[string]*SubagentDefinition),
	}
}

// LoadAgent loads a single agent configuration by name
// Springfield: "agent設定を慎重に読み込みます"
func (l *Loader) LoadAgent(name string) (*SubagentDefinition, error) {
	// Check cache first
	if cached, exists := l.cache[name]; exists {
		return cached, nil
	}
	
	// Search for agent configuration file
	configPath, err := l.findAgentConfig(name)
	if err != nil {
		return nil, fmt.Errorf("agent config not found: %w", err)
	}
	
	// Load from file
	def, err := l.LoadFromFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load agent from %s: %w", configPath, err)
	}
	
	// Cache the result
	l.cache[name] = def
	
	return def, nil
}

// LoadFromFile loads agent configuration from a specific file
func (l *Loader) LoadFromFile(path string) (*SubagentDefinition, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	
	return l.LoadFromReader(file)
}

// LoadFromReader loads agent configuration from an io.Reader
func (l *Loader) LoadFromReader(r io.Reader) (*SubagentDefinition, error) {
	var config YAMLConfig
	
	decoder := yaml.NewDecoder(r)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	
	// Validate configuration
	for _, validator := range l.validators {
		if err := validator.ValidateConfig(&config); err != nil {
			return nil, fmt.Errorf("validation failed: %w", err)
		}
	}
	
	// Convert to SubagentDefinition
	return l.convertToDefinition(&config)
}

// LoadFromString loads agent configuration from a YAML string
func (l *Loader) LoadFromString(yamlContent string) (*SubagentDefinition, error) {
	return l.LoadFromReader(strings.NewReader(yamlContent))
}

// LoadDirectory loads all agent configurations from a directory
// Krukai: "効率的に複数のagentを一括読み込み"
func (l *Loader) LoadDirectory(dir string) ([]*SubagentDefinition, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}
	
	var agents []*SubagentDefinition
	
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		// Check for YAML files
		name := entry.Name()
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}
		
		path := filepath.Join(dir, name)
		def, err := l.LoadFromFile(path)
		if err != nil {
			// Log error but continue loading other agents
			// In production, this should use proper logging
			fmt.Fprintf(os.Stderr, "Warning: failed to load %s: %v\n", path, err)
			continue
		}
		
		agents = append(agents, def)
	}
	
	return agents, nil
}

// LoadAll loads all available agent configurations
func (l *Loader) LoadAll() ([]*SubagentDefinition, error) {
	var allAgents []*SubagentDefinition
	loaded := make(map[string]bool) // Track loaded agents to avoid duplicates
	
	for _, searchPath := range l.searchPaths {
		// Check if path exists
		info, err := os.Stat(searchPath)
		if err != nil {
			continue // Skip non-existent paths
		}
		
		if !info.IsDir() {
			continue // Skip non-directories
		}
		
		agents, err := l.LoadDirectory(searchPath)
		if err != nil {
			continue // Skip directories with errors
		}
		
		// Add unique agents
		for _, agent := range agents {
			if !loaded[agent.Name] {
				allAgents = append(allAgents, agent)
				loaded[agent.Name] = true
			}
		}
	}
	
	return allAgents, nil
}

// AddValidator adds a custom configuration validator
func (l *Loader) AddValidator(validator ConfigValidator) {
	l.validators = append(l.validators, validator)
}

// ClearCache clears the loaded agent cache
func (l *Loader) ClearCache() {
	l.cache = make(map[string]*SubagentDefinition)
}

// Private helper methods

func (l *Loader) findAgentConfig(name string) (string, error) {
	// Try various file extensions
	extensions := []string{".yaml", ".yml", ""}
	
	for _, searchPath := range l.searchPaths {
		for _, ext := range extensions {
			path := filepath.Join(searchPath, name+ext)
			
			// Also try with "agent" suffix
			alternativePaths := []string{
				path,
				filepath.Join(searchPath, name+"-agent"+ext),
				filepath.Join(searchPath, name+"_agent"+ext),
			}
			
			for _, p := range alternativePaths {
				if _, err := os.Stat(p); err == nil {
					return p, nil
				}
			}
		}
	}
	
	return "", fmt.Errorf("agent configuration '%s' not found in search paths", name)
}

func (l *Loader) convertToDefinition(config *YAMLConfig) (*SubagentDefinition, error) {
	def := &SubagentDefinition{
		Name:        config.Name,
		Description: config.Description,
		Version:     config.Version,
		Tools:       config.Tools,
		Model:       config.Model,
		Color:       config.Color,
		Prompt:      config.Prompt,
		Metadata:    config.Metadata,
	}
	
	// Apply Trinity-specific configuration if present
	if config.Trinity != nil {
		// Parse Trinity role
		def.TrinityRole = parseTrinityRole(config.Trinity.Role)
		
		// Parse security level
		def.SecurityLevel = parseSecurityLevel(config.Trinity.SecurityLevel)
		
		// Set max concurrency
		def.MaxConcurrency = config.Trinity.MaxConcurrency
		if def.MaxConcurrency == 0 {
			def.MaxConcurrency = 1 // Default to 1
		}
		
		// Parse timeout
		if config.Trinity.Timeout != "" {
			timeout, err := time.ParseDuration(config.Trinity.Timeout)
			if err != nil {
				return nil, fmt.Errorf("invalid timeout duration: %w", err)
			}
			def.Timeout = timeout
		}
	}
	
	// Apply defaults if not set
	if def.TrinityRole == "" {
		def.TrinityRole = TrinityRoleOptimizer // Default to Krukai's role
	}
	
	if def.SecurityLevel == 0 {
		def.SecurityLevel = SecurityLevelInternal // Default to internal
	}
	
	if def.Timeout == 0 {
		def.Timeout = 30 * time.Second // Default timeout
	}
	
	if def.MaxConcurrency == 0 {
		def.MaxConcurrency = 1 // Default max concurrency
	}
	
	return def, nil
}

func parseTrinityRole(role string) TrinityRole {
	switch strings.ToLower(role) {
	case "strategist", "springfield":
		return TrinityRoleStrategist
	case "optimizer", "krukai":
		return TrinityRoleOptimizer
	case "auditor", "vector":
		return TrinityRoleAuditor
	case "coordinator", "trinity":
		return TrinityRoleCoordinator
	default:
		return "" // Will use default
	}
}

func parseSecurityLevel(level string) SecurityLevel {
	switch strings.ToLower(level) {
	case "public":
		return SecurityLevelPublic
	case "internal":
		return SecurityLevelInternal
	case "confidential":
		return SecurityLevelConfidential
	case "restricted":
		return SecurityLevelRestricted
	default:
		return 0 // Will use default
	}
}

// defaultConfigValidator provides basic configuration validation
// Vector: "……設定の安全性を確認……"
type defaultConfigValidator struct{}

func (v *defaultConfigValidator) ValidateConfig(config *YAMLConfig) error {
	if config.Name == "" {
		return &ValidationError{
			Field:   "name",
			Message: "agent name is required",
			Code:    "required",
		}
	}
	
	if config.Description == "" {
		return &ValidationError{
			Field:   "description",
			Message: "agent description is required",
			Code:    "required",
		}
	}
	
	if config.Prompt == "" {
		return &ValidationError{
			Field:   "prompt",
			Message: "agent prompt is required",
			Code:    "required",
		}
	}
	
	// Validate Trinity-specific configuration if present
	if config.Trinity != nil {
		if config.Trinity.MaxConcurrency < 0 {
			return &ValidationError{
				Field:   "trinity.max_concurrency",
				Message: "max concurrency must be non-negative",
				Code:    "invalid_value",
			}
		}
		
		if config.Trinity.Consensus != nil {
			if config.Trinity.Consensus.MinAgreement < 0 || config.Trinity.Consensus.MinAgreement > 1 {
				return &ValidationError{
					Field:   "trinity.consensus.min_agreement",
					Message: "min agreement must be between 0 and 1",
					Code:    "invalid_value",
				}
			}
		}
	}
	
	return nil
}

// LoaderWithTrinity creates a loader with Trinity-specific validators
// Springfield: "Trinity専用の高度な検証を提供します"
func LoaderWithTrinity() *Loader {
	loader := NewLoader()
	loader.AddValidator(&trinityValidator{})
	return loader
}

// trinityValidator provides Trinity-specific validation
type trinityValidator struct{}

func (v *trinityValidator) ValidateConfig(config *YAMLConfig) error {
	if config.Trinity == nil {
		return nil // Trinity config is optional
	}
	
	// Validate role
	validRoles := map[string]bool{
		"strategist":  true,
		"springfield": true,
		"optimizer":   true,
		"krukai":      true,
		"auditor":     true,
		"vector":      true,
		"coordinator": true,
		"trinity":     true,
	}
	
	if config.Trinity.Role != "" && !validRoles[strings.ToLower(config.Trinity.Role)] {
		return &ValidationError{
			Field:   "trinity.role",
			Message: fmt.Sprintf("invalid Trinity role: %s", config.Trinity.Role),
			Code:    "invalid_value",
		}
	}
	
	// Validate security level
	validLevels := map[string]bool{
		"public":       true,
		"internal":     true,
		"confidential": true,
		"restricted":   true,
	}
	
	if config.Trinity.SecurityLevel != "" && !validLevels[strings.ToLower(config.Trinity.SecurityLevel)] {
		return &ValidationError{
			Field:   "trinity.security_level",
			Message: fmt.Sprintf("invalid security level: %s", config.Trinity.SecurityLevel),
			Code:    "invalid_value",
		}
	}
	
	// Validate consensus configuration
	if config.Trinity.Consensus != nil {
		if len(config.Trinity.Consensus.RequiredRoles) > 0 {
			for _, role := range config.Trinity.Consensus.RequiredRoles {
				if !validRoles[strings.ToLower(role)] {
					return &ValidationError{
						Field:   "trinity.consensus.required_roles",
						Message: fmt.Sprintf("invalid role in consensus: %s", role),
						Code:    "invalid_value",
					}
				}
			}
		}
	}
	
	return nil
}