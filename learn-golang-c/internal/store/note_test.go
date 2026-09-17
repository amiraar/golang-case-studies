package store

import "testing"

func TestNoteStore_CreateAndList(t *testing.T) {
	s := NewNoteStore()
	s.Create("A", "body a", nil)
	s.Create("B", "body b", []string{"file.png"})

	notes := s.List()
	if len(notes) != 2 {
		t.Fatalf("got %d notes, want 2", len(notes))
	}
	// List() urut CreatedAt desc - "B" dibuat belakangan, harus di atas.
	if notes[0].Title != "B" {
		t.Errorf("notes[0].Title = %q, want %q", notes[0].Title, "B")
	}
}

func TestNoteStore_GetDelete(t *testing.T) {
	s := NewNoteStore()
	n := s.Create("A", "body", nil)

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

func TestNoteStore_AddAttachments(t *testing.T) {
	s := NewNoteStore()
	n := s.Create("A", "body", []string{"first.png"})

	updated, ok := s.AddAttachments(n.ID, []string{"second.png", "third.png"})
	if !ok {
		t.Fatal("expected AddAttachments to succeed on existing note")
	}
	want := []string{"first.png", "second.png", "third.png"}
	if len(updated.Attachments) != len(want) {
		t.Fatalf("got %d attachments, want %d", len(updated.Attachments), len(want))
	}
	for i, w := range want {
		if updated.Attachments[i] != w {
			t.Errorf("Attachments[%d] = %q, want %q", i, updated.Attachments[i], w)
		}
	}

	if _, ok := s.AddAttachments(9999, []string{"x.png"}); ok {
		t.Error("expected AddAttachments on unknown id to return false")
	}
}
