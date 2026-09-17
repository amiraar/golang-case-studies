package task

import (
	"path/filepath"
	"testing"
)

func TestSaveAndLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tasks.json")

	s := NewStore()
	s.Add("task 1", PriorityLow)
	s.Add("task 2", PriorityHigh)

	if err := s.SaveToFile(path); err != nil {
		t.Fatalf("SaveToFile() error = %v", err)
	}

	loaded, err := LoadFromFile(path)
	if err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}

	got := loaded.List()
	if len(got) != 2 {
		t.Fatalf("List() = %d tasks, want 2", len(got))
	}
	if got[0].Title != "task 1" || got[1].Title != "task 2" {
		t.Errorf("unexpected tasks after round trip: %+v", got)
	}

	next, err := loaded.Add("task 3", PriorityMedium)
	if err != nil {
		t.Fatalf("Add() after load error = %v", err)
	}
	if next.ID != 3 {
		t.Errorf("Add() after load ID = %d, want 3 (nextID must persist)", next.ID)
	}
}

func TestLoadFromFileMissing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "does-not-exist.json")

	s, err := LoadFromFile(path)
	if err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}
	if len(s.List()) != 0 {
		t.Errorf("expected empty store for missing file, got %d tasks", len(s.List()))
	}
}
