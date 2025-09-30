package tasks

import (
	"context"
	"errors"
	"testing"
	"time"

	"go-fred/internal/events"
	"go-fred/internal/models"
)

// mockPublisher is a mock event publisher for testing
type mockPublisher struct {
	events []events.Event
}

func (m *mockPublisher) Publish(ctx context.Context, event events.Event) error {
	m.events = append(m.events, event)
	return nil
}

func (m *mockPublisher) Close() error {
	return nil
}

func (m *mockPublisher) GetEvents() []events.Event {
	return m.events
}

func (m *mockPublisher) ClearEvents() {
	m.events = nil
}

func TestExecutorRegistry(t *testing.T) {
	registry := NewExecutorRegistry()

	// Test initial state
	types := registry.GetSupportedTypes()
	if len(types) != 0 {
		t.Errorf("Expected empty types, got %v", types)
	}

	// Test registering executor
	executor := &EchoExecutor{}
	registry.Register("echo", executor)

	// Test getting executor
	retrievedExecutor, err := registry.GetExecutor("echo")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if retrievedExecutor != executor {
		t.Error("Expected same executor instance")
	}

	// Test getting non-existent executor
	_, err = registry.GetExecutor("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent executor")
	}

	// Test supported types
	types = registry.GetSupportedTypes()
	if len(types) != 1 || types[0] != "echo" {
		t.Errorf("Expected ['echo'], got %v", types)
	}
}

func TestEchoExecutor(t *testing.T) {
	executor := &EchoExecutor{}

	// Test supported types
	types := executor.GetSupportedTypes()
	if len(types) != 1 || types[0] != "echo" {
		t.Errorf("Expected ['echo'], got %v", types)
	}

	// Test execution
	task := models.NewTask("echo", map[string]interface{}{"message": "hello"}, false)

	err := executor.Execute(context.Background(), task)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check output
	if task.Output == nil {
		t.Fatal("Expected non-nil output")
	}

	echoData, ok := task.Output["echo"]
	if !ok {
		t.Error("Expected 'echo' key in output")
	}

	echoMap, ok := echoData.(map[string]interface{})
	if !ok {
		t.Error("Expected echo data to be a map")
	}

	if echoMap["message"] != "hello" {
		t.Errorf("Expected message 'hello', got %v", echoMap["message"])
	}

	message, ok := task.Output["message"]
	if !ok {
		t.Error("Expected 'message' key in output")
	}

	if message != "Task executed successfully" {
		t.Errorf("Expected success message, got %v", message)
	}
}

func TestSleepExecutor(t *testing.T) {
	executor := &SleepExecutor{}

	// Test supported types
	types := executor.GetSupportedTypes()
	if len(types) != 1 || types[0] != "sleep" {
		t.Errorf("Expected ['sleep'], got %v", types)
	}

	// Test execution with valid duration
	task := models.NewTask("sleep", map[string]interface{}{"duration": 0.1}, false)

	start := time.Now()
	err := executor.Execute(context.Background(), task)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check that it actually slept (at least 100ms)
	if duration < 100*time.Millisecond {
		t.Errorf("Expected at least 100ms sleep, got %v", duration)
	}

	// Check output
	if task.Output == nil {
		t.Fatal("Expected non-nil output")
	}

	sleptFor, ok := task.Output["slept_for_seconds"]
	if !ok {
		t.Error("Expected 'slept_for_seconds' key in output")
	}

	if sleptFor != 0.1 {
		t.Errorf("Expected slept_for_seconds 0.1, got %v", sleptFor)
	}

	message, ok := task.Output["message"]
	if !ok {
		t.Error("Expected 'message' key in output")
	}

	if message != "Sleep completed successfully" {
		t.Errorf("Expected success message, got %v", message)
	}
}

func TestSleepExecutorInvalidInput(t *testing.T) {
	executor := &SleepExecutor{}

	// Test with invalid duration type
	task := models.NewTask("sleep", map[string]interface{}{"duration": "invalid"}, false)

	err := executor.Execute(context.Background(), task)
	if err == nil {
		t.Error("Expected error for invalid duration type")
	}
}

func TestSleepExecutorContextCancellation(t *testing.T) {
	executor := &SleepExecutor{}

	// Test with context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	task := models.NewTask("sleep", map[string]interface{}{"duration": 10}, false)

	// Cancel context after a short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := executor.Execute(ctx, task)
	if err == nil {
		t.Error("Expected error due to context cancellation")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}
}

func TestErrorExecutor(t *testing.T) {
	executor := &ErrorExecutor{}

	// Test supported types
	types := executor.GetSupportedTypes()
	if len(types) != 1 || types[0] != "error" {
		t.Errorf("Expected ['error'], got %v", types)
	}

	// Test execution with custom error message
	task := models.NewTask("error", map[string]interface{}{"message": "custom error"}, false)

	err := executor.Execute(context.Background(), task)
	if err == nil {
		t.Error("Expected error")
	}

	if err.Error() != "custom error" {
		t.Errorf("Expected 'custom error', got %v", err)
	}
}

func TestErrorExecutorDefaultMessage(t *testing.T) {
	executor := &ErrorExecutor{}

	// Test execution with invalid message type
	task := models.NewTask("error", map[string]interface{}{"message": 123}, false)

	err := executor.Execute(context.Background(), task)
	if err == nil {
		t.Error("Expected error")
	}

	if err.Error() != "Task failed as requested" {
		t.Errorf("Expected default error message, got %v", err)
	}
}

