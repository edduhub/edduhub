package placement

import (
	"context"
	"errors"
	"time"

	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"
)

type PlacementService interface {
	// Placement CRUD
	CreatePlacement(ctx context.Context, placement *models.Placement) error
	GetPlacement(ctx context.Context, collegeID, placementID int) (*models.Placement, error)
	UpdatePlacement(ctx context.Context, placement *models.Placement) error
	DeletePlacement(ctx context.Context, collegeID, placementID int) error

	// Find methods
	ListPlacementsByStudent(ctx context.Context, collegeID, studentID int, limit, offset uint64) ([]*models.Placement, error)
	ListPlacementsByCollege(ctx context.Context, collegeID int, limit, offset uint64) ([]*models.Placement, error)
	ListPlacementsByCompany(ctx context.Context, collegeID int, companyName string, limit, offset uint64) ([]*models.Placement, error)

	// Statistics
	GetPlacementStats(ctx context.Context, collegeID int) (*PlacementStats, error)
	GetCompanyStats(ctx context.Context, collegeID int) ([]*CompanyStats, error)
	GetStudentPlacementCount(ctx context.Context, collegeID, studentID int) (int, error)
}

// PlacementStats represents overall placement statistics
type PlacementStats struct {
	TotalPlacements   int     `json:"total_placements"`
	TotalStudents     int     `json:"total_students"`
	AveragePackage    float64 `json:"average_package"`
	HighestPackage    float64 `json:"highest_package"`
	LowestPackage     float64 `json:"lowest_package"`
	PlacementRate     float64 `json:"placement_rate"`
	OfferedCount      int     `json:"offered_count"`
	AcceptedCount     int     `json:"accepted_count"`
	RejectedCount     int     `json:"rejected_count"`
	OnHoldCount       int     `json:"on_hold_count"`
	UniqueCompanies   int     `json:"unique_companies"`
}

// CompanyStats represents statistics for a specific company
type CompanyStats struct {
	CompanyName     string  `json:"company_name"`
	TotalPlacements int     `json:"total_placements"`
	AveragePackage  float64 `json:"average_package"`
	HighestPackage  float64 `json:"highest_package"`
	LowestPackage   float64 `json:"lowest_package"`
}

type placementService struct {
	repo        repository.PlacementRepository
	studentRepo repository.StudentRepository
}

func NewPlacementService(
	repo repository.PlacementRepository,
	studentRepo repository.StudentRepository,
) PlacementService {
	return &placementService{
		repo:        repo,
		studentRepo: studentRepo,
	}
}

// ===========================
// Placement CRUD
// ===========================

func (s *placementService) CreatePlacement(ctx context.Context, placement *models.Placement) error {
	// Validation
	if placement.StudentID == 0 {
		return errors.New("student ID is required")
	}
	if placement.CollegeID == 0 {
		return errors.New("college ID is required")
	}
	if placement.CompanyName == "" {
		return errors.New("company name is required")
	}
	if placement.JobTitle == "" {
		return errors.New("job title is required")
	}
	if placement.Package < 0 {
		return errors.New("package must be non-negative")
	}
	if placement.PlacementDate.IsZero() {
		placement.PlacementDate = time.Now()
	}

	// Set default status if not provided
	if placement.Status == "" {
		placement.Status = "offered"
	}

	// Validate status
	validStatuses := map[string]bool{
		"offered":  true,
		"accepted": true,
		"rejected": true,
		"on-hold":  true,
	}
	if !validStatuses[placement.Status] {
		return errors.New("invalid status. must be one of: offered, accepted, rejected, on-hold")
	}

	// Verify student exists
	_, err := s.studentRepo.GetStudentByID(ctx, placement.CollegeID, placement.StudentID)
	if err != nil {
		return errors.New("student not found")
	}

	return s.repo.CreatePlacement(ctx, placement)
}

func (s *placementService) GetPlacement(ctx context.Context, collegeID, placementID int) (*models.Placement, error) {
	if collegeID == 0 || placementID == 0 {
		return nil, errors.New("invalid college ID or placement ID")
	}
	return s.repo.GetPlacementByID(ctx, collegeID, placementID)
}

func (s *placementService) UpdatePlacement(ctx context.Context, placement *models.Placement) error {
	if placement.ID == 0 || placement.CollegeID == 0 {
		return errors.New("invalid placement ID or college ID")
	}

	// Verify placement exists
	_, err := s.repo.GetPlacementByID(ctx, placement.CollegeID, placement.ID)
	if err != nil {
		return errors.New("placement not found")
	}

	// Validation
	if placement.CompanyName == "" {
		return errors.New("company name is required")
	}
	if placement.JobTitle == "" {
		return errors.New("job title is required")
	}
	if placement.Package < 0 {
		return errors.New("package must be non-negative")
	}

	// Validate status
	validStatuses := map[string]bool{
		"offered":  true,
		"accepted": true,
		"rejected": true,
		"on-hold":  true,
	}
	if !validStatuses[placement.Status] {
		return errors.New("invalid status. must be one of: offered, accepted, rejected, on-hold")
	}

	return s.repo.UpdatePlacement(ctx, placement)
}

func (s *placementService) DeletePlacement(ctx context.Context, collegeID, placementID int) error {
	if collegeID == 0 || placementID == 0 {
		return errors.New("invalid college ID or placement ID")
	}
	return s.repo.DeletePlacement(ctx, collegeID, placementID)
}

