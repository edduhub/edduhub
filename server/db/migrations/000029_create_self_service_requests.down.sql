-- Rollback: Drop student self-service requests table

DROP TRIGGER IF EXISTS update_selfservice_updated_at ON self_service_requests;
DROP TABLE IF EXISTS self_service_requests;
