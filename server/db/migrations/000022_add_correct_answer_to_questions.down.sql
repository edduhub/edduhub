-- Remove correct_answer field from questions table
ALTER TABLE questions DROP COLUMN IF EXISTS correct_answer;
