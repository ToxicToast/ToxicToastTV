package usecase

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/toxictoast/toxictoastgo/shared/kafka"

	"toxictoast/services/blog-service/internal/domain"
	"toxictoast/services/blog-service/internal/repository"
	"toxictoast/services/blog-service/pkg/config"
	"toxictoast/services/blog-service/pkg/image"
	"toxictoast/services/blog-service/pkg/storage"
)

type MediaUseCase interface {
	UploadMedia(ctx context.Context, input UploadMediaInput) (*domain.Media, error)
	GetMedia(ctx context.Context, id string) (*domain.Media, error)
	DeleteMedia(ctx context.Context, id string) error
	ListMedia(ctx context.Context, filters repository.MediaFilters) ([]domain.Media, int64, error)
}

type UploadMediaInput struct {
	Data             []byte
	OriginalFilename string
	MimeType         string
	UploadedBy       string
}

type mediaUseCase struct {
	repo          repository.MediaRepository
	storage       *storage.Storage
	kafkaProducer *kafka.Producer
	config        *config.Config
	baseURL       string // Base URL for constructing media URLs
}

func NewMediaUseCase(
	repo repository.MediaRepository,
	kafkaProducer *kafka.Producer,
	cfg *config.Config,
) (MediaUseCase, error) {
	// Initialize storage
	storageInstance, err := storage.NewStorage(
		cfg.Media.StoragePath,
		cfg.Media.AllowedTypes,
		cfg.Media.MaxSize,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Construct base URL for media files
	// In production, this would be your CDN or public URL
	baseURL := fmt.Sprintf("http://localhost:%s/media", cfg.Port)

	return &mediaUseCase{
		repo:          repo,
		storage:       storageInstance,
		kafkaProducer: kafkaProducer,
		config:        cfg,
		baseURL:       baseURL,
	}, nil
}

func (uc *mediaUseCase) UploadMedia(ctx context.Context, input UploadMediaInput) (*domain.Media, error) {
	// Process image if needed (resize large images)
	processedData := input.Data
	if uc.isImageMimeType(input.MimeType) && uc.config.Media.AutoResizeLargeImage {
		resizedData, err := uc.resizeLargeImage(input.Data)
		if err == nil && resizedData != nil {
			processedData = resizedData
		}
	}

	// Save file to storage
	filename, err := uc.storage.SaveFile(processedData, input.OriginalFilename, input.MimeType)
	if err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// Create media entity
	media := &domain.Media{
		Filename:         filepath.Base(filename),
		OriginalFilename: input.OriginalFilename,
		MimeType:         input.MimeType,
		Size:             int64(len(processedData)),
		Path:             filename,
		URL:              uc.constructURL(filename),
		UploadedBy:       input.UploadedBy,
	}

	// Extract image metadata and generate thumbnails if it's an image
	if uc.isImageMimeType(input.MimeType) {
		info, err := image.GetImageInfo(processedData)
		if err == nil {
			media.Width = info.Width
			media.Height = info.Height

			// Generate thumbnails if enabled
			if uc.config.Media.GenerateThumbnails {
				thumbnailPath, err := uc.generateThumbnail(filename)
				if err == nil && thumbnailPath != "" {
					thumbnailURL := uc.constructURL(thumbnailPath)
					media.ThumbnailURL = &thumbnailURL
				}
			}
		}
	}

	// Save to database
	if err := uc.repo.Create(ctx, media); err != nil {
		// Cleanup: delete file if database insert fails
		_ = uc.storage.DeleteFile(filename)
		return nil, fmt.Errorf("failed to create media record: %w", err)
	}

	// TODO: Publish Kafka event for media upload
	// if uc.kafkaProducer != nil { ... }

	return media, nil
}

func (uc *mediaUseCase) GetMedia(ctx context.Context, id string) (*domain.Media, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *mediaUseCase) DeleteMedia(ctx context.Context, id string) error {
	// Get media record
	media, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete from storage
	if err := uc.storage.DeleteFile(media.Path); err != nil {
		// Log error but continue with database deletion
		// The file might already be deleted or missing
	}

	// Delete thumbnail if exists
	if media.ThumbnailURL != nil {
		// Extract path from URL and delete
		// This is a simplified approach
		// In production, you'd parse the URL properly
	}

	// Delete from database
	if err := uc.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete media record: %w", err)
	}

	return nil
}

