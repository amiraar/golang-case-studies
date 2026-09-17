package checker

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadURLsFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "urls.txt")
	content := "# comment\n\ngo,https://go.dev\nhttps://example.com\n  \n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	entries, err := LoadURLsFromFile(path)
	if err != nil {
		t.Fatalf("LoadURLsFromFile() error = %v", err)
	}

	want := []URLEntry{
		{Label: "go", URL: "https://go.dev"},
		{Label: "", URL: "https://example.com"},
	}
	if len(entries) != len(want) {
		t.Fatalf("got %d entries, want %d: %+v", len(entries), len(want), entries)
	}
	for i, e := range entries {
		if e != want[i] {
			t.Errorf("entry %d = %+v, want %+v", i, e, want[i])
		}
	}
}

func TestLoadURLsFromFile_MissingFile(t *testing.T) {
	_, err := LoadURLsFromFile(filepath.Join(t.TempDir(), "missing.txt"))
	if err == nil {
		t.Fatal("LoadURLsFromFile() error = nil, want error for missing file")
	}
}

func TestLoadURLsFromFile_EmptyURLAfterComma(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "urls.txt")
	if err := os.WriteFile(path, []byte("label,\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := LoadURLsFromFile(path)
	if err == nil {
		t.Fatal("LoadURLsFromFile() error = nil, want error for empty url")
	}
}

func TestURLEntry_Name(t *testing.T) {
	if got := (URLEntry{Label: "go", URL: "https://go.dev"}).Name(); got != "go" {
		t.Errorf("Name() = %q, want %q", got, "go")
	}
	if got := (URLEntry{URL: "https://go.dev"}).Name(); got != "https://go.dev" {
		t.Errorf("Name() = %q, want %q", got, "https://go.dev")
	}
}
