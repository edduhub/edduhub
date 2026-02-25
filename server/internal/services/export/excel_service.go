package export

import (
	"context"
	"encoding/csv"
	"fmt"
	"strconv"

	"eduhub/server/internal/models"
)

// ExcelExportService defines the interface for Excel/CSV export operations
type ExcelExportService interface {
	ExportStudentsToExcel(ctx context.Context, students []*models.Student, filename string) ([]byte, error)
	ExportGradesToExcel(ctx context.Context, grades []*models.Grade, filename string) ([]byte, error)
	ExportAttendanceToExcel(ctx context.Context, attendance []*models.Attendance, filename string) ([]byte, error)
	ExportCoursesToExcel(ctx context.Context, courses []*models.Course, filename string) ([]byte, error)
	ExportFeePaymentsToExcel(ctx context.Context, payments []map[string]any, filename string) ([]byte, error)
	ExportEnrollmentToExcel(ctx context.Context, enrollments []*models.Enrollment, filename string) ([]byte, error)
}

// csvExportService implements ExcelExportService using CSV format (Excel compatible)
type csvExportService struct{}

// NewExcelExportService creates a new Excel export service instance
func NewExcelExportService() ExcelExportService {
	return &csvExportService{}
}

// ExportStudentsToExcel exports student data to CSV format (Excel compatible)
func (s *csvExportService) ExportStudentsToExcel(ctx context.Context, students []*models.Student, filename string) ([]byte, error) {
	var buf []byte
	writer := csv.NewWriter(&byteWriter{&buf})

	headers := []string{"Student ID", "User ID", "College ID", "Roll No", "Enrollment Year", "Active", "Created At", "Updated At"}
	writer.Write(headers)

	for _, student := range students {
		record := []string{
			strconv.Itoa(student.StudentID),
			strconv.Itoa(student.UserID),
			strconv.Itoa(student.CollegeID),
			student.RollNo,
			strconv.Itoa(student.EnrollmentYear),
			strconv.FormatBool(student.IsActive),
			student.CreatedAt.Format("2006-01-02 15:04:05"),
			student.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		writer.Write(record)
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("failed to write CSV: %w", err)
	}

	return buf, nil
}

// ExportGradesToExcel exports grade data to CSV format (Excel compatible)
func (s *csvExportService) ExportGradesToExcel(ctx context.Context, grades []*models.Grade, filename string) ([]byte, error) {
	var buf []byte
	writer := csv.NewWriter(&byteWriter{&buf})

	headers := []string{"Grade ID", "Student ID", "Course ID", "Assessment Type", "Assessment Name", "Obtained Marks", "Total Marks", "Percentage", "Grade", "Graded At"}
	writer.Write(headers)

	for _, grade := range grades {
		gradedAt := ""
		if grade.GradedAt != nil {
			gradedAt = grade.GradedAt.Format("2006-01-02")
		}
		gradeValue := ""
		if grade.Grade != nil {
			gradeValue = *grade.Grade
		}

		record := []string{
			strconv.Itoa(grade.ID),
			strconv.Itoa(grade.StudentID),
			strconv.Itoa(grade.CourseID),
			grade.AssessmentType,
			grade.AssessmentName,
			strconv.Itoa(grade.ObtainedMarks),
			strconv.Itoa(grade.TotalMarks),
			fmt.Sprintf("%.2f", grade.Percentage),
			gradeValue,
			gradedAt,
		}
		writer.Write(record)
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("failed to write CSV: %w", err)
	}

	return buf, nil
}

// ExportAttendanceToExcel exports attendance data to CSV format (Excel compatible)
func (s *csvExportService) ExportAttendanceToExcel(ctx context.Context, attendance []*models.Attendance, filename string) ([]byte, error) {
	var buf []byte
	writer := csv.NewWriter(&byteWriter{&buf})

	headers := []string{"Attendance ID", "Student ID", "Course ID", "Lecture ID", "Date", "Status", "Scanned At"}
	writer.Write(headers)

	for _, record := range attendance {
		csvRecord := []string{
			strconv.Itoa(record.ID),
			strconv.Itoa(record.StudentID),
			strconv.Itoa(record.CourseID),
			strconv.Itoa(record.LectureID),
			record.Date.Format("2006-01-02"),
			record.Status,
			record.ScannedAt.Format("2006-01-02 15:04:05"),
		}
		writer.Write(csvRecord)
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("failed to write CSV: %w", err)
	}

	return buf, nil
}

// ExportCoursesToExcel exports course data to CSV format (Excel compatible)
func (s *csvExportService) ExportCoursesToExcel(ctx context.Context, courses []*models.Course, filename string) ([]byte, error) {
	var buf []byte
	writer := csv.NewWriter(&byteWriter{&buf})

	headers := []string{"Course ID", "Name", "College ID", "Description", "Credits", "Instructor ID", "Created At", "Updated At"}
	writer.Write(headers)

	for _, course := range courses {
		record := []string{
			strconv.Itoa(course.ID),
			course.Name,
			strconv.Itoa(course.CollegeID),
			course.Description,
			strconv.Itoa(course.Credits),
			strconv.Itoa(course.InstructorID),
			course.CreatedAt.Format("2006-01-02"),
			course.UpdatedAt.Format("2006-01-02"),
		}
		writer.Write(record)
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("failed to write CSV: %w", err)
	}

	return buf, nil
}

// ExportFeePaymentsToExcel exports fee payment data to CSV format (Excel compatible)
func (s *csvExportService) ExportFeePaymentsToExcel(ctx context.Context, payments []map[string]any, filename string) ([]byte, error) {
	var buf []byte
	writer := csv.NewWriter(&byteWriter{&buf})

	headers := []string{"Payment ID", "Student ID", "Amount", "Payment Date", "Status", "Payment Method", "Transaction ID"}
	writer.Write(headers)

	for _, payment := range payments {
		record := []string{
			fmt.Sprintf("%v", payment["id"]),
			fmt.Sprintf("%v", payment["student_id"]),
			fmt.Sprintf("%v", payment["amount"]),
			fmt.Sprintf("%v", payment["payment_date"]),
			fmt.Sprintf("%v", payment["status"]),
			fmt.Sprintf("%v", payment["payment_method"]),
			fmt.Sprintf("%v", payment["transaction_id"]),
		}
		writer.Write(record)
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("failed to write CSV: %w", err)
	}

	return buf, nil
}

// ExportEnrollmentToExcel exports enrollment data to CSV format (Excel compatible)
func (s *csvExportService) ExportEnrollmentToExcel(ctx context.Context, enrollments []*models.Enrollment, filename string) ([]byte, error) {
	var buf []byte
	writer := csv.NewWriter(&byteWriter{&buf})

	headers := []string{"Enrollment ID", "Student ID", "Course ID", "College ID", "Enrollment Date", "Status", "Grade"}
	writer.Write(headers)

	for _, enrollment := range enrollments {
		record := []string{
			strconv.Itoa(enrollment.ID),
			strconv.Itoa(enrollment.StudentID),
			strconv.Itoa(enrollment.CourseID),
			strconv.Itoa(enrollment.CollegeID),
			enrollment.EnrollmentDate.Format("2006-01-02"),
			enrollment.Status,
			enrollment.Grade,
		}
		writer.Write(record)
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("failed to write CSV: %w", err)
	}

	return buf, nil
}

// byteWriter implements io.Writer for []byte
type byteWriter struct {
	buf *[]byte
}

func (w *byteWriter) Write(p []byte) (n int, err error) {
	*w.buf = append(*w.buf, p...)
	return len(p), nil
}
