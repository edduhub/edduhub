package batch

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"
)

type BatchResult struct {
	Success int      `json:"success"`
	Failed  int      `json:"failed"`
	Errors  []string `json:"errors,omitempty"`
}

type BatchService interface {
	ImportStudents(ctx context.Context, collegeID int, students []models.Student) (*BatchResult, error)
	ExportStudents(ctx context.Context, collegeID int, courseID *int) (string, error)
	ImportGrades(ctx context.Context, collegeID, courseID int, records [][]string) (*BatchResult, error)
	ExportGrades(ctx context.Context, collegeID, courseID int) (string, error)
	BulkEnroll(ctx context.Context, collegeID, courseID int, studentIDs []int) (*BatchResult, error)
}

type batchService struct {
	studentRepo    repository.StudentRepository
	enrollmentRepo repository.EnrollmentRepository
	gradeRepo      repository.GradeRepository
}

func NewBatchService(
	studentRepo repository.StudentRepository,
	enrollmentRepo repository.EnrollmentRepository,
	gradeRepo repository.GradeRepository,
) BatchService {
	return &batchService{
		studentRepo:    studentRepo,
		enrollmentRepo: enrollmentRepo,
		gradeRepo:      gradeRepo,
	}
}

func (s *batchService) ImportStudents(ctx context.Context, collegeID int, students []models.Student) (*BatchResult, error) {
	result := &BatchResult{}

	for _, student := range students {
		student.CollegeID = collegeID
		err := s.studentRepo.CreateStudent(ctx, &student)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to create student %s: %v", student.RollNo, err))
		} else {
			result.Success++
		}
	}

	return result, nil
}

func (s *batchService) ExportStudents(ctx context.Context, collegeID int, courseID *int) (string, error) {
	var students []*models.Student
	const pageSize uint64 = 500
	offset := uint64(0)

	for {
		var batch []*models.Student
		var err error

		if courseID != nil {
			enrollments, err := s.enrollmentRepo.FindEnrollmentsByCourse(ctx, collegeID, *courseID, pageSize, offset)
			if err != nil {
				return "", fmt.Errorf("ExportStudents: failed to fetch enrollments: %w", err)
			}
			if len(enrollments) == 0 {
				break
			}
			for _, enrollment := range enrollments {
				student, err := s.studentRepo.GetStudentByID(ctx, collegeID, enrollment.StudentID)
				if err != nil {
					return "", fmt.Errorf("ExportStudents: failed to load student %d: %w", enrollment.StudentID, err)
				}
				batch = append(batch, student)
			}
		} else {
			batch, err = s.studentRepo.FindAllStudentsByCollege(ctx, collegeID, pageSize, offset)
			if err != nil {
				return "", fmt.Errorf("ExportStudents: failed to fetch students: %w", err)
			}
			if len(batch) == 0 {
				break
			}
		}

		students = append(students, batch...)
		offset += pageSize
		if len(batch) < int(pageSize) {
			break
		}
	}

	buf := &bytes.Buffer{}
	buf.WriteString("Student ID,Roll Number,Active\n")
	for _, student := range students {
		status := "Inactive"
		if student.IsActive {
			status = "Active"
		}
		line := fmt.Sprintf("%d,%s,%s\n", student.StudentID, student.RollNo, status)
		buf.WriteString(line)
	}

	return buf.String(), nil
}

func (s *batchService) ImportGrades(ctx context.Context, collegeID, courseID int, records [][]string) (*BatchResult, error) {
	result := &BatchResult{}

	if len(records) <= 1 {
		return &BatchResult{Failed: 1, Errors: []string{"CSV file contains no data"}}, nil
	}

	for i, record := range records[1:] {
		line := i + 2
		if len(record) < 5 {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("line %d: expected at least 5 columns", line))
			continue
		}

		studentID, err := strconv.Atoi(strings.TrimSpace(record[0]))
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("line %d: invalid student_id", line))
			continue
		}

		assessmentName := strings.TrimSpace(record[1])
		assessmentType := strings.TrimSpace(record[2])
		totalMarks, err := strconv.Atoi(strings.TrimSpace(record[3]))
		if err != nil || totalMarks <= 0 {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("line %d: invalid total_marks", line))
			continue
		}

		obtainedMarks, err := strconv.Atoi(strings.TrimSpace(record[4]))
		if err != nil || obtainedMarks < 0 {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("line %d: invalid obtained_marks", line))
			continue
		}

		gradeValue := ""
		if len(record) > 5 {
			gradeValue = strings.TrimSpace(record[5])
		}

		remarks := ""
		if len(record) > 6 {
			remarks = strings.TrimSpace(record[6])
		}

		percentage := float64(obtainedMarks) / float64(totalMarks) * 100
		percentage = float64(int(percentage*100+0.5)) / 100

		grade := &models.Grade{
			StudentID:      studentID,
			CourseID:       courseID,
			CollegeID:      collegeID,
			AssessmentName: assessmentName,
			AssessmentType: assessmentType,
			TotalMarks:     totalMarks,
			ObtainedMarks:  obtainedMarks,
			Percentage:     percentage,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		if gradeValue != "" {
			grade.Grade = &gradeValue
		}
		if remarks != "" {
			grade.Remarks = &remarks
		}

		if err := s.gradeRepo.CreateGrade(ctx, grade); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("line %d: %v", line, err))
			continue
		}

		result.Success++
	}

	return result, nil
}

func (s *batchService) ExportGrades(ctx context.Context, collegeID, courseID int) (string, error) {
	grades, err := s.gradeRepo.GetGradesByCourse(ctx, collegeID, courseID)
	if err != nil {
		return "", fmt.Errorf("ExportGrades: failed to fetch grades: %w", err)
	}

	buf := &bytes.Buffer{}
	buf.WriteString("Student ID,Assessment,Type,Percentage,Grade,Remarks\n")
	for _, grade := range grades {
		gradeLetter := ""
		if grade.Grade != nil {
			gradeLetter = *grade.Grade
		}
		remarks := ""
		if grade.Remarks != nil {
			remarks = *grade.Remarks
		}
		line := fmt.Sprintf("%d,%s,%s,%.2f,%s,%s\n",
			grade.StudentID,
			escapeCSV(grade.AssessmentName),
			grade.AssessmentType,
			grade.Percentage,
			escapeCSV(gradeLetter),
			escapeCSV(remarks),
		)
		buf.WriteString(line)
	}

	return buf.String(), nil
}

func (s *batchService) BulkEnroll(ctx context.Context, collegeID, courseID int, studentIDs []int) (*BatchResult, error) {
	result := &BatchResult{}

	for _, studentID := range studentIDs {
		isEnrolled, err := s.enrollmentRepo.IsStudentEnrolled(ctx, collegeID, studentID, courseID)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to validate enrollment for student %d: %v", studentID, err))
			continue
		}
		if isEnrolled {
			continue
		}
		enrollment := &models.Enrollment{
			StudentID:      studentID,
			CourseID:       courseID,
			CollegeID:      collegeID,
			Status:         models.Active,
			EnrollmentDate: time.Now(),
		}

		err = s.enrollmentRepo.CreateEnrollment(ctx, enrollment)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to enroll student %d: %v", studentID, err))
		} else {
			result.Success++
		}
	}

	return result, nil
}

func escapeCSV(value string) string {
	if strings.ContainsAny(value, ",\n\r\"") {
		return fmt.Sprintf("\"%s\"", strings.ReplaceAll(value, "\"", "\"\""))
	}
	return value
}
