-- Migration: Create student self-service requests table
-- Canonical schema for self-service workflows

CREATE TABLE IF NOT EXISTS self_service_requests (
    id SERIAL PRIMARY KEY,
    student_id INTEGER NOT NULL REFERENCES students(student_id) ON DELETE CASCADE,
    college_id INTEGER NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('enrollment', 'schedule', 'transcript', 'document')),
    title VARCHAR(200) NOT NULL,
    description TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'processing')),
    document_type VARCHAR(50),
    delivery_method VARCHAR(20),
    admin_response TEXT,
    responded_by INTEGER,
    responded_at TIMESTAMP,
    submitted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_selfservice_student ON self_service_requests(student_id);
CREATE INDEX IF NOT EXISTS idx_selfservice_college ON self_service_requests(college_id);
CREATE INDEX IF NOT EXISTS idx_selfservice_status ON self_service_requests(status);
CREATE INDEX IF NOT EXISTS idx_selfservice_type ON self_service_requests(type);
CREATE INDEX IF NOT EXISTS idx_selfservice_submitted ON self_service_requests(submitted_at DESC);

DROP TRIGGER IF EXISTS update_selfservice_updated_at ON self_service_requests;
CREATE TRIGGER update_selfservice_updated_at
    BEFORE UPDATE ON self_service_requests
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
