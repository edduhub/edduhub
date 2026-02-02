-- Migration: Create audit logs table
-- This enables comprehensive audit trail

CREATE TABLE IF NOT EXISTS audit_logs (
    id SERIAL PRIMARY KEY,
    college_id INTEGER REFERENCES colleges(id) ON DELETE SET NULL,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    user_name VARCHAR(100),
    user_role VARCHAR(50),
    action VARCHAR(50) NOT NULL CHECK (action IN ('CREATE', 'READ', 'UPDATE', 'DELETE', 'LOGIN', 'LOGOUT', 'EXPORT', 'IMPORT', 'APPROVE', 'REJECT')),
    entity_type VARCHAR(50) NOT NULL,
    entity_id VARCHAR(100),
    entity_name VARCHAR(200),
    old_values JSONB,
    new_values JSONB,
    changes_summary TEXT,
    ip_address INET,
    user_agent TEXT,
    request_id VARCHAR(100),
    session_id VARCHAR(100),
    success BOOLEAN DEFAULT TRUE,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS audit_stats (
    id SERIAL PRIMARY KEY,
    college_id INTEGER REFERENCES colleges(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    total_actions INTEGER DEFAULT 0,
    create_count INTEGER DEFAULT 0,
    update_count INTEGER DEFAULT 0,
    delete_count INTEGER DEFAULT 0,
    login_count INTEGER DEFAULT 0,
    failed_count INTEGER DEFAULT 0,
    unique_users INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(college_id, date)
);

-- Indexes
CREATE INDEX idx_audit_college ON audit_logs(college_id);
CREATE INDEX idx_audit_user ON audit_logs(user_id);
CREATE INDEX idx_audit_action ON audit_logs(action);
CREATE INDEX idx_audit_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_created ON audit_logs(created_at);
CREATE INDEX idx_audit_ip ON audit_logs(ip_address);
CREATE INDEX idx_stats_college_date ON audit_stats(college_id, date);

-- Triggers
CREATE TRIGGER update_audit_stats_updated_at 
    BEFORE UPDATE ON audit_stats 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();
