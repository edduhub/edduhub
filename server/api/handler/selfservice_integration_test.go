package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"eduhub/server/internal/repository"
	"eduhub/server/internal/services/auth"
	"eduhub/server/internal/services/selfservice"

	"github.com/labstack/echo/v4"
)

type successEnvelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
}

func TestSelfServiceHandlerIntegrationFlow(t *testing.T) {
	ctx, db, pool := setupIntegrationDB(t,
		"users", "colleges", "students", "self_service_requests",
	)
	fixture, cleanup := seedIntegrationFixture(t, ctx, pool)
	defer cleanup()

	repo := repository.NewSelfServiceRepository(db)
	service := selfservice.NewSelfServiceService(repo)
	handler := NewSelfServiceHandler(service)
	e := echo.New()

	createBody := []byte(`{"type":"document","title":"Transcript Request","description":"Need transcript for scholarship","document_type":"transcript","delivery_method":"email"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/api/self-service/requests", bytes.NewReader(createBody))
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createRec := httptest.NewRecorder()
	createCtx := e.NewContext(createReq, createRec)
	createCtx.Set("student_id", fixture.StudentID)
	createCtx.Set("college_id", fixture.CollegeID)

	if err := handler.CreateRequest(createCtx); err != nil {
		t.Fatalf("CreateRequest returned error: %v", err)
	}
	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d body=%s", createRec.Code, createRec.Body.String())
	}

	var createResp successEnvelope
	if err := json.Unmarshal(createRec.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("failed to decode create response: %v", err)
	}
	var created struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(createResp.Data, &created); err != nil {
		t.Fatalf("failed to decode created request: %v", err)
	}
	if created.ID == 0 {
		t.Fatalf("expected created request ID")
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/self-service/requests", nil)
	listRec := httptest.NewRecorder()
	listCtx := e.NewContext(listReq, listRec)
	listCtx.Set("student_id", fixture.StudentID)
	listCtx.Set("college_id", fixture.CollegeID)
	if err := handler.GetMyRequests(listCtx); err != nil {
		t.Fatalf("GetMyRequests returned error: %v", err)
	}
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", listRec.Code, listRec.Body.String())
	}

	var listResp successEnvelope
	if err := json.Unmarshal(listRec.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("failed to decode list response: %v", err)
	}
	var listed struct {
		Total int `json:"total"`
	}
	if err := json.Unmarshal(listResp.Data, &listed); err != nil {
		t.Fatalf("failed to decode list payload: %v", err)
	}
	if listed.Total != 1 {
		t.Fatalf("expected 1 request, got %d", listed.Total)
	}

	updateBody := []byte(`{"status":"approved","response":"Processed successfully"}`)
	updateReq := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/self-service/requests/%d", created.ID), bytes.NewReader(updateBody))
	updateReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	updateRec := httptest.NewRecorder()
	updateCtx := e.NewContext(updateReq, updateRec)
	updateCtx.SetPath("/api/self-service/requests/:requestID")
	updateCtx.SetParamNames("requestID")
	updateCtx.SetParamValues(fmt.Sprintf("%d", created.ID))
	updateCtx.Set("college_id", fixture.CollegeID)
	identity := &auth.Identity{ID: fixture.AdminKratosID}
	identity.Traits.Role = "admin"
	updateCtx.Set("identity", identity)

	if err := handler.UpdateRequest(updateCtx); err != nil {
		t.Fatalf("UpdateRequest returned error: %v", err)
	}
	if updateRec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", updateRec.Code, updateRec.Body.String())
	}

	var updateResp successEnvelope
	if err := json.Unmarshal(updateRec.Body.Bytes(), &updateResp); err != nil {
		t.Fatalf("failed to decode update response: %v", err)
	}
	var updated struct {
		Status   string `json:"status"`
		Response string `json:"response"`
	}
	if err := json.Unmarshal(updateResp.Data, &updated); err != nil {
		t.Fatalf("failed to decode updated payload: %v", err)
	}
	if updated.Status != "approved" {
		t.Fatalf("expected approved status, got %s", updated.Status)
	}
	if updated.Response == "" {
		t.Fatalf("expected admin response to be persisted")
	}
}
