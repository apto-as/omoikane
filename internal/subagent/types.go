// Package subagent provides a flexible agent system inspired by Claude Code's subagents
// Springfield: "柔軟で拡張可能なアーキテクチャを提供します"
// Krukai: "パフォーマンスと型安全性を最優先に設計したわ"
// Vector: "……全ての実行パスにセキュリティチェックを配置……"
package subagent

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Agent represents a specialized AI agent with specific capabilities
// Generic type T allows for type-safe result handling
type Agent[T any] interface {
	// Core identification
	Name() string
	Description() string
	Version() string
	
	// Execution
	Execute(ctx context.Context, input string) (T, error)
	ExecuteWithOptions(ctx context.Context, input string, opts ExecutionOptions) (T, error)
	
	// Capabilities
	SupportedTools() []string
	RequiredPermissions() []Permission
	
	// Lifecycle
	Initialize(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

// SubagentDefinition defines a specialized agent configuration
type SubagentDefinition struct {
	Name        string            `json:"name" yaml:"name"`
	Description string            `json:"description" yaml:"description"`
	Version     string            `json:"version" yaml:"version"`
	Tools       []string          `json:"tools" yaml:"tools"`
	Model       string            `json:"model,omitempty" yaml:"model,omitempty"`
	Color       string            `json:"color,omitempty" yaml:"color,omitempty"`
	Prompt      string            `json:"prompt" yaml:"prompt"`
	Metadata    map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	
	// Trinity-specific fields
	TrinityRole    TrinityRole    `json:"trinity_role,omitempty" yaml:"trinity_role,omitempty"`
	SecurityLevel  SecurityLevel  `json:"security_level" yaml:"security_level"`
	MaxConcurrency int            `json:"max_concurrency" yaml:"max_concurrency"`
	Timeout        time.Duration  `json:"timeout" yaml:"timeout"`
}

// TrinityRole represents the role in Trinity system
type TrinityRole string

const (
	TrinityRoleStrategist TrinityRole = "strategist" // Springfield
	TrinityRoleOptimizer  TrinityRole = "optimizer"  // Krukai
	TrinityRoleAuditor    TrinityRole = "auditor"    // Vector
	TrinityRoleCoordinator TrinityRole = "coordinator" // Integration
)

// SecurityLevel defines the security requirements
// Vector: "……各レベルに応じた防御策を実装……"
type SecurityLevel int

const (
	SecurityLevelPublic    SecurityLevel = 0 // No sensitive data
	SecurityLevelInternal  SecurityLevel = 1 // Internal use only
	SecurityLevelConfidential SecurityLevel = 2 // Confidential data
	SecurityLevelRestricted SecurityLevel = 3 // Highly restricted
)

// Permission represents a required permission for agent execution
type Permission struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
	Scope    string `json:"scope,omitempty"`
}

// ExecutionOptions provides runtime configuration for agent execution
type ExecutionOptions struct {
	Timeout        time.Duration     `json:"timeout,omitempty"`
	MaxRetries     int               `json:"max_retries,omitempty"`
	Priority       TaskPriority      `json:"priority,omitempty"`
	ResourceLimits *ResourceLimit    `json:"resource_limits,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	ParentContext  *ExecutionContext `json:"-"` // Not serialized
}


// ResourceLimit defines resource constraints
// Vector: "……リソース制限で暴走を防ぐ……"
type ResourceLimit struct {
	MaxMemoryMB   int           `json:"max_memory_mb"`
	MaxCPUPercent float64       `json:"max_cpu_percent"`
	MaxGoroutines int           `json:"max_goroutines"`
	MaxDuration   time.Duration `json:"max_duration"`
}

// Manager manages all subagents
type Manager interface {
	// Registration
	Register(def SubagentDefinition) error
	RegisterWithValidator(def SubagentDefinition, validator AgentValidator) error
	Unregister(name string) error
	
	// Discovery
	Get(name string) (*SubagentDefinition, error)
	List() []SubagentDefinition
	ListByRole(role TrinityRole) []SubagentDefinition
	
	// Execution
	Execute(ctx context.Context, name string, input string) (interface{}, error)
	ExecuteWithOptions(ctx context.Context, name string, input string, opts ExecutionOptions) (interface{}, error)
	
	// Parallel execution
	ExecuteParallel(ctx context.Context, agents []string, input string) (map[string]interface{}, error)
	ExecuteWorkflow(ctx context.Context, workflow *Workflow) (*WorkflowResult, error)
	
	// Lifecycle
	Initialize(ctx context.Context) error
	Shutdown(ctx context.Context) error
	
	// Monitoring
	GetMetrics() *ManagerMetrics
	GetAgentMetrics(name string) (*AgentMetrics, error)
}

// AgentValidator validates agent definitions
type AgentValidator interface {
	Validate(def SubagentDefinition) error
}

// Task represents a unit of work for an agent
// Springfield: "明確で追跡可能なタスク定義"
type Task struct {
	ID          string            `json:"id"`
	AgentName   string            `json:"agent_name"`
	Input       string            `json:"input"`
	Options     ExecutionOptions  `json:"options"`
	Dependencies []string         `json:"dependencies,omitempty"` // Task IDs
	Status      TaskStatus        `json:"status"`
	Result      interface{}       `json:"result,omitempty"`
	Error       string            `json:"error,omitempty"`
	StartedAt   *time.Time        `json:"started_at,omitempty"`
	CompletedAt *time.Time        `json:"completed_at,omitempty"`
	
	// Trinity-specific
	TrinityRequirements *TrinityRequirements `json:"trinity_requirements,omitempty"`
}

// TrinityRequirements specifies Trinity-specific task requirements
type TrinityRequirements struct {
	RequireConsensus bool     `json:"require_consensus"`
	RequiredRoles    []TrinityRole `json:"required_roles"`
	MinimumAgents    int      `json:"minimum_agents"`
}

// TaskStatus represents the current state of a task
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusQueued    TaskStatus = "queued"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

// Workflow represents a complex multi-agent workflow
// Krukai: "効率的な並列実行とタスク管理"
type Workflow struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Tasks       []Task                 `json:"tasks"`
	Parallel    bool                   `json:"parallel"`
	MaxParallel int                    `json:"max_parallel"`
	Strategy    ExecutionStrategy      `json:"strategy"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ExecutionStrategy defines how tasks are executed
type ExecutionStrategy string

const (
	StrategySequential ExecutionStrategy = "sequential"
	StrategyParallel   ExecutionStrategy = "parallel"
	StrategyAdaptive   ExecutionStrategy = "adaptive" // Trinity decides
)

// WorkflowResult contains the results of workflow execution
type WorkflowResult struct {
	WorkflowID   string                 `json:"workflow_id"`
	Status       TaskStatus             `json:"status"`
	Results      map[string]interface{} `json:"results"`      // Task ID -> Result
	Errors       map[string]string      `json:"errors,omitempty"` // Task ID -> Error
	StartedAt    time.Time              `json:"started_at"`
	CompletedAt  time.Time              `json:"completed_at"`
	Duration     time.Duration          `json:"duration"`
	
	// Trinity synthesis
	TrinitySynthesis *TrinitySynthesis `json:"trinity_synthesis,omitempty"`
}

// TrinitySynthesis represents the combined analysis from Trinity
type TrinitySynthesis struct {
	Strategic   string            `json:"strategic"`   // Springfield's view
	Technical   string            `json:"technical"`   // Krukai's view
	Security    string            `json:"security"`    // Vector's view
	Consensus   string            `json:"consensus"`   // Combined recommendation
	Confidence  float64           `json:"confidence"`  // 0.0 to 1.0
	Risks       []Risk            `json:"risks,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// Risk represents an identified risk
// Vector: "……全てのリスクを識別し、記録する……"
type Risk struct {
	ID          string       `json:"id"`
	Level       RiskLevel    `json:"level"`
	Category    string       `json:"category"`
	Description string       `json:"description"`
	Mitigation  string       `json:"mitigation,omitempty"`
	Source      TrinityRole  `json:"source"` // Who identified it
}

// RiskLevel defines the severity of a risk
type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "low"
	RiskLevelMedium   RiskLevel = "medium"
	RiskLevelHigh     RiskLevel = "high"
	RiskLevelCritical RiskLevel = "critical"
)

// ExecutionContext maintains context across agent executions
type ExecutionContext struct {
	ID            string
	SessionID     string
	UserID        string
	ProjectID     string
	
	// Runtime state
	Variables     map[string]interface{}
	Secrets       map[string]string // Encrypted
	
	// Tracking
	ExecutionPath []string
	StartTime     time.Time
	
	// Concurrency control
	mu            sync.RWMutex
	cancelFunc    context.CancelFunc
}

// ManagerMetrics provides metrics for the manager
type ManagerMetrics struct {
	TotalAgents      int                    `json:"total_agents"`
	ActiveExecutions int                    `json:"active_executions"`
	TotalExecutions  int64                  `json:"total_executions"`
	FailedExecutions int64                  `json:"failed_executions"`
	AverageLatency   time.Duration          `json:"average_latency"`
	AgentMetrics     map[string]*AgentMetrics `json:"agent_metrics"`
}

// AgentMetrics provides metrics for a specific agent
// Krukai: "パフォーマンスメトリクスで最適化ポイントを特定"
type AgentMetrics struct {
	Name             string        `json:"name"`
	TotalExecutions  int64         `json:"total_executions"`
	SuccessfulExecs  int64         `json:"successful_execs"`
	FailedExecs      int64         `json:"failed_execs"`
	AverageLatency   time.Duration `json:"average_latency"`
	P95Latency       time.Duration `json:"p95_latency"`
	P99Latency       time.Duration `json:"p99_latency"`
	LastExecuted     *time.Time    `json:"last_executed,omitempty"`
	ErrorRate        float64       `json:"error_rate"`
	
	// Quality metrics
	QualityScore     float64       `json:"quality_score"`     // 0.0 to 1.0
	ConsistencyScore float64       `json:"consistency_score"` // 0.0 to 1.0
}

// Errors
var (
	ErrAgentNotFound      = fmt.Errorf("agent not found")
	ErrAgentAlreadyExists = fmt.Errorf("agent already exists")
	ErrInvalidDefinition  = fmt.Errorf("invalid agent definition")
	ErrExecutionTimeout   = fmt.Errorf("execution timeout")
	ErrResourceLimit      = fmt.Errorf("resource limit exceeded")
	ErrPermissionDenied   = fmt.Errorf("permission denied")
	ErrWorkflowFailed     = fmt.Errorf("workflow execution failed")
)

// ValidationError provides detailed validation errors
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error on %s: %s", e.Field, e.Message)
}

// ExecutionError provides detailed execution errors
type ExecutionError struct {
	AgentName string    `json:"agent_name"`
	TaskID    string    `json:"task_id,omitempty"`
	Stage     string    `json:"stage"`
	Message   string    `json:"message"`
	Cause     error     `json:"-"`
	Timestamp time.Time `json:"timestamp"`
}

func (e ExecutionError) Error() string {
	return fmt.Sprintf("[%s] execution error at %s: %s", e.AgentName, e.Stage, e.Message)
}

func (e ExecutionError) Unwrap() error {
	return e.Cause
}

// TrinityCoordinator coordinates Trinity agent execution
// Springfield: "調和の取れた協調動作を実現します"
type TrinityCoordinator interface {
	// Core Trinity operations
	AnalyzeWithTrinity(ctx context.Context, input string) (*TrinitySynthesis, error)
	
	// Individual agent access
	ExecuteSpringfield(ctx context.Context, input string) (string, error)
	ExecuteKrukai(ctx context.Context, input string) (string, error)
	ExecuteVector(ctx context.Context, input string) (string, error)
	
	// Advanced coordination
	ExecuteWithConsensus(ctx context.Context, input string, threshold float64) (*TrinitySynthesis, error)
	ExecuteWithPriority(ctx context.Context, input string, priority TrinityRole) (*TrinitySynthesis, error)
	
	// Monitoring
	GetTrinityStatus() *TrinityStatus
}

// TrinityStatus represents the current status of Trinity system
type TrinityStatus struct {
	Springfield AgentStatus `json:"springfield"`
	Krukai      AgentStatus `json:"krukai"`
	Vector      AgentStatus `json:"vector"`
	Consensus   float64     `json:"consensus_level"`
	LastSync    time.Time   `json:"last_sync"`
}

// AgentStatus represents individual agent status
type AgentStatus struct {
	Available bool          `json:"available"`
	Healthy   bool          `json:"healthy"`
	Load      float64       `json:"load"` // 0.0 to 1.0
	Latency   time.Duration `json:"latency"`
}

// Initialize package-level defaults
func init() {
	// Springfield: "初期化処理で環境を整えます"
	// Krukai: "効率的なデフォルト値を設定"
	// Vector: "……セキュリティ設定を確認……"
}