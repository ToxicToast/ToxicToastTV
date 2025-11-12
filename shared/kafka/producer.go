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

// PublishCategoryCreated publishes a category created event
func (p *Producer) PublishCategoryCreated(topic string, event CategoryCreatedEvent) error {
	return p.PublishEvent(topic, event.CategoryID, event)
}

// PublishCategoryUpdated publishes a category updated event
func (p *Producer) PublishCategoryUpdated(topic string, event CategoryUpdatedEvent) error {
	return p.PublishEvent(topic, event.CategoryID, event)
}

// PublishCategoryDeleted publishes a category deleted event
func (p *Producer) PublishCategoryDeleted(topic string, event CategoryDeletedEvent) error {
	return p.PublishEvent(topic, event.CategoryID, event)
}

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

// PublishTagCreated publishes a tag created event
func (p *Producer) PublishTagCreated(topic string, event TagCreatedEvent) error {
	return p.PublishEvent(topic, event.TagID, event)
}

// PublishTagUpdated publishes a tag updated event
func (p *Producer) PublishTagUpdated(topic string, event TagUpdatedEvent) error {
	return p.PublishEvent(topic, event.TagID, event)
}

// PublishTagDeleted publishes a tag deleted event
func (p *Producer) PublishTagDeleted(topic string, event TagDeletedEvent) error {
	return p.PublishEvent(topic, event.TagID, event)
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

// PublishCommentApproved publishes a comment approved event
func (p *Producer) PublishCommentApproved(topic string, event CommentApprovedEvent) error {
	return p.PublishEvent(topic, event.CommentID, event)
}

// PublishCommentRejected publishes a comment rejected event
func (p *Producer) PublishCommentRejected(topic string, event CommentRejectedEvent) error {
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

// PublishMediaThumbnailGenerated publishes a media thumbnail generated event
func (p *Producer) PublishMediaThumbnailGenerated(topic string, event MediaThumbnailGeneratedEvent) error {
	return p.PublishEvent(topic, event.MediaID, event)
}

// PublishLinkCreated publishes a link created event
func (p *Producer) PublishLinkCreated(topic string, event LinkCreatedEvent) error {
	return p.PublishEvent(topic, event.LinkID, event)
}

// PublishLinkUpdated publishes a link updated event
func (p *Producer) PublishLinkUpdated(topic string, event LinkUpdatedEvent) error {
	return p.PublishEvent(topic, event.LinkID, event)
}

// PublishLinkDeleted publishes a link deleted event
func (p *Producer) PublishLinkDeleted(topic string, event LinkDeletedEvent) error {
	return p.PublishEvent(topic, event.LinkID, event)
}

// PublishLinkActivated publishes a link activated event
func (p *Producer) PublishLinkActivated(topic string, event LinkActivatedEvent) error {
	return p.PublishEvent(topic, event.LinkID, event)
}

// PublishLinkDeactivated publishes a link deactivated event
func (p *Producer) PublishLinkDeactivated(topic string, event LinkDeactivatedEvent) error {
	return p.PublishEvent(topic, event.LinkID, event)
}

// PublishLinkExpired publishes a link expired event
func (p *Producer) PublishLinkExpired(topic string, event LinkExpiredEvent) error {
	return p.PublishEvent(topic, event.LinkID, event)
}

// PublishLinkClicked publishes a link clicked event
func (p *Producer) PublishLinkClicked(topic string, event LinkClickedEvent) error {
	return p.PublishEvent(topic, event.ClickID, event)
}

// PublishLinkClickFraudDetected publishes a link click fraud detected event
func (p *Producer) PublishLinkClickFraudDetected(topic string, event LinkClickFraudDetectedEvent) error {
	return p.PublishEvent(topic, event.ClickID, event)
}

// Foodfolio Category Event Publishers
func (p *Producer) PublishFoodfolioCategoryCreated(topic string, event FoodfolioCategoryCreatedEvent) error {
	return p.PublishEvent(topic, event.CategoryID, event)
}

func (p *Producer) PublishFoodfolioCategoryUpdated(topic string, event FoodfolioCategoryUpdatedEvent) error {
	return p.PublishEvent(topic, event.CategoryID, event)
}

func (p *Producer) PublishFoodfolioCategoryDeleted(topic string, event FoodfolioCategoryDeletedEvent) error {
	return p.PublishEvent(topic, event.CategoryID, event)
}

// Foodfolio Company Event Publishers
func (p *Producer) PublishFoodfolioCompanyCreated(topic string, event FoodfolioCompanyCreatedEvent) error {
	return p.PublishEvent(topic, event.CompanyID, event)
}

func (p *Producer) PublishFoodfolioCompanyUpdated(topic string, event FoodfolioCompanyUpdatedEvent) error {
	return p.PublishEvent(topic, event.CompanyID, event)
}

func (p *Producer) PublishFoodfolioCompanyDeleted(topic string, event FoodfolioCompanyDeletedEvent) error {
	return p.PublishEvent(topic, event.CompanyID, event)
}

// Foodfolio Item Event Publishers
func (p *Producer) PublishFoodfolioItemCreated(topic string, event FoodfolioItemCreatedEvent) error {
	return p.PublishEvent(topic, event.ItemID, event)
}

func (p *Producer) PublishFoodfolioItemUpdated(topic string, event FoodfolioItemUpdatedEvent) error {
	return p.PublishEvent(topic, event.ItemID, event)
}

func (p *Producer) PublishFoodfolioItemDeleted(topic string, event FoodfolioItemDeletedEvent) error {
	return p.PublishEvent(topic, event.ItemID, event)
}

// Foodfolio Item Variant Event Publishers
func (p *Producer) PublishFoodfolioVariantCreated(topic string, event FoodfolioVariantCreatedEvent) error {
	return p.PublishEvent(topic, event.VariantID, event)
}

func (p *Producer) PublishFoodfolioVariantUpdated(topic string, event FoodfolioVariantUpdatedEvent) error {
	return p.PublishEvent(topic, event.VariantID, event)
}

func (p *Producer) PublishFoodfolioVariantDeleted(topic string, event FoodfolioVariantDeletedEvent) error {
	return p.PublishEvent(topic, event.VariantID, event)
}

func (p *Producer) PublishFoodfolioVariantStockLow(topic string, event FoodfolioVariantStockLowEvent) error {
	return p.PublishEvent(topic, event.VariantID, event)
}

func (p *Producer) PublishFoodfolioVariantStockEmpty(topic string, event FoodfolioVariantStockEmptyEvent) error {
	return p.PublishEvent(topic, event.VariantID, event)
}

// Foodfolio Item Detail Event Publishers
func (p *Producer) PublishFoodfolioDetailCreated(topic string, event FoodfolioDetailCreatedEvent) error {
	return p.PublishEvent(topic, event.DetailID, event)
}

func (p *Producer) PublishFoodfolioDetailOpened(topic string, event FoodfolioDetailOpenedEvent) error {
	return p.PublishEvent(topic, event.DetailID, event)
}

func (p *Producer) PublishFoodfolioDetailExpired(topic string, event FoodfolioDetailExpiredEvent) error {
	return p.PublishEvent(topic, event.DetailID, event)
}

func (p *Producer) PublishFoodfolioDetailExpiringSoon(topic string, event FoodfolioDetailExpiringSoonEvent) error {
	return p.PublishEvent(topic, event.DetailID, event)
}

func (p *Producer) PublishFoodfolioDetailConsumed(topic string, event FoodfolioDetailConsumedEvent) error {
	return p.PublishEvent(topic, event.DetailID, event)
}

func (p *Producer) PublishFoodfolioDetailMoved(topic string, event FoodfolioDetailMovedEvent) error {
	return p.PublishEvent(topic, event.DetailID, event)
}

func (p *Producer) PublishFoodfolioDetailFrozen(topic string, event FoodfolioDetailFrozenEvent) error {
	return p.PublishEvent(topic, event.DetailID, event)
}

func (p *Producer) PublishFoodfolioDetailThawed(topic string, event FoodfolioDetailThawedEvent) error {
	return p.PublishEvent(topic, event.DetailID, event)
}

// Foodfolio Location Event Publishers
func (p *Producer) PublishFoodfolioLocationCreated(topic string, event FoodfolioLocationCreatedEvent) error {
	return p.PublishEvent(topic, event.LocationID, event)
}

func (p *Producer) PublishFoodfolioLocationUpdated(topic string, event FoodfolioLocationUpdatedEvent) error {
	return p.PublishEvent(topic, event.LocationID, event)
}

func (p *Producer) PublishFoodfolioLocationDeleted(topic string, event FoodfolioLocationDeletedEvent) error {
	return p.PublishEvent(topic, event.LocationID, event)
}

// Foodfolio Warehouse Event Publishers
func (p *Producer) PublishFoodfolioWarehouseCreated(topic string, event FoodfolioWarehouseCreatedEvent) error {
	return p.PublishEvent(topic, event.WarehouseID, event)
}

func (p *Producer) PublishFoodfolioWarehouseUpdated(topic string, event FoodfolioWarehouseUpdatedEvent) error {
	return p.PublishEvent(topic, event.WarehouseID, event)
}

func (p *Producer) PublishFoodfolioWarehouseDeleted(topic string, event FoodfolioWarehouseDeletedEvent) error {
	return p.PublishEvent(topic, event.WarehouseID, event)
}

// Foodfolio Receipt Event Publishers
func (p *Producer) PublishFoodfolioReceiptCreated(topic string, event FoodfolioReceiptCreatedEvent) error {
	return p.PublishEvent(topic, event.ReceiptID, event)
}

func (p *Producer) PublishFoodfolioReceiptScanned(topic string, event FoodfolioReceiptScannedEvent) error {
	return p.PublishEvent(topic, event.ReceiptID, event)
}

func (p *Producer) PublishFoodfolioReceiptDeleted(topic string, event FoodfolioReceiptDeletedEvent) error {
	return p.PublishEvent(topic, event.ReceiptID, event)
}

// Foodfolio Shopping List Event Publishers
func (p *Producer) PublishFoodfolioShoppinglistCreated(topic string, event FoodfolioShoppinglistCreatedEvent) error {
	return p.PublishEvent(topic, event.ShoppinglistID, event)
}

func (p *Producer) PublishFoodfolioShoppinglistUpdated(topic string, event FoodfolioShoppinglistUpdatedEvent) error {
	return p.PublishEvent(topic, event.ShoppinglistID, event)
}

func (p *Producer) PublishFoodfolioShoppinglistDeleted(topic string, event FoodfolioShoppinglistDeletedEvent) error {
	return p.PublishEvent(topic, event.ShoppinglistID, event)
}

func (p *Producer) PublishFoodfolioShoppinglistItemAdded(topic string, event FoodfolioShoppinglistItemAddedEvent) error {
	return p.PublishEvent(topic, event.ItemID, event)
}

func (p *Producer) PublishFoodfolioShoppinglistItemRemoved(topic string, event FoodfolioShoppinglistItemRemovedEvent) error {
	return p.PublishEvent(topic, event.ItemID, event)
}

func (p *Producer) PublishFoodfolioShoppinglistItemPurchased(topic string, event FoodfolioShoppinglistItemPurchasedEvent) error {
	return p.PublishEvent(topic, event.ItemID, event)
}
