BEGIN;

-- Timetable repository expects repository-era columns that are missing from the
-- runtime schema used by the current database.
CREATE TABLE IF NOT EXISTS timetable_blocks (
    id SERIAL PRIMARY KEY,
    college_id INTEGER NOT NULL REFERENCES colleges(id) ON DELETE CASCADE,
    department_id INTEGER REFERENCES departments(id) ON DELETE SET NULL,
    course_id INTEGER NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    class_id INTEGER,
    day_of_week INTEGER NOT NULL CHECK (day_of_week BETWEEN 0 AND 6),
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    room_number VARCHAR(100),
    faculty_id VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE timetable_blocks ADD COLUMN IF NOT EXISTS department_id INTEGER;
ALTER TABLE timetable_blocks ADD COLUMN IF NOT EXISTS course_id INTEGER;
ALTER TABLE timetable_blocks ADD COLUMN IF NOT EXISTS class_id INTEGER;
ALTER TABLE timetable_blocks ADD COLUMN IF NOT EXISTS room_number VARCHAR(100);
ALTER TABLE timetable_blocks ADD COLUMN IF NOT EXISTS faculty_id VARCHAR(255);
ALTER TABLE timetable_blocks ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE timetable_blocks ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

UPDATE timetable_blocks
SET room_number = COALESCE(room_number, location)
WHERE room_number IS NULL
  AND EXISTS (
      SELECT 1
      FROM information_schema.columns
      WHERE table_schema = 'public'
        AND table_name = 'timetable_blocks'
        AND column_name = 'location'
  );

CREATE INDEX IF NOT EXISTS idx_timetable_blocks_college ON timetable_blocks(college_id);
CREATE INDEX IF NOT EXISTS idx_timetable_blocks_course ON timetable_blocks(course_id);
CREATE INDEX IF NOT EXISTS idx_timetable_blocks_department ON timetable_blocks(department_id);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_trigger WHERE tgname = 'update_timetable_blocks_updated_at'
    ) THEN
        CREATE TRIGGER update_timetable_blocks_updated_at
            BEFORE UPDATE ON timetable_blocks
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();
    END IF;
END $$;

-- Fee module repositories expect a richer schema than the live database has.
ALTER TABLE fee_structures ADD COLUMN IF NOT EXISTS frequency VARCHAR(50) NOT NULL DEFAULT 'semester';
ALTER TABLE fee_structures ADD COLUMN IF NOT EXISTS semester VARCHAR(20);
ALTER TABLE fee_structures ADD COLUMN IF NOT EXISTS department_id INTEGER;
ALTER TABLE fee_structures ADD COLUMN IF NOT EXISTS course_id INTEGER;
ALTER TABLE fee_structures ADD COLUMN IF NOT EXISTS is_mandatory BOOLEAN NOT NULL DEFAULT TRUE;

ALTER TABLE fee_assignments ADD COLUMN IF NOT EXISTS waiver_amount DECIMAL(10, 2) NOT NULL DEFAULT 0;
ALTER TABLE fee_assignments ADD COLUMN IF NOT EXISTS due_date TIMESTAMPTZ;
ALTER TABLE fee_assignments ADD COLUMN IF NOT EXISTS waiver_reason TEXT;

UPDATE fee_structures
SET academic_year = COALESCE(academic_year, year::text)
WHERE academic_year IS NULL
  AND EXISTS (
      SELECT 1
      FROM information_schema.columns
      WHERE table_schema = 'public'
        AND table_name = 'fee_structures'
        AND column_name = 'year'
  );

UPDATE fee_assignments
SET due_date = COALESCE(due_date, payment_due_date)
WHERE due_date IS NULL
  AND EXISTS (
      SELECT 1
      FROM information_schema.columns
      WHERE table_schema = 'public'
        AND table_name = 'fee_assignments'
        AND column_name = 'payment_due_date'
  );

CREATE UNIQUE INDEX IF NOT EXISTS idx_fee_assignments_student_structure_unique
    ON fee_assignments(student_id, fee_structure_id);

-- Quiz/question repositories still query legacy compatibility columns.
ALTER TABLE quizzes ADD COLUMN IF NOT EXISTS time_limit_minutes INTEGER;
ALTER TABLE quizzes ADD COLUMN IF NOT EXISTS due_date TIMESTAMP;
UPDATE quizzes
SET time_limit_minutes = COALESCE(time_limit_minutes, duration_minutes),
    due_date = COALESCE(due_date, end_time)
WHERE time_limit_minutes IS NULL OR due_date IS NULL;

ALTER TABLE questions ADD COLUMN IF NOT EXISTS text TEXT;
ALTER TABLE questions ADD COLUMN IF NOT EXISTS type VARCHAR(50);
ALTER TABLE questions ADD COLUMN IF NOT EXISTS points INTEGER;
ALTER TABLE questions ADD COLUMN IF NOT EXISTS correct_answer TEXT;
UPDATE questions
SET text = COALESCE(text, question_text),
    type = COALESCE(type, question_type),
    points = COALESCE(points, marks)
WHERE text IS NULL OR type IS NULL OR points IS NULL;

