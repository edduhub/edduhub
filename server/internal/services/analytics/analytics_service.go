package analytics

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"time"

	"eduhub/server/internal/repository"
)

type StudentPerformanceMetrics struct {
	StudentID            int            `json:"student_id"`
	OverallGPA           float64        `json:"overall_gpa"`
	AttendanceRate       float64        `json:"attendance_rate"`
	AssignmentsSubmitted int            `json:"assignments_submitted"`
	AssignmentsTotal     int            `json:"assignments_total"`
	QuizzesCompleted     int            `json:"quizzes_completed"`
	AverageQuizScore     float64        `json:"average_quiz_score"`
	CourseMetrics        []CourseMetric `json:"course_metrics,omitempty"`
}

type CourseMetric struct {
	CourseID       int     `json:"course_id"`
	CourseName     string  `json:"course_name"`
	GPA            float64 `json:"gpa"`
	AttendanceRate float64 `json:"attendance_rate"`
}

type CourseAnalytics struct {
	CourseID             int     `json:"course_id"`
	TotalStudents        int     `json:"total_students"`
	AverageAttendance    float64 `json:"average_attendance"`
	AverageGrade         float64 `json:"average_grade"`
	AssignmentSubmission float64 `json:"assignment_submission_rate"`
	QuizParticipation    float64 `json:"quiz_participation_rate"`
	TopPerformers        []int   `json:"top_performers"`
	StudentsAtRisk       []int   `json:"students_at_risk"`
}

type CollegeDashboard struct {
	TotalStudents       int     `json:"total_students"`
	TotalCourses        int     `json:"total_courses"`
	TotalFaculty        int     `json:"total_faculty"`
	AverageAttendance   float64 `json:"average_attendance"`
	OverallGPA          float64 `json:"overall_gpa"`
	ActiveAnnouncements int     `json:"active_announcements"`
	UpcomingEvents      int     `json:"upcoming_events"`
}

type AttendanceTrend struct {
	Date           time.Time `json:"date"`
	AttendanceRate float64   `json:"attendance_rate"`
	TotalPresent   int       `json:"total_present"`
	TotalExpected  int       `json:"total_expected"`
}

type GradeDistribution struct {
	Grade string `json:"grade"`
	Count int    `json:"count"`
}

type AnalyticsService interface {
	GetStudentPerformance(ctx context.Context, collegeID, studentID int, courseID *int) (*StudentPerformanceMetrics, error)
	GetCourseAnalytics(ctx context.Context, collegeID, courseID int) (*CourseAnalytics, error)
	GetCollegeDashboard(ctx context.Context, collegeID int) (*CollegeDashboard, error)
	GetAttendanceTrends(ctx context.Context, collegeID int, courseID *int) ([]AttendanceTrend, error)
	GetGradeDistribution(ctx context.Context, collegeID, courseID int) ([]GradeDistribution, error)
}

type analyticsService struct {
	studentRepo    repository.StudentRepository
	attendanceRepo repository.AttendanceRepository
	gradeRepo      repository.GradeRepository
	courseRepo     repository.CourseRepository
	assignmentRepo repository.AssignmentRepository
	db             *repository.DB
}

func NewAnalyticsService(
	studentRepo repository.StudentRepository,
	attendanceRepo repository.AttendanceRepository,
	gradeRepo repository.GradeRepository,
	courseRepo repository.CourseRepository,
	assignmentRepo repository.AssignmentRepository,
	db *repository.DB,
) AnalyticsService {
	return &analyticsService{
		studentRepo:    studentRepo,
		attendanceRepo: attendanceRepo,
		gradeRepo:      gradeRepo,
		courseRepo:     courseRepo,
		assignmentRepo: assignmentRepo,
		db:             db,
	}
}

