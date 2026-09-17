package checker

import (
	"context"
	"math/rand"
	"sync"
	"time"
)

// Options: A.9 variabel/A.24 struct sebagai satu paket konfigurasi,
// dioper by value karena kecil dan tak pernah diubah setelah dibuat.
type Options struct {
	Concurrency       int           // jumlah worker goroutine (A.30) yang jalan bersamaan
	PerRequestTimeout time.Duration // A.42: batas waktu satu request, diteruskan ke Checker.CheckOne
	OverallTimeout    time.Duration // A.42: batas waktu seluruh scan, 0 = tanpa batas
	MaxJitter         time.Duration // A.39 random: jeda acak 0..MaxJitter sebelum tiap request, biar tidak membanjiri server serentak
	ProgressEvery     time.Duration // A.41 ticker: seberapa sering callback progress dipanggil
}

// ProgressFunc dipanggil dari goroutine pengumpul hasil, bukan dari worker -
// jadi aman dipakai untuk print ke stderr tanpa data race.
type ProgressFunc func(done, total int)

// CheckAll menjalankan pengecekan semua entries dengan worker pool.
//
// Pola: satu channel "jobs" dibagi ke N worker goroutine (A.30), tiap hasil
// dikirim ke channel "results" (A.31, dibuat BUFFERED A.32 sebesar
// len(entries) supaya worker tak pernah ke-block ngirim walau pengumpul
// keburu berhenti baca karena overall timeout). Goroutine terpisah menutup
// "results" (A.34 close) setelah semua worker selesai (sync.WaitGroup),
// dan goroutine pengumpul di bawah pakai range (A.34) + select (A.33)
// untuk baca hasil sambil juga dengar ticker progress dan overall timeout.
//
// Stats yang dikembalikan (A.61 sync.Mutex) dihitung dengan cara yang
// SENGAJA berbeda dari "out" di atas: worker nulis langsung ke shared
// counter yang dikunci mutex, bukan lewat channel. Di akhir fungsi kedua
// angka (len(out) vs stats.Snapshot()) seharusnya selalu sama - dua jalan
// berbeda menuju hasil yang identik.
func CheckAll(entries []URLEntry, opts Options, progress ProgressFunc) ([]CheckResult, *Stats) {
	if opts.Concurrency < 1 {
		opts.Concurrency = 1
	}

	ctx := context.Background()
	if opts.OverallTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(ctx)
		defer cancel()
		// A.41 timer/scheduler: time.AfterFunc menjadwalkan cancel() sendiri
		// begitu OverallTimeout lewat, tanpa perlu goroutine+select manual
		// khusus untuk ini - beda gaya dari context.WithTimeout supaya
		// idiom "Timer & Scheduler" (bukan cuma Ticker) ikut kepakai.
		timer := time.AfterFunc(opts.OverallTimeout, cancel)
		defer timer.Stop()
	}

	jobs := make(chan URLEntry)
	results := make(chan CheckResult, len(entries))
	checker := NewChecker()
	cache := newResultCache()
	stats := &Stats{}

	var wg sync.WaitGroup
	for i := 0; i < opts.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			runWorker(ctx, jobs, results, checker, cache, stats, opts)
		}()
	}

	go func() {
		for _, e := range entries {
			jobs <- e
		}
		close(jobs) // A.34 close: sinyal "tidak ada job lagi" ke semua worker
	}()

	go func() {
		wg.Wait()
		close(results) // A.34 close: baru aman ditutup setelah SEMUA worker selesai kirim
	}()

	var ticker *time.Ticker
	var tickC <-chan time.Time
	if opts.ProgressEvery > 0 {
		ticker = time.NewTicker(opts.ProgressEvery) // A.41 ticker
		defer ticker.Stop()
		tickC = ticker.C
	}

	out := make([]CheckResult, 0, len(entries))
	done := 0
collect:
	for {
		select { // A.33 select: tiga sumber event dimultipleks jadi satu loop
		case r, ok := <-results:
			if !ok {
				break collect // results ditutup = semua worker sudah selesai
			}
			out = append(out, r)
			done++
			if progress != nil {
				progress(done, len(entries))
			}
		case <-tickC:
			if progress != nil {
				progress(done, len(entries))
			}
		case <-ctx.Done():
			break collect // overall timeout: kembalikan hasil yang sudah masuk, bukan tunggu semuanya
		}
	}
	return out, stats
}

// runWorker: badan satu worker goroutine, dipisah dari CheckAll supaya
// tidak ada closure raksasa. Berhenti otomatis begitu "jobs" ditutup (A.34).
func runWorker(ctx context.Context, jobs <-chan URLEntry, results chan<- CheckResult, checker *Checker, cache *resultCache, stats *Stats, opts Options) {
	for entry := range jobs {
		// Cache hit: URL yang sama sudah pernah dicek worker lain - entry
		// disalin ulang biar label pemanggil saat ini yang dipakai, bukan
		// label dari worker yang pertama kali ngecek URL ini.
		if cached, ok := cache.get(entry.URL); ok {
			cached.Entry = entry
			results <- cached
			stats.Add(cached) // A.61 mutex: dicatat juga di jalur counter, biar Stats tetap merepresentasikan tiap entry input
			continue
		}

		if opts.MaxJitter > 0 {
			time.Sleep(randomJitter(opts.MaxJitter)) // A.39 random
		}
		r := checker.CheckOne(ctx, entry, opts.PerRequestTimeout)
		cache.set(entry.URL, r)
		stats.Add(r) // A.61 mutex: banyak worker nulis ke counter yang sama, dilindungi Stats.mu
		results <- r
	}
}

// randomJitter: A.39 math/rand. Sejak Go 1.20 sumber global rand sudah
// otomatis ter-seed acak per proses, jadi tak perlu rand.Seed manual lagi.
func randomJitter(max time.Duration) time.Duration {
	if max <= 0 {
		return 0
	}
	return time.Duration(rand.Int63n(int64(max)))
}
