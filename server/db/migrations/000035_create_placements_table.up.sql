-- Migration: Create placements table
-- This enables placement management feature

CREATE TABLE IF NOT EXISTS placements (
    id SERIAL PRIMARY KEY,
    college_id INTEGER NOT NULL REFERENCES colleges(id) ON DELETE CASCADE,
    company_name VARCHAR(200) NOT NULL,
    company_logo VARCHAR(500),
    job_title VARCHAR(200) NOT NULL,
    job_description TEXT,
    job_type VARCHAR(50) CHECK (job_type IN ('full_time', 'part_time', 'internship', 'contract')),
    location VARCHAR(200),
    is_remote BOOLEAN DEFAULT FALSE,
    salary_range_min DECIMAL(10,2),
    salary_range_max DECIMAL(10,2),
    salary_currency VARCHAR(3) DEFAULT 'USD',
    required_skills TEXT[],
    eligibility_criteria TEXT,
    application_deadline TIMESTAMP,
    drive_date TIMESTAMP,
    interview_mode VARCHAR(50) CHECK (interview_mode IN ('on_campus', 'virtual', 'hybrid')),
    max_applications INTEGER,
    status VARCHAR(20) DEFAULT 'open' CHECK (status IN ('open', 'closed', 'in_progress', 'completed', 'cancelled')),
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS placement_applications (
    id SERIAL PRIMARY KEY,
    placement_id INTEGER NOT NULL REFERENCES placements(id) ON DELETE CASCADE,
    student_id INTEGER NOT NULL REFERENCES students(student_id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'applied' CHECK (status IN ('applied', 'shortlisted', 'interview_scheduled', 'selected', 'rejected', 'withdrawn')),
    resume_url VARCHAR(500),
    cover_letter TEXT,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(placement_id, student_id)
);

CREATE TABLE IF NOT EXISTS placement_interviews (
    id SERIAL PRIMARY KEY,
    application_id INTEGER NOT NULL REFERENCES placement_applications(id) ON DELETE CASCADE,
    round_number INTEGER NOT NULL,
    round_name VARCHAR(100),
    scheduled_at TIMESTAMP,
    duration_minutes INTEGER,
    mode VARCHAR(50) CHECK (mode IN ('virtual', 'in_person', 'phone')),
    meeting_link VARCHAR(500),
    location VARCHAR(200),
    interviewer_name VARCHAR(100),
    interviewer_email VARCHAR(100),
    feedback TEXT,
    result VARCHAR(20) CHECK (result IN ('pending', 'passed', 'failed', 'no_show')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_placements_college ON placements(college_id);
CREATE INDEX idx_placements_status ON placements(status);
CREATE INDEX idx_placements_deadline ON placements(application_deadline);
CREATE INDEX idx_placements_company ON placements(company_name);
CREATE INDEX idx_applications_placement ON placement_applications(placement_id);
CREATE INDEX idx_applications_student ON placement_applications(student_id);
CREATE INDEX idx_applications_status ON placement_applications(status);
CREATE INDEX idx_interviews_application ON placement_interviews(application_id);

-- Triggers
CREATE TRIGGER update_placements_updated_at 
    BEFORE UPDATE ON placements 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_placement_apps_updated_at 
    BEFORE UPDATE ON placement_applications 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_placement_interviews_updated_at 
    BEFORE UPDATE ON placement_interviews 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();
