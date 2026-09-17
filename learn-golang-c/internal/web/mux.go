// Package web: server HTTP net/http murni (tanpa framework - beda dari
// learn-golang-d yang nanti pakai Echo). Semua chapter B.1-B.30 dipetakan
// ke file-file di package ini.
package web

import "net/http"

// CustomMux (B.20 Custom Multiplexer): membungkus http.ServeMux bawaan
// (yang sejak Go 1.22 sudah bisa "METHOD /path/{id}" - jadi B.20 di sini
// TIDAK dipakai untuk menambal keterbatasan matching method seperti versi
// tutorial aslinya, itu sudah beres di ServeMux sendiri) melainkan untuk
// satu hal yang ServeMux TIDAK punya: cara mendaftarkan middleware global
// lewat method (Use), bukan pembungkusan manual berlapis di main.go.
type CustomMux struct {
	http.ServeMux
	middlewares []func(http.Handler) http.Handler
}

func NewCustomMux() *CustomMux {
	return &CustomMux{}
}

// Use: middleware yang di-register PALING TERAKHIR jadi PALING LUAR
// (lihat ServeHTTP) - dieksekusi paling awal saat request masuk. Dipanggil
// main.go: Use(Logging) dulu baru Use(Recover), supaya Recover membungkus
// semuanya termasuk Logging (panic di middleware lain pun masih tertangkap).
func (m *CustomMux) Use(mw func(http.Handler) http.Handler) {
	m.middlewares = append(m.middlewares, mw)
}

// ServeHTTP (B.19 http.Handler interface): CustomMux sendiri jadi
// http.Handler yang valid, dioper langsung ke http.Server.Handler di
// main.go. Tiap middleware membungkus current secara berurutan, persis
// pola "var handler http.Handler = mux; handler = MwA(handler)" tutorial,
// hanya saja daftarnya disimpan di slice supaya jumlahnya fleksibel.
func (m *CustomMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var current http.Handler = &m.ServeMux
	for _, mw := range m.middlewares {
		current = mw(current)
	}
	current.ServeHTTP(w, r)
}