ALTER TABLE answer_options ADD COLUMN IF NOT EXISTS text TEXT;
UPDATE answer_options
SET text = COALESCE(text, option_text)
WHERE text IS NULL;

ALTER TABLE student_answers ADD COLUMN IF NOT EXISTS quiz_attempt_id INTEGER;
ALTER TABLE student_answers ADD COLUMN IF NOT EXISTS points_awarded INTEGER;
UPDATE student_answers
SET quiz_attempt_id = COALESCE(quiz_attempt_id, attempt_id),
    points_awarded = COALESCE(points_awarded, marks_awarded)
WHERE quiz_attempt_id IS NULL OR points_awarded IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_student_answers_quiz_attempt_question_unique
    ON student_answers(quiz_attempt_id, question_id);

-- Department API/repository compatibility columns.
ALTER TABLE departments ADD COLUMN IF NOT EXISTS hod VARCHAR(255);
UPDATE departments d
SET hod = COALESCE(d.hod, u.name)
FROM users u
WHERE d.head_user_id = u.id
  AND d.hod IS NULL;

-- Exam management migration historically referenced the wrong student PK, so
-- create the runtime tables with the actual students(student_id) reference.
CREATE TABLE IF NOT EXISTS exam_enrollments (
    id SERIAL PRIMARY KEY,
    exam_id INTEGER NOT NULL REFERENCES exams(id) ON DELETE CASCADE,
    student_id INTEGER NOT NULL REFERENCES students(student_id) ON DELETE CASCADE,
    college_id INTEGER NOT NULL REFERENCES colleges(id) ON DELETE CASCADE,
    enrollment_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    seat_number VARCHAR(50),
    room_number VARCHAR(50),
    question_paper_set INTEGER,
    status VARCHAR(50) DEFAULT 'enrolled' CHECK (status IN ('enrolled', 'appeared', 'absent', 'disqualified')),
    hall_ticket_generated BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (exam_id, student_id)
);

CREATE TABLE IF NOT EXISTS exam_results (
    id SERIAL PRIMARY KEY,
    exam_id INTEGER NOT NULL REFERENCES exams(id) ON DELETE CASCADE,
    student_id INTEGER NOT NULL REFERENCES students(student_id) ON DELETE CASCADE,
    college_id INTEGER NOT NULL REFERENCES colleges(id) ON DELETE CASCADE,
    marks_obtained DECIMAL(10, 2) CHECK (marks_obtained >= 0),
    grade VARCHAR(10),
    percentage DECIMAL(5, 2) CHECK (percentage >= 0 AND percentage <= 100),
    result VARCHAR(50) DEFAULT 'pending' CHECK (result IN ('pass', 'fail', 'absent', 'pending')),
    remarks TEXT,
    evaluated_by INTEGER REFERENCES users(id),
    evaluated_at TIMESTAMP,
    revaluation_status VARCHAR(50) DEFAULT 'none' CHECK (revaluation_status IN ('none', 'requested', 'in_progress', 'completed')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (exam_id, student_id)
);

CREATE TABLE IF NOT EXISTS revaluation_requests (
    id SERIAL PRIMARY KEY,
    exam_result_id INTEGER NOT NULL REFERENCES exam_results(id) ON DELETE CASCADE,
    student_id INTEGER NOT NULL REFERENCES students(student_id) ON DELETE CASCADE,
    college_id INTEGER NOT NULL REFERENCES colleges(id) ON DELETE CASCADE,
    reason TEXT NOT NULL,
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'completed')),
    previous_marks DECIMAL(10, 2) NOT NULL,
    revised_marks DECIMAL(10, 2) CHECK (revised_marks >= 0),
    reviewed_by INTEGER REFERENCES users(id),
    review_comments TEXT,
    requested_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    reviewed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (exam_result_id, student_id)
);

CREATE INDEX IF NOT EXISTS idx_exam_enrollments_exam ON exam_enrollments(exam_id);
CREATE INDEX IF NOT EXISTS idx_exam_enrollments_student ON exam_enrollments(student_id);
CREATE INDEX IF NOT EXISTS idx_exam_results_exam ON exam_results(exam_id);
CREATE INDEX IF NOT EXISTS idx_exam_results_student ON exam_results(student_id);
CREATE INDEX IF NOT EXISTS idx_revaluation_requests_result ON revaluation_requests(exam_result_id);
CREATE INDEX IF NOT EXISTS idx_revaluation_requests_student ON revaluation_requests(student_id);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_trigger WHERE tgname = 'update_exam_enrollments_runtime_updated_at'
    ) THEN
        CREATE TRIGGER update_exam_enrollments_runtime_updated_at
            BEFORE UPDATE ON exam_enrollments
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_trigger WHERE tgname = 'update_exam_results_runtime_updated_at'
    ) THEN
        CREATE TRIGGER update_exam_results_runtime_updated_at
            BEFORE UPDATE ON exam_results
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_trigger WHERE tgname = 'update_revaluation_requests_runtime_updated_at'
    ) THEN
        CREATE TRIGGER update_revaluation_requests_runtime_updated_at
            BEFORE UPDATE ON revaluation_requests
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();
    END IF;
END $$;

COMMIT;