func TestMathExecutor(t *testing.T) {
	executor := &MathExecutor{}

	// Test supported types
	types := executor.GetSupportedTypes()
	if len(types) != 1 || types[0] != "math" {
		t.Errorf("Expected ['math'], got %v", types)
	}

	tests := []struct {
		name      string
		operation string
		a         float64
		b         float64
		expected  float64
	}{
		{"add", "add", 10, 5, 15},
		{"subtract", "subtract", 10, 5, 5},
		{"multiply", "multiply", 10, 5, 50},
		{"divide", "divide", 10, 5, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := models.NewTask("math", map[string]interface{}{
				"operation": tt.operation,
				"a":         tt.a,
				"b":         tt.b,
			}, false)

			err := executor.Execute(context.Background(), task)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Check output
			if task.Output == nil {
				t.Fatal("Expected non-nil output")
			}

			operation, ok := task.Output["operation"]
			if !ok {
				t.Error("Expected 'operation' key in output")
			}
			if operation != tt.operation {
				t.Errorf("Expected operation %s, got %v", tt.operation, operation)
			}

			a, ok := task.Output["a"]
			if !ok {
				t.Error("Expected 'a' key in output")
			}
			if a != tt.a {
				t.Errorf("Expected a %v, got %v", tt.a, a)
			}

			b, ok := task.Output["b"]
			if !ok {
				t.Error("Expected 'b' key in output")
			}
			if b != tt.b {
				t.Errorf("Expected b %v, got %v", tt.b, b)
			}

			result, ok := task.Output["result"]
			if !ok {
				t.Error("Expected 'result' key in output")
			}
			if result != tt.expected {
				t.Errorf("Expected result %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestMathExecutorDivisionByZero(t *testing.T) {
	executor := &MathExecutor{}

	task := models.NewTask("math", map[string]interface{}{
		"operation": "divide",
		"a":         10,
		"b":         0,
	}, false)

	err := executor.Execute(context.Background(), task)
	if err == nil {
		t.Error("Expected error for division by zero")
	}

	if err.Error() != "division by zero" {
		t.Errorf("Expected 'division by zero' error, got %v", err)
	}
}

func TestMathExecutorInvalidOperation(t *testing.T) {
	executor := &MathExecutor{}

	task := models.NewTask("math", map[string]interface{}{
		"operation": "invalid",
		"a":         10,
		"b":         5,
	}, false)

	err := executor.Execute(context.Background(), task)
	if err == nil {
		t.Error("Expected error for invalid operation")
	}

	expectedErr := "unsupported operation: invalid"
	if err.Error() != expectedErr {
		t.Errorf("Expected '%s' error, got %v", expectedErr, err)
	}
}

func TestMathExecutorInvalidInput(t *testing.T) {
	executor := &MathExecutor{}

	tests := []struct {
		name string
		input map[string]interface{}
		expectedError string
	}{
		{
			name: "missing operation",
			input: map[string]interface{}{
				"a": 10,
				"b": 5,
			},
			expectedError: "operation must be a string",
		},
		{
			name: "invalid operation type",
			input: map[string]interface{}{
				"operation": 123,
				"a": 10,
				"b": 5,
			},
			expectedError: "operation must be a string",
		},
		{
			name: "missing a",
			input: map[string]interface{}{
				"operation": "add",
				"b": 5,
			},
			expectedError: "a must be a number",
		},
		{
			name: "invalid a type",
			input: map[string]interface{}{
				"operation": "add",
				"a": "invalid",
				"b": 5,
			},
			expectedError: "a must be a number",
		},
		{
			name: "missing b",
			input: map[string]interface{}{
				"operation": "add",
				"a": 10,
			},
			expectedError: "b must be a number",
		},
		{
			name: "invalid b type",
			input: map[string]interface{}{
				"operation": "add",
				"a": 10,
				"b": "invalid",
			},
			expectedError: "b must be a number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := models.NewTask("math", tt.input, false)

			err := executor.Execute(context.Background(), task)
			if err == nil {
				t.Error("Expected error")
			}

			if err.Error() != tt.expectedError {
				t.Errorf("Expected '%s' error, got %v", tt.expectedError, err)
			}
		})
	}
}

func TestRegisterDefaultExecutors(t *testing.T) {
	registry := NewExecutorRegistry()

	RegisterDefaultExecutors(registry)

	// Test that all default executors are registered
	expectedTypes := []string{"echo", "sleep", "error", "math"}
	types := registry.GetSupportedTypes()

	if len(types) != len(expectedTypes) {
		t.Errorf("Expected %d types, got %d", len(expectedTypes), len(types))
	}

	for _, expectedType := range expectedTypes {
		found := false
		for _, actualType := range types {
			if actualType == expectedType {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected type %s not found in %v", expectedType, types)
		}
	}

	// Test that executors can be retrieved
	for _, expectedType := range expectedTypes {
		executor, err := registry.GetExecutor(expectedType)
		if err != nil {
			t.Errorf("Unexpected error getting executor for %s: %v", expectedType, err)
		}
		if executor == nil {
			t.Errorf("Expected non-nil executor for %s", expectedType)
		}
	}
}
