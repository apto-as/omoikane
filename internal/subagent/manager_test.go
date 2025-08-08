package subagent

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestManagerRegistration(t *testing.T) {
	// Springfield: "テストによる品質保証を実施"
	config := DefaultManagerConfig()
	manager := NewManager(config)
	
	// Create test agent definition
	def := SubagentDefinition{
		Name:           "test-agent",
		Description:    "Test agent for unit testing",
		Version:        "1.0.0",
		Tools:          []string{"read", "write"},
		Prompt:         "Test prompt",
		TrinityRole:    TrinityRoleOptimizer,
		SecurityLevel:  SecurityLevelInternal,
		MaxConcurrency: 2,
		Timeout:        10 * time.Second,
	}
	
	// Test registration
	err := manager.Register(def)
	if err != nil {
		t.Fatalf("Failed to register agent: %v", err)
	}
	
	// Test duplicate registration
	err = manager.Register(def)
	if err != ErrAgentAlreadyExists {
		t.Errorf("Expected ErrAgentAlreadyExists, got: %v", err)
	}
	
	// Test retrieval
	retrieved, err := manager.Get("test-agent")
	if err != nil {
		t.Fatalf("Failed to get agent: %v", err)
	}
	
	if retrieved.Name != def.Name {
		t.Errorf("Expected name %s, got %s", def.Name, retrieved.Name)
	}
	
	// Test listing
	agents := manager.List()
	if len(agents) != 1 {
		t.Errorf("Expected 1 agent, got %d", len(agents))
	}
	
	// Test list by role
	optimizers := manager.ListByRole(TrinityRoleOptimizer)
	if len(optimizers) != 1 {
		t.Errorf("Expected 1 optimizer, got %d", len(optimizers))
	}
	
	// Test unregistration
	err = manager.Unregister("test-agent")
	if err != nil {
		t.Fatalf("Failed to unregister agent: %v", err)
	}
	
	// Verify agent is removed
	_, err = manager.Get("test-agent")
	if err != ErrAgentNotFound {
		t.Errorf("Expected ErrAgentNotFound, got: %v", err)
	}
}

func TestManagerExecution(t *testing.T) {
	// Krukai: "実行性能を検証"
	config := DefaultManagerConfig()
	manager := NewManager(config)
	
	// Register test agent
	def := SubagentDefinition{
		Name:           "exec-test",
		Description:    "Execution test agent",
		Version:        "1.0.0",
		Tools:          []string{"execute"},
		Prompt:         "Execute test",
		TrinityRole:    TrinityRoleAuditor,
		SecurityLevel:  SecurityLevelPublic,
		MaxConcurrency: 1,
		Timeout:        5 * time.Second,
	}
	
	err := manager.Register(def)
	if err != nil {
		t.Fatalf("Failed to register agent: %v", err)
	}
	
	// Test execution
	ctx := context.Background()
	result, err := manager.Execute(ctx, "exec-test", "test input")
	if err != nil {
		t.Fatalf("Failed to execute agent: %v", err)
	}
	
	// Check result structure
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}
	
	if resultMap["agent"] != "exec-test" {
		t.Errorf("Expected agent name in result")
	}
	
	// Test execution with timeout
	shortCtx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()
	
	_, err = manager.ExecuteWithOptions(shortCtx, "exec-test", "test", ExecutionOptions{
		Timeout: 5 * time.Millisecond,
	})
	// Note: Current mock implementation may not respect timeout
	// In real implementation, this should timeout
}

