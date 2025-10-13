-- Drop triggers
DROP TRIGGER IF EXISTS trigger_update_files_updated_at ON file_versions;
DROP TRIGGER IF EXISTS trigger_update_folders_updated_at ON folders;

-- Drop functions
DROP FUNCTION IF EXISTS update_files_updated_at();
DROP FUNCTION IF EXISTS update_folders_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_files_college_id;
DROP INDEX IF EXISTS idx_files_folder_id;
DROP INDEX IF EXISTS idx_files_uploaded_by;
DROP INDEX IF EXISTS idx_files_category;
DROP INDEX IF EXISTS idx_files_tags;
DROP INDEX IF EXISTS idx_files_created_at;

DROP INDEX IF EXISTS idx_file_versions_file_id;
DROP INDEX IF EXISTS idx_file_versions_is_current;
DROP INDEX IF EXISTS idx_file_versions_created_at;

DROP INDEX IF EXISTS idx_folders_college_id;
DROP INDEX IF EXISTS idx_folders_parent_id;
DROP INDEX IF EXISTS idx_folders_path;
DROP INDEX IF EXISTS idx_folders_created_by;

-- Drop constraints
ALTER TABLE files DROP CONSTRAINT IF EXISTS fk_files_current_version;
ALTER TABLE file_versions DROP CONSTRAINT IF EXISTS unique_file_version;
ALTER TABLE folders DROP CONSTRAINT IF EXISTS unique_folder_path;

-- Drop tables
DROP TABLE IF EXISTS file_versions;
DROP TABLE IF EXISTS files;
DROP TABLE IF EXISTS folders;