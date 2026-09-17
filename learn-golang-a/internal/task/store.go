package task

import (
	"fmt"
	"sort"
	"time"
)

// FilterFunc: tipe fungsi, dipakai List() supaya kriteria filter fleksibel
// tanpa banyak parameter bool.
type FilterFunc func(Task) bool

// Closure (A.21): ByStatus/ByPriority "menangkap" s/p lalu dipanggil belakangan.
// Dipanggil dari main.go saat merakit filter dari flag -status/-priority.
func ByStatus(s Status) FilterFunc {
	return func(t Task) bool { return t.Status == s }
}

func ByPriority(p Priority) FilterFunc {
	return func(t Task) bool { return t.Priority == p }
}

// UpdateOption: pola "functional options" - caller kirim opsi yang relevan
// saja (mis. WithStatus) tanpa perlu isi semua field lain.
type UpdateOption func(*Task)

func WithTitle(title string) UpdateOption {
	return func(t *Task) { t.Title = title }
}

func WithStatus(s Status) UpdateOption {
	// Panggil markStatus (unexported) supaya UpdatedAt ikut berubah.
	return func(t *Task) { t.markStatus(s) }
}

func WithPriority(p Priority) UpdateOption {
	return func(t *Task) { t.Priority = p }
}

// Repository (A.27 interface): kontrak CRUD, dipisah dari implementasi
// konkret (Store) supaya kode pemanggil tidak terikat ke satu penyimpanan.
type Repository interface {
	Add(title string, priority Priority) (Task, error)
	Get(id int) (Task, error)
	Delete(id int) error
	Update(id int, opts ...UpdateOption) (Task, error)
	List(filters ...FilterFunc) []Task
}

type Store struct {
	// map[int]Task (A.17), bukan slice: akses utamanya cari-by-ID
	// (Get/Delete/Update), jadi map = O(1) langsung lewat key.
	tasks  map[int]Task
	nextID int // unexported (A.26): auto-increment, tak boleh diubah luar package
}

// Constructor. Dipanggil main.go (lewat LoadFromFile) dan test.
func NewStore() *Store {
	return &Store{tasks: make(map[int]Task), nextID: 1}
}

// Compile-time check: pastikan *Store memenuhi Repository.
var _ Repository = (*Store)(nil)

// (Task, error): pola standar Go, bukan exception. Caller wajib cek err.
// Dipanggil dari main.go cmdAdd.
func (s *Store) Add(title string, priority Priority) (Task, error) {
	if title == "" {
		return Task{}, ErrEmptyTitle
	}

	now := time.Now()
	t := Task{
		ID:        s.nextID,
		Title:     title,
		Status:    StatusPending,
		Priority:  priority,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.tasks[t.ID] = t
	s.nextID++
	return t, nil
}

// Get/Delete/Update: pola sama - cek map lookup dua-nilai (ok), bungkus
// ErrTaskNotFound dengan %w biar errors.Is masih jalan.
// Dipanggil dari main.go: cmdShow (Get), cmdDelete (Delete), cmdUpdate (Update).
func (s *Store) Get(id int) (Task, error) {
	t, ok := s.tasks[id]
	if !ok {
		return Task{}, fmt.Errorf("id %d: %w", id, ErrTaskNotFound)
	}
	return t, nil
}

func (s *Store) Delete(id int) error {
	if _, ok := s.tasks[id]; !ok {
		return fmt.Errorf("id %d: %w", id, ErrTaskNotFound)
	}
	delete(s.tasks, id)
	return nil
}

// opts ...UpdateOption (A.20 variadic): jumlah opsi bebas, jadi slice
// otomatis. Dipanggil main.go cmdUpdate dengan opts dari flag yang diisi.
func (s *Store) Update(id int, opts ...UpdateOption) (Task, error) {
	t, ok := s.tasks[id]
	if !ok {
		return Task{}, fmt.Errorf("id %d: %w", id, ErrTaskNotFound)
	}
	for _, opt := range opts {
		opt(&t) // fungsi sebagai parameter (A.22)
	}
	s.tasks[id] = t
	return t, nil
}

// filters ...FilterFunc (A.20 variadic). Dipanggil dari main.go cmdList,
// parallel.go (lewat scanOne), dan Stream() di bawah.
// Loop predicate manual (dulu pakai label "outer"+continue) sekarang
// diganti Filter[T] generik (A.66) - matches jadi satu closure yang
// mengevaluasi semua filter, dioper ke Filter yang generik atas tipe apa
// saja, bukan cuma []Task.
func (s *Store) List(filters ...FilterFunc) []Task {
	all := make([]Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		all = append(all, t)
	}

	matches := func(t Task) bool {
		for _, f := range filters {
			if !f(t) {
				return false
			}
		}
		return true
	}
	out := Filter(all, matches)

	// Urutan map tidak pasti, jadi diurutkan manual by ID biar konsisten.
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// Stream (opsional, A.30-31+A.34): goroutine kirim tiap Task ke channel
// lalu close saat selesai; consumer cukup `for t := range store.Stream()`.
// Dipanggil dari store_test.go.
func (s *Store) Stream() <-chan Task {
	ch := make(chan Task)
	snapshot := s.List()
	go func() {
		defer close(ch)
		for _, t := range snapshot {
			ch <- t
		}
	}()
	return ch
}
