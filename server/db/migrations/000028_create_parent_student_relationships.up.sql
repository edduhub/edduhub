-- Migration: Create parent-student relationships table
-- This enables the parent portal feature

CREATE TABLE IF NOT EXISTS parent_student_relationships (
    id SERIAL PRIMARY KEY,
    parent_user_id INTEGER NOT NULL,
    student_id INTEGER NOT NULL REFERENCES students(student_id) ON DELETE CASCADE,
    college_id INTEGER NOT NULL,
    relation VARCHAR(20) NOT NULL CHECK (relation IN ('father', 'mother', 'guardian')),
    is_primary_contact BOOLEAN DEFAULT TRUE,
    receive_notifications BOOLEAN DEFAULT TRUE,
    is_verified BOOLEAN DEFAULT FALSE,
    verified_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Ensure unique parent-student pairs per college
    UNIQUE(parent_user_id, student_id, college_id)
);

-- Indexes for efficient querying
CREATE INDEX idx_parent_relationships_parent ON parent_student_relationships(parent_user_id);
CREATE INDEX idx_parent_relationships_student ON parent_student_relationships(student_id);
CREATE INDEX idx_parent_relationships_college ON parent_student_relationships(college_id);

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_parent_relationships_updated_at 
    BEFORE UPDATE ON parent_student_relationships 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();
