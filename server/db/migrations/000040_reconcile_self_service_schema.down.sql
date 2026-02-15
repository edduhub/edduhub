-- Reconciliation rollback intentionally no-op.
-- This migration normalizes divergent schemas and is not safely reversible.
SELECT 1;
