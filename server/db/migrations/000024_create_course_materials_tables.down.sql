-- Drop triggers
DROP TRIGGER IF EXISTS trigger_update_course_modules_updated_at ON course_modules;
DROP TRIGGER IF EXISTS trigger_update_course_materials_updated_at ON course_materials;

-- Drop functions
DROP FUNCTION IF EXISTS update_course_modules_updated_at();
DROP FUNCTION IF EXISTS update_course_materials_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_course_modules_course_id;
DROP INDEX IF EXISTS idx_course_modules_college_id;
DROP INDEX IF EXISTS idx_course_modules_order;
DROP INDEX IF EXISTS idx_course_modules_published;

DROP INDEX IF EXISTS idx_course_materials_course_id;
DROP INDEX IF EXISTS idx_course_materials_module_id;
DROP INDEX IF EXISTS idx_course_materials_college_id;
DROP INDEX IF EXISTS idx_course_materials_type;
DROP INDEX IF EXISTS idx_course_materials_order;
DROP INDEX IF EXISTS idx_course_materials_published;
DROP INDEX IF EXISTS idx_course_materials_due_date;
DROP INDEX IF EXISTS idx_course_materials_uploaded_by;
DROP INDEX IF EXISTS idx_course_materials_file_id;

DROP INDEX IF EXISTS idx_material_access_material_id;
DROP INDEX IF EXISTS idx_material_access_student_id;
DROP INDEX IF EXISTS idx_material_access_accessed_at;
DROP INDEX IF EXISTS idx_material_access_completed;

-- Drop constraints
ALTER TABLE course_modules DROP CONSTRAINT IF EXISTS unique_module_order_per_course;
ALTER TABLE course_materials DROP CONSTRAINT IF EXISTS check_material_type;
ALTER TABLE course_materials DROP CONSTRAINT IF EXISTS check_material_content;

-- Drop tables (in reverse order of dependencies)
DROP TABLE IF EXISTS course_material_access;
DROP TABLE IF EXISTS course_materials;
DROP TABLE IF EXISTS course_modules;
