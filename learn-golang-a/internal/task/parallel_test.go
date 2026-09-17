package task

import (
	"path/filepath"
	"testing"
	"time"
)

func TestScanFiles(t *testing.T) {
	dir := t.TempDir()

	s1 := NewStore()
	s1.Add("a", PriorityLow)
	path1 := filepath.Join(dir, "s1.json")
	if err := s1.SaveToFile(path1); err != nil {
		t.Fatalf("SaveToFile() error = %v", err)
	}

	s2 := NewStore()
	s2.Add("b", PriorityHigh)
	s2.Add("c", PriorityMedium)
	path2 := filepath.Join(dir, "s2.json")
	if err := s2.SaveToFile(path2); err != nil {
		t.Fatalf("SaveToFile() error = %v", err)
	}

	missing := filepath.Join(dir, "missing.json")

	results := ScanFiles([]string{path1, path2, missing})
	if len(results) != 3 {
		t.Fatalf("ScanFiles() returned %d results, want 3", len(results))
	}

	byPath := make(map[string]ScanResult, len(results))
	for _, r := range results {
		byPath[r.Path] = r
	}

	if len(byPath[path1].Tasks) != 1 {
		t.Errorf("%s: got %d tasks, want 1", path1, len(byPath[path1].Tasks))
	}
	if len(byPath[path2].Tasks) != 2 {
		t.Errorf("%s: got %d tasks, want 2", path2, len(byPath[path2].Tasks))
	}
	if byPath[missing].Err != nil {
		t.Errorf("missing file should load as an empty store, got err %v", byPath[missing].Err)
	}
}

func TestScanFilesWithTimeout(t *testing.T) {
	dir := t.TempDir()
	s := NewStore()
	s.Add("a", PriorityLow)
	path := filepath.Join(dir, "s.json")
	if err := s.SaveToFile(path); err != nil {
		t.Fatalf("SaveToFile() error = %v", err)
	}

	results, err := ScanFilesWithTimeout([]string{path}, time.Second)
	if err != nil {
		t.Fatalf("ScanFilesWithTimeout() error = %v", err)
	}
	if len(results) != 1 || len(results[0].Tasks) != 1 {
		t.Errorf("unexpected results: %+v", results)
	}
}

func TestDuplicateTitles(t *testing.T) {
	results := []ScanResult{
		{Path: "a.json", Tasks: []Task{{Title: "buy milk"}, {Title: "write report"}}},
		{Path: "b.json", Tasks: []Task{{Title: "buy milk"}, {Title: "clean desk"}}},
		{Path: "c.json", Tasks: []Task{{Title: "write report"}}},
	}

	got := DuplicateTitles(results)
	want := []string{"buy milk", "write report"}
	if len(got) != len(want) {
		t.Fatalf("DuplicateTitles() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("DuplicateTitles()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestDuplicateTitles_NoneFound(t *testing.T) {
	results := []ScanResult{
		{Path: "a.json", Tasks: []Task{{Title: "buy milk"}}},
		{Path: "b.json", Tasks: []Task{{Title: "clean desk"}}},
	}
	if got := DuplicateTitles(results); len(got) != 0 {
		t.Errorf("DuplicateTitles() = %v, want empty", got)
	}
}
