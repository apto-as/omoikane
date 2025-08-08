// Package subagent implements dynamic task scheduling for parallel execution
// Springfield: "効率的なタスク配分で全体最適化を図ります"
// Krukai: "動的スケジューリングで処理効率を最大化"
// Vector: "……リソース制限を厳密に監視……"
package subagent

import (
	"container/heap"
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// TaskPriority defines task execution priority levels
type TaskPriority int

const (
	PriorityLow    TaskPriority = 0
	PriorityNormal TaskPriority = 50
	PriorityHigh   TaskPriority = 100
	PriorityCritical TaskPriority = 200
)

// ScheduledTask represents a task with scheduling metadata
type ScheduledTask struct {
	Task
	Priority    TaskPriority
	Dependencies []string // Task IDs that must complete before this task
	Deadline    *time.Time
	RetryCount  int
	MaxRetries  int
	QueuedAt    time.Time
	ScheduledAt *time.Time
	EstimatedDuration time.Duration
}

// TaskScheduler manages dynamic task scheduling and execution
type TaskScheduler interface {
	// Schedule adds a task to the scheduler
	Schedule(ctx context.Context, task *ScheduledTask) error
	
	// ScheduleBatch adds multiple tasks at once
	ScheduleBatch(ctx context.Context, tasks []*ScheduledTask) error
	
	// Execute starts the scheduler and processes tasks
	Execute(ctx context.Context) error
	
	// GetStatus returns current scheduler status
	GetStatus() *SchedulerStatus
	
	// Stop gracefully stops the scheduler
	Stop(ctx context.Context) error
	
	// SetMaxWorkers adjusts the number of concurrent workers
	SetMaxWorkers(count int)
}

// SchedulerStatus contains scheduler runtime information
type SchedulerStatus struct {
	ActiveWorkers   int32
	MaxWorkers      int32
	QueuedTasks     int32
	CompletedTasks  int64
	FailedTasks     int64
	TotalLatency    time.Duration
	AverageLatency  time.Duration
	CurrentLoad     float64
	StartedAt       time.Time
	LastTaskAt      *time.Time
}

// DynamicScheduler implements TaskScheduler with dynamic worker pool
// Krukai: "効率的なgoroutineプールで最適化"
type DynamicScheduler struct {
	manager        Manager
	taskQueue      *PriorityQueue
	workers        sync.Pool
	activeWorkers  atomic.Int32
	maxWorkers     atomic.Int32
	completedTasks atomic.Int64
	failedTasks    atomic.Int64
	totalLatency   atomic.Int64
	
	// Dependency tracking
	dependencies   map[string][]string  // task ID -> dependent task IDs
	completedDeps  map[string]bool      // completed task IDs
	depMutex       sync.RWMutex
	
	// Task results
	results        map[string]interface{}
	resultsMutex   sync.RWMutex
	
	// Control
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	startedAt      time.Time
	lastTaskAt     atomic.Value // *time.Time
}

// NewDynamicScheduler creates a new dynamic task scheduler
func NewDynamicScheduler(manager Manager, maxWorkers int) *DynamicScheduler {
	if maxWorkers <= 0 {
		maxWorkers = 10 // Default to Claude Code's limit
	}
	
	scheduler := &DynamicScheduler{
		manager:       manager,
		taskQueue:     NewPriorityQueue(),
		dependencies:  make(map[string][]string),
		completedDeps: make(map[string]bool),
		results:       make(map[string]interface{}),
		startedAt:     time.Now(),
	}
	
	scheduler.maxWorkers.Store(int32(maxWorkers))
	
	// Initialize worker pool
	scheduler.workers = sync.Pool{
		New: func() interface{} {
			return &taskWorker{
				scheduler: scheduler,
			}
		},
	}
	
	return scheduler
}

// Schedule adds a task to the scheduler
func (s *DynamicScheduler) Schedule(ctx context.Context, task *ScheduledTask) error {
	// Vector: "……タスクの妥当性を検証……"
	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}
	
	if task.ID == "" {
		task.ID = fmt.Sprintf("task-%d", time.Now().UnixNano())
	}
	
	if task.Priority == 0 {
		task.Priority = PriorityNormal
	}
	
	task.QueuedAt = time.Now()
	
	// Track dependencies
	if len(task.Dependencies) > 0 {
		s.depMutex.Lock()
		s.dependencies[task.ID] = task.Dependencies
		s.depMutex.Unlock()
	}
	
	// Add to priority queue
	heap.Push(s.taskQueue, task)
	
	// Wake up a worker if available
	s.wakeWorker()
	
	return nil
}

