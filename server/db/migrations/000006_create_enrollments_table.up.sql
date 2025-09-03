BEGIN;

CREATE TABLE IF NOT EXISTS enrollments (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    student_id INT NOT NULL,
    course_id INT NOT NULL,
    college_id INT NOT NULL, -- Denormalized? Or specific to enrollment context?
    enrollment_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status VARCHAR(50) NOT NULL DEFAULT 'Active', -- e.g., Active, Completed, Dropped
    grade VARCHAR(10), -- Nullable
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Foreign Keys
    CONSTRAINT fk_enrollments_student
        FOREIGN KEY (student_id)
        REFERENCES students(student_id)
        ON DELETE CASCADE, -- If student is deleted, remove their enrollments
    CONSTRAINT fk_enrollments_course
        FOREIGN KEY (course_id)
        REFERENCES courses(id)
        ON DELETE RESTRICT, -- Don't delete course if students are enrolled
    CONSTRAINT fk_enrollments_college
        FOREIGN KEY (college_id)
        REFERENCES colleges(id)
        ON DELETE RESTRICT, -- Or CASCADE if enrollments are college-specific

    -- Prevent duplicate enrollments for the same student in the same course
    UNIQUE (student_id, course_id)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_enrollments_student_id ON enrollments (student_id);
CREATE INDEX IF NOT EXISTS idx_enrollments_course_id ON enrollments (course_id);
CREATE INDEX IF NOT EXISTS idx_enrollments_college_id ON enrollments (college_id);
CREATE INDEX IF NOT EXISTS idx_enrollments_student_course ON enrollments (student_id, course_id); -- For unique constraint

COMMIT;
