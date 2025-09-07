BEGIN;

CREATE TABLE IF NOT EXISTS assignments (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    course_id INT NOT NULL,
    college_id INT NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    due_date TIMESTAMPTZ,
    max_points INT NOT NULL DEFAULT 100,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_assignments_course
        FOREIGN KEY (course_id)
        REFERENCES courses(id)
        ON DELETE CASCADE, -- If a course is deleted, its assignments are also deleted.

    CONSTRAINT fk_assignments_college
        FOREIGN KEY (college_id)
        REFERENCES colleges(id)
        ON DELETE RESTRICT -- Prevent deleting a college if it has assignments.
);

CREATE TABLE IF NOT EXISTS assignment_submissions (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    assignment_id INT NOT NULL,
    student_id INT NOT NULL,
    submission_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    content_text TEXT,
    file_path VARCHAR(512), -- Path to the stored file (e.g., in S3 or local storage)
    grade INT,
    feedback TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_submissions_assignment
        FOREIGN KEY (assignment_id)
        REFERENCES assignments(id)
        ON DELETE CASCADE, -- If an assignment is deleted, delete submissions.

    CONSTRAINT fk_submissions_student
        FOREIGN KEY (student_id)
        REFERENCES students(student_id)
        ON DELETE CASCADE, -- If a student is deleted, delete their submissions.

    UNIQUE (assignment_id, student_id) -- A student can only have one submission per assignment.
);

CREATE INDEX IF NOT EXISTS idx_assignments_course_id ON assignments (course_id);
CREATE INDEX IF NOT EXISTS idx_assignment_submissions_assignment_id ON assignment_submissions (assignment_id);
CREATE INDEX IF NOT EXISTS idx_assignment_submissions_student_id ON assignment_submissions (student_id);

COMMIT;
