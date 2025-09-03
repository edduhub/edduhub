BEGIN;

-- Add college_id column to courses table
ALTER TABLE courses
ADD COLUMN IF NOT EXISTS college_id INT NOT NULL;

-- Add foreign key constraint to colleges table
ALTER TABLE courses
ADD CONSTRAINT IF NOT EXISTS fk_courses_college
FOREIGN KEY (college_id)
REFERENCES colleges(id)
ON DELETE RESTRICT;

-- Create index for college_id
CREATE INDEX IF NOT EXISTS idx_courses_college_id ON courses (college_id);

COMMIT;