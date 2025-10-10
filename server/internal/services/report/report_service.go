package report

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"

	"github.com/jung-kurt/gofpdf"
)

type ReportService interface {
	GenerateGradeCard(ctx context.Context, collegeID, studentID int, semester *int) ([]byte, error)
	GenerateTranscript(ctx context.Context, collegeID, studentID int) ([]byte, error)
	GenerateAttendanceReport(ctx context.Context, collegeID, courseID int) ([]byte, error)
	GenerateCourseReport(ctx context.Context, collegeID, courseID int) ([]byte, error)
}

type reportService struct {
	studentRepo    repository.StudentRepository
	gradeRepo      repository.GradeRepository
	attendanceRepo repository.AttendanceRepository
	enrollmentRepo repository.EnrollmentRepository
	courseRepo     repository.CourseRepository
}

func NewReportService(
	studentRepo repository.StudentRepository,
	gradeRepo repository.GradeRepository,
	attendanceRepo repository.AttendanceRepository,
	enrollmentRepo repository.EnrollmentRepository,
	courseRepo repository.CourseRepository,
) ReportService {
	return &reportService{
		studentRepo:    studentRepo,
		gradeRepo:      gradeRepo,
		attendanceRepo: attendanceRepo,
		enrollmentRepo: enrollmentRepo,
		courseRepo:     courseRepo,
	}
}

func (s *reportService) GenerateGradeCard(ctx context.Context, collegeID, studentID int, semester *int) ([]byte, error) {
	student, err := s.studentRepo.GetStudentByID(ctx, collegeID, studentID)
	if err != nil {
		return nil, fmt.Errorf("GenerateGradeCard: failed to load student: %w", err)
	}
	if student == nil {
		return nil, fmt.Errorf("GenerateGradeCard: student %d not found", studentID)
	}

	grades, err := s.gradeRepo.GetGradesByStudent(ctx, collegeID, studentID)
	if err != nil {
		return nil, fmt.Errorf("GenerateGradeCard: failed to load grades: %w", err)
	}

	pdf := newPDF("Grade Card")
	addHeader(pdf, fmt.Sprintf("Grade Card - Student #%d", studentID))
	pdf.Ln(4)
	pdf.SetFont("Helvetica", "", 10)
	pdf.MultiCell(0, 5, fmt.Sprintf("Roll Number: %s\nCollege ID: %d", student.RollNo, collegeID), gofpdf.BorderNone, gofpdf.AlignLeft, false)

	if len(grades) == 0 {
		pdf.Ln(6)
		pdf.CellFormat(0, 6, "No grades available", gofpdf.BorderNone, 1, gofpdf.AlignLeft, false, 0, "")
		return outputPDF(pdf)
	}

	pdf.Ln(4)
	drawTableHeader(pdf, []string{"Assessment", "Course ID", "Type", "Obtained", "Total", "Percentage", "Grade"})
	for _, grade := range grades {
		percentage := fmt.Sprintf("%.2f%%", grade.Percentage)
		gradeLetter := "-"
		if grade.Grade != nil {
			gradeLetter = *grade.Grade
		}
		row := []string{
			grade.AssessmentName,
			fmt.Sprintf("%d", grade.CourseID),
			grade.AssessmentType,
			fmt.Sprintf("%d", grade.ObtainedMarks),
			fmt.Sprintf("%d", grade.TotalMarks),
			percentage,
			gradeLetter,
		}
		drawTableRow(pdf, row)
	}

	return outputPDF(pdf)
}

