// Package benchmark provides comprehensive performance benchmarking for subagent system
// Trinity統合: "三位一体によるパフォーマンス最適化の証明"
// 並列実行のベンチマーク、メモリ使用量測定、goroutine効率検証、レイテンシ測定
package benchmark

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/apto-as/omoikane/internal/subagent"
)

// BenchmarkSuite provides comprehensive performance testing
type BenchmarkSuite struct {
	manager        subagent.Manager
	config         subagent.ManagerConfig
	baselineMetrics *PerformanceBaseline
}

// PerformanceBaseline stores baseline performance metrics
type PerformanceBaseline struct {
	SingleExecutionLatency time.Duration
	ParallelThroughput     float64
	MemoryPerOperation     uint64
	GoroutinesPerOperation int
}

// NewBenchmarkSuite creates a new benchmark suite
func NewBenchmarkSuite() *BenchmarkSuite {
	config := subagent.DefaultManagerConfig()
	config.MaxConcurrentAgents = 50 // Krukai's optimization
	config.EnableMetrics = true
	
	return &BenchmarkSuite{
		manager: subagent.NewManager(config),
		config:  config,
	}
}

// Krukai: "並列実行の完璧な効率性を実証"
func BenchmarkParallelExecution(b *testing.B) {
	suite := NewBenchmarkSuite()
	
	// Register benchmark agents
	agentNames := make([]string, 10)
	for i := 0; i < 10; i++ {
		def := subagent.SubagentDefinition{
			Name:          fmt.Sprintf("parallel-bench-%d", i),
			Description:   "Parallel benchmark agent",
			Version:       "1.0.0",
			Tools:         []string{"benchmark"},
			Prompt:        "Parallel execution benchmark",
			TrinityRole:   subagent.TrinityRoleOptimizer,
			SecurityLevel: subagent.SecurityLevelPublic,
			Timeout:       5 * time.Second,
		}
		
		err := suite.manager.Register(def)
		if err != nil {
			b.Fatalf("Failed to register agent %d: %v", i, err)
		}
		agentNames[i] = def.Name
	}
	
	ctx := context.Background()
	
	b.ResetTimer()
	b.ReportAllocs()
	
	// Test parallel execution performance
	b.Run("ParallelExecution", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, err := suite.manager.ExecuteParallel(ctx, agentNames, "benchmark input")
				if err != nil {
					b.Errorf("Parallel execution failed: %v", err)
				}
			}
		})
	})
	
	// Measure throughput
	b.Run("Throughput", func(b *testing.B) {
		start := time.Now()
		for i := 0; i < b.N; i++ {
			_, err := suite.manager.ExecuteParallel(ctx, agentNames, fmt.Sprintf("throughput-test-%d", i))
			if err != nil {
				b.Errorf("Throughput test failed: %v", err)
			}
		}
		duration := time.Since(start)
		throughput := float64(b.N) / duration.Seconds()
		b.ReportMetric(throughput, "ops/sec")
	})
}

// Springfield: "システム全体の調和とスケーラビリティを測定"
func BenchmarkScalability(b *testing.B) {
	agentCounts := []int{1, 5, 10, 20, 50}
	
	for _, count := range agentCounts {
		b.Run(fmt.Sprintf("Agents_%d", count), func(b *testing.B) {
			suite := NewBenchmarkSuite()
			
			// Register agents
			agentNames := make([]string, count)
			for i := 0; i < count; i++ {
				def := subagent.SubagentDefinition{
					Name:          fmt.Sprintf("scale-test-%d", i),
					Description:   "Scalability test agent",
					Version:       "1.0.0",
					Tools:         []string{"scale"},
					Prompt:        "Scalability test",
					TrinityRole:   subagent.TrinityRoleStrategist,
					SecurityLevel: subagent.SecurityLevelPublic,
				}
				
				err := suite.manager.Register(def)
				if err != nil {
					b.Fatalf("Failed to register agent %d: %v", i, err)
				}
				agentNames[i] = def.Name
			}
			
			ctx := context.Background()
			
			b.ResetTimer()
			b.ReportAllocs()
			
			for i := 0; i < b.N; i++ {
				_, err := suite.manager.ExecuteParallel(ctx, agentNames, "scalability test")
				if err != nil {
					b.Errorf("Scalability test failed with %d agents: %v", count, err)
				}
			}
		})
	}
}

