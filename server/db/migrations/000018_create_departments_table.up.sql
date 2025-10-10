CREATE TABLE IF NOT EXISTS departments (
    id SERIAL PRIMARY KEY,
    college_id INTEGER NOT NULL REFERENCES colleges(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    code VARCHAR(20) NOT NULL,
    description TEXT,
    head_user_id INTEGER,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(college_id, code)
);

CREATE INDEX idx_departments_college ON departments(college_id);
CREATE INDEX idx_departments_head ON departments(head_user_id);
CREATE INDEX idx_departments_active ON departments(is_active);