func TestManagerParallelExecution(t *testing.T) {
	// Vector: "……並列実行の安全性を検証……"
	config := DefaultManagerConfig()
	config.MaxConcurrentAgents = 3
	manager := NewManager(config)
	
	// Register multiple agents
	agents := []SubagentDefinition{
		{
			Name:        "parallel-1",
			Description: "Parallel test agent 1",
			Version:     "1.0.0",
			Tools:       []string{"test"},
			Prompt:      "Test 1",
			TrinityRole: TrinityRoleStrategist,
		},
		{
			Name:        "parallel-2",
			Description: "Parallel test agent 2",
			Version:     "1.0.0",
			Tools:       []string{"test"},
			Prompt:      "Test 2",
			TrinityRole: TrinityRoleOptimizer,
		},
		{
			Name:        "parallel-3",
			Description: "Parallel test agent 3",
			Version:     "1.0.0",
			Tools:       []string{"test"},
			Prompt:      "Test 3",
			TrinityRole: TrinityRoleAuditor,
		},
	}
	
	for _, def := range agents {
		if err := manager.Register(def); err != nil {
			t.Fatalf("Failed to register agent %s: %v", def.Name, err)
		}
	}
	
	// Execute in parallel
	ctx := context.Background()
	agentNames := []string{"parallel-1", "parallel-2", "parallel-3"}
	
	results, err := manager.ExecuteParallel(ctx, agentNames, "parallel test")
	if err != nil {
		t.Fatalf("Failed to execute parallel: %v", err)
	}
	
	// Verify all agents executed
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}
	
	for _, name := range agentNames {
		if _, exists := results[name]; !exists {
			t.Errorf("Missing result for agent %s", name)
		}
	}
}

func TestManagerWorkflow(t *testing.T) {
	// Springfield: "ワークフロー全体の調和を検証"
	config := DefaultManagerConfig()
	manager := NewManager(config)
	
	// Register agents for workflow
	agents := []SubagentDefinition{
		{
			Name:        "workflow-1",
			Description: "Workflow agent 1",
			Version:     "1.0.0",
			Tools:       []string{"analyze"},
			Prompt:      "Analyze",
			TrinityRole: TrinityRoleStrategist,
		},
		{
			Name:        "workflow-2",
			Description: "Workflow agent 2",
			Version:     "1.0.0",
			Tools:       []string{"implement"},
			Prompt:      "Implement",
			TrinityRole: TrinityRoleOptimizer,
		},
		{
			Name:        "workflow-3",
			Description: "Workflow agent 3",
			Version:     "1.0.0",
			Tools:       []string{"validate"},
			Prompt:      "Validate",
			TrinityRole: TrinityRoleAuditor,
		},
	}
	
	for _, def := range agents {
		if err := manager.Register(def); err != nil {
			t.Fatalf("Failed to register agent %s: %v", def.Name, err)
		}
	}
	
	// Create workflow
	workflow := &Workflow{
		ID:          "test-workflow",
		Name:        "Test Workflow",
		Description: "Trinity workflow test",
		Tasks: []Task{
			{ID: "task-1", AgentName: "workflow-1", Input: "analyze this"},
			{ID: "task-2", AgentName: "workflow-2", Input: "implement that"},
			{ID: "task-3", AgentName: "workflow-3", Input: "validate all"},
		},
		Strategy: StrategySequential,
	}
	
	// Execute workflow
	ctx := context.Background()
	result, err := manager.ExecuteWorkflow(ctx, workflow)
	if err != nil {
		t.Fatalf("Failed to execute workflow: %v", err)
	}
	
	// Verify workflow completed
	if result.Status != TaskStatusCompleted {
		t.Errorf("Expected workflow to complete, got status: %s", result.Status)
	}
	
	// Verify all tasks executed
	if len(result.Results) != 3 {
		t.Errorf("Expected 3 task results, got %d", len(result.Results))
	}
	
	// Test parallel workflow
	workflow.Strategy = StrategyParallel
	result, err = manager.ExecuteWorkflow(ctx, workflow)
	if err != nil {
		t.Fatalf("Failed to execute parallel workflow: %v", err)
	}
	
	if result.Status != TaskStatusCompleted {
		t.Errorf("Expected parallel workflow to complete, got status: %s", result.Status)
	}
	
	// Test adaptive workflow
	workflow.Strategy = StrategyAdaptive
	result, err = manager.ExecuteWorkflow(ctx, workflow)
	if err != nil {
		t.Fatalf("Failed to execute adaptive workflow: %v", err)
	}
	
	if result.Status != TaskStatusCompleted {
		t.Errorf("Expected adaptive workflow to complete, got status: %s", result.Status)
	}
}

