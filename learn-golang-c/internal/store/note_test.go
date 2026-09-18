package store

import "testing"

// newTestNoteStore: DB SQLite in-memory (":memory:") - schema dibuat sama
// seperti OpenDB, tapi tidak menyentuh disk sama sekali. Tiap test dapat
// koneksi/DB terpisah, jadi tidak perlu dibersihkan manual.
func newTestNoteStore(t *testing.T) *NoteStore {
	t.Helper()
	db, err := OpenDB(":memory:")
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	s, err := NewNoteStore(db)
	if err != nil {
		t.Fatalf("NewNoteStore: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestNoteStore_CreateAndList(t *testing.T) {
	s := newTestNoteStore(t)
	if _, err := s.Create("A", "body a", nil); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := s.Create("B", "body b", []string{"file.png"}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	notes, err := s.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(notes) != 2 {
		t.Fatalf("got %d notes, want 2", len(notes))
	}
	// List() urut CreatedAt desc (tie-break ID desc) - "B" dibuat belakangan,
	// harus di atas.
	if notes[0].Title != "B" {
		t.Errorf("notes[0].Title = %q, want %q", notes[0].Title, "B")
	}
}

func TestNoteStore_GetDelete(t *testing.T) {
	s := newTestNoteStore(t)
	n, err := s.Create("A", "body", nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if _, err := s.Get(n.ID); err != nil {
		t.Fatalf("expected note to exist: %v", err)
	}
	ok, err := s.Delete(n.ID)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !ok {
		t.Fatal("expected Delete to succeed")
	}
	if _, err := s.Get(n.ID); err == nil {
		t.Error("expected note to be gone after Delete")
	}
	ok, err = s.Delete(n.ID)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if ok {
		t.Error("expected second Delete of same id to return false")
	}
}

func TestNoteStore_AddAttachments(t *testing.T) {
	s := newTestNoteStore(t)
	n, err := s.Create("A", "body", []string{"first.png"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	updated, err := s.AddAttachments(n.ID, []string{"second.png", "third.png"})
	if err != nil {
		t.Fatalf("expected AddAttachments to succeed on existing note: %v", err)
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

	if _, err := s.AddAttachments(9999, []string{"x.png"}); err == nil {
		t.Error("expected AddAttachments on unknown id to return an error")
	}
}
