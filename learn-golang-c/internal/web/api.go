package web

import (
	"encoding/json"
	"net/http"

	"learn-golang-c/internal/store"
)

// api.go: endpoint JSON murni untuk Habit, dipisah dari habits.go (yang
// isinya handler HTML form) supaya kontras dua gaya request/response
// (B.11-B.13 form vs B.14-B.15 JSON) kelihatan jelas di struktur kode,
// bukan cuma dicontohkan dalam satu handler campur aduk.
type APIHandler struct {
	store *store.HabitStore
}

func NewAPIHandler(s *store.HabitStore) *APIHandler {
	return &APIHandler{store: s}
}

// ListHabits (B.15 HTTP Response: JSON): GET /api/habits.
func (h *APIHandler) ListHabits(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// json.NewEncoder(w) langsung ke ResponseWriter (io.Writer) lebih
	// efisien dibanding json.Marshal+w.Write karena tak perlu buffer
	// perantara berisi seluruh []byte hasil encode.
	if err := json.NewEncoder(w).Encode(h.store.List()); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
	}
}

type createHabitPayload struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CreateHabit (B.14 HTTP Request JSON Payload): POST /api/habits.
// json.NewDecoder(r.Body), bukan io.ReadAll+json.Unmarshal - Body adalah
// io.Reader/stream, Decoder baca langsung dari situ tanpa nampung seluruh
// body di memori dulu.
func (h *APIHandler) CreateHabit(w http.ResponseWriter, r *http.Request) {
	var payload createHabitPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, "payload JSON tidak valid: "+err.Error())
		return
	}
	if payload.Name == "" {
		writeJSONError(w, http.StatusBadRequest, "name wajib diisi")
		return
	}

	hb := h.store.Create(payload.Name, payload.Description)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(hb)
}
