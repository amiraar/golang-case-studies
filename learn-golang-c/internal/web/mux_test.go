package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestCustomMux_MiddlewareOrder (B.20): middleware yang di-Use PALING
// TERAKHIR harus jadi PALING LUAR - dieksekusi paling awal. Dibuktikan
// lewat urutan tulis ke slice bersama, bukan cuma asumsi dari komentar.
func TestCustomMux_MiddlewareOrder(t *testing.T) {
	var order []string

	track := func(name string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name)
				next.ServeHTTP(w, r)
			})
		}
	}

	mux := NewCustomMux()
	mux.Use(track("first-registered"))
	mux.Use(track("second-registered"))
	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	want := []string{"second-registered", "first-registered", "handler"}
	if len(order) != len(want) {
		t.Fatalf("order = %v, want %v", order, want)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Errorf("order[%d] = %q, want %q (full: %v)", i, order[i], want[i], order)
		}
	}
}

func TestCustomMux_RecoverStopsPanic(t *testing.T) {
	mux := NewCustomMux()
	mux.Use(RecoverMiddleware)
	mux.HandleFunc("GET /boom", func(w http.ResponseWriter, r *http.Request) {
		panic("kaboom")
	})

	req := httptest.NewRequest(http.MethodGet, "/boom", nil)
	w := httptest.NewRecorder()

	// Tanpa RecoverMiddleware, panic ini akan menjatuhkan test process.
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}
