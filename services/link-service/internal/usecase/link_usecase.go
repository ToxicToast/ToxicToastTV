package usecase

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/url"
	"time"

	"toxictoast/services/link-service/internal/domain"
	"toxictoast/services/link-service/internal/repository"
	"toxictoast/services/link-service/pkg/config"
)

const (
	shortCodeLength = 6
	shortCodeChars  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// LinkUseCase defines the interface for link business logic
type LinkUseCase interface {
	CreateLink(ctx context.Context, input CreateLinkInput) (*domain.Link, string, error)
	GetLink(ctx context.Context, id string) (*domain.Link, error)
	GetLinkByShortCode(ctx context.Context, shortCode string) (*domain.Link, error)
	UpdateLink(ctx context.Context, id string, input UpdateLinkInput) (*domain.Link, error)
	DeleteLink(ctx context.Context, id string) error
	ListLinks(ctx context.Context, filters repository.LinkFilters) ([]domain.Link, int64, error)
	IncrementClick(ctx context.Context, shortCode string) (int, error)
	GetLinkStats(ctx context.Context, linkID string) (*repository.LinkStats, error)
}

// CreateLinkInput defines the input for creating a new link
type CreateLinkInput struct {
	OriginalURL string
	CustomAlias *string
	Title       *string
	Description *string
	ExpiresAt   *time.Time
}

// UpdateLinkInput defines the input for updating a link
type UpdateLinkInput struct {
	OriginalURL *string
	CustomAlias *string
	Title       *string
	Description *string
	ExpiresAt   *time.Time
	IsActive    *bool
}

type linkUseCase struct {
	linkRepo repository.LinkRepository
	config   *config.Config
}

// NewLinkUseCase creates a new instance of LinkUseCase
func NewLinkUseCase(linkRepo repository.LinkRepository, cfg *config.Config) LinkUseCase {
	return &linkUseCase{
		linkRepo: linkRepo,
		config:   cfg,
	}
}

func (uc *linkUseCase) CreateLink(ctx context.Context, input CreateLinkInput) (*domain.Link, string, error) {
	// Validate URL
	if err := validateURL(input.OriginalURL); err != nil {
		return nil, "", fmt.Errorf("invalid URL: %w", err)
	}

	// Generate or validate short code
	var shortCode string
	var err error

	if input.CustomAlias != nil && *input.CustomAlias != "" {
		// Use custom alias if provided
		shortCode = *input.CustomAlias

		// Check if custom alias already exists
		exists, err := uc.linkRepo.ShortCodeExists(ctx, shortCode)
		if err != nil {
			return nil, "", fmt.Errorf("failed to check custom alias: %w", err)
		}
		if exists {
			return nil, "", fmt.Errorf("custom alias already exists")
		}
	} else {
		// Generate random short code
		shortCode, err = uc.generateUniqueShortCode(ctx)
		if err != nil {
			return nil, "", fmt.Errorf("failed to generate short code: %w", err)
		}
	}

	// Create link entity
	link := &domain.Link{
		OriginalURL: input.OriginalURL,
		ShortCode:   shortCode,
		CustomAlias: input.CustomAlias,
		Title:       input.Title,
		Description: input.Description,
		ExpiresAt:   input.ExpiresAt,
		IsActive:    true,
		ClickCount:  0,
	}

	// Save to database
	if err := uc.linkRepo.Create(ctx, link); err != nil {
		return nil, "", fmt.Errorf("failed to create link: %w", err)
	}

	// Generate full short URL
	shortURL := fmt.Sprintf("%s/%s", uc.config.BaseURL, shortCode)

	return link, shortURL, nil
}

func (uc *linkUseCase) GetLink(ctx context.Context, id string) (*domain.Link, error) {
	return uc.linkRepo.GetByID(ctx, id)
}

func (uc *linkUseCase) GetLinkByShortCode(ctx context.Context, shortCode string) (*domain.Link, error) {
	return uc.linkRepo.GetByShortCode(ctx, shortCode)
}

func (uc *linkUseCase) UpdateLink(ctx context.Context, id string, input UpdateLinkInput) (*domain.Link, error) {
	// Get existing link
	link, err := uc.linkRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if input.OriginalURL != nil {
		if err := validateURL(*input.OriginalURL); err != nil {
			return nil, fmt.Errorf("invalid URL: %w", err)
		}
		link.OriginalURL = *input.OriginalURL
	}

	if input.CustomAlias != nil {
		// Check if new custom alias already exists (excluding current link)
		exists, err := uc.linkRepo.ShortCodeExists(ctx, *input.CustomAlias)
		if err != nil {
			return nil, fmt.Errorf("failed to check custom alias: %w", err)
		}
		if exists && *input.CustomAlias != link.ShortCode {
			return nil, fmt.Errorf("custom alias already exists")
		}
		link.CustomAlias = input.CustomAlias
		link.ShortCode = *input.CustomAlias
	}

	if input.Title != nil {
		link.Title = input.Title
	}

	if input.Description != nil {
		link.Description = input.Description
	}

	if input.ExpiresAt != nil {
		link.ExpiresAt = input.ExpiresAt
	}

	if input.IsActive != nil {
		link.IsActive = *input.IsActive
	}

	// Save to database
	if err := uc.linkRepo.Update(ctx, link); err != nil {
		return nil, fmt.Errorf("failed to update link: %w", err)
	}

	return link, nil
}

func (uc *linkUseCase) DeleteLink(ctx context.Context, id string) error {
	// Check if link exists
	_, err := uc.linkRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete from database
	if err := uc.linkRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete link: %w", err)
	}

	return nil
}

func (uc *linkUseCase) ListLinks(ctx context.Context, filters repository.LinkFilters) ([]domain.Link, int64, error) {
	return uc.linkRepo.List(ctx, filters)
}

func (uc *linkUseCase) IncrementClick(ctx context.Context, shortCode string) (int, error) {
	// Get link by short code
	link, err := uc.linkRepo.GetByShortCode(ctx, shortCode)
	if err != nil {
		return 0, err
	}

	// Check if link is available
	if !link.IsAvailable() {
		return 0, fmt.Errorf("link is not available")
	}

	// Increment click count
	if err := uc.linkRepo.IncrementClicks(ctx, link.ID); err != nil {
		return 0, fmt.Errorf("failed to increment clicks: %w", err)
	}

	return link.ClickCount + 1, nil
}

func (uc *linkUseCase) GetLinkStats(ctx context.Context, linkID string) (*repository.LinkStats, error) {
	return uc.linkRepo.GetStats(ctx, linkID)
}

// Helper methods

func (uc *linkUseCase) generateUniqueShortCode(ctx context.Context) (string, error) {
	maxAttempts := 10
	for i := 0; i < maxAttempts; i++ {
		shortCode, err := generateRandomShortCode(shortCodeLength)
		if err != nil {
			return "", err
		}

		// Check if short code exists
		exists, err := uc.linkRepo.ShortCodeExists(ctx, shortCode)
		if err != nil {
			return "", err
		}

		if !exists {
			return shortCode, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique short code after %d attempts", maxAttempts)
}

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
