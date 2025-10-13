package file

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"
	"eduhub/server/internal/services/storage"

	"github.com/google/uuid"
)

type FileService interface {
	UploadFile(ctx context.Context, collegeID, userID int, file io.Reader, filename, contentType string, size int64, category, description string, folderID *int, tags []string) (*models.FileWithVersion, error)
	GetFile(ctx context.Context, collegeID, fileID int) (*models.FileWithVersion, error)
	ListFiles(ctx context.Context, collegeID int, folderID *int, category *string, limit, offset int) ([]*models.FileWithVersion, error)
	UpdateFile(ctx context.Context, collegeID, fileID int, name, description *string, category *string, folderID *int, tags *[]string, isPublic *bool) error
	DeleteFile(ctx context.Context, collegeID, fileID int) error

	GetFileVersions(ctx context.Context, fileID int) ([]*models.FileVersion, error)
	UploadNewVersion(ctx context.Context, collegeID, fileID, userID int, file io.Reader, filename, contentType string, size int64, comment string) (*models.FileVersion, error)
	SetCurrentVersion(ctx context.Context, collegeID, fileID, versionID int) error
	GetFileURL(ctx context.Context, collegeID, fileID int) (string, error)

	CreateFolder(ctx context.Context, collegeID, userID int, name string, parentID *int) (*models.Folder, error)
	GetFolder(ctx context.Context, collegeID, folderID int) (*models.Folder, error)
	ListFolders(ctx context.Context, collegeID int, parentID *int) ([]*models.Folder, error)
	UpdateFolder(ctx context.Context, collegeID, folderID int, name string, parentID *int) error
	DeleteFolder(ctx context.Context, collegeID, folderID int) error

	SearchFiles(ctx context.Context, collegeID int, query string, category *string, limit, offset int) ([]*models.FileWithVersion, error)
	GetFilesByTags(ctx context.Context, collegeID int, tags []string, limit, offset int) ([]*models.FileWithVersion, error)
}

type fileService struct {
	fileRepo    repository.FileRepository
	storageSvc  storage.StorageService
}

func NewFileService(fileRepo repository.FileRepository, storageSvc storage.StorageService) FileService {
	return &fileService{
		fileRepo:   fileRepo,
		storageSvc: storageSvc,
	}
}

func (s *fileService) UploadFile(ctx context.Context, collegeID, userID int, file io.Reader, filename, contentType string, size int64, category, description string, folderID *int, tags []string) (*models.FileWithVersion, error) {
	// Validate inputs
	if filename == "" {
		return nil, fmt.Errorf("filename is required")
	}
	if size <= 0 {
		return nil, fmt.Errorf("invalid file size")
	}

	// Generate unique object key
	objectKey := s.generateObjectKey(collegeID, userID, category, filename)

	// Calculate hash for integrity checking
	hash, err := s.calculateHash(file)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate file hash: %w", err)
	}

	// Reset file reader for upload
	if seeker, ok := file.(io.Seeker); ok {
		seeker.Seek(0, io.SeekStart)
	}

	// Upload to storage
	_, err = s.storageSvc.UploadFile(ctx, objectKey, file, size, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Create file record
	fileModel := &models.File{
		CollegeID:   collegeID,
		Name:        strings.TrimSuffix(filename, filepath.Ext(filename)),
		Description: description,
		Category:    category,
		FolderID:    folderID,
		UploadedBy:  userID,
		IsPublic:    false,
		Tags:        tags,
	}

	err = s.fileRepo.CreateFile(ctx, fileModel)
	if err != nil {
		// Cleanup uploaded file on error
		s.storageSvc.DeleteFile(ctx, objectKey)
		return nil, fmt.Errorf("failed to create file record: %w", err)
	}

	// Create initial version
	version := &models.FileVersion{
		FileID:      fileModel.ID,
		Version:     1,
		ObjectKey:   objectKey,
		Filename:    filename,
		Size:        size,
		ContentType: contentType,
		Hash:        hash,
		UploadedBy:  userID,
		Comment:     "Initial upload",
		IsCurrent:   true,
	}

	err = s.fileRepo.CreateFileVersion(ctx, version)
	if err != nil {
		// Cleanup on error
		s.fileRepo.DeleteFile(ctx, collegeID, fileModel.ID)
		s.storageSvc.DeleteFile(ctx, objectKey)
		return nil, fmt.Errorf("failed to create file version: %w", err)
	}

	// Update file with current version ID
	fileModel.CurrentVersionID = version.ID
	err = s.fileRepo.UpdateFile(ctx, fileModel)
	if err != nil {
		return nil, fmt.Errorf("failed to update file current version: %w", err)
	}

	// Return file with version info
	return s.fileRepo.GetFileByID(ctx, collegeID, fileModel.ID)
}

