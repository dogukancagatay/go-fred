package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

// Task represents a task in the system
type Task struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Status      TaskStatus             `json:"status"`
	Input       map[string]interface{} `json:"input"`
	Output      map[string]interface{} `json:"output,omitempty"`
	Error       string                 `json:"error,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Duration    *time.Duration         `json:"duration_ms,omitempty"`
	IsAsync     bool                   `json:"is_async"`
}

// TaskRequest represents a request to create a task
type TaskRequest struct {
	Type   string                 `json:"type" binding:"required"`
	Input  map[string]interface{} `json:"input"`
	Async  bool                   `json:"async,omitempty"`
}

// TaskResponse represents the response for a task
type TaskResponse struct {
	Task *Task `json:"task"`
}

// TaskListResponse represents the response for listing tasks
type TaskListResponse struct {
	Tasks []Task `json:"tasks"`
	Total int    `json:"total"`
}

// NewTask creates a new task with the given parameters
func NewTask(taskType string, input map[string]interface{}, isAsync bool) *Task {
	now := time.Now()
	return &Task{
		ID:        uuid.New().String(),
		Type:      taskType,
		Status:    TaskStatusPending,
		Input:     input,
		CreatedAt: now,
		IsAsync:   isAsync,
	}
}

// Start marks the task as started
func (t *Task) Start() {
	now := time.Now()
	t.Status = TaskStatusRunning
	t.StartedAt = &now
}

// Complete marks the task as completed with the given output
func (t *Task) Complete(output map[string]interface{}) {
	now := time.Now()
	t.Status = TaskStatusCompleted
	t.CompletedAt = &now
	t.Output = output

	if t.StartedAt != nil {
		duration := now.Sub(*t.StartedAt)
		t.Duration = &duration
	}
}

// Fail marks the task as failed with the given error
func (t *Task) Fail(err error) {
	now := time.Now()
	t.Status = TaskStatusFailed
	t.CompletedAt = &now
	t.Error = err.Error()

	if t.StartedAt != nil {
		duration := now.Sub(*t.StartedAt)
		t.Duration = &duration
	}
}

// Cancel marks the task as cancelled
func (t *Task) Cancel() {
	now := time.Now()
	t.Status = TaskStatusCancelled
	t.CompletedAt = &now

	if t.StartedAt != nil {
		duration := now.Sub(*t.StartedAt)
		t.Duration = &duration
	}
}

// IsFinished returns true if the task is in a finished state
func (t *Task) IsFinished() bool {
	return t.Status == TaskStatusCompleted ||
		   t.Status == TaskStatusFailed ||
		   t.Status == TaskStatusCancelled
}

// ToJSON converts the task to JSON string
func (t *Task) ToJSON() (string, error) {
	data, err := json.Marshal(t)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FromJSON creates a task from JSON string
func FromJSON(jsonStr string) (*Task, error) {
	var task Task
	err := json.Unmarshal([]byte(jsonStr), &task)
	if err != nil {
		return nil, err
	}
	return &task, nil
}
