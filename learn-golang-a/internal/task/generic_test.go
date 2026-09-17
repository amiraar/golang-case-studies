package task

import (
	"reflect"
	"testing"
)

func TestFilter(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5, 6}
	even := Filter(nums, func(n int) bool { return n%2 == 0 })
	if want := []int{2, 4, 6}; !reflect.DeepEqual(even, want) {
		t.Errorf("Filter(nums, even) = %v, want %v", even, want)
	}

	words := []string{"go", "rust", "c"}
	short := Filter(words, func(s string) bool { return len(s) <= 2 })
	if want := []string{"go", "c"}; !reflect.DeepEqual(short, want) {
		t.Errorf("Filter(words, short) = %v, want %v", short, want)
	}
}

func TestFilter_NoMatch(t *testing.T) {
	out := Filter([]int{1, 2, 3}, func(n int) bool { return n > 10 })
	if len(out) != 0 {
		t.Errorf("Filter() = %v, want empty slice", out)
	}
}

func TestSet(t *testing.T) {
	s := NewSet("a", "b", "a")
	if s.Len() != 2 {
		t.Fatalf("Len() = %d, want 2", s.Len())
	}
	if !s.Has("a") || !s.Has("b") {
		t.Error("Has() = false for an item that was added")
	}
	if s.Has("c") {
		t.Error("Has(\"c\") = true, want false")
	}

	s.Add("c")
	if s.Len() != 3 {
		t.Errorf("Len() after Add = %d, want 3", s.Len())
	}

	s.Remove("a")
	if s.Has("a") {
		t.Error("Has(\"a\") = true after Remove")
	}
	if s.Len() != 2 {
		t.Errorf("Len() after Remove = %d, want 2", s.Len())
	}
}

func TestSet_IntType(t *testing.T) {
	s := NewSet(1, 2, 3)
	if !s.Has(2) {
		t.Error("Has(2) = false, want true")
	}
}
