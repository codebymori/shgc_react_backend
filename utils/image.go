package utils

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	MaxImageSize = 5 * 1024 * 1024 // 5MB
	UploadDir    = "./uploads"
)

var (
	// Allowed image MIME types
	allowedMimeTypes = map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/heic": true,
		"image/webp": true,
	}

	// Allowed file extensions
	allowedExtensions = map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".heic": true,
		".webp": true,
	}

	// Magic bytes for image validation
	// This is used to verify the actual file content, not just the extension
	imageMagicBytes = map[string][]byte{
		"image/jpeg": {0xFF, 0xD8, 0xFF},
		"image/png":  {0x89, 0x50, 0x4E, 0x47},
		"image/heic": {0x00, 0x00, 0x00}, // HEIC starts with ftyp in bytes 4-8
		"image/webp": {0x52, 0x49, 0x46, 0x46}, // WebP starts with RIFF (bytes 0-3) and WEBP (bytes 8-11)
	}
)

// ImageUploadResult contains the result of an image upload
type ImageUploadResult struct {
	Filename string
	Path     string
	URL      string
	Size     int64
}

// ValidateImageFile validates the uploaded image file
func ValidateImageFile(file multipart.File, header *multipart.FileHeader) error {
	// Check file size
	if header.Size > MaxImageSize {
		return errors.New("file size exceeds maximum allowed size of 5MB")
	}

	if header.Size == 0 {
		return errors.New("file is empty")
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedExtensions[ext] {
		return fmt.Errorf("file extension %s is not allowed. Only .jpg, .jpeg, .png, .heic, and .webp are allowed", ext)
	}

	// Read the first 512 bytes to detect content type
	buffer := make([]byte, 512)
	_, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return errors.New("failed to read file")
	}

	// Reset file pointer to beginning
	if _, err := file.Seek(0, 0); err != nil {
		return errors.New("failed to reset file pointer")
	}

	// Detect content type using http.DetectContentType
	contentType := http.DetectContentType(buffer)
	
	// Validate MIME type
	if !allowedMimeTypes[contentType] {
		// Special handling for HEIC as it might be detected as application/octet-stream
		if ext == ".heic" {
			// Additional HEIC validation by checking file signature
			if !isValidHEIC(buffer) {
				return fmt.Errorf("file content type %s does not match HEIC format", contentType)
			}
		} else {
			return fmt.Errorf("file content type %s is not allowed. Only JPEG, PNG, HEIC, and WebP images are allowed", contentType)
		}
	}

	// Verify magic bytes match the detected content type
	if !verifyMagicBytes(buffer, contentType, ext) {
		return errors.New("file content does not match its extension. Possible file manipulation detected")
	}

	return nil
}

// verifyMagicBytes checks if the file's magic bytes match the expected format
func verifyMagicBytes(buffer []byte, contentType string, ext string) bool {
	switch contentType {
	case "image/jpeg":
		// JPEG magic bytes: FF D8 FF
		return len(buffer) >= 3 && buffer[0] == 0xFF && buffer[1] == 0xD8 && buffer[2] == 0xFF
	case "image/png":
		// PNG magic bytes: 89 50 4E 47
		return len(buffer) >= 4 && buffer[0] == 0x89 && buffer[1] == 0x50 && buffer[2] == 0x4E && buffer[3] == 0x47
	case "image/webp":
		// WebP magic bytes: RIFF (0-3) ... WEBP (8-11)
		return len(buffer) >= 12 && 
			buffer[0] == 0x52 && buffer[1] == 0x49 && buffer[2] == 0x46 && buffer[3] == 0x46 && // RIFF
			buffer[8] == 0x57 && buffer[9] == 0x45 && buffer[10] == 0x42 && buffer[11] == 0x50 // WEBP
	default:
		// For HEIC and other formats
		if ext == ".heic" {
			return isValidHEIC(buffer)
		}
	}
	return false
}

// isValidHEIC validates HEIC file format by checking its signature
func isValidHEIC(buffer []byte) bool {
	// HEIC files have 'ftyp' at bytes 4-7 and 'heic' or 'mif1' at bytes 8-11
	if len(buffer) < 12 {
		return false
	}
	
	// Check for ftyp box
	if !(buffer[4] == 'f' && buffer[5] == 't' && buffer[6] == 'y' && buffer[7] == 'p') {
		return false
	}
	
	// Check for heic or mif1 brand
	heicBrand := buffer[8] == 'h' && buffer[9] == 'e' && buffer[10] == 'i' && buffer[11] == 'c'
	mif1Brand := buffer[8] == 'm' && buffer[9] == 'i' && buffer[10] == 'f' && buffer[11] == '1'
	
	return heicBrand || mif1Brand
}

