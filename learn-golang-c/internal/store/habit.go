package store

import (
	"sort"
	"sync"
	"time"
)

type Habit struct {
	ID          int
	Name        string
	Description string
	CreatedAt   time.Time
}

// HabitStore: check-in disimpan terpisah dari Habit sendiri
// (habitID -> tanggal("2006-01-02") -> selesai/tidak) supaya satu habit
// bisa punya riwayat check-in banyak hari tanpa Habit membengkak jadi
// slice yang terus tumbuh di dalam struct-nya sendiri.
type HabitStore struct {
	mu       sync.RWMutex
	habits   map[int]Habit
	checkins map[int]map[string]bool
	nextID   int
}

func NewHabitStore() *HabitStore {
	return &HabitStore{
		habits:   make(map[int]Habit),
		checkins: make(map[int]map[string]bool),
		nextID:   1,
	}
}

func (s *HabitStore) Create(name, description string) Habit {
	s.mu.Lock()
	defer s.mu.Unlock()

	h := Habit{ID: s.nextID, Name: name, Description: description, CreatedAt: time.Now()}
	s.habits[h.ID] = h
	s.checkins[h.ID] = make(map[string]bool)
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

// ToggleCheckIn: dipanggil POST /habits/{id}/checkin (habits.go) DAN
// POST /api/habits/{id}/checkin (api.go) - toggle, bukan set, supaya satu
// tombol/satu endpoint cukup untuk tandai selesai maupun batalkan.
// ok=false kalau habit id tidak ada.
func (s *HabitStore) ToggleCheckIn(id int, date string) (done bool, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	days, exists := s.checkins[id]
	if !exists {
		return false, false
	}
	days[date] = !days[date]
	return days[date], true
}

func (s *HabitStore) IsCheckedIn(id int, date string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.checkins[id][date]
}

// Streak: hitung mundur dari hari ini, berhenti begitu ketemu hari yang
// belum di-check-in. Dipanggil dashboard ("/") dan Stats (B.28/B.30,
// stats.go) - sengaja diekspos sebagai method murni (tanpa I/O) supaya
// gampang dipanggil berkali-kali dari goroutine simulasi kerja berat di
// stats.go tanpa efek samping.
func (s *HabitStore) Streak(id int) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	days, ok := s.checkins[id]
	if !ok {
		return 0
	}

	streak := 0
	d := time.Now()
	for {
		key := d.Format("2006-01-02")
		if !days[key] {
			break
		}
		streak++
		d = d.AddDate(0, 0, -1)
	}
	return streak
}
