BEGIN;

DROP TRIGGER IF EXISTS update_revaluation_requests_runtime_updated_at ON revaluation_requests;
DROP TRIGGER IF EXISTS update_exam_results_runtime_updated_at ON exam_results;
DROP TRIGGER IF EXISTS update_exam_enrollments_runtime_updated_at ON exam_enrollments;
DROP TRIGGER IF EXISTS update_timetable_blocks_updated_at ON timetable_blocks;

DROP TABLE IF EXISTS revaluation_requests;
DROP TABLE IF EXISTS exam_results;
DROP TABLE IF EXISTS exam_enrollments;

DROP INDEX IF EXISTS idx_student_answers_quiz_attempt_question_unique;
DROP INDEX IF EXISTS idx_fee_assignments_student_structure_unique;
DROP INDEX IF EXISTS idx_timetable_blocks_department;
DROP INDEX IF EXISTS idx_timetable_blocks_course;
DROP INDEX IF EXISTS idx_timetable_blocks_college;

ALTER TABLE departments DROP COLUMN IF EXISTS hod;
ALTER TABLE student_answers DROP COLUMN IF EXISTS points_awarded;
ALTER TABLE student_answers DROP COLUMN IF EXISTS quiz_attempt_id;
ALTER TABLE answer_options DROP COLUMN IF EXISTS text;
ALTER TABLE questions DROP COLUMN IF EXISTS correct_answer;
ALTER TABLE questions DROP COLUMN IF EXISTS points;
ALTER TABLE questions DROP COLUMN IF EXISTS type;
ALTER TABLE questions DROP COLUMN IF EXISTS text;
ALTER TABLE quizzes DROP COLUMN IF EXISTS due_date;
ALTER TABLE quizzes DROP COLUMN IF EXISTS time_limit_minutes;
ALTER TABLE fee_assignments DROP COLUMN IF EXISTS due_date;
ALTER TABLE fee_assignments DROP COLUMN IF EXISTS waiver_reason;
ALTER TABLE fee_assignments DROP COLUMN IF EXISTS waiver_amount;
ALTER TABLE fee_structures DROP COLUMN IF EXISTS is_mandatory;
ALTER TABLE fee_structures DROP COLUMN IF EXISTS course_id;
ALTER TABLE fee_structures DROP COLUMN IF EXISTS department_id;
ALTER TABLE fee_structures DROP COLUMN IF EXISTS semester;
ALTER TABLE fee_structures DROP COLUMN IF EXISTS frequency;
ALTER TABLE timetable_blocks DROP COLUMN IF EXISTS faculty_id;
ALTER TABLE timetable_blocks DROP COLUMN IF EXISTS room_number;
ALTER TABLE timetable_blocks DROP COLUMN IF EXISTS class_id;
ALTER TABLE timetable_blocks DROP COLUMN IF EXISTS department_id;

COMMIT;
