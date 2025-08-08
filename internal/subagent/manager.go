// Package subagent implements the subagent management system
// Springfield: "エージェント間の調和を保ちながら管理します"
// Krukai: "効率的な並列実行と状態管理を実装"
// Vector: "……全ての実行を監視し、異常を検出……"
package subagent

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

// manager implements the Manager interface with high-performance optimizations
// Krukai: "完璧に最適化されたmanager実装"
type manager struct {
	// Agent registry with optimized access
	registry   map[string]*SubagentDefinition
	registryMu sync.RWMutex

	// Agent validators
	validators []AgentValidator

	// High-performance execution control
	semaphore    *semaphore.Weighted // Limit concurrent executions
	maxWorkers   int64
	goroutinePool *goroutinePool     // Custom goroutine pool

	// Optimized metrics with atomic operations
	metrics    *ManagerMetrics
	metricsMu  sync.RWMutex

	// Agent-specific metrics with lock-free operations
	agentMetrics   map[string]*AgentMetrics
	agentMetricsMu sync.RWMutex

	// Object pooling for memory efficiency
	execContextPool sync.Pool
	resultPool      sync.Pool

	// Performance monitoring
	performanceTracker *PerformanceTracker

	// Lifecycle with enhanced control
	ctx        context.Context
	cancelFunc context.CancelFunc
	wg         sync.WaitGroup
	healthCheck *healthChecker

	// Configuration
	config ManagerConfig
}

// ManagerConfig contains configuration for the manager
type ManagerConfig struct {
	MaxConcurrentAgents int           `json:"max_concurrent_agents" yaml:"max_concurrent_agents"`
	DefaultTimeout      time.Duration `json:"default_timeout" yaml:"default_timeout"`
	EnableMetrics       bool          `json:"enable_metrics" yaml:"enable_metrics"`
	MetricsInterval     time.Duration `json:"metrics_interval" yaml:"metrics_interval"`
	
	// Trinity-specific
	TrinityMode     bool `json:"trinity_mode" yaml:"trinity_mode"`
	RequireConsensus bool `json:"require_consensus" yaml:"require_consensus"`
}

// DefaultManagerConfig returns optimized default configuration
// Krukai: "パフォーマンス最適化済みのデフォルト設定"
func DefaultManagerConfig() ManagerConfig {
	return ManagerConfig{
		MaxConcurrentAgents: 50, // Krukai enhancement: 50 agents support
		DefaultTimeout:      30 * time.Second,
		EnableMetrics:       true,
		MetricsInterval:     30 * time.Second, // More frequent metrics
		TrinityMode:        true,
		RequireConsensus:   false,
	}
}

// NewManager creates a new high-performance subagent manager
// Krukai: "最高効率の並列実行マネージャーを構築"
func NewManager(config ManagerConfig) Manager {
	ctx, cancel := context.WithCancel(context.Background())

	m := &manager{
		registry:     make(map[string]*SubagentDefinition),
		validators:   []AgentValidator{},
		semaphore:    semaphore.NewWeighted(int64(config.MaxConcurrentAgents)),
		maxWorkers:   int64(config.MaxConcurrentAgents),
		goroutinePool: newGoroutinePool(int(config.MaxConcurrentAgents)),
		metrics:      &ManagerMetrics{
			AgentMetrics: make(map[string]*AgentMetrics),
		},
		agentMetrics: make(map[string]*AgentMetrics),
		ctx:          ctx,
		cancelFunc:   cancel,
		config:       config,
		performanceTracker: newPerformanceTracker(),
		healthCheck: newHealthChecker(config.MaxConcurrentAgents),
	}

	// Initialize object pools for memory efficiency
	m.execContextPool.New = func() interface{} {
		return &ExecutionContext{
			Variables: make(map[string]interface{}),
			Secrets:   make(map[string]string),
		}
	}
	
	m.resultPool.New = func() interface{} {
		return make(map[string]interface{})
	}

	// Add default validators
	m.validators = append(m.validators, &defaultValidator{})

	// Start background services
	if config.EnableMetrics {
		m.wg.Add(3)
		go m.metricsCollector()
		go m.performanceMonitor()
		go m.healthMonitor()
	}

	return m
}

// Register registers a new subagent
func (m *manager) Register(def SubagentDefinition) error {
	return m.RegisterWithValidator(def, nil)
}

