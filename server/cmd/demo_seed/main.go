package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"

	"eduhub/server/api/handler"
	"eduhub/server/internal/config"
	"eduhub/server/internal/repository"
	services "eduhub/server/internal/services"
	authsvc "eduhub/server/internal/services/auth"
	filesvc "eduhub/server/internal/services/file"
	storagesvc "eduhub/server/internal/services/storage"
	storageclient "eduhub/server/internal/storage"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

const (
	demoCollegeName       = "EduHub Demo College"
	demoCollegeExternalID = "eduhub-demo-college"
	demoDefaultPassword   = "EduHub#2026!LocalSeed$A7q2"
)

type demoUser struct {
	ID       int
	KratosID string
	Email    string
	Role     string
	First    string
	Last     string
}

type demoFixture struct {
	CollegeID    int
	Admin        demoUser
	Faculty      demoUser
	Student      demoUser
	Parent       demoUser
	StudentID    int
	DepartmentID int
	CourseID     int
	LectureID    int
	AssignmentID int
	QuizID       int
	AttemptID    int
	ExamID       int
	FileID       int
}

type demoAccountSpec struct {
	Email            string
	Role             string
	RegistrationRole string
	First            string
	Last             string
	RollNo           string
}

var demoAccounts = []demoAccountSpec{
	{
		Email:            "admin.demo@eduhub.local",
		Role:             "admin",
		RegistrationRole: "admin",
		First:            "Demo",
		Last:             "Admin",
		RollNo:           "ADM-001",
	},
	{
		Email:            "faculty.demo@eduhub.local",
		Role:             "faculty",
		RegistrationRole: "faculty",
		First:            "Demo",
		Last:             "Faculty",
		RollNo:           "FAC-001",
	},
	{
		Email:            "student.demo@eduhub.local",
		Role:             "student",
		RegistrationRole: "student",
		First:            "Demo",
		Last:             "Student",
		RollNo:           "STU-001",
	},
	{
		Email:            "parent.demo@eduhub.local",
		Role:             "parent",
		RegistrationRole: "student",
		First:            "Demo",
		Last:             "Parent",
		RollNo:           "PAR-001",
	},
}

