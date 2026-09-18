package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func signTestToken(t *testing.T, key []byte, username string, expired bool) string {
	t.Helper()
	exp := time.Now().Add(time.Hour)
	if expired {
		exp = time.Now().Add(-time.Hour)
	}
	claims := AppClaims{
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(exp)},
		Username:         username,
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(key)
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func TestJWTMiddleware_ValidToken(t *testing.T) {
	e := testEcho()
	key := []byte("test-secret")
	mw := JWTMiddleware(key)

	var gotUsername string
	next := func(c echo.Context) error {
		gotUsername = UsernameFromContext(c)
		return c.NoContent(http.StatusOK)
	}

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+signTestToken(t, key, "admin", false))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := mw(next)(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotUsername != "admin" {
		t.Errorf("username in context = %q, want %q", gotUsername, "admin")
	}
}

func TestJWTMiddleware_MissingHeader(t *testing.T) {
	e := testEcho()
	mw := JWTMiddleware([]byte("test-secret"))
	next := func(c echo.Context) error { return c.NoContent(http.StatusOK) }

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := mw(next)(c)
	if err == nil {
		t.Fatal("expected error for missing Authorization header")
	}
	e.HTTPErrorHandler(err, c)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestJWTMiddleware_ExpiredToken(t *testing.T) {
	e := testEcho()
	key := []byte("test-secret")
	mw := JWTMiddleware(key)
	next := func(c echo.Context) error { return c.NoContent(http.StatusOK) }

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+signTestToken(t, key, "admin", true))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := mw(next)(c)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
	e.HTTPErrorHandler(err, c)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestJWTMiddleware_WrongSigningKey(t *testing.T) {
	e := testEcho()
	mw := JWTMiddleware([]byte("correct-secret"))
	next := func(c echo.Context) error { return c.NoContent(http.StatusOK) }

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+signTestToken(t, []byte("wrong-secret"), "admin", false))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := mw(next)(c)
	if err == nil {
		t.Fatal("expected error for token signed with wrong key")
	}
	e.HTTPErrorHandler(err, c)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}
