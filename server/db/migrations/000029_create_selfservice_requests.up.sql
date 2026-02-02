-- Migration: Create self-service requests table
-- This enables the student self-service portal feature

CREATE TABLE IF NOT EXISTS self_service_requests (
    id SERIAL PRIMARY KEY,
    student_id INTEGER NOT NULL REFERENCES students(student_id) ON DELETE CASCADE,
    college_id INTEGER NOT NULL,
    request_type VARCHAR(30) NOT NULL CHECK (request_type IN ('enrollment', 'schedule', 'transcript', 'document', 'other')),
    title VARCHAR(200) NOT NULL,
    description TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'processing', 'completed')),
    
    -- For document requests
    document_type VARCHAR(50),
    delivery_method VARCHAR(20) CHECK (delivery_method IN ('pickup', 'email', 'postal')),
    
    -- For enrollment requests
    requested_course_id INTEGER REFERENCES courses(course_id),
    
    -- For schedule requests
    current_schedule TEXT,
    requested_schedule TEXT,
    
    -- Admin response
    admin_response TEXT,
    responded_by INTEGER,
    responded_at TIMESTAMP,
    
    -- Tracking
    submitted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_selfservice_student ON self_service_requests(student_id);
CREATE INDEX idx_selfservice_college ON self_service_requests(college_id);
CREATE INDEX idx_selfservice_status ON self_service_requests(status);
CREATE INDEX idx_selfservice_type ON self_service_requests(request_type);
CREATE INDEX idx_selfservice_submitted ON self_service_requests(submitted_at DESC);

-- Trigger for updated_at
CREATE TRIGGER update_selfservice_requests_updated_at 
    BEFORE UPDATE ON self_service_requests 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();
