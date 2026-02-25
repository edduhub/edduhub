package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"eduhub/server/internal/repository"
	"eduhub/server/internal/services/auth"
	"eduhub/server/internal/services/facultytools"

	"github.com/labstack/echo/v4"
)

func TestFacultyToolsIntegrationFlow(t *testing.T) {
	ctx, db, pool := setupIntegrationDB(t,
		"users", "colleges", "students", "courses",
		"grading_rubrics", "rubric_criteria",
		"faculty_office_hours", "office_hour_bookings",
	)
	fixture, cleanup := seedIntegrationFixture(t, ctx, pool)
	defer cleanup()

	repo := repository.NewFacultyToolsRepository(db)
	service := facultytools.NewFacultyToolsService(repo)
	handler := NewFacultyToolsHandler(service)
	e := echo.New()

	facultyIdentity := &auth.Identity{ID: fixture.FacultyKratosID}
	facultyIdentity.Traits.Role = "faculty"
	studentIdentity := &auth.Identity{ID: fixture.StudentKratosID}
	studentIdentity.Traits.Role = "student"

	createRubric := []byte(`{"name":"Assignment Rubric","max_score":100,"is_template":false,"is_active":true,"criteria":[{"name":"Clarity","weight":50,"max_score":50,"sort_order":1},{"name":"Depth","weight":50,"max_score":50,"sort_order":2}]}`)
	rubricReq := httptest.NewRequest(http.MethodPost, "/api/faculty-tools/rubrics", bytes.NewReader(createRubric))
	rubricReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rubricRec := httptest.NewRecorder()
	rubricCtx := e.NewContext(rubricReq, rubricRec)
	rubricCtx.Set("college_id", fixture.CollegeID)
	rubricCtx.Set("identity", facultyIdentity)

	if err := handler.CreateRubric(rubricCtx); err != nil {
		t.Fatalf("CreateRubric returned error: %v", err)
	}
	if rubricRec.Code != http.StatusCreated {
		t.Fatalf("expected 201 for rubric create, got %d body=%s", rubricRec.Code, rubricRec.Body.String())
	}

	listRubricsReq := httptest.NewRequest(http.MethodGet, "/api/faculty-tools/rubrics", nil)
	listRubricsRec := httptest.NewRecorder()
	listRubricsCtx := e.NewContext(listRubricsReq, listRubricsRec)
	listRubricsCtx.Set("college_id", fixture.CollegeID)
	listRubricsCtx.Set("identity", facultyIdentity)
	if err := handler.ListRubrics(listRubricsCtx); err != nil {
		t.Fatalf("ListRubrics returned error: %v", err)
	}
	if listRubricsRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for list rubrics, got %d body=%s", listRubricsRec.Code, listRubricsRec.Body.String())
	}

	var rubricListResp successEnvelope
	if err := json.Unmarshal(listRubricsRec.Body.Bytes(), &rubricListResp); err != nil {
		t.Fatalf("failed decoding rubric list response: %v", err)
	}
	var rubricListPayload struct {
		Total int `json:"total"`
	}
	if err := json.Unmarshal(rubricListResp.Data, &rubricListPayload); err != nil {
		t.Fatalf("failed decoding rubric list payload: %v", err)
	}
	if rubricListPayload.Total < 1 {
		t.Fatalf("expected at least one rubric")
	}

	createOfficeHour := []byte(`{"day_of_week":1,"start_time":"09:00","end_time":"10:00","location":"Room 301","is_virtual":false,"max_students":2,"is_active":true}`)
	officeReq := httptest.NewRequest(http.MethodPost, "/api/faculty-tools/office-hours", bytes.NewReader(createOfficeHour))
	officeReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	officeRec := httptest.NewRecorder()
	officeCtx := e.NewContext(officeReq, officeRec)
	officeCtx.Set("college_id", fixture.CollegeID)
	officeCtx.Set("identity", facultyIdentity)

	if err := handler.CreateOfficeHour(officeCtx); err != nil {
		t.Fatalf("CreateOfficeHour returned error: %v", err)
	}
	if officeRec.Code != http.StatusCreated {
		t.Fatalf("expected 201 for office hour create, got %d body=%s", officeRec.Code, officeRec.Body.String())
	}

	var officeResp successEnvelope
	if err := json.Unmarshal(officeRec.Body.Bytes(), &officeResp); err != nil {
		t.Fatalf("failed decoding office-hour response: %v", err)
	}
	var officePayload struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(officeResp.Data, &officePayload); err != nil {
		t.Fatalf("failed decoding office-hour payload: %v", err)
	}
	if officePayload.ID == 0 {
		t.Fatalf("expected created office-hour ID")
	}

	bookingDateOne := time.Now().Add(24 * time.Hour).Format("2006-01-02")
	createBookingOne := fmt.Appendf(nil, `{"office_hour_id":%d,"booking_date":"%s","purpose":"Need project guidance"}`,
		officePayload.ID,
		bookingDateOne,
	)
	bookingOneReq := httptest.NewRequest(http.MethodPost, "/api/faculty-tools/bookings", bytes.NewReader(createBookingOne))
	bookingOneReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	bookingOneRec := httptest.NewRecorder()
	bookingOneCtx := e.NewContext(bookingOneReq, bookingOneRec)
	bookingOneCtx.Set("college_id", fixture.CollegeID)
	bookingOneCtx.Set("student_id", fixture.StudentID)
	bookingOneCtx.Set("identity", studentIdentity)

	if err := handler.CreateBooking(bookingOneCtx); err != nil {
		t.Fatalf("CreateBooking returned error: %v", err)
	}
	if bookingOneRec.Code != http.StatusCreated {
		t.Fatalf("expected 201 for booking create, got %d body=%s", bookingOneRec.Code, bookingOneRec.Body.String())
	}

	var bookingOneResp successEnvelope
	if err := json.Unmarshal(bookingOneRec.Body.Bytes(), &bookingOneResp); err != nil {
		t.Fatalf("failed decoding booking response: %v", err)
	}
	var bookingOnePayload struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(bookingOneResp.Data, &bookingOnePayload); err != nil {
		t.Fatalf("failed decoding booking payload: %v", err)
	}

	updateBookingOne := []byte(`{"status":"completed","notes":"Session done"}`)
	updateBookingReq := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/faculty-tools/bookings/%d/status", bookingOnePayload.ID), bytes.NewReader(updateBookingOne))
	updateBookingReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	updateBookingRec := httptest.NewRecorder()
	updateBookingCtx := e.NewContext(updateBookingReq, updateBookingRec)
	updateBookingCtx.SetPath("/api/faculty-tools/bookings/:bookingID/status")
	updateBookingCtx.SetParamNames("bookingID")
	updateBookingCtx.SetParamValues(fmt.Sprintf("%d", bookingOnePayload.ID))
	updateBookingCtx.Set("college_id", fixture.CollegeID)
	updateBookingCtx.Set("identity", facultyIdentity)
	if err := handler.UpdateBookingStatus(updateBookingCtx); err != nil {
		t.Fatalf("UpdateBookingStatus (faculty) returned error: %v", err)
	}
	if updateBookingRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for booking status update, got %d body=%s", updateBookingRec.Code, updateBookingRec.Body.String())
	}

	bookingDateTwo := time.Now().Add(48 * time.Hour).Format("2006-01-02")
	createBookingTwo := fmt.Appendf(nil, `{"office_hour_id":%d,"booking_date":"%s","purpose":"Need revision help"}`,
		officePayload.ID,
		bookingDateTwo,
	)
	bookingTwoReq := httptest.NewRequest(http.MethodPost, "/api/faculty-tools/bookings", bytes.NewReader(createBookingTwo))
	bookingTwoReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	bookingTwoRec := httptest.NewRecorder()
	bookingTwoCtx := e.NewContext(bookingTwoReq, bookingTwoRec)
	bookingTwoCtx.Set("college_id", fixture.CollegeID)
	bookingTwoCtx.Set("student_id", fixture.StudentID)
	bookingTwoCtx.Set("identity", studentIdentity)
	if err := handler.CreateBooking(bookingTwoCtx); err != nil {
		t.Fatalf("CreateBooking(second) returned error: %v", err)
	}

	var bookingTwoResp successEnvelope
	if err := json.Unmarshal(bookingTwoRec.Body.Bytes(), &bookingTwoResp); err != nil {
		t.Fatalf("failed decoding second booking response: %v", err)
	}
	var bookingTwoPayload struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(bookingTwoResp.Data, &bookingTwoPayload); err != nil {
		t.Fatalf("failed decoding second booking payload: %v", err)
	}

	cancelBookingReq := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/faculty-tools/bookings/%d/status", bookingTwoPayload.ID), bytes.NewReader([]byte(`{"status":"cancelled"}`)))
	cancelBookingReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	cancelBookingRec := httptest.NewRecorder()
	cancelBookingCtx := e.NewContext(cancelBookingReq, cancelBookingRec)
	cancelBookingCtx.SetPath("/api/faculty-tools/bookings/:bookingID/status")
	cancelBookingCtx.SetParamNames("bookingID")
	cancelBookingCtx.SetParamValues(fmt.Sprintf("%d", bookingTwoPayload.ID))
	cancelBookingCtx.Set("college_id", fixture.CollegeID)
	cancelBookingCtx.Set("student_id", fixture.StudentID)
	cancelBookingCtx.Set("identity", studentIdentity)
	if err := handler.UpdateBookingStatus(cancelBookingCtx); err != nil {
		t.Fatalf("UpdateBookingStatus(student cancel) returned error: %v", err)
	}
	if cancelBookingRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for booking cancel, got %d body=%s", cancelBookingRec.Code, cancelBookingRec.Body.String())
	}

	listByOfficeReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/faculty-tools/office-hours/%d/bookings", officePayload.ID), nil)
	listByOfficeRec := httptest.NewRecorder()
	listByOfficeCtx := e.NewContext(listByOfficeReq, listByOfficeRec)
	listByOfficeCtx.SetPath("/api/faculty-tools/office-hours/:officeHourID/bookings")
	listByOfficeCtx.SetParamNames("officeHourID")
	listByOfficeCtx.SetParamValues(fmt.Sprintf("%d", officePayload.ID))
	listByOfficeCtx.Set("college_id", fixture.CollegeID)
	listByOfficeCtx.Set("identity", facultyIdentity)
	if err := handler.ListBookingsByOfficeHour(listByOfficeCtx); err != nil {
		t.Fatalf("ListBookingsByOfficeHour returned error: %v", err)
	}
	if listByOfficeRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for listing office-hour bookings, got %d body=%s", listByOfficeRec.Code, listByOfficeRec.Body.String())
	}
}