func main() {
	seedOnly := flag.Bool("seed-only", false, "seed demo data without verifying logins")
	verifyOnly := flag.Bool("verify-only", false, "verify seeded demo logins without reseeding")
	flag.Parse()

	if *seedOnly && *verifyOnly {
		fatalf("seed-only and verify-only cannot be used together")
	}

	loadEnvFiles()

	if *verifyOnly {
		if err := verifyDemoLogins(); err != nil {
			fatalf("demo login verification failed: %v", err)
		}
		fmt.Println("Demo login verification completed")
		return
	}

	ctx := context.Background()
	db, err := config.LoadDatabaseWithRetry(5)
	if err != nil {
		fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	_, err = seedDemo(ctx, db)
	if err != nil {
		fatalf("demo seed failed: %v", err)
	}

	if !*seedOnly {
		if err := verifyDemoLogins(); err != nil {
			fatalf("demo login verification failed: %v", err)
		}
	}

	fmt.Println("Demo seed completed")
	fmt.Println("Accounts:")
	for _, account := range demoAccounts {
		fmt.Printf("  %-7s %s / %s\n", account.Role, account.Email, demoDefaultPassword)
	}
}

func seedDemo(ctx context.Context, db *repository.DB) (*demoFixture, error) {
	fixture := &demoFixture{}
	kratos := authsvc.NewKratosService()

	collegeID, err := ensureCollege(ctx, db, demoCollegeName, demoCollegeExternalID)
	if err != nil {
		return nil, err
	}
	fixture.CollegeID = collegeID

	fixture.Admin, err = ensureUser(ctx, db, kratos, demoAccounts[0])
	if err != nil {
		return nil, fmt.Errorf("seed admin user: %w", err)
	}
	fixture.Faculty, err = ensureUser(ctx, db, kratos, demoAccounts[1])
	if err != nil {
		return nil, fmt.Errorf("seed faculty user: %w", err)
	}
	fixture.Student, err = ensureUser(ctx, db, kratos, demoAccounts[2])
	if err != nil {
		return nil, fmt.Errorf("seed student user: %w", err)
	}
	fixture.Parent, err = ensureUser(ctx, db, kratos, demoAccounts[3])
	if err != nil {
		return nil, fmt.Errorf("seed parent user: %w", err)
	}

	if err := ensureProfile(ctx, db, fixture.CollegeID, fixture.Admin, "Platform administrator"); err != nil {
		return nil, err
	}
	if err := ensureProfile(ctx, db, fixture.CollegeID, fixture.Faculty, "Faculty advisor for the demo cohort"); err != nil {
		return nil, err
	}
	if err := ensureProfile(ctx, db, fixture.CollegeID, fixture.Student, "Computer science student for demo flows"); err != nil {
		return nil, err
	}
	if err := ensureProfile(ctx, db, fixture.CollegeID, fixture.Parent, "Primary parent contact for the demo student"); err != nil {
		return nil, err
	}

	fixture.StudentID, err = ensureStudent(ctx, db, fixture.CollegeID, fixture.Student, "STU-001")
	if err != nil {
		return nil, fmt.Errorf("seed student profile row: %w", err)
	}
	if err := ensureParentRelationship(ctx, db, fixture.CollegeID, fixture.Parent.ID, fixture.StudentID); err != nil {
		return nil, fmt.Errorf("seed parent relationship: %w", err)
	}
	if err := ensureRoleAssignments(ctx, db, fixture.CollegeID, fixture); err != nil {
		return nil, fmt.Errorf("seed role assignments: %w", err)
	}

	fixture.DepartmentID, err = ensureDepartment(ctx, db, fixture.CollegeID, fixture.Faculty.ID)
	if err != nil {
		return nil, fmt.Errorf("seed department: %w", err)
	}
	fixture.CourseID, err = ensureCourse(ctx, db, fixture.CollegeID, fixture.Faculty.ID)
	if err != nil {
		return nil, fmt.Errorf("seed course: %w", err)
	}
	if err := ensureEnrollment(ctx, db, fixture.CollegeID, fixture.StudentID, fixture.CourseID); err != nil {
		return nil, fmt.Errorf("seed enrollment: %w", err)
	}

	fixture.LectureID, err = ensureLecture(ctx, db, fixture.CollegeID, fixture.CourseID)
	if err != nil {
		return nil, fmt.Errorf("seed lecture: %w", err)
	}
	if err := ensureAttendance(ctx, db, fixture.CollegeID, fixture.StudentID, fixture.CourseID, fixture.LectureID); err != nil {
		return nil, fmt.Errorf("seed attendance: %w", err)
	}
	if err := ensureTimetable(ctx, db, fixture.CollegeID, fixture.DepartmentID, fixture.CourseID, fixture.Faculty.KratosID); err != nil {
		return nil, fmt.Errorf("seed timetable: %w", err)
	}

	fixture.AssignmentID, err = ensureAssignment(ctx, db, fixture.CollegeID, fixture.CourseID)
	if err != nil {
		return nil, fmt.Errorf("seed assignment: %w", err)
	}
	if err := ensureAssignmentSubmission(ctx, db, fixture.StudentID, fixture.AssignmentID); err != nil {
		return nil, fmt.Errorf("seed assignment submission: %w", err)
	}

	fixture.QuizID, err = ensureQuiz(ctx, db, fixture.CollegeID, fixture.CourseID)
	if err != nil {
		return nil, fmt.Errorf("seed quiz: %w", err)
	}
	questionIDs, optionIDs, err := ensureQuizQuestions(ctx, db, fixture.QuizID)
	if err != nil {
		return nil, fmt.Errorf("seed quiz questions: %w", err)
	}
	fixture.AttemptID, err = ensureQuizAttempt(ctx, db, fixture.CollegeID, fixture.StudentID, fixture.QuizID)
	if err != nil {
		return nil, fmt.Errorf("seed quiz attempt: %w", err)
	}
	if err := ensureStudentAnswers(ctx, db, fixture.AttemptID, questionIDs, optionIDs); err != nil {
		return nil, fmt.Errorf("seed student answers: %w", err)
	}

	if err := ensureGrades(ctx, db, fixture.CollegeID, fixture.StudentID, fixture.CourseID, fixture.Faculty); err != nil {
		return nil, fmt.Errorf("seed grades: %w", err)
	}
	if err := ensureCalendarEvent(ctx, db, fixture.CollegeID, fixture.CourseID, fixture.Faculty); err != nil {
		return nil, fmt.Errorf("seed calendar event: %w", err)
	}
	if err := ensureAnnouncement(ctx, db, fixture.CollegeID, fixture.CourseID, fixture.Faculty); err != nil {
		return nil, fmt.Errorf("seed announcement: %w", err)
	}
	if err := ensureFees(ctx, db, fixture.CollegeID, fixture.StudentID, fixture.Admin.ID, fixture.CourseID, fixture.DepartmentID); err != nil {
		return nil, fmt.Errorf("seed fees: %w", err)
	}

	fixture.ExamID, err = ensureExamFlow(ctx, db, fixture)
	if err != nil {
		return nil, fmt.Errorf("seed exam flow: %w", err)
	}
	if err := ensureFacultyTools(ctx, db, fixture); err != nil {
		return nil, fmt.Errorf("seed faculty tools: %w", err)
	}
	if err := ensureSelfService(ctx, db, fixture); err != nil {
		return nil, fmt.Errorf("seed self service: %w", err)
	}
	if err := ensureForum(ctx, db, fixture); err != nil {
		return nil, fmt.Errorf("seed forum: %w", err)
	}
	if err := ensurePlacements(ctx, db, fixture); err != nil {
		return nil, fmt.Errorf("seed placements: %w", err)
	}
	if err := ensureWebhooks(ctx, db, fixture); err != nil {
		return nil, fmt.Errorf("seed webhooks: %w", err)
	}
	if err := ensureAudit(ctx, db, fixture); err != nil {
		return nil, fmt.Errorf("seed audit: %w", err)
	}
	if err := ensureNotifications(ctx, db, fixture); err != nil {
		return nil, fmt.Errorf("seed notifications: %w", err)
	}

	fixture.FileID, err = ensureFileArtifacts(ctx, db, fixture)
	if err != nil {
		return nil, fmt.Errorf("seed file artifacts: %w", err)
	}

	return fixture, nil
}

type registrationClient interface {
	InitiateRegistrationFlow(ctx context.Context) (map[string]any, error)
	CompleteRegistration(ctx context.Context, flowID string, regReq authsvc.RegistrationRequest) (*authsvc.Identity, error)
	GetIdentity(ctx context.Context, identityID string) (*authsvc.Identity, error)
	FindIdentityByEmail(ctx context.Context, email string) (*authsvc.Identity, error)
	DeleteIdentity(ctx context.Context, identityID string) error
}

func loadEnvFiles() {
	for _, path := range []string{".env", "server/.env.local"} {
		_ = godotenv.Load(path)
	}
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

func ensureCollege(ctx context.Context, db *repository.DB, name string, externalID string) (int, error) {
	var id int
	err := db.Pool.QueryRow(ctx, `
		SELECT id
		FROM colleges
		WHERE name = $1 OR external_id = $2
		ORDER BY CASE WHEN name = $1 THEN 0 ELSE 1 END, id
		LIMIT 1`, name, externalID).Scan(&id)
	if err != nil && !errorsIsNoRows(err) {
		return 0, fmt.Errorf("lookup college: %w", err)
	}

	if errorsIsNoRows(err) {
		err = db.Pool.QueryRow(ctx, `
			INSERT INTO colleges (name, external_id, address, city, state, country)
			VALUES ($1, $2, '123 Demo Street', 'Bengaluru', 'KA', 'India')
			RETURNING id`, name, externalID).Scan(&id)
		if err != nil {
			return 0, fmt.Errorf("insert college: %w", err)
		}
	}

	_, err = db.Pool.Exec(ctx, `
		UPDATE colleges
		SET name = $2,
		    external_id = $3,
		    address = '123 Demo Street',
		    city = 'Bengaluru',
		    state = 'KA',
		    country = 'India',
		    updated_at = NOW()
		WHERE id = $1`, id, name, externalID)
	return id, err
}

func ensureUser(ctx context.Context, db *repository.DB, kratos registrationClient, spec demoAccountSpec) (demoUser, error) {
	user := demoUser{
		Email: spec.Email,
		Role:  spec.Role,
		First: spec.First,
		Last:  spec.Last,
	}

	err := db.Pool.QueryRow(ctx, `SELECT id, kratos_identity_id FROM users WHERE email = $1`, spec.Email).Scan(&user.ID, &user.KratosID)
	if err != nil && !errorsIsNoRows(err) {
		return user, fmt.Errorf("lookup user %s: %w", spec.Email, err)
	}

	identity, err := reconcileKratosIdentity(ctx, kratos, user.KratosID, spec)
	if err != nil {
		return user, err
	}

	fullName := strings.TrimSpace(spec.First + " " + spec.Last)
	err = db.Pool.QueryRow(ctx, `
		INSERT INTO users (kratos_identity_id, name, role, email, is_active)
		VALUES ($1, $2, $3, $4, TRUE)
		ON CONFLICT (email)
		DO UPDATE SET kratos_identity_id = EXCLUDED.kratos_identity_id,
		              name = EXCLUDED.name,
		              role = EXCLUDED.role,
		              is_active = TRUE,
		              updated_at = NOW()
		RETURNING id`, identity.ID, fullName, spec.Role, spec.Email).Scan(&user.ID)
	if err != nil {
		return user, fmt.Errorf("upsert user %s: %w", spec.Email, err)
	}

	user.KratosID = identity.ID
	return user, nil
}

func reconcileKratosIdentity(ctx context.Context, kratos registrationClient, currentIdentityID string, spec demoAccountSpec) (*authsvc.Identity, error) {
	expected := authsvc.Traits{
		Email:   spec.Email,
		Name:    authsvc.Name{First: spec.First, Last: spec.Last},
		Role:    spec.RegistrationRole,
		College: authsvc.College{ID: demoCollegeExternalID, Name: demoCollegeName},
		RollNo:  spec.RollNo,
	}

	var currentIdentity *authsvc.Identity
	if currentIdentityID != "" {
		identity, err := kratos.GetIdentity(ctx, currentIdentityID)
		if err == nil {
			currentIdentity = identity
		}
	}

	identityByEmail, err := kratos.FindIdentityByEmail(ctx, spec.Email)
	if err != nil {
		return nil, fmt.Errorf("lookup kratos identity for %s: %w", spec.Email, err)
	}

	if currentIdentity != nil && identityMatches(currentIdentity, expected) {
		return currentIdentity, nil
	}

	if currentIdentity == nil && identityByEmail != nil && identityMatches(identityByEmail, expected) {
		return identityByEmail, nil
	}

	deleted := map[string]struct{}{}
	for _, identityID := range []string{currentIdentityID, identityID(identityByEmail)} {
		if identityID == "" {
			continue
		}
		if _, seen := deleted[identityID]; seen {
			continue
		}
		if err := kratos.DeleteIdentity(ctx, identityID); err != nil && !strings.Contains(strings.ToLower(err.Error()), "not found") {
			return nil, fmt.Errorf("delete stale kratos identity %s for %s: %w", identityID, spec.Email, err)
		}
		deleted[identityID] = struct{}{}
	}

	identity, err := registerIdentity(ctx, kratos, expected)
	if err != nil {
		return nil, fmt.Errorf("register kratos identity for %s: %w", spec.Email, err)
	}
	return identity, nil
}

func registerIdentity(ctx context.Context, kratos registrationClient, traits authsvc.Traits) (*authsvc.Identity, error) {
	flow, err := kratos.InitiateRegistrationFlow(ctx)
	if err != nil {
		return nil, fmt.Errorf("initiate registration flow: %w", err)
	}

	flowID, _ := flow["id"].(string)
	if flowID == "" {
		return nil, fmt.Errorf("registration flow id missing")
	}

	return kratos.CompleteRegistration(ctx, flowID, authsvc.RegistrationRequest{
		Method:   "password",
		Password: demoDefaultPassword,
		Traits:   traits,
	})
}

func identityMatches(identity *authsvc.Identity, expected authsvc.Traits) bool {
	if identity == nil {
		return false
	}

	return identity.Traits.Email == expected.Email &&
		identity.Traits.Name.First == expected.Name.First &&
		identity.Traits.Name.Last == expected.Name.Last &&
		identity.Traits.Role == expected.Role &&
		identity.Traits.College.ID == expected.College.ID &&
		identity.Traits.College.Name == expected.College.Name &&
		identity.Traits.RollNo == expected.RollNo
}

func identityID(identity *authsvc.Identity) string {
	if identity == nil {
		return ""
	}
	return identity.ID
}

func ensureStudent(ctx context.Context, db *repository.DB, collegeID int, user demoUser, rollNo string) (int, error) {
	var studentID int
	err := db.Pool.QueryRow(ctx, `
		INSERT INTO students (user_id, college_id, kratos_identity_id, enrollment_year, roll_no, is_active)
		VALUES ($1, $2, $3, 2025, $4, TRUE)
		ON CONFLICT (user_id)
		DO UPDATE SET college_id = EXCLUDED.college_id,
		              kratos_identity_id = EXCLUDED.kratos_identity_id,
		              roll_no = EXCLUDED.roll_no,
		              is_active = TRUE,
		              updated_at = NOW()
		RETURNING student_id`, user.ID, collegeID, user.KratosID, rollNo).Scan(&studentID)
	return studentID, err
}

func ensureProfile(ctx context.Context, db *repository.DB, collegeID int, user demoUser, bio string) error {
	_, err := db.Pool.Exec(ctx, `
		INSERT INTO profiles (user_id, college_id, first_name, last_name, bio, profile_image, phone_number, address, last_active, preferences, social_links)
		VALUES ($1, $2, $3, $4, $5, '', '+91-9999999999', 'Demo Address', NOW(), '{}'::jsonb, '{}'::jsonb)
		ON CONFLICT (user_id)
		DO UPDATE SET first_name = EXCLUDED.first_name,
		              last_name = EXCLUDED.last_name,
		              bio = EXCLUDED.bio,
		              profile_image = EXCLUDED.profile_image,
		              phone_number = EXCLUDED.phone_number,
		              address = EXCLUDED.address,
		              last_active = NOW(),
		              updated_at = NOW()`,
		user.ID, collegeID, user.First, user.Last, bio)
	return err
}

func ensureParentRelationship(ctx context.Context, db *repository.DB, collegeID, parentUserID, studentID int) error {
	_, err := db.Pool.Exec(ctx, `
		INSERT INTO parent_student_relationships (parent_user_id, student_id, college_id, relation, is_primary_contact, receive_notifications, is_verified, verified_at)
		VALUES ($1, $2, $3, 'guardian', TRUE, TRUE, TRUE, NOW())
		ON CONFLICT (parent_user_id, student_id, college_id)
		DO UPDATE SET is_primary_contact = TRUE, receive_notifications = TRUE, is_verified = TRUE, verified_at = NOW(), updated_at = NOW()`,
		parentUserID, studentID, collegeID)
	return err
}

func ensureRoleAssignments(ctx context.Context, db *repository.DB, collegeID int, fixture *demoFixture) error {
	roles := []struct {
		name   string
		userID int
	}{
		{name: "admin", userID: fixture.Admin.ID},
		{name: "faculty", userID: fixture.Faculty.ID},
		{name: "student", userID: fixture.Student.ID},
		{name: "parent", userID: fixture.Parent.ID},
	}

	var parentRoleID int
	err := db.Pool.QueryRow(ctx, `SELECT id FROM roles WHERE name = 'parent' LIMIT 1`).Scan(&parentRoleID)
	if errorsIsNoRows(err) {
		err = db.Pool.QueryRow(ctx, `INSERT INTO roles (name, description, college_id, is_system_role) VALUES ('parent', 'Parent portal access', $1, FALSE) RETURNING id`, collegeID).Scan(&parentRoleID)
	}
	if err != nil {
		return err
	}
	_ = parentRoleID

	for _, roleAssignment := range roles {
		var roleID int
		if err := db.Pool.QueryRow(ctx, `SELECT id FROM roles WHERE name = $1 LIMIT 1`, roleAssignment.name).Scan(&roleID); err != nil {
			return err
		}
		if _, err := db.Pool.Exec(ctx, `
			INSERT INTO user_role_assignments (user_id, role_id, assigned_by)
			VALUES ($1, $2, $3)
			ON CONFLICT (user_id, role_id) DO NOTHING`, roleAssignment.userID, roleID, fixture.Admin.ID); err != nil {
			return err
		}
	}
	return nil
}

func ensureDepartment(ctx context.Context, db *repository.DB, collegeID int, facultyUserID int) (int, error) {
	const code = "CSE"
	var id int
	err := db.Pool.QueryRow(ctx, `SELECT id FROM departments WHERE college_id = $1 AND code = $2`, collegeID, code).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errorsIsNoRows(err) {
		return 0, err
	}
	desc := "Department of Computer Science and Engineering"
	err = db.Pool.QueryRow(ctx, `
		INSERT INTO departments (college_id, name, code, description, head_user_id, hod, is_active)
		VALUES ($1, 'Computer Science', $2, $3, $4, 'Demo Faculty', TRUE)
		RETURNING id`, collegeID, code, desc, facultyUserID).Scan(&id)
	return id, err
}

func ensureCourse(ctx context.Context, db *repository.DB, collegeID int, facultyUserID int) (int, error) {
	const name = "Foundations of Software Engineering"
	var id int
	err := db.Pool.QueryRow(ctx, `SELECT id FROM courses WHERE college_id = $1 AND name = $2`, collegeID, name).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errorsIsNoRows(err) {
		return 0, err
	}
	err = db.Pool.QueryRow(ctx, `
		INSERT INTO courses (name, description, college_id, credits, instructor_id)
		VALUES ($1, 'Core software engineering concepts for the demo cohort', $2, 4, $3)
		RETURNING id`, name, collegeID, facultyUserID).Scan(&id)
	return id, err
}

func ensureEnrollment(ctx context.Context, db *repository.DB, collegeID, studentID, courseID int) error {
	_, err := db.Pool.Exec(ctx, `
		INSERT INTO enrollments (student_id, course_id, college_id, status, grade)
		VALUES ($1, $2, $3, 'Active', 'A')
		ON CONFLICT (student_id, course_id)
		DO UPDATE SET status = 'Active', grade = 'A', updated_at = NOW()`, studentID, courseID, collegeID)
	return err
}

func ensureLecture(ctx context.Context, db *repository.DB, collegeID, courseID int) (int, error) {
	const title = "Sprint Planning and Estimation"
	var id int
	err := db.Pool.QueryRow(ctx, `SELECT id FROM lectures WHERE college_id = $1 AND course_id = $2 AND title = $3`, collegeID, courseID, title).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errorsIsNoRows(err) {
		return 0, err
	}
	start := time.Now().Add(24 * time.Hour).Truncate(time.Minute)
	end := start.Add(90 * time.Minute)
	err = db.Pool.QueryRow(ctx, `
		INSERT INTO lectures (course_id, college_id, title, description, start_time, end_time, meeting_link)
		VALUES ($1, $2, $3, 'Weekly lecture for demo timetable and attendance flows', $4, $5, 'https://meet.example.com/demo')
		RETURNING id`, courseID, collegeID, title, start, end).Scan(&id)
	return id, err
}

func ensureAttendance(ctx context.Context, db *repository.DB, collegeID, studentID, courseID, lectureID int) error {
	date := time.Now().Add(-24 * time.Hour).Format("2006-01-02")
	_, err := db.Pool.Exec(ctx, `
		INSERT INTO attendance (student_id, course_id, college_id, lecture_id, date, status)
		VALUES ($1, $2, $3, $4, $5, 'Present')
		ON CONFLICT (student_id, course_id, lecture_id, date, college_id)
		DO UPDATE SET status = 'Present', updated_at = NOW()`, studentID, courseID, collegeID, lectureID, date)
	return err
}

func ensureTimetable(ctx context.Context, db *repository.DB, collegeID, departmentID, courseID int, facultyID string) error {
	for _, entry := range []struct {
		day   int
		start string
		end   string
		room  string
	}{
		{day: 1, start: "09:00:00", end: "10:30:00", room: "A-101"},
		{day: 3, start: "11:00:00", end: "12:30:00", room: "A-102"},
	} {
		var existingID int
		err := db.Pool.QueryRow(ctx, `
			SELECT id
			FROM timetable_blocks
			WHERE college_id = $1 AND course_id = $2 AND day_of_week = $3 AND start_time = $4::time
			LIMIT 1`, collegeID, courseID, entry.day, entry.start).Scan(&existingID)
		if err == nil {
			if _, updateErr := db.Pool.Exec(ctx, `
				UPDATE timetable_blocks
				SET faculty_id = $2,
				    room_number = $3,
				    end_time = $4::time,
				    updated_at = NOW()
				WHERE id = $1`, existingID, facultyID, entry.room, entry.end); updateErr != nil {
				return updateErr
			}
			continue
		}
		if !errorsIsNoRows(err) {
			return err
		}
		if _, err := db.Pool.Exec(ctx, `
			INSERT INTO timetable_blocks (college_id, department_id, course_id, class_id, day_of_week, start_time, end_time, room_number, faculty_id)
			VALUES ($1, $2, $3, NULL, $4, $5::time, $6::time, $7, $8)`,
			collegeID, departmentID, courseID, entry.day, entry.start, entry.end, entry.room, facultyID); err != nil {
			return err
		}
	}
	return nil
}

func ensureAssignment(ctx context.Context, db *repository.DB, collegeID, courseID int) (int, error) {
	const title = "Architecture Case Study"
	var id int
	err := db.Pool.QueryRow(ctx, `SELECT id FROM assignments WHERE college_id = $1 AND course_id = $2 AND title = $3`, collegeID, courseID, title).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errorsIsNoRows(err) {
		return 0, err
	}
	dueDate := time.Now().Add(72 * time.Hour)
	err = db.Pool.QueryRow(ctx, `
		INSERT INTO assignments (course_id, college_id, title, description, due_date, max_points)
		VALUES ($1, $2, $3, 'Submit a short architecture review for the demo platform', $4, 100)
		RETURNING id`, courseID, collegeID, title, dueDate).Scan(&id)
	return id, err
}

func ensureAssignmentSubmission(ctx context.Context, db *repository.DB, studentID, assignmentID int) error {
	_, err := db.Pool.Exec(ctx, `
		INSERT INTO assignment_submissions (assignment_id, student_id, content_text, grade, feedback)
		VALUES ($1, $2, 'Reviewed the service boundaries and auth flow.', 92, 'Strong submission with clear reasoning.')
		ON CONFLICT (assignment_id, student_id)
		DO UPDATE SET content_text = EXCLUDED.content_text, grade = EXCLUDED.grade, feedback = EXCLUDED.feedback, updated_at = NOW()`, assignmentID, studentID)
	return err
}

func ensureQuiz(ctx context.Context, db *repository.DB, collegeID, courseID int) (int, error) {
	const title = "Sprint Review Quiz"
	var id int
	err := db.Pool.QueryRow(ctx, `SELECT id FROM quizzes WHERE college_id = $1 AND course_id = $2 AND title = $3`, collegeID, courseID, title).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errorsIsNoRows(err) {
		return 0, err
	}
	start := time.Now().Add(-2 * time.Hour)
	end := time.Now().Add(7 * 24 * time.Hour)
	err = db.Pool.QueryRow(ctx, `
		INSERT INTO quizzes (course_id, college_id, title, description, duration_minutes, total_marks, passing_marks, start_time, end_time, is_active, time_limit_minutes, due_date)
		VALUES ($1, $2, $3, 'Short quiz covering sprint planning and review concepts', 30, 20, 10, $4, $5, TRUE, 30, $5)
		RETURNING id`, courseID, collegeID, title, start, end).Scan(&id)
	return id, err
}

func ensureQuizQuestions(ctx context.Context, db *repository.DB, quizID int) ([]int, []int, error) {
	questionTexts := []string{
		"Which artifact captures the sprint goal?",
		"Velocity is best described as what?",
	}
	questionIDs := make([]int, 0, len(questionTexts))
	correctOptionIDs := make([]int, 0, len(questionTexts))
	for index, text := range questionTexts {
		var qid int
		err := db.Pool.QueryRow(ctx, `SELECT id FROM questions WHERE quiz_id = $1 AND COALESCE(text, question_text) = $2`, quizID, text).Scan(&qid)
		if errorsIsNoRows(err) {
			err = db.Pool.QueryRow(ctx, `
				INSERT INTO questions (quiz_id, question_text, question_type, marks, order_num, text, type, points, correct_answer)
				VALUES ($1, $2, 'multiple_choice', 10, $3, $2, 'multiple_choice', 10, NULL)
				RETURNING id`, quizID, text, index+1).Scan(&qid)
		}
		if err != nil {
			return nil, nil, err
		}
		questionIDs = append(questionIDs, qid)

		var options [][]any
		if index == 0 {
			options = [][]any{{"The sprint backlog summary", true}, {"The team charter", false}, {"The release burndown", false}}
		} else {
			options = [][]any{{"The amount of work completed per sprint", true}, {"The remaining budget", false}, {"The test coverage percentage", false}}
		}
		var correctID int
		for order, option := range options {
			var optionID int
			err = db.Pool.QueryRow(ctx, `SELECT id FROM answer_options WHERE question_id = $1 AND COALESCE(text, option_text) = $2`, qid, option[0]).Scan(&optionID)
			if errorsIsNoRows(err) {
				err = db.Pool.QueryRow(ctx, `
					INSERT INTO answer_options (question_id, option_text, is_correct, order_num, text)
					VALUES ($1, $2, $3, $4, $2)
					RETURNING id`, qid, option[0], option[1], order+1).Scan(&optionID)
			}
			if err != nil {
				return nil, nil, err
			}
			if option[1].(bool) {
				correctID = optionID
			}
		}
		correctOptionIDs = append(correctOptionIDs, correctID)
	}
	return questionIDs, correctOptionIDs, nil
}

func ensureQuizAttempt(ctx context.Context, db *repository.DB, collegeID, studentID, quizID int) (int, error) {
	var id int
	err := db.Pool.QueryRow(ctx, `SELECT id FROM quiz_attempts WHERE college_id = $1 AND student_id = $2 AND quiz_id = $3`, collegeID, studentID, quizID).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errorsIsNoRows(err) {
		return 0, err
	}
	start := time.Now().Add(-90 * time.Minute)
	end := time.Now().Add(-60 * time.Minute)
	err = db.Pool.QueryRow(ctx, `
		INSERT INTO quiz_attempts (quiz_id, student_id, college_id, start_time, end_time, score, status)
		VALUES ($1, $2, $3, $4, $5, 18, 'graded')
		RETURNING id`, quizID, studentID, collegeID, start, end).Scan(&id)
	return id, err
}

func ensureStudentAnswers(ctx context.Context, db *repository.DB, attemptID int, questionIDs []int, optionIDs []int) error {
	for i, questionID := range questionIDs {
		_, err := db.Pool.Exec(ctx, `
			INSERT INTO student_answers (attempt_id, quiz_attempt_id, question_id, selected_option_id, answer_text, is_correct, marks_awarded, points_awarded)
			VALUES ($1, $1, $2, $3, NULL, TRUE, 9, 9)
			ON CONFLICT (quiz_attempt_id, question_id)
			DO UPDATE SET selected_option_id = EXCLUDED.selected_option_id,
			              is_correct = EXCLUDED.is_correct,
			              marks_awarded = EXCLUDED.marks_awarded,
			              points_awarded = EXCLUDED.points_awarded,
			              updated_at = NOW()`, attemptID, questionID, optionIDs[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func ensureGrades(ctx context.Context, db *repository.DB, collegeID, studentID, courseID int, faculty demoUser) error {
	entries := []struct {
		name     string
		typeName string
		total    int
		obtained int
		grade    string
	}{
		{name: "Architecture Case Study", typeName: "assignment", total: 100, obtained: 92, grade: "A"},
		{name: "Sprint Review Quiz", typeName: "quiz", total: 20, obtained: 18, grade: "A"},
	}
	for _, entry := range entries {
		var existingID int
		err := db.Pool.QueryRow(ctx,
			`SELECT id FROM grades WHERE student_id = $1 AND course_id = $2 AND assessment_name = $3 LIMIT 1`,
			studentID, courseID, entry.name,
		).Scan(&existingID)
		if err == nil {
			continue
		}
		if !errorsIsNoRows(err) {
			return err
		}
		percentage := float64(entry.obtained) / float64(entry.total) * 100
		_, err = db.Pool.Exec(ctx, `
			INSERT INTO grades (student_id, course_id, college_id, assessment_name, assessment_type, total_marks, obtained_marks, percentage, grade, remarks, graded_by, graded_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'Seeded for demo readiness', $10, NOW())`,
			studentID, courseID, collegeID, entry.name, entry.typeName, entry.total, entry.obtained, percentage, entry.grade, faculty.First+" "+faculty.Last,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func ensureCalendarEvent(ctx context.Context, db *repository.DB, collegeID, courseID int, faculty demoUser) error {
	_, err := db.Pool.Exec(ctx, `
		INSERT INTO calendar_events (college_id, course_id, title, description, event_type, start_time, end_time, location, created_by)
		SELECT $1, $2, 'Demo Stakeholder Review', 'Calendar item for the demo walkthrough', 'meeting', NOW() + INTERVAL '2 days', NOW() + INTERVAL '2 days 2 hours', 'Main Conference Room', $3
		WHERE NOT EXISTS (
			SELECT 1 FROM calendar_events WHERE college_id = $1 AND title = 'Demo Stakeholder Review'
		)`, collegeID, courseID, faculty.First+" "+faculty.Last)
	return err
}

func ensureAnnouncement(ctx context.Context, db *repository.DB, collegeID, courseID int, faculty demoUser) error {
	_, err := db.Pool.Exec(ctx, `
		INSERT INTO announcements (college_id, course_id, title, content, priority, is_published, published_at, created_by)
		SELECT $1, $2, 'Demo Week Schedule', 'All demo stakeholders can use these seeded records to validate academic workflows.', 'high', TRUE, NOW(), $3
		WHERE NOT EXISTS (
			SELECT 1 FROM announcements WHERE college_id = $1 AND title = 'Demo Week Schedule'
		)`, collegeID, courseID, faculty.First+" "+faculty.Last)
	return err
}

func ensureFees(ctx context.Context, db *repository.DB, collegeID, studentID, adminUserID, courseID, departmentID int) error {
	_ = adminUserID
	_ = courseID
	_ = departmentID

	var structureID int
	err := db.Pool.QueryRow(ctx, `SELECT id FROM fee_structures WHERE college_id = $1 AND name = 'Semester Tuition'`, collegeID).Scan(&structureID)
	if errorsIsNoRows(err) {
		due := time.Now().Add(15 * 24 * time.Hour)
		err = db.Pool.QueryRow(ctx, `
			INSERT INTO fee_structures (college_id, name, description, amount, currency, fee_type, academic_year, is_active, due_date, frequency, semester, is_mandatory)
			VALUES ($1, 'Semester Tuition', 'Primary semester tuition fee', 1500.00, 'USD', 'tuition', '2025-2026', TRUE, $2, 'semester', 'Semester 6', TRUE)
			RETURNING id`, collegeID, due).Scan(&structureID)
	}
	if err != nil {
		return err
	}

	var assignmentID int
	err = db.Pool.QueryRow(ctx, `SELECT id FROM fee_assignments WHERE student_id = $1 AND fee_structure_id = $2`, studentID, structureID).Scan(&assignmentID)
	if errorsIsNoRows(err) {
		due := time.Now().Add(15 * 24 * time.Hour)
		paymentDueDate := due.Format("2006-01-02")
		err = db.Pool.QueryRow(ctx, `
			INSERT INTO fee_assignments (college_id, student_id, fee_structure_id, amount, status, paid_amount, payment_due_date, is_waived, waiver_reason, currency, waiver_amount, due_date)
			VALUES ($1, $2, $3, 1500.00, 'partial', 500.00, $4::date, FALSE, NULL, 'USD', 0, $5)
			RETURNING id`, collegeID, studentID, structureID, paymentDueDate, due).Scan(&assignmentID)
	}
	if err != nil {
		return err
	}

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO fee_payments (college_id, fee_assignment_id, student_id, amount, payment_method, payment_status, transaction_id, payment_date)
		SELECT $1, $2, $3, 500.00, 'online', 'completed', 'demo-fee-payment-001', NOW() - INTERVAL '1 day'
		WHERE NOT EXISTS (SELECT 1 FROM fee_payments WHERE transaction_id = 'demo-fee-payment-001')`, collegeID, assignmentID, studentID)
	return err
}

func ensureExamFlow(ctx context.Context, db *repository.DB, fixture *demoFixture) (int, error) {
	var roomID int
	err := db.Pool.QueryRow(ctx, `SELECT id FROM exam_rooms WHERE college_id = $1 AND room_number = 'ER-1'`, fixture.CollegeID).Scan(&roomID)
	if errorsIsNoRows(err) {
		err = db.Pool.QueryRow(ctx, `
			INSERT INTO exam_rooms (college_id, room_number, room_name, capacity, location, facilities, is_active)
			VALUES ($1, 'ER-1', 'Exam Hall 1', 60, 'Block A', 'Projector,AC', TRUE)
			RETURNING id`, fixture.CollegeID).Scan(&roomID)
	}
	if err != nil {
		return 0, err
	}

	var examID int
	err = db.Pool.QueryRow(ctx, `SELECT id FROM exams WHERE college_id = $1 AND course_id = $2 AND title = 'Midterm Practical'`, fixture.CollegeID, fixture.CourseID).Scan(&examID)
	if errorsIsNoRows(err) {
		start := time.Now().Add(5 * 24 * time.Hour).Truncate(time.Minute)
		end := start.Add(2 * time.Hour)
		err = db.Pool.QueryRow(ctx, `
			INSERT INTO exams (college_id, course_id, title, description, exam_type, start_time, end_time, duration, total_marks, passing_marks, room_id, status, instructions, allowed_materials, question_paper_sets, created_by)
			VALUES ($1, $2, 'Midterm Practical', 'Hands-on exam for the demo course', 'practical', $3, $4, 120, 100, 40, $5, 'scheduled', 'Bring your laptop', 'Notebook', 1, $6)
			RETURNING id`, fixture.CollegeID, fixture.CourseID, start, end, roomID, fixture.Faculty.ID).Scan(&examID)
	}
	if err != nil {
		return 0, err
	}

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO exam_enrollments (exam_id, student_id, college_id, seat_number, room_number, question_paper_set, status, hall_ticket_generated)
		VALUES ($1, $2, $3, 'A12', 'ER-1', 1, 'enrolled', TRUE)
		ON CONFLICT (exam_id, student_id)
		DO UPDATE SET seat_number = 'A12', room_number = 'ER-1', hall_ticket_generated = TRUE, updated_at = NOW()`, examID, fixture.StudentID, fixture.CollegeID)
	if err != nil {
		return 0, err
	}

	var resultID int
	err = db.Pool.QueryRow(ctx, `SELECT id FROM exam_results WHERE exam_id = $1 AND student_id = $2`, examID, fixture.StudentID).Scan(&resultID)
	if errorsIsNoRows(err) {
		marks := 88.0
		percentage := 88.0
		grade := "A"
		err = db.Pool.QueryRow(ctx, `
			INSERT INTO exam_results (exam_id, student_id, college_id, marks_obtained, grade, percentage, result, remarks, evaluated_by, evaluated_at, revaluation_status)
			VALUES ($1, $2, $3, $4, $5, $6, 'pass', 'Strong practical performance', $7, NOW(), 'requested')
			RETURNING id`, examID, fixture.StudentID, fixture.CollegeID, marks, grade, percentage, fixture.Faculty.ID).Scan(&resultID)
	}
	if err != nil {
		return 0, err
	}

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO revaluation_requests (exam_result_id, student_id, college_id, reason, status, previous_marks, revised_marks, reviewed_by, review_comments, requested_at, reviewed_at)
		SELECT $1, $2, $3, 'Requesting review of the final rubric criterion.', 'approved', 88.0, 90.0, $4, 'Adjusted after manual review', NOW() - INTERVAL '12 hours', NOW()
		WHERE NOT EXISTS (SELECT 1 FROM revaluation_requests WHERE exam_result_id = $1 AND student_id = $2)`, resultID, fixture.StudentID, fixture.CollegeID, fixture.Admin.ID)
	if err != nil {
		return 0, err
	}

	return examID, nil
}

func ensureFacultyTools(ctx context.Context, db *repository.DB, fixture *demoFixture) error {
	var rubricID int
	err := db.Pool.QueryRow(ctx, `SELECT id FROM grading_rubrics WHERE college_id = $1 AND faculty_id = $2 AND name = 'Presentation Rubric'`, fixture.CollegeID, fixture.Faculty.ID).Scan(&rubricID)
	if errorsIsNoRows(err) {
		err = db.Pool.QueryRow(ctx, `
			INSERT INTO grading_rubrics (faculty_id, college_id, name, description, course_id, is_template, is_active, max_score)
			VALUES ($1, $2, 'Presentation Rubric', 'Used for student presentation grading', $3, TRUE, TRUE, 100)
			RETURNING id`, fixture.Faculty.ID, fixture.CollegeID, fixture.CourseID).Scan(&rubricID)
	}
	if err != nil {
		return err
	}
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO rubric_criteria (rubric_id, name, description, weight, max_score, sort_order)
		SELECT $1, 'Clarity', 'Explains architecture choices clearly', 0.50, 50, 1
		WHERE NOT EXISTS (SELECT 1 FROM rubric_criteria WHERE rubric_id = $1 AND name = 'Clarity')`, rubricID)
	if err != nil {
		return err
	}

	var officeHourID int
	err = db.Pool.QueryRow(ctx, `SELECT id FROM faculty_office_hours WHERE college_id = $1 AND faculty_id = $2 AND day_of_week = 2 AND start_time = '14:00:00'::time`, fixture.CollegeID, fixture.Faculty.ID).Scan(&officeHourID)
	if errorsIsNoRows(err) {
		err = db.Pool.QueryRow(ctx, `
			INSERT INTO faculty_office_hours (faculty_id, college_id, day_of_week, start_time, end_time, location, is_virtual, virtual_link, max_students, is_active)
			VALUES ($1, $2, 2, '14:00:00', '15:00:00', 'Faculty Cabin', TRUE, 'https://meet.example.com/faculty-office-hour', 5, TRUE)
			RETURNING id`, fixture.Faculty.ID, fixture.CollegeID).Scan(&officeHourID)
	}
	if err != nil {
		return err
	}

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO office_hour_bookings (office_hour_id, student_id, booking_date, start_time, end_time, purpose, status, notes)
		VALUES ($1, $2, CURRENT_DATE + 1, '14:00:00', '14:20:00', 'Discuss the case study feedback', 'confirmed', 'Seeded booking for demo')
		ON CONFLICT (office_hour_id, booking_date, start_time)
		DO UPDATE SET purpose = EXCLUDED.purpose, status = EXCLUDED.status, notes = EXCLUDED.notes, updated_at = NOW()`, officeHourID, fixture.StudentID)
	return err
}

func ensureSelfService(ctx context.Context, db *repository.DB, fixture *demoFixture) error {
	_, err := db.Pool.Exec(ctx, `
		INSERT INTO self_service_requests (student_id, college_id, type, title, description, status, document_type, delivery_method, admin_response, responded_by, responded_at)
		SELECT $1, $2, 'document', 'Transcript Copy', 'Need a transcript copy for internship verification.', 'approved', 'transcript', 'email', 'Approved and queued for delivery.', $3, NOW()
		WHERE NOT EXISTS (
			SELECT 1 FROM self_service_requests WHERE student_id = $1 AND title = 'Transcript Copy'
		)`, fixture.StudentID, fixture.CollegeID, fixture.Admin.ID)
	return err
}

func ensureForum(ctx context.Context, db *repository.DB, fixture *demoFixture) error {
	var threadID int
	err := db.Pool.QueryRow(ctx, `SELECT id FROM forum_threads WHERE college_id = $1 AND title = 'How should we structure the service layer?'`, fixture.CollegeID).Scan(&threadID)
	if errorsIsNoRows(err) {
		err = db.Pool.QueryRow(ctx, `
			INSERT INTO forum_threads (college_id, course_id, category, author_id, title, content, is_pinned, is_locked, view_count, reply_count, last_reply_at, last_reply_by, tags)
			VALUES ($1, $2, 'academic', $3, 'How should we structure the service layer?', 'Looking for advice on separating handlers, services, and repositories.', FALSE, FALSE, 12, 1, NOW(), $4, ARRAY['architecture','backend'])
			RETURNING id`, fixture.CollegeID, fixture.CourseID, fixture.Student.ID, fixture.Faculty.ID).Scan(&threadID)
	}
	if err != nil {
		return err
	}
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO forum_replies (thread_id, parent_id, content, author_id, is_accepted_answer, like_count, college_id)
		SELECT $1, NULL, 'Start with handlers for HTTP concerns, services for workflow logic, and repositories for persistence.', $2, TRUE, 5, $3
		WHERE NOT EXISTS (
			SELECT 1 FROM forum_replies WHERE thread_id = $1 AND author_id = $2
		)`, threadID, fixture.Faculty.ID, fixture.CollegeID)
	return err
}

func ensurePlacements(ctx context.Context, db *repository.DB, fixture *demoFixture) error {
	var placementID int
	err := db.Pool.QueryRow(ctx, `SELECT id FROM placements WHERE college_id = $1 AND company_name = 'Acme Systems'`, fixture.CollegeID).Scan(&placementID)
	if errorsIsNoRows(err) {
		deadline := time.Now().Add(10 * 24 * time.Hour)
		driveDate := time.Now().Add(14 * 24 * time.Hour)
		err = db.Pool.QueryRow(ctx, `
			INSERT INTO placements (college_id, company_name, job_title, job_description, job_type, location, is_remote, salary_range_min, salary_range_max, salary_currency, required_skills, eligibility_criteria, application_deadline, drive_date, interview_mode, max_applications, status, created_by)
			VALUES ($1, 'Acme Systems', 'Software Engineer Intern', 'Internship focused on backend systems and internal tooling.', 'internship', 'Bengaluru', TRUE, 1200, 1800, 'USD', ARRAY['Go','React','PostgreSQL'], 'Final-year students with 75% attendance', $2, $3, 'virtual', 50, 'open', $4)
			RETURNING id`, fixture.CollegeID, deadline, driveDate, fixture.Admin.ID).Scan(&placementID)
	}
	if err != nil {
		return err
	}

	var applicationID int
	err = db.Pool.QueryRow(ctx, `SELECT id FROM placement_applications WHERE placement_id = $1 AND student_id = $2`, placementID, fixture.StudentID).Scan(&applicationID)
	if errorsIsNoRows(err) {
		err = db.Pool.QueryRow(ctx, `
			INSERT INTO placement_applications (placement_id, student_id, status, resume_url, cover_letter)
			VALUES ($1, $2, 'shortlisted', 'https://example.com/demo-resume.pdf', 'Excited to contribute to backend platform work.')
			RETURNING id`, placementID, fixture.StudentID).Scan(&applicationID)
	}
	if err != nil {
		return err
	}

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO placement_interviews (application_id, round_number, round_name, scheduled_at, duration_minutes, mode, meeting_link, interviewer_name, interviewer_email, feedback, result)
		SELECT $1, 1, 'Technical Screen', NOW() + INTERVAL '16 days', 45, 'virtual', 'https://meet.example.com/acme-interview', 'Jordan Lee', 'jordan.lee@acme.example', 'Strong fundamentals and communication.', 'pending'
		WHERE NOT EXISTS (SELECT 1 FROM placement_interviews WHERE application_id = $1 AND round_number = 1)`, applicationID)
	return err
}

func ensureWebhooks(ctx context.Context, db *repository.DB, fixture *demoFixture) error {
	var webhookID int
	err := db.Pool.QueryRow(ctx, `SELECT id FROM webhooks WHERE college_id = $1 AND name = 'Demo Events Webhook'`, fixture.CollegeID).Scan(&webhookID)
	if errorsIsNoRows(err) {
		err = db.Pool.QueryRow(ctx, `
			INSERT INTO webhooks (college_id, name, description, url, secret, event_types, is_active, last_triggered_at, last_payload, last_response_status, last_response_body, created_by)
			VALUES ($1, 'Demo Events Webhook', 'Used to demonstrate webhook delivery history', 'https://hooks.example.com/eduhub-demo', 'demo-secret', ARRAY['assignment.created','student.updated'], TRUE, NOW(), '{"event":"assignment.created"}'::jsonb, 200, 'ok', $2)
			RETURNING id`, fixture.CollegeID, fixture.Admin.ID).Scan(&webhookID)
	}
	if err != nil {
		return err
	}
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO webhook_deliveries (webhook_id, event_type, payload, attempt_number, response_status, response_body, delivered_at)
		SELECT $1, 'assignment.created', '{"assignment":"Architecture Case Study"}'::jsonb, 1, 200, 'ok', NOW()
		WHERE NOT EXISTS (SELECT 1 FROM webhook_deliveries WHERE webhook_id = $1 AND event_type = 'assignment.created')`, webhookID)
	return err
}

func ensureAudit(ctx context.Context, db *repository.DB, fixture *demoFixture) error {
	_, err := db.Pool.Exec(ctx, `
		INSERT INTO audit_logs (college_id, user_id, action, entity_type, entity_id, changes, ip_address, user_agent, timestamp)
		SELECT $1, $2, 'CREATE', 'announcement', 1, '{"title":"Demo Week Schedule","priority":"high"}'::jsonb, '127.0.0.1', 'demo-seed', NOW()
		WHERE NOT EXISTS (
			SELECT 1 FROM audit_logs WHERE college_id = $1 AND entity_type = 'announcement' AND entity_id = 1
		)`, fixture.CollegeID, fixture.Admin.ID)
	if err != nil {
		return err
	}
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO audit_stats (college_id, date, total_actions, create_count, update_count, delete_count, login_count, failed_count, unique_users)
		VALUES ($1, CURRENT_DATE, 12, 5, 4, 0, 3, 0, 4)
		ON CONFLICT (college_id, date)
		DO UPDATE SET total_actions = EXCLUDED.total_actions,
		              create_count = EXCLUDED.create_count,
		              update_count = EXCLUDED.update_count,
		              login_count = EXCLUDED.login_count,
		              unique_users = EXCLUDED.unique_users,
		              updated_at = NOW()`, fixture.CollegeID)
	return err
}

func ensureNotifications(ctx context.Context, db *repository.DB, fixture *demoFixture) error {
	notifications := []struct {
		userID  int
		title   string
		message string
		typeVal string
	}{
		{fixture.Admin.ID, "System Health Green", "All demo services are responding normally.", "success"},
		{fixture.Faculty.ID, "Review Pending Submissions", "One assignment submission is ready for review.", "info"},
		{fixture.Student.ID, "New Exam Result Available", "Your practical exam result has been published.", "success"},
		{fixture.Parent.ID, "Attendance Update", "Your linked student has an updated attendance summary.", "info"},
	}
	for _, item := range notifications {
		var existingID int
		err := db.Pool.QueryRow(ctx,
			`SELECT id FROM notifications WHERE college_id = $1 AND user_id = $2 AND title = $3 LIMIT 1`,
			fixture.CollegeID, item.userID, item.title,
		).Scan(&existingID)
		if err == nil {
			continue
		}
		if !errorsIsNoRows(err) {
			return err
		}
		_, err = db.Pool.Exec(ctx, `
			INSERT INTO notifications (college_id, user_id, title, message, type, is_read, is_published, published_at, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, FALSE, TRUE, NOW(), NOW(), NOW())`,
			fixture.CollegeID, item.userID, item.title, item.message, item.typeVal,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func ensureFileArtifacts(ctx context.Context, db *repository.DB, fixture *demoFixture) (int, error) {
	storageCfg, err := config.LoadStorageConfig()
	if err != nil {
		return 0, err
	}
	minioClient, err := storageclient.NewMinioClient(&storageclient.MinioConfig{
		Endpoint:  storageCfg.Endpoint,
		AccessKey: storageCfg.AccessKey,
		SecretKey: storageCfg.SecretKey,
		UseSSL:    storageCfg.UseSSL,
		Bucket:    storageCfg.Bucket,
		Region:    storageCfg.Region,
	})
	if err != nil {
		return 0, err
	}
	fileRepo := repository.NewFileRepository(db)
	storageService := storagesvc.NewStorageService(minioClient.Client(), storageCfg.Bucket, storageCfg.Endpoint, storageCfg.UseSSL)
	fileService := filesvc.NewFileService(fileRepo, storageService)

	var folderID int
	err = db.Pool.QueryRow(ctx, `SELECT id FROM folders WHERE college_id = $1 AND (path = 'demo' OR path = '/demo') LIMIT 1`, fixture.CollegeID).Scan(&folderID)
	if errorsIsNoRows(err) {
		folder, createErr := fileService.CreateFolder(ctx, fixture.CollegeID, fixture.Admin.ID, "demo", nil)
		if createErr != nil {
			return 0, createErr
		}
		folderID = folder.ID
	} else if err != nil {
		return 0, err
	}

	var fileID int
	err = db.Pool.QueryRow(ctx, `SELECT id FROM files WHERE college_id = $1 AND name = 'demo-handbook'`, fixture.CollegeID).Scan(&fileID)
	if err == nil {
		return fileID, nil
	}
	if !errorsIsNoRows(err) {
		return 0, err
	}

	content := []byte("EduHub demo handbook\n\nThis file is seeded to validate the files module and MinIO integration.\n")
	uploaded, err := fileService.UploadFile(
		ctx,
		fixture.CollegeID,
		fixture.Admin.ID,
		bytes.NewReader(content),
		"demo-handbook.txt",
		"text/plain",
		int64(len(content)),
		"document",
		"Seeded handbook for file previews",
		&folderID,
		[]string{"demo", "handbook"},
	)
	if err != nil {
		return 0, err
	}

	moduleID, err := ensureCourseModule(ctx, db, fixture, uploaded.ID)
	if err != nil {
		return 0, err
	}
	if err := ensureCourseMaterial(ctx, db, fixture, moduleID, uploaded.ID); err != nil {
		return 0, err
	}

	return uploaded.ID, nil
}

func ensureCourseModule(ctx context.Context, db *repository.DB, fixture *demoFixture, fileID int) (int, error) {
	var moduleID int
	err := db.Pool.QueryRow(ctx, `SELECT id FROM course_modules WHERE course_id = $1 AND title = 'Week 1'`, fixture.CourseID).Scan(&moduleID)
	if errorsIsNoRows(err) {
		err = db.Pool.QueryRow(ctx, `
			INSERT INTO course_modules (course_id, title, description, display_order, is_published, college_id)
			VALUES ($1, 'Week 1', 'Seeded module for the course materials flow', 1, TRUE, $2)
			RETURNING id`, fixture.CourseID, fixture.CollegeID).Scan(&moduleID)
	}
	return moduleID, err
}

func ensureCourseMaterial(ctx context.Context, db *repository.DB, fixture *demoFixture, moduleID, fileID int) error {
	_, err := db.Pool.Exec(ctx, `
		INSERT INTO course_materials (course_id, title, description, type, file_id, module_id, display_order, is_published, published_at, uploaded_by, college_id)
		SELECT $1, 'Demo Handbook', 'Seeded file-backed course material.', 'document', $2, $3, 1, TRUE, NOW(), $4, $5
		WHERE NOT EXISTS (
			SELECT 1 FROM course_materials WHERE course_id = $1 AND title = 'Demo Handbook'
		)`, fixture.CourseID, fileID, moduleID, fixture.Admin.ID, fixture.CollegeID)
	return err
}

func verifyDemoLogins() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("load config for login verification: %w", err)
	}
	defer cfg.DB.Close()

	svcs := services.NewServices(cfg)
	defer svcs.WebSocketService.Stop()

	e := echo.New()
	e.POST("/auth/login", handler.NewAuthHandler(svcs.Auth).DirectLogin)

	for _, account := range demoAccounts {
		if err := verifyDemoLogin(e, account); err != nil {
			return err
		}
	}

	return nil
}

func verifyDemoLogin(e *echo.Echo, account demoAccountSpec) error {
	payload, err := json.Marshal(map[string]string{
		"email":    account.Email,
		"password": demoDefaultPassword,
	})
	if err != nil {
		return fmt.Errorf("marshal login payload for %s: %w", account.Email, err)
	}

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		return fmt.Errorf("POST /auth/login failed for %s with HTTP %d: %s", account.Email, rec.Code, strings.TrimSpace(rec.Body.String()))
	}

	var response struct {
		Success bool `json:"success"`
		Data    struct {
			Token string `json:"token"`
			User  struct {
				Email     string `json:"email"`
				Role      string `json:"role"`
				CollegeID string `json:"collegeId"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		return fmt.Errorf("decode login response for %s: %w", account.Email, err)
	}

	if !response.Success || response.Data.Token == "" {
		return fmt.Errorf("POST /auth/login returned an incomplete session for %s", account.Email)
	}
	if response.Data.User.Email != account.Email {
		return fmt.Errorf("POST /auth/login returned email %q for %s", response.Data.User.Email, account.Email)
	}
	if response.Data.User.Role != account.Role {
		return fmt.Errorf("POST /auth/login returned role %q for %s, expected %q", response.Data.User.Role, account.Email, account.Role)
	}
	if response.Data.User.CollegeID != demoCollegeExternalID {
		return fmt.Errorf("POST /auth/login returned college %q for %s, expected %q", response.Data.User.CollegeID, account.Email, demoCollegeExternalID)
	}

	return nil
}

func errorsIsNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
