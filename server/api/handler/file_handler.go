package handler

import (
	"encoding/json"
	"path/filepath"
	"strconv"
	"strings"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/services/file"

	"github.com/labstack/echo/v4"
)

type FileHandler struct {
	fileService file.FileService
}

func NewFileHandler(fileService file.FileService) *FileHandler {
	return &FileHandler{
		fileService: fileService,
	}
}

// UploadFile handles file upload with versioning support
func (h *FileHandler) UploadFile(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	userID, err := helpers.ExtractUserID(c)
	if err != nil {
		return helpers.Error(c, "user ID required", 401)
	}

	// Get file from form
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return helpers.Error(c, "file is required", 400)
	}

	// Get other form parameters
	category := c.FormValue("category")
	if category == "" {
		category = "document"
	}

	description := c.FormValue("description")
	folderIDStr := c.FormValue("folder_id")
	var folderID *int
	if folderIDStr != "" {
		if id, err := strconv.Atoi(folderIDStr); err == nil {
			folderID = &id
		}
	}

	tagsStr := c.FormValue("tags")
	var tags []string
	if tagsStr != "" {
		json.Unmarshal([]byte(tagsStr), &tags)
	}

	// Validate file size (50MB limit for versioned files)
	if fileHeader.Size > 50*1024*1024 {
		return helpers.Error(c, "file size exceeds 50MB limit", 400)
	}

	// Validate file type
	allowedTypes := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
		".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
		".txt": true, ".zip": true, ".rar": true, ".ppt": true, ".pptx": true,
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !allowedTypes[ext] {
		return helpers.Error(c, "file type not allowed", 400)
	}

	// Open file
	src, err := fileHeader.Open()
	if err != nil {
		return helpers.Error(c, "failed to open file", 500)
	}
	defer src.Close()

	// Upload file
	fileWithVersion, err := h.fileService.UploadFile(
		c.Request().Context(),
		collegeID,
		userID,
		src,
		fileHeader.Filename,
		fileHeader.Header.Get("Content-Type"),
		fileHeader.Size,
		category,
		description,
		folderID,
		tags,
	)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, fileWithVersion, 201)
}

// GetFile retrieves file information with current version
func (h *FileHandler) GetFile(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	fileIDStr := c.Param("fileID")
	fileID, err := strconv.Atoi(fileIDStr)
	if err != nil {
		return helpers.Error(c, "invalid file ID", 400)
	}

	fileWithVersion, err := h.fileService.GetFile(c.Request().Context(), collegeID, fileID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, fileWithVersion, 200)
}

// ListFiles lists files with optional filtering
func (h *FileHandler) ListFiles(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	folderIDStr := c.QueryParam("folder_id")
	var folderID *int
	if folderIDStr != "" {
		if id, err := strconv.Atoi(folderIDStr); err == nil {
			folderID = &id
		}
	}

	category := c.QueryParam("category")
	var categoryPtr *string
	if category != "" {
		categoryPtr = &category
	}

	limitStr := c.QueryParam("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offsetStr := c.QueryParam("offset")
	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	files, err := h.fileService.ListFiles(c.Request().Context(), collegeID, folderID, categoryPtr, limit, offset)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, files, 200)
}

// UpdateFile updates file metadata
func (h *FileHandler) UpdateFile(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	fileIDStr := c.Param("fileID")
	fileID, err := strconv.Atoi(fileIDStr)
	if err != nil {
		return helpers.Error(c, "invalid file ID", 400)
	}

	var req struct {
		Name        *string  `json:"name"`
		Description *string  `json:"description"`
		Category    *string  `json:"category"`
		FolderID    *int     `json:"folder_id"`
		Tags        *[]string `json:"tags"`
		IsPublic    *bool    `json:"is_public"`
	}

	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	err = h.fileService.UpdateFile(
		c.Request().Context(),
		collegeID,
		fileID,
		req.Name,
		req.Description,
		req.Category,
		req.FolderID,
		req.Tags,
		req.IsPublic,
	)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "File updated successfully", 200)
}

// DeleteFile deletes a file and all its versions
func (h *FileHandler) DeleteFile(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	fileIDStr := c.Param("fileID")
	fileID, err := strconv.Atoi(fileIDStr)
	if err != nil {
		return helpers.Error(c, "invalid file ID", 400)
	}

	err = h.fileService.DeleteFile(c.Request().Context(), collegeID, fileID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "File deleted successfully", 200)
}

// GetFileVersions retrieves all versions of a file
func (h *FileHandler) GetFileVersions(c echo.Context) error {
	fileIDStr := c.Param("fileID")
	fileID, err := strconv.Atoi(fileIDStr)
	if err != nil {
		return helpers.Error(c, "invalid file ID", 400)
	}

	versions, err := h.fileService.GetFileVersions(c.Request().Context(), fileID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, versions, 200)
}

