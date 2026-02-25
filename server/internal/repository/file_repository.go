package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"eduhub/server/internal/models"

	"github.com/georgysavva/scany/v2/pgxscan"
)

type FileRepository interface {
	CreateFile(ctx context.Context, file *models.File) error
	GetFileByID(ctx context.Context, collegeID, fileID int) (*models.FileWithVersion, error)
	ListFiles(ctx context.Context, collegeID int, folderID *int, category *string, limit, offset int) ([]*models.FileWithVersion, error)
	UpdateFile(ctx context.Context, file *models.File) error
	DeleteFile(ctx context.Context, collegeID, fileID int) error

	CreateFileVersion(ctx context.Context, version *models.FileVersion) error
	GetFileVersions(ctx context.Context, fileID int) ([]*models.FileVersion, error)
	GetCurrentVersion(ctx context.Context, fileID int) (*models.FileVersion, error)
	SetCurrentVersion(ctx context.Context, fileID, versionID int) error

	CreateFolder(ctx context.Context, folder *models.Folder) error
	GetFolderByID(ctx context.Context, collegeID, folderID int) (*models.Folder, error)
	ListFolders(ctx context.Context, collegeID int, parentID *int) ([]*models.Folder, error)
	UpdateFolder(ctx context.Context, folder *models.Folder) error
	DeleteFolder(ctx context.Context, collegeID, folderID int) error
	GetFolderPath(ctx context.Context, folderID int) (string, error)

	SearchFiles(ctx context.Context, collegeID int, query string, category *string, limit, offset int) ([]*models.FileWithVersion, error)
	GetFilesByTags(ctx context.Context, collegeID int, tags []string, limit, offset int) ([]*models.FileWithVersion, error)
}

type fileRepository struct {
	db *DB
}

func NewFileRepository(db *DB) FileRepository {
	return &fileRepository{db: db}
}

