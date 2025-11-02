-- Create exam_rooms table for managing physical exam rooms/halls
CREATE TABLE IF NOT EXISTS exam_rooms (
    id SERIAL PRIMARY KEY,
    college_id INTEGER NOT NULL REFERENCES colleges(id) ON DELETE CASCADE,
    room_number VARCHAR(50) NOT NULL,
    room_name VARCHAR(255) NOT NULL,
    capacity INTEGER NOT NULL CHECK (capacity > 0),
    location VARCHAR(255),
    facilities TEXT, -- JSON or comma-separated list of facilities
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(college_id, room_number)
);

-- Create index for faster room lookups
CREATE INDEX idx_exam_rooms_college ON exam_rooms(college_id);
CREATE INDEX idx_exam_rooms_active ON exam_rooms(college_id, is_active);

-- Create exams table for managing formal examinations
CREATE TABLE IF NOT EXISTS exams (
    id SERIAL PRIMARY KEY,
    college_id INTEGER NOT NULL REFERENCES colleges(id) ON DELETE CASCADE,
    course_id INTEGER NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    exam_type VARCHAR(50) NOT NULL CHECK (exam_type IN ('midterm', 'final', 'quiz', 'practical')),
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    duration INTEGER NOT NULL CHECK (duration > 0), -- Duration in minutes
    total_marks DECIMAL(10, 2) NOT NULL CHECK (total_marks > 0),
    passing_marks DECIMAL(10, 2) NOT NULL CHECK (passing_marks >= 0),
    room_id INTEGER REFERENCES exam_rooms(id) ON DELETE SET NULL,
    status VARCHAR(50) DEFAULT 'scheduled' CHECK (status IN ('scheduled', 'ongoing', 'completed', 'cancelled')),
    instructions TEXT,
    allowed_materials TEXT,
    question_paper_sets INTEGER DEFAULT 1 CHECK (question_paper_sets > 0),
    created_by INTEGER NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_exam_time CHECK (end_time > start_time),
    CONSTRAINT check_passing_marks CHECK (passing_marks <= total_marks)
);

-- Create indexes for exam queries
CREATE INDEX idx_exams_college ON exams(college_id);
CREATE INDEX idx_exams_course ON exams(course_id);
CREATE INDEX idx_exams_status ON exams(status);
CREATE INDEX idx_exams_start_time ON exams(start_time);

-- Create exam_enrollments table for student exam registrations
CREATE TABLE IF NOT EXISTS exam_enrollments (
    id SERIAL PRIMARY KEY,
    exam_id INTEGER NOT NULL REFERENCES exams(id) ON DELETE CASCADE,
    student_id INTEGER NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    college_id INTEGER NOT NULL REFERENCES colleges(id) ON DELETE CASCADE,
    enrollment_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    seat_number VARCHAR(50),
    room_number VARCHAR(50),
    question_paper_set INTEGER,
    status VARCHAR(50) DEFAULT 'enrolled' CHECK (status IN ('enrolled', 'appeared', 'absent', 'disqualified')),
    hall_ticket_generated BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(exam_id, student_id)
);

-- Create indexes for enrollment queries
CREATE INDEX idx_exam_enrollments_exam ON exam_enrollments(exam_id);
CREATE INDEX idx_exam_enrollments_student ON exam_enrollments(student_id);
CREATE INDEX idx_exam_enrollments_college ON exam_enrollments(college_id);
CREATE INDEX idx_exam_enrollments_status ON exam_enrollments(status);

-- Create exam_results table for storing exam results
CREATE TABLE IF NOT EXISTS exam_results (
    id SERIAL PRIMARY KEY,
    exam_id INTEGER NOT NULL REFERENCES exams(id) ON DELETE CASCADE,
    student_id INTEGER NOT NULL REFERENCES students(id) ON DELETE CASCADE,
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
    UNIQUE(exam_id, student_id)
);

-- Create indexes for result queries
CREATE INDEX idx_exam_results_exam ON exam_results(exam_id);
CREATE INDEX idx_exam_results_student ON exam_results(student_id);
CREATE INDEX idx_exam_results_college ON exam_results(college_id);
CREATE INDEX idx_exam_results_result ON exam_results(result);

-- Create revaluation_requests table for exam re-evaluation requests
CREATE TABLE IF NOT EXISTS revaluation_requests (
    id SERIAL PRIMARY KEY,
    exam_result_id INTEGER NOT NULL REFERENCES exam_results(id) ON DELETE CASCADE,
    student_id INTEGER NOT NULL REFERENCES students(id) ON DELETE CASCADE,
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
    UNIQUE(exam_result_id, student_id)
);

-- Create indexes for revaluation queries
CREATE INDEX idx_revaluation_requests_result ON revaluation_requests(exam_result_id);
CREATE INDEX idx_revaluation_requests_student ON revaluation_requests(student_id);
CREATE INDEX idx_revaluation_requests_status ON revaluation_requests(status);
CREATE INDEX idx_revaluation_requests_college ON revaluation_requests(college_id);

-- Create function to auto-update updated_at timestamp
CREATE OR REPLACE FUNCTION update_exam_management_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for auto-updating updated_at
CREATE TRIGGER update_exam_rooms_timestamp
    BEFORE UPDATE ON exam_rooms
    FOR EACH ROW
    EXECUTE FUNCTION update_exam_management_updated_at();

CREATE TRIGGER update_exams_timestamp
    BEFORE UPDATE ON exams
    FOR EACH ROW
    EXECUTE FUNCTION update_exam_management_updated_at();

CREATE TRIGGER update_exam_enrollments_timestamp
    BEFORE UPDATE ON exam_enrollments
    FOR EACH ROW
    EXECUTE FUNCTION update_exam_management_updated_at();

CREATE TRIGGER update_exam_results_timestamp
    BEFORE UPDATE ON exam_results
    FOR EACH ROW
    EXECUTE FUNCTION update_exam_management_updated_at();

CREATE TRIGGER update_revaluation_requests_timestamp
    BEFORE UPDATE ON revaluation_requests
    FOR EACH ROW
    EXECUTE FUNCTION update_exam_management_updated_at();

-- Add comments for documentation
COMMENT ON TABLE exams IS 'Stores formal examination information';
COMMENT ON TABLE exam_enrollments IS 'Tracks student enrollments in exams with seat allocation';
COMMENT ON TABLE exam_results IS 'Stores exam results and evaluation status';
COMMENT ON TABLE revaluation_requests IS 'Manages exam re-evaluation/recheck requests';
COMMENT ON TABLE exam_rooms IS 'Physical rooms/halls for conducting exams';