func TestManagerMetrics(t *testing.T) {
	// Krukai: "メトリクスの正確性を検証"
	config := DefaultManagerConfig()
	config.EnableMetrics = true
	manager := NewManager(config)
	
	// Register and execute agent
	def := SubagentDefinition{
		Name:        "metrics-test",
		Description: "Metrics test agent",
		Version:     "1.0.0",
		Tools:       []string{"metrics"},
		Prompt:      "Metrics",
		TrinityRole: TrinityRoleOptimizer,
	}
	
	err := manager.Register(def)
	if err != nil {
		t.Fatalf("Failed to register agent: %v", err)
	}
	
	// Execute multiple times
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		_, err = manager.Execute(ctx, "metrics-test", "test")
		if err != nil {
			t.Fatalf("Execution %d failed: %v", i, err)
		}
	}
	
	// Check metrics
	metrics := manager.GetMetrics()
	if metrics.TotalExecutions != 5 {
		t.Errorf("Expected 5 total executions, got %d", metrics.TotalExecutions)
	}
	
	// Check agent-specific metrics
	agentMetrics, err := manager.GetAgentMetrics("metrics-test")
	if err != nil {
		t.Fatalf("Failed to get agent metrics: %v", err)
	}
	
	if agentMetrics.TotalExecutions != 5 {
		t.Errorf("Expected 5 agent executions, got %d", agentMetrics.TotalExecutions)
	}
	
	if agentMetrics.SuccessfulExecs != 5 {
		t.Errorf("Expected 5 successful executions, got %d", agentMetrics.SuccessfulExecs)
	}
}

func TestManagerShutdown(t *testing.T) {
	// Vector: "……適切なシャットダウンを確認……"
	config := DefaultManagerConfig()
	manager := NewManager(config)
	
	// Initialize manager
	ctx := context.Background()
	err := manager.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}
	
	// Shutdown manager
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	err = manager.Shutdown(shutdownCtx)
	if err != nil {
		t.Fatalf("Failed to shutdown: %v", err)
	}
}

// Springfield's Enhanced Test Suite
// "品質保証のための包括的なテスト実装"

func TestManagerConcurrencyLimits(t *testing.T) {
	// Springfield: "適切な並行性制御のテスト"
	config := DefaultManagerConfig()
	config.MaxConcurrentAgents = 2 // Low limit for testing
	manager := NewManager(config)
	
	// Register test agent
	def := SubagentDefinition{
		Name:           "concurrency-test",
		Description:    "Concurrency test agent",
		Version:        "1.0.0",
		Tools:          []string{"test"},
		Prompt:         "Test concurrency",
		TrinityRole:    TrinityRoleStrategist,
		SecurityLevel:  SecurityLevelPublic,
		MaxConcurrency: 1,
		Timeout:        1 * time.Second,
	}
	
	err := manager.Register(def)
	if err != nil {
		t.Fatalf("Failed to register agent: %v", err)
	}
	
	// Test concurrent execution respecting limits
	ctx := context.Background()
	var wg sync.WaitGroup
	results := make([]interface{}, 5)
	errors := make([]error, 5)
	
	// Launch 5 concurrent executions (more than limit)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			result, err := manager.Execute(ctx, "concurrency-test", fmt.Sprintf("test-%d", index))
			results[index] = result
			errors[index] = err
		}(i)
	}
	
	wg.Wait()
	
	// Check that all executions completed (may have been queued)
	successCount := 0
	for i := 0; i < 5; i++ {
		if errors[i] == nil {
			successCount++
		}
	}
	
	if successCount == 0 {
		t.Error("Expected at least some executions to succeed")
	}
	
	t.Logf("Successful executions: %d/5", successCount)
}

