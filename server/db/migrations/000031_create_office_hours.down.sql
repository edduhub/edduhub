-- Rollback: Drop faculty office hours tables

DROP TRIGGER IF EXISTS update_bookings_updated_at ON office_hour_bookings;
DROP TRIGGER IF EXISTS update_office_hours_updated_at ON faculty_office_hours;

DROP TABLE IF EXISTS office_hour_bookings;
DROP TABLE IF EXISTS faculty_office_hours;
