package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"toxictoast/services/webhook-service/internal/domain"
	"toxictoast/services/webhook-service/internal/usecase"

	"github.com/segmentio/kafka-go"
	"github.com/toxictoast/toxictoastgo/shared/logger"
)

type KafkaConsumer struct {
	reader         *kafka.Reader
	deliveryUC     *usecase.DeliveryUseCase
	wg             sync.WaitGroup
	ctx            context.Context
	cancel         context.CancelFunc
	messageChannel chan kafka.Message
	workerCount    int
}

type Config struct {
	Brokers      []string
	Topics       []string
	GroupID      string
	WorkerCount  int
}

func NewKafkaConsumer(
	config Config,
	deliveryUC *usecase.DeliveryUseCase,
) *KafkaConsumer {
	if config.WorkerCount == 0 {
		config.WorkerCount = 5
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Create reader with GroupTopics for multiple topics
	var reader *kafka.Reader
	if len(config.Topics) > 0 {
		reader = kafka.NewReader(kafka.ReaderConfig{
			Brokers:     config.Brokers,
			GroupTopics: config.Topics, // Use GroupTopics for consumer groups
			GroupID:     config.GroupID,
			MinBytes:    10e3,
			MaxBytes:    10e6,
			StartOffset: kafka.LastOffset,
		})
	} else {
		// Fallback to single topic if no topics provided
		reader = kafka.NewReader(kafka.ReaderConfig{
			Brokers:     config.Brokers,
			Topic:       "default-topic",
			GroupID:     config.GroupID,
			MinBytes:    10e3,
			MaxBytes:    10e6,
			StartOffset: kafka.LastOffset,
		})
	}

	return &KafkaConsumer{
		reader:         reader,
		deliveryUC:     deliveryUC,
		ctx:            ctx,
		cancel:         cancel,
		messageChannel: make(chan kafka.Message, 100),
		workerCount:    config.WorkerCount,
	}
}

// Start starts the Kafka consumer
func (c *KafkaConsumer) Start() error {
	logger.Info(fmt.Sprintf("Starting Kafka consumer with %d workers", c.workerCount))

	// Start message reader
	c.wg.Add(1)
	go c.readMessages()

	// Start worker pool
	for i := 0; i < c.workerCount; i++ {
		c.wg.Add(1)
		go c.worker(i)
	}

	logger.Info("Kafka consumer started successfully")
	return nil
}

// Stop gracefully stops the Kafka consumer
func (c *KafkaConsumer) Stop() error {
	logger.Info("Stopping Kafka consumer...")
	c.cancel()
	close(c.messageChannel)
	c.wg.Wait()

	if err := c.reader.Close(); err != nil {
		logger.Error(fmt.Sprintf("Error closing Kafka reader: %v", err))
		return err
	}

	logger.Info("Kafka consumer stopped")
	return nil
}

// readMessages reads messages from Kafka and sends them to the message channel
func (c *KafkaConsumer) readMessages() {
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			msg, err := c.reader.FetchMessage(c.ctx)
			if err != nil {
				if err == context.Canceled {
					return
				}
				logger.Error(fmt.Sprintf("Error fetching message from Kafka: %v", err))
				continue
			}

			// Send to workers
			select {
			case c.messageChannel <- msg:
			case <-c.ctx.Done():
				return
			}
		}
	}
}

// worker processes messages from the message channel
func (c *KafkaConsumer) worker(id int) {
	defer c.wg.Done()
	logger.Info(fmt.Sprintf("Kafka worker %d started", id))

	for {
		select {
		case <-c.ctx.Done():
			logger.Info(fmt.Sprintf("Kafka worker %d stopped", id))
			return
		case msg, ok := <-c.messageChannel:
			if !ok {
				logger.Info(fmt.Sprintf("Kafka worker %d stopped (channel closed)", id))
				return
			}

			if err := c.processMessage(msg); err != nil {
				logger.Error(fmt.Sprintf("Worker %d failed to process message: %v", id, err))
			} else {
				// Commit message after successful processing
				if err := c.reader.CommitMessages(c.ctx, msg); err != nil {
					logger.Error(fmt.Sprintf("Failed to commit message: %v", err))
				}
			}
		}
	}
}

// processMessage processes a single Kafka message
func (c *KafkaConsumer) processMessage(msg kafka.Message) error {
	logger.Info(fmt.Sprintf("Processing message from topic %s (partition %d, offset %d)",
		msg.Topic, msg.Partition, msg.Offset))

	// Parse event from message
	var event domain.Event
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		logger.Error(fmt.Sprintf("Failed to unmarshal event: %v", err))
		// Still commit the message to avoid reprocessing invalid JSON
		return nil
	}

	// Set event type from topic if not present
	if event.Type == "" {
		event.Type = msg.Topic
	}

	// Set source if not present
	if event.Source == "" {
		event.Source = "kafka"
	}

	// Process event and queue deliveries
	if err := c.deliveryUC.ProcessEvent(c.ctx, &event); err != nil {
		return fmt.Errorf("failed to process event: %w", err)
	}

	logger.Info(fmt.Sprintf("Successfully processed event %s (type: %s)", event.ID, event.Type))
	return nil
}
