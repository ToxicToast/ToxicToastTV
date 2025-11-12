package consumer

import (
	"context"
	"log"
	"time"

	"github.com/IBM/sarama"

	"toxictoast/services/sse-service/internal/broker"
	"toxictoast/services/sse-service/internal/domain"
	"toxictoast/services/sse-service/pkg/config"
)

// KafkaConsumer consumes events from Kafka and forwards them to the SSE broker
type KafkaConsumer struct {
	config       *config.KafkaConfig
	consumerGroup sarama.ConsumerGroup
	broker       *broker.Broker
	stopChan     chan struct{}
}

// NewKafkaConsumer creates a new Kafka consumer
func NewKafkaConsumer(cfg *config.KafkaConfig, b *broker.Broker) (*KafkaConsumer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_6_0_0
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	if cfg.AutoOffsetReset == "earliest" {
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	}

	consumerGroup, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID, config)
	if err != nil {
		return nil, err
	}

	return &KafkaConsumer{
		config:       cfg,
		consumerGroup: consumerGroup,
		broker:       b,
		stopChan:     make(chan struct{}),
	}, nil
}

// Start starts consuming Kafka messages
func (kc *KafkaConsumer) Start(ctx context.Context) {
	log.Printf("🎧 Kafka consumer started")
	log.Printf("   Brokers: %v", kc.config.Brokers)
	log.Printf("   Topics: %v", kc.config.Topics)
	log.Printf("   Group ID: %s", kc.config.GroupID)

	handler := &consumerGroupHandler{
		broker: kc.broker,
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("🛑 Kafka consumer context cancelled")
				return
			case <-kc.stopChan:
				log.Println("🛑 Kafka consumer stopped")
				return
			default:
				// Consume messages
				if err := kc.consumerGroup.Consume(ctx, kc.config.Topics, handler); err != nil {
					log.Printf("❌ Error from consumer: %v", err)
					time.Sleep(5 * time.Second) // Wait before retry
				}
			}
		}
	}()
}

// Stop stops the Kafka consumer
func (kc *KafkaConsumer) Stop() error {
	log.Println("🛑 Stopping Kafka consumer...")

	// Close consumer group first (this will trigger context cancellation)
	if err := kc.consumerGroup.Close(); err != nil {
		log.Printf("⚠️  Error closing consumer group: %v", err)
	}

	// Signal stop and close channel
	select {
	case <-kc.stopChan:
		// Already closed
	default:
		close(kc.stopChan)
	}

	log.Println("✅ Kafka consumer stopped")
	return nil
}

// consumerGroupHandler implements sarama.ConsumerGroupHandler
type consumerGroupHandler struct {
	broker *broker.Broker
}

func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	log.Println("✅ Kafka consumer group session setup")
	return nil
}

func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	log.Println("🧹 Kafka consumer group session cleanup")
	return nil
}

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			// Parse event from Kafka message
			event, err := domain.ParseEvent(message.Value, message.Topic)
			if err != nil {
				log.Printf("⚠️  Failed to parse event from topic %s: %v", message.Topic, err)
				session.MarkMessage(message, "")
				continue
			}

			// Publish event to SSE broker
			h.broker.PublishEvent(event)

			// Mark message as processed
			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}
