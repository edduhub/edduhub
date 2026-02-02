-- Migration: Create student self-service requests table
-- This enables the self-service portal feature

CREATE TABLE IF NOT EXISTS self_service_requests (
    id SERIAL PRIMARY KEY,
    student_id INTEGER NOT NULL REFERENCES students(student_id) ON DELETE CASCADE,
    college_id INTEGER NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('enrollment', 'schedule', 'transcript', 'document')),
    title VARCHAR(200) NOT NULL,
    description TEXT NOT NULL,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'processing')),
    document_type VARCHAR(50),
    delivery_method VARCHAR(20),
    admin_response TEXT,
    responded_by INTEGER,
    responded_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_selfservice_student ON self_service_requests(student_id);
CREATE INDEX idx_selfservice_college ON self_service_requests(college_id);
CREATE INDEX idx_selfservice_status ON self_service_requests(status);
CREATE INDEX idx_selfservice_type ON self_service_requests(type);

-- Trigger
CREATE TRIGGER update_selfservice_updated_at 
    BEFORE UPDATE ON self_service_requests 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();
