BEGIN;

CREATE TABLE IF NOT EXISTS announcements (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    title VARCHAR(150) NOT NULL,
    content TEXT NOT NULL,
    college_id INT NOT NULL,
    user_id INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_college
        FOREIGN KEY(college_id)
        REFERENCES colleges(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_user
        FOREIGN KEY(user_id)
        REFERENCES users(id)
        ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_announcements_college_id ON announcements (college_id);
CREATE INDEX IF NOT EXISTS idx_announcements_user_id ON announcements (user_id);


COMMIT;
