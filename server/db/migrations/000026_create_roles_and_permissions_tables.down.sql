BEGIN;

-- Drop tables in reverse order due to foreign key constraints
DROP TABLE IF EXISTS user_role_assignments;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;

COMMIT;
