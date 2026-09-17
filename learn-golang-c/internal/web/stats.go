package web

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"learn-golang-c/internal/store"
)

// stats.go: satu endpoint yang sengaja menggabung tiga chapter sekaligus -
// B.28 (http.TimeoutHandler dipasang PEMBUNGKUS di main.go), B.29 (baca
// username dari context, diisi BasicAuthMiddleware), B.30 (handler sendiri
// memantau context supaya kerja beratnya bisa berhenti lebih awal).
//
// Poin penting (dicatat eksplisit karena gampang salah paham): B.27/B.28
// http.TimeoutHandler HANYA memutus respons ke client setelah durasi
// habis - ia TIDAK membatalkan context ataupun menghentikan goroutine
// handler yang masih jalan di belakang layar. Supaya kerja berat itu
// betulan berhenti (bukan cuma responsnya "dibuang"), handler ini bikin
// context.WithTimeout SENDIRI dan mengecek ctx.Done() di setiap langkah
// perhitungan (buildSummary) - itu juga berarti ctx ini ikut ter-cancel
// kalau CLIENT memutus koneksi lebih dulu (B.30), beda dari r.Context()
// polos yang TimeoutHandler biarkan terus hidup sampai request asli selesai.
type HabitsStatsHandler struct {
	store  *store.HabitStore
	render *Renderer
	// perLambat: delay simulasi "kerja berat" per habit (mis. query
	// eksternal). Field, bukan konstanta, supaya test bisa mempercepatnya.
	perHabitDelay time.Duration
}

func NewHabitsStatsHandler(s *store.HabitStore, rd *Renderer, perHabitDelay time.Duration) *HabitsStatsHandler {
	return &HabitsStatsHandler{store: s, render: rd, perHabitDelay: perHabitDelay}
}

type statsResult struct {
	summary string
	err     error
}

// buildSummary: loop, bukan satu time.Sleep besar - tiap iterasi cek
// ctx.Done() dulu (select) sebelum lanjut ke habit berikutnya, jadi
// pembatalan beneran menghentikan progres di tengah, bukan cuma menahan
// hasil akhir yang sudah telanjur selesai dihitung.
func (h *HabitsStatsHandler) buildSummary(ctx context.Context) (string, error) {
	habits := h.store.List()
	var b strings.Builder
	for _, hb := range habits {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(h.perHabitDelay):
		}
		fmt.Fprintf(&b, "%s: streak %d hari\n", hb.Name, h.store.Streak(hb.ID))
	}
	return b.String(), nil
}

// Stats: GET /habits/stats, dipasang main.go lewat http.TimeoutHandler(3s).
func (h *HabitsStatsHandler) Stats(w http.ResponseWriter, r *http.Request) {
	// Timeout sendiri (bukan cuma andalkan http.TimeoutHandler di luar) -
	// lihat komentar di atas struct.
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resultCh := make(chan statsResult, 1)
	go func() {
		summary, err := h.buildSummary(ctx)
		resultCh <- statsResult{summary, err}
	}()

	select {
	case res := <-resultCh:
		if res.err != nil {
			http.Error(w, "stats dibatalkan: "+res.err.Error(), http.StatusRequestTimeout)
			return
		}
		username, _ := UsernameFromContext(r.Context())
		data := struct {
			Summary  string
			Username string
		}{Summary: res.summary, Username: username}
		h.render.Page(w, r, "habits_stats.html", data)
	case <-ctx.Done():
		http.Error(w, "stats dibatalkan: "+ctx.Err().Error(), http.StatusRequestTimeout)
	}
}
