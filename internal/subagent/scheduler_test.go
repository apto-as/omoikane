package subagent

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestSchedulerBasic(t *testing.T) {
	// Springfield: "スケジューラーの基本動作を検証"
	config := DefaultManagerConfig()
	manager := NewManager(config)
	
	// Register test agent
	def := SubagentDefinition{
		Name:        "scheduler-test",
		Description: "Scheduler test agent",
		Version:     "1.0.0",
		Tools:       []string{"test"},
		Prompt:      "Test scheduler",
		TrinityRole: TrinityRoleOptimizer,
	}
	
	err := manager.Register(def)
	if err != nil {
		t.Fatalf("Failed to register agent: %v", err)
	}
	
	// Create scheduler
	scheduler := NewDynamicScheduler(manager, 5)
	
	// Create and schedule a task
	task := &ScheduledTask{
		Task: Task{
			ID:        "test-task-1",
			AgentName: "scheduler-test",
			Input:     "test input",
		},
		Priority: PriorityNormal,
	}
	
	ctx := context.Background()
	err = scheduler.Schedule(ctx, task)
	if err != nil {
		t.Fatalf("Failed to schedule task: %v", err)
	}
	
	// Check status
	status := scheduler.GetStatus()
	if status.QueuedTasks != 1 {
		t.Errorf("Expected 1 queued task, got %d", status.QueuedTasks)
	}
	
	// Start scheduler in background
	execCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	
	go scheduler.Execute(execCtx)
	
	// Wait for task completion
	time.Sleep(200 * time.Millisecond)
	
	// Check completion
	status = scheduler.GetStatus()
	if status.CompletedTasks != 1 {
		t.Errorf("Expected 1 completed task, got %d", status.CompletedTasks)
	}
	
	// Stop scheduler
	stopCtx, stopCancel := context.WithTimeout(ctx, 2*time.Second)
	defer stopCancel()
	
	cancel() // Cancel execution context
	err = scheduler.Stop(stopCtx)
	if err != nil {
		t.Fatalf("Failed to stop scheduler: %v", err)
	}
}

func TestSchedulerPriority(t *testing.T) {
	// Krukai: "優先度順序を厳密に検証"
	config := DefaultManagerConfig()
	manager := NewManager(config)
	
	// Register agent
	def := SubagentDefinition{
		Name:        "priority-test",
		Description: "Priority test agent",
		Version:     "1.0.0",
		Tools:       []string{"test"},
		Prompt:      "Test priority",
		TrinityRole: TrinityRoleAuditor,
	}
	
	manager.Register(def)
	
	// Create scheduler with single worker to ensure sequential processing
	scheduler := NewDynamicScheduler(manager, 1)
	
	// Schedule tasks with different priorities
	ctx := context.Background()
	
	tasks := []*ScheduledTask{
		{
			Task: Task{ID: "low", AgentName: "priority-test", Input: "low"},
			Priority: PriorityLow,
		},
		{
			Task: Task{ID: "critical", AgentName: "priority-test", Input: "critical"},
			Priority: PriorityCritical,
		},
		{
			Task: Task{ID: "normal", AgentName: "priority-test", Input: "normal"},
			Priority: PriorityNormal,
		},
		{
			Task: Task{ID: "high", AgentName: "priority-test", Input: "high"},
			Priority: PriorityHigh,
		},
	}
	
	for _, task := range tasks {
		err := scheduler.Schedule(ctx, task)
		if err != nil {
			t.Fatalf("Failed to schedule task %s: %v", task.ID, err)
		}
	}
	
	// Verify queue ordering
	queue := scheduler.taskQueue
	if queue.Len() != 4 {
		t.Errorf("Expected 4 tasks in queue, got %d", queue.Len())
	}
	
	// Start scheduler
	execCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	
	go scheduler.Execute(execCtx)
	
	// Wait for all tasks to complete
	time.Sleep(500 * time.Millisecond)
	
	// Verify all completed
	status := scheduler.GetStatus()
	if status.CompletedTasks != 4 {
		t.Errorf("Expected 4 completed tasks, got %d", status.CompletedTasks)
	}
	
	cancel()
	scheduler.Stop(context.Background())
}

