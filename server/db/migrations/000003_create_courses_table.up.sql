BEGIN;

CREATE TABLE IF NOT EXISTS courses (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    college_id INT NOT NULL,
    credits INT NOT NULL,
    instructor_id INT NOT NULL, -- Required foreign key to instructor
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Foreign key to users table (assuming instructors are users)
    CONSTRAINT fk_courses_instructor
        FOREIGN KEY (instructor_id)
        REFERENCES users(id)
        ON DELETE RESTRICT,
    -- Foreign key to colleges table
    CONSTRAINT fk_courses_college
        FOREIGN KEY (college_id)
        REFERENCES colleges(id)
        ON DELETE RESTRICT
    );

-- Index for foreign key and potential lookups by name
CREATE INDEX IF NOT EXISTS idx_courses_instructor_id ON courses (instructor_id);
CREATE INDEX IF NOT EXISTS idx_courses_name ON courses (name);
CREATE INDEX IF NOT EXISTS idx_courses_college_id ON courses (college_id);

COMMIT;