// ScheduleBatch adds multiple tasks at once
func (s *DynamicScheduler) ScheduleBatch(ctx context.Context, tasks []*ScheduledTask) error {
	// Springfield: "バッチ処理で効率化を図ります"
	for _, task := range tasks {
		if err := s.Schedule(ctx, task); err != nil {
			return fmt.Errorf("failed to schedule task %s: %w", task.ID, err)
		}
	}
	return nil
}

// Execute starts the scheduler and processes tasks
func (s *DynamicScheduler) Execute(ctx context.Context) error {
	s.ctx, s.cancel = context.WithCancel(ctx)
	
	// Start worker management goroutine
	s.wg.Add(1)
	go s.manageWorkers()
	
	// Start initial workers
	initialWorkers := int(s.maxWorkers.Load() / 2)
	if initialWorkers < 1 {
		initialWorkers = 1
	}
	
	for i := 0; i < initialWorkers; i++ {
		s.startWorker()
	}
	
	// Wait for context cancellation
	<-s.ctx.Done()
	
	// Wait for all workers to finish
	s.wg.Wait()
	
	return nil
}

// GetStatus returns current scheduler status
func (s *DynamicScheduler) GetStatus() *SchedulerStatus {
	completedCount := s.completedTasks.Load()
	totalLatencyNanos := s.totalLatency.Load()
	
	var avgLatency time.Duration
	if completedCount > 0 {
		avgLatency = time.Duration(totalLatencyNanos / completedCount)
	}
	
	activeWorkers := s.activeWorkers.Load()
	maxWorkers := s.maxWorkers.Load()
	var load float64
	if maxWorkers > 0 {
		load = float64(activeWorkers) / float64(maxWorkers)
	}
	
	var lastTask *time.Time
	if val := s.lastTaskAt.Load(); val != nil {
		lastTask = val.(*time.Time)
	}
	
	return &SchedulerStatus{
		ActiveWorkers:  activeWorkers,
		MaxWorkers:     maxWorkers,
		QueuedTasks:    int32(s.taskQueue.Len()),
		CompletedTasks: completedCount,
		FailedTasks:    s.failedTasks.Load(),
		TotalLatency:   time.Duration(totalLatencyNanos),
		AverageLatency: avgLatency,
		CurrentLoad:    load,
		StartedAt:      s.startedAt,
		LastTaskAt:     lastTask,
	}
}

// Stop gracefully stops the scheduler
func (s *DynamicScheduler) Stop(ctx context.Context) error {
	// Vector: "……安全な停止手順を実行……"
	if s.cancel != nil {
		s.cancel()
	}
	
	// Wait for workers with timeout
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("scheduler stop timeout: %w", ctx.Err())
	}
}

// SetMaxWorkers adjusts the number of concurrent workers
func (s *DynamicScheduler) SetMaxWorkers(count int) {
	if count < 1 {
		count = 1
	}
	s.maxWorkers.Store(int32(count))
	
	// Wake workers if we increased the limit
	currentActive := s.activeWorkers.Load()
	if int32(count) > currentActive {
		for i := currentActive; i < int32(count) && s.taskQueue.Len() > 0; i++ {
			s.wakeWorker()
		}
	}
}

// Private methods

func (s *DynamicScheduler) manageWorkers() {
	defer s.wg.Done()
	
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.adjustWorkers()
		}
	}
}

func (s *DynamicScheduler) adjustWorkers() {
	// Krukai: "動的にワーカー数を最適化"
	queueSize := s.taskQueue.Len()
	activeWorkers := s.activeWorkers.Load()
	maxWorkers := s.maxWorkers.Load()
	
	// Scale up if queue is growing
	if queueSize > int(activeWorkers)*2 && activeWorkers < maxWorkers {
		newWorkers := min(int(maxWorkers-activeWorkers), queueSize/2)
		for i := 0; i < newWorkers; i++ {
			s.startWorker()
		}
	}
	
	// Scale down handled by workers exiting when idle
}

func (s *DynamicScheduler) startWorker() {
	// Ensure context is initialized
	if s.ctx == nil {
		return
	}
	
	s.wg.Add(1)
	s.activeWorkers.Add(1)
	
	go func() {
		defer s.wg.Done()
		defer s.activeWorkers.Add(-1)
		
		worker := s.workers.Get().(*taskWorker)
		defer s.workers.Put(worker)
		
		worker.run(s.ctx)
	}()
}

func (s *DynamicScheduler) wakeWorker() {
	// Try to start a new worker if under limit
	if s.activeWorkers.Load() < s.maxWorkers.Load() {
		s.startWorker()
	}
}

func (s *DynamicScheduler) canExecuteTask(task *ScheduledTask) bool {
	// Check dependencies
	if len(task.Dependencies) == 0 {
		return true
	}
	
	s.depMutex.RLock()
	defer s.depMutex.RUnlock()
	
	for _, dep := range task.Dependencies {
		if !s.completedDeps[dep] {
			return false
		}
	}
	
	return true
}