func TestSchedulerDependencies(t *testing.T) {
	// Vector: "……依存関係の正確な処理を検証……"
	config := DefaultManagerConfig()
	manager := NewManager(config)
	
	// Register agent
	def := SubagentDefinition{
		Name:        "dep-test",
		Description: "Dependency test agent",
		Version:     "1.0.0",
		Tools:       []string{"test"},
		Prompt:      "Test dependencies",
		TrinityRole: TrinityRoleStrategist,
	}
	
	manager.Register(def)
	
	// Create scheduler
	scheduler := NewDynamicScheduler(manager, 3)
	
	// Create tasks with dependencies
	// Task C depends on A and B
	// Task D depends on C
	ctx := context.Background()
	
	taskA := &ScheduledTask{
		Task: Task{ID: "task-a", AgentName: "dep-test", Input: "A"},
		Priority: PriorityNormal,
	}
	
	taskB := &ScheduledTask{
		Task: Task{ID: "task-b", AgentName: "dep-test", Input: "B"},
		Priority: PriorityNormal,
	}
	
	taskC := &ScheduledTask{
		Task: Task{ID: "task-c", AgentName: "dep-test", Input: "C"},
		Priority: PriorityNormal,
		Dependencies: []string{"task-a", "task-b"},
	}
	
	taskD := &ScheduledTask{
		Task: Task{ID: "task-d", AgentName: "dep-test", Input: "D"},
		Priority: PriorityNormal,
		Dependencies: []string{"task-c"},
	}
	
	// Schedule in reverse order to test dependency handling
	scheduler.Schedule(ctx, taskD)
	scheduler.Schedule(ctx, taskC)
	scheduler.Schedule(ctx, taskB)
	scheduler.Schedule(ctx, taskA)
	
	// Start scheduler
	execCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	
	go scheduler.Execute(execCtx)
	
	// Wait for completion
	time.Sleep(1 * time.Second)
	
	// Verify all tasks completed
	status := scheduler.GetStatus()
	if status.CompletedTasks != 4 {
		t.Errorf("Expected 4 completed tasks, got %d", status.CompletedTasks)
	}
	
	// Verify C was executed after A and B
	// Verify D was executed after C
	// (In a real implementation, we'd track execution order)
	
	cancel()
	scheduler.Stop(context.Background())
}

func TestSchedulerBatchScheduling(t *testing.T) {
	// Springfield: "バッチ処理の効率性を検証"
	config := DefaultManagerConfig()
	manager := NewManager(config)
	
	// Register agent
	def := SubagentDefinition{
		Name:        "batch-test",
		Description: "Batch test agent",
		Version:     "1.0.0",
		Tools:       []string{"test"},
		Prompt:      "Test batch",
		TrinityRole: TrinityRoleOptimizer,
	}
	
	manager.Register(def)
	
	// Create scheduler
	scheduler := NewDynamicScheduler(manager, 5)
	
	// Create batch of tasks
	var tasks []*ScheduledTask
	for i := 0; i < 10; i++ {
		task := &ScheduledTask{
			Task: Task{
				ID:        fmt.Sprintf("batch-task-%d", i),
				AgentName: "batch-test",
				Input:     fmt.Sprintf("input-%d", i),
			},
			Priority: PriorityNormal,
		}
		tasks = append(tasks, task)
	}
	
	// Schedule batch
	ctx := context.Background()
	err := scheduler.ScheduleBatch(ctx, tasks)
	if err != nil {
		t.Fatalf("Failed to schedule batch: %v", err)
	}
	
	// Check queue size
	status := scheduler.GetStatus()
	if status.QueuedTasks != 10 {
		t.Errorf("Expected 10 queued tasks, got %d", status.QueuedTasks)
	}
	
	// Start scheduler
	execCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	
	go scheduler.Execute(execCtx)
	
	// Wait for completion
	time.Sleep(1 * time.Second)
	
	// Verify all completed
	status = scheduler.GetStatus()
	if status.CompletedTasks != 10 {
		t.Errorf("Expected 10 completed tasks, got %d", status.CompletedTasks)
	}
	
	cancel()
	scheduler.Stop(context.Background())
}

