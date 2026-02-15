-- Reconciliation migration: normalize self_service_requests schema to canonical form
-- Safe for databases that previously used request_type-based schema.

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = 'public'
          AND table_name = 'self_service_requests'
    ) THEN
        RETURN;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'self_service_requests'
          AND column_name = 'request_type'
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'self_service_requests'
          AND column_name = 'type'
    ) THEN
        ALTER TABLE self_service_requests ADD COLUMN type VARCHAR(50);
        UPDATE self_service_requests
        SET type = CASE request_type
            WHEN 'other' THEN 'document'
            ELSE request_type
        END;
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'self_service_requests'
          AND column_name = 'submitted_at'
    ) THEN
        ALTER TABLE self_service_requests ADD COLUMN submitted_at TIMESTAMP;
        UPDATE self_service_requests
        SET submitted_at = COALESCE(created_at, CURRENT_TIMESTAMP)
        WHERE submitted_at IS NULL;
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'self_service_requests'
          AND column_name = 'admin_response'
    ) THEN
        ALTER TABLE self_service_requests ADD COLUMN admin_response TEXT;
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'self_service_requests'
          AND column_name = 'delivery_method'
    ) THEN
        ALTER TABLE self_service_requests ADD COLUMN delivery_method VARCHAR(20);
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'self_service_requests'
          AND column_name = 'document_type'
    ) THEN
        ALTER TABLE self_service_requests ADD COLUMN document_type VARCHAR(50);
    END IF;

    UPDATE self_service_requests
    SET status = 'approved'
    WHERE status = 'completed';

    ALTER TABLE self_service_requests
        ALTER COLUMN submitted_at SET DEFAULT CURRENT_TIMESTAMP,
        ALTER COLUMN submitted_at SET NOT NULL;

    ALTER TABLE self_service_requests
        ALTER COLUMN type SET DEFAULT 'document';
    UPDATE self_service_requests SET type = 'document' WHERE type IS NULL;
    ALTER TABLE self_service_requests
        ALTER COLUMN type SET NOT NULL,
        ALTER COLUMN status SET DEFAULT 'pending',
        ALTER COLUMN status SET NOT NULL;

    ALTER TABLE self_service_requests DROP COLUMN IF EXISTS request_type;
    ALTER TABLE self_service_requests DROP COLUMN IF EXISTS requested_course_id;
    ALTER TABLE self_service_requests DROP COLUMN IF EXISTS current_schedule;
    ALTER TABLE self_service_requests DROP COLUMN IF EXISTS requested_schedule;
    ALTER TABLE self_service_requests DROP COLUMN IF EXISTS completed_at;

    BEGIN
        ALTER TABLE self_service_requests
            ADD CONSTRAINT self_service_requests_type_check_v2
            CHECK (type IN ('enrollment', 'schedule', 'transcript', 'document'));
    EXCEPTION
        WHEN duplicate_object THEN NULL;
    END;

    BEGIN
        ALTER TABLE self_service_requests
            ADD CONSTRAINT self_service_requests_status_check_v2
            CHECK (status IN ('pending', 'approved', 'rejected', 'processing'));
    EXCEPTION
        WHEN duplicate_object THEN NULL;
    END;

    CREATE INDEX IF NOT EXISTS idx_selfservice_submitted ON self_service_requests(submitted_at DESC);
END $$;
