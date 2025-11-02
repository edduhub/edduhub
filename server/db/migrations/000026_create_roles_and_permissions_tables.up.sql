BEGIN;

-- Create roles table
CREATE TABLE IF NOT EXISTS roles (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    college_id INT,
    is_system_role BOOLEAN NOT NULL DEFAULT FALSE, -- System roles cannot be deleted
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_roles_college FOREIGN KEY (college_id) REFERENCES colleges(id) ON DELETE CASCADE
);

-- Create permissions table
CREATE TABLE IF NOT EXISTS permissions (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    resource VARCHAR(100) NOT NULL, -- e.g., 'users', 'courses', 'grades'
    action VARCHAR(50) NOT NULL, -- e.g., 'create', 'read', 'update', 'delete'
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_resource_action UNIQUE(resource, action)
);

-- Create role_permissions junction table
CREATE TABLE IF NOT EXISTS role_permissions (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    role_id INT NOT NULL,
    permission_id INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_role_permissions_role FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    CONSTRAINT fk_role_permissions_permission FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE,
    CONSTRAINT unique_role_permission UNIQUE(role_id, permission_id)
);

-- Create user_role_assignments table (many-to-many: users can have multiple roles)
CREATE TABLE IF NOT EXISTS user_role_assignments (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id INT NOT NULL,
    role_id INT NOT NULL,
    assigned_by INT, -- User ID of who assigned the role
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ, -- Optional expiration date
    CONSTRAINT fk_user_role_assignments_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_user_role_assignments_role FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    CONSTRAINT fk_user_role_assignments_assigned_by FOREIGN KEY (assigned_by) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT unique_user_role UNIQUE(user_id, role_id)
);

-- Create indexes for better query performance
CREATE INDEX idx_roles_college_id ON roles(college_id);
CREATE INDEX idx_roles_name ON roles(name);
CREATE INDEX idx_permissions_resource ON permissions(resource);
CREATE INDEX idx_permissions_action ON permissions(action);
CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions(permission_id);
CREATE INDEX idx_user_role_assignments_user_id ON user_role_assignments(user_id);
CREATE INDEX idx_user_role_assignments_role_id ON user_role_assignments(role_id);

-- Insert default system roles
INSERT INTO roles (name, description, is_system_role) VALUES
('admin', 'System administrator with full access', TRUE),
('faculty', 'Faculty member with teaching privileges', TRUE),
('student', 'Student with learning privileges', TRUE),
('staff', 'Staff member with limited administrative access', TRUE);

-- Insert default permissions
INSERT INTO permissions (name, resource, action, description) VALUES
-- User management
('users.create', 'users', 'create', 'Create new users'),
('users.read', 'users', 'read', 'View user information'),
('users.update', 'users', 'update', 'Update user information'),
('users.delete', 'users', 'delete', 'Delete users'),

-- Course management
('courses.create', 'courses', 'create', 'Create new courses'),
('courses.read', 'courses', 'read', 'View course information'),
('courses.update', 'courses', 'update', 'Update course information'),
('courses.delete', 'courses', 'delete', 'Delete courses'),
('courses.enroll', 'courses', 'enroll', 'Enroll students in courses'),

-- Student management
('students.create', 'students', 'create', 'Create student records'),
('students.read', 'students', 'read', 'View student information'),
('students.update', 'students', 'update', 'Update student information'),
('students.delete', 'students', 'delete', 'Delete student records'),

-- Attendance management
('attendance.mark', 'attendance', 'mark', 'Mark student attendance'),
('attendance.read', 'attendance', 'read', 'View attendance records'),
('attendance.update', 'attendance', 'update', 'Update attendance records'),

-- Grade management
('grades.create', 'grades', 'create', 'Create grade entries'),
('grades.read', 'grades', 'read', 'View grades'),
('grades.update', 'grades', 'update', 'Update grades'),
('grades.delete', 'grades', 'delete', 'Delete grade entries'),

