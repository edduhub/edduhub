-- Drop triggers
DROP TRIGGER IF EXISTS update_revaluation_requests_timestamp ON revaluation_requests;
DROP TRIGGER IF EXISTS update_exam_results_timestamp ON exam_results;
DROP TRIGGER IF EXISTS update_exam_enrollments_timestamp ON exam_enrollments;
DROP TRIGGER IF EXISTS update_exams_timestamp ON exams;
DROP TRIGGER IF EXISTS update_exam_rooms_timestamp ON exam_rooms;

-- Drop function
DROP FUNCTION IF EXISTS update_exam_management_updated_at();

-- Drop tables in reverse order (respecting foreign key constraints)
DROP TABLE IF EXISTS revaluation_requests;
DROP TABLE IF EXISTS exam_results;
DROP TABLE IF EXISTS exam_enrollments;
DROP TABLE IF EXISTS exams;
DROP TABLE IF EXISTS exam_rooms;
