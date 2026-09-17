package web

import (
	"context"
	"log"
	"net/http"
	"time"
)

// contextKey (B.29 HTTP Handler Context Value): tipe khusus, BUKAN string
// polos, supaya key ini tidak pernah tabrakan dengan context key dari
// package lain (mis. kalau nanti import package pihak ketiga yang juga
// pakai context.WithValue dengan key string "username").
type contextKey string

const contextKeyUsername contextKey = "username"

// UsernameFromContext: dipanggil handler mana pun yang mau tahu siapa yang
// login (stats.go menampilkannya di halaman). ok=false kalau request tidak
// lewat BasicAuthMiddleware (mis. route publik).
func UsernameFromContext(ctx context.Context) (string, bool) {
	u, ok := ctx.Value(contextKeyUsername).(string)
	return u, ok
}

// LoggingMiddleware: dipasang global lewat CustomMux.Use di main.go.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s (%s)", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}

// RecoverMiddleware (B.26 Custom Error Handler, bagian panic recovery):
// recover() HARUS dipanggil langsung di dalam fungsi yang di-defer, tidak
// bisa "defer recover()" - itu sebabnya bentuknya defer func(){ ... }().
// Dipasang paling luar (lihat mux.go Use) supaya panic di middleware lain
// atau di handler mana pun tetap tertangkap, server tidak crash.
func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic recovered: %v", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// BasicAuthMiddleware (B.18 HTTP Basic Auth): factory, bukan middleware
// langsung, karena butuh username/password dari config (main.go) sebagai
// closure. Dipasang PER-ROUTE (bukan global lewat CustomMux.Use) di
// main.go - hanya route yang mengubah data (create/delete/checkin) yang
// diproteksi, GET/list tetap publik.
func BasicAuthMiddleware(username, password string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, p, ok := r.BasicAuth()
			if !ok || u != username || p != password {
				w.Header().Set("WWW-Authenticate", `Basic realm="notetracker"`)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			// B.29: simpan username ke context supaya handler di belakang
			// (mis. stats.go) bisa menyapa user tanpa parse header ulang.
			ctx := context.WithValue(r.Context(), contextKeyUsername, u)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
