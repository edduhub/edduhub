package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time" // Needed for time.Now() and time.Time fields

	"eduhub/server/internal/models" // Your models package

	"eduhub/server/internal/repository/db"

	"github.com/jackc/pgx/v4"
)

// --- Updated models.Student struct (assuming these fields exist in your DB) ---

type StudentRepository interface {
	CreateStudent(ctx context.Context, student *models.Student) error
	GetStudentByRollNo(ctx context.Context, collegeID int, rollNo string) (*models.Student, error)
	GetStudentByID(ctx context.Context, collegeID int, studentID int) (*models.Student, error) // Note: studentID is the primary key 'id' here
	UpdateStudent(ctx context.Context, model *models.Student) error
	FreezeStudent(ctx context.Context, rollNo string) error   // Renamed param to match casing
	UnFreezeStudent(ctx context.Context, rollNo string) error // Renamed param to match casing
	FindByKratosID(ctx context.Context, kratosID string) (*models.Student, error)
	DeleteStudent(ctx context.Context, collegeID int, studentID int) error

	// Find methods with pagination
	FindAllStudentsByCollege(ctx context.Context, collegeID int, limit, offset uint64) ([]*models.Student, error)
	CountStudentsByCollege(ctx context.Context, collegeID int) (int, error)
	// Note: GetStudentByID signature is a bit unusual if ID is the primary key;
	// typically you only need the ID. Assuming you intend to filter by both ID and CollegeID.
}

// studentRepository now holds a direct reference to *DB
type studentRepository struct {
	DB *DB
	*db.Queries
}

// NewStudentRepository receives the *DB directly
func NewStudentRepository(database *DB) StudentRepository {
	return &studentRepository{
		DB: database,
		Queries: db.New(database.Pool),
	}
}

const studentTable = "students" // Define your table name

// CreateStudent inserts a new student record into the database.
func (s *studentRepository) CreateStudent(ctx context.Context, student *models.Student) error {
	// Set timestamps if they are zero-valued
	now := time.Now()
	if student.CreatedAt.IsZero() {
		student.CreatedAt = now
	}
	if student.UpdatedAt.IsZero() {
		student.UpdatedAt = now
	}

	// Use sqlc generated code
	params := db.CreateStudentParams{
		UserID:           int32(student.UserID),
		CollegeID:        int32(student.CollegeID),
		KratosIdentityID: student.KratosIdentityID,
		EnrollmentYear:   sql.NullInt32{Int32: int32(student.EnrollmentYear), Valid: student.EnrollmentYear > 0},
		RollNo:           student.RollNo,
		IsActive:         student.IsActive,
		CreatedAt:        student.CreatedAt,
		UpdatedAt:        student.UpdatedAt,
	}

	result, err := s.Queries.CreateStudent(ctx, params)
	if err != nil {
		return fmt.Errorf("CreateStudent: failed to execute query: %w", err)
	}

	// Update the student struct with the returned values
	student.StudentID = int(result.StudentID)
	student.UserID = int(result.UserID)
	student.CollegeID = int(result.CollegeID)
	student.KratosIdentityID = result.KratosIdentityID
	student.EnrollmentYear = int(result.EnrollmentYear.Int32)
	student.RollNo = result.RollNo
	student.IsActive = result.IsActive
	student.CreatedAt = result.CreatedAt
	student.UpdatedAt = result.UpdatedAt

	return nil // Success
}

// GetStudentByRollNo retrieves a student by their roll number.
func (s *studentRepository) GetStudentByRollNo(ctx context.Context, collegeID int, rollNo string) (*models.Student, error) {
	// Use sqlc generated code
	params := db.GetStudentByRollNoParams{
		RollNo:    rollNo,
		CollegeID: int32(collegeID),
	}

	result, err := s.Queries.GetStudentByRollNo(ctx, params)
	if err != nil {
		if err == pgx.ErrNoRows {
			// Return a specific error or nil, nil as per your error handling strategy
			return nil, fmt.Errorf("GetStudentByRollNo: student with rollNo %s not found in college %d", rollNo, collegeID)
		}
		// Any other error during execution or scanning
		return nil, fmt.Errorf("GetStudentByRollNo: failed to execute query or scan for college %d, rollNo %s: %w", collegeID, rollNo, err)
	}

	// Convert db.Student to models.Student
	student := &models.Student{
		StudentID:        int(result.StudentID),
		UserID:           int(result.UserID),
		CollegeID:        int(result.CollegeID),
		KratosIdentityID: result.KratosIdentityID,
		EnrollmentYear:   int(result.EnrollmentYear.Int32),
		RollNo:           result.RollNo,
		IsActive:         result.IsActive,
		CreatedAt:        result.CreatedAt,
		UpdatedAt:        result.UpdatedAt,
	}

	return student, nil // Success
}

// DeleteStudent removes a student record by its ID, scoped by collegeID.
func (s *studentRepository) DeleteStudent(ctx context.Context, collegeID int, studentID int) error {
	// Use sqlc generated code
	params := db.DeleteStudentParams{
		StudentID: int32(studentID),
		CollegeID: int32(collegeID),
	}

	err := s.Queries.DeleteStudent(ctx, params)
	if err != nil {
		// Consider foreign key constraint errors (e.g., if student has enrollments)
		return fmt.Errorf("DeleteStudent: failed to execute query: %w", err)
	}

	return nil
}

