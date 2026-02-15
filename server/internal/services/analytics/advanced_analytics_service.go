package analytics

import (
	"context"
	"fmt"
	"math"
	"time"

	"eduhub/server/internal/config"
	"eduhub/server/internal/repository"
)

type AdvancedAnalyticsService interface {
	GetStudentProgression(ctx context.Context, collegeID, studentID int) (*StudentProgression, error)
	GetCourseEngagement(ctx context.Context, collegeID, courseID int) (*CourseEngagement, error)
	GetPredictiveInsights(ctx context.Context, collegeID int) (*PredictiveInsights, error)
	GetLearningAnalytics(ctx context.Context, collegeID int, startDate, endDate *time.Time) (*LearningAnalytics, error)
	GetPerformanceTrends(ctx context.Context, collegeID int, entityType string, entityID int) ([]PerformanceTrend, error)
	GetComparativeAnalysis(ctx context.Context, collegeID int, courseIDs []int) (*ComparativeAnalysis, error)
}

type StudentProgression struct {
	StudentID        int                    `json:"student_id"`
	OverallTrend     string                 `json:"overall_trend"` // improving, declining, stable
	GradeProgression []GradeProgressPoint   `json:"grade_progression"`
	AttendanceTrend  []AttendanceTrendPoint `json:"attendance_trend"`
	SkillDevelopment []SkillPoint           `json:"skill_development"`
	Recommendations  []string               `json:"recommendations"`
}

type GradeProgressPoint struct {
	Date       time.Time `json:"date"`
	AverageGPA float64   `json:"average_gpa"`
	CourseID   int       `json:"course_id,omitempty"`
}

type AttendanceTrendPoint struct {
	Date           time.Time `json:"date"`
	AttendanceRate float64   `json:"attendance_rate"`
	CourseID       int       `json:"course_id,omitempty"`
}

type SkillPoint struct {
	Skill    string    `json:"skill"`
	Level    float64   `json:"level"`
	Date     time.Time `json:"date"`
	CourseID int       `json:"course_id,omitempty"`
}

type CourseEngagement struct {
	CourseID            int               `json:"course_id"`
	TotalStudents       int               `json:"total_students"`
	ActiveStudents      int               `json:"active_students"`
	EngagementRate      float64           `json:"engagement_rate"`
	ActivityBreakdown   map[string]int    `json:"activity_breakdown"`
	PeakActivityHours   []int             `json:"peak_activity_hours"`
	DropoutRiskStudents []int             `json:"dropout_risk_students"`
	EngagementTimeline  []EngagementPoint `json:"engagement_timeline"`
}

type EngagementPoint struct {
	Date            time.Time `json:"date"`
	ActiveUsers     int       `json:"active_users"`
	AssignmentsDone int       `json:"assignments_done"`
	QuizzesTaken    int       `json:"quizzes_taken"`
	ForumPosts      int       `json:"forum_posts"`
}

type PredictiveInsights struct {
	AtRiskStudents        []RiskStudent          `json:"at_risk_students"`
	CourseCompletionRates []CompletionRate       `json:"course_completion_rates"`
	GradePredictions      []GradePrediction      `json:"grade_predictions"`
	AttendancePredictions []AttendancePrediction `json:"attendance_predictions"`
	Recommendations       []string               `json:"recommendations"`
}

type RiskStudent struct {
	StudentID     int      `json:"student_id"`
	RiskLevel     string   `json:"risk_level"` // high, medium, low
	RiskFactors   []string `json:"risk_factors"`
	Probability   float64  `json:"probability"`
	Interventions []string `json:"interventions"`
}

type CompletionRate struct {
	CourseID       int     `json:"course_id"`
	CompletionRate float64 `json:"completion_rate"`
	PredictedRate  float64 `json:"predicted_rate"`
	TimeToComplete int     `json:"time_to_complete_days"`
}

type GradePrediction struct {
	StudentID    int     `json:"student_id"`
	CourseID     int     `json:"course_id"`
	PredictedGPA float64 `json:"predicted_gpa"`
	Confidence   float64 `json:"confidence"`
}

type AttendancePrediction struct {
	StudentID           int     `json:"student_id"`
	CourseID            int     `json:"course_id"`
	PredictedAttendance float64 `json:"predicted_attendance"`
	Confidence          float64 `json:"confidence"`
}

type LearningAnalytics struct {
	Period               string              `json:"period"`
	TotalStudents        int                 `json:"total_students"`
	TotalCourses         int                 `json:"total_courses"`
	AverageEngagement    float64             `json:"average_engagement"`
	TopPerformingCourses []CoursePerformance `json:"top_performing_courses"`
	LearningPatterns     []LearningPattern   `json:"learning_patterns"`
	ResourceUtilization  map[string]int      `json:"resource_utilization"`
	TimeSpentAnalytics   []TimeSpentPoint    `json:"time_spent_analytics"`
}

type CoursePerformance struct {
	CourseID       int     `json:"course_id"`
	CourseName     string  `json:"course_name"`
	AverageGrade   float64 `json:"average_grade"`
	CompletionRate float64 `json:"completion_rate"`
	EngagementRate float64 `json:"engagement_rate"`
}

type LearningPattern struct {
	Pattern     string  `json:"pattern"`
	Students    int     `json:"students"`
	SuccessRate float64 `json:"success_rate"`
}

type TimeSpentPoint struct {
	Date       time.Time `json:"date"`
	HoursSpent float64   `json:"hours_spent"`
	Activity   string    `json:"activity"`
}