func (r *fileRepository) CreateFile(ctx context.Context, file *models.File) error {
	query := `
		INSERT INTO files (college_id, name, description, category, folder_id, uploaded_by, is_public, tags)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`

	tagsJSON, err := json.Marshal(file.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	err = r.db.Pool.QueryRow(ctx, query,
		file.CollegeID, file.Name, file.Description, file.Category,
		file.FolderID, file.UploadedBy, file.IsPublic, tagsJSON,
	).Scan(&file.ID, &file.CreatedAt, &file.UpdatedAt)

	return err
}

func (r *fileRepository) GetFileByID(ctx context.Context, collegeID, fileID int) (*models.FileWithVersion, error) {
	query := `
		SELECT f.*, fv.*, fol.*
		FROM files f
		LEFT JOIN file_versions fv ON f.current_version_id = fv.id
		LEFT JOIN folders fol ON f.folder_id = fol.id
		WHERE f.id = $1 AND f.college_id = $2`

	var file models.FileWithVersion
	err := pgxscan.Get(ctx, r.db.Pool, &file, query, fileID, collegeID)
	if err != nil {
		return nil, err
	}

	// Get all versions
	versions, err := r.GetFileVersions(ctx, fileID)
	if err != nil {
		return nil, err
	}
	// Convert []*models.FileVersion to []models.FileVersion
	fileVersions := make([]models.FileVersion, len(versions))
	for i, v := range versions {
		fileVersions[i] = *v
	}
	file.Versions = fileVersions

	return &file, nil
}

func (r *fileRepository) ListFiles(ctx context.Context, collegeID int, folderID *int, category *string, limit, offset int) ([]*models.FileWithVersion, error) {
	query := `
		SELECT f.*, fv.*, fol.*
		FROM files f
		LEFT JOIN file_versions fv ON f.current_version_id = fv.id
		LEFT JOIN folders fol ON f.folder_id = fol.id
		WHERE f.college_id = $1`

	args := []any{collegeID}
	argCount := 1

	if folderID != nil {
		argCount++
		query += fmt.Sprintf(" AND f.folder_id = $%d", argCount)
		args = append(args, *folderID)
	}

	if category != nil {
		argCount++
		query += fmt.Sprintf(" AND f.category = $%d", argCount)
		args = append(args, *category)
	}

	query += fmt.Sprintf(" ORDER BY f.created_at DESC LIMIT $%d OFFSET $%d", argCount+1, argCount+2)
	args = append(args, limit, offset)

	var files []*models.FileWithVersion
	err := pgxscan.Select(ctx, r.db.Pool, &files, query, args...)
	return files, err
}

func (r *fileRepository) UpdateFile(ctx context.Context, file *models.File) error {
	query := `
		UPDATE files
		SET name = $1, description = $2, category = $3, folder_id = $4,
		    is_public = $5, tags = $6, updated_at = NOW()
		WHERE id = $7 AND college_id = $8`

	tagsJSON, err := json.Marshal(file.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	_, err = r.db.Pool.Exec(ctx, query,
		file.Name, file.Description, file.Category, file.FolderID,
		file.IsPublic, tagsJSON, file.ID, file.CollegeID)

	return err
}

func (r *fileRepository) DeleteFile(ctx context.Context, collegeID, fileID int) error {
	query := `DELETE FROM files WHERE id = $1 AND college_id = $2`
	_, err := r.db.Pool.Exec(ctx, query, fileID, collegeID)
	return err
}

func (r *fileRepository) CreateFileVersion(ctx context.Context, version *models.FileVersion) error {
	query := `
		INSERT INTO file_versions (file_id, version, object_key, filename, size, content_type, hash, uploaded_by, comment, is_current)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at`

	err := r.db.Pool.QueryRow(ctx, query,
		version.FileID, version.Version, version.ObjectKey, version.Filename,
		version.Size, version.ContentType, version.Hash, version.UploadedBy,
		version.Comment, version.IsCurrent,
	).Scan(&version.ID, &version.CreatedAt)

	return err
}

func (r *fileRepository) GetFileVersions(ctx context.Context, fileID int) ([]*models.FileVersion, error) {
	query := `
		SELECT * FROM file_versions
		WHERE file_id = $1
		ORDER BY version DESC`

	var versions []*models.FileVersion
	err := pgxscan.Select(ctx, r.db.Pool, &versions, query, fileID)
	return versions, err
}

func (r *fileRepository) GetCurrentVersion(ctx context.Context, fileID int) (*models.FileVersion, error) {
	query := `
		SELECT fv.* FROM file_versions fv
		JOIN files f ON fv.id = f.current_version_id
		WHERE f.id = $1`

	var version models.FileVersion
	err := pgxscan.Get(ctx, r.db.Pool, &version, query, fileID)
	return &version, err
}

func (r *fileRepository) SetCurrentVersion(ctx context.Context, fileID, versionID int) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Update all versions for this file to not current
	_, err = tx.Exec(ctx, "UPDATE file_versions SET is_current = false WHERE file_id = $1", fileID)
	if err != nil {
		return err
	}

	// Set the specified version as current
	_, err = tx.Exec(ctx, "UPDATE file_versions SET is_current = true WHERE id = $1 AND file_id = $2", versionID, fileID)
	if err != nil {
		return err
	}

	// Update the file's current_version_id
	_, err = tx.Exec(ctx, "UPDATE files SET current_version_id = $1 WHERE id = $2", versionID, fileID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *fileRepository) CreateFolder(ctx context.Context, folder *models.Folder) error {
	query := `
		INSERT INTO folders (college_id, name, parent_id, path, created_by)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	err := r.db.Pool.QueryRow(ctx, query,
		folder.CollegeID, folder.Name, folder.ParentID, folder.Path, folder.CreatedBy,
	).Scan(&folder.ID, &folder.CreatedAt, &folder.UpdatedAt)

	return err
}

func (r *fileRepository) GetFolderByID(ctx context.Context, collegeID, folderID int) (*models.Folder, error) {
	query := `SELECT * FROM folders WHERE id = $1 AND college_id = $2`

	var folder models.Folder
	err := pgxscan.Get(ctx, r.db.Pool, &folder, query, folderID, collegeID)
	return &folder, err
}

func (r *fileRepository) ListFolders(ctx context.Context, collegeID int, parentID *int) ([]*models.Folder, error) {
	query := `SELECT * FROM folders WHERE college_id = $1`
	args := []any{collegeID}

	if parentID != nil {
		query += " AND parent_id = $2"
		args = append(args, *parentID)
	}

	query += " ORDER BY name"

	var folders []*models.Folder
	err := pgxscan.Select(ctx, r.db.Pool, &folders, query, args...)
	return folders, err
}

func (r *fileRepository) UpdateFolder(ctx context.Context, folder *models.Folder) error {
	query := `
		UPDATE folders
		SET name = $1, parent_id = $2, path = $3, updated_at = NOW()
		WHERE id = $4 AND college_id = $5`

	_, err := r.db.Pool.Exec(ctx, query,
		folder.Name, folder.ParentID, folder.Path, folder.ID, folder.CollegeID)

	return err
}

func (r *fileRepository) DeleteFolder(ctx context.Context, collegeID, folderID int) error {
	query := `DELETE FROM folders WHERE id = $1 AND college_id = $2`
	_, err := r.db.Pool.Exec(ctx, query, folderID, collegeID)
	return err
}

func (r *fileRepository) GetFolderPath(ctx context.Context, folderID int) (string, error) {
	query := `SELECT path FROM folders WHERE id = $1`

	var path string
	err := r.db.Pool.QueryRow(ctx, query, folderID).Scan(&path)
	return path, err
}

func (r *fileRepository) SearchFiles(ctx context.Context, collegeID int, query string, category *string, limit, offset int) ([]*models.FileWithVersion, error) {
	sqlQuery := `
		SELECT f.*, fv.*, fol.*
		FROM files f
		LEFT JOIN file_versions fv ON f.current_version_id = fv.id
		LEFT JOIN folders fol ON f.folder_id = fol.id
		WHERE f.college_id = $1 AND (
			f.name ILIKE $2 OR
			f.description ILIKE $2 OR
			fv.filename ILIKE $2
		)`

	args := []any{collegeID, "%" + query + "%"}
	argCount := 2

	if category != nil {
		argCount++
		sqlQuery += fmt.Sprintf(" AND f.category = $%d", argCount)
		args = append(args, *category)
	}

	sqlQuery += fmt.Sprintf(" ORDER BY f.created_at DESC LIMIT $%d OFFSET $%d", argCount+1, argCount+2)
	args = append(args, limit, offset)

	var files []*models.FileWithVersion
	err := pgxscan.Select(ctx, r.db.Pool, &files, sqlQuery, args...)
	return files, err
}

func (r *fileRepository) GetFilesByTags(ctx context.Context, collegeID int, tags []string, limit, offset int) ([]*models.FileWithVersion, error) {
	query := `
		SELECT f.*, fv.*, fol.*
		FROM files f
		LEFT JOIN file_versions fv ON f.current_version_id = fv.id
		LEFT JOIN folders fol ON f.folder_id = fol.id
		WHERE f.college_id = $1 AND f.tags ?| $2
		ORDER BY f.created_at DESC
		LIMIT $3 OFFSET $4`

	var files []*models.FileWithVersion
	err := pgxscan.Select(ctx, r.db.Pool, &files, query, collegeID, tags, limit, offset)
	return files, err
}
