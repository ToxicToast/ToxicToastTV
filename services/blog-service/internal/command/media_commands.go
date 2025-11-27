package command

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/toxictoast/toxictoastgo/shared/cqrs"
	"github.com/toxictoast/toxictoastgo/shared/kafka"

	"toxictoast/services/blog-service/internal/domain"
	"toxictoast/services/blog-service/internal/repository"
	"toxictoast/services/blog-service/pkg/config"
	"toxictoast/services/blog-service/pkg/image"
	"toxictoast/services/blog-service/pkg/storage"
)

// ============================================================================
// Commands
// ============================================================================

// UploadMediaCommand uploads a media file
type UploadMediaCommand struct {
	cqrs.BaseCommand
	Data             []byte `json:"-"` // Binary data not serialized to JSON
	OriginalFilename string `json:"original_filename"`
	MimeType         string `json:"mime_type"`
	UploadedBy       string `json:"uploaded_by"`
}

func (c *UploadMediaCommand) CommandName() string {
	return "upload_media"
}

func (c *UploadMediaCommand) Validate() error {
	if len(c.Data) == 0 {
		return errors.New("data is required")
	}
	if c.OriginalFilename == "" {
		return errors.New("original_filename is required")
	}
	if c.MimeType == "" {
		return errors.New("mime_type is required")
	}
	if c.UploadedBy == "" {
		return errors.New("uploaded_by is required")
	}
	return nil
}

// DeleteMediaCommand deletes a media file
type DeleteMediaCommand struct {
	cqrs.BaseCommand
}

func (c *DeleteMediaCommand) CommandName() string {
	return "delete_media"
}

func (c *DeleteMediaCommand) Validate() error {
	if c.AggregateID == "" {
		return errors.New("media_id is required")
	}
	return nil
}

// ============================================================================
// Command Handlers
// ============================================================================

// UploadMediaHandler handles media uploads
type UploadMediaHandler struct {
	mediaRepo     repository.MediaRepository
	storage       *storage.Storage
	kafkaProducer *kafka.Producer
	config        *config.Config
	baseURL       string
}

func NewUploadMediaHandler(
	mediaRepo repository.MediaRepository,
	kafkaProducer *kafka.Producer,
	cfg *config.Config,
) (*UploadMediaHandler, error) {
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
	baseURL := fmt.Sprintf("http://localhost:%s/media", cfg.Port)

	return &UploadMediaHandler{
		mediaRepo:     mediaRepo,
		storage:       storageInstance,
		kafkaProducer: kafkaProducer,
		config:        cfg,
		baseURL:       baseURL,
	}, nil
}

func (h *UploadMediaHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	uploadCmd := cmd.(*UploadMediaCommand)

	// Process image if needed (resize large images)
	processedData := uploadCmd.Data
	if h.isImageMimeType(uploadCmd.MimeType) && h.config.Media.AutoResizeLargeImage {
		resizedData, err := h.resizeLargeImage(uploadCmd.Data)
		if err == nil && resizedData != nil {
			processedData = resizedData
		}
	}

	// Save file to storage
	filename, err := h.storage.SaveFile(processedData, uploadCmd.OriginalFilename, uploadCmd.MimeType)
	if err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	// Generate UUID for media
	mediaID := uuid.New().String()

	// Create media entity
	media := &domain.Media{
		ID:               mediaID,
		Filename:         filepath.Base(filename),
		OriginalFilename: uploadCmd.OriginalFilename,
		MimeType:         uploadCmd.MimeType,
		Size:             int64(len(processedData)),
		Path:             filename,
		URL:              h.constructURL(filename),
		UploadedBy:       uploadCmd.UploadedBy,
	}

	// Extract image metadata and generate thumbnails if it's an image
	if h.isImageMimeType(uploadCmd.MimeType) {
		info, err := image.GetImageInfo(processedData)
		if err == nil {
			media.Width = info.Width
			media.Height = info.Height

			// Generate thumbnails if enabled
			if h.config.Media.GenerateThumbnails {
				thumbnailPath, err := h.generateThumbnail(filename)
				if err == nil && thumbnailPath != "" {
					thumbnailURL := h.constructURL(thumbnailPath)
					media.ThumbnailURL = &thumbnailURL

					// Publish thumbnail generated event
					if h.kafkaProducer != nil {
						event := kafka.MediaThumbnailGeneratedEvent{
							MediaID:      media.ID,
							ThumbnailURL: thumbnailURL,
							GeneratedAt:  time.Now(),
						}
						if err := h.kafkaProducer.PublishMediaThumbnailGenerated("blog.media.thumbnail.generated", event); err != nil {
							fmt.Printf("Warning: Failed to publish media thumbnail generated event: %v\n", err)
						}
					}
				}
			}
		}
	}

	// Save to database
	if err := h.mediaRepo.Create(ctx, media); err != nil {
		// Cleanup: delete file if database insert fails
		_ = h.storage.DeleteFile(filename)
		return fmt.Errorf("failed to create media record: %w", err)
	}

	// Store media ID in command result
	uploadCmd.AggregateID = mediaID

	// Publish Kafka event
	if h.kafkaProducer != nil {
		event := kafka.MediaUploadedEvent{
			MediaID:          media.ID,
			Filename:         media.Filename,
			OriginalFilename: media.OriginalFilename,
			MimeType:         media.MimeType,
			Size:             media.Size,
			URL:              media.URL,
			UploadedBy:       media.UploadedBy,
			UploadedAt:       media.CreatedAt,
		}
		if err := h.kafkaProducer.PublishMediaUploaded("blog.media.uploaded", event); err != nil {
			fmt.Printf("Warning: Failed to publish media uploaded event: %v\n", err)
		}
	}

	return nil
}