type PerformanceTrend struct {
	Date       time.Time `json:"date"`
	Metric     string    `json:"metric"`
	Value      float64   `json:"value"`
	ChangeRate float64   `json:"change_rate"`
}

type ComparativeAnalysis struct {
	CourseComparisons []CourseComparison `json:"course_comparisons"`
	OverallInsights   []string           `json:"overall_insights"`
	Recommendations   []string           `json:"recommendations"`
}

type CourseComparison struct {
	CourseID1       int                `json:"course_id_1"`
	CourseID2       int                `json:"course_id_2"`
	CourseName1     string             `json:"course_name_1"`
	CourseName2     string             `json:"course_name_2"`
	Metrics         map[string]float64 `json:"metrics"`
	SignificantDiff []string           `json:"significant_differences"`
}

type advancedAnalyticsService struct {
	db              *repository.DB
	basicAnalytics  AnalyticsService
	analyticsConfig *config.AnalyticsConfig
}

func NewAdvancedAnalyticsService(db *repository.DB, basicAnalytics AnalyticsService) AdvancedAnalyticsService {
	return &advancedAnalyticsService{
		db:              db,
		basicAnalytics:  basicAnalytics,
		analyticsConfig: config.LoadAnalyticsConfig(),
	}
}

func (s *advancedAnalyticsService) GetStudentProgression(ctx context.Context, collegeID, studentID int) (*StudentProgression, error) {
	progression := &StudentProgression{StudentID: studentID}

	// Get grade progression over time
	gradeProgression, err := s.getGradeProgression(ctx, collegeID, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get grade progression: %w", err)
	}
	progression.GradeProgression = gradeProgression

	// Get attendance trends
	attendanceTrend, err := s.getAttendanceTrends(ctx, collegeID, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get attendance trends: %w", err)
	}
	progression.AttendanceTrend = attendanceTrend

	// Analyze overall trend
	progression.OverallTrend = s.analyzeOverallTrend(gradeProgression, attendanceTrend)

	// Get skill development data
	skillDevelopment, err := s.getSkillDevelopment(ctx, collegeID, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get skill development: %w", err)
	}
	progression.SkillDevelopment = skillDevelopment

	// Generate recommendations
	progression.Recommendations = s.generateStudentRecommendations(progression.OverallTrend, gradeProgression, attendanceTrend)

	return progression, nil
}

func (s *advancedAnalyticsService) GetCourseEngagement(ctx context.Context, collegeID, courseID int) (*CourseEngagement, error) {
	engagement := &CourseEngagement{CourseID: courseID}

	// Get basic enrollment data
	totalStudents, err := s.getTotalStudents(ctx, collegeID, courseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total students: %w", err)
	}
	engagement.TotalStudents = totalStudents

	// Calculate active students (students with recent activity)
	activeStudents, err := s.getActiveStudents(ctx, collegeID, courseID, 30) // last 30 days
	if err != nil {
		return nil, fmt.Errorf("failed to get active students: %w", err)
	}
	engagement.ActiveStudents = activeStudents

	if totalStudents > 0 {
		engagement.EngagementRate = float64(activeStudents) / float64(totalStudents) * 100
	}

	// Get activity breakdown
	activityBreakdown, err := s.getActivityBreakdown(ctx, collegeID, courseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity breakdown: %w", err)
	}
	engagement.ActivityBreakdown = activityBreakdown

	// Get peak activity hours
	peakHours, err := s.getPeakActivityHours(ctx, collegeID, courseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get peak activity hours: %w", err)
	}
	engagement.PeakActivityHours = peakHours

	// Identify students at risk of dropping out
	dropoutRisk, err := s.getDropoutRiskStudents(ctx, collegeID, courseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dropout risk students: %w", err)
	}
	engagement.DropoutRiskStudents = dropoutRisk

	// Get engagement timeline
	timeline, err := s.getEngagementTimeline(ctx, collegeID, courseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get engagement timeline: %w", err)
	}
	engagement.EngagementTimeline = timeline

	return engagement, nil
}

func (s *advancedAnalyticsService) GetPredictiveInsights(ctx context.Context, collegeID int) (*PredictiveInsights, error) {
	insights := &PredictiveInsights{}

	// Identify at-risk students
	atRiskStudents, err := s.identifyAtRiskStudents(ctx, collegeID)
	if err != nil {
		return nil, fmt.Errorf("failed to identify at-risk students: %w", err)
	}
	insights.AtRiskStudents = atRiskStudents

	// Predict course completion rates
	completionRates, err := s.predictCourseCompletionRates(ctx, collegeID)
	if err != nil {
		return nil, fmt.Errorf("failed to predict completion rates: %w", err)
	}
	insights.CourseCompletionRates = completionRates

	// Generate recommendations
	insights.Recommendations = s.generatePredictiveRecommendations(atRiskStudents, completionRates)

	return insights, nil
}

func (s *advancedAnalyticsService) GetLearningAnalytics(ctx context.Context, collegeID int, startDate, endDate *time.Time) (*LearningAnalytics, error) {
	analytics := &LearningAnalytics{}

	// Set period
	if startDate != nil && endDate != nil {
		analytics.Period = fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	} else {
		analytics.Period = "Last 30 days"
	}

	// Get basic metrics
	totalStudents, err := s.getTotalStudentsInCollege(ctx, collegeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total students: %w", err)
	}
	analytics.TotalStudents = totalStudents

	totalCourses, err := s.getTotalCoursesInCollege(ctx, collegeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total courses: %w", err)
	}
	analytics.TotalCourses = totalCourses

	// Get top performing courses
	topCourses, err := s.getTopPerformingCourses(ctx, collegeID, 5)
	if err != nil {
		return nil, fmt.Errorf("failed to get top performing courses: %w", err)
	}
	analytics.TopPerformingCourses = topCourses

	// Identify learning patterns
	patterns, err := s.identifyLearningPatterns(ctx, collegeID)
	if err != nil {
		return nil, fmt.Errorf("failed to identify learning patterns: %w", err)
	}
	analytics.LearningPatterns = patterns

	return analytics, nil
}

