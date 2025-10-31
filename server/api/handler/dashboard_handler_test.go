package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"eduhub/server/internal/models"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock services
type MockStudentService struct {
	mock.Mock
}

func (m *MockStudentService) FindByKratosID(ctx context.Context, kratosID string) (*models.Student, error) {
	args := m.Called(ctx, kratosID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Student), args.Error(1)
}

func (m *MockStudentService) GetStudentDetailedProfile(ctx context.Context, collegeID int, studentID int) (interface{}, error) {
	args := m.Called(ctx, collegeID, studentID)
	return args.Get(0), args.Error(1)
}

func (m *MockStudentService) UpdateStudentPartial(ctx context.Context, collegeID int, studentID int, req *models.UpdateStudentRequest) error {
	args := m.Called(ctx, collegeID, studentID, req)
	return args.Error(0)
}

func (m *MockStudentService) ListStudents(ctx context.Context, collegeID int, limit, offset uint64) ([]*models.Student, error) {
	args := m.Called(ctx, collegeID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Student), args.Error(1)
}

func (m *MockStudentService) CreateStudent(ctx context.Context, student *models.Student) error {
	args := m.Called(ctx, student)
	return args.Error(0)
}

func (m *MockStudentService) DeleteStudent(ctx context.Context, collegeID int, studentID int) error {
	args := m.Called(ctx, collegeID, studentID)
	return args.Error(0)
}

func (m *MockStudentService) FreezeStudent(ctx context.Context, collegeID int, studentID int) error {
	args := m.Called(ctx, collegeID, studentID)
	return args.Error(0)
}

type MockEnrollmentService struct {
	mock.Mock
}

func (m *MockEnrollmentService) FindEnrollmentsByStudent(ctx context.Context, collegeID int, studentID int, limit, offset uint64) ([]*models.Enrollment, error) {
	args := m.Called(ctx, collegeID, studentID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Enrollment), args.Error(1)
}

func (m *MockEnrollmentService) CreateEnrollment(ctx context.Context, enrollment *models.Enrollment) error {
	args := m.Called(ctx, enrollment)
	return args.Error(0)
}

func (m *MockEnrollmentService) IsStudentEnrolled(ctx context.Context, collegeID, studentID, courseID int) (bool, error) {
	args := m.Called(ctx, collegeID, studentID, courseID)
	return args.Bool(0), args.Error(1)
}

func (m *MockEnrollmentService) UpdateEnrollment(ctx context.Context, enrollment *models.Enrollment) error {
	args := m.Called(ctx, enrollment)
	return args.Error(0)
}

func (m *MockEnrollmentService) UpdateEnrollmentStatus(ctx context.Context, collegeID, enrollmentID int, status string) error {
	args := m.Called(ctx, collegeID, enrollmentID, status)
	return args.Error(0)
}

func (m *MockEnrollmentService) DeleteEnrollment(ctx context.Context, collegeID int, enrollmentID int) error {
	args := m.Called(ctx, collegeID, enrollmentID)
	return args.Error(0)
}

func (m *MockEnrollmentService) FindEnrollmentsByCourse(ctx context.Context, collegeID int, courseID int, limit, offset uint64) ([]*models.Enrollment, error) {
	args := m.Called(ctx, collegeID, courseID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Enrollment), args.Error(1)
}

func (m *MockEnrollmentService) GetEnrollmentByID(ctx context.Context, collegeID, enrollmentID int) (*models.Enrollment, error) {
	args := m.Called(ctx, collegeID, enrollmentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Enrollment), args.Error(1)
}

type MockCourseService struct {
	mock.Mock
}

func (m *MockCourseService) GetCourseByID(ctx context.Context, collegeID int, courseID int) (*models.Course, error) {
	args := m.Called(ctx, collegeID, courseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Course), args.Error(1)
}

func (m *MockCourseService) FindAllCourses(ctx context.Context, collegeID int, limit, offset uint64) ([]*models.Course, error) {
	args := m.Called(ctx, collegeID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Course), args.Error(1)
}

