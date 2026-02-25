package repository

import (
	"context"
	"eduhub/server/internal/models"
	"errors"
	"fmt"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
)

// CourseMaterialRepository defines interface for course materials and modules database operations
type CourseMaterialRepository interface {
	// Module operations
	CreateModule(ctx context.Context, module *models.CourseModule) error
	GetModuleByID(ctx context.Context, collegeID, moduleID int) (*models.CourseModule, error)
	ListModulesByCourse(ctx context.Context, collegeID, courseID int) ([]*models.CourseModule, error)
	UpdateModule(ctx context.Context, module *models.CourseModule) error
	DeleteModule(ctx context.Context, collegeID, moduleID int) error

	// Material operations
	CreateMaterial(ctx context.Context, material *models.CourseMaterial) error
	GetMaterialByID(ctx context.Context, collegeID, materialID int) (*models.CourseMaterial, error)
	GetMaterialWithDetails(ctx context.Context, collegeID, materialID int) (*models.CourseMaterialWithDetails, error)
	ListMaterialsByCourse(ctx context.Context, collegeID, courseID int, onlyPublished bool) ([]*models.CourseMaterial, error)
	ListMaterialsByCourseWithDetails(ctx context.Context, collegeID, courseID int, onlyPublished bool) ([]*models.CourseMaterialWithDetails, error)
	ListMaterialsByModule(ctx context.Context, collegeID, moduleID int) ([]*models.CourseMaterial, error)
	ListMaterialsByModuleWithDetails(ctx context.Context, collegeID, moduleID int, onlyPublished bool) ([]*models.CourseMaterialWithDetails, error)
	UpdateMaterial(ctx context.Context, material *models.CourseMaterial) error
	DeleteMaterial(ctx context.Context, collegeID, materialID int) error

	// Access tracking operations
	LogAccess(ctx context.Context, materialID, studentID int, durationSeconds int, completed bool) error
	GetAccessStats(ctx context.Context, materialID int) (map[string]any, error)
	GetStudentProgress(ctx context.Context, courseID, studentID int) (map[string]any, error)
}

type courseMaterialRepository struct {
	DB *DB
}

func NewCourseMaterialRepository(db *DB) CourseMaterialRepository {
	return &courseMaterialRepository{DB: db}
}

// --- Module Methods ---

func (r *courseMaterialRepository) CreateModule(ctx context.Context, module *models.CourseModule) error {
	now := time.Now()
	module.CreatedAt = now
	module.UpdatedAt = now

	sql := `INSERT INTO course_modules (course_id, title, description, display_order, is_published, college_id, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING id`

	temp := struct {
		ID int `db:"id"`
	}{}

	err := pgxscan.Get(ctx, r.DB.Pool, &temp, sql,
		module.CourseID, module.Title, module.Description, module.Order,
		module.IsPublished, module.CollegeID, module.CreatedAt, module.UpdatedAt)

	if err != nil {
		return fmt.Errorf("CreateModule: failed to execute query or scan ID: %w", err)
	}
	module.ID = temp.ID
	return nil
}

func (r *courseMaterialRepository) GetModuleByID(ctx context.Context, collegeID, moduleID int) (*models.CourseModule, error) {
	module := &models.CourseModule{}
	sql := `SELECT id, course_id, title, description, display_order, is_published, college_id, created_at, updated_at
			FROM course_modules
			WHERE id = $1 AND college_id = $2`

	err := pgxscan.Get(ctx, r.DB.Pool, module, sql, moduleID, collegeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetModuleByID: module with ID %d not found for college ID %d", moduleID, collegeID)
		}
		return nil, fmt.Errorf("GetModuleByID: failed to execute query or scan: %w", err)
	}
	return module, nil
}

func (r *courseMaterialRepository) ListModulesByCourse(ctx context.Context, collegeID, courseID int) ([]*models.CourseModule, error) {
	modules := []*models.CourseModule{}
	sql := `SELECT id, course_id, title, description, display_order, is_published, college_id, created_at, updated_at
			FROM course_modules
			WHERE course_id = $1 AND college_id = $2
			ORDER BY display_order ASC, created_at ASC`

	err := pgxscan.Select(ctx, r.DB.Pool, &modules, sql, courseID, collegeID)
	if err != nil {
		return nil, fmt.Errorf("ListModulesByCourse: failed to execute query or scan: %w", err)
	}
	return modules, nil
}

