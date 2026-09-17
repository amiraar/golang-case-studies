package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"learn-golang-c/internal/store"
)

func TestAPIHandler_CreateAndList(t *testing.T) {
	s := store.NewHabitStore()
	h := NewAPIHandler(s)

	createReq := httptest.NewRequest(http.MethodPost, "/api/habits", strings.NewReader(`{"name":"Meditasi","description":"10 menit"}`))
	createW := httptest.NewRecorder()
	h.CreateHabit(createW, createReq)

	if createW.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want 201, body=%s", createW.Code, createW.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/habits", nil)
	listW := httptest.NewRecorder()
	h.ListHabits(listW, listReq)

	var habits []store.Habit
	if err := json.Unmarshal(listW.Body.Bytes(), &habits); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(habits) != 1 || habits[0].Name != "Meditasi" {
		t.Errorf("got %+v, want one habit named Meditasi", habits)
	}
}

func TestAPIHandler_CreateHabit_MissingName(t *testing.T) {
	s := store.NewHabitStore()
	h := NewAPIHandler(s)

	req := httptest.NewRequest(http.MethodPost, "/api/habits", strings.NewReader(`{"description":"tanpa nama"}`))
	w := httptest.NewRecorder()
	h.CreateHabit(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestAPIHandler_CreateHabit_MalformedJSON(t *testing.T) {
	s := store.NewHabitStore()
	h := NewAPIHandler(s)

	req := httptest.NewRequest(http.MethodPost, "/api/habits", strings.NewReader(`not json`))
	w := httptest.NewRecorder()
	h.CreateHabit(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}