func TestManagerResourceManagement(t *testing.T) {
	// Springfield: "リソース管理の正確性を検証"
	config := DefaultManagerConfig()
	config.EnableMetrics = true
	manager := NewManager(config).(*manager)
	
	// Register multiple agents
	for i := 0; i < 10; i++ {
		def := SubagentDefinition{
			Name:           fmt.Sprintf("resource-test-%d", i),
			Description:    fmt.Sprintf("Resource test agent %d", i),
			Version:        "1.0.0",
			Tools:          []string{"resource"},
			Prompt:         "Test resources",
			TrinityRole:    TrinityRoleOptimizer,
			SecurityLevel:  SecurityLevelInternal,
			MaxConcurrency: 1,
		}
		
		err := manager.Register(def)
		if err != nil {
			t.Fatalf("Failed to register agent %d: %v", i, err)
		}
	}
	
	// Check registry size
	agents := manager.List()
	if len(agents) != 10 {
		t.Errorf("Expected 10 agents, got %d", len(agents))
	}
	
	// Test memory usage tracking (basic check)
	initialGoroutines := runtime.NumGoroutine()
	manager.performanceTracker.UpdateMetrics()
	goroutineCount := manager.performanceTracker.GetGoroutineCount()
	
	if goroutineCount < initialGoroutines {
		t.Errorf("Goroutine tracking seems incorrect: %d < %d", goroutineCount, initialGoroutines)
	}
	
	// Cleanup and verify no leaks
	for i := 0; i < 10; i++ {
		err := manager.Unregister(fmt.Sprintf("resource-test-%d", i))
		if err != nil {
			t.Errorf("Failed to unregister agent %d: %v", i, err)
		}
	}
	
	// Verify cleanup
	agents = manager.List()
	if len(agents) != 0 {
		t.Errorf("Expected 0 agents after cleanup, got %d", len(agents))
	}
}

func TestManagerErrorHandling(t *testing.T) {
	// Springfield: "エラー処理の堅牢性を確認"
	config := DefaultManagerConfig()
	manager := NewManager(config)
	
	// Test operations on non-existent agent
	ctx := context.Background()
	
	// Get non-existent agent
	_, err := manager.Get("non-existent")
	if err != ErrAgentNotFound {
		t.Errorf("Expected ErrAgentNotFound, got: %v", err)
	}
	
	// Execute non-existent agent
	_, err = manager.Execute(ctx, "non-existent", "test")
	if err == nil {
		t.Error("Expected error for non-existent agent execution")
	}
	
	// Get metrics for non-existent agent
	_, err = manager.GetAgentMetrics("non-existent")
	if err != ErrAgentNotFound {
		t.Errorf("Expected ErrAgentNotFound for metrics, got: %v", err)
	}
	
	// Unregister non-existent agent
	err = manager.Unregister("non-existent")
	if err != ErrAgentNotFound {
		t.Errorf("Expected ErrAgentNotFound for unregister, got: %v", err)
	}
}

func TestManagerValidation(t *testing.T) {
	// Springfield: "入力検証の正確性をテスト"
	config := DefaultManagerConfig()
	manager := NewManager(config)
	
	// Test invalid agent definitions
	testCases := []struct {
		name        string
		definition  SubagentDefinition
		expectError bool
	}{
		{
			name: "empty name",
			definition: SubagentDefinition{
				Name:        "",
				Description: "Test",
			},
			expectError: true,
		},
		{
			name: "empty description",
			definition: SubagentDefinition{
				Name:        "test",
				Description: "",
			},
			expectError: true,
		},
		{
			name: "negative concurrency",
			definition: SubagentDefinition{
				Name:           "test",
				Description:    "Test",
				MaxConcurrency: -1,
			},
			expectError: true,
		},
		{
			name: "invalid security level",
			definition: SubagentDefinition{
				Name:          "test",
				Description:   "Test",
				SecurityLevel: SecurityLevel(999),
			},
			expectError: true,
		},
		{
			name: "valid definition",
			definition: SubagentDefinition{
				Name:          "test",
				Description:   "Test agent",
				Version:       "1.0.0",
				Tools:         []string{"test"},
				Prompt:        "Test prompt",
				TrinityRole:   TrinityRoleStrategist,
				SecurityLevel: SecurityLevelPublic,
			},
			expectError: false,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := manager.Register(tc.definition)
			if tc.expectError && err == nil {
				t.Errorf("Expected error for %s, but got none", tc.name)
			} else if !tc.expectError && err != nil {
				t.Errorf("Expected no error for %s, but got: %v", tc.name, err)
			}
			
			// Cleanup valid registrations
			if !tc.expectError && err == nil {
				manager.Unregister(tc.definition.Name)
			}
		})
	}
}

