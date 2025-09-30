package tasks

import (
	"context"
	"fmt"
	"time"

	"go-fred-rest/internal/models"
)

// EchoExecutor is a simple executor that echoes the input
type EchoExecutor struct{}

// Execute implements the TaskExecutor interface
func (e *EchoExecutor) Execute(ctx context.Context, task *models.Task) error {
	// Simulate some work
	time.Sleep(100 * time.Millisecond)

	// Echo the input as output
	task.Output = map[string]interface{}{
		"echo": task.Input,
		"message": "Task executed successfully",
	}

	return nil
}

// GetSupportedTypes returns the supported task types
func (e *EchoExecutor) GetSupportedTypes() []string {
	return []string{"echo"}
}

// SleepExecutor is an executor that sleeps for a specified duration
type SleepExecutor struct{}

// Execute implements the TaskExecutor interface
func (s *SleepExecutor) Execute(ctx context.Context, task *models.Task) error {
	// Get sleep duration from input
	duration, ok := task.Input["duration"].(float64)
	if !ok {
		return fmt.Errorf("duration must be a number")
	}

	sleepDuration := time.Duration(duration) * time.Second

	// Check for context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(sleepDuration):
		// Sleep completed
	}

	task.Output = map[string]interface{}{
		"slept_for_seconds": duration,
		"message": "Sleep completed successfully",
	}

	return nil
}

// GetSupportedTypes returns the supported task types
func (s *SleepExecutor) GetSupportedTypes() []string {
	return []string{"sleep"}
}

// ErrorExecutor is an executor that always fails
type ErrorExecutor struct{}

// Execute implements the TaskExecutor interface
func (e *ErrorExecutor) Execute(ctx context.Context, task *models.Task) error {
	// Get error message from input
	message, ok := task.Input["message"].(string)
	if !ok {
		message = "Task failed as requested"
	}

	return fmt.Errorf("%s", message)
}

// GetSupportedTypes returns the supported task types
func (e *ErrorExecutor) GetSupportedTypes() []string {
	return []string{"error"}
}

// MathExecutor is an executor that performs basic math operations
type MathExecutor struct{}

// Execute implements the TaskExecutor interface
func (m *MathExecutor) Execute(ctx context.Context, task *models.Task) error {
	// Get operation and operands from input
	operation, ok := task.Input["operation"].(string)
	if !ok {
		return fmt.Errorf("operation must be a string")
	}

	a, ok := task.Input["a"].(float64)
	if !ok {
		return fmt.Errorf("a must be a number")
	}

	b, ok := task.Input["b"].(float64)
	if !ok {
		return fmt.Errorf("b must be a number")
	}

	var result float64

	switch operation {
	case "add":
		result = a + b
	case "subtract":
		result = a - b
	case "multiply":
		result = a * b
	case "divide":
		if b == 0 {
			return fmt.Errorf("division by zero")
		}
		result = a / b
	default:
		return fmt.Errorf("unsupported operation: %s", operation)
	}

	task.Output = map[string]interface{}{
		"operation": operation,
		"a": a,
		"b": b,
		"result": result,
	}

	return nil
}

// GetSupportedTypes returns the supported task types
func (m *MathExecutor) GetSupportedTypes() []string {
	return []string{"math"}
}

// RegisterDefaultExecutors registers the default task executors
func RegisterDefaultExecutors(registry *ExecutorRegistry) {
	registry.Register("echo", &EchoExecutor{})
	registry.Register("sleep", &SleepExecutor{})
	registry.Register("error", &ErrorExecutor{})
	registry.Register("math", &MathExecutor{})
}