func (s *advancedAnalyticsService) GetPerformanceTrends(ctx context.Context, collegeID int, entityType string, entityID int) ([]PerformanceTrend, error) {
	trends := make([]PerformanceTrend, 0)

	switch entityType {
	case "student":
		return s.getStudentPerformanceTrends(ctx, collegeID, entityID)
	case "course":
		return s.getCoursePerformanceTrends(ctx, collegeID, entityID)
	default:
		return trends, fmt.Errorf("invalid entity type: %s", entityType)
	}
}

func (s *advancedAnalyticsService) GetComparativeAnalysis(ctx context.Context, collegeID int, courseIDs []int) (*ComparativeAnalysis, error) {
	analysis := &ComparativeAnalysis{}

	comparisons := make([]CourseComparison, 0)

	// Compare courses pairwise
	for i := 0; i < len(courseIDs); i++ {
		for j := i + 1; j < len(courseIDs); j++ {
			comparison, err := s.compareCourses(ctx, collegeID, courseIDs[i], courseIDs[j])
			if err != nil {
				continue // Skip failed comparisons
			}
			comparisons = append(comparisons, *comparison)
		}
	}

	analysis.CourseComparisons = comparisons
	analysis.OverallInsights = s.generateComparativeInsights(comparisons)
	analysis.Recommendations = s.generateComparativeRecommendations(comparisons)

	return analysis, nil
}

// Helper methods (simplified implementations)