// UploadNewVersion uploads a new version of an existing file
func (h *FileHandler) UploadNewVersion(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	userID, err := helpers.ExtractUserID(c)
	if err != nil {
		return helpers.Error(c, "user ID required", 401)
	}

	fileIDStr := c.Param("fileID")
	fileID, err := strconv.Atoi(fileIDStr)
	if err != nil {
		return helpers.Error(c, "invalid file ID", 400)
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return helpers.Error(c, "file is required", 400)
	}

	comment := c.FormValue("comment")

	// Validate file size
	if fileHeader.Size > 50*1024*1024 {
		return helpers.Error(c, "file size exceeds 50MB limit", 400)
	}

	src, err := fileHeader.Open()
	if err != nil {
		return helpers.Error(c, "failed to open file", 500)
	}
	defer src.Close()

	version, err := h.fileService.UploadNewVersion(
		c.Request().Context(),
		collegeID,
		fileID,
		userID,
		src,
		fileHeader.Filename,
		fileHeader.Header.Get("Content-Type"),
		fileHeader.Size,
		comment,
	)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, version, 201)
}

// SetCurrentVersion sets a specific version as the current version
func (h *FileHandler) SetCurrentVersion(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	fileIDStr := c.Param("fileID")
	fileID, err := strconv.Atoi(fileIDStr)
	if err != nil {
		return helpers.Error(c, "invalid file ID", 400)
	}

	versionIDStr := c.Param("versionID")
	versionID, err := strconv.Atoi(versionIDStr)
	if err != nil {
		return helpers.Error(c, "invalid version ID", 400)
	}

	err = h.fileService.SetCurrentVersion(c.Request().Context(), collegeID, fileID, versionID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "Current version updated successfully", 200)
}

// GetFileURL generates a presigned URL for file download
func (h *FileHandler) GetFileURL(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	fileIDStr := c.Param("fileID")
	fileID, err := strconv.Atoi(fileIDStr)
	if err != nil {
		return helpers.Error(c, "invalid file ID", 400)
	}

	url, err := h.fileService.GetFileURL(c.Request().Context(), collegeID, fileID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, map[string]string{"url": url}, 200)
}

// CreateFolder creates a new folder
func (h *FileHandler) CreateFolder(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	userID, err := helpers.ExtractUserID(c)
	if err != nil {
		return helpers.Error(c, "user ID required", 401)
	}

	var req struct {
		Name     string `json:"name" validate:"required"`
		ParentID *int   `json:"parent_id"`
	}

	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	folder, err := h.fileService.CreateFolder(c.Request().Context(), collegeID, userID, req.Name, req.ParentID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, folder, 201)
}

// GetFolder retrieves folder information
func (h *FileHandler) GetFolder(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	folderIDStr := c.Param("folderID")
	folderID, err := strconv.Atoi(folderIDStr)
	if err != nil {
		return helpers.Error(c, "invalid folder ID", 400)
	}

	folder, err := h.fileService.GetFolder(c.Request().Context(), collegeID, folderID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, folder, 200)
}

// ListFolders lists folders
func (h *FileHandler) ListFolders(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	parentIDStr := c.QueryParam("parent_id")
	var parentID *int
	if parentIDStr != "" {
		if id, err := strconv.Atoi(parentIDStr); err == nil {
			parentID = &id
		}
	}

	folders, err := h.fileService.ListFolders(c.Request().Context(), collegeID, parentID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, folders, 200)
}

// UpdateFolder updates folder information
func (h *FileHandler) UpdateFolder(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	folderIDStr := c.Param("folderID")
	folderID, err := strconv.Atoi(folderIDStr)
	if err != nil {
		return helpers.Error(c, "invalid folder ID", 400)
	}

	var req struct {
		Name     string `json:"name" validate:"required"`
		ParentID *int   `json:"parent_id"`
	}

	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	err = h.fileService.UpdateFolder(c.Request().Context(), collegeID, folderID, req.Name, req.ParentID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "Folder updated successfully", 200)
}

// DeleteFolder deletes a folder
func (h *FileHandler) DeleteFolder(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	folderIDStr := c.Param("folderID")
	folderID, err := strconv.Atoi(folderIDStr)
	if err != nil {
		return helpers.Error(c, "invalid folder ID", 400)
	}

	err = h.fileService.DeleteFolder(c.Request().Context(), collegeID, folderID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "Folder deleted successfully", 200)
}

// SearchFiles searches files by query
func (h *FileHandler) SearchFiles(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	query := c.QueryParam("q")
	if query == "" {
		return helpers.Error(c, "search query is required", 400)
	}

	category := c.QueryParam("category")
	var categoryPtr *string
	if category != "" {
		categoryPtr = &category
	}

	limitStr := c.QueryParam("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offsetStr := c.QueryParam("offset")
	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	files, err := h.fileService.SearchFiles(c.Request().Context(), collegeID, query, categoryPtr, limit, offset)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, files, 200)
}

// GetFilesByTags retrieves files by tags
func (h *FileHandler) GetFilesByTags(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	tagsParam := c.QueryParam("tags")
	if tagsParam == "" {
		return helpers.Error(c, "tags parameter is required", 400)
	}

	var tags []string
	if err := json.Unmarshal([]byte(tagsParam), &tags); err != nil {
		return helpers.Error(c, "invalid tags format", 400)
	}

	limitStr := c.QueryParam("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offsetStr := c.QueryParam("offset")
	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	files, err := h.fileService.GetFilesByTags(c.Request().Context(), collegeID, tags, limit, offset)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, files, 200)
}