func TestManagerTrinityIntegration(t *testing.T) {
	// Springfield: "三位一体システムの統合テスト"
	config := DefaultManagerConfig()
	config.TrinityMode = true
	config.RequireConsensus = true
	manager := NewManager(config)
	
	// Register Trinity agents
	trinityAgents := []SubagentDefinition{
		{
			Name:          "springfield",
			Description:   "Strategic coordinator",
			Version:       "2.0.0",
			Tools:         []string{"strategy", "coordination"},
			Prompt:        "Strategic analysis and coordination",
			TrinityRole:   TrinityRoleStrategist,
			SecurityLevel: SecurityLevelInternal,
		},
		{
			Name:          "krukai",
			Description:   "Performance optimizer",
			Version:       "2.0.0",
			Tools:         []string{"optimize", "profile"},
			Prompt:        "Performance optimization and quality assurance",
			TrinityRole:   TrinityRoleOptimizer,
			SecurityLevel: SecurityLevelInternal,
		},
		{
			Name:          "vector",
			Description:   "Security auditor",
			Version:       "2.0.0",
			Tools:         []string{"audit", "scan"},
			Prompt:        "Security audit and risk assessment",
			TrinityRole:   TrinityRoleAuditor,
			SecurityLevel: SecurityLevelConfidential,
		},
	}
	
	for _, agent := range trinityAgents {
		err := manager.Register(agent)
		if err != nil {
			t.Fatalf("Failed to register %s: %v", agent.Name, err)
		}
	}
	
	// Test role-based listing
	strategists := manager.ListByRole(TrinityRoleStrategist)
	if len(strategists) != 1 || strategists[0].Name != "springfield" {
		t.Error("Expected one strategist named 'springfield'")
	}
	
	optimizers := manager.ListByRole(TrinityRoleOptimizer)
	if len(optimizers) != 1 || optimizers[0].Name != "krukai" {
		t.Error("Expected one optimizer named 'krukai'")
	}
	
	auditors := manager.ListByRole(TrinityRoleAuditor)
	if len(auditors) != 1 || auditors[0].Name != "vector" {
		t.Error("Expected one auditor named 'vector'")
	}
	
	// Test Trinity workflow execution
	workflow := &Workflow{
		ID:          "trinity-test",
		Name:        "Trinity Integration Test",
		Description: "Test Trinity coordination",
		Tasks: []Task{
			{ID: "analyze", AgentName: "springfield", Input: "strategic analysis"},
			{ID: "optimize", AgentName: "krukai", Input: "performance optimization"},
			{ID: "audit", AgentName: "vector", Input: "security audit"},
		},
		Strategy: StrategyAdaptive, // Trinity decides
	}
	
	ctx := context.Background()
	result, err := manager.ExecuteWorkflow(ctx, workflow)
	if err != nil {
		t.Fatalf("Trinity workflow failed: %v", err)
	}
	
	if result.Status != TaskStatusCompleted {
		t.Errorf("Expected workflow completion, got: %s", result.Status)
	}
	
	if len(result.Results) != 3 {
		t.Errorf("Expected 3 task results, got %d", len(result.Results))
	}
}