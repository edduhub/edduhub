package course_material

import (
	"context"
	"fmt"
	"time"

	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"

	"github.com/go-playground/validator/v10"
)

type CourseMaterialService interface {
	// Module operations
	CreateModule(ctx context.Context, courseID, collegeID, userID int, req *models.CreateCourseModuleRequest) (*models.CourseModule, error)
	GetModule(ctx context.Context, collegeID, moduleID int) (*models.CourseModule, error)
	ListModules(ctx context.Context, collegeID, courseID int) ([]*models.CourseModule, error)
	UpdateModule(ctx context.Context, collegeID, moduleID int, req *models.UpdateCourseModuleRequest) error
	DeleteModule(ctx context.Context, collegeID, moduleID int) error

	// Material operations
	CreateMaterial(ctx context.Context, courseID, collegeID, userID int, req *models.CreateCourseMaterialRequest) (*models.CourseMaterial, error)
	GetMaterial(ctx context.Context, collegeID, materialID int) (*models.CourseMaterialWithDetails, error)
	ListMaterials(ctx context.Context, collegeID, courseID int, moduleID *int, onlyPublished bool) ([]*models.CourseMaterialWithDetails, error)
	UpdateMaterial(ctx context.Context, collegeID, materialID int, req *models.UpdateCourseMaterialRequest) error
	DeleteMaterial(ctx context.Context, collegeID, materialID int) error
	PublishMaterial(ctx context.Context, collegeID, materialID int) error
	UnpublishMaterial(ctx context.Context, collegeID, materialID int) error

	// Access tracking
	LogMaterialAccess(ctx context.Context, materialID, studentID int, durationSeconds int, completed bool) error
	GetMaterialAccessStats(ctx context.Context, materialID int) (map[string]interface{}, error)
	GetStudentProgress(ctx context.Context, courseID, studentID int) (map[string]interface{}, error)
}

type courseMaterialService struct {
	courseRepo    repository.CourseRepository
	materialRepo  repository.CourseMaterialRepository
	fileRepo      repository.FileRepository
	studentRepo   repository.StudentRepository
	validate      *validator.Validate
}

func NewCourseMaterialService(
	courseRepo repository.CourseRepository,
	materialRepo repository.CourseMaterialRepository,
	fileRepo repository.FileRepository,
	studentRepo repository.StudentRepository,
) CourseMaterialService {
	return &courseMaterialService{
		courseRepo:   courseRepo,
		materialRepo: materialRepo,
		fileRepo:     fileRepo,
		studentRepo:  studentRepo,
		validate:     validator.New(),
	}
}

