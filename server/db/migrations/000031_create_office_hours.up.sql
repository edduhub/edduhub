-- Migration: Create faculty office hours table
-- This enables faculty tools office hours feature

CREATE TABLE IF NOT EXISTS faculty_office_hours (
    id SERIAL PRIMARY KEY,
    faculty_id INTEGER NOT NULL,
    college_id INTEGER NOT NULL,
    day_of_week INTEGER NOT NULL CHECK (day_of_week BETWEEN 0 AND 6),
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    location VARCHAR(200),
    is_virtual BOOLEAN DEFAULT FALSE,
    virtual_link VARCHAR(500),
    max_students INTEGER DEFAULT 1,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Ensure no overlapping times for same faculty on same day
    UNIQUE(faculty_id, day_of_week, start_time)
);

CREATE TABLE IF NOT EXISTS office_hour_bookings (
    id SERIAL PRIMARY KEY,
    office_hour_id INTEGER NOT NULL REFERENCES faculty_office_hours(id) ON DELETE CASCADE,
    student_id INTEGER NOT NULL REFERENCES students(student_id) ON DELETE CASCADE,
    booking_date DATE NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    purpose TEXT,
    status VARCHAR(20) DEFAULT 'confirmed' CHECK (status IN ('confirmed', 'cancelled', 'completed', 'no_show')),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Prevent double bookings
    UNIQUE(office_hour_id, booking_date, start_time)
);

-- Indexes
CREATE INDEX idx_office_hours_faculty ON faculty_office_hours(faculty_id);
CREATE INDEX idx_office_hours_college ON faculty_office_hours(college_id);
CREATE INDEX idx_office_hours_day ON faculty_office_hours(day_of_week);
CREATE INDEX idx_bookings_office_hour ON office_hour_bookings(office_hour_id);
CREATE INDEX idx_bookings_student ON office_hour_bookings(student_id);
CREATE INDEX idx_bookings_date ON office_hour_bookings(booking_date);

-- Triggers
CREATE TRIGGER update_office_hours_updated_at 
    BEFORE UPDATE ON faculty_office_hours 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_bookings_updated_at 
    BEFORE UPDATE ON office_hour_bookings 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();
