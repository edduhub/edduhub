BEGIN;

ALTER TABLE timetable_blocks
    ALTER COLUMN faculty_id TYPE INTEGER
    USING NULLIF(faculty_id, '')::INTEGER;

ALTER TABLE timetable_blocks
    ALTER COLUMN day_of_week TYPE VARCHAR(50)
    USING day_of_week::TEXT;

COMMIT;
