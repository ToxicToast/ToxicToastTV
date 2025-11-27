package command

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	"github.com/toxictoast/toxictoastgo/shared/kafka"

	"toxictoast/services/link-service/internal/domain"
	"toxictoast/services/link-service/internal/repository"
	"toxictoast/services/link-service/pkg/config"
)

const (
	shortCodeLength = 6
	shortCodeChars  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// ============================================================================
// Commands
// ============================================================================

// CreateLinkCommand creates a new link with optional custom alias
type CreateLinkCommand struct {
	cqrs.BaseCommand
	OriginalURL string
	CustomAlias *string
	Title       *string
	Description *string
	ExpiresAt   *time.Time
	ShortURL    string // Will be set by handler
}

func (c *CreateLinkCommand) CommandName() string {
	return "create_link"
}

func (c *CreateLinkCommand) Validate() error {
	if c.OriginalURL == "" {
		return fmt.Errorf("original_url is required")
	}
	return validateURL(c.OriginalURL)
}

// UpdateLinkCommand updates an existing link
type UpdateLinkCommand struct {
	cqrs.BaseCommand
	OriginalURL *string
	CustomAlias *string
	Title       *string
	Description *string
	ExpiresAt   *time.Time
	IsActive    *bool
}

func (c *UpdateLinkCommand) CommandName() string {
	return "update_link"
}

func (c *UpdateLinkCommand) Validate() error {
	if c.AggregateID == "" {
		return fmt.Errorf("link_id is required")
	}
	if c.OriginalURL != nil {
		return validateURL(*c.OriginalURL)
	}
	return nil
}

// DeleteLinkCommand deletes a link
type DeleteLinkCommand struct {
	cqrs.BaseCommand
}

func (c *DeleteLinkCommand) CommandName() string {
	return "delete_link"
}

func (c *DeleteLinkCommand) Validate() error {
	if c.AggregateID == "" {
		return fmt.Errorf("link_id is required")
	}
	return nil
}

// IncrementClickCommand increments click count for a link
type IncrementClickCommand struct {
	cqrs.BaseCommand
	ShortCode string
	NewCount  int // Will be set by handler
}

func (c *IncrementClickCommand) CommandName() string {
	return "increment_click"
}

func (c *IncrementClickCommand) Validate() error {
	if c.ShortCode == "" {
		return fmt.Errorf("short_code is required")
	}
	return nil
}

// DeactivateExpiredLinkCommand deactivates an expired link (used by scheduler)
type DeactivateExpiredLinkCommand struct {
	cqrs.BaseCommand
	Link *domain.Link // The link to deactivate
}

func (c *DeactivateExpiredLinkCommand) CommandName() string {
	return "deactivate_expired_link"
}

func (c *DeactivateExpiredLinkCommand) Validate() error {
	if c.Link == nil {
		return fmt.Errorf("link is required")
	}
	return nil
}

// ============================================================================
// Command Handlers
// ============================================================================

// CreateLinkHandler handles link creation
type CreateLinkHandler struct {
	linkRepo      repository.LinkRepository
	kafkaProducer *kafka.Producer
	config        *config.Config
}

func NewCreateLinkHandler(linkRepo repository.LinkRepository, kafkaProducer *kafka.Producer, cfg *config.Config) *CreateLinkHandler {
	return &CreateLinkHandler{
		linkRepo:      linkRepo,
		kafkaProducer: kafkaProducer,
		config:        cfg,
	}
}

func (h *CreateLinkHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	createCmd := cmd.(*CreateLinkCommand)

	// Generate or validate short code
	var shortCode string
	var err error

	if createCmd.CustomAlias != nil && *createCmd.CustomAlias != "" {
		// Use custom alias if provided
		shortCode = *createCmd.CustomAlias

		// Check if custom alias already exists
		exists, err := h.linkRepo.ShortCodeExists(ctx, shortCode)
		if err != nil {
			return fmt.Errorf("failed to check custom alias: %w", err)
		}
		if exists {
			return fmt.Errorf("custom alias already exists")
		}
	} else {
		// Generate random short code
		shortCode, err = h.generateUniqueShortCode(ctx)
		if err != nil {
			return fmt.Errorf("failed to generate short code: %w", err)
		}
	}

	// Generate UUID for link
	linkID := uuid.New().String()
	createCmd.AggregateID = linkID

	// Create link entity
	link := &domain.Link{
		ID:          linkID,
		OriginalURL: createCmd.OriginalURL,
		ShortCode:   shortCode,
		CustomAlias: createCmd.CustomAlias,
		Title:       createCmd.Title,
		Description: createCmd.Description,
		ExpiresAt:   createCmd.ExpiresAt,
		IsActive:    true,
		ClickCount:  0,
	}

	// Save to database
	if err := h.linkRepo.Create(ctx, link); err != nil {
		return fmt.Errorf("failed to create link: %w", err)
	}

	// Generate full short URL
	createCmd.ShortURL = fmt.Sprintf("%s/%s", h.config.BaseURL, shortCode)

	// Publish Kafka event
	if h.kafkaProducer != nil {
		event := kafka.LinkCreatedEvent{
			LinkID:      link.ID,
			OriginalURL: link.OriginalURL,
			ShortCode:   link.ShortCode,
			CustomAlias: link.CustomAlias,
			Title:       link.Title,
			Description: link.Description,
			ExpiresAt:   link.ExpiresAt,
			IsActive:    link.IsActive,
			CreatedAt:   link.CreatedAt,
		}
		topic := "link.created"
		if err := h.kafkaProducer.PublishLinkCreated(topic, event); err != nil {
			log.Printf("Warning: Failed to publish link created event: %v", err)
		}
	}

	return nil
}

func (h *CreateLinkHandler) generateUniqueShortCode(ctx context.Context) (string, error) {
	maxAttempts := 10
	for i := 0; i < maxAttempts; i++ {
		shortCode, err := generateRandomShortCode(shortCodeLength)
		if err != nil {
			return "", err
		}

		// Check if short code exists
		exists, err := h.linkRepo.ShortCodeExists(ctx, shortCode)
		if err != nil {
			return "", err
		}

		if !exists {
			return shortCode, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique short code after %d attempts", maxAttempts)
}

// UpdateLinkHandler handles link updates
type UpdateLinkHandler struct {
	linkRepo      repository.LinkRepository
	kafkaProducer *kafka.Producer
}

func NewUpdateLinkHandler(linkRepo repository.LinkRepository, kafkaProducer *kafka.Producer) *UpdateLinkHandler {
	return &UpdateLinkHandler{
		linkRepo:      linkRepo,
		kafkaProducer: kafkaProducer,
	}
}

func (h *UpdateLinkHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	updateCmd := cmd.(*UpdateLinkCommand)

	// Get existing link
	link, err := h.linkRepo.GetByID(ctx, updateCmd.AggregateID)
	if err != nil {
		return err
	}

	// Track activation status change
	oldIsActive := link.IsActive
	activationChanged := false

	// Update fields if provided
	if updateCmd.OriginalURL != nil {
		link.OriginalURL = *updateCmd.OriginalURL
	}

	if updateCmd.CustomAlias != nil {
		// Check if new custom alias already exists (excluding current link)
		exists, err := h.linkRepo.ShortCodeExists(ctx, *updateCmd.CustomAlias)
		if err != nil {
			return fmt.Errorf("failed to check custom alias: %w", err)
		}
		if exists && *updateCmd.CustomAlias != link.ShortCode {
			return fmt.Errorf("custom alias already exists")
		}
		link.CustomAlias = updateCmd.CustomAlias
		link.ShortCode = *updateCmd.CustomAlias
	}

	if updateCmd.Title != nil {
		link.Title = updateCmd.Title
	}

	if updateCmd.Description != nil {
		link.Description = updateCmd.Description
	}

	if updateCmd.ExpiresAt != nil {
		link.ExpiresAt = updateCmd.ExpiresAt
	}

	if updateCmd.IsActive != nil {
		if oldIsActive != *updateCmd.IsActive {
			activationChanged = true
		}
		link.IsActive = *updateCmd.IsActive
	}

	// Save to database
	if err := h.linkRepo.Update(ctx, link); err != nil {
		return fmt.Errorf("failed to update link: %w", err)
	}

	// Publish Kafka events
	if h.kafkaProducer != nil {
		// Publish link updated event
		event := kafka.LinkUpdatedEvent{
			LinkID:      link.ID,
			OriginalURL: link.OriginalURL,
			ShortCode:   link.ShortCode,
			CustomAlias: link.CustomAlias,
			Title:       link.Title,
			Description: link.Description,
			ExpiresAt:   link.ExpiresAt,
			IsActive:    link.IsActive,
			UpdatedAt:   link.UpdatedAt,
		}
		topic := "link.updated"
		if err := h.kafkaProducer.PublishLinkUpdated(topic, event); err != nil {
			log.Printf("Warning: Failed to publish link updated event: %v", err)
		}

		// Publish activation/deactivation events if status changed
		if activationChanged {
			if link.IsActive {
				activatedEvent := kafka.LinkActivatedEvent{
					LinkID:      link.ID,
					ShortCode:   link.ShortCode,
					ActivatedAt: time.Now(),
				}
				activatedTopic := "link.activated"
				if err := h.kafkaProducer.PublishLinkActivated(activatedTopic, activatedEvent); err != nil {
					log.Printf("Warning: Failed to publish link activated event: %v", err)
				}
			} else {
				deactivatedEvent := kafka.LinkDeactivatedEvent{
					LinkID:        link.ID,
					ShortCode:     link.ShortCode,
					DeactivatedAt: time.Now(),
				}
				deactivatedTopic := "link.deactivated"
				if err := h.kafkaProducer.PublishLinkDeactivated(deactivatedTopic, deactivatedEvent); err != nil {
					log.Printf("Warning: Failed to publish link deactivated event: %v", err)
				}
			}
		}
	}

	return nil
}

// DeleteLinkHandler handles link deletion
type DeleteLinkHandler struct {
	linkRepo      repository.LinkRepository
	kafkaProducer *kafka.Producer
}

func NewDeleteLinkHandler(linkRepo repository.LinkRepository, kafkaProducer *kafka.Producer) *DeleteLinkHandler {
	return &DeleteLinkHandler{
		linkRepo:      linkRepo,
		kafkaProducer: kafkaProducer,
	}
}

func (h *DeleteLinkHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	deleteCmd := cmd.(*DeleteLinkCommand)

	// Check if link exists
	link, err := h.linkRepo.GetByID(ctx, deleteCmd.AggregateID)
	if err != nil {
		return err
	}

	// Delete from database
	if err := h.linkRepo.Delete(ctx, deleteCmd.AggregateID); err != nil {
		return fmt.Errorf("failed to delete link: %w", err)
	}

	// Publish Kafka event
	if h.kafkaProducer != nil {
		event := kafka.LinkDeletedEvent{
			LinkID:    link.ID,
			ShortCode: link.ShortCode,
			DeletedAt: time.Now(),
		}
		topic := "link.deleted"
		if err := h.kafkaProducer.PublishLinkDeleted(topic, event); err != nil {
			log.Printf("Warning: Failed to publish link deleted event: %v", err)
		}
	}

	return nil
}

// IncrementClickHandler increments link click count
type IncrementClickHandler struct {
	linkRepo repository.LinkRepository
}

func NewIncrementClickHandler(linkRepo repository.LinkRepository) *IncrementClickHandler {
	return &IncrementClickHandler{
		linkRepo: linkRepo,
	}
}

func (h *IncrementClickHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	incrementCmd := cmd.(*IncrementClickCommand)

	// Get link by short code
	link, err := h.linkRepo.GetByShortCode(ctx, incrementCmd.ShortCode)
	if err != nil {
		return err
	}

	// Store link ID
	incrementCmd.AggregateID = link.ID

	// Check if link is available
	if !link.IsAvailable() {
		return fmt.Errorf("link is not available")
	}

	// Increment click count
	if err := h.linkRepo.IncrementClicks(ctx, link.ID); err != nil {
		return fmt.Errorf("failed to increment clicks: %w", err)
	}

	// Set new count in command
	incrementCmd.NewCount = link.ClickCount + 1

	return nil
}

// DeactivateExpiredLinkHandler deactivates an expired link
type DeactivateExpiredLinkHandler struct {
	linkRepo      repository.LinkRepository
	kafkaProducer *kafka.Producer
}

func NewDeactivateExpiredLinkHandler(linkRepo repository.LinkRepository, kafkaProducer *kafka.Producer) *DeactivateExpiredLinkHandler {
	return &DeactivateExpiredLinkHandler{
		linkRepo:      linkRepo,
		kafkaProducer: kafkaProducer,
	}
}

func (h *DeactivateExpiredLinkHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	deactivateCmd := cmd.(*DeactivateExpiredLinkCommand)

	link := deactivateCmd.Link

	// Mark as inactive
	link.IsActive = false
	link.UpdatedAt = time.Now()

	// Save to database
	if err := h.linkRepo.Update(ctx, link); err != nil {
		return fmt.Errorf("failed to deactivate expired link: %w", err)
	}

	// Publish link expired event
	if h.kafkaProducer != nil {
		var expiresAt time.Time
		if link.ExpiresAt != nil {
			expiresAt = *link.ExpiresAt
		}

		event := kafka.LinkExpiredEvent{
			LinkID:    link.ID,
			ShortCode: link.ShortCode,
			ExpiresAt: expiresAt,
		}
		topic := "link.expired"
		if err := h.kafkaProducer.PublishLinkExpired(topic, event); err != nil {
			log.Printf("Warning: Failed to publish link expired event: %v", err)
		}
	}

	return nil
}

// ============================================================================
// Helper Functions
// ============================================================================

func generateRandomShortCode(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	for i, b := range bytes {
		bytes[i] = shortCodeChars[b%byte(len(shortCodeChars))]
	}

	return string(bytes), nil
}

func validateURL(rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("URL must use http or https scheme")
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("URL must have a valid host")
	}

	return nil
}