func (s *fileService) GetFile(ctx context.Context, collegeID, fileID int) (*models.FileWithVersion, error) {
	return s.fileRepo.GetFileByID(ctx, collegeID, fileID)
}

func (s *fileService) ListFiles(ctx context.Context, collegeID int, folderID *int, category *string, limit, offset int) ([]*models.FileWithVersion, error) {
	return s.fileRepo.ListFiles(ctx, collegeID, folderID, category, limit, offset)
}

func (s *fileService) UpdateFile(ctx context.Context, collegeID, fileID int, name, description *string, category *string, folderID *int, tags *[]string, isPublic *bool) error {
	file, err := s.fileRepo.GetFileByID(ctx, collegeID, fileID)
	if err != nil {
		return err
	}

	if name != nil {
		file.Name = *name
	}
	if description != nil {
		file.Description = *description
	}
	if category != nil {
		file.Category = *category
	}
	if folderID != nil {
		file.FolderID = folderID
	}
	if tags != nil {
		file.Tags = *tags
	}
	if isPublic != nil {
		file.IsPublic = *isPublic
	}

	return s.fileRepo.UpdateFile(ctx, &file.File)
}

func (s *fileService) DeleteFile(ctx context.Context, collegeID, fileID int) error {
	// Get file to find all versions for cleanup
	file, err := s.fileRepo.GetFileByID(ctx, collegeID, fileID)
	if err != nil {
		return err
	}

	// Delete all versions from storage
	for _, version := range file.Versions {
		s.storageSvc.DeleteFile(ctx, version.ObjectKey)
	}

	// Delete from database (cascade will handle versions)
	return s.fileRepo.DeleteFile(ctx, collegeID, fileID)
}

func (s *fileService) GetFileVersions(ctx context.Context, fileID int) ([]*models.FileVersion, error) {
	return s.fileRepo.GetFileVersions(ctx, fileID)
}

func (s *fileService) UploadNewVersion(ctx context.Context, collegeID, fileID, userID int, fileReader io.Reader, filename, contentType string, size int64, comment string) (*models.FileVersion, error) {
	// Get current file
	file, err := s.fileRepo.GetFileByID(ctx, collegeID, fileID)
	if err != nil {
		return nil, err
	}

	// Get next version number
	versions, err := s.fileRepo.GetFileVersions(ctx, fileID)
	if err != nil {
		return nil, err
	}

	nextVersion := 1
	if len(versions) > 0 {
		nextVersion = versions[0].Version + 1
	}

	// Generate object key
	objectKey := s.generateObjectKey(collegeID, userID, file.Category, filename)

	// Calculate hash
	hash, err := s.calculateHash(fileReader)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate file hash: %w", err)
	}

	// Reset file reader
	if seeker, ok := fileReader.(io.Seeker); ok {
		seeker.Seek(0, io.SeekStart)
	}

	// Upload to storage
	_, err = s.storageSvc.UploadFile(ctx, objectKey, fileReader, size, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Create version record
	version := &models.FileVersion{
		FileID:      fileID,
		Version:     nextVersion,
		ObjectKey:   objectKey,
		Filename:    filename,
		Size:        size,
		ContentType: contentType,
		Hash:        hash,
		UploadedBy:  userID,
		Comment:     comment,
		IsCurrent:   false,
	}

	err = s.fileRepo.CreateFileVersion(ctx, version)
	if err != nil {
		s.storageSvc.DeleteFile(ctx, objectKey)
		return nil, fmt.Errorf("failed to create file version: %w", err)
	}

	return version, nil
}

