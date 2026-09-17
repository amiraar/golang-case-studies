package task

import (
	"errors"
	"testing"
)

func TestStoreAdd(t *testing.T) {
	cases := []struct {
		name    string
		title   string
		wantErr error
	}{
		{"valid title", "buy milk", nil},
		{"empty title", "", ErrEmptyTitle},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewStore()
			got, err := s.Add(tc.title, PriorityMedium)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("Add() error = %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Add() unexpected error: %v", err)
			}
			if got.ID != 1 || got.Status != StatusPending {
				t.Errorf("Add() = %+v, want ID=1 Status=pending", got)
			}
		})
	}
}

func TestStoreDeleteAndGet(t *testing.T) {
	s := NewStore()
	t1, _ := s.Add("task 1", PriorityLow)

	if err := s.Delete(t1.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := s.Get(t1.ID); !errors.Is(err, ErrTaskNotFound) {
		t.Errorf("Get() after delete error = %v, want ErrTaskNotFound", err)
	}
	if err := s.Delete(t1.ID); !errors.Is(err, ErrTaskNotFound) {
		t.Errorf("Delete() twice error = %v, want ErrTaskNotFound", err)
	}
}

func TestStoreUpdate(t *testing.T) {
	s := NewStore()
	t1, _ := s.Add("task 1", PriorityLow)

	updated, err := s.Update(t1.ID, WithStatus(StatusDone), WithPriority(PriorityHigh))
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Status != StatusDone || updated.Priority != PriorityHigh {
		t.Errorf("Update() = %+v, want status=done priority=high", updated)
	}

	if _, err := s.Update(999, WithStatus(StatusDone)); !errors.Is(err, ErrTaskNotFound) {
		t.Errorf("Update() unknown id error = %v, want ErrTaskNotFound", err)
	}
}

func TestStoreListFilters(t *testing.T) {
	s := NewStore()
	a, _ := s.Add("a", PriorityLow)
	b, _ := s.Add("b", PriorityHigh)
	if _, err := s.Update(b.ID, WithStatus(StatusDone)); err != nil {
		t.Fatalf("setup Update() error = %v", err)
	}

	cases := []struct {
		name    string
		filters []FilterFunc
		wantIDs []int
	}{
		{"no filter", nil, []int{a.ID, b.ID}},
		{"by status done", []FilterFunc{ByStatus(StatusDone)}, []int{b.ID}},
		{"by priority low", []FilterFunc{ByPriority(PriorityLow)}, []int{a.ID}},
		{"combined no match", []FilterFunc{ByStatus(StatusDone), ByPriority(PriorityLow)}, []int{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := s.List(tc.filters...)
			if len(got) != len(tc.wantIDs) {
				t.Fatalf("List() = %d tasks, want %d", len(got), len(tc.wantIDs))
			}
			for i, id := range tc.wantIDs {
				if got[i].ID != id {
					t.Errorf("List()[%d].ID = %d, want %d", i, got[i].ID, id)
				}
			}
		})
	}
}

func TestStoreStream(t *testing.T) {
	s := NewStore()
	s.Add("a", PriorityLow)
	s.Add("b", PriorityHigh)

	count := 0
	for range s.Stream() {
		count++
	}
	if count != 2 {
		t.Errorf("Stream() yielded %d tasks, want 2", count)
	}
}