// CreateModule creates a new course module
func (s *courseMaterialService) CreateModule(ctx context.Context, courseID, collegeID, userID int, req *models.CreateCourseModuleRequest) (*models.CourseModule, error) {
	if err := s.validate.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Verify course exists and belongs to college
	course, err := s.courseRepo.GetCourseByID(ctx, collegeID, courseID)
	if err != nil {
		return nil, fmt.Errorf("course not found: %w", err)
	}
	if course == nil {
		return nil, fmt.Errorf("course %d not found in college %d", courseID, collegeID)
	}

	module := &models.CourseModule{
		CourseID:    courseID,
		Title:       req.Title,
		Description: req.Description,
		Order:       req.Order,
		IsPublished: req.IsPublished,
		CollegeID:   collegeID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.materialRepo.CreateModule(ctx, module); err != nil {
		return nil, fmt.Errorf("failed to create module: %w", err)
	}

	return module, nil
}

// GetModule retrieves a course module
func (s *courseMaterialService) GetModule(ctx context.Context, collegeID, moduleID int) (*models.CourseModule, error) {
	module, err := s.materialRepo.GetModuleByID(ctx, collegeID, moduleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get module: %w", err)
	}
	if module == nil {
		return nil, fmt.Errorf("module not found")
	}
	return module, nil
}

// ListModules lists all modules for a course
func (s *courseMaterialService) ListModules(ctx context.Context, collegeID, courseID int) ([]*models.CourseModule, error) {
	modules, err := s.materialRepo.ListModulesByCourse(ctx, collegeID, courseID)
	if err != nil {
		return nil, fmt.Errorf("failed to list modules: %w", err)
	}
	return modules, nil
}

// UpdateModule updates a course module
func (s *courseMaterialService) UpdateModule(ctx context.Context, collegeID, moduleID int, req *models.UpdateCourseModuleRequest) error {
	if err := s.validate.Struct(req); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Get existing module
	module, err := s.materialRepo.GetModuleByID(ctx, collegeID, moduleID)
	if err != nil {
		return fmt.Errorf("module not found: %w", err)
	}
	if module == nil {
		return fmt.Errorf("module not found")
	}

	// Update fields
	if req.Title != nil {
		module.Title = *req.Title
	}
	if req.Description != nil {
		module.Description = *req.Description
	}
	if req.Order != nil {
		module.Order = *req.Order
	}
	if req.IsPublished != nil {
		module.IsPublished = *req.IsPublished
	}
	module.UpdatedAt = time.Now()

	if err := s.materialRepo.UpdateModule(ctx, module); err != nil {
		return fmt.Errorf("failed to update module: %w", err)
	}

	return nil
}

// DeleteModule deletes a course module
func (s *courseMaterialService) DeleteModule(ctx context.Context, collegeID, moduleID int) error {
	// Check if module has materials
	materials, err := s.materialRepo.ListMaterialsByModule(ctx, collegeID, moduleID)
	if err != nil {
		return fmt.Errorf("failed to check module materials: %w", err)
	}
	if len(materials) > 0 {
		return fmt.Errorf("cannot delete module with %d materials", len(materials))
	}

	if err := s.materialRepo.DeleteModule(ctx, collegeID, moduleID); err != nil {
		return fmt.Errorf("failed to delete module: %w", err)
	}

	return nil
}

// CreateMaterial creates a new course material
func (s *courseMaterialService) CreateMaterial(ctx context.Context, courseID, collegeID, userID int, req *models.CreateCourseMaterialRequest) (*models.CourseMaterial, error) {
	if err := s.validate.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Verify course exists
	course, err := s.courseRepo.GetCourseByID(ctx, collegeID, courseID)
	if err != nil || course == nil {
		return nil, fmt.Errorf("course not found")
	}

	// Verify module if specified
	if req.ModuleID != nil {
		module, err := s.materialRepo.GetModuleByID(ctx, collegeID, *req.ModuleID)
		if err != nil || module == nil {
			return nil, fmt.Errorf("module not found")
		}
		if module.CourseID != courseID {
			return nil, fmt.Errorf("module does not belong to course")
		}
	}

	// Verify file if specified
	if req.FileID != nil {
		file, err := s.fileRepo.GetFile(ctx, collegeID, *req.FileID)
		if err != nil || file == nil {
			return nil, fmt.Errorf("file not found")
		}
	}

	material := &models.CourseMaterial{
		CourseID:    courseID,
		Title:       req.Title,
		Description: req.Description,
		Type:        req.Type,
		FileID:      req.FileID,
		ExternalURL: req.ExternalURL,
		ModuleID:    req.ModuleID,
		Order:       req.Order,
		IsPublished: req.IsPublished,
		DueDate:     req.DueDate,
		UploadedBy:  userID,
		CollegeID:   collegeID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if material.IsPublished {
		now := time.Now()
		material.PublishedAt = &now
	}

	if err := s.materialRepo.CreateMaterial(ctx, material); err != nil {
		return nil, fmt.Errorf("failed to create material: %w", err)
	}

	return material, nil
}

// GetMaterial retrieves a course material with details
func (s *courseMaterialService) GetMaterial(ctx context.Context, collegeID, materialID int) (*models.CourseMaterialWithDetails, error) {
	material, err := s.materialRepo.GetMaterialWithDetails(ctx, collegeID, materialID)
	if err != nil {
		return nil, fmt.Errorf("failed to get material: %w", err)
	}
	if material == nil {
		return nil, fmt.Errorf("material not found")
	}
	return material, nil
}

// ListMaterials lists all materials for a course
func (s *courseMaterialService) ListMaterials(ctx context.Context, collegeID, courseID int, moduleID *int, onlyPublished bool) ([]*models.CourseMaterialWithDetails, error) {
	var materials []*models.CourseMaterialWithDetails
	var err error

	if moduleID != nil {
		materials, err = s.materialRepo.ListMaterialsByModuleWithDetails(ctx, collegeID, *moduleID, onlyPublished)
	} else {
		materials, err = s.materialRepo.ListMaterialsByCourseWithDetails(ctx, collegeID, courseID, onlyPublished)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list materials: %w", err)
	}

	return materials, nil
}

// UpdateMaterial updates a course material
func (s *courseMaterialService) UpdateMaterial(ctx context.Context, collegeID, materialID int, req *models.UpdateCourseMaterialRequest) error {
	if err := s.validate.Struct(req); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Get existing material
	existingMaterial, err := s.materialRepo.GetMaterialByID(ctx, collegeID, materialID)
	if err != nil || existingMaterial == nil {
		return fmt.Errorf("material not found")
	}

	// Update fields
	if req.Title != nil {
		existingMaterial.Title = *req.Title
	}
	if req.Description != nil {
		existingMaterial.Description = *req.Description
	}
	if req.Type != nil {
		existingMaterial.Type = *req.Type
	}
	if req.FileID != nil {
		existingMaterial.FileID = req.FileID
	}
	if req.ExternalURL != nil {
		existingMaterial.ExternalURL = req.ExternalURL
	}
	if req.ModuleID != nil {
		existingMaterial.ModuleID = req.ModuleID
	}
	if req.Order != nil {
		existingMaterial.Order = *req.Order
	}
	if req.IsPublished != nil {
		existingMaterial.IsPublished = *req.IsPublished
		if *req.IsPublished && existingMaterial.PublishedAt == nil {
			now := time.Now()
			existingMaterial.PublishedAt = &now
		}
	}
	if req.DueDate != nil {
		existingMaterial.DueDate = req.DueDate
	}
	existingMaterial.UpdatedAt = time.Now()

	if err := s.materialRepo.UpdateMaterial(ctx, existingMaterial); err != nil {
		return fmt.Errorf("failed to update material: %w", err)
	}

	return nil
}

// DeleteMaterial deletes a course material
func (s *courseMaterialService) DeleteMaterial(ctx context.Context, collegeID, materialID int) error {
	if err := s.materialRepo.DeleteMaterial(ctx, collegeID, materialID); err != nil {
		return fmt.Errorf("failed to delete material: %w", err)
	}
	return nil
}

// PublishMaterial publishes a material
func (s *courseMaterialService) PublishMaterial(ctx context.Context, collegeID, materialID int) error {
	material, err := s.materialRepo.GetMaterialByID(ctx, collegeID, materialID)
	if err != nil || material == nil {
		return fmt.Errorf("material not found")
	}

	material.IsPublished = true
	now := time.Now()
	material.PublishedAt = &now
	material.UpdatedAt = time.Now()

	if err := s.materialRepo.UpdateMaterial(ctx, material); err != nil {
		return fmt.Errorf("failed to publish material: %w", err)
	}

	return nil
}

// UnpublishMaterial unpublishes a material
func (s *courseMaterialService) UnpublishMaterial(ctx context.Context, collegeID, materialID int) error {
	material, err := s.materialRepo.GetMaterialByID(ctx, collegeID, materialID)
	if err != nil || material == nil {
		return fmt.Errorf("material not found")
	}

	material.IsPublished = false
	material.UpdatedAt = time.Now()

	if err := s.materialRepo.UpdateMaterial(ctx, material); err != nil {
		return fmt.Errorf("failed to unpublish material: %w", err)
	}

	return nil
}

// LogMaterialAccess logs student access to material
func (s *courseMaterialService) LogMaterialAccess(ctx context.Context, materialID, studentID int, durationSeconds int, completed bool) error {
	return s.materialRepo.LogAccess(ctx, materialID, studentID, durationSeconds, completed)
}

// GetMaterialAccessStats gets access statistics for a material
func (s *courseMaterialService) GetMaterialAccessStats(ctx context.Context, materialID int) (map[string]interface{}, error) {
	stats, err := s.materialRepo.GetAccessStats(ctx, materialID)
	if err != nil {
		return nil, fmt.Errorf("failed to get access stats: %w", err)
	}
	return stats, nil
}

// GetStudentProgress gets student's progress in a course
func (s *courseMaterialService) GetStudentProgress(ctx context.Context, courseID, studentID int) (map[string]interface{}, error) {
	progress, err := s.materialRepo.GetStudentProgress(ctx, courseID, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get student progress: %w", err)
	}
	return progress, nil
}
