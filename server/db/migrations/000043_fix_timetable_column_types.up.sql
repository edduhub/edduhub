BEGIN;

ALTER TABLE timetable_blocks
    ALTER COLUMN day_of_week TYPE INTEGER
    USING CASE
        WHEN day_of_week ~ '^[0-9]+$' THEN day_of_week::INTEGER
        WHEN LOWER(day_of_week) = 'monday' THEN 1
        WHEN LOWER(day_of_week) = 'tuesday' THEN 2
        WHEN LOWER(day_of_week) = 'wednesday' THEN 3
        WHEN LOWER(day_of_week) = 'thursday' THEN 4
        WHEN LOWER(day_of_week) = 'friday' THEN 5
        WHEN LOWER(day_of_week) = 'saturday' THEN 6
        WHEN LOWER(day_of_week) = 'sunday' THEN 0
        ELSE 0
    END;

ALTER TABLE timetable_blocks
    ALTER COLUMN faculty_id TYPE VARCHAR(255)
    USING faculty_id::TEXT;

COMMIT;
