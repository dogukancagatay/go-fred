package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewTask(t *testing.T) {
	taskType := "echo"
	input := map[string]interface{}{"message": "hello"}
	isAsync := true

	task := NewTask(taskType, input, isAsync)

	if task.ID == "" {
		t.Error("Expected non-empty ID")
	}
	if task.Type != taskType {
		t.Errorf("Expected type %s, got %s", taskType, task.Type)
	}
	if task.Status != TaskStatusPending {
		t.Errorf("Expected status %s, got %s", TaskStatusPending, task.Status)
	}
	if task.Input["message"] != "hello" {
		t.Errorf("Expected input message 'hello', got %v", task.Input["message"])
	}
	if task.IsAsync != isAsync {
		t.Errorf("Expected IsAsync %v, got %v", isAsync, task.IsAsync)
	}
	if task.CreatedAt.IsZero() {
		t.Error("Expected non-zero CreatedAt")
	}
}

func TestTaskStart(t *testing.T) {
	task := NewTask("echo", map[string]interface{}{}, false)

	task.Start()

	if task.Status != TaskStatusRunning {
		t.Errorf("Expected status %s, got %s", TaskStatusRunning, task.Status)
	}
	if task.StartedAt == nil {
		t.Error("Expected non-nil StartedAt")
	}
	if task.StartedAt.IsZero() {
		t.Error("Expected non-zero StartedAt")
	}
}

func TestTaskComplete(t *testing.T) {
	task := NewTask("echo", map[string]interface{}{}, false)
	task.Start()

	// Wait a bit to ensure duration is non-zero
	time.Sleep(10 * time.Millisecond)

	output := map[string]interface{}{"result": "success"}
	task.Complete(output)

	if task.Status != TaskStatusCompleted {
		t.Errorf("Expected status %s, got %s", TaskStatusCompleted, task.Status)
	}
	if task.CompletedAt == nil {
		t.Error("Expected non-nil CompletedAt")
	}
	if task.CompletedAt.IsZero() {
		t.Error("Expected non-zero CompletedAt")
	}
	if task.Output["result"] != "success" {
		t.Errorf("Expected output result 'success', got %v", task.Output["result"])
	}
	if task.Duration == nil {
		t.Error("Expected non-nil Duration")
	}
	if *task.Duration <= 0 {
		t.Errorf("Expected positive duration, got %v", *task.Duration)
	}
}

func TestTaskFail(t *testing.T) {
	task := NewTask("echo", map[string]interface{}{}, false)
	task.Start()

	// Wait a bit to ensure duration is non-zero
	time.Sleep(10 * time.Millisecond)

	testErr := &testError{message: "task failed"}
	task.Fail(testErr)

	if task.Status != TaskStatusFailed {
		t.Errorf("Expected status %s, got %s", TaskStatusFailed, task.Status)
	}
	if task.CompletedAt == nil {
		t.Error("Expected non-nil CompletedAt")
	}
	if task.CompletedAt.IsZero() {
		t.Error("Expected non-zero CompletedAt")
	}
	if task.Error != "task failed" {
		t.Errorf("Expected error 'task failed', got %s", task.Error)
	}
	if task.Duration == nil {
		t.Error("Expected non-nil Duration")
	}
	if *task.Duration <= 0 {
		t.Errorf("Expected positive duration, got %v", *task.Duration)
	}
}

func TestTaskCancel(t *testing.T) {
	task := NewTask("echo", map[string]interface{}{}, false)
	task.Start()

	// Wait a bit to ensure duration is non-zero
	time.Sleep(10 * time.Millisecond)

	task.Cancel()

	if task.Status != TaskStatusCancelled {
		t.Errorf("Expected status %s, got %s", TaskStatusCancelled, task.Status)
	}
	if task.CompletedAt == nil {
		t.Error("Expected non-nil CompletedAt")
	}
	if task.CompletedAt.IsZero() {
		t.Error("Expected non-zero CompletedAt")
	}
	if task.Duration == nil {
		t.Error("Expected non-nil Duration")
	}
	if *task.Duration <= 0 {
		t.Errorf("Expected positive duration, got %v", *task.Duration)
	}
}