func (r *courseMaterialRepository) UpdateModule(ctx context.Context, module *models.CourseModule) error {
	module.UpdatedAt = time.Now()

	sql := `UPDATE course_modules
			SET title = $1, description = $2, display_order = $3, is_published = $4, updated_at = $5
			WHERE id = $6 AND college_id = $7`

	cmdTag, err := r.DB.Pool.Exec(ctx, sql,
		module.Title, module.Description, module.Order,
		module.IsPublished, module.UpdatedAt, module.ID, module.CollegeID)

	if err != nil {
		return fmt.Errorf("UpdateModule: failed to execute query: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateModule: no module found with ID %d for college ID %d", module.ID, module.CollegeID)
	}
	return nil
}

func (r *courseMaterialRepository) DeleteModule(ctx context.Context, collegeID, moduleID int) error {
	sql := `DELETE FROM course_modules WHERE id = $1 AND college_id = $2`
	cmdTag, err := r.DB.Pool.Exec(ctx, sql, moduleID, collegeID)
	if err != nil {
		return fmt.Errorf("DeleteModule: failed to execute query: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("DeleteModule: no module found with ID %d for college ID %d", moduleID, collegeID)
	}
	return nil
}

// --- Material Methods ---

func (r *courseMaterialRepository) CreateMaterial(ctx context.Context, material *models.CourseMaterial) error {
	now := time.Now()
	material.CreatedAt = now
	material.UpdatedAt = now

	sql := `INSERT INTO course_materials
			(course_id, title, description, type, file_id, external_url, module_id,
			 display_order, is_published, published_at, due_date, uploaded_by, college_id,
			 created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
			RETURNING id`

	temp := struct {
		ID int `db:"id"`
	}{}

	err := pgxscan.Get(ctx, r.DB.Pool, &temp, sql,
		material.CourseID, material.Title, material.Description, material.Type,
		material.FileID, material.ExternalURL, material.ModuleID, material.Order,
		material.IsPublished, material.PublishedAt, material.DueDate, material.UploadedBy,
		material.CollegeID, material.CreatedAt, material.UpdatedAt)

	if err != nil {
		return fmt.Errorf("CreateMaterial: failed to execute query or scan ID: %w", err)
	}
	material.ID = temp.ID
	return nil
}

func (r *courseMaterialRepository) GetMaterialByID(ctx context.Context, collegeID, materialID int) (*models.CourseMaterial, error) {
	material := &models.CourseMaterial{}
	sql := `SELECT id, course_id, title, description, type, file_id, external_url, module_id,
			display_order, is_published, published_at, due_date, uploaded_by, college_id,
			created_at, updated_at
			FROM course_materials
			WHERE id = $1 AND college_id = $2`

	err := pgxscan.Get(ctx, r.DB.Pool, material, sql, materialID, collegeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetMaterialByID: material with ID %d not found for college ID %d", materialID, collegeID)
		}
		return nil, fmt.Errorf("GetMaterialByID: failed to execute query or scan: %w", err)
	}
	return material, nil
}

func (r *courseMaterialRepository) GetMaterialWithDetails(ctx context.Context, collegeID, materialID int) (*models.CourseMaterialWithDetails, error) {
	material := &models.CourseMaterialWithDetails{}
	sql := `SELECT
				cm.id, cm.course_id, cm.title, cm.description, cm.type, cm.file_id,
				cm.external_url, cm.module_id, cm.display_order, cm.is_published,
				cm.published_at, cm.due_date, cm.uploaded_by, cm.college_id,
				cm.created_at, cm.updated_at,
				f.filename, f.file_path, f.file_size, f.mime_type,
				mod.title as module_title
			FROM course_materials cm
			LEFT JOIN files f ON cm.file_id = f.id
			LEFT JOIN course_modules mod ON cm.module_id = mod.id
			WHERE cm.id = $1 AND cm.college_id = $2`

	err := pgxscan.Get(ctx, r.DB.Pool, material, sql, materialID, collegeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetMaterialWithDetails: material with ID %d not found for college ID %d", materialID, collegeID)
		}
		return nil, fmt.Errorf("GetMaterialWithDetails: failed to execute query or scan: %w", err)
	}
	return material, nil
}