// Vector: "メモリリークと資源効率の厳密な検証"
func BenchmarkMemoryEfficiency(b *testing.B) {
	suite := NewBenchmarkSuite()
	
	// Register memory test agent
	def := subagent.SubagentDefinition{
		Name:          "memory-bench",
		Description:   "Memory efficiency benchmark",
		Version:       "1.0.0",
		Tools:         []string{"memory"},
		Prompt:        "Memory benchmark",
		TrinityRole:   subagent.TrinityRoleAuditor,
		SecurityLevel: subagent.SecurityLevelPublic,
	}
	
	err := suite.manager.Register(def)
	if err != nil {
		b.Fatalf("Failed to register memory benchmark agent: %v", err)
	}
	
	// Measure initial memory
	var m1 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)
	initialAlloc := m1.Alloc
	initialGoroutines := runtime.NumGoroutine()
	
	ctx := context.Background()
	
	b.ResetTimer()
	b.ReportAllocs()
	
	// Memory efficiency test
	b.Run("MemoryPerOperation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := suite.manager.Execute(ctx, "memory-bench", fmt.Sprintf("memory-test-%d", i))
			if err != nil {
				b.Errorf("Memory test failed: %v", err)
			}
		}
	})
	
	// Measure final memory
	runtime.GC()
	runtime.GC() // Double GC to ensure cleanup
	time.Sleep(100 * time.Millisecond)
	
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)
	finalAlloc := m2.Alloc
	finalGoroutines := runtime.NumGoroutine()
	
	memoryGrowth := finalAlloc - initialAlloc
	goroutineGrowth := finalGoroutines - initialGoroutines
	
	// Report memory metrics
	b.ReportMetric(float64(memoryGrowth), "bytes/total")
	b.ReportMetric(float64(memoryGrowth)/float64(b.N), "bytes/op")
	b.ReportMetric(float64(goroutineGrowth), "goroutines/total")
	
	// Verify no significant leaks
	maxAcceptableGrowth := uint64(1024 * 1024) // 1MB
	if memoryGrowth > maxAcceptableGrowth {
		b.Errorf("Potential memory leak: grew by %d bytes", memoryGrowth)
	}
	
	maxAcceptableGoroutines := 10
	if goroutineGrowth > maxAcceptableGoroutines {
		b.Errorf("Potential goroutine leak: grew by %d goroutines", goroutineGrowth)
	}
}

// Krukai: "レイテンシ最適化の精密測定"
func BenchmarkLatency(b *testing.B) {
	suite := NewBenchmarkSuite()
	
	// Register latency test agent
	def := subagent.SubagentDefinition{
		Name:          "latency-bench",
		Description:   "Latency benchmark agent",
		Version:       "1.0.0",
		Tools:         []string{"latency"},
		Prompt:        "Latency benchmark",
		TrinityRole:   subagent.TrinityRoleOptimizer,
		SecurityLevel: subagent.SecurityLevelPublic,
		Timeout:       1 * time.Second,
	}
	
	err := suite.manager.Register(def)
	if err != nil {
		b.Fatalf("Failed to register latency benchmark agent: %v", err)
	}
	
	ctx := context.Background()
	
	// Warmup
	for i := 0; i < 10; i++ {
		suite.manager.Execute(ctx, "latency-bench", "warmup")
	}
	
	// Collect latency measurements
	latencies := make([]time.Duration, b.N)
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		start := time.Now()
		_, err := suite.manager.Execute(ctx, "latency-bench", fmt.Sprintf("latency-test-%d", i))
		latency := time.Since(start)
		latencies[i] = latency
		
		if err != nil {
			b.Errorf("Latency test failed: %v", err)
		}
	}
	
	// Calculate percentiles
	if len(latencies) > 0 {
		// Sort latencies for percentile calculation
		for i := 0; i < len(latencies)-1; i++ {
			for j := i + 1; j < len(latencies); j++ {
				if latencies[i] > latencies[j] {
					latencies[i], latencies[j] = latencies[j], latencies[i]
				}
			}
		}
		
		p50 := latencies[len(latencies)*50/100]
		p95 := latencies[len(latencies)*95/100]
		p99 := latencies[len(latencies)*99/100]
		
		b.ReportMetric(float64(p50.Nanoseconds()), "p50_ns")
		b.ReportMetric(float64(p95.Nanoseconds()), "p95_ns")
		b.ReportMetric(float64(p99.Nanoseconds()), "p99_ns")
	}
}