// ===========================
// Find Methods
// ===========================

func (s *placementService) ListPlacementsByStudent(ctx context.Context, collegeID, studentID int, limit, offset uint64) ([]*models.Placement, error) {
	if collegeID == 0 || studentID == 0 {
		return nil, errors.New("invalid college ID or student ID")
	}
	if limit == 0 {
		limit = 50
	}
	return s.repo.FindPlacementsByStudent(ctx, collegeID, studentID, limit, offset)
}

func (s *placementService) ListPlacementsByCollege(ctx context.Context, collegeID int, limit, offset uint64) ([]*models.Placement, error) {
	if collegeID == 0 {
		return nil, errors.New("college ID is required")
	}
	if limit == 0 {
		limit = 50
	}
	return s.repo.FindPlacementsByCollege(ctx, collegeID, limit, offset)
}

func (s *placementService) ListPlacementsByCompany(ctx context.Context, collegeID int, companyName string, limit, offset uint64) ([]*models.Placement, error) {
	if collegeID == 0 {
		return nil, errors.New("college ID is required")
	}
	if companyName == "" {
		return nil, errors.New("company name is required")
	}
	if limit == 0 {
		limit = 50
	}
	return s.repo.FindPlacementsByCompany(ctx, collegeID, companyName, limit, offset)
}

// ===========================
// Statistics
// ===========================

func (s *placementService) GetPlacementStats(ctx context.Context, collegeID int) (*PlacementStats, error) {
	if collegeID == 0 {
		return nil, errors.New("college ID is required")
	}

	// Get all placements (with a high limit to get all records)
	placements, err := s.repo.FindPlacementsByCollege(ctx, collegeID, 10000, 0)
	if err != nil {
		return nil, err
	}

	stats := &PlacementStats{
		TotalPlacements: len(placements),
		LowestPackage:   999999999, // Initialize with high value
	}

	if len(placements) == 0 {
		return stats, nil
	}

	uniqueStudents := make(map[int]bool)
	uniqueCompanies := make(map[string]bool)
	var totalPackage float64

	for _, placement := range placements {
		uniqueStudents[placement.StudentID] = true
		uniqueCompanies[placement.CompanyName] = true
		totalPackage += placement.Package

		// Track highest and lowest packages
		if placement.Package > stats.HighestPackage {
			stats.HighestPackage = placement.Package
		}
		if placement.Package < stats.LowestPackage {
			stats.LowestPackage = placement.Package
		}

		// Count by status
		switch placement.Status {
		case "offered":
			stats.OfferedCount++
		case "accepted":
			stats.AcceptedCount++
		case "rejected":
			stats.RejectedCount++
		case "on-hold":
			stats.OnHoldCount++
		}
	}

	stats.TotalStudents = len(uniqueStudents)
	stats.UniqueCompanies = len(uniqueCompanies)
	stats.AveragePackage = totalPackage / float64(stats.TotalPlacements)

	// Calculate placement rate (accepted placements / total students * 100)
	if stats.TotalStudents > 0 {
		stats.PlacementRate = float64(stats.AcceptedCount) / float64(stats.TotalStudents) * 100
	}

	if stats.LowestPackage == 999999999 {
		stats.LowestPackage = 0
	}

	return stats, nil
}

func (s *placementService) GetCompanyStats(ctx context.Context, collegeID int) ([]*CompanyStats, error) {
	if collegeID == 0 {
		return nil, errors.New("college ID is required")
	}

	// Get all placements
	placements, err := s.repo.FindPlacementsByCollege(ctx, collegeID, 10000, 0)
	if err != nil {
		return nil, err
	}

	// Group placements by company
	companyMap := make(map[string]*CompanyStats)

	for _, placement := range placements {
		if companyMap[placement.CompanyName] == nil {
			companyMap[placement.CompanyName] = &CompanyStats{
				CompanyName:    placement.CompanyName,
				LowestPackage:  999999999,
			}
		}

		stats := companyMap[placement.CompanyName]
		stats.TotalPlacements++

		if placement.Package > stats.HighestPackage {
			stats.HighestPackage = placement.Package
		}
		if placement.Package < stats.LowestPackage {
			stats.LowestPackage = placement.Package
		}
	}

	// Calculate averages
	for companyName, stats := range companyMap {
		companyPlacements, _ := s.repo.FindPlacementsByCompany(ctx, collegeID, companyName, 10000, 0)
		var total float64
		for _, p := range companyPlacements {
			total += p.Package
		}
		if len(companyPlacements) > 0 {
			stats.AveragePackage = total / float64(len(companyPlacements))
		}

		if stats.LowestPackage == 999999999 {
			stats.LowestPackage = 0
		}
	}

	// Convert map to slice
	result := make([]*CompanyStats, 0, len(companyMap))
	for _, stats := range companyMap {
		result = append(result, stats)
	}

	return result, nil
}

func (s *placementService) GetStudentPlacementCount(ctx context.Context, collegeID, studentID int) (int, error) {
	if collegeID == 0 || studentID == 0 {
		return 0, errors.New("invalid college ID or student ID")
	}
	return s.repo.CountPlacementsByStudent(ctx, collegeID, studentID)
}
