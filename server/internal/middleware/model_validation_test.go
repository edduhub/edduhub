package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests exercise ValidateStruct against structs that mirror the project's
// model update-request types with their validate tags.

// --- Course-like validation ---

type testCourse struct {
	Name         string `json:"name" validate:"required,minlen=3,maxlen=100"`
	CollegeID    int    `json:"college_id" validate:"required,gte=1"`
	Credits      int    `json:"credits" validate:"required,gte=1,lte=5"`
	InstructorID int    `json:"instructor_id" validate:"required,gte=1"`
}

func TestValidateModel_Course(t *testing.T) {
	t.Run("valid course passes", func(t *testing.T) {
		c := testCourse{Name: "Data Structures", CollegeID: 1, Credits: 3, InstructorID: 1}
		assert.NoError(t, ValidateStruct(&c))
	})

	t.Run("fails when name is empty", func(t *testing.T) {
		c := testCourse{Name: "", CollegeID: 1, Credits: 3, InstructorID: 1}
		require.Error(t, ValidateStruct(&c))
	})

	t.Run("fails when name too short", func(t *testing.T) {
		c := testCourse{Name: "DS", CollegeID: 1, Credits: 3, InstructorID: 1}
		require.Error(t, ValidateStruct(&c))
	})

	t.Run("fails when credits out of range", func(t *testing.T) {
		c := testCourse{Name: "Data Structures", CollegeID: 1, Credits: 6, InstructorID: 1}
		require.Error(t, ValidateStruct(&c))
	})

	t.Run("fails when college_id is zero", func(t *testing.T) {
		c := testCourse{Name: "Data Structures", CollegeID: 0, Credits: 3, InstructorID: 1}
		require.Error(t, ValidateStruct(&c))
	})
}

// --- Update request (pointer fields) ---

type testUpdateCourseRequest struct {
	Name    *string `json:"name" validate:"omitempty,minlen=3,maxlen=100"`
	Credits *int    `json:"credits" validate:"omitempty,gte=1,lte=5"`
}

func TestValidateModel_UpdateCourseRequest(t *testing.T) {
	t.Run("nil fields are skipped", func(t *testing.T) {
		req := testUpdateCourseRequest{}
		assert.NoError(t, ValidateStruct(&req))
	})

	t.Run("valid name update", func(t *testing.T) {
		name := "New Name"
		req := testUpdateCourseRequest{Name: &name}
		assert.NoError(t, ValidateStruct(&req))
	})

	t.Run("name too short", func(t *testing.T) {
		name := "AB"
		req := testUpdateCourseRequest{Name: &name}
		require.Error(t, ValidateStruct(&req))
	})

	t.Run("valid credits", func(t *testing.T) {
		credits := 4
		req := testUpdateCourseRequest{Credits: &credits}
		assert.NoError(t, ValidateStruct(&req))
	})

	t.Run("invalid credits", func(t *testing.T) {
		credits := 10
		req := testUpdateCourseRequest{Credits: &credits}
		require.Error(t, ValidateStruct(&req))
	})
}

// --- User update request ---

type testUpdateUserRequest struct {
	Name             *string `json:"name" validate:"omitempty,min=1,max=100"`
	Email            *string `json:"email" validate:"omitempty,email"`
	KratosIdentityID *string `json:"kratos_identity_id" validate:"omitempty,len=36"`
}

func TestValidateModel_UpdateUserRequest(t *testing.T) {
	t.Run("nil fields are skipped", func(t *testing.T) {
		assert.NoError(t, ValidateStruct(&testUpdateUserRequest{}))
	})

	t.Run("valid email", func(t *testing.T) {
		email := "user@example.com"
		assert.NoError(t, ValidateStruct(&testUpdateUserRequest{Email: &email}))
	})

	t.Run("invalid email", func(t *testing.T) {
		email := "not-an-email"
		require.Error(t, ValidateStruct(&testUpdateUserRequest{Email: &email}))
	})

	t.Run("valid kratos ID length", func(t *testing.T) {
		id := "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
		assert.NoError(t, ValidateStruct(&testUpdateUserRequest{KratosIdentityID: &id}))
	})

	t.Run("invalid kratos ID length", func(t *testing.T) {
		id := "too-short"
		require.Error(t, ValidateStruct(&testUpdateUserRequest{KratosIdentityID: &id}))
	})
}

// --- Student update request ---

type testUpdateStudentRequest struct {
	CollegeID      *int    `json:"college_id" validate:"omitempty,gte=1"`
	EnrollmentYear *int    `json:"enrollment_year" validate:"omitempty,gte=1947"`
	RollNo         *string `json:"roll_no" validate:"omitempty,min=1,max=50"`
}

