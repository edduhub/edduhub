-- Rollback: Drop faculty grading rubrics tables

DROP TRIGGER IF EXISTS update_levels_updated_at ON rubric_performance_levels;
DROP TRIGGER IF EXISTS update_criteria_updated_at ON rubric_criteria;
DROP TRIGGER IF EXISTS update_rubrics_updated_at ON grading_rubrics;

DROP TABLE IF EXISTS rubric_performance_levels;
DROP TABLE IF EXISTS rubric_criteria;
DROP TABLE IF EXISTS grading_rubrics;