func (s *DynamicScheduler) getNextTask() *ScheduledTask {
	for s.taskQueue.Len() > 0 {
		task := heap.Pop(s.taskQueue).(*ScheduledTask)
		
		if s.canExecuteTask(task) {
			return task
		}
		
		// Re-queue task if dependencies not met
		heap.Push(s.taskQueue, task)
		time.Sleep(10 * time.Millisecond) // Brief pause to avoid busy loop
		
		// Check if there are other tasks
		if s.taskQueue.Len() == 1 {
			// Only dependent tasks remain
			return nil
		}
	}
	
	return nil
}

func (s *DynamicScheduler) executeTask(ctx context.Context, task *ScheduledTask) error {
	now := time.Now()
	task.ScheduledAt = &now
	
	// Execute via manager
	result, err := s.manager.Execute(ctx, task.AgentName, task.Input)
	
	// Record completion
	executionTime := time.Since(now)
	s.totalLatency.Add(int64(executionTime))
	
	if err != nil {
		s.failedTasks.Add(1)
		
		// Retry logic
		if task.RetryCount < task.MaxRetries {
			task.RetryCount++
			task.Priority = PriorityHigh // Boost priority for retry
			s.Schedule(ctx, task)
			return fmt.Errorf("task %s failed, retrying: %w", task.ID, err)
		}
		
		return fmt.Errorf("task %s failed after %d retries: %w", task.ID, task.MaxRetries, err)
	}
	
	s.completedTasks.Add(1)
	completionTime := time.Now()
	s.lastTaskAt.Store(&completionTime)
	
	// Store result
	s.resultsMutex.Lock()
	s.results[task.ID] = result
	s.resultsMutex.Unlock()
	
	// Mark as completed for dependencies
	s.depMutex.Lock()
	s.completedDeps[task.ID] = true
	s.depMutex.Unlock()
	
	return nil
}

// taskWorker processes tasks from the queue
type taskWorker struct {
	scheduler *DynamicScheduler
}

func (w *taskWorker) run(ctx context.Context) {
	idleTimeout := 30 * time.Second
	timer := time.NewTimer(idleTimeout)
	defer timer.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			// Exit if idle for too long
			return
		default:
			task := w.scheduler.getNextTask()
			if task == nil {
				// No tasks available, wait briefly
				time.Sleep(100 * time.Millisecond)
				timer.Reset(idleTimeout)
				continue
			}
			
			// Reset idle timer
			timer.Stop()
			
			// Execute task
			taskCtx := ctx
			if task.Deadline != nil {
				var cancel context.CancelFunc
				taskCtx, cancel = context.WithDeadline(ctx, *task.Deadline)
				defer cancel()
			}
			
			if err := w.scheduler.executeTask(taskCtx, task); err != nil {
				// Log error (in production, use proper logging)
				fmt.Printf("Task execution error: %v\n", err)
			}
			
			timer.Reset(idleTimeout)
		}
	}
}

// PriorityQueue implements a priority queue for tasks
type PriorityQueue struct {
	items []*ScheduledTask
	mu    sync.Mutex
}

// NewPriorityQueue creates a new priority queue
func NewPriorityQueue() *PriorityQueue {
	return &PriorityQueue{
		items: make([]*ScheduledTask, 0),
	}
}

func (pq *PriorityQueue) Len() int {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	return len(pq.items)
}

func (pq *PriorityQueue) Less(i, j int) bool {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	
	if i < 0 || j < 0 || i >= len(pq.items) || j >= len(pq.items) {
		return false
	}
	
	// Higher priority first
	if pq.items[i].Priority != pq.items[j].Priority {
		return pq.items[i].Priority > pq.items[j].Priority
	}
	
	// Earlier deadline first
	if pq.items[i].Deadline != nil && pq.items[j].Deadline != nil {
		return pq.items[i].Deadline.Before(*pq.items[j].Deadline)
	}
	
	// Earlier queued first
	return pq.items[i].QueuedAt.Before(pq.items[j].QueuedAt)
}

func (pq *PriorityQueue) Swap(i, j int) {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	if i < 0 || j < 0 || i >= len(pq.items) || j >= len(pq.items) {
		return
	}
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
}

func (pq *PriorityQueue) Push(x interface{}) {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	pq.items = append(pq.items, x.(*ScheduledTask))
}

func (pq *PriorityQueue) Pop() interface{} {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	
	n := len(pq.items)
	if n == 0 {
		return nil
	}
	
	item := pq.items[n-1]
	pq.items = pq.items[:n-1]
	return item
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}