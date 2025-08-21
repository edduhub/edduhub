-- name: CreateStudent :one
INSERT INTO students (
    user_id,
    college_id,
    kratos_identity_id,
    enrollment_year,
    roll_no,
    is_active,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING student_id, user_id, college_id, kratos_identity_id, enrollment_year, roll_no, is_active, created_at, updated_at;

-- name: GetStudentByRollNo :one
SELECT student_id, user_id, college_id, kratos_identity_id, enrollment_year, roll_no, is_active, created_at, updated_at
FROM students
WHERE roll_no = $1 AND college_id = $2;

-- name: GetStudentByID :one
SELECT student_id, user_id, college_id, kratos_identity_id, enrollment_year, roll_no, is_active, created_at, updated_at
FROM students
WHERE student_id = $1 AND college_id = $2;

-- name: UpdateStudent :exec
UPDATE students
SET user_id = $1,
    college_id = $2,
    kratos_identity_id = $3,
    enrollment_year = $4,
    roll_no = $5,
    is_active = $6,
    updated_at = $7
WHERE student_id = $8;

-- name: FreezeStudent :exec
UPDATE students
SET is_active = false,
    updated_at = NOW()
WHERE roll_no = $1;

-- name: UnFreezeStudent :exec
UPDATE students
SET is_active = true,
    updated_at = NOW()
WHERE roll_no = $1;

-- name: FindByKratosID :one
SELECT student_id, user_id, college_id, kratos_identity_id, enrollment_year, roll_no, is_active, created_at, updated_at
FROM students
WHERE kratos_identity_id = $1;

-- name: DeleteStudent :exec
DELETE FROM students
WHERE student_id = $1 AND college_id = $2;

-- name: FindAllStudentsByCollege :many
SELECT student_id, user_id, college_id, kratos_identity_id, enrollment_year, roll_no, is_active, created_at, updated_at
FROM students
WHERE college_id = $1
ORDER BY roll_no ASC
LIMIT $2 OFFSET $3;

-- name: CountStudentsByCollege :one
SELECT COUNT(*) as count
FROM students
WHERE college_id = $1;