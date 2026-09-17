package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"learn-golang-c/internal/store"
)

type HabitsHandler struct {
	store  *store.HabitStore
	render *Renderer
}

func NewHabitsHandler(s *store.HabitStore, rd *Renderer) *HabitsHandler {
	return &HabitsHandler{store: s, render: rd}
}

// habitRow: bentuk data khusus view (bukan store.Habit apa adanya) karena
// halaman butuh info turunan (sudah check-in hari ini? berapa streak-nya?)
// yang bukan bagian dari struct penyimpanan.
type habitRow struct {
	store.Habit
	CheckedInToday bool
	Streak         int
}

func (h *HabitsHandler) rows() []habitRow {
	today := time.Now().Format("2006-01-02")
	habits := h.store.List()
	rows := make([]habitRow, len(habits))
	for i, hb := range habits {
		rows[i] = habitRow{
			Habit:          hb,
			CheckedInToday: h.store.IsCheckedIn(hb.ID, today),
			Streak:         h.store.Streak(hb.ID),
		}
	}
	return rows
}

// List: GET /habits.
func (h *HabitsHandler) List(w http.ResponseWriter, r *http.Request) {
	data := struct{ Habits []habitRow }{Habits: h.rows()}
	h.render.Page(w, r, "habits_list.html", data, "habit_row.html")
}

// Dashboard: GET / - gabungan ringkas notes+habits, dipakai sebagai
// halaman B.1/B.2 (route root pertama) yang berkembang jadi dashboard
// sungguhan, bukan sekadar "hello world" statis.
func (h *HabitsHandler) Dashboard(notesStore *store.NoteStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			NotFoundHandler(h.render)(w, r)
			return
		}
		notes := notesStore.List()
		if len(notes) > 5 {
			notes = notes[:5]
		}
		data := struct {
			Habits     []habitRow
			RecentNote []store.Note
		}{
			Habits:     h.rows(),
			RecentNote: notes,
		}
		h.render.Page(w, r, "dashboard.html", data, "habit_row.html", "note_card.html")
	}
}

func (h *HabitsHandler) NewForm(w http.ResponseWriter, r *http.Request) {
	h.render.Page(w, r, "habit_form.html", nil)
}

// Create: POST /habits (form biasa, lewat browser - lihat api.go untuk
// versi JSON-nya, B.14).
func (h *HabitsHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "name wajib diisi", http.StatusBadRequest)
		return
	}
	hb := h.store.Create(name, r.FormValue("description"))
	SetFlash(w, fmt.Sprintf("Habit %q berhasil dibuat", hb.Name))
	http.Redirect(w, r, "/habits", http.StatusSeeOther)
}

func (h *HabitsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if h.store.Delete(id) {
		SetFlash(w, "Habit dihapus")
	}
	http.Redirect(w, r, "/habits", http.StatusSeeOther)
}

// Checkin (B.15 JSON response): POST /habits/{id}/checkin. Dipanggil
// fetch() kecil di habits_list.html (progressive enhancement - tombol
// toggle tanpa reload halaman), makanya balasnya JSON, bukan redirect.
func (h *HabitsHandler) Checkin(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "id tidak valid")
		return
	}

	today := time.Now().Format("2006-01-02")
	done, ok := h.store.ToggleCheckIn(id, today)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "habit tidak ditemukan")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"id":     id,
		"date":   today,
		"done":   done,
		"streak": h.store.Streak(id),
	})
}