type MockGradesService struct {
	mock.Mock
}

func (m *MockGradesService) GetGrades(ctx context.Context, filter models.GradeFilter) ([]*models.Grade, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Grade), args.Error(1)
}

func (m *MockGradesService) CreateGrade(ctx context.Context, grade *models.Grade) error {
	args := m.Called(ctx, grade)
	return args.Error(0)
}

func (m *MockGradesService) GetGradeByID(ctx context.Context, gradeID int, collegeID int) (*models.Grade, error) {
	args := m.Called(ctx, gradeID, collegeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Grade), args.Error(1)
}

func (m *MockGradesService) UpdateGrade(ctx context.Context, grade *models.Grade) error {
	args := m.Called(ctx, grade)
	return args.Error(0)
}

func (m *MockGradesService) UpdateGradePartial(ctx context.Context, collegeID int, gradeID int, req *models.UpdateGradeRequest) error {
	args := m.Called(ctx, collegeID, gradeID, req)
	return args.Error(0)
}

func (m *MockGradesService) DeleteGrade(ctx context.Context, gradeID int, collegeID int) error {
	args := m.Called(ctx, gradeID, collegeID)
	return args.Error(0)
}

func (m *MockGradesService) GetGradesByCourse(ctx context.Context, collegeID int, courseID int) ([]*models.Grade, error) {
	args := m.Called(ctx, collegeID, courseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Grade), args.Error(1)
}

func (m *MockGradesService) GetGradesByStudent(ctx context.Context, collegeID int, studentID int) ([]*models.Grade, error) {
	args := m.Called(ctx, collegeID, studentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Grade), args.Error(1)
}

type MockAttendanceService struct {
	mock.Mock
}

func (m *MockAttendanceService) GetAttendanceByCourseAndStudent(ctx context.Context, collegeID int, courseID int, studentID int) ([]*models.Attendance, error) {
	args := m.Called(ctx, collegeID, courseID, studentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Attendance), args.Error(1)
}

type MockAssignmentService struct {
	mock.Mock
}

func (m *MockAssignmentService) ListAssignments(ctx context.Context, filter models.AssignmentFilter) ([]*models.Assignment, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Assignment), args.Error(1)
}

func (m *MockAssignmentService) GetSubmissionByStudentAndAssignment(ctx context.Context, collegeID int, studentID int, assignmentID int) (*models.AssignmentSubmission, error) {
	args := m.Called(ctx, collegeID, studentID, assignmentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AssignmentSubmission), args.Error(1)
}

func (m *MockAssignmentService) CountPendingSubmissionsByCollege(ctx context.Context, collegeID int) (int, error) {
	args := m.Called(ctx, collegeID)
	return args.Int(0), args.Error(1)
}

type MockCalendarService struct {
	mock.Mock
}

func (m *MockCalendarService) GetEvents(ctx context.Context, filter models.CalendarBlockFilter) ([]*models.CalendarBlock, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.CalendarBlock), args.Error(1)
}

type MockAnnouncementService struct {
	mock.Mock
}

func (m *MockAnnouncementService) GetAnnouncements(ctx context.Context, filter models.AnnouncementFilter) ([]*models.Announcement, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Announcement), args.Error(1)
}

type MockAnalyticsService struct {
	mock.Mock
}

type MockAuditService struct {
	mock.Mock
}

