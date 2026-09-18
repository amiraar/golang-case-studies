package store

import (
	"testing"
	"time"
)

func TestHabitStore_ToggleCheckIn(t *testing.T) {
	s := NewHabitStore()
	h := s.Create("Olahraga")
	today := time.Now().Format("2006-01-02")

	done, ok := s.ToggleCheckIn(h.ID, today)
	if !ok || !done {
		t.Fatalf("first toggle: done=%v ok=%v, want true/true", done, ok)
	}
	done, ok = s.ToggleCheckIn(h.ID, today)
	if !ok || done {
		t.Fatalf("second toggle: done=%v ok=%v, want false/true", done, ok)
	}
}

func TestHabitStore_ToggleCheckIn_UnknownID(t *testing.T) {
	s := NewHabitStore()
	if _, ok := s.ToggleCheckIn(999, "2026-01-01"); ok {
		t.Error("expected ok=false for unknown habit id")
	}
}

func TestHabitStore_Streak(t *testing.T) {
	s := NewHabitStore()
	h := s.Create("Baca")

	today := time.Now()
	for i := 0; i < 3; i++ {
		date := today.AddDate(0, 0, -i).Format("2006-01-02")
		s.ToggleCheckIn(h.ID, date)
	}

	if got := s.Streak(h.ID); got != 3 {
		t.Errorf("Streak() = %d, want 3", got)
	}
}

func TestHabitStore_Delete(t *testing.T) {
	s := NewHabitStore()
	h := s.Create("X")
	if !s.Delete(h.ID) {
		t.Fatal("expected Delete to succeed")
	}
	if _, ok := s.Get(h.ID); ok {
		t.Error("expected habit to be gone after Delete")
	}
}