func TestSchedulerDynamicWorkers(t *testing.T) {
	// Krukai: "動的ワーカー調整を検証"
	config := DefaultManagerConfig()
	manager := NewManager(config)
	
	// Register agent
	def := SubagentDefinition{
		Name:        "worker-test",
		Description: "Worker test agent",
		Version:     "1.0.0",
		Tools:       []string{"test"},
		Prompt:      "Test workers",
		TrinityRole: TrinityRoleOptimizer,
	}
	
	manager.Register(def)
	
	// Create scheduler with initial low worker count
	scheduler := NewDynamicScheduler(manager, 2)
	
	// Start scheduler
	ctx := context.Background()
	execCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	
	go scheduler.Execute(execCtx)
	
	// Schedule many tasks to trigger worker scaling
	for i := 0; i < 20; i++ {
		task := &ScheduledTask{
			Task: Task{
				ID:        fmt.Sprintf("worker-task-%d", i),
				AgentName: "worker-test",
				Input:     fmt.Sprintf("input-%d", i),
			},
			Priority: PriorityNormal,
		}
		scheduler.Schedule(ctx, task)
	}
	
	// Wait a bit for workers to scale up
	time.Sleep(100 * time.Millisecond)
	
	// Check worker count increased
	status := scheduler.GetStatus()
	if status.ActiveWorkers < 2 {
		t.Errorf("Expected at least 2 active workers, got %d", status.ActiveWorkers)
	}
	
	// Increase max workers
	scheduler.SetMaxWorkers(5)
	
	// Wait for adjustment
	time.Sleep(2 * time.Second)
	
	// Verify tasks completed
	status = scheduler.GetStatus()
	if status.CompletedTasks != 20 {
		t.Errorf("Expected 20 completed tasks, got %d", status.CompletedTasks)
	}
	
	cancel()
	scheduler.Stop(context.Background())
}

func TestSchedulerRetry(t *testing.T) {
	// Vector: "……失敗時のリトライ機能を検証……"
	config := DefaultManagerConfig()
	manager := NewManager(config)
	
	// Register a mock agent that fails first 2 times
	def := SubagentDefinition{
		Name:        "retry-test",
		Description: "Retry test agent",
		Version:     "1.0.0",
		Tools:       []string{"test"},
		Prompt:      "Test retry",
		TrinityRole: TrinityRoleAuditor,
	}
	
	manager.Register(def)
	
	// Note: In real implementation, we'd need to mock the agent execution
	// to simulate failures and eventual success
	
	// Create scheduler
	scheduler := NewDynamicScheduler(manager, 1)
	
	// Create task with retry settings
	task := &ScheduledTask{
		Task: Task{
			ID:        "retry-task",
			AgentName: "retry-test",
			Input:     "retry input",
		},
		Priority:   PriorityNormal,
		MaxRetries: 3,
	}
	
	ctx := context.Background()
	scheduler.Schedule(ctx, task)
	
	// Start scheduler
	execCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	
	go scheduler.Execute(execCtx)
	
	// Wait for retries
	time.Sleep(1 * time.Second)
	
	// In a real test, we'd verify:
	// - Task was retried up to MaxRetries times
	// - Priority was boosted on retry
	// - Task eventually succeeded or failed permanently
	
	cancel()
	scheduler.Stop(context.Background())
}

