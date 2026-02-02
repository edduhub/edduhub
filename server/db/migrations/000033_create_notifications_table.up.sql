-- Migration: Create notifications table
-- This enables notification system

CREATE TABLE IF NOT EXISTS notifications (
    id SERIAL PRIMARY KEY,
    college_id INTEGER NOT NULL REFERENCES colleges(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    type VARCHAR(20) DEFAULT 'info' CHECK (type IN ('info', 'warning', 'success', 'error', 'urgent')),
    target_audience VARCHAR(50)[] DEFAULT '{}',
    sender_id INTEGER REFERENCES users(id),
    sender_name VARCHAR(100),
    related_entity_type VARCHAR(50),
    related_entity_id INTEGER,
    is_broadcast BOOLEAN DEFAULT FALSE,
    is_published BOOLEAN DEFAULT FALSE,
    published_at TIMESTAMP,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Notification recipients table
CREATE TABLE IF NOT EXISTS notification_recipients (
    id SERIAL PRIMARY KEY,
    notification_id INTEGER NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_read BOOLEAN DEFAULT FALSE,
    read_at TIMESTAMP,
    is_email_sent BOOLEAN DEFAULT FALSE,
    email_sent_at TIMESTAMP,
    is_push_sent BOOLEAN DEFAULT FALSE,
    push_sent_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(notification_id, user_id)
);

-- User notification preferences
CREATE TABLE IF NOT EXISTS notification_preferences (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    email_enabled BOOLEAN DEFAULT TRUE,
    push_enabled BOOLEAN DEFAULT TRUE,
    sms_enabled BOOLEAN DEFAULT FALSE,
    urgent_email BOOLEAN DEFAULT TRUE,
    urgent_push BOOLEAN DEFAULT TRUE,
    info_email BOOLEAN DEFAULT TRUE,
    info_push BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(user_id)
);

-- Indexes
CREATE INDEX idx_notifications_college ON notifications(college_id);
CREATE INDEX idx_notifications_type ON notifications(type);
CREATE INDEX idx_notifications_published ON notifications(is_published, published_at);
CREATE INDEX idx_notifications_expires ON notifications(expires_at);
CREATE INDEX idx_recipients_user ON notification_recipients(user_id);
CREATE INDEX idx_recipients_notification ON notification_recipients(notification_id);
CREATE INDEX idx_recipients_read ON notification_recipients(is_read);
CREATE INDEX idx_notif_prefs_user ON notification_preferences(user_id);

-- Triggers
CREATE TRIGGER update_notifications_updated_at 
    BEFORE UPDATE ON notifications 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_notif_prefs_updated_at 
    BEFORE UPDATE ON notification_preferences 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();
