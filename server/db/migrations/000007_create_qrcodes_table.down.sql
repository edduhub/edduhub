BEGIN;

DROP INDEX IF EXISTS idx_qrcodes_student_id;
DROP INDEX IF EXISTS idx_qrcodes_expires_at;
DROP TABLE IF EXISTS qrcodes;

COMMIT;
