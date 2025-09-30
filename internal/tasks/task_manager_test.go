package tasks

import (
	"context"
	"testing"
	"time"

	"go-fred-rest/internal/models"
)

func TestTaskManagerCreateTask(t *testing.T) {
	registry := NewExecutorRegistry()
	RegisterDefaultExecutors(registry)

	mockPub := &mockPublisher{}
	taskManager := NewTaskManager(registry, mockPub, 5)

	// Test creating a valid task
	task, err := taskManager.CreateTask("echo", map[string]interface{}{"message": "hello"}, false)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if task == nil {
		t.Fatal("Expected non-nil task")
	}

	if task.Type != "echo" {
		t.Errorf("Expected type 'echo', got %s", task.Type)
	}

	if task.Status != models.TaskStatusPending {
		t.Errorf("Expected status 'pending', got %s", task.Status)
	}

	if task.IsAsync != false {
		t.Errorf("Expected IsAsync false, got %v", task.IsAsync)
	}

	// Test that task was stored
	retrievedTask, err := taskManager.GetTask(task.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if retrievedTask.ID != task.ID {
		t.Errorf("Expected task ID %s, got %s", task.ID, retrievedTask.ID)
	}

	// Test that event was published
	events := mockPub.GetEvents()
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}

	if events[0].Type != events.EventTypeTaskCreated {
		t.Errorf("Expected event type %s, got %s", events.EventTypeTaskCreated, events[0].Type)
	}
}

func TestTaskManagerCreateTaskInvalidType(t *testing.T) {
	registry := NewExecutorRegistry()
	RegisterDefaultExecutors(registry)

	mockPub := &mockPublisher{}
	taskManager := NewTaskManager(registry, mockPub, 5)

	// Test creating task with invalid type
	_, err := taskManager.CreateTask("invalid", map[string]interface{}{}, false)
	if err == nil {
		t.Error("Expected error for invalid task type")
	}
}

func TestTaskManagerGetTask(t *testing.T) {
	registry := NewExecutorRegistry()
	RegisterDefaultExecutors(registry)

	mockPub := &mockPublisher{}
	taskManager := NewTaskManager(registry, mockPub, 5)

	// Test getting non-existent task
	_, err := taskManager.GetTask("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent task")
	}

	// Create a task
	task, err := taskManager.CreateTask("echo", map[string]interface{}{}, false)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Test getting existing task
	retrievedTask, err := taskManager.GetTask(task.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if retrievedTask.ID != task.ID {
		t.Errorf("Expected task ID %s, got %s", task.ID, retrievedTask.ID)
	}
}

func TestTaskManagerListTasks(t *testing.T) {
	registry := NewExecutorRegistry()
	RegisterDefaultExecutors(registry)

	mockPub := &mockPublisher{}
	taskManager := NewTaskManager(registry, mockPub, 5)

	// Test empty list
	tasks := taskManager.ListTasks()
	if len(tasks) != 0 {
		t.Errorf("Expected empty list, got %d tasks", len(tasks))
	}

	// Create some tasks
	task1, err := taskManager.CreateTask("echo", map[string]interface{}{}, false)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	task2, err := taskManager.CreateTask("sleep", map[string]interface{}{}, true)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Test listing tasks
	tasks = taskManager.ListTasks()
	if len(tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(tasks))
	}

	// Check that both tasks are in the list
	found1, found2 := false, false
	for _, task := range tasks {
		if task.ID == task1.ID {
			found1 = true
		}
		if task.ID == task2.ID {
			found2 = true
		}
	}

	if !found1 {
		t.Error("Task 1 not found in list")
	}
	if !found2 {
		t.Error("Task 2 not found in list")
	}
}

