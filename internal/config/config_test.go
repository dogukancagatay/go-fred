package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Create a temporary config file
	configContent := `
server:
  host: "test-host"
  port: 9090

events:
  publisher: "kafka"
  kafka:
    brokers: ["localhost:9092"]
    topic: "test-topic"

tasks:
  max_concurrent: 5
  timeout_seconds: 600
`

	tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write config content: %v", err)
	}
	tmpFile.Close()

	// Test loading config
	config, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test server config
	if config.Server.Host != "test-host" {
		t.Errorf("Expected host 'test-host', got '%s'", config.Server.Host)
	}
	if config.Server.Port != 9090 {
		t.Errorf("Expected port 9090, got %d", config.Server.Port)
	}

	// Test events config
	if config.Events.Publisher != "kafka" {
		t.Errorf("Expected publisher 'kafka', got '%s'", config.Events.Publisher)
	}
	if len(config.Events.Kafka.Brokers) != 1 || config.Events.Kafka.Brokers[0] != "localhost:9092" {
		t.Errorf("Expected broker 'localhost:9092', got %v", config.Events.Kafka.Brokers)
	}
	if config.Events.Kafka.Topic != "test-topic" {
		t.Errorf("Expected topic 'test-topic', got '%s'", config.Events.Kafka.Topic)
	}

	// Test tasks config
	if config.Tasks.MaxConcurrent != 5 {
		t.Errorf("Expected max_concurrent 5, got %d", config.Tasks.MaxConcurrent)
	}
	if config.Tasks.TimeoutSeconds != 600 {
		t.Errorf("Expected timeout_seconds 600, got %d", config.Tasks.TimeoutSeconds)
	}
}

func TestLoadWithDefaults(t *testing.T) {
	// Create a minimal config file
	configContent := `
server:
  port: 8080
`

	tmpFile, err := os.CreateTemp("", "test-config-minimal-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write config content: %v", err)
	}
	tmpFile.Close()

	// Test loading config
	config, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test defaults
	if config.Server.Host != "localhost" {
		t.Errorf("Expected default host 'localhost', got '%s'", config.Server.Host)
	}
	if config.Server.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", config.Server.Port)
	}
	if config.Events.Publisher != "noop" {
		t.Errorf("Expected default publisher 'noop', got '%s'", config.Events.Publisher)
	}
	if config.Tasks.MaxConcurrent != 10 {
		t.Errorf("Expected default max_concurrent 10, got %d", config.Tasks.MaxConcurrent)
	}
	if config.Tasks.TimeoutSeconds != 300 {
		t.Errorf("Expected default timeout_seconds 300, got %d", config.Tasks.TimeoutSeconds)
	}
}

func TestLoadEmptyFile(t *testing.T) {
	// Create an empty config file
	tmpFile, err := os.CreateTemp("", "test-config-empty-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Test loading config
	config, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test defaults
	if config.Server.Host != "localhost" {
		t.Errorf("Expected default host 'localhost', got '%s'", config.Server.Host)
	}
	if config.Server.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", config.Server.Port)
	}
	if config.Events.Publisher != "noop" {
		t.Errorf("Expected default publisher 'noop', got '%s'", config.Events.Publisher)
	}
}

func TestLoadInvalidFile(t *testing.T) {
	// Test loading non-existent file
	_, err := Load("non-existent-file.yaml")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	// Create a file with invalid YAML
	configContent := `
server:
  host: "test-host"
  port: invalid-port
`

	tmpFile, err := os.CreateTemp("", "test-config-invalid-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write config content: %v", err)
	}
	tmpFile.Close()

	// Test loading invalid config
	_, err = Load(tmpFile.Name())
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}
}

func TestGetAddress(t *testing.T) {
	config := &Config{
		Server: ServerConfig{
			Host: "example.com",
			Port: 9000,
		},
	}

	expected := "example.com:9000"
	actual := config.GetAddress()
	if actual != expected {
		t.Errorf("Expected address '%s', got '%s'", expected, actual)
	}
}
