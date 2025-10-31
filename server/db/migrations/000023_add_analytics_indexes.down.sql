BEGIN;

-- Drop analytics optimization indexes
DROP INDEX IF EXISTS idx_attendance_college_course_date;
DROP INDEX IF EXISTS idx_attendance_student_college_course;
DROP INDEX IF EXISTS idx_attendance_status_date;
DROP INDEX IF EXISTS idx_attendance_college_date_status;
DROP INDEX IF EXISTS idx_grades_college_course_student;
DROP INDEX IF EXISTS idx_grades_college_course_percentage;
DROP INDEX IF EXISTS idx_grades_student_college;
DROP INDEX IF EXISTS idx_grades_type_college_course;
DROP INDEX IF EXISTS idx_assignments_college_course_due;
DROP INDEX IF EXISTS idx_assignment_submissions_college;
DROP INDEX IF EXISTS idx_quiz_attempts_college_status;
DROP INDEX IF EXISTS idx_quiz_attempts_college_quiz_student;
DROP INDEX IF EXISTS idx_attendance_analytics_composite;
DROP INDEX IF EXISTS idx_grades_analytics_composite;
DROP INDEX IF EXISTS idx_attendance_present_recent;
DROP INDEX IF EXISTS idx_grades_recent;
DROP INDEX IF EXISTS idx_enrollments_college_course;
DROP INDEX IF EXISTS idx_enrollments_student_college;

COMMIT;