package server

import (
	"context"
	"testing"
	"time"

	"go-fred-rest/internal/config"
	"go-fred-rest/internal/events"
	"go-fred-rest/internal/tasks"
)

func TestNew(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Events: config.EventsConfig{
			Publisher: "noop",
		},
		Tasks: config.TasksConfig{
			MaxConcurrent: 10,
		},
	}

	server := New(cfg)

	if server == nil {
		t.Fatal("Expected non-nil server")
	}

	if server.config != cfg {
		t.Error("Expected config to be set")
	}

	if server.router == nil {
		t.Error("Expected non-nil router")
	}

	if server.taskManager == nil {
		t.Error("Expected non-nil task manager")
	}

	if server.eventPub == nil {
		t.Error("Expected non-nil event publisher")
	}
}

func TestServerStart(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 0, // Use random port for testing
		},
		Events: config.EventsConfig{
			Publisher: "noop",
		},
		Tasks: config.TasksConfig{
			MaxConcurrent: 10,
		},
	}

	server := New(cfg)

	// Start server in a goroutine
	serverStarted := make(chan bool)
	go func() {
		serverStarted <- true
		server.Start()
	}()

	// Wait for server to start
	select {
	case <-serverStarted:
		// Server started successfully
	case <-time.After(1 * time.Second):
		t.Fatal("Server failed to start within timeout")
	}

	// Stop the server
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := server.Stop(ctx)
	if err != nil {
		t.Errorf("Unexpected error stopping server: %v", err)
	}
}

func TestServerStop(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 0, // Use random port for testing
		},
		Events: config.EventsConfig{
			Publisher: "noop",
		},
		Tasks: config.TasksConfig{
			MaxConcurrent: 10,
		},
	}

	server := New(cfg)

	// Test stopping server that hasn't been started
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := server.Stop(ctx)
	if err != nil {
		t.Errorf("Unexpected error stopping unstarted server: %v", err)
	}
}

func TestCorsMiddleware(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 0,
		},
		Events: config.EventsConfig{
			Publisher: "noop",
		},
		Tasks: config.TasksConfig{
			MaxConcurrent: 10,
		},
	}

	server := New(cfg)

	// Test that CORS middleware is applied
	// This is tested indirectly through the handlers_test.go file
	// where we test the OPTIONS request handling
	if server.router == nil {
		t.Error("Expected non-nil router")
	}
}

func TestServerWithKafkaConfig(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 0,
		},
		Events: config.EventsConfig{
			Publisher: "kafka",
			Kafka: config.KafkaConfig{
				Brokers: []string{"localhost:9092"},
				Topic:   "test-topic",
			},
		},
		Tasks: config.TasksConfig{
			MaxConcurrent: 10,
		},
	}

	// This should not panic even with Kafka config
	// (Kafka connection will fail, but server creation should succeed)
	server := New(cfg)

	if server == nil {
		t.Fatal("Expected non-nil server")
	}

	if server.eventPub == nil {
		t.Error("Expected non-nil event publisher")
	}
}

func TestServerWithInvalidEventPublisher(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 0,
		},
		Events: config.EventsConfig{
			Publisher: "invalid",
		},
		Tasks: config.TasksConfig{
			MaxConcurrent: 10,
		},
	}

	// This should panic due to invalid event publisher
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for invalid event publisher")
		}
	}()

	New(cfg)
}

func TestServerTaskManagerIntegration(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 0,
		},
		Events: config.EventsConfig{
			Publisher: "noop",
		},
		Tasks: config.TasksConfig{
			MaxConcurrent: 5,
		},
	}

	server := New(cfg)

	// Test that task manager is properly configured
	if server.taskManager == nil {
		t.Fatal("Expected non-nil task manager")
	}

	// Test that default executors are registered
	registry := server.taskManager.(*tasks.TaskManager)
	// We can't access the registry directly, but we can test through task creation
	task, err := server.taskManager.CreateTask("echo", map[string]interface{}{"message": "test"}, false)
	if err != nil {
		t.Fatalf("Unexpected error creating task: %v", err)
	}

	if task == nil {
		t.Fatal("Expected non-nil task")
	}

	if task.Type != "echo" {
		t.Errorf("Expected task type 'echo', got %s", task.Type)
	}
}

func TestServerEventPublisherIntegration(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 0,
		},
		Events: config.EventsConfig{
			Publisher: "noop",
		},
		Tasks: config.TasksConfig{
			MaxConcurrent: 10,
		},
	}

	server := New(cfg)

	// Test that event publisher is properly configured
	if server.eventPub == nil {
		t.Fatal("Expected non-nil event publisher")
	}

	// Test that event publisher works
	ctx := context.Background()
	event := events.Event{
		ID:        "test-id",
		Type:      "test.type",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"key": "value"},
		Source:    "test",
	}

	err := server.eventPub.Publish(ctx, event)
	if err != nil {
		t.Errorf("Unexpected error publishing event: %v", err)
	}

	// Test that event publisher can be closed
	err = server.eventPub.Close()
	if err != nil {
		t.Errorf("Unexpected error closing event publisher: %v", err)
	}
}
