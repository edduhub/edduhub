package repository

import "eduhub/server/internal/storage"

type Repository struct {
	AttendanceRepository AttendanceRepository
	StudentRepository    StudentRepository
	UserRepository       UserRepository
	EnrollmentRepository EnrollmentRepository
	PlacementRepository  PlacementRepository  // Added Placement
	QuizRepository       QuizRepository       // Added Quiz
	DepartmentRepository DepartmentRepository // Added Department
	ProfileRepository    ProfileRepository    // Added Profile
	CourseRepository     CourseRepository
	AssignmentRepository AssignmentRepository // Added Assignment
	LectureRepository    LectureRepository
	CollegeRepository    CollegeRepository
	GradeRepository      GradeRepository
}

// NewRepository creates a new repository with all required sub-repositories
// It needs a bun.DB instance to create the base repositories
func NewRepository(DB *DB, minioCient *storage.MinioClient) *Repository {
	// Create type-specific database repositories
	attendanceRepo := NewAttendanceRepository(DB.Pool)
	studentRepo := NewStudentRepository(DB)
	userRepo := NewUserRepository(DB)
	enrollmentRepo := NewEnrollmentRepository(DB)
	placementRepo := NewPlacementRepository(DB)   // Instantiate Placement
	quizRepo := NewQuizRepository(DB)             // Instantiate Quiz
	departmentRepo := NewDepartmentRepository(DB) // Instantiate Department
	profileRepo := NewProfileRepository(DB)       // Instantiate Profile
	courseRepo := NewCourseRepository(DB)
	assignmentRepo := NewAssignmentRepository(DB, minioCient) // Instantiate Assignment
	lectureRepo := NewLectureRepository(DB)
	collegeRepo := NewCollegeRepository(DB)
	gradeRepo := NewGradeRepository(DB)
	return &Repository{
		AttendanceRepository: attendanceRepo,
		StudentRepository:    studentRepo,
		UserRepository:       userRepo,
		EnrollmentRepository: enrollmentRepo,
		PlacementRepository:  placementRepo,
		QuizRepository:       quizRepo,
		DepartmentRepository: departmentRepo,
		ProfileRepository:    profileRepo,
		CourseRepository:     courseRepo,
		AssignmentRepository: assignmentRepo,
		LectureRepository:    lectureRepo,
		CollegeRepository:    collegeRepo,
		GradeRepository:      gradeRepo,
	}
}
