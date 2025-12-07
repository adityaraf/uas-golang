package utils

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type FileUploadConfig struct {
	UploadPath       string
	MaxFileSize      int64
	AllowedFileTypes []string
}

var DefaultUploadConfig = FileUploadConfig{
	UploadPath:       "./uploads/achievements",
	MaxFileSize:      5 * 1024 * 1024, // 5MB
	AllowedFileTypes: []string{".pdf", ".jpg", ".jpeg", ".png", ".doc", ".docx"},
}

// SaveUploadedFile menyimpan file yang diupload
func SaveUploadedFile(file *multipart.FileHeader, config FileUploadConfig) (string, error) {
	// Validasi ukuran file
	if file.Size > config.MaxFileSize {
		return "", fmt.Errorf("ukuran file terlalu besar. Maksimal %d MB", config.MaxFileSize/(1024*1024))
	}

	// Validasi tipe file
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !isAllowedFileType(ext, config.AllowedFileTypes) {
		return "", fmt.Errorf("tipe file tidak diizinkan. Hanya: %v", config.AllowedFileTypes)
	}

	// Buat folder jika belum ada
	if err := os.MkdirAll(config.UploadPath, 0755); err != nil {
		return "", fmt.Errorf("gagal membuat folder upload: %v", err)
	}

	// Generate unique filename
	uniqueFilename := generateUniqueFilename(file.Filename)
	filepath := filepath.Join(config.UploadPath, uniqueFilename)

	// Buka file source
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("gagal membuka file: %v", err)
	}
	defer src.Close()

	// Buat file destination
	dst, err := os.Create(filepath)
	if err != nil {
		return "", fmt.Errorf("gagal membuat file: %v", err)
	}
	defer dst.Close()

	// Copy file
	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("gagal menyimpan file: %v", err)
	}

	return filepath, nil
}

// SaveMultipleFiles menyimpan multiple files
func SaveMultipleFiles(form *multipart.Form, fieldName string, config FileUploadConfig) ([]string, error) {
	files := form.File[fieldName]
	if len(files) == 0 {
		return []string{}, nil
	}

	var savedFiles []string
	for _, file := range files {
		filepath, err := SaveUploadedFile(file, config)
		if err != nil {
			// Rollback: hapus file yang sudah tersimpan
			for _, savedFile := range savedFiles {
				os.Remove(savedFile)
			}
			return nil, err
		}
		savedFiles = append(savedFiles, filepath)
	}

	return savedFiles, nil
}

// DeleteFile menghapus file
func DeleteFile(filepath string) error {
	if filepath == "" {
		return nil
	}
	return os.Remove(filepath)
}

// isAllowedFileType mengecek apakah tipe file diizinkan
func isAllowedFileType(ext string, allowedTypes []string) bool {
	for _, allowedType := range allowedTypes {
		if ext == allowedType {
			return true
		}
	}
	return false
}

// generateUniqueFilename generate nama file yang unik
func generateUniqueFilename(originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	nameWithoutExt := strings.TrimSuffix(originalFilename, ext)

	// Sanitize filename
	nameWithoutExt = strings.ReplaceAll(nameWithoutExt, " ", "_")

	// Generate unique name dengan timestamp dan UUID
	timestamp := time.Now().Format("20060102_150405")
	uniqueID := uuid.New().String()[:8]

	return fmt.Sprintf("%s_%s_%s%s", nameWithoutExt, timestamp, uniqueID, ext)
}

// GetFileInfo mendapatkan informasi file
func GetFileInfo(file *multipart.FileHeader) (filename string, size int64, mimetype string) {
	return file.Filename, file.Size, file.Header.Get("Content-Type")
}