// Trinity統合: "三位一体ワークフローの完全性能評価"
func BenchmarkTrinityWorkflow(b *testing.B) {
	suite := NewBenchmarkSuite()
	suite.config.TrinityMode = true
	
	// Register Trinity agents
	trinityAgents := []subagent.SubagentDefinition{
		{
			Name:          "trinity-springfield",
			Description:   "Strategic coordinator",
			Version:       "2.0.0",
			Tools:         []string{"strategy"},
			Prompt:        "Strategic analysis",
			TrinityRole:   subagent.TrinityRoleStrategist,
			SecurityLevel: subagent.SecurityLevelInternal,
		},
		{
			Name:          "trinity-krukai",
			Description:   "Performance optimizer",
			Version:       "2.0.0",
			Tools:         []string{"optimize"},
			Prompt:        "Performance optimization",
			TrinityRole:   subagent.TrinityRoleOptimizer,
			SecurityLevel: subagent.SecurityLevelInternal,
		},
		{
			Name:          "trinity-vector",
			Description:   "Security auditor",
			Version:       "2.0.0",
			Tools:         []string{"audit"},
			Prompt:        "Security audit",
			TrinityRole:   subagent.TrinityRoleAuditor,
			SecurityLevel: subagent.SecurityLevelConfidential,
		},
	}
	
	for _, agent := range trinityAgents {
		err := suite.manager.Register(agent)
		if err != nil {
			b.Fatalf("Failed to register Trinity agent %s: %v", agent.Name, err)
		}
	}
	
	ctx := context.Background()
	
	// Test different workflow strategies
	strategies := []struct {
		name     string
		strategy subagent.ExecutionStrategy
	}{
		{"Sequential", subagent.StrategySequential},
		{"Parallel", subagent.StrategyParallel},
		{"Adaptive", subagent.StrategyAdaptive},
	}
	
	for _, strat := range strategies {
		b.Run(fmt.Sprintf("Strategy_%s", strat.name), func(b *testing.B) {
			workflow := &subagent.Workflow{
				ID:          "trinity-benchmark",
				Name:        "Trinity Benchmark Workflow",
				Description: "Performance benchmark for Trinity workflow",
				Tasks: []subagent.Task{
					{ID: "analyze", AgentName: "trinity-springfield", Input: "strategic analysis"},
					{ID: "optimize", AgentName: "trinity-krukai", Input: "performance optimization"},
					{ID: "audit", AgentName: "trinity-vector", Input: "security audit"},
				},
				Strategy: strat.strategy,
			}
			
			b.ResetTimer()
			b.ReportAllocs()
			
			for i := 0; i < b.N; i++ {
				result, err := suite.manager.ExecuteWorkflow(ctx, workflow)
				if err != nil {
					b.Errorf("Trinity workflow failed: %v", err)
				}
				if result.Status != subagent.TaskStatusCompleted {
					b.Errorf("Workflow not completed: %s", result.Status)
				}
			}
		})
	}
}

// Springfield: "高負荷環境での安定性テスト"
func BenchmarkHighLoadStability(b *testing.B) {
	suite := NewBenchmarkSuite()
	
	// Register high-load test agent
	def := subagent.SubagentDefinition{
		Name:          "highload-bench",
		Description:   "High load stability benchmark",
		Version:       "1.0.0",
		Tools:         []string{"highload"},
		Prompt:        "High load test",
		TrinityRole:   subagent.TrinityRoleStrategist,
		SecurityLevel: subagent.SecurityLevelPublic,
		MaxConcurrency: 5,
		Timeout:       10 * time.Second,
	}
	
	err := suite.manager.Register(def)
	if err != nil {
		b.Fatalf("Failed to register high load agent: %v", err)
	}
	
	ctx := context.Background()
	concurrency := 100
	
	b.Run("HighConcurrency", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var wg sync.WaitGroup
			errors := make(chan error, concurrency)
			
			// Launch concurrent operations
			for j := 0; j < concurrency; j++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					_, err := suite.manager.Execute(ctx, "highload-bench", fmt.Sprintf("highload-%d", id))
					if err != nil {
						select {
						case errors <- err:
						default:
						}
					}
				}(j)
			}
			
			wg.Wait()
			close(errors)
			
			// Check for errors
			errorCount := 0
			for err := range errors {
				errorCount++
				if errorCount == 1 { // Only report first error
					b.Errorf("High load test error: %v", err)
				}
			}
			
			// Report error rate
			errorRate := float64(errorCount) / float64(concurrency) * 100
			b.ReportMetric(errorRate, "error_rate_%")
		}
	})
}

