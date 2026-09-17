package web

import (
	"encoding/json"
	"net/http"
)

// NotFoundHandler (B.26 Custom Error Handler): http.ServeMux modern (Go
// 1.22+) sebetulnya sudah otomatis balas 404 untuk path yang tak
// ke-match, TAPI itu 404 polos bawaan net/http (plain text "404 page not
// found"). Fungsi ini dipakai main.go untuk route "/" (catch-all subtree,
// persis idiom tutorial) supaya path tak dikenal dapat halaman 404 yang
// konsisten dengan tampilan situs, bukan teks polos.
func NotFoundHandler(rd *Renderer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		rd.Page(w, r, "404.html", nil)
	}
}

// writeJSONError: dipakai api.go supaya semua error endpoint JSON
// konsisten bentuknya ({"error": "..."}), bukan campur antara
// http.Error (text/plain) dan body JSON.
func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
