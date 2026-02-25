package handler

import (
	"fmt"
	"path/filepath"
	"strings"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/services/storage"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type FileUploadHandler struct {
	storageService storage.StorageService
}

func NewFileUploadHandler(storageService storage.StorageService) *FileUploadHandler {
	return &FileUploadHandler{
		storageService: storageService,
	}
}

// UploadFile handles file upload requests
func (h *FileUploadHandler) UploadFile(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	userID, err := helpers.ExtractUserID(c)
	if err != nil {
		return helpers.Error(c, "user ID required", 401)
	}

	// Get file from form
	file, err := c.FormFile("file")
	if err != nil {
		return helpers.Error(c, "file is required", 400)
	}

	// Get upload type (profile, assignment, document, etc.)
	uploadType := c.FormValue("type")
	if uploadType == "" {
		uploadType = "document"
	}

	// Validate file size (10MB limit)
	if file.Size > 10*1024*1024 {
		return helpers.Error(c, "file size exceeds 10MB limit", 400)
	}

	// Validate file type
	allowedTypes := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".pdf":  true,
		".doc":  true,
		".docx": true,
		".xls":  true,
		".xlsx": true,
		".txt":  true,
		".zip":  true,
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedTypes[ext] {
		return helpers.Error(c, "file type not allowed", 400)
	}

	// Open file
	src, err := file.Open()
	if err != nil {
		return helpers.Error(c, "failed to open file", 500)
	}
	defer src.Close()

	// Generate unique filename
	uniqueFilename := fmt.Sprintf("%s_%s%s", uuid.New().String(), filepath.Base(file.Filename), ext)

	// Build storage path: college_id/user_id/type/filename
	objectKey := fmt.Sprintf("%d/%d/%s/%s", collegeID, userID, uploadType, uniqueFilename)

	// Upload to storage
	fileURL, err := h.storageService.UploadFile(c.Request().Context(), objectKey, src, file.Size, file.Header.Get("Content-Type"))
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, map[string]any{
		"url":      fileURL,
		"filename": file.Filename,
		"size":     file.Size,
		"type":     uploadType,
	}, 201)
}

// DeleteFile deletes a file from storage
func (h *FileUploadHandler) DeleteFile(c echo.Context) error {
	objectKey := c.QueryParam("key")
	if objectKey == "" {
		return helpers.Error(c, "object key is required", 400)
	}

	err := h.storageService.DeleteFile(c.Request().Context(), objectKey)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "File deleted successfully", 200)
}

// GetFileURL generates a presigned URL for file download
func (h *FileUploadHandler) GetFileURL(c echo.Context) error {
	objectKey := c.QueryParam("key")
	if objectKey == "" {
		return helpers.Error(c, "object key is required", 400)
	}

	url, err := h.storageService.GetFileURL(c.Request().Context(), objectKey)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, map[string]string{"url": url}, 200)
}
