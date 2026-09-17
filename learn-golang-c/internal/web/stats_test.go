package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"learn-golang-c/internal/store"
)

func TestHabitsStatsHandler_CompletesAndShowsUsername(t *testing.T) {
	s := store.NewHabitStore()
	s.Create("Olahraga", "")
	// delay kecil (bukan 0) supaya select di buildSummary tetap melewati
	// cabang time.After, bukan langsung selesai tanpa benar-benar menunggu.
	h := NewHabitsStatsHandler(s, testRenderer(), 5*time.Millisecond)

	req := httptest.NewRequest(http.MethodGet, "/habits/stats", nil)
	ctx := context.WithValue(req.Context(), contextKeyUsername, "admin") // B.29, biasanya diisi BasicAuthMiddleware
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	h.Stats(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "admin") {
		t.Error("expected page to show username from context")
	}
	if !strings.Contains(w.Body.String(), "Olahraga") {
		t.Error("expected page to show habit summary")
	}
}

// TestHabitsStatsHandler_ClientCancelStopsEarly (B.30): context di-cancel
// SEBELUM handler sempat selesai - dibuktikan lewat status 408, bukan
// nunggu penuh sampai buildSummary selesai wajar.
func TestHabitsStatsHandler_ClientCancelStopsEarly(t *testing.T) {
	s := store.NewHabitStore()
	for i := 0; i < 5; i++ {
		s.Create("Habit", "")
	}
	// Delay lumayan (100ms x 5 habit = 500ms) supaya cancel yang dikirim
	// setelah 10ms PASTI mendahului penyelesaian normal.
	h := NewHabitsStatsHandler(s, testRenderer(), 100*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest(http.MethodGet, "/habits/stats", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()
	h.Stats(w, req)

	if w.Code != http.StatusRequestTimeout {
		t.Errorf("status = %d, want 408 (context cancelled)", w.Code)
	}
}