// RegisterWithValidator registers a new subagent with custom validator
func (m *manager) RegisterWithValidator(def SubagentDefinition, validator AgentValidator) error {
	// Validate definition
	if err := m.validateDefinition(def); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Custom validation if provided
	if validator != nil {
		if err := validator.Validate(def); err != nil {
			return fmt.Errorf("custom validation failed: %w", err)
		}
	}

	// Register the agent
	m.registryMu.Lock()
	defer m.registryMu.Unlock()

	if _, exists := m.registry[def.Name]; exists {
		return ErrAgentAlreadyExists
	}

	// Set defaults
	if def.MaxConcurrency == 0 {
		def.MaxConcurrency = 1
	}
	if def.Timeout == 0 {
		def.Timeout = m.config.DefaultTimeout
	}

	// Store definition
	m.registry[def.Name] = &def

	// Initialize metrics
	if m.config.EnableMetrics {
		m.initAgentMetrics(def.Name)
	}

	return nil
}

// Unregister removes a subagent
func (m *manager) Unregister(name string) error {
	m.registryMu.Lock()
	defer m.registryMu.Unlock()

	if _, exists := m.registry[name]; !exists {
		return ErrAgentNotFound
	}

	delete(m.registry, name)
	
	// Clean up metrics
	m.agentMetricsMu.Lock()
	delete(m.agentMetrics, name)
	m.agentMetricsMu.Unlock()

	return nil
}

// Get retrieves a subagent definition
func (m *manager) Get(name string) (*SubagentDefinition, error) {
	m.registryMu.RLock()
	defer m.registryMu.RUnlock()

	def, exists := m.registry[name]
	if !exists {
		return nil, ErrAgentNotFound
	}

	// Return a copy to prevent external modification
	defCopy := *def
	return &defCopy, nil
}

// List returns all registered subagents
func (m *manager) List() []SubagentDefinition {
	m.registryMu.RLock()
	defer m.registryMu.RUnlock()

	agents := make([]SubagentDefinition, 0, len(m.registry))
	for _, def := range m.registry {
		agents = append(agents, *def)
	}

	return agents
}

// ListByRole returns subagents with specific Trinity role
func (m *manager) ListByRole(role TrinityRole) []SubagentDefinition {
	m.registryMu.RLock()
	defer m.registryMu.RUnlock()

	agents := make([]SubagentDefinition, 0)
	for _, def := range m.registry {
		if def.TrinityRole == role {
			agents = append(agents, *def)
		}
	}

	return agents
}

// Execute executes a single agent
func (m *manager) Execute(ctx context.Context, name string, input string) (interface{}, error) {
	return m.ExecuteWithOptions(ctx, name, input, ExecutionOptions{})
}

