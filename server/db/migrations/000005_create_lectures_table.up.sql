BEGIN;

CREATE TABLE IF NOT EXISTS lectures (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    course_id INT NOT NULL,
    college_id INT NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    meeting_link VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_lectures_college
        FOREIGN KEY (college_id)
        REFERENCES colleges(id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_lectures_course
        FOREIGN KEY (course_id)
        REFERENCES courses(id)
        ON DELETE CASCADE -- If course is deleted, maybe delete lectures? Or RESTRICT?
    -- CONSTRAINT fk_lectures_qrcode -- Add if qr_code_id is indeed a FK
    --    FOREIGN KEY (qr_code_id)
    --    REFERENCES qrcodes(id)
    --    ON DELETE SET NULL
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_lectures_college_id ON lectures (college_id);
CREATE INDEX IF NOT EXISTS idx_lectures_course_id ON lectures (course_id);
-- CREATE INDEX IF NOT EXISTS idx_lectures_qr_code_id ON lectures (qr_code_id);
CREATE INDEX IF NOT EXISTS idx_lectures_start_time ON lectures (start_time);
CREATE INDEX IF NOT EXISTS idx_lectures_end_time ON lectures (end_time);

COMMIT;