func (h *UploadMediaHandler) constructURL(filename string) string {
	// Convert Windows path separators to URL path separators
	urlPath := strings.ReplaceAll(filename, "\\", "/")
	return fmt.Sprintf("%s/%s", h.baseURL, urlPath)
}

func (h *UploadMediaHandler) isImageMimeType(mimeType string) bool {
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

func (h *UploadMediaHandler) generateThumbnail(originalPath string) (string, error) {
	// Get full path to original file
	fullPath := h.storage.GetFilePath(originalPath)

	// Generate medium thumbnail (300x300) - good balance for most uses
	thumbnailPath := image.GetThumbnailPath(originalPath, image.ThumbnailMedium)
	thumbnailFullPath := h.storage.GetFilePath(thumbnailPath)

	// Create thumbnail
	if err := image.GenerateThumbnail(fullPath, thumbnailFullPath, image.ThumbnailMedium); err != nil {
		return "", fmt.Errorf("failed to generate thumbnail: %w", err)
	}

	return thumbnailPath, nil
}

func (h *UploadMediaHandler) resizeLargeImage(data []byte) ([]byte, error) {
	// Get image info
	info, err := image.GetImageInfo(data)
	if err != nil {
		return nil, err
	}

	// Check if resize is needed
	maxWidth := h.config.Media.MaxImageWidth
	maxHeight := h.config.Media.MaxImageHeight

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

// DeleteMediaHandler handles media deletion
type DeleteMediaHandler struct {
	mediaRepo     repository.MediaRepository
	storage       *storage.Storage
	kafkaProducer *kafka.Producer
	config        *config.Config
}

func NewDeleteMediaHandler(
	mediaRepo repository.MediaRepository,
	kafkaProducer *kafka.Producer,
	cfg *config.Config,
) (*DeleteMediaHandler, error) {
	// Initialize storage
	storageInstance, err := storage.NewStorage(
		cfg.Media.StoragePath,
		cfg.Media.AllowedTypes,
		cfg.Media.MaxSize,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	return &DeleteMediaHandler{
		mediaRepo:     mediaRepo,
		storage:       storageInstance,
		kafkaProducer: kafkaProducer,
		config:        cfg,
	}, nil
}

func (h *DeleteMediaHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	deleteCmd := cmd.(*DeleteMediaCommand)

	// Get media record
	media, err := h.mediaRepo.GetByID(ctx, deleteCmd.AggregateID)
	if err != nil {
		return fmt.Errorf("media not found: %w", err)
	}

	// Delete from storage
	if err := h.storage.DeleteFile(media.Path); err != nil {
		// Log error but continue with database deletion
		// The file might already be deleted or missing
		fmt.Printf("Warning: Failed to delete file from storage: %v\n", err)
	}

	// Delete thumbnail if exists
	if media.ThumbnailURL != nil {
		// Extract path from URL and delete
		// This is a simplified approach
		// In production, you'd parse the URL properly
	}

	// Delete from database
	if err := h.mediaRepo.Delete(ctx, deleteCmd.AggregateID); err != nil {
		return fmt.Errorf("failed to delete media record: %w", err)
	}

	// Publish Kafka event
	if h.kafkaProducer != nil {
		event := kafka.MediaDeletedEvent{
			MediaID:   media.ID,
			Filename:  media.Filename,
			DeletedAt: time.Now(),
		}
		if err := h.kafkaProducer.PublishMediaDeleted("blog.media.deleted", event); err != nil {
			fmt.Printf("Warning: Failed to publish media deleted event: %v\n", err)
		}
	}

	return nil
}