// SaveImage saves the validated image file to the upload directory
func SaveImage(file multipart.File, header *multipart.FileHeader, subfolder string) (*ImageUploadResult, error) {
	// Validate the image first
	if err := ValidateImageFile(file, header); err != nil {
		return nil, err
	}

	// Create upload directory if it doesn't exist
	uploadPath := filepath.Join(UploadDir, subfolder)
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		return nil, errors.New("failed to create upload directory")
	}

	// Generate unique filename
	ext := strings.ToLower(filepath.Ext(header.Filename))
	filename := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().Unix(), ext)
	fullPath := filepath.Join(uploadPath, filename)

	// Create the file
	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, errors.New("failed to create file")
	}
	defer dst.Close()

	// Copy the uploaded file to the destination
	if _, err := io.Copy(dst, file); err != nil {
		// Clean up the file if copy fails
		os.Remove(fullPath)
		return nil, errors.New("failed to save file")
	}

	// Verify the saved file one more time
	savedFile, err := os.Open(fullPath)
	if err != nil {
		os.Remove(fullPath)
		return nil, errors.New("failed to verify saved file")
	}
	defer savedFile.Close()

	// Create a temporary header for validation
	tempHeader := &multipart.FileHeader{
		Filename: filename,
		Size:     header.Size,
	}

	if err := ValidateImageFile(savedFile, tempHeader); err != nil {
		os.Remove(fullPath)
		return nil, fmt.Errorf("saved file validation failed: %v", err)
	}

	// Get file info
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		return nil, errors.New("failed to get file info")
	}

	// Return relative URL path
	imageURL := fmt.Sprintf("/uploads/%s/%s", subfolder, filename)

	result := &ImageUploadResult{
		Filename: filename,
		Path:     fullPath,
		URL:      imageURL,
		Size:     fileInfo.Size(),
	}

	return result, nil
}

// DeleteImage deletes an image file
func DeleteImage(imagePath string) error {
	if imagePath == "" {
		return nil
	}

	// Convert URL path to file system path
	// e.g., /uploads/news/image.jpg -> ./uploads/news/image.jpg
	filePath := strings.TrimPrefix(imagePath, "/")
	filePath = filepath.Join(".", filePath)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil // File doesn't exist, consider it as success
	}

	// Delete the file
	if err := os.Remove(filePath); err != nil {
		return errors.New("failed to delete image file")
	}

	return nil
}

// ValidateImageBuffer validates image content from a byte buffer
// This is useful for additional validation after reading the entire file
func ValidateImageBuffer(buffer []byte, filename string) error {
	if len(buffer) == 0 {
		return errors.New("empty file buffer")
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if !allowedExtensions[ext] {
		return fmt.Errorf("invalid file extension: %s", ext)
	}

	contentType := http.DetectContentType(buffer)
	if !allowedMimeTypes[contentType] && ext != ".heic" {
		return fmt.Errorf("invalid content type: %s", contentType)
	}

	if !verifyMagicBytes(buffer, contentType, ext) {
		return errors.New("file content validation failed")
	}

	return nil
}

// ReadAndValidateImage reads the entire image file and validates it
func ReadAndValidateImage(file multipart.File, header *multipart.FileHeader) ([]byte, error) {
	// First validation
	if err := ValidateImageFile(file, header); err != nil {
		return nil, err
	}

	// Reset file pointer
	if _, err := file.Seek(0, 0); err != nil {
		return nil, errors.New("failed to reset file pointer")
	}

	// Read entire file into buffer
	buffer := new(bytes.Buffer)
	size, err := buffer.ReadFrom(file)
	if err != nil {
		return nil, errors.New("failed to read file")
	}

	if size != header.Size {
		return nil, errors.New("file size mismatch")
	}

	// Validate the buffer
	if err := ValidateImageBuffer(buffer.Bytes(), header.Filename); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// PrependBaseURL adds BASE_URL to image URL if it's not already a full URL
// It also handles fixing stale absolute URLs that point to local uploads
func PrependBaseURL(imageURL, baseURL string) string {
	if imageURL == "" || baseURL == "" {
		return imageURL
	}
	
	// If the URL is a local upload path (contains /uploads/), 
	// ensure it uses the current BASE_URL regardless of what's stored in DB
	if idx := strings.Index(imageURL, "/uploads/"); idx != -1 {
		// Strip everything before /uploads/ (including old domain/port)
		// e.g., "http://localhost:8000/uploads/img.jpg" -> "/uploads/img.jpg"
		cleanPath := imageURL[idx:]
		return baseURL + cleanPath
	}

	// Check if URL is already absolute (external URL)
	if strings.HasPrefix(imageURL, "http://") || strings.HasPrefix(imageURL, "https://") {
		return imageURL
	}
	return baseURL + imageURL
}
