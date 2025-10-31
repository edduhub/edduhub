BEGIN;

-- Additional indexes for analytics performance optimization
-- These indexes are specifically designed to improve the performance of analytics queries

-- Attendance analytics indexes
CREATE INDEX IF NOT EXISTS idx_attendance_college_course_date ON attendance (college_id, course_id, date);
CREATE INDEX IF NOT EXISTS idx_attendance_student_college_course ON attendance (student_id, college_id, course_id);
CREATE INDEX IF NOT EXISTS idx_attendance_status_date ON attendance (status, date) WHERE status IN ('Present', 'Absent');
CREATE INDEX IF NOT EXISTS idx_attendance_college_date_status ON attendance (college_id, date, status);

-- Grades analytics indexes
CREATE INDEX IF NOT EXISTS idx_grades_college_course_student ON grades (college_id, course_id, student_id);
CREATE INDEX IF NOT EXISTS idx_grades_college_course_percentage ON grades (college_id, course_id, percentage);
CREATE INDEX IF NOT EXISTS idx_grades_student_college ON grades (student_id, college_id);
CREATE INDEX IF NOT EXISTS idx_grades_type_college_course ON grades (assessment_type, college_id, course_id);

-- Assignment analytics indexes
CREATE INDEX IF NOT EXISTS idx_assignments_college_course_due ON assignments (college_id, course_id, due_date);
CREATE INDEX IF NOT EXISTS idx_assignment_submissions_college ON assignment_submissions (assignment_id, student_id) 
WHERE assignment_id IN (SELECT id FROM assignments WHERE college_id IS NOT NULL);

-- Quiz analytics indexes
CREATE INDEX IF NOT EXISTS idx_quiz_attempts_college_status ON quiz_attempts (college_id, status);
CREATE INDEX IF NOT EXISTS idx_quiz_attempts_college_quiz_student ON quiz_attempts (college_id, quiz_id, student_id);

-- Composite indexes for complex analytics queries
CREATE INDEX IF NOT EXISTS idx_attendance_analytics_composite ON attendance (college_id, course_id, student_id, date, status);
CREATE INDEX IF NOT EXISTS idx_grades_analytics_composite ON grades (college_id, course_id, student_id, percentage, assessment_type);

-- Partial indexes for common analytics filters
CREATE INDEX IF NOT EXISTS idx_attendance_present_recent ON attendance (college_id, course_id, date) 
WHERE status = 'Present' AND date >= CURRENT_DATE - INTERVAL '90 days';

CREATE INDEX IF NOT EXISTS idx_grades_recent ON grades (college_id, course_id, created_at) 
WHERE created_at >= CURRENT_DATE - INTERVAL '90 days';

-- Enrollment analytics indexes (for course analytics)
CREATE INDEX IF NOT EXISTS idx_enrollments_college_course ON enrollments (college_id, course_id);
CREATE INDEX IF NOT EXISTS idx_enrollments_student_college ON enrollments (student_id, college_id);

COMMIT;