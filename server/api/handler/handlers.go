package handler

import (
	"eduhub/server/internal/services"
)

type Handlers struct {
	Auth         *AuthHandler
	Attendance   *AttendanceHandler
	Announcement *AnnouncementHandler
	// User       *UserHandler
	Student *StudentHandler
	College *CollegeHandler
	Course  *CourseHandler
	Lecture *LectureHandler
	Quiz    *QuizHandler
	// other handlers
}

func NewHandlers(services *services.Services) *Handlers {
	return &Handlers{
		Auth:         NewAuthHandler(services.Auth),
		Attendance:   NewAttendanceHandler(services.Attendance),
		Announcement: NewAnnouncementHandler(services.Announcement),
		// User:       &UserHandler{authService: services.Auth},
		Student: NewStudentHandler(services.StudentService),
		College: NewCollegeHandler(services.CollegeService),
		Course:  NewCourseHandler(services.CourseService),
		Lecture: NewLectureHandler(services.LectureService),
		Quiz:    NewQuizHandler(services.QuizService),
	}
}