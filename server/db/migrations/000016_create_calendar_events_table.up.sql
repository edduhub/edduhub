CREATE TABLE IF NOT EXISTS calendar_events (
    id SERIAL PRIMARY KEY,
    college_id INTEGER NOT NULL REFERENCES colleges(id) ON DELETE CASCADE,
    course_id INTEGER REFERENCES courses(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    event_type VARCHAR(50) NOT NULL CHECK (event_type IN ('exam', 'holiday', 'meeting', 'deadline', 'event', 'other')),
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    location VARCHAR(255),
    is_recurring BOOLEAN DEFAULT FALSE,
    recurrence_pattern VARCHAR(50),
    created_by VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_time_range CHECK (end_time > start_time)
);

CREATE INDEX idx_calendar_events_college ON calendar_events(college_id);
CREATE INDEX idx_calendar_events_course ON calendar_events(course_id);
CREATE INDEX idx_calendar_events_type ON calendar_events(event_type);
CREATE INDEX idx_calendar_events_time ON calendar_events(start_time, end_time);
