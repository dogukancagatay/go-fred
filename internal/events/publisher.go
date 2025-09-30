package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"go-fred/internal/config"

	"github.com/segmentio/kafka-go"
)

// Event represents a system event
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
	Source    string                 `json:"source"`
}

// Publisher defines the interface for event publishing
type Publisher interface {
	Publish(ctx context.Context, event Event) error
	Close() error
}

// NewPublisher creates a new event publisher based on configuration
func NewPublisher(cfg *config.EventsConfig) (Publisher, error) {
	switch cfg.Publisher {
	case "kafka":
		return NewKafkaPublisher(cfg.Kafka)
	case "noop", "":
		return NewNoOpPublisher(), nil
	default:
		return nil, fmt.Errorf("unsupported event publisher: %s", cfg.Publisher)
	}
}

// NoOpPublisher is a no-operation publisher that only logs events
type NoOpPublisher struct{}

// NewNoOpPublisher creates a new no-op publisher
func NewNoOpPublisher() *NoOpPublisher {
	return &NoOpPublisher{}
}

// Publish logs the event
func (p *NoOpPublisher) Publish(ctx context.Context, event Event) error {
	eventJSON, _ := json.MarshalIndent(event, "", "  ")
	log.Printf("Event published (no-op): %s", string(eventJSON))
	return nil
}

// Close does nothing for no-op publisher
func (p *NoOpPublisher) Close() error {
	return nil
}

// KafkaPublisher publishes events to Kafka
type KafkaPublisher struct {
	writer *kafka.Writer
	topic  string
}

// NewKafkaPublisher creates a new Kafka publisher
func NewKafkaPublisher(cfg config.KafkaConfig) (*KafkaPublisher, error) {
	if len(cfg.Brokers) == 0 {
		return nil, fmt.Errorf("kafka brokers not configured")
	}
	if cfg.Topic == "" {
		return nil, fmt.Errorf("kafka topic not configured")
	}

	writer := &kafka.Writer{
		Addr:      kafka.TCP(cfg.Brokers...),
		Topic:     cfg.Topic,
		Balancer:  &kafka.LeastBytes{},
		BatchSize: 1,
	}

	return &KafkaPublisher{
		writer: writer,
		topic:  cfg.Topic,
	}, nil
}

// Publish sends the event to Kafka
func (p *KafkaPublisher) Publish(ctx context.Context, event Event) error {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	message := kafka.Message{
		Key:   []byte(event.ID),
		Value: eventJSON,
		Time:  event.Timestamp,
	}

	if err := p.writer.WriteMessages(ctx, message); err != nil {
		return fmt.Errorf("failed to write message to kafka: %w", err)
	}

	log.Printf("Event published to Kafka: %s (topic: %s)", event.ID, p.topic)
	return nil
}

// Close closes the Kafka writer
func (p *KafkaPublisher) Close() error {
	return p.writer.Close()
}
