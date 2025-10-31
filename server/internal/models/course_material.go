package models

import "time"

// CourseMaterial represents learning materials attached to a course
type CourseMaterial struct {
	ID          int       `json:"id" db:"id"`
	CourseID    int       `json:"courseId" db:"course_id"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	Type        string    `json:"type" db:"type"` // document, video, link, assignment, quiz
	FileID      *int      `json:"fileId,omitempty" db:"file_id"`
	FileURL     *string   `json:"fileUrl,omitempty" db:"file_url"`
	ExternalURL *string   `json:"externalUrl,omitempty" db:"external_url"`
	ModuleID    *int      `json:"moduleId,omitempty" db:"module_id"`
	Order       int       `json:"order" db:"display_order"`
	IsPublished bool      `json:"isPublished" db:"is_published"`
	PublishedAt *time.Time `json:"publishedAt,omitempty" db:"published_at"`
	DueDate     *time.Time `json:"dueDate,omitempty" db:"due_date"`
	UploadedBy  int       `json:"uploadedBy" db:"uploaded_by"`
	CollegeID   int       `json:"collegeId" db:"college_id"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}

// CourseModule represents a grouping/section of course materials
type CourseModule struct {
	ID          int       `json:"id" db:"id"`
	CourseID    int       `json:"courseId" db:"course_id"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	Order       int       `json:"order" db:"display_order"`
	IsPublished bool      `json:"isPublished" db:"is_published"`
	CollegeID   int       `json:"collegeId" db:"college_id"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}

// CourseMaterialWithDetails includes additional details for display
type CourseMaterialWithDetails struct {
	CourseMaterial
	FileName     *string `json:"fileName,omitempty"`
	FileSize     *int64  `json:"fileSize,omitempty"`
	ModuleName   *string `json:"moduleName,omitempty"`
	UploaderName string  `json:"uploaderName"`
}

// CreateCourseMaterialRequest represents the request to create a material
type CreateCourseMaterialRequest struct {
	Title       string     `json:"title" validate:"required,min=1,max=255"`
	Description string     `json:"description"`
	Type        string     `json:"type" validate:"required,oneof=document video link assignment quiz"`
	FileID      *int       `json:"fileId"`
	ExternalURL *string    `json:"externalUrl"`
	ModuleID    *int       `json:"moduleId"`
	Order       int        `json:"order"`
	IsPublished bool       `json:"isPublished"`
	DueDate     *time.Time `json:"dueDate"`
}

// UpdateCourseMaterialRequest represents the request to update a material
type UpdateCourseMaterialRequest struct {
	Title       *string    `json:"title,omitempty" validate:"omitempty,min=1,max=255"`
	Description *string    `json:"description,omitempty"`
	Type        *string    `json:"type,omitempty" validate:"omitempty,oneof=document video link assignment quiz"`
	FileID      *int       `json:"fileId,omitempty"`
	ExternalURL *string    `json:"externalUrl,omitempty"`
	ModuleID    *int       `json:"moduleId,omitempty"`
	Order       *int       `json:"order,omitempty"`
	IsPublished *bool      `json:"isPublished,omitempty"`
	DueDate     *time.Time `json:"dueDate,omitempty"`
}

// CreateCourseModuleRequest represents the request to create a module
type CreateCourseModuleRequest struct {
	Title       string `json:"title" validate:"required,min=1,max=255"`
	Description string `json:"description"`
	Order       int    `json:"order"`
	IsPublished bool   `json:"isPublished"`
}

// UpdateCourseModuleRequest represents the request to update a module
type UpdateCourseModuleRequest struct {
	Title       *string `json:"title,omitempty" validate:"omitempty,min=1,max=255"`
	Description *string `json:"description,omitempty"`
	Order       *int    `json:"order,omitempty"`
	IsPublished *bool   `json:"isPublished,omitempty"`
}
