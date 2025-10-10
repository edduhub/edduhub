CREATE TABLE IF NOT EXISTS grades (
    id SERIAL PRIMARY KEY,
    student_id INTEGER NOT NULL REFERENCES students(student_id) ON DELETE CASCADE,
    course_id INTEGER NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    college_id INTEGER NOT NULL REFERENCES colleges(id) ON DELETE CASCADE,
    assessment_name VARCHAR(200) NOT NULL,
    assessment_type VARCHAR(50) NOT NULL CHECK (assessment_type IN ('quiz', 'assignment', 'midterm', 'final', 'project', 'other')),
    total_marks INTEGER NOT NULL,
    obtained_marks INTEGER NOT NULL,
    percentage DECIMAL(5,2),
    grade VARCHAR(5),
    remarks TEXT,
    graded_by VARCHAR(255),
    graded_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_marks CHECK (obtained_marks <= total_marks AND obtained_marks >= 0)
);

CREATE INDEX idx_grades_student ON grades(student_id);
CREATE INDEX idx_grades_course ON grades(course_id);
CREATE INDEX idx_grades_college ON grades(college_id);
CREATE INDEX idx_grades_type ON grades(assessment_type);
