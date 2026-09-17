package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBasicAuthMiddleware(t *testing.T) {
	var gotUsername string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUsername, _ = UsernameFromContext(r.Context()) // B.29
		w.WriteHeader(http.StatusOK)
	})
	protected := BasicAuthMiddleware("admin", "secret")(inner)

	t.Run("no credentials -> 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		w := httptest.NewRecorder()
		protected.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want 401", w.Code)
		}
	})

	t.Run("wrong password -> 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		req.SetBasicAuth("admin", "wrong")
		w := httptest.NewRecorder()
		protected.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want 401", w.Code)
		}
	})

	t.Run("correct credentials -> 200 and username in context", func(t *testing.T) {
		gotUsername = ""
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		req.SetBasicAuth("admin", "secret")
		w := httptest.NewRecorder()
		protected.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want 200", w.Code)
		}
		if gotUsername != "admin" {
			t.Errorf("UsernameFromContext = %q, want %q", gotUsername, "admin")
		}
	})
}
