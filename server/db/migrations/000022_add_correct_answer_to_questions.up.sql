-- Add correct_answer field to questions table for short answer grading
ALTER TABLE questions ADD COLUMN IF NOT EXISTS correct_answer TEXT;

-- Add comment explaining the field usage
COMMENT ON COLUMN questions.correct_answer IS 'Stores the correct answer for short answer questions. Can contain multiple acceptable answers separated by semicolons for flexible matching.';