func (uc *mediaUseCase) ListMedia(ctx context.Context, filters repository.MediaFilters) ([]domain.Media, int64, error) {
	return uc.repo.List(ctx, filters)
}

// Helper methods

func (uc *mediaUseCase) constructURL(filename string) string {
	// Convert Windows path separators to URL path separators
	urlPath := strings.ReplaceAll(filename, "\\", "/")
	return fmt.Sprintf("%s/%s", uc.baseURL, urlPath)
}

func (uc *mediaUseCase) isImageMimeType(mimeType string) bool {
	imageMimeTypes := []string{
		"image/jpeg",
		"image/png",
		"image/gif",
		"image/webp",
		"image/svg+xml",
	}

	mimeType = strings.ToLower(mimeType)
	for _, t := range imageMimeTypes {
		if t == mimeType {
			return true
		}
	}

	return false
}

// generateThumbnail creates a thumbnail for an image
func (uc *mediaUseCase) generateThumbnail(originalPath string) (string, error) {
	// Get full path to original file
	fullPath := uc.storage.GetFilePath(originalPath)

	// Generate medium thumbnail (300x300) - good balance for most uses
	thumbnailPath := image.GetThumbnailPath(originalPath, image.ThumbnailMedium)
	thumbnailFullPath := uc.storage.GetFilePath(thumbnailPath)

	// Create thumbnail
	if err := image.GenerateThumbnail(fullPath, thumbnailFullPath, image.ThumbnailMedium); err != nil {
		return "", fmt.Errorf("failed to generate thumbnail: %w", err)
	}

	return thumbnailPath, nil
}

// generateMultipleThumbnails creates multiple thumbnail sizes
func (uc *mediaUseCase) generateMultipleThumbnails(originalPath string) ([]string, error) {
	fullPath := uc.storage.GetFilePath(originalPath)

	// Define sizes to generate
	sizes := []image.ThumbnailSize{
		image.ThumbnailSmall,  // 150x150
		image.ThumbnailMedium, // 300x300
		image.ThumbnailLarge,  // 600x600
	}

	var thumbnailPaths []string

	for _, size := range sizes {
		thumbnailPath := image.GetThumbnailPath(originalPath, size)
		thumbnailFullPath := uc.storage.GetFilePath(thumbnailPath)

		if err := image.GenerateThumbnail(fullPath, thumbnailFullPath, size); err != nil {
			// Log error but continue with other sizes
			continue
		}

		thumbnailPaths = append(thumbnailPaths, thumbnailPath)
	}

	if len(thumbnailPaths) == 0 {
		return nil, fmt.Errorf("failed to generate any thumbnails")
	}

	return thumbnailPaths, nil
}

// resizeLargeImage resizes an image if it exceeds max dimensions
func (uc *mediaUseCase) resizeLargeImage(data []byte) ([]byte, error) {
	// Get image info
	info, err := image.GetImageInfo(data)
	if err != nil {
		return nil, err
	}

	// Check if resize is needed
	maxWidth := uc.config.Media.MaxImageWidth
	maxHeight := uc.config.Media.MaxImageHeight

	if info.Width <= maxWidth && info.Height <= maxHeight {
		// No resize needed
		return nil, nil
	}

	// Calculate new dimensions while maintaining aspect ratio
	newWidth, newHeight := image.CalculateThumbnailDimensions(
		info.Width, info.Height,
		maxWidth, maxHeight,
	)

	// Resize image
	opts := image.ResizeOptions{
		Width:   newWidth,
		Height:  newHeight,
		Fit:     true,
		Quality: 85,
	}

	resizedData, err := image.ResizeImageFromBytes(data, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to resize image: %w", err)
	}

	return resizedData, nil
}