func (s *analyticsService) GetStudentPerformance(ctx context.Context, collegeID, studentID int, courseID *int) (*StudentPerformanceMetrics, error) {
	metrics := &StudentPerformanceMetrics{StudentID: studentID}

	avgPercentage, err := s.averageGradePercentage(ctx, collegeID, studentID, courseID)
	if err != nil {
		return nil, err
	}
	metrics.OverallGPA = percentageToGPA(avgPercentage)

	attendanceRate, err := s.attendanceRate(ctx, collegeID, studentID, courseID)
	if err != nil {
		return nil, err
	}
	metrics.AttendanceRate = attendanceRate

	submitted, totalAssignments, err := s.assignmentStats(ctx, collegeID, studentID, courseID)
	if err != nil {
		return nil, err
	}
	metrics.AssignmentsSubmitted = submitted
	metrics.AssignmentsTotal = totalAssignments

	quizzesCompleted, averageQuizScore, err := s.quizStats(ctx, collegeID, studentID, courseID)
	if err != nil {
		return nil, err
	}
	metrics.QuizzesCompleted = quizzesCompleted
	metrics.AverageQuizScore = averageQuizScore

	// Populate course metrics only when courseID filter is not provided
	if courseID == nil {
		courseMetrics, err := s.studentCourseMetrics(ctx, collegeID, studentID)
		if err != nil {
			return nil, err
		}
		metrics.CourseMetrics = courseMetrics
	}

	return metrics, nil
}

func (s *analyticsService) GetCourseAnalytics(ctx context.Context, collegeID, courseID int) (*CourseAnalytics, error) {
	analytics := &CourseAnalytics{CourseID: courseID}

	totalStudents, err := s.countEnrollments(ctx, collegeID, courseID)
	if err != nil {
		return nil, err
	}
	analytics.TotalStudents = totalStudents

	avgAttendance, err := s.courseAttendanceRate(ctx, collegeID, courseID)
	if err != nil {
		return nil, err
	}
	analytics.AverageAttendance = avgAttendance

	avgGrade, err := s.courseAverageGrade(ctx, collegeID, courseID)
	if err != nil {
		return nil, err
	}
	analytics.AverageGrade = percentageToGPA(avgGrade)

	assignmentSubmissionRate, err := s.courseAssignmentSubmissionRate(ctx, collegeID, courseID, totalStudents)
	if err != nil {
		return nil, err
	}
	analytics.AssignmentSubmission = assignmentSubmissionRate

	quizParticipation, err := s.courseQuizParticipation(ctx, collegeID, courseID, totalStudents)
	if err != nil {
		return nil, err
	}
	analytics.QuizParticipation = quizParticipation

	topPerformers, err := s.topPerformers(ctx, collegeID, courseID, 5)
	if err != nil {
		return nil, err
	}
	analytics.TopPerformers = topPerformers

	studentsAtRisk, err := s.studentsAtRisk(ctx, collegeID, courseID)
	if err != nil {
		return nil, err
	}
	analytics.StudentsAtRisk = studentsAtRisk

	return analytics, nil
}

func (s *analyticsService) GetCollegeDashboard(ctx context.Context, collegeID int) (*CollegeDashboard, error) {
	dashboard := &CollegeDashboard{}

	if err := s.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM students WHERE college_id = $1`, collegeID).Scan(&dashboard.TotalStudents); err != nil {
		return nil, fmt.Errorf("GetCollegeDashboard: failed to compute total students: %w", err)
	}

	if err := s.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM courses WHERE college_id = $1`, collegeID).Scan(&dashboard.TotalCourses); err != nil {
		return nil, fmt.Errorf("GetCollegeDashboard: failed to compute total courses: %w", err)
	}

	if err := s.db.Pool.QueryRow(ctx, `SELECT COUNT(DISTINCT instructor_id) FROM courses WHERE college_id = $1`, collegeID).Scan(&dashboard.TotalFaculty); err != nil {
		return nil, fmt.Errorf("GetCollegeDashboard: failed to compute faculty count: %w", err)
	}

	avgAttendance, err := s.overallAttendanceRate(ctx, collegeID)
	if err != nil {
		return nil, err
	}
	dashboard.AverageAttendance = avgAttendance

	avgPercentage, err := s.overallAveragePercentage(ctx, collegeID)
	if err != nil {
		return nil, err
	}
	dashboard.OverallGPA = percentageToGPA(avgPercentage)

	if err := s.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM announcements WHERE college_id = $1 AND is_published = TRUE AND (expires_at IS NULL OR expires_at > NOW())`, collegeID).Scan(&dashboard.ActiveAnnouncements); err != nil {
		return nil, fmt.Errorf("GetCollegeDashboard: failed to count announcements: %w", err)
	}

	if err := s.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM calendar_events WHERE college_id = $1 AND start_time >= NOW()`, collegeID).Scan(&dashboard.UpcomingEvents); err != nil {
		return nil, fmt.Errorf("GetCollegeDashboard: failed to count events: %w", err)
	}

	return dashboard, nil
}