func (s *reportService) GenerateTranscript(ctx context.Context, collegeID, studentID int) ([]byte, error) {
	student, err := s.studentRepo.GetStudentByID(ctx, collegeID, studentID)
	if err != nil {
		return nil, fmt.Errorf("GenerateTranscript: failed to load student: %w", err)
	}
	if student == nil {
		return nil, fmt.Errorf("GenerateTranscript: student %d not found", studentID)
	}

	grades, err := s.gradeRepo.GetGradesByStudent(ctx, collegeID, studentID)
	if err != nil {
		return nil, fmt.Errorf("GenerateTranscript: failed to fetch grades: %w", err)
	}

	byCourse := make(map[int][]*models.Grade)
	for _, grade := range grades {
		byCourse[grade.CourseID] = append(byCourse[grade.CourseID], grade)
	}

	pdf := newPDF("Transcript")
	addHeader(pdf, fmt.Sprintf("Official Transcript - Student #%d", studentID))
	pdf.Ln(4)
	pdf.MultiCell(0, 5, fmt.Sprintf("Roll Number: %s\nIssued On: %s", student.RollNo, time.Now().Format("02 Jan 2006")), gofpdf.BorderNone, gofpdf.AlignLeft, false)

	if len(grades) == 0 {
		pdf.Ln(6)
		pdf.CellFormat(0, 6, "No academic records available", gofpdf.BorderNone, 1, gofpdf.AlignLeft, false, 0, "")
		return outputPDF(pdf)
	}

	courseIDs := make([]int, 0, len(byCourse))
	for courseID := range byCourse {
		courseIDs = append(courseIDs, courseID)
	}
	sort.Ints(courseIDs)

	for _, courseID := range courseIDs {
		courseGrades := byCourse[courseID]
		pdf.Ln(6)
		pdf.SetFont("Helvetica", "B", 11)
		pdf.CellFormat(0, 6, fmt.Sprintf("Course #%d", courseID), gofpdf.BorderNone, 1, gofpdf.AlignLeft, false, 0, "")
		drawTableHeader(pdf, []string{"Assessment", "Type", "Percentage", "Grade"})
		var total float64
		for _, grade := range courseGrades {
			total += grade.Percentage
			gradeLetter := "-"
			if grade.Grade != nil {
				gradeLetter = *grade.Grade
			}
			row := []string{
				grade.AssessmentName,
				grade.AssessmentType,
				fmt.Sprintf("%.2f%%", grade.Percentage),
				gradeLetter,
			}
			drawTableRow(pdf, row)
		}
		avg := total / float64(len(courseGrades))
		pdf.SetFont("Helvetica", "I", 9)
		pdf.CellFormat(0, 5, fmt.Sprintf("Course Average: %.2f%%", avg), gofpdf.BorderNone, 1, gofpdf.AlignRight, false, 0, "")
	}

	return outputPDF(pdf)
}

func (s *reportService) GenerateAttendanceReport(ctx context.Context, collegeID, courseID int) ([]byte, error) {
	records, err := s.attendanceRepo.GetAttendanceByCourse(ctx, collegeID, courseID, 10000, 0)
	if err != nil {
		return nil, fmt.Errorf("GenerateAttendanceReport: failed to load attendance: %w", err)
	}

	if len(records) == 0 {
		return generateSimplePDF(fmt.Sprintf("Attendance Report for Course %d", courseID), []string{"No attendance records found"})
	}

	perStudent := make(map[int]struct {
		present int
		total   int
	})
	for _, record := range records {
		entry := perStudent[record.StudentID]
		if strings.EqualFold(record.Status, "present") {
			entry.present++
		}
		entry.total++
		perStudent[record.StudentID] = entry
	}

	pdf := newPDF("Attendance Report")
	addHeader(pdf, fmt.Sprintf("Attendance Report - Course #%d", courseID))
	pdf.Ln(4)
	drawTableHeader(pdf, []string{"Student ID", "Sessions", "Present", "Attendance %"})

	studentIDs := make([]int, 0, len(perStudent))
	for id := range perStudent {
		studentIDs = append(studentIDs, id)
	}
	sort.Ints(studentIDs)

	for _, id := range studentIDs {
		entry := perStudent[id]
		percentage := 0.0
		if entry.total > 0 {
			percentage = float64(entry.present) / float64(entry.total) * 100
		}
		row := []string{
			fmt.Sprintf("%d", id),
			fmt.Sprintf("%d", entry.total),
			fmt.Sprintf("%d", entry.present),
			fmt.Sprintf("%.2f%%", percentage),
		}
		drawTableRow(pdf, row)
	}

	return outputPDF(pdf)
}

