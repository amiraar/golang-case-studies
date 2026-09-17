package checker

import "sync"

// Stats (A.61 sync.Mutex): agregat total/ok/failed, di-update LANGSUNG oleh
// tiap worker goroutine di pool.go lewat Add - pendekatan yang beda dari
// channel "results" di pool.go: channel dipakai buat MENGALIRKAN nilai dari
// banyak goroutine ke satu pembaca; mutex di sini dipakai buat MELINDUNGI
// satu shared state (tiga counter) yang ditulis LANGSUNG oleh banyak
// goroutine sekaligus. Dua pendekatan concurrency Go yang sama-sama valid,
// dipilih sesuai kebutuhan - bukan salah satu yang "lebih benar".
type Stats struct {
	mu     sync.Mutex
	total  int
	ok     int
	failed int
}

// Add dipanggil dari goroutine worker manapun (pool.go) secara bersamaan.
// Tanpa mu.Lock(), s.total++ dari beberapa goroutine berbarengan adalah
// data race klasik: baca-ubah-tulis yang bisa saling menimpa, hasil akhir
// jadi lebih kecil dari jumlah request yang sebenarnya selesai.
func (s *Stats) Add(r CheckResult) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.total++
	if r.OK() {
		s.ok++
	} else {
		s.failed++
	}
}

// Snapshot: baca ketiga counter di bawah lock yang SAMA, supaya pembaca
// tak pernah lihat kombinasi angka yang "setengah ke-update" (mis. total
// sudah naik tapi ok/failed belum, kalau masing-masing punya lock sendiri).
func (s *Stats) Snapshot() (total, ok, failed int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.total, s.ok, s.failed
}