func TestValidateModel_UpdateStudentRequest(t *testing.T) {
	t.Run("nil fields are skipped", func(t *testing.T) {
		assert.NoError(t, ValidateStruct(&testUpdateStudentRequest{}))
	})

	t.Run("valid enrollment year", func(t *testing.T) {
		year := 2020
		assert.NoError(t, ValidateStruct(&testUpdateStudentRequest{EnrollmentYear: &year}))
	})

	t.Run("invalid enrollment year", func(t *testing.T) {
		year := 1900
		require.Error(t, ValidateStruct(&testUpdateStudentRequest{EnrollmentYear: &year}))
	})

	t.Run("invalid college ID zero", func(t *testing.T) {
		cid := 0
		require.Error(t, ValidateStruct(&testUpdateStudentRequest{CollegeID: &cid}))
	})
}

// --- Quiz update request ---

type testUpdateQuizRequest struct {
	Title            *string `json:"title" validate:"omitempty,min=1,max=100"`
	TimeLimitMinutes *int    `json:"time_limit_minutes" validate:"omitempty,gte=0"`
}

func TestValidateModel_UpdateQuizRequest(t *testing.T) {
	t.Run("nil fields are skipped", func(t *testing.T) {
		assert.NoError(t, ValidateStruct(&testUpdateQuizRequest{}))
	})

	t.Run("valid title", func(t *testing.T) {
		title := "Quiz 1"
		assert.NoError(t, ValidateStruct(&testUpdateQuizRequest{Title: &title}))
	})

	t.Run("valid time limit", func(t *testing.T) {
		tl := 30
		assert.NoError(t, ValidateStruct(&testUpdateQuizRequest{TimeLimitMinutes: &tl}))
	})
}

// --- Question update with oneof ---

type testQuizType string

type testUpdateQuestionRequest struct {
	Type   *testQuizType `json:"type" validate:"omitempty,oneof=MultipleChoice TrueFalse ShortAnswer"`
	Points *int          `json:"points" validate:"omitempty,gte=0,lte=100"`
}

func TestValidateModel_UpdateQuestionRequest(t *testing.T) {
	t.Run("nil fields are skipped", func(t *testing.T) {
		assert.NoError(t, ValidateStruct(&testUpdateQuestionRequest{}))
	})

	t.Run("valid type", func(t *testing.T) {
		qt := testQuizType("MultipleChoice")
		assert.NoError(t, ValidateStruct(&testUpdateQuestionRequest{Type: &qt}))
	})

	t.Run("invalid type", func(t *testing.T) {
		qt := testQuizType("FillInBlank")
		require.Error(t, ValidateStruct(&testUpdateQuestionRequest{Type: &qt}))
	})

	t.Run("valid points", func(t *testing.T) {
		p := 10
		assert.NoError(t, ValidateStruct(&testUpdateQuestionRequest{Points: &p}))
	})

	t.Run("points out of range", func(t *testing.T) {
		p := 200
		require.Error(t, ValidateStruct(&testUpdateQuestionRequest{Points: &p}))
	})
}

// --- StudentAttendanceStatus-like ---

type testStudentAttendanceStatus struct {
	StudentID int    `json:"student_id" validate:"required,gt=0"`
	Status    string `json:"status" validate:"required,oneof=Present Absent"`
}

func TestValidateModel_StudentAttendanceStatus(t *testing.T) {
	t.Run("valid present", func(t *testing.T) {
		assert.NoError(t, ValidateStruct(&testStudentAttendanceStatus{StudentID: 1, Status: "Present"}))
	})

	t.Run("valid absent", func(t *testing.T) {
		assert.NoError(t, ValidateStruct(&testStudentAttendanceStatus{StudentID: 2, Status: "Absent"}))
	})

	t.Run("invalid status", func(t *testing.T) {
		require.Error(t, ValidateStruct(&testStudentAttendanceStatus{StudentID: 1, Status: "Late"}))
	})

	t.Run("missing student_id", func(t *testing.T) {
		require.Error(t, ValidateStruct(&testStudentAttendanceStatus{StudentID: 0, Status: "Present"}))
	})

	t.Run("missing status", func(t *testing.T) {
		require.Error(t, ValidateStruct(&testStudentAttendanceStatus{StudentID: 1, Status: ""}))
	})
}

// --- QuizAttempt-like ---

type testQuizAttempt struct {
	StudentID int `json:"student_id" validate:"required"`
	QuizID    int `json:"quiz_id" validate:"required"`
	CollegeID int `json:"college_id" validate:"required"`
	CourseID  int `json:"course_id" validate:"required"`
}

func TestValidateModel_QuizAttempt(t *testing.T) {
	t.Run("valid attempt", func(t *testing.T) {
		qa := testQuizAttempt{StudentID: 1, QuizID: 1, CollegeID: 1, CourseID: 1}
		assert.NoError(t, ValidateStruct(&qa))
	})
}
