BEGIN;

-- Drop foreign key constraint
ALTER TABLE courses 
DROP CONSTRAINT IF EXISTS fk_courses_college;

-- Drop index
DROP INDEX IF EXISTS idx_courses_college_id;

-- Drop college_id column
ALTER TABLE courses 
DROP COLUMN IF EXISTS college_id;

COMMIT;