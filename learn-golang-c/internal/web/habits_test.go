package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"learn-golang-c/internal/store"
)

func TestHabitsHandler_Create(t *testing.T) {
	s := store.NewHabitStore()
	h := NewHabitsHandler(s, testRenderer())

	form := url.Values{"name": {"Olahraga"}, "description": {"pagi hari"}}
	req := httptest.NewRequest(http.MethodPost, "/habits", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.Create(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303, body=%s", w.Code, w.Body.String())
	}
	if len(s.List()) != 1 {
		t.Fatalf("got %d habits, want 1", len(s.List()))
	}
}

func TestHabitsHandler_Checkin_TogglesAndReturnsJSON(t *testing.T) {
	s := store.NewHabitStore()
	hb := s.Create("Baca", "")
	h := NewHabitsHandler(s, testRenderer())

	req := httptest.NewRequest(http.MethodPost, "/habits/1/checkin", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	h.Checkin(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", w.Code, w.Body.String())
	}
	var resp struct {
		Done   bool `json:"done"`
		Streak int  `json:"streak"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v (body=%s)", err, w.Body.String())
	}
	if !resp.Done || resp.Streak != 1 {
		t.Errorf("got done=%v streak=%d, want done=true streak=1", resp.Done, resp.Streak)
	}
	_ = hb
}

func TestHabitsHandler_Checkin_UnknownID(t *testing.T) {
	s := store.NewHabitStore()
	h := NewHabitsHandler(s, testRenderer())

	req := httptest.NewRequest(http.MethodPost, "/habits/999/checkin", nil)
	req.SetPathValue("id", "999")
	w := httptest.NewRecorder()
	h.Checkin(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}