func (s *analyticsService) GetAttendanceTrends(ctx context.Context, collegeID int, courseID *int) ([]AttendanceTrend, error) {
	query := `SELECT date, 
        COALESCE(SUM(CASE WHEN status = 'Present' THEN 1 ELSE 0 END),0) AS present,
        COUNT(*) AS expected
        FROM attendance
        WHERE college_id = $1 AND date >= (CURRENT_DATE - INTERVAL '14 day')`
	args := []interface{}{collegeID}

	if courseID != nil {
		query += " AND course_id = $2"
		args = append(args, *courseID)
	}

	query += " GROUP BY date ORDER BY date"

	rows, err := s.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("GetAttendanceTrends: failed to query attendance: %w", err)
	}
	defer rows.Close()

	trends := make([]AttendanceTrend, 0)
	for rows.Next() {
		var trend AttendanceTrend
		var present, expected int
		if err := rows.Scan(&trend.Date, &present, &expected); err != nil {
			return nil, fmt.Errorf("GetAttendanceTrends: failed to scan row: %w", err)
		}

		trend.TotalPresent = present
		trend.TotalExpected = expected
		if expected > 0 {
			trend.AttendanceRate = roundFloat(float64(present) / float64(expected) * 100)
		}

		trends = append(trends, trend)
	}

	return trends, nil
}

func (s *analyticsService) GetGradeDistribution(ctx context.Context, collegeID, courseID int) ([]GradeDistribution, error) {
	query := `SELECT bucket, COUNT(*) FROM (
        SELECT CASE
            WHEN percentage >= 85 THEN 'A'
            WHEN percentage >= 70 THEN 'B'
            WHEN percentage >= 55 THEN 'C'
            WHEN percentage >= 40 THEN 'D'
            ELSE 'F'
        END AS bucket
        FROM grades
        WHERE college_id = $1 AND course_id = $2
    ) AS buckets
    GROUP BY bucket
    ORDER BY bucket`

	rows, err := s.db.Pool.Query(ctx, query, collegeID, courseID)
	if err != nil {
		return nil, fmt.Errorf("GetGradeDistribution: failed to query distribution: %w", err)
	}
	defer rows.Close()

	distribution := make([]GradeDistribution, 0)
	for rows.Next() {
		var gd GradeDistribution
		if err := rows.Scan(&gd.Grade, &gd.Count); err != nil {
			return nil, fmt.Errorf("GetGradeDistribution: failed to scan row: %w", err)
		}
		distribution = append(distribution, gd)
	}

	return distribution, nil
}

func (s *analyticsService) averageGradePercentage(ctx context.Context, collegeID, studentID int, courseID *int) (float64, error) {
	query := `SELECT COALESCE(AVG(percentage),0) FROM grades WHERE college_id = $1 AND student_id = $2`
	args := []interface{}{collegeID, studentID}

	if courseID != nil {
		query += " AND course_id = $3"
		args = append(args, *courseID)
	}

	var avg sql.NullFloat64
	if err := s.db.Pool.QueryRow(ctx, query, args...).Scan(&avg); err != nil {
		return 0, fmt.Errorf("averageGradePercentage: query failed: %w", err)
	}

	if avg.Valid {
		return roundFloat(avg.Float64), nil
	}

	return 0, nil
}

