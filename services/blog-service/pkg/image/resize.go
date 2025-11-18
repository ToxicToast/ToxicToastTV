package image

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
)

// ThumbnailSize represents a thumbnail configuration
type ThumbnailSize struct {
	Name   string // e.g., "small", "medium", "large"
	Width  int
	Height int
	Suffix string // e.g., "_small", "_medium"
}

// Common thumbnail sizes
var (
	ThumbnailSmall = ThumbnailSize{
		Name:   "small",
		Width:  150,
		Height: 150,
		Suffix: "_thumb_small",
	}
	ThumbnailMedium = ThumbnailSize{
		Name:   "medium",
		Width:  300,
		Height: 300,
		Suffix: "_thumb_medium",
	}
	ThumbnailLarge = ThumbnailSize{
		Name:   "large",
		Width:  600,
		Height: 600,
		Suffix: "_thumb_large",
	}
)

// ResizeOptions configures image resizing behavior
type ResizeOptions struct {
	Width   int
	Height  int
	Fit     bool // If true, fit image within dimensions while maintaining aspect ratio
	Quality int  // JPEG quality (1-100), default 85
}

// ResizeImage resizes an image from a file
func ResizeImage(sourcePath, destPath string, opts ResizeOptions) error {
	// Open source image
	src, err := imaging.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}

	// Resize image
	var resized *image.NRGBA
	if opts.Fit {
		// Fit within dimensions while maintaining aspect ratio
		resized = imaging.Fit(src, opts.Width, opts.Height, imaging.Lanczos)
	} else {
		// Resize to exact dimensions (may distort)
		resized = imaging.Resize(src, opts.Width, opts.Height, imaging.Lanczos)
	}

	// Save resized image
	return saveImage(resized, destPath, opts.Quality)
}

// ResizeImageFromBytes resizes an image from byte data
func ResizeImageFromBytes(data []byte, opts ResizeOptions) ([]byte, error) {
	// Decode image
	src, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Resize image
	var resized *image.NRGBA
	if opts.Fit {
		resized = imaging.Fit(src, opts.Width, opts.Height, imaging.Lanczos)
	} else {
		resized = imaging.Resize(src, opts.Width, opts.Height, imaging.Lanczos)
	}

	// Encode to bytes
	return encodeImage(resized, format, opts.Quality)
}

// GenerateThumbnail creates a square thumbnail from an image file
func GenerateThumbnail(sourcePath, destPath string, size ThumbnailSize) error {
	// Open source image
	src, err := imaging.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}

	// Create square thumbnail (crop to center)
	thumbnail := imaging.Fill(src, size.Width, size.Height, imaging.Center, imaging.Lanczos)

	// Save thumbnail
	return saveImage(thumbnail, destPath, 85)
}

// GenerateThumbnailFromBytes creates a thumbnail from byte data
func GenerateThumbnailFromBytes(data []byte, size ThumbnailSize) ([]byte, error) {
	// Decode image
	src, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Create square thumbnail
	thumbnail := imaging.Fill(src, size.Width, size.Height, imaging.Center, imaging.Lanczos)

	// Encode to bytes
	return encodeImage(thumbnail, format, 85)
}

// GenerateMultipleThumbnails creates multiple thumbnail sizes for an image
func GenerateMultipleThumbnails(sourcePath string, sizes []ThumbnailSize) ([]string, error) {
	// Open source image once
	src, err := imaging.Open(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open image: %w", err)
	}

	var thumbnailPaths []string
	ext := filepath.Ext(sourcePath)
	basePathWithoutExt := strings.TrimSuffix(sourcePath, ext)

	for _, size := range sizes {
		// Generate thumbnail path
		thumbnailPath := basePathWithoutExt + size.Suffix + ext

		// Create thumbnail
		thumbnail := imaging.Fill(src, size.Width, size.Height, imaging.Center, imaging.Lanczos)

		// Save thumbnail
		if err := saveImage(thumbnail, thumbnailPath, 85); err != nil {
			return thumbnailPaths, fmt.Errorf("failed to save thumbnail %s: %w", size.Name, err)
		}

		thumbnailPaths = append(thumbnailPaths, thumbnailPath)
	}

	return thumbnailPaths, nil
}

// CropImage crops an image to specified dimensions
func CropImage(sourcePath, destPath string, x, y, width, height int) error {
	src, err := imaging.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}

	// Crop image
	cropped := imaging.Crop(src, image.Rect(x, y, x+width, y+height))

	// Save cropped image
	return saveImage(cropped, destPath, 90)
}

// saveImage saves an image to disk with appropriate format
func saveImage(img image.Image, path string, quality int) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Determine format from extension
	ext := strings.ToLower(filepath.Ext(path))

	// Set default quality if not specified
	if quality == 0 {
		quality = 85
	}

	// Save based on format
	switch ext {
	case ".jpg", ".jpeg":
		return imaging.Save(img, path, imaging.JPEGQuality(quality))
	case ".png":
		return imaging.Save(img, path)
	case ".gif":
		return imaging.Save(img, path)
	default:
		return imaging.Save(img, path)
	}
}

// encodeImage encodes an image to bytes
func encodeImage(img image.Image, format string, quality int) ([]byte, error) {
	var buf bytes.Buffer

	if quality == 0 {
		quality = 85
	}

	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
		if err != nil {
			return nil, err
		}
	case "png":
		err := png.Encode(&buf, img)
		if err != nil {
			return nil, err
		}
	case "gif":
		err := gif.Encode(&buf, img, nil)
		if err != nil {
			return nil, err
		}
	default:
		// Default to JPEG
		err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
		if err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

// GetThumbnailPath generates the path for a thumbnail
func GetThumbnailPath(originalPath string, size ThumbnailSize) string {
	ext := filepath.Ext(originalPath)
	basePathWithoutExt := strings.TrimSuffix(originalPath, ext)
	return basePathWithoutExt + size.Suffix + ext
}

// CalculateThumbnailDimensions calculates dimensions to fit within max size while maintaining aspect ratio
func CalculateThumbnailDimensions(originalWidth, originalHeight, maxWidth, maxHeight int) (int, int) {
	if originalWidth <= maxWidth && originalHeight <= maxHeight {
		return originalWidth, originalHeight
	}

	aspectRatio := float64(originalWidth) / float64(originalHeight)

	var newWidth, newHeight int

	if originalWidth > originalHeight {
		newWidth = maxWidth
		newHeight = int(float64(maxWidth) / aspectRatio)
	} else {
		newHeight = maxHeight
		newWidth = int(float64(maxHeight) * aspectRatio)
	}

	return newWidth, newHeight
}