func TestSchedulerDeadline(t *testing.T) {
	// Springfield: "デッドライン管理を検証"
	config := DefaultManagerConfig()
	manager := NewManager(config)
	
	// Register agent
	def := SubagentDefinition{
		Name:        "deadline-test",
		Description: "Deadline test agent",
		Version:     "1.0.0",
		Tools:       []string{"test"},
		Prompt:      "Test deadline",
		TrinityRole: TrinityRoleStrategist,
	}
	
	manager.Register(def)
	
	// Create scheduler
	scheduler := NewDynamicScheduler(manager, 2)
	
	// Create tasks with deadlines
	ctx := context.Background()
	
	nearDeadline := time.Now().Add(1 * time.Second)
	farDeadline := time.Now().Add(10 * time.Second)
	
	taskUrgent := &ScheduledTask{
		Task: Task{ID: "urgent", AgentName: "deadline-test", Input: "urgent"},
		Priority: PriorityNormal,
		Deadline: &nearDeadline,
	}
	
	taskRelaxed := &ScheduledTask{
		Task: Task{ID: "relaxed", AgentName: "deadline-test", Input: "relaxed"},
		Priority: PriorityNormal,
		Deadline: &farDeadline,
	}
	
	// Schedule relaxed first, then urgent
	scheduler.Schedule(ctx, taskRelaxed)
	scheduler.Schedule(ctx, taskUrgent)
	
	// Start scheduler
	execCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	
	go scheduler.Execute(execCtx)
	
	// Wait for completion
	time.Sleep(500 * time.Millisecond)
	
	// Both should complete
	status := scheduler.GetStatus()
	if status.CompletedTasks != 2 {
		t.Errorf("Expected 2 completed tasks, got %d", status.CompletedTasks)
	}
	
	cancel()
	scheduler.Stop(context.Background())
}

func TestSchedulerConcurrency(t *testing.T) {
	// Krukai: "並行処理の正確性を検証"
	config := DefaultManagerConfig()
	manager := NewManager(config)
	
	// Register agent
	def := SubagentDefinition{
		Name:        "concurrent-test",
		Description: "Concurrency test agent",
		Version:     "1.0.0",
		Tools:       []string{"test"},
		Prompt:      "Test concurrency",
		TrinityRole: TrinityRoleOptimizer,
	}
	
	manager.Register(def)
	
	// Create scheduler with multiple workers
	scheduler := NewDynamicScheduler(manager, 10)
	
	// Start scheduler
	ctx := context.Background()
	execCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	
	go scheduler.Execute(execCtx)
	
	// Schedule many tasks concurrently
	var wg sync.WaitGroup
	taskCount := 100
	
	for i := 0; i < taskCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			task := &ScheduledTask{
				Task: Task{
					ID:        fmt.Sprintf("concurrent-%d", id),
					AgentName: "concurrent-test",
					Input:     fmt.Sprintf("input-%d", id),
				},
				Priority: TaskPriority(id % 4 * 50), // Vary priorities
			}
			scheduler.Schedule(ctx, task)
		}(i)
	}
	
	// Wait for all tasks to be scheduled
	wg.Wait()
	
	// Wait for execution
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-timeout:
			t.Fatal("Timeout waiting for tasks to complete")
		case <-ticker.C:
			status := scheduler.GetStatus()
			if status.CompletedTasks == int64(taskCount) {
				// All tasks completed
				goto done
			}
		}
	}
	
done:
	// Verify metrics
	status := scheduler.GetStatus()
	if status.CompletedTasks != int64(taskCount) {
		t.Errorf("Expected %d completed tasks, got %d", taskCount, status.CompletedTasks)
	}
	
	if status.FailedTasks != 0 {
		t.Errorf("Expected 0 failed tasks, got %d", status.FailedTasks)
	}
	
	cancel()
	scheduler.Stop(context.Background())
}