func (s *analyticsService) attendanceRate(ctx context.Context, collegeID, studentID int, courseID *int) (float64, error) {
	query := `SELECT 
        COALESCE(SUM(CASE WHEN status = 'Present' THEN 1 ELSE 0 END),0) AS present,
        COUNT(*) AS total
        FROM attendance WHERE college_id = $1 AND student_id = $2`
	args := []interface{}{collegeID, studentID}

	if courseID != nil {
		query += " AND course_id = $3"
		args = append(args, *courseID)
	}

	var present, total int
	if err := s.db.Pool.QueryRow(ctx, query, args...).Scan(&present, &total); err != nil {
		return 0, fmt.Errorf("attendanceRate: query failed: %w", err)
	}

	if total == 0 {
		return 0, nil
	}

	return roundFloat(float64(present) / float64(total) * 100), nil
}

func (s *analyticsService) assignmentStats(ctx context.Context, collegeID, studentID int, courseID *int) (int, int, error) {
	submissionQuery := `SELECT COUNT(*) FROM assignment_submissions s
        JOIN assignments a ON a.id = s.assignment_id
        WHERE s.student_id = $1 AND a.college_id = $2`
	submissionArgs := []interface{}{studentID, collegeID}

	totalQuery := `SELECT COUNT(DISTINCT a.id) FROM assignments a
        JOIN enrollments e ON e.course_id = a.course_id AND e.college_id = a.college_id
        WHERE e.student_id = $1 AND a.college_id = $2`
	totalArgs := []interface{}{studentID, collegeID}

	if courseID != nil {
		submissionQuery += " AND a.course_id = $3"
		submissionArgs = append(submissionArgs, *courseID)

		totalQuery += " AND a.course_id = $3"
		totalArgs = append(totalArgs, *courseID)
	}

	var submitted int
	if err := s.db.Pool.QueryRow(ctx, submissionQuery, submissionArgs...).Scan(&submitted); err != nil {
		return 0, 0, fmt.Errorf("assignmentStats: failed to count submissions: %w", err)
	}

	var total int
	if err := s.db.Pool.QueryRow(ctx, totalQuery, totalArgs...).Scan(&total); err != nil {
		return 0, 0, fmt.Errorf("assignmentStats: failed to count assignments: %w", err)
	}

	return submitted, total, nil
}

func (s *analyticsService) quizStats(ctx context.Context, collegeID, studentID int, courseID *int) (int, float64, error) {
	query := `SELECT COUNT(*), COALESCE(AVG(score),0) FROM quiz_attempts qa
        JOIN quizzes q ON q.id = qa.quiz_id
        WHERE qa.college_id = $1 AND qa.student_id = $2 AND qa.status IN ('submitted','graded')`
	args := []interface{}{collegeID, studentID}

	if courseID != nil {
		query += " AND q.course_id = $3"
		args = append(args, *courseID)
	}

	var count int
	var avg sql.NullFloat64
	if err := s.db.Pool.QueryRow(ctx, query, args...).Scan(&count, &avg); err != nil {
		return 0, 0, fmt.Errorf("quizStats: query failed: %w", err)
	}

	if avg.Valid {
		return count, roundFloat(avg.Float64), nil
	}

	return count, 0, nil
}