func (s *advancedAnalyticsService) getGradeProgression(ctx context.Context, collegeID, studentID int) ([]GradeProgressPoint, error) {
	query := `
		SELECT DATE_TRUNC('month', created_at) as month, AVG(percentage) as avg_grade
		FROM grades
		WHERE college_id = $1 AND student_id = $2
		GROUP BY DATE_TRUNC('month', created_at)
		ORDER BY month DESC
		LIMIT 12`

	rows, err := s.db.Pool.Query(ctx, query, collegeID, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	points := make([]GradeProgressPoint, 0)
	for rows.Next() {
		var point GradeProgressPoint
		if err := rows.Scan(&point.Date, &point.AverageGPA); err != nil {
			continue
		}
		point.AverageGPA = PercentageToGPA(point.AverageGPA)
		points = append(points, point)
	}

	return points, nil
}

func (s *advancedAnalyticsService) getAttendanceTrends(ctx context.Context, collegeID, studentID int) ([]AttendanceTrendPoint, error) {
	query := `
		SELECT DATE_TRUNC('week', date) as week,
		       COALESCE(SUM(CASE WHEN status = 'Present' THEN 1 ELSE 0 END)::float / COUNT(*) * 100, 0) as attendance_rate
		FROM attendance
		WHERE college_id = $1 AND student_id = $2
		GROUP BY DATE_TRUNC('week', date)
		ORDER BY week DESC
		LIMIT 12`

	rows, err := s.db.Pool.Query(ctx, query, collegeID, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	points := make([]AttendanceTrendPoint, 0)
	for rows.Next() {
		var point AttendanceTrendPoint
		if err := rows.Scan(&point.Date, &point.AttendanceRate); err != nil {
			continue
		}
		points = append(points, point)
	}

	return points, nil
}

func (s *advancedAnalyticsService) analyzeOverallTrend(grades []GradeProgressPoint, attendance []AttendanceTrendPoint) string {
	if len(grades) < 2 {
		return "insufficient_data"
	}

	// Simple trend analysis - compare first half with second half
	mid := len(grades) / 2
	firstHalf := grades[mid:]
	secondHalf := grades[:mid]

	var firstAvg, secondAvg float64
	for _, g := range firstHalf {
		firstAvg += g.AverageGPA
	}
	firstAvg /= float64(len(firstHalf))

	for _, g := range secondHalf {
		secondAvg += g.AverageGPA
	}
	secondAvg /= float64(len(secondHalf))

	if secondAvg > firstAvg+0.2 {
		return "improving"
	} else if secondAvg < firstAvg-0.2 {
		return "declining"
	}
	return "stable"
}

func (s *advancedAnalyticsService) generateStudentRecommendations(trend string, grades []GradeProgressPoint, attendance []AttendanceTrendPoint) []string {
	recommendations := make([]string, 0)

	switch trend {
	case "declining":
		recommendations = append(recommendations, "Consider additional tutoring support")
		recommendations = append(recommendations, "Review study habits and time management")
	case "improving":
		recommendations = append(recommendations, "Continue current study strategies")
		recommendations = append(recommendations, "Consider advanced coursework")
	}

	// Check attendance
	if len(attendance) > 0 {
		recent := attendance[0]
		if recent.AttendanceRate < 75 {
			recommendations = append(recommendations, "Improve attendance - currently below 75%")
		}
	}

	return recommendations
}

func (s *advancedAnalyticsService) getTotalStudents(ctx context.Context, collegeID, courseID int) (int, error) {
	var count int
	err := s.db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM enrollments WHERE college_id = $1 AND course_id = $2", collegeID, courseID).Scan(&count)
	return count, err
}

func (s *advancedAnalyticsService) getActiveStudents(ctx context.Context, collegeID, courseID, days int) (int, error) {
	query := `
		SELECT COUNT(DISTINCT student_id) FROM (
			SELECT student_id FROM attendance WHERE college_id = $1 AND course_id = $2 AND date >= CURRENT_DATE - INTERVAL '%d days'
			UNION
			SELECT student_id FROM assignment_submissions s JOIN assignments a ON a.id = s.assignment_id
			WHERE a.college_id = $1 AND a.course_id = $2 AND s.created_at >= CURRENT_DATE - INTERVAL '%d days'
			UNION
			SELECT student_id FROM quiz_attempts qa JOIN quizzes q ON q.id = qa.quiz_id
			WHERE qa.college_id = $1 AND q.course_id = $2 AND qa.created_at >= CURRENT_DATE - INTERVAL '%d days'
		) active_students`

	query = fmt.Sprintf(query, days, days, days)

	var count int
	err := s.db.Pool.QueryRow(ctx, query, collegeID, courseID).Scan(&count)
	return count, err
}

func (s *advancedAnalyticsService) getActivityBreakdown(ctx context.Context, collegeID, courseID int) (map[string]int, error) {
	breakdown := make(map[string]int)

	// Count assignments
	var assignments int
	err := s.db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM assignments WHERE college_id = $1 AND course_id = $2", collegeID, courseID).Scan(&assignments)
	if err != nil {
		return nil, err
	}
	breakdown["assignments"] = assignments

	// Count quizzes
	var quizzes int
	err = s.db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM quizzes WHERE college_id = $1 AND course_id = $2", collegeID, courseID).Scan(&quizzes)
	if err != nil {
		return nil, err
	}
	breakdown["quizzes"] = quizzes

	// Count lectures
	var lectures int
	err = s.db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM lectures WHERE college_id = $1 AND course_id = $2", collegeID, courseID).Scan(&lectures)
	if err != nil {
		return nil, err
	}
	breakdown["lectures"] = lectures

	return breakdown, nil
}

func (s *advancedAnalyticsService) getPeakActivityHours(ctx context.Context, collegeID, courseID int) ([]int, error) {
	query := `
		SELECT EXTRACT(hour FROM created_at) as hour, COUNT(*) as activity_count
		FROM (
			SELECT created_at FROM assignment_submissions s JOIN assignments a ON a.id = s.assignment_id WHERE a.college_id = $1 AND a.course_id = $2
			UNION ALL
			SELECT created_at FROM quiz_attempts qa JOIN quizzes q ON q.id = qa.quiz_id WHERE qa.college_id = $1 AND q.course_id = $2
		) activities
		GROUP BY EXTRACT(hour FROM created_at)
		ORDER BY activity_count DESC
		LIMIT 3`

	rows, err := s.db.Pool.Query(ctx, query, collegeID, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hours := make([]int, 0)
	for rows.Next() {
		var hour int
		var count int
		if err := rows.Scan(&hour, &count); err != nil {
			continue
		}
		hours = append(hours, hour)
	}

	return hours, nil
}

func (s *advancedAnalyticsService) getDropoutRiskStudents(ctx context.Context, collegeID, courseID int) ([]int, error) {
	query := `
		SELECT DISTINCT e.student_id
		FROM enrollments e
		WHERE e.college_id = $1 AND e.course_id = $2
		AND (
			-- Low attendance (< 60%)
			EXISTS (
				SELECT 1 FROM attendance a
				WHERE a.college_id = e.college_id AND a.course_id = e.course_id AND a.student_id = e.student_id
				GROUP BY a.student_id
				HAVING COALESCE(SUM(CASE WHEN a.status = 'Present' THEN 1 ELSE 0 END)::float / NULLIF(COUNT(*),0),0) < 0.6
			)
			OR
			-- Low grades (< 50%)
			EXISTS (
				SELECT 1 FROM grades g
				WHERE g.college_id = e.college_id AND g.course_id = e.course_id AND g.student_id = e.student_id
				GROUP BY g.student_id
				HAVING COALESCE(AVG(g.percentage),0) < 50
			)
			OR
			-- No recent activity (last 14 days)
			NOT EXISTS (
				SELECT 1 FROM attendance a
				WHERE a.college_id = e.college_id AND a.course_id = e.course_id AND a.student_id = e.student_id
				AND a.date >= CURRENT_DATE - INTERVAL '14 days'
			)
		)`

	rows, err := s.db.Pool.Query(ctx, query, collegeID, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	students := make([]int, 0)
	for rows.Next() {
		var studentID int
		if err := rows.Scan(&studentID); err != nil {
			continue
		}
		students = append(students, studentID)
	}

	return students, nil
}

func (s *advancedAnalyticsService) identifyAtRiskStudents(ctx context.Context, collegeID int) ([]RiskStudent, error) {
	query := `
		SELECT
			s.id as student_id,
			COALESCE(AVG(g.percentage), 0) as avg_grade,
			COALESCE(AVG(CASE WHEN a.status = 'Present' THEN 100 ELSE 0 END), 0) as attendance_rate,
			COUNT(DISTINCT CASE WHEN g.created_at >= CURRENT_DATE - INTERVAL '30 days' THEN g.id END) as recent_grades,
			COUNT(DISTINCT CASE WHEN a.date >= CURRENT_DATE - INTERVAL '30 days' THEN a.id END) as recent_attendance
		FROM students s
		LEFT JOIN grades g ON g.student_id = s.id AND g.college_id = $1
		LEFT JOIN attendance a ON a.student_id = s.id AND a.college_id = $1
		WHERE s.college_id = $1
		GROUP BY s.id
		HAVING (
			COALESCE(AVG(g.percentage), 0) < 60 OR
			COALESCE(AVG(CASE WHEN a.status = 'Present' THEN 100 ELSE 0 END), 0) < 70 OR
			COUNT(DISTINCT CASE WHEN g.created_at >= CURRENT_DATE - INTERVAL '30 days' THEN g.id END) = 0
		)`

	rows, err := s.db.Pool.Query(ctx, query, collegeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	riskStudents := make([]RiskStudent, 0)
	for rows.Next() {
		var studentID int
		var avgGrade, attendanceRate float64
		var recentGrades, recentAttendance int

		if err := rows.Scan(&studentID, &avgGrade, &attendanceRate, &recentGrades, &recentAttendance); err != nil {
			continue
		}

		riskFactors := make([]string, 0)
		if avgGrade < 60 {
			riskFactors = append(riskFactors, "Low grades")
		}
		if attendanceRate < 70 {
			riskFactors = append(riskFactors, "Poor attendance")
		}
		if recentGrades == 0 {
			riskFactors = append(riskFactors, "No recent assessments")
		}

		riskLevel := "medium"
		probability := s.calculateRiskProbability(avgGrade, attendanceRate, recentGrades, recentAttendance)
		if probability >= s.analyticsConfig.RiskLevelHighThreshold {
			riskLevel = "high"
		} else if probability < s.analyticsConfig.RiskLevelLowThreshold {
			riskLevel = "low"
		}

		interventions := make([]string, 0)
		if riskLevel == "high" {
			interventions = append(interventions, "Schedule meeting with academic advisor")
			interventions = append(interventions, "Provide additional tutoring support")
		} else {
			interventions = append(interventions, "Monitor progress closely")
			interventions = append(interventions, "Send encouragement notifications")
		}

		riskStudents = append(riskStudents, RiskStudent{
			StudentID:     studentID,
			RiskLevel:     riskLevel,
			RiskFactors:   riskFactors,
			Probability:   probability,
			Interventions: interventions,
		})
	}

	return riskStudents, nil
}

func (s *advancedAnalyticsService) predictCourseCompletionRates(ctx context.Context, collegeID int) ([]CompletionRate, error) {
	query := `
		SELECT
			c.id as course_id,
			c.name as course_name,
			COUNT(DISTINCT e.student_id) as enrolled,
			COUNT(DISTINCT CASE WHEN g.percentage >= 40 THEN e.student_id END) as completed,
			AVG(EXTRACT(epoch FROM (CURRENT_DATE - c.created_at))/86400) as avg_duration_days
		FROM courses c
		LEFT JOIN enrollments e ON e.course_id = c.id AND e.college_id = c.college_id
		LEFT JOIN grades g ON g.course_id = c.id AND g.student_id = e.student_id AND g.college_id = c.college_id
		WHERE c.college_id = $1
		GROUP BY c.id, c.name, c.created_at`

	rows, err := s.db.Pool.Query(ctx, query, collegeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	completionRates := make([]CompletionRate, 0)
	for rows.Next() {
		var courseID int
		var courseName string
		var enrolled, completed int
		var avgDuration float64

		if err := rows.Scan(&courseID, &courseName, &enrolled, &completed, &avgDuration); err != nil {
			continue
		}

		completionRate := 0.0
		if enrolled > 0 {
			completionRate = float64(completed) / float64(enrolled) * 100
		}
		predictedRate := completionRate * 1.1 // Simplified prediction

		completionRates = append(completionRates, CompletionRate{
			CourseID:       courseID,
			CompletionRate: completionRate,
			PredictedRate:  predictedRate,
			TimeToComplete: int(avgDuration),
		})
	}

	return completionRates, nil
}

func (s *advancedAnalyticsService) generatePredictiveRecommendations(atRiskStudents []RiskStudent, completionRates []CompletionRate) []string {
	recommendations := make([]string, 0)

	if len(atRiskStudents) > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Monitor %d at-risk students closely", len(atRiskStudents)))
	}

	lowCompletionCourses := 0
	for _, rate := range completionRates {
		if rate.CompletionRate < 70 {
			lowCompletionCourses++
		}
	}

	if lowCompletionCourses > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Review %d courses with low completion rates", lowCompletionCourses))
	}

	return recommendations
}

func (s *advancedAnalyticsService) calculateRiskProbability(avgGrade, attendanceRate float64, recentGrades, recentAttendance int) float64 {
	cfg := s.analyticsConfig
	score := 0.0

	switch {
	case avgGrade < 40:
		score += cfg.RiskWeightGradeVeryLow
	case avgGrade < 60:
		score += cfg.RiskWeightGradeLow
	case avgGrade < 70:
		score += cfg.RiskWeightGradeMedium
	}

	switch {
	case attendanceRate < 50:
		score += cfg.RiskWeightAttendanceVeryLow
	case attendanceRate < 70:
		score += cfg.RiskWeightAttendanceLow
	case attendanceRate < 80:
		score += cfg.RiskWeightAttendanceMedium
	}

	switch {
	case recentGrades == 0:
		score += cfg.RiskWeightNoRecentGrades
	case recentGrades < 2:
		score += cfg.RiskWeightFewRecentGrades
	}

	if recentAttendance == 0 {
		score += cfg.RiskWeightNoRecentAttendance
	}

	if score < cfg.RiskMinScore {
		score = cfg.RiskMinScore
	}
	if score > cfg.RiskMaxScore {
		score = cfg.RiskMaxScore
	}

	return math.Round(score*100) / 100
}

func (s *advancedAnalyticsService) getTotalStudentsInCollege(ctx context.Context, collegeID int) (int, error) {
	var count int
	err := s.db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM students WHERE college_id = $1", collegeID).Scan(&count)
	return count, err
}

func (s *advancedAnalyticsService) getTotalCoursesInCollege(ctx context.Context, collegeID int) (int, error) {
	var count int
	err := s.db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM courses WHERE college_id = $1", collegeID).Scan(&count)
	return count, err
}

func (s *advancedAnalyticsService) getTopPerformingCourses(ctx context.Context, collegeID int, limit int) ([]CoursePerformance, error) {
	query := `
		SELECT
			c.id, c.name,
			COALESCE(AVG(g.percentage), 0) as avg_grade,
			COUNT(DISTINCT CASE WHEN g.percentage >= 40 THEN e.student_id END)::float / COUNT(DISTINCT e.student_id) * 100 as completion_rate
		FROM courses c
		LEFT JOIN enrollments e ON e.course_id = c.id AND e.college_id = c.college_id
		LEFT JOIN grades g ON g.course_id = c.id AND g.student_id = e.student_id AND g.college_id = c.college_id
		WHERE c.college_id = $1
		GROUP BY c.id, c.name
		ORDER BY avg_grade DESC
		LIMIT $2`

	rows, err := s.db.Pool.Query(ctx, query, collegeID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	courses := make([]CoursePerformance, 0)
	for rows.Next() {
		var course CoursePerformance
		if err := rows.Scan(&course.CourseID, &course.CourseName, &course.AverageGrade, &course.CompletionRate); err != nil {
			continue
		}
		courses = append(courses, course)
	}

	return courses, nil
}

func (s *advancedAnalyticsService) identifyLearningPatterns(ctx context.Context, collegeID int) ([]LearningPattern, error) {
	patterns := make([]LearningPattern, 0)

	// Pattern 1: High attendance correlation with success
	highAttendanceQuery := `
		SELECT
			COUNT(DISTINCT s.id) as student_count,
			COALESCE(AVG(CASE WHEN g.percentage >= 60 THEN 100 ELSE 0 END), 0) as success_rate
		FROM students s
		JOIN attendance a ON a.student_id = s.id AND a.college_id = s.college_id
		LEFT JOIN grades g ON g.student_id = s.id AND g.college_id = s.college_id
		WHERE s.college_id = $1
		GROUP BY s.id
		HAVING COALESCE(AVG(CASE WHEN a.status = 'Present' THEN 100 ELSE 0 END), 0) >= 80`

	var studentCount int
	var successRate float64
	err := s.db.Pool.QueryRow(ctx, highAttendanceQuery, collegeID).Scan(&studentCount, &successRate)
	if err == nil && studentCount > 0 {
		patterns = append(patterns, LearningPattern{
			Pattern:     "High Attendance (≥80%)",
			Students:    studentCount,
			SuccessRate: successRate,
		})
	}

	// Pattern 2: Regular assignment submission
	regularSubmissionQuery := `
		SELECT
			COUNT(DISTINCT s.student_id) as student_count,
			COALESCE(AVG(CASE WHEN g.percentage >= 60 THEN 100 ELSE 0 END), 0) as success_rate
		FROM (
			SELECT student_id, COUNT(*) as submission_count
			FROM assignment_submissions
			WHERE college_id = $1 AND created_at >= CURRENT_DATE - INTERVAL '60 days'
			GROUP BY student_id
			HAVING COUNT(*) >= 5
		) s
		LEFT JOIN grades g ON g.student_id = s.student_id AND g.college_id = $1`

	err = s.db.Pool.QueryRow(ctx, regularSubmissionQuery, collegeID).Scan(&studentCount, &successRate)
	if err == nil && studentCount > 0 {
		patterns = append(patterns, LearningPattern{
			Pattern:     "Regular Assignment Submissions",
			Students:    studentCount,
			SuccessRate: successRate,
		})
	}

	// Pattern 3: Active quiz participation
	activeQuizQuery := `
		SELECT
			COUNT(DISTINCT qa.student_id) as student_count,
			COALESCE(AVG(CASE WHEN g.percentage >= 60 THEN 100 ELSE 0 END), 0) as success_rate
		FROM (
			SELECT student_id, COUNT(*) as quiz_count
			FROM quiz_attempts
			WHERE college_id = $1 AND created_at >= CURRENT_DATE - INTERVAL '60 days'
			GROUP BY student_id
			HAVING COUNT(*) >= 3
		) qa
		LEFT JOIN grades g ON g.student_id = qa.student_id AND g.college_id = $1`

	err = s.db.Pool.QueryRow(ctx, activeQuizQuery, collegeID).Scan(&studentCount, &successRate)
	if err == nil && studentCount > 0 {
		patterns = append(patterns, LearningPattern{
			Pattern:     "Active Quiz Participation",
			Students:    studentCount,
			SuccessRate: successRate,
		})
	}

	return patterns, nil
}

func (s *advancedAnalyticsService) compareCourses(ctx context.Context, collegeID, courseID1, courseID2 int) (*CourseComparison, error) {
	comparison := &CourseComparison{
		CourseID1: courseID1,
		CourseID2: courseID2,
		Metrics:   make(map[string]float64),
	}

	// Get course names
	var name1, name2 string
	err := s.db.Pool.QueryRow(ctx, "SELECT name FROM courses WHERE id = $1 AND college_id = $2", courseID1, collegeID).Scan(&name1)
	if err != nil {
		return nil, err
	}
	err = s.db.Pool.QueryRow(ctx, "SELECT name FROM courses WHERE id = $1 AND college_id = $2", courseID2, collegeID).Scan(&name2)
	if err != nil {
		return nil, err
	}

	comparison.CourseName1 = name1
	comparison.CourseName2 = name2

	// Compare various metrics
	metrics := []string{"avg_grade", "attendance_rate", "completion_rate", "engagement_rate"}

	for _, metric := range metrics {
		var value1, value2 float64

		switch metric {
		case "avg_grade":
			s.db.Pool.QueryRow(ctx, "SELECT COALESCE(AVG(percentage),0) FROM grades WHERE course_id = $1 AND college_id = $2", courseID1, collegeID).Scan(&value1)
			s.db.Pool.QueryRow(ctx, "SELECT COALESCE(AVG(percentage),0) FROM grades WHERE course_id = $1 AND college_id = $2", courseID2, collegeID).Scan(&value2)
		case "attendance_rate":
			s.db.Pool.QueryRow(ctx, "SELECT COALESCE(AVG(CASE WHEN status = 'Present' THEN 100 ELSE 0 END),0) FROM attendance WHERE course_id = $1 AND college_id = $2", courseID1, collegeID).Scan(&value1)
			s.db.Pool.QueryRow(ctx, "SELECT COALESCE(AVG(CASE WHEN status = 'Present' THEN 100 ELSE 0 END),0) FROM attendance WHERE course_id = $1 AND college_id = $2", courseID2, collegeID).Scan(&value2)
		}

		comparison.Metrics[metric+"_1"] = value1
		comparison.Metrics[metric+"_2"] = value2

		// Check for significant differences
		if abs(value1-value2) > 10 { // More than 10% difference
			if value1 > value2 {
				comparison.SignificantDiff = append(comparison.SignificantDiff, fmt.Sprintf("%s performs better in %s (%.1f vs %.1f)", name1, metric, value1, value2))
			} else {
				comparison.SignificantDiff = append(comparison.SignificantDiff, fmt.Sprintf("%s performs better in %s (%.1f vs %.1f)", name2, metric, value2, value1))
			}
		}
	}

	return comparison, nil
}

func (s *advancedAnalyticsService) generateComparativeInsights(comparisons []CourseComparison) []string {
	insights := make([]string, 0)

	if len(comparisons) == 0 {
		return insights
	}

	totalComparisons := len(comparisons)
	significantDiffs := 0
	for _, comp := range comparisons {
		significantDiffs += len(comp.SignificantDiff)
	}

	if significantDiffs > 0 {
		insights = append(insights, fmt.Sprintf("Found %d significant performance differences across %d course comparisons", significantDiffs, totalComparisons))
	}

	return insights
}

func (s *advancedAnalyticsService) generateComparativeRecommendations(comparisons []CourseComparison) []string {
	recommendations := make([]string, 0)

	// Analyze patterns and generate recommendations
	// This is a simplified implementation

	if len(comparisons) > 0 {
		recommendations = append(recommendations, "Consider sharing best practices between courses with significant performance differences")
		recommendations = append(recommendations, "Review course materials and teaching methods for underperforming courses")
	}

	return recommendations
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func (s *advancedAnalyticsService) getStudentPerformanceTrends(ctx context.Context, collegeID, studentID int) ([]PerformanceTrend, error) {
	trends := make([]PerformanceTrend, 0)

	// Grade trends over time
	gradeQuery := `
		SELECT
			DATE_TRUNC('week', created_at) as week,
			AVG(percentage) as avg_grade
		FROM grades
		WHERE college_id = $1 AND student_id = $2
		GROUP BY DATE_TRUNC('week', created_at)
		ORDER BY week DESC
		LIMIT 12`

	rows, err := s.db.Pool.Query(ctx, gradeQuery, collegeID, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prevValue float64
	first := true
	for rows.Next() {
		var date time.Time
		var value float64
		if err := rows.Scan(&date, &value); err != nil {
			continue
		}

		changeRate := 0.0
		if !first {
			changeRate = ((value - prevValue) / prevValue) * 100
		}

		trends = append(trends, PerformanceTrend{
			Date:       date,
			Metric:     "average_grade",
			Value:      value,
			ChangeRate: changeRate,
		})

		prevValue = value
		first = false
	}

	// Attendance trends
	attendanceQuery := `
		SELECT
			DATE_TRUNC('week', date) as week,
			COALESCE(AVG(CASE WHEN status = 'Present' THEN 100 ELSE 0 END), 0) as attendance_rate
		FROM attendance
		WHERE college_id = $1 AND student_id = $2
		GROUP BY DATE_TRUNC('week', date)
		ORDER BY week DESC
		LIMIT 12`

	rows, err = s.db.Pool.Query(ctx, attendanceQuery, collegeID, studentID)
	if err != nil {
		return trends, nil
	}
	defer rows.Close()

	prevValue = 0
	first = true
	for rows.Next() {
		var date time.Time
		var value float64
		if err := rows.Scan(&date, &value); err != nil {
			continue
		}

		changeRate := 0.0
		if !first && prevValue > 0 {
			changeRate = ((value - prevValue) / prevValue) * 100
		}

		trends = append(trends, PerformanceTrend{
			Date:       date,
			Metric:     "attendance_rate",
			Value:      value,
			ChangeRate: changeRate,
		})

		prevValue = value
		first = false
	}

	return trends, nil
}

func (s *advancedAnalyticsService) getCoursePerformanceTrends(ctx context.Context, collegeID, courseID int) ([]PerformanceTrend, error) {
	trends := make([]PerformanceTrend, 0)

	// Average grade trends for course
	gradeQuery := `
		SELECT
			DATE_TRUNC('week', created_at) as week,
			AVG(percentage) as avg_grade
		FROM grades
		WHERE college_id = $1 AND course_id = $2
		GROUP BY DATE_TRUNC('week', created_at)
		ORDER BY week DESC
		LIMIT 12`

	rows, err := s.db.Pool.Query(ctx, gradeQuery, collegeID, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prevValue float64
	first := true
	for rows.Next() {
		var date time.Time
		var value float64
		if err := rows.Scan(&date, &value); err != nil {
			continue
		}

		changeRate := 0.0
		if !first && prevValue > 0 {
			changeRate = ((value - prevValue) / prevValue) * 100
		}

		trends = append(trends, PerformanceTrend{
			Date:       date,
			Metric:     "course_average_grade",
			Value:      value,
			ChangeRate: changeRate,
		})

		prevValue = value
		first = false
	}

	// Enrollment trends
	enrollmentQuery := `
		SELECT
			DATE_TRUNC('month', created_at) as month,
			COUNT(*) as enrollment_count
		FROM enrollments
		WHERE college_id = $1 AND course_id = $2
		GROUP BY DATE_TRUNC('month', created_at)
		ORDER BY month DESC
		LIMIT 6`

	rows, err = s.db.Pool.Query(ctx, enrollmentQuery, collegeID, courseID)
	if err != nil {
		return trends, nil
	}
	defer rows.Close()

	prevValue = 0
	first = true
	for rows.Next() {
		var date time.Time
		var value float64
		if err := rows.Scan(&date, &value); err != nil {
			continue
		}

		changeRate := 0.0
		if !first && prevValue > 0 {
			changeRate = ((value - prevValue) / prevValue) * 100
		}

		trends = append(trends, PerformanceTrend{
			Date:       date,
			Metric:     "enrollment_count",
			Value:      value,
			ChangeRate: changeRate,
		})

		prevValue = value
		first = false
	}

	return trends, nil
}

// getSkillDevelopment calculates skill development over time based on grades and performance
func (s *advancedAnalyticsService) getSkillDevelopment(ctx context.Context, collegeID, studentID int) ([]SkillPoint, error) {
	query := `
		SELECT 
			c.name as skill_area,
			DATE_TRUNC('month', g.created_at) as month,
			AVG(g.percentage) as avg_score
		FROM grades g
		JOIN courses c ON c.id = g.course_id AND c.college_id = g.college_id
		WHERE g.college_id = $1 AND g.student_id = $2
		GROUP BY c.name, DATE_TRUNC('month', g.created_at)
		ORDER BY month DESC
		LIMIT 12`

	rows, err := s.db.Pool.Query(ctx, query, collegeID, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	skillPoints := make([]SkillPoint, 0)
	for rows.Next() {
		var skillArea string
		var month time.Time
		var avgScore float64

		if err := rows.Scan(&skillArea, &month, &avgScore); err != nil {
			continue
		}

		skillPoints = append(skillPoints, SkillPoint{
			Skill: skillArea,
			Level: avgScore / 25, // Convert to 0-4 scale
			Date:  month,
		})
	}

	return skillPoints, nil
}

// getEngagementTimeline retrieves engagement data over time for a course
func (s *advancedAnalyticsService) getEngagementTimeline(ctx context.Context, collegeID, courseID int) ([]EngagementPoint, error) {
	query := `
		SELECT 
			DATE_TRUNC('week', activity_date) as week,
			active_users,
			assignments_done,
			quizzes_taken,
			forum_posts
		FROM (
			-- Attendance activity
			SELECT date as activity_date, COUNT(DISTINCT student_id) as active_users, 0 as assignments_done, 0 as quizzes_taken, 0 as forum_posts
			FROM attendance WHERE college_id = $1 AND course_id = $2
			GROUP BY date
			UNION ALL
			-- Assignment submissions
			SELECT DATE(s.created_at) as activity_date, 0 as active_users, COUNT(*) as assignments_done, 0 as quizzes_taken, 0 as forum_posts
			FROM assignment_submissions s 
			JOIN assignments a ON a.id = s.assignment_id 
			WHERE a.college_id = $1 AND a.course_id = $2
			GROUP BY DATE(s.created_at)
			UNION ALL
			-- Quiz attempts
			SELECT DATE(qa.created_at) as activity_date, 0 as active_users, 0 as assignments_done, COUNT(*) as quizzes_taken, 0 as forum_posts
			FROM quiz_attempts qa 
			JOIN quizzes q ON q.id = qa.quiz_id 
			WHERE qa.college_id = $1 AND q.course_id = $2
			GROUP BY DATE(qa.created_at)
		) weekly_activity
		GROUP BY DATE_TRUNC('week', activity_date), active_users, assignments_done, quizzes_taken, forum_posts
		ORDER BY week DESC
		LIMIT 12`

	rows, err := s.db.Pool.Query(ctx, query, collegeID, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	timeline := make([]EngagementPoint, 0)
	for rows.Next() {
		var week time.Time
		var activeUsers, assignmentsDone, quizzesTaken, forumPosts int

		if err := rows.Scan(&week, &activeUsers, &assignmentsDone, &quizzesTaken, &forumPosts); err != nil {
			continue
		}

		timeline = append(timeline, EngagementPoint{
			Date:            week,
			ActiveUsers:     activeUsers,
			AssignmentsDone: assignmentsDone,
			QuizzesTaken:    quizzesTaken,
			ForumPosts:      forumPosts,
		})
	}

	return timeline, nil
}
