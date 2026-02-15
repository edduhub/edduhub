package handler

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"eduhub/server/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

type integrationFixture struct {
	CollegeID       int
	AdminUserID     int
	FacultyUserID   int
	StudentUserID   int
	StudentID       int
	CourseID        int
	AdminKratosID   string
	FacultyKratosID string
	StudentKratosID string
}

func setupIntegrationDB(t *testing.T, requiredTables ...string) (context.Context, *repository.DB, *pgxpool.Pool) {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration test")
	}

	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}
	t.Cleanup(pool.Close)

	for _, table := range requiredTables {
		var exists bool
		err := pool.QueryRow(context.Background(), `
			SELECT EXISTS (
				SELECT 1
				FROM information_schema.tables
				WHERE table_schema = 'public' AND table_name = $1
			)`, table,
		).Scan(&exists)
		if err != nil {
			t.Fatalf("failed checking table %s: %v", table, err)
		}
		if !exists {
			t.Skipf("table %s not found, skipping integration test", table)
		}
	}

	return context.Background(), &repository.DB{Pool: pool}, pool
}

func seedIntegrationFixture(t *testing.T, ctx context.Context, pool *pgxpool.Pool) (integrationFixture, func()) {
	t.Helper()
	suffix := time.Now().UnixNano()

	fixture := integrationFixture{
		AdminKratosID:   fmt.Sprintf("kratos-admin-%d", suffix),
		FacultyKratosID: fmt.Sprintf("kratos-faculty-%d", suffix),
		StudentKratosID: fmt.Sprintf("kratos-student-%d", suffix),
	}

	err := pool.QueryRow(ctx,
		`INSERT INTO colleges (name, city, state, country) VALUES ($1, 'Test City', 'TS', 'Testland') RETURNING id`,
		fmt.Sprintf("Test College %d", suffix),
	).Scan(&fixture.CollegeID)
	if err != nil {
		t.Fatalf("failed creating college: %v", err)
	}

	err = pool.QueryRow(ctx,
		`INSERT INTO users (kratos_identity_id, name, role, email, is_active)
		 VALUES ($1, 'Admin User', 'admin', $2, TRUE) RETURNING id`,
		fixture.AdminKratosID,
		fmt.Sprintf("admin-%d@example.com", suffix),
	).Scan(&fixture.AdminUserID)
	if err != nil {
		t.Fatalf("failed creating admin user: %v", err)
	}

	err = pool.QueryRow(ctx,
		`INSERT INTO users (kratos_identity_id, name, role, email, is_active)
		 VALUES ($1, 'Faculty User', 'faculty', $2, TRUE) RETURNING id`,
		fixture.FacultyKratosID,
		fmt.Sprintf("faculty-%d@example.com", suffix),
	).Scan(&fixture.FacultyUserID)
	if err != nil {
		t.Fatalf("failed creating faculty user: %v", err)
	}

	err = pool.QueryRow(ctx,
		`INSERT INTO users (kratos_identity_id, name, role, email, is_active)
		 VALUES ($1, 'Student User', 'student', $2, TRUE) RETURNING id`,
		fixture.StudentKratosID,
		fmt.Sprintf("student-%d@example.com", suffix),
	).Scan(&fixture.StudentUserID)
	if err != nil {
		t.Fatalf("failed creating student user: %v", err)
	}

	err = pool.QueryRow(ctx,
		`INSERT INTO students (user_id, college_id, kratos_identity_id, enrollment_year, roll_no, is_active)
		 VALUES ($1, $2, $3, 2025, $4, TRUE) RETURNING student_id`,
		fixture.StudentUserID,
		fixture.CollegeID,
		fixture.StudentKratosID,
		fmt.Sprintf("ROLL-%d", suffix),
	).Scan(&fixture.StudentID)
	if err != nil {
		t.Fatalf("failed creating student: %v", err)
	}

	err = pool.QueryRow(ctx,
		`INSERT INTO courses (name, description, college_id, credits, instructor_id)
		 VALUES ($1, 'Integration test course', $2, 3, $3) RETURNING id`,
		fmt.Sprintf("Course-%d", suffix),
		fixture.CollegeID,
		fixture.FacultyUserID,
	).Scan(&fixture.CourseID)
	if err != nil {
		t.Fatalf("failed creating course: %v", err)
	}

	_, err = pool.Exec(ctx,
		`INSERT INTO enrollments (student_id, course_id, college_id, status)
		 VALUES ($1, $2, $3, 'Active')`,
		fixture.StudentID,
		fixture.CourseID,
		fixture.CollegeID,
	)
	if err != nil {
		t.Fatalf("failed creating enrollment: %v", err)
	}

	cleanup := func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM self_service_requests WHERE college_id = $1`, fixture.CollegeID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM office_hour_bookings WHERE office_hour_id IN (SELECT id FROM faculty_office_hours WHERE college_id = $1)`, fixture.CollegeID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM faculty_office_hours WHERE college_id = $1`, fixture.CollegeID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM rubric_criteria WHERE rubric_id IN (SELECT id FROM grading_rubrics WHERE college_id = $1)`, fixture.CollegeID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM grading_rubrics WHERE college_id = $1`, fixture.CollegeID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM assignment_submissions WHERE student_id = $1`, fixture.StudentID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM assignments WHERE college_id = $1`, fixture.CollegeID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM grades WHERE college_id = $1 AND student_id = $2`, fixture.CollegeID, fixture.StudentID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM attendance WHERE college_id = $1 AND student_id = $2`, fixture.CollegeID, fixture.StudentID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM enrollments WHERE college_id = $1 AND student_id = $2`, fixture.CollegeID, fixture.StudentID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM courses WHERE id = $1`, fixture.CourseID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM students WHERE student_id = $1`, fixture.StudentID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id IN ($1, $2, $3)`, fixture.AdminUserID, fixture.FacultyUserID, fixture.StudentUserID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM colleges WHERE id = $1`, fixture.CollegeID)
	}

	return fixture, cleanup
}