// TestGetStudentDashboard_Success tests the successful retrieval of student dashboard
func TestGetStudentDashboard_Success(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/student/dashboard", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set context values for college ID and user ID
	c.Set("collegeID", "1")
	c.Set("userID", "test-user-id")

	// Create mocks
	mockStudentService := new(MockStudentService)
	mockEnrollmentService := new(MockEnrollmentService)
	mockCourseService := new(MockCourseService)
	mockGradesService := new(MockGradesService)
	mockAttendanceService := new(MockAttendanceService)
	mockAssignmentService := new(MockAssignmentService)
	mockCalendarService := new(MockCalendarService)
	mockAnnouncementService := new(MockAnnouncementService)

	// Create handler
	handler := &DashboardHandler{
		studentService:      mockStudentService,
		courseService:       mockCourseService,
		attendanceService:   mockAttendanceService,
		announcementService: mockAnnouncementService,
		calendarService:     mockCalendarService,
		assignmentService:   mockAssignmentService,
		enrollmentService:   mockEnrollmentService,
		gradesService:       mockGradesService,
	}

	// Setup mock student
	mockStudent := &models.Student{
		ID:        1,
		RollNo:    "ST001",
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		Semester:  3,
	}

	// Setup mock enrollments
	mockEnrollments := []*models.Enrollment{
		{ID: 1, StudentID: 1, CourseID: 1, Status: "active"},
		{ID: 2, StudentID: 1, CourseID: 2, Status: "active"},
	}

	// Setup mock courses
	mockCourse1 := &models.Course{
		ID:      1,
		Code:    "CS101",
		Name:    "Introduction to Computer Science",
		Credits: 3,
	}
	mockCourse2 := &models.Course{
		ID:      2,
		Code:    "MATH201",
		Name:    "Calculus II",
		Credits: 4,
	}

	// Setup mock grades
	studentID := 1
	courseID1 := 1
	courseID2 := 2
	collegeID := 1
	mockGrades1 := []*models.Grade{
		{ID: 1, StudentID: studentID, CourseID: courseID1, Percentage: 85.0},
		{ID: 2, StudentID: studentID, CourseID: courseID1, Percentage: 90.0},
	}
	mockGrades2 := []*models.Grade{
		{ID: 3, StudentID: studentID, CourseID: courseID2, Percentage: 78.0},
	}

	// Setup mock attendance
	mockAttendance1 := []*models.Attendance{
		{ID: 1, Status: "present"},
		{ID: 2, Status: "present"},
		{ID: 3, Status: "absent"},
		{ID: 4, Status: "present"},
	}
	mockAttendance2 := []*models.Attendance{
		{ID: 5, Status: "present"},
		{ID: 6, Status: "present"},
	}

	// Setup mock assignments
	futureDate := time.Now().Add(7 * 24 * time.Hour).Format(time.RFC3339)
	mockAssignments := []*models.Assignment{
		{ID: 1, CourseID: 1, Title: "Assignment 1", DueDate: futureDate, MaxScore: 100},
		{ID: 2, CourseID: 2, Title: "Assignment 2", DueDate: futureDate, MaxScore: 100},
	}

	// Setup mock calendar events
	mockEvents := []*models.CalendarBlock{
		{ID: 1, Title: "Midterm Exam", Type: "exam"},
	}

	// Setup mock announcements
	mockAnnouncements := []*models.Announcement{
		{ID: 1, Title: "Important Notice", Content: "Test content", Priority: "high"},
	}

	// Configure expectations
	mockStudentService.On("FindByKratosID", mock.Anything, "test-user-id").Return(mockStudent, nil)
	mockEnrollmentService.On("FindEnrollmentsByStudent", mock.Anything, 1, 1, uint64(100), uint64(0)).Return(mockEnrollments, nil)
	mockCourseService.On("GetCourseByID", mock.Anything, 1, 1).Return(mockCourse1, nil)
	mockCourseService.On("GetCourseByID", mock.Anything, 1, 2).Return(mockCourse2, nil)

	// Setup grade filter mocks
	mockGradesService.On("GetGrades", mock.Anything, mock.MatchedBy(func(filter models.GradeFilter) bool {
		return filter.StudentID != nil && *filter.StudentID == studentID &&
		       filter.CourseID != nil && *filter.CourseID == courseID1 &&
		       filter.CollegeID != nil && *filter.CollegeID == collegeID
	})).Return(mockGrades1, nil)

	mockGradesService.On("GetGrades", mock.Anything, mock.MatchedBy(func(filter models.GradeFilter) bool {
		return filter.StudentID != nil && *filter.StudentID == studentID &&
		       filter.CourseID != nil && *filter.CourseID == courseID2 &&
		       filter.CollegeID != nil && *filter.CollegeID == collegeID
	})).Return(mockGrades2, nil)

	mockGradesService.On("GetGrades", mock.Anything, mock.MatchedBy(func(filter models.GradeFilter) bool {
		return filter.StudentID != nil && *filter.StudentID == studentID &&
		       filter.Limit == 10
	})).Return(append(mockGrades1, mockGrades2...), nil)

	mockAttendanceService.On("GetAttendanceByCourseAndStudent", mock.Anything, 1, 1, 1).Return(mockAttendance1, nil)
	mockAttendanceService.On("GetAttendanceByCourseAndStudent", mock.Anything, 1, 2, 1).Return(mockAttendance2, nil)
	mockAssignmentService.On("ListAssignments", mock.Anything, mock.Anything).Return(mockAssignments, nil)
	mockAssignmentService.On("GetSubmissionByStudentAndAssignment", mock.Anything, 1, 1, 1).Return(nil, errors.New("not found"))
	mockAssignmentService.On("GetSubmissionByStudentAndAssignment", mock.Anything, 1, 1, 2).Return(nil, errors.New("not found"))
	mockCalendarService.On("GetEvents", mock.Anything, mock.Anything).Return(mockEvents, nil)
	mockAnnouncementService.On("GetAnnouncements", mock.Anything, mock.Anything).Return(mockAnnouncements, nil)

	// Execute
	err := handler.GetStudentDashboard(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Parse response
	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify response structure
	assert.Contains(t, response, "data")
	data := response["data"].(map[string]interface{})

	assert.Contains(t, data, "student")
	assert.Contains(t, data, "academicOverview")
	assert.Contains(t, data, "courses")
	assert.Contains(t, data, "assignments")
	assert.Contains(t, data, "recentGrades")

	// Verify student data
	studentData := data["student"].(map[string]interface{})
	assert.Equal(t, "ST001", studentData["rollNo"])
	assert.Equal(t, "John", studentData["firstName"])

	// Verify academic overview
	academicOverview := data["academicOverview"].(map[string]interface{})
	assert.NotNil(t, academicOverview["gpa"])
	assert.Equal(t, float64(7), academicOverview["totalCredits"])

	// Verify mocks were called
	mockStudentService.AssertExpectations(t)
	mockEnrollmentService.AssertExpectations(t)
	mockCourseService.AssertExpectations(t)
	mockGradesService.AssertExpectations(t)
}

// TestGetStudentDashboard_StudentNotFound tests when student is not found
func TestGetStudentDashboard_StudentNotFound(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/student/dashboard", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.Set("collegeID", "1")
	c.Set("userID", "nonexistent-user")

	mockStudentService := new(MockStudentService)
	handler := &DashboardHandler{
		studentService: mockStudentService,
	}

	// Configure expectation - student not found
	mockStudentService.On("FindByKratosID", mock.Anything, "nonexistent-user").Return(nil, errors.New("student not found"))

	// Execute
	err := handler.GetStudentDashboard(c)

	// Assert
	assert.Error(t, err)
	mockStudentService.AssertExpectations(t)
}

// TestCalculateGradePoint tests the GPA calculation function
func TestCalculateGradePoint(t *testing.T) {
	tests := []struct {
		name       string
		percentage float64
		expected   float64
	}{
		{"Grade A (90+)", 95.0, 4.0},
		{"Grade A- (85-89)", 87.0, 3.7},
		{"Grade B+ (80-84)", 82.0, 3.3},
		{"Grade B (75-79)", 77.0, 3.0},
		{"Grade B- (70-74)", 72.0, 2.7},
		{"Grade C+ (65-69)", 67.0, 2.3},
		{"Grade C (60-64)", 62.0, 2.0},
		{"Grade C- (55-59)", 57.0, 1.7},
		{"Grade D (50-54)", 52.0, 1.0},
		{"Grade F (<50)", 45.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateGradePoint(tt.percentage)
			assert.Equal(t, tt.expected, result)
		})
	}
}
