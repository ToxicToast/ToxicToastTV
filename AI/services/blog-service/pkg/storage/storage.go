package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Storage handles file operations for media uploads
type Storage struct {
	basePath     string
	allowedTypes []string
	maxSize      int64
}

// NewStorage creates a new storage instance
func NewStorage(basePath string, allowedTypes []string, maxSize int64) (*Storage, error) {
	// Create base directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &Storage{
		basePath:     basePath,
		allowedTypes: allowedTypes,
		maxSize:      maxSize,
	}, nil
}

// SaveFile saves a file to disk and returns the filename
func (s *Storage) SaveFile(data []byte, originalFilename, mimeType string) (string, error) {
	// Validate MIME type
	if !s.isAllowedMimeType(mimeType) {
		return "", fmt.Errorf("mime type not allowed: %s", mimeType)
	}

	// Validate file size
	if int64(len(data)) > s.maxSize {
		return "", fmt.Errorf("file size exceeds maximum allowed size of %d bytes", s.maxSize)
	}

	// Generate unique filename
	ext := filepath.Ext(originalFilename)
	filename := fmt.Sprintf("%s-%s%s", time.Now().Format("20060102"), uuid.New().String(), ext)

	// Create year/month subdirectory
	subDir := time.Now().Format("2006/01")
	fullPath := filepath.Join(s.basePath, subDir)
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create subdirectory: %w", err)
	}

	// Full file path
	filePath := filepath.Join(fullPath, filename)

	// Write file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	// Return relative path (from basePath)
	relativePath := filepath.Join(subDir, filename)
	return relativePath, nil
}

// DeleteFile removes a file from disk
func (s *Storage) DeleteFile(filename string) error {
	fullPath := filepath.Join(s.basePath, filename)

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filename)
	}

	// Delete file
	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// GetFilePath returns the full path to a file
func (s *Storage) GetFilePath(filename string) string {
	return filepath.Join(s.basePath, filename)
}

// FileExists checks if a file exists
func (s *Storage) FileExists(filename string) bool {
	fullPath := filepath.Join(s.basePath, filename)
	_, err := os.Stat(fullPath)
	return err == nil
}

// GetFileSize returns the size of a file in bytes
func (s *Storage) GetFileSize(filename string) (int64, error) {
	fullPath := filepath.Join(s.basePath, filename)
	info, err := os.Stat(fullPath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// CopyFile copies a file to a new location
func (s *Storage) CopyFile(src, dst string) error {
	srcPath := filepath.Join(s.basePath, src)
	dstPath := filepath.Join(s.basePath, dst)

	// Create destination directory if needed
	dstDir := filepath.Dir(dstPath)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Open source file
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// Copy data
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

// isAllowedMimeType checks if a MIME type is allowed
func (s *Storage) isAllowedMimeType(mimeType string) bool {
	if len(s.allowedTypes) == 0 {
		return true // No restrictions
	}

	mimeType = strings.ToLower(mimeType)
	for _, allowed := range s.allowedTypes {
		if strings.ToLower(allowed) == mimeType {
			return true
		}
	}

	return false
}

// GetMimeTypeFromExtension returns the MIME type for a file extension
func GetMimeTypeFromExtension(ext string) string {
	ext = strings.ToLower(ext)
	mimeTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".svg":  "image/svg+xml",
		".pdf":  "application/pdf",
		".mp4":  "video/mp4",
		".webm": "video/webm",
	}

	if mimeType, ok := mimeTypes[ext]; ok {
		return mimeType
	}

	return "application/octet-stream"
}
