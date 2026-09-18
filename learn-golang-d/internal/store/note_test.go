package store

import "testing"

func TestNoteStore_CreateAndList(t *testing.T) {
	s := NewNoteStore()
	s.Create("A", "body a")
	s.Create("B", "body b")

	notes := s.List()
	if len(notes) != 2 {
		t.Fatalf("got %d notes, want 2", len(notes))
	}
	if notes[0].Title != "B" {
		t.Errorf("notes[0].Title = %q, want %q", notes[0].Title, "B")
	}
}

func TestNoteStore_GetDelete(t *testing.T) {
	s := NewNoteStore()
	n := s.Create("A", "body")

	if _, ok := s.Get(n.ID); !ok {
		t.Fatal("expected note to exist")
	}
	if !s.Delete(n.ID) {
		t.Fatal("expected Delete to succeed")
	}
	if _, ok := s.Get(n.ID); ok {
		t.Error("expected note to be gone after Delete")
	}
	if s.Delete(n.ID) {
		t.Error("expected second Delete of same id to return false")
	}
}
