package tasks

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go-fred/internal/events"
	"go-fred/internal/models"
)

// TaskExecutor defines the interface for executing tasks
type TaskExecutor interface {
	Execute(ctx context.Context, task *models.Task) error
	GetSupportedTypes() []string
}

// ExecutorRegistry manages task executors
type ExecutorRegistry struct {
	executors map[string]TaskExecutor
	mu        sync.RWMutex
}

// NewExecutorRegistry creates a new executor registry
func NewExecutorRegistry() *ExecutorRegistry {
	return &ExecutorRegistry{
		executors: make(map[string]TaskExecutor),
	}
}

// Register registers a task executor for a specific task type
func (r *ExecutorRegistry) Register(taskType string, executor TaskExecutor) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.executors[taskType] = executor
}

// GetExecutor returns the executor for the given task type
func (r *ExecutorRegistry) GetExecutor(taskType string) (TaskExecutor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	executor, exists := r.executors[taskType]
	if !exists {
		return nil, fmt.Errorf("no executor found for task type: %s", taskType)
	}
	return executor, nil
}

// GetSupportedTypes returns all supported task types
func (r *ExecutorRegistry) GetSupportedTypes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]string, 0, len(r.executors))
	for taskType := range r.executors {
		types = append(types, taskType)
	}
	return types
}

// TaskManager manages task execution and storage
type TaskManager struct {
	registry     *ExecutorRegistry
	eventPub     events.Publisher
	tasks        map[string]*models.Task
	mu           sync.RWMutex
	maxConcurrent int
	semaphore    chan struct{}
}

// NewTaskManager creates a new task manager
func NewTaskManager(registry *ExecutorRegistry, eventPub events.Publisher, maxConcurrent int) *TaskManager {
	return &TaskManager{
		registry:      registry,
		eventPub:      eventPub,
		tasks:         make(map[string]*models.Task),
		maxConcurrent: maxConcurrent,
		semaphore:     make(chan struct{}, maxConcurrent),
	}
}

// CreateTask creates a new task
func (tm *TaskManager) CreateTask(taskType string, input map[string]interface{}, isAsync bool) (*models.Task, error) {
	// Check if executor exists for this task type
	_, err := tm.registry.GetExecutor(taskType)
	if err != nil {
		return nil, err
	}

	task := models.NewTask(taskType, input, isAsync)

	tm.mu.Lock()
	tm.tasks[task.ID] = task
	tm.mu.Unlock()

	// Publish task created event
	ctx := context.Background()
	events.PublishTaskCreated(ctx, tm.eventPub, task.ID, taskType, isAsync)

	return task, nil
}

// GetTask retrieves a task by ID
func (tm *TaskManager) GetTask(taskID string) (*models.Task, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	task, exists := tm.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}
	return task, nil
}

// ListTasks returns all tasks
func (tm *TaskManager) ListTasks() []*models.Task {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tasks := make([]*models.Task, 0, len(tm.tasks))
	for _, task := range tm.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

// ExecuteTask executes a task synchronously
func (tm *TaskManager) ExecuteTask(ctx context.Context, taskID string) error {
	task, err := tm.GetTask(taskID)
	if err != nil {
		return err
	}

	if task.IsFinished() {
		return fmt.Errorf("task %s is already finished", taskID)
	}

	// Acquire semaphore
	select {
	case tm.semaphore <- struct{}{}:
		defer func() { <-tm.semaphore }()
	case <-ctx.Done():
		return ctx.Err()
	}

	return tm.executeTaskInternal(ctx, task)
}

// ExecuteTaskAsync executes a task asynchronously
func (tm *TaskManager) ExecuteTaskAsync(ctx context.Context, taskID string) error {
	task, err := tm.GetTask(taskID)
	if err != nil {
		return err
	}

	if task.IsFinished() {
		return fmt.Errorf("task %s is already finished", taskID)
	}

	// Start execution in background
	go func() {
		// Acquire semaphore
		tm.semaphore <- struct{}{}
		defer func() { <-tm.semaphore }()

		// Create new context for background execution
		bgCtx := context.Background()
		tm.executeTaskInternal(bgCtx, task)
	}()

	return nil
}

// executeTaskInternal performs the actual task execution
func (tm *TaskManager) executeTaskInternal(ctx context.Context, task *models.Task) error {
	startTime := time.Now()

	// Mark task as started
	task.Start()
	events.PublishTaskStarted(ctx, tm.eventPub, task.ID)

	// Get executor for task type
	executor, err := tm.registry.GetExecutor(task.Type)
	if err != nil {
		task.Fail(err)
		events.PublishTaskFailed(ctx, tm.eventPub, task.ID, time.Since(startTime), err)
		return err
	}

	// Execute the task
	err = executor.Execute(ctx, task)

	duration := time.Since(startTime)

	if err != nil {
		task.Fail(err)
		events.PublishTaskFailed(ctx, tm.eventPub, task.ID, duration, err)
		return err
	}

	// Task completed successfully
	task.Complete(task.Output)
	events.PublishTaskCompleted(ctx, tm.eventPub, task.ID, duration, task.Output)

	return nil
}

// CancelTask cancels a running task
func (tm *TaskManager) CancelTask(taskID string) error {
	task, err := tm.GetTask(taskID)
	if err != nil {
		return err
	}

	if task.IsFinished() {
		return fmt.Errorf("task %s is already finished", taskID)
	}

	task.Cancel()

	ctx := context.Background()
	events.PublishTaskCancelled(ctx, tm.eventPub, taskID)

	return nil
}
