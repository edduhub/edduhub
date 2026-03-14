BEGIN;

ALTER TABLE colleges
    ADD COLUMN IF NOT EXISTS external_id VARCHAR(255);

CREATE UNIQUE INDEX IF NOT EXISTS idx_colleges_external_id_unique
    ON colleges(external_id)
    WHERE external_id IS NOT NULL;

COMMIT;