func (s *analyticsService) studentCourseMetrics(ctx context.Context, collegeID, studentID int) ([]CourseMetric, error) {
	query := `SELECT c.id, c.name,
		COALESCE(AVG(g.percentage),0) AS avg_percentage,
		COALESCE(SUM(CASE WHEN a.status = 'Present' THEN 1 ELSE 0 END)::float / NULLIF(COUNT(a.id),0) * 100, 0) AS attendance_rate
        FROM courses c
        JOIN enrollments e ON e.course_id = c.id AND e.student_id = $2
        LEFT JOIN grades g ON g.course_id = c.id AND g.student_id = e.student_id AND g.college_id = $1
        LEFT JOIN attendance a ON a.course_id = c.id AND a.student_id = e.student_id AND a.college_id = $1
        WHERE c.college_id = $1 AND e.college_id = $1
        GROUP BY c.id, c.name`

	rows, err := s.db.Pool.Query(ctx, query, collegeID, studentID)
	if err != nil {
		return nil, fmt.Errorf("studentCourseMetrics: query failed: %w", err)
	}
	defer rows.Close()

	metrics := make([]CourseMetric, 0)
	for rows.Next() {
		var cm CourseMetric
		var avgPercentage sql.NullFloat64
		var attendance sql.NullFloat64

		if err := rows.Scan(&cm.CourseID, &cm.CourseName, &avgPercentage, &attendance); err != nil {
			return nil, fmt.Errorf("studentCourseMetrics: scan failed: %w", err)
		}

		if avgPercentage.Valid {
			cm.GPA = percentageToGPA(avgPercentage.Float64)
		}
		if attendance.Valid {
			cm.AttendanceRate = roundFloat(attendance.Float64)
		}

		metrics = append(metrics, cm)
	}

	return metrics, nil
}

