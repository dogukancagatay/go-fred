package events

import (
	"context"
	"testing"
	"time"

	"go-fred/internal/config"
)

func TestNewPublisher(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.EventsConfig
		expectError bool
	}{
		{
			name: "noop publisher",
			config: &config.EventsConfig{
				Publisher: "noop",
			},
			expectError: false,
		},
		{
			name: "empty publisher defaults to noop",
			config: &config.EventsConfig{
				Publisher: "",
			},
			expectError: false,
		},
		{
			name: "kafka publisher with valid config",
			config: &config.EventsConfig{
				Publisher: "kafka",
				Kafka: config.KafkaConfig{
					Brokers: []string{"localhost:9092"},
					Topic:   "test-topic",
				},
			},
			expectError: false,
		},
		{
			name: "kafka publisher with no brokers",
			config: &config.EventsConfig{
				Publisher: "kafka",
				Kafka: config.KafkaConfig{
					Brokers: []string{},
					Topic:   "test-topic",
				},
			},
			expectError: true,
		},
		{
			name: "kafka publisher with no topic",
			config: &config.EventsConfig{
				Publisher: "kafka",
				Kafka: config.KafkaConfig{
					Brokers: []string{"localhost:9092"},
					Topic:   "",
				},
			},
			expectError: true,
		},
		{
			name: "unsupported publisher",
			config: &config.EventsConfig{
				Publisher: "unsupported",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			publisher, err := NewPublisher(tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				if publisher != nil {
					t.Errorf("Expected nil publisher but got %v", publisher)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if publisher == nil {
					t.Error("Expected publisher but got nil")
				}
			}
		})
	}
}

func TestNoOpPublisher(t *testing.T) {
	publisher := NewNoOpPublisher()
	if publisher == nil {
		t.Fatal("Expected publisher but got nil")
	}

	ctx := context.Background()
	event := Event{
		ID:        "test-id",
		Type:      "test.type",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"key": "value"},
		Source:    "test",
	}

	// Test publish
	err := publisher.Publish(ctx, event)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Test close
	err = publisher.Close()
	if err != nil {
		t.Errorf("Unexpected error on close: %v", err)
	}
}

func TestEventBuilder(t *testing.T) {
	builder := NewEventBuilder("test.type")
	if builder == nil {
		t.Fatal("Expected builder but got nil")
	}

	event := builder.
		WithData("key1", "value1").
		WithData("key2", 42).
		WithTaskID("task-123").
		WithError(nil).
		Build()

	if event.Type != "test.type" {
		t.Errorf("Expected type 'test.type', got '%s'", event.Type)
	}
	if event.ID == "" {
		t.Error("Expected non-empty ID")
	}
	if event.Source != "go-fred" {
		t.Errorf("Expected source 'go-fred', got '%s'", event.Source)
	}
	if event.Data["key1"] != "value1" {
		t.Errorf("Expected key1 'value1', got '%v'", event.Data["key1"])
	}
	if event.Data["key2"] != 42 {
		t.Errorf("Expected key2 42, got '%v'", event.Data["key2"])
	}
	if event.Data["task_id"] != "task-123" {
		t.Errorf("Expected task_id 'task-123', got '%v'", event.Data["task_id"])
	}
}

func TestEventBuilderWithError(t *testing.T) {
	testErr := &testError{message: "test error"}
	builder := NewEventBuilder("test.type")
	event := builder.WithError(testErr).Build()

	if event.Data["error"] != "test error" {
		t.Errorf("Expected error 'test error', got '%v'", event.Data["error"])
	}
}

func TestEventBuilderWithDuration(t *testing.T) {
	duration := 5 * time.Second
	builder := NewEventBuilder("test.type")
	event := builder.WithDuration(duration).Build()

	if event.Data["duration_ms"] != int64(5000) {
		t.Errorf("Expected duration_ms 5000, got '%v'", event.Data["duration_ms"])
	}
}

func TestPublishTaskCreated(t *testing.T) {
	publisher := NewNoOpPublisher()
	ctx := context.Background()

	err := PublishTaskCreated(ctx, publisher, "task-123", "echo", true)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestPublishTaskStarted(t *testing.T) {
	publisher := NewNoOpPublisher()
	ctx := context.Background()

	err := PublishTaskStarted(ctx, publisher, "task-123")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestPublishTaskCompleted(t *testing.T) {
	publisher := NewNoOpPublisher()
	ctx := context.Background()
	duration := 2 * time.Second
	result := map[string]interface{}{"status": "success"}

	err := PublishTaskCompleted(ctx, publisher, "task-123", duration, result)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestPublishTaskFailed(t *testing.T) {
	publisher := NewNoOpPublisher()
	ctx := context.Background()
	duration := 1 * time.Second
	testErr := &testError{message: "task failed"}

	err := PublishTaskFailed(ctx, publisher, "task-123", duration, testErr)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestPublishTaskCancelled(t *testing.T) {
	publisher := NewNoOpPublisher()
	ctx := context.Background()

	err := PublishTaskCancelled(ctx, publisher, "task-123")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestPublishCustomEvent(t *testing.T) {
	publisher := NewNoOpPublisher()
	ctx := context.Background()
	data := map[string]interface{}{
		"custom_key": "custom_value",
		"number":     123,
	}

	err := PublishCustomEvent(ctx, publisher, "custom.type", data)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// testError is a simple error type for testing
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}