// ExecuteWithOptions executes a single agent with options
func (m *manager) ExecuteWithOptions(ctx context.Context, name string, input string, opts ExecutionOptions) (interface{}, error) {
	// Get agent definition
	def, err := m.Get(name)
	if err != nil {
		return nil, err
	}

	// Apply timeout
	timeout := def.Timeout
	if opts.Timeout > 0 {
		timeout = opts.Timeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Acquire semaphore
	if err := m.semaphore.Acquire(ctx, 1); err != nil {
		return nil, fmt.Errorf("failed to acquire semaphore: %w", err)
	}
	defer m.semaphore.Release(1)

	// Track metrics
	startTime := time.Now()
	m.incrementExecution(name)

	// Execute agent (simulation for now)
	result, err := m.executeAgent(ctx, def, input, opts)
	
	// Update metrics
	m.updateMetrics(name, time.Since(startTime), err)

	if err != nil {
		return nil, &ExecutionError{
			AgentName: name,
			Stage:     "execution",
			Message:   err.Error(),
			Cause:     err,
			Timestamp: time.Now(),
		}
	}

	return result, nil
}

// ExecuteParallel executes multiple agents in parallel
// Krukai: "最大効率で並列実行を制御"
func (m *manager) ExecuteParallel(ctx context.Context, agents []string, input string) (map[string]interface{}, error) {
	results := make(map[string]interface{})
	resultsMu := sync.Mutex{}

	g, ctx := errgroup.WithContext(ctx)

	for _, agentName := range agents {
		agentName := agentName // Capture loop variable

		g.Go(func() error {
			result, err := m.Execute(ctx, agentName, input)
			if err != nil {
				// Store error but don't fail entire group
				resultsMu.Lock()
				results[agentName] = map[string]interface{}{
					"error": err.Error(),
				}
				resultsMu.Unlock()
				return nil // Continue with other agents
			}

			resultsMu.Lock()
			results[agentName] = result
			resultsMu.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("parallel execution failed: %w", err)
	}

	return results, nil
}

// ExecuteWorkflow executes a complex workflow
// Springfield: "ワークフロー全体を調和させます"
func (m *manager) ExecuteWorkflow(ctx context.Context, workflow *Workflow) (*WorkflowResult, error) {
	result := &WorkflowResult{
		WorkflowID: workflow.ID,
		Status:     TaskStatusRunning,
		Results:    make(map[string]interface{}),
		Errors:     make(map[string]string),
		StartedAt:  time.Now(),
	}

	switch workflow.Strategy {
	case StrategySequential:
		return m.executeSequentialWorkflow(ctx, workflow, result)
	case StrategyParallel:
		return m.executeParallelWorkflow(ctx, workflow, result)
	case StrategyAdaptive:
		return m.executeAdaptiveWorkflow(ctx, workflow, result)
	default:
		return nil, fmt.Errorf("unknown execution strategy: %s", workflow.Strategy)
	}
}

// Initialize initializes the manager
func (m *manager) Initialize(ctx context.Context) error {
	// Initialize any required resources
	return nil
}

// Shutdown shuts down the manager
func (m *manager) Shutdown(ctx context.Context) error {
	// Cancel internal context
	m.cancelFunc()

	// Wait for goroutines with timeout
	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// GetMetrics returns manager metrics
func (m *manager) GetMetrics() *ManagerMetrics {
	m.metricsMu.RLock()
	defer m.metricsMu.RUnlock()

	// Return a copy
	metricsCopy := *m.metrics
	
	// Copy agent metrics
	metricsCopy.AgentMetrics = make(map[string]*AgentMetrics)
	m.agentMetricsMu.RLock()
	for name, metrics := range m.agentMetrics {
		agentMetricsCopy := *metrics
		metricsCopy.AgentMetrics[name] = &agentMetricsCopy
	}
	m.agentMetricsMu.RUnlock()

	return &metricsCopy
}

// GetAgentMetrics returns metrics for a specific agent
func (m *manager) GetAgentMetrics(name string) (*AgentMetrics, error) {
	m.agentMetricsMu.RLock()
	defer m.agentMetricsMu.RUnlock()

	metrics, exists := m.agentMetrics[name]
	if !exists {
		return nil, ErrAgentNotFound
	}

	// Return a copy
	metricsCopy := *metrics
	return &metricsCopy, nil
}

// Private helper methods

func (m *manager) validateDefinition(def SubagentDefinition) error {
	if def.Name == "" {
		return &ValidationError{
			Field:   "name",
			Message: "agent name is required",
			Code:    "required",
		}
	}

	if def.Description == "" {
		return &ValidationError{
			Field:   "description",
			Message: "agent description is required",
			Code:    "required",
		}
	}

	// Run all validators
	for _, validator := range m.validators {
		if err := validator.Validate(def); err != nil {
			return err
		}
	}

	return nil
}

func (m *manager) initAgentMetrics(name string) {
	m.agentMetricsMu.Lock()
	defer m.agentMetricsMu.Unlock()

	m.agentMetrics[name] = &AgentMetrics{
		Name: name,
	}
}

func (m *manager) incrementExecution(name string) {
	m.metricsMu.Lock()
	m.metrics.ActiveExecutions++
	m.metrics.TotalExecutions++
	m.metricsMu.Unlock()

	m.agentMetricsMu.Lock()
	if metrics, exists := m.agentMetrics[name]; exists {
		atomic.AddInt64(&metrics.TotalExecutions, 1)
	}
	m.agentMetricsMu.Unlock()
}

func (m *manager) updateMetrics(name string, duration time.Duration, err error) {
	m.metricsMu.Lock()
	m.metrics.ActiveExecutions--
	if err != nil {
		m.metrics.FailedExecutions++
	}
	m.metricsMu.Unlock()

	m.agentMetricsMu.Lock()
	if metrics, exists := m.agentMetrics[name]; exists {
		if err != nil {
			atomic.AddInt64(&metrics.FailedExecs, 1)
		} else {
			atomic.AddInt64(&metrics.SuccessfulExecs, 1)
		}
		
		// Update latency (simplified - should use proper statistics)
		now := time.Now()
		metrics.LastExecuted = &now
		metrics.AverageLatency = (metrics.AverageLatency + duration) / 2
	}
	m.agentMetricsMu.Unlock()
}

func (m *manager) executeAgent(ctx context.Context, def *SubagentDefinition, input string, opts ExecutionOptions) (interface{}, error) {
	// This is a placeholder for actual agent execution
	// In real implementation, this would:
	// 1. Load the agent implementation
	// 2. Apply resource limits
	// 3. Execute with proper isolation
	// 4. Return structured results

	// Vector: "……実行を監視中……"
	select {
	case <-ctx.Done():
		return nil, ErrExecutionTimeout
	case <-time.After(100 * time.Millisecond): // Simulate work
		// Return mock result based on agent type
		return map[string]interface{}{
			"agent":   def.Name,
			"input":   input,
			"result":  fmt.Sprintf("Processed by %s", def.Name),
			"trinity": def.TrinityRole,
		}, nil
	}
}

func (m *manager) executeSequentialWorkflow(ctx context.Context, workflow *Workflow, result *WorkflowResult) (*WorkflowResult, error) {
	for _, task := range workflow.Tasks {
		taskResult, err := m.Execute(ctx, task.AgentName, task.Input)
		if err != nil {
			result.Errors[task.ID] = err.Error()
			result.Status = TaskStatusFailed
			return result, err
		}
		result.Results[task.ID] = taskResult
	}

	result.Status = TaskStatusCompleted
	result.CompletedAt = time.Now()
	result.Duration = result.CompletedAt.Sub(result.StartedAt)

	return result, nil
}

func (m *manager) executeParallelWorkflow(ctx context.Context, workflow *Workflow, result *WorkflowResult) (*WorkflowResult, error) {
	// Execute all tasks in parallel
	taskNames := make([]string, len(workflow.Tasks))
	for i, task := range workflow.Tasks {
		taskNames[i] = task.AgentName
	}

	results, err := m.ExecuteParallel(ctx, taskNames, workflow.Tasks[0].Input)
	if err != nil {
		result.Status = TaskStatusFailed
		return result, err
	}

	// Map results back to task IDs
	for _, task := range workflow.Tasks {
		result.Results[task.ID] = results[task.AgentName]
	}

	result.Status = TaskStatusCompleted
	result.CompletedAt = time.Now()
	result.Duration = result.CompletedAt.Sub(result.StartedAt)

	return result, nil
}

func (m *manager) executeAdaptiveWorkflow(ctx context.Context, workflow *Workflow, result *WorkflowResult) (*WorkflowResult, error) {
	// Trinity decides the best execution strategy
	// Springfield: "状況に応じて最適な実行戦略を選択します"
	
	// Simple heuristic: use parallel if more than 3 tasks
	if len(workflow.Tasks) > 3 {
		return m.executeParallelWorkflow(ctx, workflow, result)
	}
	return m.executeSequentialWorkflow(ctx, workflow, result)
}

func (m *manager) metricsCollector() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.config.MetricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			// Collect and aggregate metrics
			// This is where we would send metrics to monitoring systems
		}
	}
}

// defaultValidator provides basic validation
type defaultValidator struct{}

func (v *defaultValidator) Validate(def SubagentDefinition) error {
	// Basic validation rules
	if len(def.Name) > 100 {
		return &ValidationError{
			Field:   "name",
			Message: "agent name too long",
			Code:    "max_length",
		}
	}

	if def.MaxConcurrency < 0 {
		return &ValidationError{
			Field:   "max_concurrency",
			Message: "max concurrency must be positive",
			Code:    "invalid_value",
		}
	}

	// Vector: "……セキュリティレベルを検証……"
	if def.SecurityLevel < SecurityLevelPublic || def.SecurityLevel > SecurityLevelRestricted {
		return &ValidationError{
			Field:   "security_level",
			Message: "invalid security level",
			Code:    "invalid_value",
		}
	}

	return nil
}

// Krukai's High-Performance Extensions
// "完璧に最適化されたgoroutineプールとパフォーマンス監視システム"

// goroutinePool manages a pool of goroutines for efficient execution
type goroutinePool struct {
	jobs     chan func()
	workers  int
	quit     chan struct{}
	wg       sync.WaitGroup
}

func newGoroutinePool(workers int) *goroutinePool {
	pool := &goroutinePool{
		jobs:    make(chan func(), workers*2), // Buffer to prevent blocking
		workers: workers,
		quit:    make(chan struct{}),
	}
	
	// Start worker goroutines
	for i := 0; i < workers; i++ {
		pool.wg.Add(1)
		go pool.worker()
	}
	
	return pool
}

func (p *goroutinePool) worker() {
	defer p.wg.Done()
	
	for {
		select {
		case job := <-p.jobs:
			if job != nil {
				job()
			}
		case <-p.quit:
			return
		}
	}
}

func (p *goroutinePool) Submit(job func()) bool {
	select {
	case p.jobs <- job:
		return true
	default:
		return false // Pool is full
	}
}

func (p *goroutinePool) Close() {
	close(p.quit)
	p.wg.Wait()
	close(p.jobs)
}

// PerformanceTracker tracks detailed performance metrics
type PerformanceTracker struct {
	cpuUsage       atomic.Value // float64
	memUsage       atomic.Value // uint64 (bytes)
	goroutineCount atomic.Value // int
	gcCount        atomic.Value // uint32
	lastUpdate     atomic.Value // time.Time
}

func newPerformanceTracker() *PerformanceTracker {
	return &PerformanceTracker{}
}

func (pt *PerformanceTracker) UpdateMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	pt.memUsage.Store(m.Alloc)
	pt.goroutineCount.Store(runtime.NumGoroutine())
	pt.gcCount.Store(m.NumGC)
	pt.lastUpdate.Store(time.Now())
}

