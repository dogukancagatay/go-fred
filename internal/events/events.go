package events

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// EventType constants for different event types
const (
	EventTypeTaskCreated   = "task.created"
	EventTypeTaskStarted   = "task.started"
	EventTypeTaskCompleted = "task.completed"
	EventTypeTaskFailed    = "task.failed"
	EventTypeTaskCancelled = "task.cancelled"
)

// EventBuilder helps build events with common patterns
type EventBuilder struct {
	event Event
}

// NewEventBuilder creates a new event builder
func NewEventBuilder(eventType string) *EventBuilder {
	return &EventBuilder{
		event: Event{
			ID:        uuid.New().String(),
			Type:      eventType,
			Timestamp: time.Now(),
			Data:      make(map[string]interface{}),
			Source:    "go-fred",
		},
	}
}

// WithData adds data to the event
func (b *EventBuilder) WithData(key string, value interface{}) *EventBuilder {
	b.event.Data[key] = value
	return b
}

// WithTaskID adds task ID to the event data
func (b *EventBuilder) WithTaskID(taskID string) *EventBuilder {
	b.event.Data["task_id"] = taskID
	return b
}

// WithError adds error information to the event data
func (b *EventBuilder) WithError(err error) *EventBuilder {
	b.event.Data["error"] = err.Error()
	return b
}

// WithDuration adds duration information to the event data
func (b *EventBuilder) WithDuration(duration time.Duration) *EventBuilder {
	b.event.Data["duration_ms"] = duration.Milliseconds()
	return b
}

// Build returns the built event
func (b *EventBuilder) Build() Event {
	return b.event
}

// PublishTaskCreated publishes a task created event
func PublishTaskCreated(ctx context.Context, publisher Publisher, taskID, taskType string, isAsync bool) error {
	event := NewEventBuilder(EventTypeTaskCreated).
		WithTaskID(taskID).
		WithData("task_type", taskType).
		WithData("is_async", isAsync).
		Build()

	return publisher.Publish(ctx, event)
}

// PublishTaskStarted publishes a task started event
func PublishTaskStarted(ctx context.Context, publisher Publisher, taskID string) error {
	event := NewEventBuilder(EventTypeTaskStarted).
		WithTaskID(taskID).
		Build()

	return publisher.Publish(ctx, event)
}

// PublishTaskCompleted publishes a task completed event
func PublishTaskCompleted(ctx context.Context, publisher Publisher, taskID string, duration time.Duration, result interface{}) error {
	event := NewEventBuilder(EventTypeTaskCompleted).
		WithTaskID(taskID).
		WithDuration(duration).
		WithData("result", result).
		Build()

	return publisher.Publish(ctx, event)
}

// PublishTaskFailed publishes a task failed event
func PublishTaskFailed(ctx context.Context, publisher Publisher, taskID string, duration time.Duration, err error) error {
	event := NewEventBuilder(EventTypeTaskFailed).
		WithTaskID(taskID).
		WithDuration(duration).
		WithError(err).
		Build()

	return publisher.Publish(ctx, event)
}

// PublishTaskCancelled publishes a task cancelled event
func PublishTaskCancelled(ctx context.Context, publisher Publisher, taskID string) error {
	event := NewEventBuilder(EventTypeTaskCancelled).
		WithTaskID(taskID).
		Build()

	return publisher.Publish(ctx, event)
}

// PublishCustomEvent publishes a custom event with the given type and data
func PublishCustomEvent(ctx context.Context, publisher Publisher, eventType string, data map[string]interface{}) error {
	builder := NewEventBuilder(eventType)
	for key, value := range data {
		builder.WithData(key, value)
	}

	event := builder.Build()
	return publisher.Publish(ctx, event)
}
