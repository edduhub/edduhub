package models

import (
	"time"
)

type FileVersion struct {
	ID          int       `json:"id" db:"id"`
	FileID      int       `json:"file_id" db:"file_id"`
	Version     int       `json:"version" db:"version"`
	ObjectKey   string    `json:"object_key" db:"object_key"`
	Filename    string    `json:"filename" db:"filename"`
	Size        int64     `json:"size" db:"size"`
	ContentType string    `json:"content_type" db:"content_type"`
	Hash        string    `json:"hash" db:"hash"` // SHA256 hash for integrity checking
	UploadedBy  int       `json:"uploaded_by" db:"uploaded_by"`
	Comment     string    `json:"comment" db:"comment"`
	IsCurrent   bool      `json:"is_current" db:"is_current"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type File struct {
	ID          int       `json:"id" db:"id"`
	CollegeID   int       `json:"college_id" db:"college_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Category    string    `json:"category" db:"category"` // assignment, lecture, profile, document, etc.
	FolderID    *int      `json:"folder_id" db:"folder_id"`
	CurrentVersionID int  `json:"current_version_id" db:"current_version_id"`
	UploadedBy  int       `json:"uploaded_by" db:"uploaded_by"`
	IsPublic    bool      `json:"is_public" db:"is_public"`
	Tags        []string  `json:"tags" db:"tags"` // JSON array of tags
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type Folder struct {
	ID        int       `json:"id" db:"id"`
	CollegeID int       `json:"college_id" db:"college_id"`
	Name      string    `json:"name" db:"name"`
	ParentID  *int      `json:"parent_id" db:"parent_id"`
	Path      string    `json:"path" db:"path"` // Full path like "/course1/assignments/week1"
	CreatedBy int       `json:"created_by" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type FileWithVersion struct {
	File
	CurrentVersion *FileVersion `json:"current_version"`
	Versions       []FileVersion `json:"versions,omitempty"`
	Folder         *Folder       `json:"folder,omitempty"`
}