// Vector: "リソース制限とセキュリティ検証のベンチマーク"
func BenchmarkResourceLimits(b *testing.B) {
	config := subagent.DefaultManagerConfig()
	config.MaxConcurrentAgents = 5 // Tight limit
	suite := &BenchmarkSuite{
		manager: subagent.NewManager(config),
		config:  config,
	}
	
	// Register resource-limited agent
	def := subagent.SubagentDefinition{
		Name:          "resource-bench",
		Description:   "Resource limits benchmark",
		Version:       "1.0.0",
		Tools:         []string{"resource"},
		Prompt:        "Resource test",
		TrinityRole:   subagent.TrinityRoleAuditor,
		SecurityLevel: subagent.SecurityLevelPublic,
		MaxConcurrency: 1,
		Timeout:       2 * time.Second,
	}
	
	err := suite.manager.Register(def)
	if err != nil {
		b.Fatalf("Failed to register resource benchmark agent: %v", err)
	}
	
	ctx := context.Background()
	
	b.Run("ResourceConstraints", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				// This should respect resource limits
				_, err := suite.manager.Execute(ctx, "resource-bench", "resource test")
				if err != nil {
					// Some executions may fail due to resource limits - this is expected
					b.Logf("Resource-limited execution: %v", err)
				}
			}
		})
	})
}

// Performance regression test
func BenchmarkRegressionTest(b *testing.B) {
	suite := NewBenchmarkSuite()
	
	// Register regression test agent
	def := subagent.SubagentDefinition{
		Name:          "regression-bench",
		Description:   "Performance regression test",
		Version:       "1.0.0",
		Tools:         []string{"regression"},
		Prompt:        "Regression test",
		TrinityRole:   subagent.TrinityRoleOptimizer,
		SecurityLevel: subagent.SecurityLevelPublic,
	}
	
	err := suite.manager.Register(def)
	if err != nil {
		b.Fatalf("Failed to register regression test agent: %v", err)
	}
	
	ctx := context.Background()
	
	// Expected performance thresholds
	maxLatency := 100 * time.Millisecond
	minThroughput := 10.0 // ops/sec
	
	b.Run("PerformanceThresholds", func(b *testing.B) {
		start := time.Now()
		
		for i := 0; i < b.N; i++ {
			opStart := time.Now()
			_, err := suite.manager.Execute(ctx, "regression-bench", fmt.Sprintf("regression-%d", i))
			opLatency := time.Since(opStart)
			
			if err != nil {
				b.Errorf("Regression test failed: %v", err)
			}
			
			if opLatency > maxLatency {
				b.Errorf("Operation too slow: %v > %v", opLatency, maxLatency)
			}
		}
		
		totalDuration := time.Since(start)
		actualThroughput := float64(b.N) / totalDuration.Seconds()
		
		if actualThroughput < minThroughput {
			b.Errorf("Throughput too low: %.2f < %.2f ops/sec", actualThroughput, minThroughput)
		}
		
		b.ReportMetric(actualThroughput, "ops/sec")
	})
}

// Comprehensive benchmark suite runner
func BenchmarkComprehensiveSuite(b *testing.B) {
	// Trinity: "全体的なシステムパフォーマンスの統合評価"
	b.Run("ParallelExecution", BenchmarkParallelExecution)
	b.Run("Scalability", BenchmarkScalability)
	b.Run("MemoryEfficiency", BenchmarkMemoryEfficiency)
	b.Run("Latency", BenchmarkLatency)
	b.Run("TrinityWorkflow", BenchmarkTrinityWorkflow)
	b.Run("HighLoadStability", BenchmarkHighLoadStability)
	b.Run("ResourceLimits", BenchmarkResourceLimits)
	b.Run("RegressionTest", BenchmarkRegressionTest)
}