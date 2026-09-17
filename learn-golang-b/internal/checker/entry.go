package checker

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// URLEntry: satu baris input. Label opsional (A.9 variabel, A.10 tipe data
// string) cuma dipakai untuk label tampilan, tidak memengaruhi request.
type URLEntry struct {
	Label string
	URL   string
}

// LoadURLsFromFile (A.50 file): format tiap baris "url" atau "label,url".
// Baris kosong dan diawali "#" dilewati supaya file input bisa dikomentari,
// mirip .gitignore.
func LoadURLsFromFile(path string) ([]URLEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("load urls: %w", err)
	}
	defer f.Close() // A.36 defer: file selalu ditutup walau ada error di tengah scan

	var entries []URLEntry
	scanner := bufio.NewScanner(f)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		label, url := "", line
		if idx := strings.Index(line, ","); idx != -1 {
			label = strings.TrimSpace(line[:idx])
			url = strings.TrimSpace(line[idx+1:])
		}
		if url == "" {
			return nil, fmt.Errorf("load urls: line %d: empty url", lineNo)
		}
		entries = append(entries, URLEntry{Label: label, URL: url})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("load urls: %w", err)
	}
	return entries, nil
}

// Name: label kalau ada, kalau tidak fallback ke URL. Dipakai result.go
// biar output tetap enak dibaca meski file input tak kasih label.
func (e URLEntry) Name() string {
	if e.Label != "" {
		return e.Label
	}
	return e.URL
}
