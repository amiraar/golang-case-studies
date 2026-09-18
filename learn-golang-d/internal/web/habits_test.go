package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"

	"learn-golang-d/internal/store"
)

func TestHabitsHandler_CreateAndCheckin(t *testing.T) {
	e := testEcho()
	h := NewHabitsHandler(store.NewHabitStore())

	req := httptest.NewRequest(http.MethodPost, "/habits", strings.NewReader(`{"name":"Olahraga"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.Create(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201, body=%s", rec.Code, rec.Body.String())
	}

	req2 := httptest.NewRequest(http.MethodPost, "/habits/1/checkin", nil)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	c2.SetParamNames("id")
	c2.SetParamValues("1")
	if err := h.Checkin(c2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec2.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", rec2.Code, rec2.Body.String())
	}
	if !strings.Contains(rec2.Body.String(), `"done":true`) {
		t.Errorf("expected done:true in response, got %s", rec2.Body.String())
	}
}

func TestHabitsHandler_Checkin_UnknownID(t *testing.T) {
	e := testEcho()
	h := NewHabitsHandler(store.NewHabitStore())

	req := httptest.NewRequest(http.MethodPost, "/habits/999/checkin", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("999")

	err := h.Checkin(c)
	if err == nil {
		t.Fatal("expected not-found error")
	}
	e.HTTPErrorHandler(err, c)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}