func (r *courseMaterialRepository) ListMaterialsByCourse(ctx context.Context, collegeID, courseID int, onlyPublished bool) ([]*models.CourseMaterial, error) {
	materials := []*models.CourseMaterial{}

	sql := `SELECT id, course_id, title, description, type, file_id, external_url, module_id,
			display_order, is_published, published_at, due_date, uploaded_by, college_id,
			created_at, updated_at
			FROM course_materials
			WHERE course_id = $1 AND college_id = $2`

	if onlyPublished {
		sql += ` AND is_published = true`
	}

	sql += ` ORDER BY module_id NULLS FIRST, display_order ASC, created_at ASC`

	err := pgxscan.Select(ctx, r.DB.Pool, &materials, sql, courseID, collegeID)
	if err != nil {
		return nil, fmt.Errorf("ListMaterialsByCourse: failed to execute query or scan: %w", err)
	}
	return materials, nil
}

func (r *courseMaterialRepository) ListMaterialsByCourseWithDetails(ctx context.Context, collegeID, courseID int, onlyPublished bool) ([]*models.CourseMaterialWithDetails, error) {
	materials := []*models.CourseMaterialWithDetails{}

	sql := `SELECT
				cm.id, cm.course_id, cm.title, cm.description, cm.type, cm.file_id,
				cm.external_url, cm.module_id, cm.display_order, cm.is_published,
				cm.published_at, cm.due_date, cm.uploaded_by, cm.college_id,
				cm.created_at, cm.updated_at,
				f.filename, f.file_path, f.file_size, f.mime_type,
				mod.title as module_title
			FROM course_materials cm
			LEFT JOIN files f ON cm.file_id = f.id
			LEFT JOIN course_modules mod ON cm.module_id = mod.id
			WHERE cm.course_id = $1 AND cm.college_id = $2`

	if onlyPublished {
		sql += ` AND cm.is_published = true`
	}

	sql += ` ORDER BY cm.module_id NULLS FIRST, cm.display_order ASC, cm.created_at ASC`

	err := pgxscan.Select(ctx, r.DB.Pool, &materials, sql, courseID, collegeID)
	if err != nil {
		return nil, fmt.Errorf("ListMaterialsByCourseWithDetails: failed to execute query or scan: %w", err)
	}
	return materials, nil
}

func (r *courseMaterialRepository) ListMaterialsByModule(ctx context.Context, collegeID, moduleID int) ([]*models.CourseMaterial, error) {
	materials := []*models.CourseMaterial{}
	sql := `SELECT id, course_id, title, description, type, file_id, external_url, module_id,
			display_order, is_published, published_at, due_date, uploaded_by, college_id,
			created_at, updated_at
			FROM course_materials
			WHERE module_id = $1 AND college_id = $2
			ORDER BY display_order ASC, created_at ASC`

	err := pgxscan.Select(ctx, r.DB.Pool, &materials, sql, moduleID, collegeID)
	if err != nil {
		return nil, fmt.Errorf("ListMaterialsByModule: failed to execute query or scan: %w", err)
	}
	return materials, nil
}

func (r *courseMaterialRepository) ListMaterialsByModuleWithDetails(ctx context.Context, collegeID, moduleID int, onlyPublished bool) ([]*models.CourseMaterialWithDetails, error) {
	materials := []*models.CourseMaterialWithDetails{}

	sql := `SELECT
				cm.id, cm.course_id, cm.title, cm.description, cm.type, cm.file_id,
				cm.external_url, cm.module_id, cm.display_order, cm.is_published,
				cm.published_at, cm.due_date, cm.uploaded_by, cm.college_id,
				cm.created_at, cm.updated_at,
				f.filename, f.file_path, f.file_size, f.mime_type,
				mod.title as module_title
			FROM course_materials cm
			LEFT JOIN files f ON cm.file_id = f.id
			LEFT JOIN course_modules mod ON cm.module_id = mod.id
			WHERE cm.module_id = $1 AND cm.college_id = $2`

	if onlyPublished {
		sql += ` AND cm.is_published = true`
	}

	sql += ` ORDER BY cm.display_order ASC, cm.created_at ASC`

	err := pgxscan.Select(ctx, r.DB.Pool, &materials, sql, moduleID, collegeID)
	if err != nil {
		return nil, fmt.Errorf("ListMaterialsByModuleWithDetails: failed to execute query or scan: %w", err)
	}
	return materials, nil
}

