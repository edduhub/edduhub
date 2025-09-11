package attendance

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/skip2/go-qrcode"
)

// studentId in the request body
// course id in the request body and lecture id obtained from qr code
// need to check if student is enrolled in the course
// need to check if the student is enrolled in the lecture
type QRCodeData struct {
	CourseID  int       `json:"course_id"`
	LectureID int       `json:"lecture_id"`
	TimeStamp time.Time `json:"time_stamp"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (a *attendanceService) GenerateQRCode(ctx context.Context, collegeID int, courseID int, lectureID int) (string, error) {
	// generate qr code
	// return qr code
	// need to check if student is enrolled in the course for marking attendance

	now := time.Now()
	expiresAt := now.Add(30 * time.Minute)
	qrCodeData := QRCodeData{
		CourseID:  courseID,
		LectureID: lectureID,
		TimeStamp: now,
		ExpiresAt: expiresAt,
	}
	jsonData, err := json.Marshal(qrCodeData)
	if err != nil {
		return "", err
	}
	qrBytes, err := qrcode.Encode(string(jsonData), qrcode.Medium, 256)
	if err != nil {
		return "", err
	}
	qrbase64 := base64.StdEncoding.EncodeToString(qrBytes)
	return qrbase64, nil
}

// process qr and take values from it to mark attendance(process qr and chaning state)
func (a *attendanceService) ProcessQRCode(ctx context.Context, collegeID int, studentID int, qrCodeContent string) error {
	var qrData QRCodeData
	if err := json.Unmarshal([]byte(qrCodeContent), &qrData); err != nil {
		// Consider logging the actual error here for debugging
		return errors.New("invalid qr code content")
	}

	// Check if the QR code has expired
	if time.Now().After(qrData.ExpiresAt) {
		return errors.New("qr code expired")
	}

	// Attempt to mark attendance
	marked, err := a.MarkAttendance(ctx, collegeID, studentID, qrData.CourseID, qrData.LectureID)
	if err != nil {
		// Return the specific error from MarkAttendance
		return fmt.Errorf("failed to mark attendance: %w", err)
	}
	if !marked {
		return errors.New("unable to mark attendance (check enrollment or if already marked)")
	}

	// Attendance marked successfully
	return nil
}

/// process qr takes qr input and marks attendance
// mark attendance verifies student and marks attendance