func TestTaskIsFinished(t *testing.T) {
	tests := []struct {
		name     string
		status   TaskStatus
		expected bool
	}{
		{"pending", TaskStatusPending, false},
		{"running", TaskStatusRunning, false},
		{"completed", TaskStatusCompleted, true},
		{"failed", TaskStatusFailed, true},
		{"cancelled", TaskStatusCancelled, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := NewTask("echo", map[string]interface{}{}, false)
			task.Status = tt.status

			result := task.IsFinished()
			if result != tt.expected {
				t.Errorf("Expected IsFinished() to return %v for status %s, got %v",
					tt.expected, tt.status, result)
			}
		})
	}
}

func TestTaskToJSON(t *testing.T) {
	task := NewTask("echo", map[string]interface{}{"message": "hello"}, false)
	task.Start()
	task.Complete(map[string]interface{}{"result": "success"})

	jsonStr, err := task.ToJSON()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if jsonStr == "" {
		t.Error("Expected non-empty JSON string")
	}

	// Verify it's valid JSON by unmarshaling
	var unmarshaled map[string]interface{}
	err = json.Unmarshal([]byte(jsonStr), &unmarshaled)
	if err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	if unmarshaled["type"] != "echo" {
		t.Errorf("Expected type 'echo', got %v", unmarshaled["type"])
	}
	if unmarshaled["status"] != "completed" {
		t.Errorf("Expected status 'completed', got %v", unmarshaled["status"])
	}
}

func TestFromJSON(t *testing.T) {
	// Create a task and convert to JSON
	originalTask := NewTask("echo", map[string]interface{}{"message": "hello"}, false)
	originalTask.Start()
	originalTask.Complete(map[string]interface{}{"result": "success"})

	jsonStr, err := originalTask.ToJSON()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Convert back from JSON
	restoredTask, err := FromJSON(jsonStr)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Compare key fields
	if restoredTask.ID != originalTask.ID {
		t.Errorf("Expected ID %s, got %s", originalTask.ID, restoredTask.ID)
	}
	if restoredTask.Type != originalTask.Type {
		t.Errorf("Expected type %s, got %s", originalTask.Type, restoredTask.Type)
	}
	if restoredTask.Status != originalTask.Status {
		t.Errorf("Expected status %s, got %s", originalTask.Status, restoredTask.Status)
	}
	if restoredTask.IsAsync != originalTask.IsAsync {
		t.Errorf("Expected IsAsync %v, got %v", originalTask.IsAsync, restoredTask.IsAsync)
	}
}

func TestFromJSONInvalid(t *testing.T) {
	_, err := FromJSON("invalid json")
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestTaskRequestValidation(t *testing.T) {
	// Test valid request
	validRequest := TaskRequest{
		Type:  "echo",
		Input: map[string]interface{}{"message": "hello"},
		Async: false,
	}

	if validRequest.Type == "" {
		t.Error("Expected non-empty Type")
	}

	// Test with empty type (should be caught by binding validation)
	invalidRequest := TaskRequest{
		Type:  "",
		Input: map[string]interface{}{"message": "hello"},
		Async: false,
	}

	if invalidRequest.Type != "" {
		t.Error("Expected empty Type for invalid request")
	}
}

func TestTaskResponse(t *testing.T) {
	task := NewTask("echo", map[string]interface{}{"message": "hello"}, false)
	response := TaskResponse{Task: task}

	if response.Task == nil {
		t.Error("Expected non-nil Task")
	}
	if response.Task.Type != "echo" {
		t.Errorf("Expected task type 'echo', got %s", response.Task.Type)
	}
}

func TestTaskListResponse(t *testing.T) {
	tasks := []*Task{
		NewTask("echo", map[string]interface{}{"message": "hello1"}, false),
		NewTask("sleep", map[string]interface{}{"duration": 5}, true),
	}

	response := TaskListResponse{
		Tasks: make([]Task, len(tasks)),
		Total: len(tasks),
	}

	for i, task := range tasks {
		response.Tasks[i] = *task
	}

	if response.Total != 2 {
		t.Errorf("Expected total 2, got %d", response.Total)
	}
	if len(response.Tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(response.Tasks))
	}
	if response.Tasks[0].Type != "echo" {
		t.Errorf("Expected first task type 'echo', got %s", response.Tasks[0].Type)
	}
	if response.Tasks[1].Type != "sleep" {
		t.Errorf("Expected second task type 'sleep', got %s", response.Tasks[1].Type)
	}
}

// testError is a simple error type for testing
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}