func (r *courseMaterialRepository) UpdateMaterial(ctx context.Context, material *models.CourseMaterial) error {
	material.UpdatedAt = time.Now()

	sql := `UPDATE course_materials
			SET title = $1, description = $2, type = $3, file_id = $4, external_url = $5,
				module_id = $6, display_order = $7, is_published = $8, published_at = $9,
				due_date = $10, updated_at = $11
			WHERE id = $12 AND college_id = $13`

	cmdTag, err := r.DB.Pool.Exec(ctx, sql,
		material.Title, material.Description, material.Type, material.FileID,
		material.ExternalURL, material.ModuleID, material.Order, material.IsPublished,
		material.PublishedAt, material.DueDate, material.UpdatedAt,
		material.ID, material.CollegeID)

	if err != nil {
		return fmt.Errorf("UpdateMaterial: failed to execute query: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateMaterial: no material found with ID %d for college ID %d", material.ID, material.CollegeID)
	}
	return nil
}

func (r *courseMaterialRepository) DeleteMaterial(ctx context.Context, collegeID, materialID int) error {
	sql := `DELETE FROM course_materials WHERE id = $1 AND college_id = $2`
	cmdTag, err := r.DB.Pool.Exec(ctx, sql, materialID, collegeID)
	if err != nil {
		return fmt.Errorf("DeleteMaterial: failed to execute query: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("DeleteMaterial: no material found with ID %d for college ID %d", materialID, collegeID)
	}
	return nil
}

// --- Access Tracking Methods ---

func (r *courseMaterialRepository) LogAccess(ctx context.Context, materialID, studentID int, durationSeconds int, completed bool) error {
	sql := `INSERT INTO course_material_access
			(material_id, student_id, accessed_at, duration_seconds, completed)
			VALUES ($1, $2, NOW(), $3, $4)`

	_, err := r.DB.Pool.Exec(ctx, sql, materialID, studentID, durationSeconds, completed)
	if err != nil {
		return fmt.Errorf("LogAccess: failed to execute query: %w", err)
	}
	return nil
}

func (r *courseMaterialRepository) GetAccessStats(ctx context.Context, materialID int) (map[string]any, error) {
	sql := `SELECT
				COUNT(DISTINCT student_id) as unique_students,
				COUNT(*) as total_accesses,
				AVG(duration_seconds) as avg_duration,
				SUM(CASE WHEN completed THEN 1 ELSE 0 END) as completion_count
			FROM course_material_access
			WHERE material_id = $1`

	var stats struct {
		UniqueStudents  int     `db:"unique_students"`
		TotalAccesses   int     `db:"total_accesses"`
		AvgDuration     float64 `db:"avg_duration"`
		CompletionCount int     `db:"completion_count"`
	}

	err := pgxscan.Get(ctx, r.DB.Pool, &stats, sql, materialID)
	if err != nil {
		return nil, fmt.Errorf("GetAccessStats: failed to execute query or scan: %w", err)
	}

	result := map[string]any{
		"uniqueStudents":  stats.UniqueStudents,
		"totalAccesses":   stats.TotalAccesses,
		"avgDuration":     stats.AvgDuration,
		"completionCount": stats.CompletionCount,
	}

	return result, nil
}

func (r *courseMaterialRepository) GetStudentProgress(ctx context.Context, courseID, studentID int) (map[string]any, error) {
	sql := `SELECT
				COUNT(DISTINCT cm.id) as total_materials,
				COUNT(DISTINCT CASE WHEN cma.completed THEN cm.id END) as completed_materials,
				AVG(cma.duration_seconds) as avg_duration
			FROM course_materials cm
			LEFT JOIN course_material_access cma ON cm.id = cma.material_id AND cma.student_id = $2
			WHERE cm.course_id = $1 AND cm.is_published = true`

	var progress struct {
		TotalMaterials     int     `db:"total_materials"`
		CompletedMaterials int     `db:"completed_materials"`
		AvgDuration        float64 `db:"avg_duration"`
	}

	err := pgxscan.Get(ctx, r.DB.Pool, &progress, sql, courseID, studentID)
	if err != nil {
		return nil, fmt.Errorf("GetStudentProgress: failed to execute query or scan: %w", err)
	}

	completionPercentage := 0.0
	if progress.TotalMaterials > 0 {
		completionPercentage = (float64(progress.CompletedMaterials) / float64(progress.TotalMaterials)) * 100
	}

	result := map[string]any{
		"totalMaterials":       progress.TotalMaterials,
		"completedMaterials":   progress.CompletedMaterials,
		"completionPercentage": completionPercentage,
		"avgDuration":          progress.AvgDuration,
	}

	return result, nil
}