// FindAllStudentsByCollege retrieves a paginated list of students for a specific college.
func (s *studentRepository) FindAllStudentsByCollege(ctx context.Context, collegeID int, limit, offset uint64) ([]*models.Student, error) {
	// Use sqlc generated code
	params := db.FindAllStudentsByCollegeParams{
		CollegeID: int32(collegeID),
		Limit:     int32(limit),
		Offset:    int32(offset),
	}

	results, err := s.Queries.FindAllStudentsByCollege(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("FindAllStudentsByCollege: failed to execute query or scan: %w", err)
	}

	// Convert []db.Student to []*models.Student
	students := make([]*models.Student, len(results))
	for i, result := range results {
		students[i] = &models.Student{
			StudentID:        int(result.StudentID),
			UserID:           int(result.UserID),
			CollegeID:        int(result.CollegeID),
			KratosIdentityID: result.KratosIdentityID,
			EnrollmentYear:   int(result.EnrollmentYear.Int32),
			RollNo:           result.RollNo,
			IsActive:         result.IsActive,
			CreatedAt:        result.CreatedAt,
			UpdatedAt:        result.UpdatedAt,
		}
	}

	return students, nil
}

// CountStudentsByCollege counts the total number of students within a specific college.
func (s *studentRepository) CountStudentsByCollege(ctx context.Context, collegeID int) (int, error) {
	// Use sqlc generated code
	count, err := s.Queries.CountStudentsByCollege(ctx, int32(collegeID))
	if err != nil {
		return 0, fmt.Errorf("CountStudentsByCollege: failed to execute query or scan: %w", err)
	}

	return int(count), nil
}

// GetStudentByID retrieves a student by their ID, filtered by college ID.
func (s *studentRepository) GetStudentByID(ctx context.Context, collegeID int, studentID int) (*models.Student, error) {
	// Use sqlc generated code
	params := db.GetStudentByIDParams{
		StudentID: int32(studentID),
		CollegeID: int32(collegeID),
	}

	result, err := s.Queries.GetStudentByID(ctx, params)
	if err != nil {
		if err == pgx.ErrNoRows {
			// Return nil, nil for not found
			return nil, nil
		}
		// Any other error during execution or scanning
		return nil, fmt.Errorf("GetStudentByID: failed to execute query or scan: %w", err)
	}

	// Convert db.Student to models.Student
	student := &models.Student{
		StudentID:        int(result.StudentID),
		UserID:           int(result.UserID),
		CollegeID:        int(result.CollegeID),
		KratosIdentityID: result.KratosIdentityID,
		EnrollmentYear:   int(result.EnrollmentYear.Int32),
		RollNo:           result.RollNo,
		IsActive:         result.IsActive,
		CreatedAt:        result.CreatedAt,
		UpdatedAt:        result.UpdatedAt,
	}

	return student, nil // Success
}

// UpdateStudent updates an existing student record.
func (s *studentRepository) UpdateStudent(ctx context.Context, model *models.Student) error {
	// Update the UpdatedAt timestamp
	model.UpdatedAt = time.Now()

	// Use sqlc generated code
	params := db.UpdateStudentParams{
		UserID:           int32(model.UserID),
		CollegeID:        int32(model.CollegeID),
		KratosIdentityID: model.KratosIdentityID,
		EnrollmentYear:   sql.NullInt32{Int32: int32(model.EnrollmentYear), Valid: model.EnrollmentYear > 0},
		RollNo:           model.RollNo,
		IsActive:         model.IsActive,
		UpdatedAt:        model.UpdatedAt,
		StudentID:        int32(model.StudentID),
	}

	err := s.Queries.UpdateStudent(ctx, params)
	if err != nil {
		return fmt.Errorf("UpdateStudent: failed to execute query: %w", err)
	}

	return nil // Success
}

// FreezeStudent sets the IsActive status of a student to false based on their roll number.
func (s *studentRepository) FreezeStudent(ctx context.Context, rollNo string) error {
	// Use sqlc generated code
	err := s.Queries.FreezeStudent(ctx, rollNo)
	if err != nil {
		return fmt.Errorf("FreezeStudent: failed to execute query: %w", err)
	}

	return nil // Success
}

// UnFreezeStudent sets the IsActive status of a student to true based on their roll number.
func (s *studentRepository) UnFreezeStudent(ctx context.Context, rollNo string) error {
	// Use sqlc generated code
	err := s.Queries.UnFreezeStudent(ctx, rollNo)
	if err != nil {
		return fmt.Errorf("UnFreezeStudent: failed to execute query: %w", err)
	}

	return nil // Success
}

// FindByKratosID retrieves a student record by their Kratos identity ID.
func (s *studentRepository) FindByKratosID(ctx context.Context, kratosID string) (*models.Student, error) {
	// Use sqlc generated code
	result, err := s.Queries.FindByKratosID(ctx, kratosID)
	if err != nil {
		if err == pgx.ErrNoRows {
			// Return nil, nil for not found, consistent with your original code
			return nil, nil
		}
		// Any other error during execution or scanning
		return nil, fmt.Errorf("FindByKratosID: failed to execute query or scan: %w", err)
	}

	// Convert db.Student to models.Student
	student := &models.Student{
		StudentID:        int(result.StudentID),
		UserID:           int(result.UserID),
		CollegeID:        int(result.CollegeID),
		KratosIdentityID: result.KratosIdentityID,
		EnrollmentYear:   int(result.EnrollmentYear.Int32),
		RollNo:           result.RollNo,
		IsActive:         result.IsActive,
		CreatedAt:        result.CreatedAt,
		UpdatedAt:        result.UpdatedAt,
	}

	return student, nil // Success
}
