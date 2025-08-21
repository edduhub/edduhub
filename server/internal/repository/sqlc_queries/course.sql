-- name: CreateCourse :one
INSERT INTO courses (
    name,
    description,
    credits,
    instructor_id,
    college_id,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING id, name, description, credits, instructor_id, college_id, created_at, updated_at;

-- name: FindCourseByID :one
SELECT id, name, description, credits, instructor_id, college_id, created_at, updated_at
FROM courses
WHERE id = $1 AND college_id = $2;

-- name: UpdateCourse :exec
UPDATE courses
SET name = $1,
    description = $2,
    credits = $3,
    instructor_id = $4,
    updated_at = $5
WHERE id = $6 AND college_id = $7;

-- name: DeleteCourse :exec
DELETE FROM courses
WHERE id = $1 AND college_id = $2;

-- name: FindAllCourses :many
SELECT id, name, description, credits, instructor_id, college_id, created_at, updated_at
FROM courses
WHERE college_id = $1
ORDER BY name ASC
LIMIT $2 OFFSET $3;

-- name: FindCoursesByInstructor :many
SELECT id, name, description, credits, instructor_id, college_id, created_at, updated_at
FROM courses
WHERE college_id = $1 AND instructor_id = $2
ORDER BY name ASC
LIMIT $3 OFFSET $4;

-- name: CountCoursesByCollege :one
SELECT COUNT(*) as count
FROM courses
WHERE college_id = $1;

-- name: CountCoursesByInstructor :one
SELECT COUNT(*) as count
FROM courses
WHERE college_id = $1 AND instructor_id = $2;