func (pt *PerformanceTracker) GetCPUUsage() float64 {
	if val := pt.cpuUsage.Load(); val != nil {
		return val.(float64)
	}
	return 0.0
}

func (pt *PerformanceTracker) GetMemoryUsage() uint64 {
	if val := pt.memUsage.Load(); val != nil {
		return val.(uint64)
	}
	return 0
}

func (pt *PerformanceTracker) GetGoroutineCount() int {
	if val := pt.goroutineCount.Load(); val != nil {
		return val.(int)
	}
	return 0
}

// healthChecker monitors system health
type healthChecker struct {
	maxConcurrency int
	alertThreshold float64
	alerts         chan string
	mu            sync.RWMutex
	healthy       bool
}

func newHealthChecker(maxConcurrency int) *healthChecker {
	return &healthChecker{
		maxConcurrency: maxConcurrency,
		alertThreshold: 0.8, // Alert at 80% capacity
		alerts:         make(chan string, 10),
		healthy:        true,
	}
}

func (hc *healthChecker) CheckHealth(activeExecs int) bool {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	
	utilizationRate := float64(activeExecs) / float64(hc.maxConcurrency)
	
	if utilizationRate > hc.alertThreshold {
		select {
		case hc.alerts <- fmt.Sprintf("High utilization: %.2f%%", utilizationRate*100):
		default: // Don't block if alert channel is full
		}
	}
	
	return hc.healthy
}

func (hc *healthChecker) GetAlerts() <-chan string {
	return hc.alerts
}

// Enhanced metrics collection with performance monitoring
func (m *manager) performanceMonitor() {
	defer m.wg.Done()
	
	ticker := time.NewTicker(5 * time.Second) // Frequent performance checks
	defer ticker.Stop()
	
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.performanceTracker.UpdateMetrics()
			
			// Check for resource leaks
			goroutines := m.performanceTracker.GetGoroutineCount()
			if goroutines > int(m.maxWorkers)*2 {
				// Log warning about potential goroutine leak
			}
		}
	}
}

func (m *manager) healthMonitor() {
	defer m.wg.Done()
	
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.metricsMu.RLock()
			activeExecs := m.metrics.ActiveExecutions
			m.metricsMu.RUnlock()
			
			m.healthCheck.CheckHealth(activeExecs)
			
		case alert := <-m.healthCheck.GetAlerts():
			// Handle health alerts (could integrate with monitoring system)
			_ = alert
		}
	}
}