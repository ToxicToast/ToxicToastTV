package command

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	"github.com/toxictoast/toxictoastgo/shared/kafka"

	"toxictoast/services/link-service/internal/domain"
	"toxictoast/services/link-service/internal/repository"
)

// ============================================================================
// Commands
// ============================================================================

// RecordClickCommand records a click on a link
type RecordClickCommand struct {
	cqrs.BaseCommand
	LinkID     string
	IPAddress  string
	UserAgent  string
	Referer    *string
	Country    *string
	City       *string
	DeviceType *string
}

func (c *RecordClickCommand) CommandName() string {
	return "record_click"
}

func (c *RecordClickCommand) Validate() error {
	if c.LinkID == "" {
		return fmt.Errorf("link_id is required")
	}
	if c.IPAddress == "" {
		return fmt.Errorf("ip_address is required")
	}
	if c.UserAgent == "" {
		return fmt.Errorf("user_agent is required")
	}
	return nil
}

// ============================================================================
// Command Handlers
// ============================================================================

// RecordClickHandler handles click recording
type RecordClickHandler struct {
	clickRepo     repository.ClickRepository
	linkRepo      repository.LinkRepository
	kafkaProducer *kafka.Producer
}

func NewRecordClickHandler(clickRepo repository.ClickRepository, linkRepo repository.LinkRepository, kafkaProducer *kafka.Producer) *RecordClickHandler {
	return &RecordClickHandler{
		clickRepo:     clickRepo,
		linkRepo:      linkRepo,
		kafkaProducer: kafkaProducer,
	}
}

func (h *RecordClickHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	recordCmd := cmd.(*RecordClickCommand)

	// Verify link exists
	link, err := h.linkRepo.GetByID(ctx, recordCmd.LinkID)
	if err != nil {
		return fmt.Errorf("link not found: %w", err)
	}

	// Check if link is available
	if !link.IsAvailable() {
		return fmt.Errorf("link is not available")
	}

	// Generate UUID for click
	clickID := uuid.New().String()
	recordCmd.AggregateID = clickID

	// Create click entity
	click := &domain.Click{
		ID:         clickID,
		LinkID:     recordCmd.LinkID,
		IPAddress:  recordCmd.IPAddress,
		UserAgent:  recordCmd.UserAgent,
		Referer:    recordCmd.Referer,
		Country:    recordCmd.Country,
		City:       recordCmd.City,
		DeviceType: recordCmd.DeviceType,
		ClickedAt:  time.Now(),
	}

	// Save to database
	if err := h.clickRepo.Create(ctx, click); err != nil {
		return fmt.Errorf("failed to record click: %w", err)
	}

	// Publish Kafka event
	if h.kafkaProducer != nil {
		event := kafka.LinkClickedEvent{
			ClickID:    click.ID,
			LinkID:     click.LinkID,
			ShortCode:  link.ShortCode,
			IPAddress:  click.IPAddress,
			UserAgent:  click.UserAgent,
			Referer:    click.Referer,
			Country:    click.Country,
			City:       click.City,
			DeviceType: click.DeviceType,
			ClickedAt:  click.ClickedAt,
		}
		topic := "link.clicked"
		if err := h.kafkaProducer.PublishLinkClicked(topic, event); err != nil {
			log.Printf("Warning: Failed to publish link clicked event: %v", err)
		}
	}

	return nil
}
