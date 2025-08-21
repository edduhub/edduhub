-- name: GetAttendanceByCourse :many
SELECT id, student_id, course_id, college_id, date, status, scanned_at, lecture_id
FROM attendance
WHERE college_id = $1 AND course_id = $2
ORDER BY date DESC, student_id ASC
LIMIT $3 OFFSET $4;

-- name: MarkAttendance :one
INSERT INTO attendance (
    student_id,
    course_id,
    college_id,
    lecture_id,
    date,
    status,
    scanned_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) ON CONFLICT (student_id, course_id, lecture_id, date, college_id)
DO UPDATE SET scanned_at = EXCLUDED.scanned_at, status = EXCLUDED.status
RETURNING *;

-- name: UpdateAttendance :exec
UPDATE attendance
SET status = $1
WHERE college_id = $2 AND student_id = $3 AND course_id = $4 AND lecture_id = $5;

-- name: GetAttendanceStudentInCourse :many
SELECT id, student_id, course_id, college_id, date, status, scanned_at, lecture_id
FROM attendance
WHERE college_id = $1 AND student_id = $2 AND course_id = $3
ORDER BY date DESC, scanned_at DESC
LIMIT $4 OFFSET $5;

-- name: GetAttendanceStudent :many
SELECT id, student_id, course_id, college_id, date, status, scanned_at, lecture_id
FROM attendance
WHERE college_id = $1 AND student_id = $2
ORDER BY date DESC, course_id ASC, scanned_at DESC
LIMIT $3 OFFSET $4;

-- name: GetAttendanceByLecture :many
SELECT id, student_id, course_id, college_id, date, status, scanned_at, lecture_id
FROM attendance
WHERE college_id = $1 AND lecture_id = $2 AND course_id = $3
ORDER BY student_id ASC, scanned_at ASC
LIMIT $4 OFFSET $5;

-- name: FreezeAttendance :exec
UPDATE attendance
SET status = $1
WHERE college_id = $2 AND student_id = $3;

-- name: UnFreezeAttendance :exec
UPDATE attendance
SET status = $1
WHERE college_id = $2 AND student_id = $3 AND status = $4;

-- name: SetAttendanceStatus :exec
INSERT INTO attendance (
    student_id,
    course_id,
    college_id,
    lecture_id,
    status,
    scanned_at
) VALUES (
    $1, $2, $3, $4, $5, $6
) ON CONFLICT (student_id, course_id, college_id, lecture_id)
DO UPDATE SET status = EXCLUDED.status, scanned_at = EXCLUDED.scanned_at;