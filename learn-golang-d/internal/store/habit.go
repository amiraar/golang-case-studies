package store

import (
	"sort"
	"sync"
	"time"
)

type Habit struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type HabitStore struct {
	mu       sync.RWMutex
	habits   map[int]Habit
	checkins map[int]map[string]bool // habitID -> "YYYY-MM-DD" -> done
	nextID   int
}

func NewHabitStore() *HabitStore {
	return &HabitStore{
		habits:   make(map[int]Habit),
		checkins: make(map[int]map[string]bool),
		nextID:   1,
	}
}

func (s *HabitStore) Create(name string) Habit {
	s.mu.Lock()
	defer s.mu.Unlock()

	h := Habit{ID: s.nextID, Name: name, CreatedAt: time.Now()}
	s.habits[h.ID] = h
	s.nextID++
	return h
}

func (s *HabitStore) List() []Habit {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]Habit, 0, len(s.habits))
	for _, h := range s.habits {
		out = append(out, h)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (s *HabitStore) Get(id int) (Habit, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	h, ok := s.habits[id]
	return h, ok
}

func (s *HabitStore) Delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.habits[id]; !ok {
		return false
	}
	delete(s.habits, id)
	delete(s.checkins, id)
	return true
}

// ToggleCheckIn: dipanggil handler POST /habits/{id}/checkin. Sama seperti
// learn-golang-c - toggle, bukan cuma set true, supaya checkin salah bisa
// dibatalkan tanpa endpoint terpisah.
func (s *HabitStore) ToggleCheckIn(id int, date string) (done, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.habits[id]; !exists {
		return false, false
	}
	if s.checkins[id] == nil {
		s.checkins[id] = make(map[string]bool)
	}
	s.checkins[id][date] = !s.checkins[id][date]
	return s.checkins[id][date], true
}

func (s *HabitStore) IsCheckedIn(id int, date string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.checkins[id][date]
}

// Streak: hitung mundur dari hari ini, berhenti di hari pertama yang belum
// checked-in.
func (s *HabitStore) Streak(id int) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	days := s.checkins[id]
	streak := 0
	for d := time.Now(); ; d = d.AddDate(0, 0, -1) {
		if !days[d.Format("2006-01-02")] {
			break
		}
		streak++
	}
	return streak
}
