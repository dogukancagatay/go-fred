package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server ServerConfig `yaml:"server"`
	Events EventsConfig `yaml:"events"`
	Tasks  TasksConfig  `yaml:"tasks"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// EventsConfig holds event publisher configuration
type EventsConfig struct {
	Publisher string      `yaml:"publisher"`
	Kafka     KafkaConfig `yaml:"kafka"`
}

// KafkaConfig holds Kafka-specific configuration
type KafkaConfig struct {
	Brokers []string `yaml:"brokers"`
	Topic   string   `yaml:"topic"`
}

// TasksConfig holds task execution configuration
type TasksConfig struct {
	MaxConcurrent  int `yaml:"max_concurrent"`
	TimeoutSeconds int `yaml:"timeout_seconds"`
}

// Load reads and parses the configuration file
func Load(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults
	if config.Server.Host == "" {
		config.Server.Host = "localhost"
	}
	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}
	if config.Events.Publisher == "" {
		config.Events.Publisher = "noop"
	}
	if config.Tasks.MaxConcurrent == 0 {
		config.Tasks.MaxConcurrent = 10
	}
	if config.Tasks.TimeoutSeconds == 0 {
		config.Tasks.TimeoutSeconds = 300
	}

	return &config, nil
}

// GetAddress returns the server address
func (c *Config) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}
