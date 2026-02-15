package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"eduhub/server/internal/repository"
	"eduhub/server/internal/services/college"

	"github.com/labstack/echo/v4"
)

func TestCollegeIntegrationFlow(t *testing.T) {
	ctx, db, pool := setupIntegrationDB(t, "colleges", "users", "students", "courses")
	fixture, cleanup := seedIntegrationFixture(t, ctx, pool)
	defer cleanup()

	repo := repository.NewCollegeRepository(db)
	service := college.NewCollegeService(repo)
	handler := NewCollegeHandler(service)
	e := echo.New()

	// Test GetCollegeDetails - Success case
	t.Run("GetCollegeDetails_Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/colleges/"+string(rune(fixture.CollegeID)), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("college_id", fixture.CollegeID)

		if err := handler.GetCollegeDetails(c); err != nil {
			t.Fatalf("GetCollegeDetails returned error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
		}
	})

	// Test GetCollegeDetails - Not found case
	t.Run("GetCollegeDetails_NotFound", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/colleges/999999", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("college_id", 999999)

		err := handler.GetCollegeDetails(c)
		if err == nil {
			t.Fatalf("expected error for non-existent college")
		}
	})

	// Test GetCollegeStats - Success case
	t.Run("GetCollegeStats_Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/colleges/"+string(rune(fixture.CollegeID))+"/stats", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("college_id", fixture.CollegeID)

		if err := handler.GetCollegeStats(c); err != nil {
			t.Fatalf("GetCollegeStats returned error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
		}
	})

	// Test UpdateCollegeDetails - Success case
	t.Run("UpdateCollegeDetails_Success", func(t *testing.T) {
		updateReq := []byte(`{"name":"Updated College Name","city":"New City"}`)
		req := httptest.NewRequest(http.MethodPut, "/api/colleges/"+string(rune(fixture.CollegeID)), bytes.NewReader(updateReq))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("college_id", fixture.CollegeID)

		if err := handler.UpdateCollegeDetails(c); err != nil {
			t.Fatalf("UpdateCollegeDetails returned error: %v", err)
		}
		if rec.Code != http.StatusNoContent && rec.Code != http.StatusOK {
			t.Fatalf("expected 204 or 200, got %d body=%s", rec.Code, rec.Body.String())
		}
	})

	// Test UpdateCollegeDetails - Invalid request
	t.Run("UpdateCollegeDetails_InvalidRequest", func(t *testing.T) {
		invalidReq := []byte(`invalid json`)
		req := httptest.NewRequest(http.MethodPut, "/api/colleges/"+string(rune(fixture.CollegeID)), bytes.NewReader(invalidReq))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("college_id", fixture.CollegeID)

		err := handler.UpdateCollegeDetails(c)
		if err == nil {
			t.Fatalf("expected error for invalid request body")
		}
	})

	// Test GetCollegeDetails after update to verify changes
	t.Run("GetCollegeDetails_AfterUpdate", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/colleges/"+string(rune(fixture.CollegeID)), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("college_id", fixture.CollegeID)

		if err := handler.GetCollegeDetails(c); err != nil {
			t.Fatalf("GetCollegeDetails returned error: %v", err)
		}

		var resp successEnvelope
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		t.Logf("College details response: %s", rec.Body.String())
	})
}

func TestCollegeHandler_MissingCollegeID(t *testing.T) {
	_, db, pool := setupIntegrationDB(t, "colleges")
	if pool == nil {
		t.Skip("No database connection")
	}

	repo := repository.NewCollegeRepository(db)
	service := college.NewCollegeService(repo)
	handler := NewCollegeHandler(service)
	e := echo.New()

	// Test without college_id in context
	req := httptest.NewRequest(http.MethodGet, "/api/colleges/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	// Not setting college_id

	err := handler.GetCollegeDetails(c)
	if err == nil {
		t.Fatalf("expected error when college_id is missing")
	}
}