func TestTaskManagerExecuteTask(t *testing.T) {
	registry := NewExecutorRegistry()
	RegisterDefaultExecutors(registry)

	mockPub := &mockPublisher{}
	taskManager := NewTaskManager(registry, mockPub, 5)

	// Create a task
	task, err := taskManager.CreateTask("echo", map[string]interface{}{"message": "hello"}, false)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Execute the task
	err = taskManager.ExecuteTask(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check that task was executed
	executedTask, err := taskManager.GetTask(task.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if executedTask.Status != models.TaskStatusCompleted {
		t.Errorf("Expected status 'completed', got %s", executedTask.Status)
	}

	if executedTask.Output == nil {
		t.Error("Expected non-nil output")
	}

	// Check that events were published
	events := mockPub.GetEvents()
	if len(events) < 3 { // created, started, completed
		t.Errorf("Expected at least 3 events, got %d", len(events))
	}

	// Check event types
	eventTypes := make(map[string]bool)
	for _, event := range events {
		eventTypes[event.Type] = true
	}

	if !eventTypes[events.EventTypeTaskCreated] {
		t.Error("Expected task.created event")
	}
	if !eventTypes[events.EventTypeTaskStarted] {
		t.Error("Expected task.started event")
	}
	if !eventTypes[events.EventTypeTaskCompleted] {
		t.Error("Expected task.completed event")
	}
}

func TestTaskManagerExecuteTaskNonExistent(t *testing.T) {
	registry := NewExecutorRegistry()
	RegisterDefaultExecutors(registry)

	mockPub := &mockPublisher{}
	taskManager := NewTaskManager(registry, mockPub, 5)

	// Test executing non-existent task
	err := taskManager.ExecuteTask(context.Background(), "non-existent")
	if err == nil {
		t.Error("Expected error for non-existent task")
	}
}

func TestTaskManagerExecuteTaskAlreadyFinished(t *testing.T) {
	registry := NewExecutorRegistry()
	RegisterDefaultExecutors(registry)

	mockPub := &mockPublisher{}
	taskManager := NewTaskManager(registry, mockPub, 5)

	// Create and execute a task
	task, err := taskManager.CreateTask("echo", map[string]interface{}{}, false)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	err = taskManager.ExecuteTask(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Try to execute again
	err = taskManager.ExecuteTask(context.Background(), task.ID)
	if err == nil {
		t.Error("Expected error for already finished task")
	}
}

func TestTaskManagerExecuteTaskAsync(t *testing.T) {
	registry := NewExecutorRegistry()
	RegisterDefaultExecutors(registry)

	mockPub := &mockPublisher{}
	taskManager := NewTaskManager(registry, mockPub, 5)

	// Create a task
	task, err := taskManager.CreateTask("sleep", map[string]interface{}{"duration": 0.1}, false)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Execute the task asynchronously
	err = taskManager.ExecuteTaskAsync(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Wait a bit for the task to complete
	time.Sleep(200 * time.Millisecond)

	// Check that task was executed
	executedTask, err := taskManager.GetTask(task.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if executedTask.Status != models.TaskStatusCompleted {
		t.Errorf("Expected status 'completed', got %s", executedTask.Status)
	}
}

func TestTaskManagerCancelTask(t *testing.T) {
	registry := NewExecutorRegistry()
	RegisterDefaultExecutors(registry)

	mockPub := &mockPublisher{}
	taskManager := NewTaskManager(registry, mockPub, 5)

	// Create a task
	task, err := taskManager.CreateTask("echo", map[string]interface{}{}, false)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Start the task
	task.Start()

	// Cancel the task
	err = taskManager.CancelTask(task.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check that task was cancelled
	cancelledTask, err := taskManager.GetTask(task.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if cancelledTask.Status != models.TaskStatusCancelled {
		t.Errorf("Expected status 'cancelled', got %s", cancelledTask.Status)
	}

	// Check that event was published
	events := mockPub.GetEvents()
	if len(events) < 2 { // created, cancelled
		t.Errorf("Expected at least 2 events, got %d", len(events))
	}

	// Check for cancelled event
	foundCancelled := false
	for _, event := range events {
		if event.Type == events.EventTypeTaskCancelled {
			foundCancelled = true
			break
		}
	}

	if !foundCancelled {
		t.Error("Expected task.cancelled event")
	}
}

func TestTaskManagerCancelTaskNonExistent(t *testing.T) {
	registry := NewExecutorRegistry()
	RegisterDefaultExecutors(registry)

	mockPub := &mockPublisher{}
	taskManager := NewTaskManager(registry, mockPub, 5)

	// Test cancelling non-existent task
	err := taskManager.CancelTask("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent task")
	}
}

func TestTaskManagerCancelTaskAlreadyFinished(t *testing.T) {
	registry := NewExecutorRegistry()
	RegisterDefaultExecutors(registry)

	mockPub := &mockPublisher{}
	taskManager := NewTaskManager(registry, mockPub, 5)

	// Create and execute a task
	task, err := taskManager.CreateTask("echo", map[string]interface{}{}, false)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	err = taskManager.ExecuteTask(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Try to cancel the finished task
	err = taskManager.CancelTask(task.ID)
	if err == nil {
		t.Error("Expected error for already finished task")
	}
}

func TestTaskManagerConcurrentExecution(t *testing.T) {
	registry := NewExecutorRegistry()
	RegisterDefaultExecutors(registry)

	mockPub := &mockPublisher{}
	taskManager := NewTaskManager(registry, mockPub, 2) // Max 2 concurrent

	// Create multiple tasks
	tasks := make([]*models.Task, 3)
	for i := 0; i < 3; i++ {
		task, err := taskManager.CreateTask("sleep", map[string]interface{}{"duration": 0.1}, false)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		tasks[i] = task
	}

	// Execute all tasks concurrently
	start := time.Now()
	for _, task := range tasks {
		go func(taskID string) {
			taskManager.ExecuteTask(context.Background(), taskID)
		}(task.ID)
	}

	// Wait for all tasks to complete
	time.Sleep(500 * time.Millisecond)

	// Check that all tasks completed
	for _, task := range tasks {
		completedTask, err := taskManager.GetTask(task.ID)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if completedTask.Status != models.TaskStatusCompleted {
			t.Errorf("Expected task %s to be completed, got %s", task.ID, completedTask.Status)
		}
	}

	// Check that execution was limited by semaphore
	// With max 2 concurrent and 3 tasks, total time should be at least 200ms
	duration := time.Since(start)
	if duration < 200*time.Millisecond {
		t.Errorf("Expected at least 200ms duration due to concurrency limit, got %v", duration)
	}
}
