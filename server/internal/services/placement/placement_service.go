package placement

import (
	"context"
	"errors"

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
	TotalPlacements int     `json:"total_placements"`
	TotalStudents   int     `json:"total_students"`
	AveragePackage  float64 `json:"average_package"`
	HighestPackage  float64 `json:"highest_package"`
	LowestPackage   float64 `json:"lowest_package"`
	PlacementRate   float64 `json:"placement_rate"`
	OpenCount       int     `json:"open_count"`
	ClosedCount     int     `json:"closed_count"`
	InProgressCount int     `json:"in_progress_count"`
	CompletedCount  int     `json:"completed_count"`
	CancelledCount  int     `json:"cancelled_count"`
	UniqueCompanies int     `json:"unique_companies"`
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
	if placement.CollegeID == 0 {
		return errors.New("college ID is required")
	}
	if placement.CompanyName == "" {
		return errors.New("company name is required")
	}
	if placement.JobTitle == "" {
		return errors.New("job title is required")
	}
	if placement.Status == "" {
		placement.Status = models.PlacementStatusOpen
	}

	validStatuses := map[string]bool{
		models.PlacementStatusOpen:       true,
		models.PlacementStatusClosed:     true,
		models.PlacementStatusInProgress: true,
		models.PlacementStatusCompleted:  true,
		models.PlacementStatusCancelled:  true,
	}
	if !validStatuses[placement.Status] {
		return errors.New("invalid status. must be one of: open, closed, in_progress, completed, cancelled")
	}

	if placement.SalaryCurrency == "" {
		placement.SalaryCurrency = "USD"
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

	// Validate status
	validStatuses := map[string]bool{
		models.PlacementStatusOpen:       true,
		models.PlacementStatusClosed:     true,
		models.PlacementStatusInProgress: true,
		models.PlacementStatusCompleted:  true,
		models.PlacementStatusCancelled:  true,
	}
	if !validStatuses[placement.Status] {
		return errors.New("invalid status. must be one of: open, closed, in_progress, completed, cancelled")
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

	placements, err := s.repo.FindPlacementsByCollege(ctx, collegeID, 10000, 0)
	if err != nil {
		return nil, err
	}

	stats := &PlacementStats{
		TotalPlacements: len(placements),
		LowestPackage:   999999999,
	}

	if len(placements) == 0 {
		return stats, nil
	}

	uniqueCompanies := make(map[string]bool)
	var totalPackage float64
	var packageCount int

	for _, placement := range placements {
		uniqueCompanies[placement.CompanyName] = true

		// Calculate average package from salary range
		if placement.SalaryRangeMax != nil && *placement.SalaryRangeMax > 0 {
			avgPkg := *placement.SalaryRangeMax
			if placement.SalaryRangeMin != nil && *placement.SalaryRangeMin > 0 {
				avgPkg = (*placement.SalaryRangeMin + *placement.SalaryRangeMax) / 2
			}
			totalPackage += avgPkg
			packageCount++

			if avgPkg > stats.HighestPackage {
				stats.HighestPackage = avgPkg
			}
			if avgPkg < stats.LowestPackage {
				stats.LowestPackage = avgPkg
			}
		}

		// Count by status
		switch placement.Status {
		case models.PlacementStatusOpen:
			stats.OpenCount++
		case models.PlacementStatusClosed:
			stats.ClosedCount++
		case models.PlacementStatusInProgress:
			stats.InProgressCount++
		case models.PlacementStatusCompleted:
			stats.CompletedCount++
		case models.PlacementStatusCancelled:
			stats.CancelledCount++
		}
	}

	stats.UniqueCompanies = len(uniqueCompanies)
	if packageCount > 0 {
		stats.AveragePackage = totalPackage / float64(packageCount)
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

	placements, err := s.repo.FindPlacementsByCollege(ctx, collegeID, 10000, 0)
	if err != nil {
		return nil, err
	}

	companyMap := make(map[string]*CompanyStats)
	companyTotals := make(map[string]float64)
	companyCounts := make(map[string]int)

	for _, placement := range placements {
		stats, exists := companyMap[placement.CompanyName]
		if !exists {
			var initialPkg float64 = 0
			if placement.SalaryRangeMax != nil {
				initialPkg = *placement.SalaryRangeMax
			}
			stats = &CompanyStats{
				CompanyName:    placement.CompanyName,
				LowestPackage:  initialPkg,
				HighestPackage: initialPkg,
			}
			companyMap[placement.CompanyName] = stats
		}

		stats.TotalPlacements++

		if placement.SalaryRangeMax != nil && *placement.SalaryRangeMax > 0 {
			avgPkg := *placement.SalaryRangeMax
			if placement.SalaryRangeMin != nil && *placement.SalaryRangeMin > 0 {
				avgPkg = (*placement.SalaryRangeMin + *placement.SalaryRangeMax) / 2
			}
			companyTotals[placement.CompanyName] += avgPkg
			companyCounts[placement.CompanyName]++

			if avgPkg > stats.HighestPackage {
				stats.HighestPackage = avgPkg
			}
			if avgPkg < stats.LowestPackage {
				stats.LowestPackage = avgPkg
			}
		}
	}

	for companyName, stats := range companyMap {
		if companyCounts[companyName] > 0 {
			stats.AveragePackage = companyTotals[companyName] / float64(companyCounts[companyName])
		}
	}

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
