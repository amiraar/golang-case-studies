// Package store: penyimpanan in-memory untuk Note dan Habit. Web server
// menangani banyak request bersamaan (tiap request = goroutine milik
// net/http), jadi struct di sini pakai sync.RWMutex (mirip A.61 di
// learn-golang-b, tapi RWMutex karena rasio baca:tulis di web app biasanya
// jomplang - banyak GET/list, jarang POST/create) supaya aman diakses
// paralel tanpa data race.
package store

import (
	"sort"
	"sync"
	"time"
)

// Note: domain "habit/note tracker" - catatan bebas, opsional ada lampiran
// file (diisi lewat B.13 file upload di internal/web/notes.go).
type Note struct {
	ID          int
	Title       string
	Body        string
	Attachments []string // nama file di bawah Upload.Dir, nil/kosong = tanpa lampiran
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type NoteStore struct {
	mu     sync.RWMutex
	notes  map[int]Note
	nextID int
}

func NewNoteStore() *NoteStore {
	return &NoteStore{notes: make(map[int]Note), nextID: 1}
}

// Create: dipanggil handler POST /notes (notes.go) setelah form di-parse.
func (s *NoteStore) Create(title, body string, attachments []string) Note {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	n := Note{
		ID:          s.nextID,
		Title:       title,
		Body:        body,
		Attachments: attachments,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	s.notes[n.ID] = n
	s.nextID++
	return n
}

// AddAttachments: dipanggil handler POST /notes/{id}/attachments (B.16,
// notes.go) setelah tiap part tersimpan ke disk - menambah nama file ke
// note yang SUDAH ada, terpisah dari Create karena upload lampiran susulan
// adalah aksi berbeda dari membuat note baru.
func (s *NoteStore) AddAttachments(id int, names []string) (Note, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	n, ok := s.notes[id]
	if !ok {
		return Note{}, false
	}
	n.Attachments = append(n.Attachments, names...)
	n.UpdatedAt = time.Now()
	s.notes[id] = n
	return n, true
}

// List: terbaru dulu (CreatedAt desc) - lebih natural untuk halaman note
// dibanding urut ID seperti Project 1, karena user mau lihat catatan
// terbaru di atas. Tie-break pakai ID desc (bukan cuma CreatedAt) karena
// dua Create() beruntun bisa dapat time.Now() yang identik kalau resolusi
// clock OS-nya kasar - ID yang auto-increment di bawah lock TIDAK PERNAH
// kembar, jadi urutan tetap deterministik walau timestamp-nya sama.
func (s *NoteStore) List() []Note {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]Note, 0, len(s.notes))
	for _, n := range s.notes {
		out = append(out, n)
	}
	sort.Slice(out, func(i, j int) bool {
		if !out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].CreatedAt.After(out[j].CreatedAt)
		}
		return out[i].ID > out[j].ID
	})
	return out
}

func (s *NoteStore) Get(id int) (Note, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	n, ok := s.notes[id]
	return n, ok
}

func (s *NoteStore) Delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.notes[id]; !ok {
		return false
	}
	delete(s.notes, id)
	return true
}
