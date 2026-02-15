package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"eduhub/server/internal/repository"
	"eduhub/server/internal/services/assignment"
	"eduhub/server/internal/services/attendance"
	"eduhub/server/internal/services/auth"
	"eduhub/server/internal/services/email"
	"eduhub/server/internal/services/grades"
	"eduhub/server/internal/services/student"

	"github.com/labstack/echo/v4"
)

func TestParentDashboardMetricsIntegration(t *testing.T) {
	ctx, db, pool := setupIntegrationDB(t,
		"users", "colleges", "students", "courses", "enrollments",
		"attendance", "grades", "assignments", "assignment_submissions",
	)
	fixture, cleanup := seedIntegrationFixture(t, ctx, pool)
	defer cleanup()

	_, err := pool.Exec(ctx,
		`INSERT INTO attendance (student_id, course_id, college_id, lecture_id, date, status)
		 VALUES ($1, $2, $3, 1, CURRENT_DATE, 'Present'),
		        ($1, $2, $3, 2, CURRENT_DATE - INTERVAL '1 day', 'Absent')`,
		fixture.StudentID,
		fixture.CourseID,
		fixture.CollegeID,
	)
	if err != nil {
		t.Fatalf("failed seeding attendance: %v", err)
	}

	_, err = pool.Exec(ctx,
		`INSERT INTO grades (student_id, course_id, college_id, assessment_name, assessment_type, total_marks, obtained_marks, percentage)
		 VALUES ($1, $2, $3, 'Midterm', 'midterm', 100, 82, 82.0)`,
		fixture.StudentID,
		fixture.CourseID,
		fixture.CollegeID,
	)
	if err != nil {
		t.Fatalf("failed seeding grade: %v", err)
	}

	var pendingAssignmentID int
	err = pool.QueryRow(ctx,
		`INSERT INTO assignments (course_id, college_id, title, description, due_date, max_points)
		 VALUES ($1, $2, 'Pending Assignment', 'Needs submission', NOW() + INTERVAL '7 days', 100)
		 RETURNING id`,
		fixture.CourseID,
		fixture.CollegeID,
	).Scan(&pendingAssignmentID)
	if err != nil {
		t.Fatalf("failed seeding pending assignment: %v", err)
	}

	var submittedAssignmentID int
	err = pool.QueryRow(ctx,
		`INSERT INTO assignments (course_id, college_id, title, description, due_date, max_points)
		 VALUES ($1, $2, 'Submitted Assignment', 'Already submitted', NOW() + INTERVAL '10 days', 100)
		 RETURNING id`,
		fixture.CourseID,
		fixture.CollegeID,
	).Scan(&submittedAssignmentID)
	if err != nil {
		t.Fatalf("failed seeding submitted assignment: %v", err)
	}

	_, err = pool.Exec(ctx,
		`INSERT INTO assignment_submissions (assignment_id, student_id, submission_time, content_text)
		 VALUES ($1, $2, NOW(), 'submitted work')`,
		submittedAssignmentID,
		fixture.StudentID,
	)
	if err != nil {
		t.Fatalf("failed seeding assignment submission: %v", err)
	}

	studentRepo := repository.NewStudentRepository(db)
	attendanceRepo := repository.NewAttendanceRepository(db.Pool)
	enrollmentRepo := repository.NewEnrollmentRepository(db)
	profileRepo := repository.NewProfileRepository(db)
	gradeRepo := repository.NewGradeRepository(db)
	courseRepo := repository.NewCourseRepository(db)
	assignmentRepo := repository.NewAssignmentRepository(db, nil)

	studentService := student.NewstudentService(studentRepo, attendanceRepo, enrollmentRepo, profileRepo, gradeRepo)
	attendanceService := attendance.NewAttendanceService(attendanceRepo, studentRepo, enrollmentRepo)
	gradeService := grades.NewGradeServices(gradeRepo, studentRepo, enrollmentRepo, courseRepo)
	assignmentService := assignment.NewAssignmentService(assignmentRepo, nil)
	emailService := email.NewEmailService("", "", "", "", "")

	handler := NewParentHandler(studentService, attendanceService, gradeService, assignmentService, emailService, db)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/parent/children/%d/dashboard", fixture.StudentID), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/parent/children/:studentID/dashboard")
	c.SetParamNames("studentID")
	c.SetParamValues(fmt.Sprintf("%d", fixture.StudentID))
	c.Set("college_id", fixture.CollegeID)

	identity := &auth.Identity{ID: fixture.AdminKratosID}
	identity.Traits.Role = "admin"
	identity.Traits.College.ID = fmt.Sprintf("%d", fixture.CollegeID)
	c.Set("identity", identity)

	if err := handler.GetChildDashboard(c); err != nil {
		t.Fatalf("GetChildDashboard returned error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp successEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed decoding response: %v", err)
	}

	var payload struct {
		Metrics map[string]any `json:"metrics"`
	}
	if err := json.Unmarshal(resp.Data, &payload); err != nil {
		t.Fatalf("failed decoding payload: %v", err)
	}

	attendanceRate, ok := payload.Metrics["attendanceRate"].(float64)
	if !ok || attendanceRate <= 0 {
		t.Fatalf("expected attendanceRate > 0, got %#v", payload.Metrics["attendanceRate"])
	}

	pendingAssignments, ok := payload.Metrics["pendingAssignments"].(float64)
	if !ok || pendingAssignments < 1 {
		t.Fatalf("expected pendingAssignments >= 1, got %#v", payload.Metrics["pendingAssignments"])
	}

	averageGrade, ok := payload.Metrics["averageGrade"].(float64)
	if !ok || averageGrade <= 0 {
		t.Fatalf("expected averageGrade > 0, got %#v", payload.Metrics["averageGrade"])
	}
}
