package models

import (
	"time"
)

type Attendance struct {
	ID        int       `db:"id" json:"ID"`
	StudentID int       `db:"student_id" json:"studentID"`
	CourseID  int       `db:"course_id" json:"courseId"`
	CollegeID int       `db:"college_id" json:"collegeID"`
	Date      time.Time `db:"date" json:"date"`
	Status    string    `db:"status" json:"status"`
	ScannedAt time.Time `db:"scanned_at" json:"scannedAt"`
	LectureID int       `db:"lecture_id" json:"lectureID"`
}

// StudentAttendanceStatus is used for bulk attendance marking requests.
type StudentAttendanceStatus struct {
	StudentID int    `json:"student_id" validate:"required,gt=0"`
	Status    string `json:"status" validate:"required,oneof=Present Absent"` // Ensure status is either Present or Absent
}