-- Assignment management
('assignments.create', 'assignments', 'create', 'Create assignments'),
('assignments.read', 'assignments', 'read', 'View assignments'),
('assignments.update', 'assignments', 'update', 'Update assignments'),
('assignments.delete', 'assignments', 'delete', 'Delete assignments'),
('assignments.submit', 'assignments', 'submit', 'Submit assignments'),
('assignments.grade', 'assignments', 'grade', 'Grade assignment submissions'),

-- Quiz management
('quizzes.create', 'quizzes', 'create', 'Create quizzes'),
('quizzes.read', 'quizzes', 'read', 'View quizzes'),
('quizzes.update', 'quizzes', 'update', 'Update quizzes'),
('quizzes.delete', 'quizzes', 'delete', 'Delete quizzes'),
('quizzes.attempt', 'quizzes', 'attempt', 'Attempt quizzes'),

-- Announcement management
('announcements.create', 'announcements', 'create', 'Create announcements'),
('announcements.read', 'announcements', 'read', 'View announcements'),
('announcements.update', 'announcements', 'update', 'Update announcements'),
('announcements.delete', 'announcements', 'delete', 'Delete announcements'),

-- Department management
('departments.create', 'departments', 'create', 'Create departments'),
('departments.read', 'departments', 'read', 'View departments'),
('departments.update', 'departments', 'update', 'Update departments'),
('departments.delete', 'departments', 'delete', 'Delete departments'),

-- Fee management
('fees.create', 'fees', 'create', 'Create fee records'),
('fees.read', 'fees', 'read', 'View fee information'),
('fees.update', 'fees', 'update', 'Update fee records'),
('fees.pay', 'fees', 'pay', 'Make fee payments'),

-- Timetable management
('timetable.create', 'timetable', 'create', 'Create timetable entries'),
('timetable.read', 'timetable', 'read', 'View timetable'),
('timetable.update', 'timetable', 'update', 'Update timetable entries'),
('timetable.delete', 'timetable', 'delete', 'Delete timetable entries'),

-- Role management
('roles.create', 'roles', 'create', 'Create roles'),
('roles.read', 'roles', 'read', 'View roles'),
('roles.update', 'roles', 'update', 'Update roles'),
('roles.delete', 'roles', 'delete', 'Delete roles'),
('roles.assign', 'roles', 'assign', 'Assign roles to users'),

-- Permission management
('permissions.read', 'permissions', 'read', 'View permissions'),
('permissions.assign', 'permissions', 'assign', 'Assign permissions to roles'),

-- Analytics
('analytics.read', 'analytics', 'read', 'View analytics and reports');

-- Assign permissions to admin role (full access)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'admin';

-- Assign permissions to faculty role
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.name IN (
    'courses.read', 'courses.update',
    'students.read',
    'attendance.mark', 'attendance.read', 'attendance.update',
    'grades.create', 'grades.read', 'grades.update',
    'assignments.create', 'assignments.read', 'assignments.update', 'assignments.delete', 'assignments.grade',
    'quizzes.create', 'quizzes.read', 'quizzes.update', 'quizzes.delete',
    'announcements.create', 'announcements.read', 'announcements.update', 'announcements.delete',
    'timetable.read',
    'analytics.read'
)
WHERE r.name = 'faculty';

-- Assign permissions to student role
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.name IN (
    'courses.read',
    'students.read',
    'attendance.read',
    'grades.read',
    'assignments.read', 'assignments.submit',
    'quizzes.read', 'quizzes.attempt',
    'announcements.read',
    'fees.read', 'fees.pay',
    'timetable.read'
)
WHERE r.name = 'student';

-- Assign permissions to staff role
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.name IN (
    'courses.read',
    'students.read', 'students.create', 'students.update',
    'announcements.read',
    'departments.read',
    'fees.read', 'fees.create', 'fees.update',
    'timetable.read'
)
WHERE r.name = 'staff';

COMMIT;
