package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"

	"learn-golang-d/internal/store"
)

// testEcho: satu instance Echo per test dengan Validator+ErrorHandler
// terpasang, persis seperti main.go - supaya c.Validate() dan alur error
// di test benar-benar menguji wiring produksi, bukan tiruan.
func testEcho() *echo.Echo {
	e := echo.New()
	e.Validator = NewCustomValidator()
	e.HTTPErrorHandler = ErrorHandler
	return e
}

func TestNotesHandler_Create(t *testing.T) {
	e := testEcho()
	h := NewNotesHandler(store.NewNoteStore())

	req := httptest.NewRequest(http.MethodPost, "/notes", strings.NewReader(`{"title":"Judul","body":"isi"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Create(c); err != nil {
		e.HTTPErrorHandler(err, c)
	}

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201, body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Judul") {
		t.Error("expected response body to contain note title")
	}
}

func TestNotesHandler_Create_MissingTitle(t *testing.T) {
	e := testEcho()
	h := NewNotesHandler(store.NewNoteStore())

	req := httptest.NewRequest(http.MethodPost, "/notes", strings.NewReader(`{"body":"tanpa judul"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Create(c)
	if err == nil {
		t.Fatal("expected validation error when title is missing")
	}
	e.HTTPErrorHandler(err, c)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400, body=%s", rec.Code, rec.Body.String())
	}
}

func TestNotesHandler_Detail_NotFound(t *testing.T) {
	e := testEcho()
	h := NewNotesHandler(store.NewNoteStore())

	req := httptest.NewRequest(http.MethodGet, "/notes/99", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("99")

	err := h.Detail(c)
	if err == nil {
		t.Fatal("expected not-found error")
	}
	e.HTTPErrorHandler(err, c)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestNotesHandler_List(t *testing.T) {
	e := testEcho()
	s := store.NewNoteStore()
	s.Create("Judul A", "isi A")
	h := NewNotesHandler(s)

	req := httptest.NewRequest(http.MethodGet, "/notes", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.List(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Judul A") {
		t.Error("expected response to contain note title")
	}
}
