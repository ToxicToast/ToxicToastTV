package image

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"

	_ "golang.org/x/image/webp"
)

// ImageInfo contains metadata about an image
type ImageInfo struct {
	Width  int
	Height int
	Format string
}

// GetImageInfo extracts metadata from image data
func GetImageInfo(data []byte) (*ImageInfo, error) {
	reader := bytes.NewReader(data)
	return GetImageInfoFromReader(reader)
}

// GetImageInfoFromReader extracts metadata from an io.Reader
func GetImageInfoFromReader(reader io.Reader) (*ImageInfo, error) {
	config, format, err := image.DecodeConfig(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return &ImageInfo{
		Width:  config.Width,
		Height: config.Height,
		Format: format,
	}, nil
}

// GetImageInfoFromFile extracts metadata from a file path
func GetImageInfoFromFile(filePath string) (*ImageInfo, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return GetImageInfoFromReader(file)
}

// IsImage checks if data represents a valid image
func IsImage(data []byte) bool {
	_, err := GetImageInfo(data)
	return err == nil
}

// ValidateImageSize checks if image dimensions are within acceptable limits
func ValidateImageSize(width, height, maxWidth, maxHeight int) error {
	if maxWidth > 0 && width > maxWidth {
		return fmt.Errorf("image width %d exceeds maximum of %d", width, maxWidth)
	}

	if maxHeight > 0 && height > maxHeight {
		return fmt.Errorf("image height %d exceeds maximum of %d", height, maxHeight)
	}

	return nil
}
