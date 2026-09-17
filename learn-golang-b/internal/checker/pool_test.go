package checker

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestCheckAll_AllOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	entries := []URLEntry{
		{Label: "a", URL: srv.URL},
		{Label: "b", URL: srv.URL},
		{Label: "c", URL: srv.URL},
	}
	opts := Options{Concurrency: 2, PerRequestTimeout: time.Second}

	results, stats := CheckAll(entries, opts, nil)
	if len(results) != len(entries) {
		t.Fatalf("got %d results, want %d", len(results), len(entries))
	}
	for _, r := range results {
		if !r.OK() {
			t.Errorf("%s: unexpected error %v", r.Entry.Name(), r.Err)
		}
	}

	// Stats (A.61 mutex) dan results (channel) harus sepakat - dua cara
	// berbeda menghitung hal yang sama.
	total, ok, failed := stats.Snapshot()
	if total != len(entries) || ok != len(entries) || failed != 0 {
		t.Errorf("stats.Snapshot() = (%d, %d, %d), want (%d, %d, 0)", total, ok, failed, len(entries), len(entries))
	}
}

func TestCheckAll_ProgressCallback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	entries := []URLEntry{{Label: "a", URL: srv.URL}, {Label: "b", URL: srv.URL}}
	opts := Options{Concurrency: 2, PerRequestTimeout: time.Second}

	calls := 0
	CheckAll(entries, opts, func(done, total int) {
		calls++
		if total != len(entries) {
			t.Errorf("progress total = %d, want %d", total, len(entries))
		}
	})
	if calls == 0 {
		t.Error("progress callback was never called")
	}
}

func TestCheckAll_OverallTimeoutReturnsPartialResults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// URL beda-beda (bukan srv.URL yang sama diulang) supaya result cache
	// (A.61) tak ikut men-dedupe entry-entry ini - test ini murni menguji
	// overall timeout, bukan cache.
	entries := []URLEntry{
		{URL: srv.URL + "/1"}, {URL: srv.URL + "/2"}, {URL: srv.URL + "/3"}, {URL: srv.URL + "/4"},
	}
	opts := Options{
		Concurrency:       1, // satu worker + 4 entry lambat -> overall timeout pasti kepotong sebelum semua selesai
		PerRequestTimeout: time.Second,
		OverallTimeout:    150 * time.Millisecond,
	}

	results, _ := CheckAll(entries, opts, nil)
	if len(results) >= len(entries) {
		t.Fatalf("got %d results, want fewer than %d (overall timeout should cut the scan short)", len(results), len(entries))
	}
}

func TestCheckAll_DefaultsConcurrencyToOne(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	entries := []URLEntry{{URL: srv.URL}}
	opts := Options{Concurrency: 0, PerRequestTimeout: time.Second}

	results, _ := CheckAll(entries, opts, nil)
	if len(results) != 1 || !results[0].OK() {
		t.Fatalf("unexpected results with Concurrency=0: %+v", results)
	}
}

func TestCheckAll_DuplicateURLsHitServerOnce(t *testing.T) {
	var hits int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&hits, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	entries := []URLEntry{
		{Label: "first", URL: srv.URL},
		{Label: "second", URL: srv.URL}, // sama persis URL-nya, cuma label beda
		{Label: "third", URL: srv.URL},
	}
	// Concurrency: 1 - satu worker memproses ketiganya berurutan, jadi
	// entry ke-2/3 PASTI ketemu entry pertama sudah selesai & ke-cache.
	// Dengan concurrency > 1, tiga worker bisa mulai bersamaan sebelum
	// ada yang sempat mengisi cache (thundering herd) - itu skenario beda
	// yang butuh singleflight (C.37), di luar cakupan mutex sederhana ini.
	opts := Options{Concurrency: 1, PerRequestTimeout: time.Second}

	results, _ := CheckAll(entries, opts, nil)
	if len(results) != len(entries) {
		t.Fatalf("got %d results, want %d", len(results), len(entries))
	}
	if got := atomic.LoadInt64(&hits); got != 1 {
		t.Errorf("server got %d hits, want 1 (result cache should dedupe repeated URLs)", got)
	}

	// Label tiap result harus tetap sesuai entry aslinya walau hasilnya dari cache.
	gotLabels := make(map[string]bool, len(results))
	for _, r := range results {
		gotLabels[r.Entry.Label] = true
	}
	for _, want := range []string{"first", "second", "third"} {
		if !gotLabels[want] {
			t.Errorf("missing result for label %q", want)
		}
	}
}

func TestRandomJitter(t *testing.T) {
	if got := randomJitter(0); got != 0 {
		t.Errorf("randomJitter(0) = %v, want 0", got)
	}
	for i := 0; i < 20; i++ {
		if got := randomJitter(10 * time.Millisecond); got < 0 || got >= 10*time.Millisecond {
			t.Fatalf("randomJitter(10ms) = %v, out of range [0, 10ms)", got)
		}
	}
}