func (s *analyticsService) countEnrollments(ctx context.Context, collegeID, courseID int) (int, error) {
	var total int
	if err := s.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM enrollments WHERE college_id = $1 AND course_id = $2`, collegeID, courseID).Scan(&total); err != nil {
		return 0, fmt.Errorf("countEnrollments: query failed: %w", err)
	}
	return total, nil
}

func (s *analyticsService) courseAttendanceRate(ctx context.Context, collegeID, courseID int) (float64, error) {
	var present, total int
	if err := s.db.Pool.QueryRow(ctx, `SELECT COALESCE(SUM(CASE WHEN status = 'Present' THEN 1 ELSE 0 END),0) AS present,
        COUNT(*) AS total FROM attendance WHERE college_id = $1 AND course_id = $2`, collegeID, courseID).Scan(&present, &total); err != nil {
		return 0, fmt.Errorf("courseAttendanceRate: query failed: %w", err)
	}
	if total == 0 {
		return 0, nil
	}
	return roundFloat(float64(present) / float64(total) * 100), nil
}

func (s *analyticsService) courseAverageGrade(ctx context.Context, collegeID, courseID int) (float64, error) {
	var avg sql.NullFloat64
	if err := s.db.Pool.QueryRow(ctx, `SELECT COALESCE(AVG(percentage),0) FROM grades WHERE college_id = $1 AND course_id = $2`, collegeID, courseID).Scan(&avg); err != nil {
		return 0, fmt.Errorf("courseAverageGrade: query failed: %w", err)
	}
	if avg.Valid {
		return roundFloat(avg.Float64), nil
	}
	return 0, nil
}

func (s *analyticsService) courseAssignmentSubmissionRate(ctx context.Context, collegeID, courseID, totalStudents int) (float64, error) {
	if totalStudents == 0 {
		return 0, nil
	}

	var totalAssignments int
	if err := s.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM assignments WHERE college_id = $1 AND course_id = $2`, collegeID, courseID).Scan(&totalAssignments); err != nil {
		return 0, fmt.Errorf("courseAssignmentSubmissionRate: failed to count assignments: %w", err)
	}

	if totalAssignments == 0 {
		return 0, nil
	}

	var submissions int
	if err := s.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM assignment_submissions s
        JOIN assignments a ON a.id = s.assignment_id
        WHERE a.college_id = $1 AND a.course_id = $2`, collegeID, courseID).Scan(&submissions); err != nil {
		return 0, fmt.Errorf("courseAssignmentSubmissionRate: failed to count submissions: %w", err)
	}

	denominator := totalAssignments * totalStudents
	if denominator == 0 {
		return 0, nil
	}

	return roundFloat(float64(submissions) / float64(denominator) * 100), nil
}

func (s *analyticsService) courseQuizParticipation(ctx context.Context, collegeID, courseID, totalStudents int) (float64, error) {
	if totalStudents == 0 {
		return 0, nil
	}

	var totalQuizzes int
	if err := s.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM quizzes WHERE college_id = $1 AND course_id = $2`, collegeID, courseID).Scan(&totalQuizzes); err != nil {
		return 0, fmt.Errorf("courseQuizParticipation: failed to count quizzes: %w", err)
	}

	if totalQuizzes == 0 {
		return 0, nil
	}

	var attempts int
	if err := s.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM quiz_attempts qa
        WHERE qa.college_id = $1 AND qa.quiz_id IN (SELECT id FROM quizzes WHERE college_id = $1 AND course_id = $2)
        AND qa.status IN ('submitted','graded')`, collegeID, courseID).Scan(&attempts); err != nil {
		return 0, fmt.Errorf("courseQuizParticipation: failed to count attempts: %w", err)
	}

	denominator := totalQuizzes * totalStudents
	if denominator == 0 {
		return 0, nil
	}

	return roundFloat(float64(attempts) / float64(denominator) * 100), nil
}

func (s *analyticsService) topPerformers(ctx context.Context, collegeID, courseID, limit int) ([]int, error) {
	query := `SELECT student_id FROM grades WHERE college_id = $1 AND course_id = $2
        ORDER BY percentage DESC LIMIT $3`
	rows, err := s.db.Pool.Query(ctx, query, collegeID, courseID, limit)
	if err != nil {
		return nil, fmt.Errorf("topPerformers: query failed: %w", err)
	}
	defer rows.Close()

	performers := make([]int, 0)
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("topPerformers: scan failed: %w", err)
		}
		performers = append(performers, id)
	}

	return performers, nil
}

func (s *analyticsService) studentsAtRisk(ctx context.Context, collegeID, courseID int) ([]int, error) {
	query := `SELECT student_id FROM enrollments e
        WHERE e.college_id = $1 AND e.course_id = $2
        AND (
            EXISTS (
                SELECT 1 FROM attendance a
                WHERE a.college_id = e.college_id AND a.course_id = e.course_id AND a.student_id = e.student_id
                GROUP BY a.student_id
                HAVING COALESCE(SUM(CASE WHEN a.status = 'Present' THEN 1 ELSE 0 END)::float / NULLIF(COUNT(*),0),0) < 0.6
            )
            OR EXISTS (
                SELECT 1 FROM grades g
                WHERE g.college_id = e.college_id AND g.course_id = e.course_id AND g.student_id = e.student_id
                GROUP BY g.student_id
                HAVING COALESCE(AVG(g.percentage),0) < 60
            )
        )`

	rows, err := s.db.Pool.Query(ctx, query, collegeID, courseID)
	if err != nil {
		return nil, fmt.Errorf("studentsAtRisk: query failed: %w", err)
	}
	defer rows.Close()

	ids := make([]int, 0)
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("studentsAtRisk: scan failed: %w", err)
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func (s *analyticsService) overallAttendanceRate(ctx context.Context, collegeID int) (float64, error) {
	var present, total int
	if err := s.db.Pool.QueryRow(ctx, `SELECT COALESCE(SUM(CASE WHEN status = 'Present' THEN 1 ELSE 0 END),0), COUNT(*)
        FROM attendance WHERE college_id = $1`, collegeID).Scan(&present, &total); err != nil {
		return 0, fmt.Errorf("overallAttendanceRate: query failed: %w", err)
	}

	if total == 0 {
		return 0, nil
	}

	return roundFloat(float64(present) / float64(total) * 100), nil
}

func (s *analyticsService) overallAveragePercentage(ctx context.Context, collegeID int) (float64, error) {
	var avg sql.NullFloat64
	if err := s.db.Pool.QueryRow(ctx, `SELECT COALESCE(AVG(percentage),0) FROM grades WHERE college_id = $1`, collegeID).Scan(&avg); err != nil {
		return 0, fmt.Errorf("overallAveragePercentage: query failed: %w", err)
	}

	if avg.Valid {
		return roundFloat(avg.Float64), nil
	}

	return 0, nil
}

// roundFloat rounds a float64 to 2 decimal places
func roundFloat(val float64) float64 {
	return math.Round(val*100) / 100
}