func (s *reportService) GenerateCourseReport(ctx context.Context, collegeID, courseID int) ([]byte, error) {
	grades, err := s.gradeRepo.GetGradesByCourse(ctx, collegeID, courseID)
	if err != nil {
		return nil, fmt.Errorf("GenerateCourseReport: failed to fetch grades: %w", err)
	}
	attendance, err := s.attendanceRepo.GetAttendanceByCourse(ctx, collegeID, courseID, 10000, 0)
	if err != nil {
		return nil, fmt.Errorf("GenerateCourseReport: failed to fetch attendance: %w", err)
	}
	enrollments, err := s.enrollmentRepo.FindEnrollmentsByCourse(ctx, collegeID, courseID, 10000, 0)
	if err != nil {
		return nil, fmt.Errorf("GenerateCourseReport: failed to fetch enrollments: %w", err)
	}

	avgGrade := 0.0
	gradeDistribution := make(map[string]int)
	if len(grades) > 0 {
		total := 0.0
		for _, grade := range grades {
			total += grade.Percentage
			bucket := gradeBucket(grade.Percentage)
			gradeDistribution[bucket]++
		}
		avgGrade = total / float64(len(grades))
	}

	attSummary := summarizeAttendance(attendance)

	pdf := newPDF("Course Report")
	addHeader(pdf, fmt.Sprintf("Comprehensive Course Report - Course #%d", courseID))
	pdf.Ln(6)
	pdf.SetFont("Helvetica", "", 10)
	pdf.CellFormat(0, 5, fmt.Sprintf("Total Enrollments: %d", len(enrollments)), gofpdf.BorderNone, 1, gofpdf.AlignLeft, false, 0, "")
	pdf.CellFormat(0, 5, fmt.Sprintf("Average Grade Percentage: %.2f%%", avgGrade), gofpdf.BorderNone, 1, gofpdf.AlignLeft, false, 0, "")
	pdf.CellFormat(0, 5, fmt.Sprintf("Average Attendance: %.2f%%", attSummary.overall), gofpdf.BorderNone, 1, gofpdf.AlignLeft, false, 0, "")

	pdf.Ln(4)
	pdf.SetFont("Helvetica", "B", 11)
	pdf.CellFormat(0, 6, "Grade Distribution", gofpdf.BorderNone, 1, gofpdf.AlignLeft, false, 0, "")
	drawTableHeader(pdf, []string{"Grade", "Count"})
	for _, bucket := range []string{"A", "B", "C", "D", "F"} {
		row := []string{bucket, fmt.Sprintf("%d", gradeDistribution[bucket])}
		drawTableRow(pdf, row)
	}

	pdf.Ln(6)
	pdf.SetFont("Helvetica", "B", 11)
	pdf.CellFormat(0, 6, "Attendance Summary", gofpdf.BorderNone, 1, gofpdf.AlignLeft, false, 0, "")
	drawTableHeader(pdf, []string{"Student ID", "Attendance %"})
	for _, entry := range attSummary.byStudent {
		row := []string{
			fmt.Sprintf("%d", entry.studentID),
			fmt.Sprintf("%.2f%%", entry.percentage),
		}
		drawTableRow(pdf, row)
	}

	return outputPDF(pdf)
}

func newPDF(title string) *gofpdf.Fpdf {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetTitle(title, false)
	pdf.AddPage()
	return pdf
}

func addHeader(pdf *gofpdf.Fpdf, title string) {
	pdf.SetFont("Helvetica", "B", 14)
	pdf.CellFormat(0, 8, title, gofpdf.BorderNone, 1, gofpdf.AlignCenter, false, 0, "")
}

func drawTableHeader(pdf *gofpdf.Fpdf, headers []string) {
	pdf.SetFont("Helvetica", "B", 9)
	for _, header := range headers {
		pdf.CellFormat(0, 6, header, gofpdf.BorderBottom, 0, gofpdf.AlignLeft, false, 0, "")
	}
	pdf.Ln(-1)
}

func drawTableRow(pdf *gofpdf.Fpdf, columns []string) {
	pdf.SetFont("Helvetica", "", 9)
	for _, col := range columns {
		pdf.CellFormat(0, 6, col, gofpdf.BorderNone, 0, gofpdf.AlignLeft, false, 0, "")
	}
	pdf.Ln(-1)
}

func outputPDF(pdf *gofpdf.Fpdf) ([]byte, error) {
	buf := &bytes.Buffer{}
	if err := pdf.Output(buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func generateSimplePDF(title string, lines []string) ([]byte, error) {
	pdf := newPDF(title)
	addHeader(pdf, title)
	pdf.Ln(6)
	for _, line := range lines {
		pdf.CellFormat(0, 6, line, gofpdf.BorderNone, 1, gofpdf.AlignLeft, false, 0, "")
	}
	return outputPDF(pdf)
}

func gradeBucket(percentage float64) string {
	switch {
	case percentage >= 85:
		return "A"
	case percentage >= 70:
		return "B"
	case percentage >= 55:
		return "C"
	case percentage >= 40:
		return "D"
	default:
		return "F"
	}
}

type attendanceSummary struct {
	byStudent []struct {
		studentID  int
		percentage float64
	}
	overall float64
}

func summarizeAttendance(records []*models.Attendance) attendanceSummary {
	if len(records) == 0 {
		return attendanceSummary{}
	}

	perStudent := make(map[int]struct {
		present int
		total   int
	})
	var totalPresent, totalSessions int

	for _, record := range records {
		entry := perStudent[record.StudentID]
		if strings.EqualFold(record.Status, "present") {
			entry.present++
			totalPresent++
		}
		entry.total++
		totalSessions++
		perStudent[record.StudentID] = entry
	}

	students := make([]struct {
		studentID  int
		percentage float64
	}, 0, len(perStudent))
	for id, entry := range perStudent {
		percentage := 0.0
		if entry.total > 0 {
			percentage = float64(entry.present) / float64(entry.total) * 100
		}
		students = append(students, struct {
			studentID  int
			percentage float64
		}{studentID: id, percentage: percentage})
	}
	sort.Slice(students, func(i, j int) bool { return students[i].studentID < students[j].studentID })

	overall := 0.0
	if totalSessions > 0 {
		overall = float64(totalPresent) / float64(totalSessions) * 100
	}

	return attendanceSummary{
		byStudent: students,
		overall:   overall,
	}
}
