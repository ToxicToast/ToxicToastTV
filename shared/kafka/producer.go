package kafka

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"
)

type Producer struct {
	producer sarama.SyncProducer
	brokers  []string
}

func NewProducer(brokers []string) (*Producer, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	saramaConfig.Producer.Retry.Max = 5
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.Compression = sarama.CompressionSnappy
	saramaConfig.Producer.Timeout = 10 * time.Second

	producer, err := sarama.NewSyncProducer(brokers, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	log.Printf("Kafka producer connected to brokers: %v", brokers)

	return &Producer{
		producer: producer,
		brokers:  brokers,
	}, nil
}

func (p *Producer) Close() error {
	if p.producer != nil {
		return p.producer.Close()
	}
	return nil
}

// PublishEvent publishes an event to a Kafka topic
func (p *Producer) PublishEvent(topic string, key string, event interface{}) error {
	eventData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(eventData),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("event_type"),
				Value: []byte(topic),
			},
			{
				Key:   []byte("timestamp"),
				Value: []byte(time.Now().Format(time.RFC3339)),
			},
		},
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message to Kafka: %w", err)
	}

	log.Printf("Event published to topic %s (partition: %d, offset: %d)", topic, partition, offset)
	return nil
}

// Helper methods for common event publishing patterns
// Services should pass their configured topic names

// PublishPostCreated publishes a post created event
func (p *Producer) PublishPostCreated(topic string, event PostCreatedEvent) error {
	return p.PublishEvent(topic, event.PostID, event)
}

// PublishPostUpdated publishes a post updated event
func (p *Producer) PublishPostUpdated(topic string, event PostUpdatedEvent) error {
	return p.PublishEvent(topic, event.PostID, event)
}

// PublishPostPublished publishes a post published event
func (p *Producer) PublishPostPublished(topic string, event PostPublishedEvent) error {
	return p.PublishEvent(topic, event.PostID, event)
}

// PublishPostDeleted publishes a post deleted event
func (p *Producer) PublishPostDeleted(topic string, event PostDeletedEvent) error {
	return p.PublishEvent(topic, event.PostID, event)
}

// PublishCommentCreated publishes a comment created event
func (p *Producer) PublishCommentCreated(topic string, event CommentCreatedEvent) error {
	return p.PublishEvent(topic, event.CommentID, event)
}

// PublishCommentModerated publishes a comment moderated event
func (p *Producer) PublishCommentModerated(topic string, event CommentModeratedEvent) error {
	return p.PublishEvent(topic, event.CommentID, event)
}

// PublishCommentDeleted publishes a comment deleted event
func (p *Producer) PublishCommentDeleted(topic string, event CommentDeletedEvent) error {
	return p.PublishEvent(topic, event.CommentID, event)
}

// PublishMediaUploaded publishes a media uploaded event
func (p *Producer) PublishMediaUploaded(topic string, event MediaUploadedEvent) error {
	return p.PublishEvent(topic, event.MediaID, event)
}

// PublishMediaDeleted publishes a media deleted event
func (p *Producer) PublishMediaDeleted(topic string, event MediaDeletedEvent) error {
	return p.PublishEvent(topic, event.MediaID, event)
}
