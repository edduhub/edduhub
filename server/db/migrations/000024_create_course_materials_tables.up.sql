-- Create course_modules table
CREATE TABLE course_modules (
    id SERIAL PRIMARY KEY,
    course_id INTEGER NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    display_order INTEGER NOT NULL DEFAULT 0,
    is_published BOOLEAN NOT NULL DEFAULT false,
    college_id INTEGER NOT NULL REFERENCES colleges(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create course_materials table
CREATE TABLE course_materials (
    id SERIAL PRIMARY KEY,
    course_id INTEGER NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL, -- document, video, link, assignment, quiz, etc.
    file_id INTEGER REFERENCES files(id) ON DELETE SET NULL,
    external_url TEXT,
    module_id INTEGER REFERENCES course_modules(id) ON DELETE SET NULL,
    display_order INTEGER NOT NULL DEFAULT 0,
    is_published BOOLEAN NOT NULL DEFAULT false,
    published_at TIMESTAMP WITH TIME ZONE,
    due_date TIMESTAMP WITH TIME ZONE,
    uploaded_by INTEGER NOT NULL REFERENCES users(id),
    college_id INTEGER NOT NULL REFERENCES colleges(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create course_material_access table for tracking student access
CREATE TABLE course_material_access (
    id SERIAL PRIMARY KEY,
    material_id INTEGER NOT NULL REFERENCES course_materials(id) ON DELETE CASCADE,
    student_id INTEGER NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    accessed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    duration_seconds INTEGER DEFAULT 0,
    completed BOOLEAN DEFAULT false
);

-- Create indexes for better performance
CREATE INDEX idx_course_modules_course_id ON course_modules(course_id);
CREATE INDEX idx_course_modules_college_id ON course_modules(college_id);
CREATE INDEX idx_course_modules_order ON course_modules(display_order);
CREATE INDEX idx_course_modules_published ON course_modules(is_published);

CREATE INDEX idx_course_materials_course_id ON course_materials(course_id);
CREATE INDEX idx_course_materials_module_id ON course_materials(module_id);
CREATE INDEX idx_course_materials_college_id ON course_materials(college_id);
CREATE INDEX idx_course_materials_type ON course_materials(type);
CREATE INDEX idx_course_materials_order ON course_materials(display_order);
CREATE INDEX idx_course_materials_published ON course_materials(is_published);
CREATE INDEX idx_course_materials_due_date ON course_materials(due_date);
CREATE INDEX idx_course_materials_uploaded_by ON course_materials(uploaded_by);
CREATE INDEX idx_course_materials_file_id ON course_materials(file_id);

CREATE INDEX idx_material_access_material_id ON course_material_access(material_id);
CREATE INDEX idx_material_access_student_id ON course_material_access(student_id);
CREATE INDEX idx_material_access_accessed_at ON course_material_access(accessed_at);
CREATE INDEX idx_material_access_completed ON course_material_access(material_id, student_id, completed);

-- Create trigger to update course_modules.updated_at
CREATE OR REPLACE FUNCTION update_course_modules_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_course_modules_updated_at
    BEFORE UPDATE ON course_modules
    FOR EACH ROW
    EXECUTE FUNCTION update_course_modules_updated_at();

-- Create trigger to update course_materials.updated_at
CREATE OR REPLACE FUNCTION update_course_materials_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_course_materials_updated_at
    BEFORE UPDATE ON course_materials
    FOR EACH ROW
    EXECUTE FUNCTION update_course_materials_updated_at();

-- Add constraints
ALTER TABLE course_modules ADD CONSTRAINT unique_module_order_per_course
    UNIQUE (course_id, display_order);

-- Add check constraint for material type
ALTER TABLE course_materials ADD CONSTRAINT check_material_type
    CHECK (type IN ('document', 'video', 'link', 'assignment', 'quiz', 'presentation', 'audio', 'image', 'other'));

-- Add check constraint to ensure either file_id or external_url is provided for certain types
ALTER TABLE course_materials ADD CONSTRAINT check_material_content
    CHECK (
        (type IN ('link') AND external_url IS NOT NULL) OR
        (type IN ('document', 'video', 'presentation', 'audio', 'image') AND (file_id IS NOT NULL OR external_url IS NOT NULL)) OR
        (type IN ('assignment', 'quiz', 'other'))
    );
