package checker

import (
	"context"
	"net/http"
	"time"
)

// Checker membungkus http.Client (A.55 client HTTP) supaya gampang dites
// pakai httptest.Server tanpa menyentuh jaringan asli.
type Checker struct {
	httpClient *http.Client
}

func NewChecker() *Checker {
	return &Checker{httpClient: &http.Client{}}
}

// CheckOne mengecek satu URL dengan batas waktu (A.35 channel timeout).
// Request jalan di goroutine terpisah (A.30) dan kirim hasilnya lewat
// channel "done" (A.31); select (A.33) menunggu SALAH SATU dari dua hal
// duluan: hasil datang, atau ctx keburu habis waktu.
//
// ctx dipakai (bukan time.After polos seperti versi paling dasar di
// tutorial A.35) supaya saat timeout, request HTTP yang masih jalan di
// goroutine benar-benar dibatalkan lewat http.NewRequestWithContext -
// goroutine-nya ikut selesai, tidak menggantung nunggu jaringan selamanya.
func (c *Checker) CheckOne(ctx context.Context, entry URLEntry, timeout time.Duration) CheckResult {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel() // A.36: pastikan timer internal context ikut dibersihkan

	start := time.Now()
	done := make(chan CheckResult, 1) // A.32 buffered: goroutine tetap bisa kirim & selesai walau sisi select sudah keburu ambil jalur ctx.Done()

	go func() {
		done <- c.do(ctx, entry, start)
	}()

	select {
	case r := <-done:
		return r
	case <-ctx.Done():
		return CheckResult{
			Entry:     entry,
			CheckedAt: start,
			Duration:  time.Since(start),
			Err:       ctx.Err(),
		}
	}
}

// do: request HTTP sebenarnya, dipanggil dari goroutine di CheckOne.
func (c *Checker) do(ctx context.Context, entry URLEntry, start time.Time) CheckResult {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, entry.URL, nil)
	if err != nil {
		return CheckResult{Entry: entry, CheckedAt: start, Duration: time.Since(start), Err: err}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return CheckResult{Entry: entry, CheckedAt: start, Duration: time.Since(start), Err: err}
	}
	defer resp.Body.Close() // A.36 defer: body wajib ditutup biar koneksi bisa dipakai ulang

	result := CheckResult{
		Entry:      entry,
		StatusCode: resp.StatusCode,
		Duration:   time.Since(start),
		CheckedAt:  start,
	}

	// A.40 parsing time: header Date HTTP formatnya RFC1123, beda dari RFC3339
	// bawaan time.Time.String(). Gagal parse bukan error fatal - ServerDate
	// cuma tetap zero value.
	if dateHeader := resp.Header.Get("Date"); dateHeader != "" {
		if t, err := time.Parse(time.RFC1123, dateHeader); err == nil {
			result.ServerDate = t
		}
	}
	return result
}
