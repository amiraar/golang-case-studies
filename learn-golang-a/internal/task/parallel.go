package task

import (
	"fmt"
	"sort"
	"time"
)

// Fitur opsional (A.30-A.35): scan banyak file task sekaligus secara paralel.

// ScanResult: hasil scan satu file.
type ScanResult struct {
	Path  string
	Tasks []Task
	Err   error
}

// Dipanggil di dalam goroutine oleh ScanFiles/ScanFilesWithTimeout,
// dipisah supaya logikanya tak duplikat.
func scanOne(path string) ScanResult {
	s, err := LoadFromFile(path)
	if err != nil {
		return ScanResult{Path: path, Err: err}
	}
	return ScanResult{Path: path, Tasks: s.List()}
}

// ScanFiles (A.30 goroutine + A.32 buffered channel). Dipanggil main.go
// cmdScan saat tanpa -timeout-ms.
// Goroutine dipakai karena tiap file adalah I/O (nunggu disk); baca
// paralel bikin total waktu ~= file paling lambat, bukan jumlah semuanya.
// Channel dibuat BUFFERED sebesar jumlah file supaya tiap goroutine bisa
// kirim hasil dan langsung selesai tanpa menunggu penerima siap.
func ScanFiles(paths []string) []ScanResult {
	resultsCh := make(chan ScanResult, len(paths))
	for _, p := range paths {
		// path dikirim sebagai parameter closure, bukan pakai p langsung -
		// wajib untuk goroutine dalam loop, agar tiap goroutine dapat
		// nilai p miliknya sendiri, bukan p versi terakhir dari loop.
		go func(path string) {
			resultsCh <- scanOne(path)
		}(p)
	}

	// Terima persis len(paths) kali karena jumlah pengirim sudah pasti.
	results := make([]ScanResult, 0, len(paths))
	for i := 0; i < len(paths); i++ {
		results = append(results, <-resultsCh)
	}
	return results
}

// ScanFilesWithTimeout = ScanFiles + batas waktu. Dipanggil main.go
// cmdScan saat -timeout-ms > 0.
// A.33 select + A.35 timeout: menunggu salah satu channel siap duluan -
// ada hasil masuk, atau waktu habis. Cara idiomatik Go untuk "tunggu,
// tapi jangan lebih dari X detik".
func ScanFilesWithTimeout(paths []string, timeout time.Duration) ([]ScanResult, error) {
	resultsCh := make(chan ScanResult, len(paths))
	for _, p := range paths {
		go func(path string) {
			resultsCh <- scanOne(path)
		}(p)
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop() // A.36: bersihkan timer apa pun hasil akhirnya

	results := make([]ScanResult, 0, len(paths))
	for i := 0; i < len(paths); i++ {
		select {
		case r := <-resultsCh:
			results = append(results, r)
		case <-timer.C:
			// Waktu habis: kembalikan hasil yang sudah masuk, bukan dibuang semua.
			return results, fmt.Errorf("scan timed out after %s with %d/%d file(s) done", timeout, len(results), len(paths))
		}
	}
	return results, nil
}

// DuplicateTitles (A.66 generics): pakai Set[string] generik buat cari
// judul task yang muncul di lebih dari satu file saat scan gabungan -
// berguna buat ketahuan isi file yang tumpang tindih sebelum digabung
// manual. Dipanggil main.go cmdScan setelah ScanFiles/ScanFilesWithTimeout.
func DuplicateTitles(results []ScanResult) []string {
	seen := NewSet[string]()
	dup := NewSet[string]()
	for _, r := range results {
		for _, t := range r.Tasks {
			if seen.Has(t.Title) {
				dup.Add(t.Title)
			} else {
				seen.Add(t.Title)
			}
		}
	}
	out := dup.Slice()
	sort.Strings(out) // Set.Slice() urutannya acak (asal map) - diurutkan biar output konsisten
	return out
}
