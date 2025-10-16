BEGIN;

-- Drop indexes first
DROP INDEX IF EXISTS idx_assignment_submissions_student_id;
DROP INDEX IF EXISTS idx_assignment_submissions_assignment_id;
DROP INDEX IF EXISTS idx_assignments_course_id;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS assignment_submissions;
DROP TABLE IF EXISTS assignments;

COMMIT;