func (s *fileService) SetCurrentVersion(ctx context.Context, collegeID, fileID, versionID int) error {
	return s.fileRepo.SetCurrentVersion(ctx, fileID, versionID)
}

func (s *fileService) GetFileURL(ctx context.Context, collegeID, fileID int) (string, error) {
	file, err := s.fileRepo.GetFileByID(ctx, collegeID, fileID)
	if err != nil {
		return "", err
	}

	if file.CurrentVersion == nil {
		return "", fmt.Errorf("no current version found for file")
	}

	return s.storageSvc.GetFileURL(ctx, file.CurrentVersion.ObjectKey)
}

func (s *fileService) CreateFolder(ctx context.Context, collegeID, userID int, name string, parentID *int) (*models.Folder, error) {
	if name == "" {
		return nil, fmt.Errorf("folder name is required")
	}

	path := name
	if parentID != nil {
		parentPath, err := s.fileRepo.GetFolderPath(ctx, *parentID)
		if err != nil {
			return nil, fmt.Errorf("failed to get parent folder path: %w", err)
		}
		path = parentPath + "/" + name
	}

	folder := &models.Folder{
		CollegeID: collegeID,
		Name:      name,
		ParentID:  parentID,
		Path:      path,
		CreatedBy: userID,
	}

	err := s.fileRepo.CreateFolder(ctx, folder)
	if err != nil {
		return nil, fmt.Errorf("failed to create folder: %w", err)
	}

	return folder, nil
}

func (s *fileService) GetFolder(ctx context.Context, collegeID, folderID int) (*models.Folder, error) {
	return s.fileRepo.GetFolderByID(ctx, collegeID, folderID)
}

func (s *fileService) ListFolders(ctx context.Context, collegeID int, parentID *int) ([]*models.Folder, error) {
	return s.fileRepo.ListFolders(ctx, collegeID, parentID)
}

func (s *fileService) UpdateFolder(ctx context.Context, collegeID, folderID int, name string, parentID *int) error {
	folder, err := s.fileRepo.GetFolderByID(ctx, collegeID, folderID)
	if err != nil {
		return err
	}

	folder.Name = name
	folder.ParentID = parentID

	// Update path
	path := name
	if parentID != nil {
		parentPath, err := s.fileRepo.GetFolderPath(ctx, *parentID)
		if err != nil {
			return fmt.Errorf("failed to get parent folder path: %w", err)
		}
		path = parentPath + "/" + name
	}
	folder.Path = path

	return s.fileRepo.UpdateFolder(ctx, folder)
}

func (s *fileService) DeleteFolder(ctx context.Context, collegeID, folderID int) error {
	return s.fileRepo.DeleteFolder(ctx, collegeID, folderID)
}

func (s *fileService) SearchFiles(ctx context.Context, collegeID int, query string, category *string, limit, offset int) ([]*models.FileWithVersion, error) {
	return s.fileRepo.SearchFiles(ctx, collegeID, query, category, limit, offset)
}

func (s *fileService) GetFilesByTags(ctx context.Context, collegeID int, tags []string, limit, offset int) ([]*models.FileWithVersion, error) {
	return s.fileRepo.GetFilesByTags(ctx, collegeID, tags, limit, offset)
}

func (s *fileService) generateObjectKey(collegeID, userID int, category, filename string) string {
	uniqueID := uuid.New().String()
	return fmt.Sprintf("%d/%d/%s/%s_%s", collegeID, userID, category, uniqueID, filename)
}

func (s *fileService) calculateHash(reader io.Reader) (string, error) {
	hasher := sha256.New()
	if _, err := io.Copy(hasher, reader); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}