package checker

import (
	"context"
	"sync"
	"time"
)

// File ini pola pipeline eksplisit (A.63 pipeline, A.64 fan-in/fan-out,
// A.65 context cancellation pipeline) - tiga stage terpisah, beda dari
// CheckAll di pool.go yang menggabung semuanya jadi satu fungsi worker
// pool. Dipakai cmd/pipeline, BUKAN dipanggil CheckAll - dua cara berbeda
// menyelesaikan masalah yang sama (concurrent URL check), sengaja
// dipertahankan berdampingan sebagai perbandingan.

// GenerateStage (stage 1, A.63 pipeline): ubah slice jadi channel, satu
// nilai dikirim tiap kali konsumen siap nerima. select+ctx.Done() (A.65)
// supaya generator berhenti ngirim begitu ctx dibatalkan - tanpa ini,
// generator bisa ke-block selamanya ngirim ke channel yang sudah tak ada
// pembacanya (goroutine leak).
func GenerateStage(ctx context.Context, entries []URLEntry) <-chan URLEntry {
	out := make(chan URLEntry)
	go func() {
		defer close(out) // A.34 close: sinyal "tidak ada nilai lagi" ke stage berikutnya
		for _, e := range entries {
			select {
			case out <- e:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// CheckStage (stage 2, A.64 fan-out): dipanggil BERKALI-KALI oleh caller
// (lihat cmd/pipeline) dengan channel "in" yang SAMA setiap kali - beberapa
// goroutine rebutan baca dari satu channel input = fan-out. Tiap panggilan
// dapat channel output sendiri-sendiri; digabung lagi di MergeStage.
func (c *Checker) CheckStage(ctx context.Context, in <-chan URLEntry, timeout time.Duration) <-chan CheckResult {
	out := make(chan CheckResult)
	go func() {
		defer close(out)
		for entry := range in { // A.34 range: goroutine ini berhenti sendiri begitu GenerateStage close(out)-nya
			r := c.CheckOne(ctx, entry, timeout)
			select {
			case out <- r:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// MergeStage (stage 3, A.64 fan-in): kebalikan fan-out - N channel input
// digabung jadi 1 channel output. sync.WaitGroup (A.60) menunggu semua
// channel input habis sebelum close(out) (A.34), supaya konsumen di
// belakangnya cukup pakai range biasa tanpa tahu ada berapa worker di depan.
func MergeStage(ctx context.Context, cs ...<-chan CheckResult) <-chan CheckResult {
	out := make(chan CheckResult)
	var wg sync.WaitGroup
	wg.Add(len(cs))

	for _, c := range cs {
		go func(c <-chan CheckResult) {
			defer wg.Done()
			for r := range c {
				select {
				case out <- r:
				case <-ctx.Done():
					return
				}
			}
		}(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
