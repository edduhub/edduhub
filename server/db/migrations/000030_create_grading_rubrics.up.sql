-- Migration: Create faculty grading rubrics table
-- This enables faculty rubrics management

CREATE TABLE IF NOT EXISTS grading_rubrics (
    id SERIAL PRIMARY KEY,
    faculty_id INTEGER NOT NULL,
    college_id INTEGER NOT NULL,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    course_id INTEGER,
    is_template BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    max_score INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS rubric_criteria (
    id SERIAL PRIMARY KEY,
    rubric_id INTEGER NOT NULL REFERENCES grading_rubrics(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    weight DECIMAL(5,2) NOT NULL DEFAULT 1.0,
    max_score INTEGER NOT NULL,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS rubric_performance_levels (
    id SERIAL PRIMARY KEY,
    rubric_id INTEGER NOT NULL REFERENCES grading_rubrics(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    points INTEGER NOT NULL,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_rubrics_faculty ON grading_rubrics(faculty_id);
CREATE INDEX idx_rubrics_college ON grading_rubrics(college_id);
CREATE INDEX idx_rubrics_course ON grading_rubrics(course_id);
CREATE INDEX idx_rubric_criteria_rubric ON rubric_criteria(rubric_id);
CREATE INDEX idx_rubric_levels_rubric ON rubric_performance_levels(rubric_id);

-- Triggers
CREATE TRIGGER update_rubrics_updated_at 
    BEFORE UPDATE ON grading_rubrics 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_criteria_updated_at 
    BEFORE UPDATE ON rubric_criteria 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_levels_updated_at 
    BEFORE UPDATE ON rubric_performance_levels 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();
