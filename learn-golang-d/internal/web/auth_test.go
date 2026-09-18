package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestAuthHandler_Login_Success(t *testing.T) {
	e := testEcho()
	h := NewAuthHandler("admin", "admin123", []byte("test-secret"), 60)

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"username":"admin","password":"admin123"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Login(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "token") {
		t.Error("expected response to contain a token field")
	}
}

func TestAuthHandler_Login_WrongPassword(t *testing.T) {
	e := testEcho()
	h := NewAuthHandler("admin", "admin123", []byte("test-secret"), 60)

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"username":"admin","password":"salah"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Login(c)
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
	e.HTTPErrorHandler(err, c)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestAuthHandler_Login_MissingFields(t *testing.T) {
	e := testEcho()
	h := NewAuthHandler("admin", "admin123", []byte("test-secret"), 60)

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Login(c)
	if err == nil {
		t.Fatal("expected validation error for empty payload")
	}
	e.HTTPErrorHandler(err, c)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}
