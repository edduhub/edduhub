BEGIN;

DROP INDEX IF EXISTS idx_colleges_external_id_unique;

ALTER TABLE colleges
    DROP COLUMN IF EXISTS external_id;

COMMIT;
