// Package store: penyimpanan in-memory, sengaja tetap sederhana (bukan
// SQL) karena fokus belajar Project 4 ada di lapisan framework (C block),
// bukan di lapisan storage - kalau storage juga diganti sekaligus, sulit
// membedakan mana perubahan akibat Echo dan mana akibat DB.
package store

import (
	"sort"
	"sync"
	"time"
)

type Note struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type NoteStore struct {
	mu     sync.RWMutex
	notes  map[int]Note
	nextID int
}

func NewNoteStore() *NoteStore {
	return &NoteStore{notes: make(map[int]Note), nextID: 1}
}

func (s *NoteStore) Create(title, body string) Note {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	n := Note{ID: s.nextID, Title: title, Body: body, CreatedAt: now, UpdatedAt: now}
	s.notes[n.ID] = n
	s.nextID++
	return n
}

// List: terbaru dulu, tie-break ID desc - alasan sama seperti
// learn-golang-c/internal/store/note.go (dua Create() beruntun bisa
// dapat time.Now() identik di clock beresolusi kasar).
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
