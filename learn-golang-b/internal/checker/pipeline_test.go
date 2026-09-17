package checker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPipeline_GenerateCheckMerge(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	entries := []URLEntry{
		{Label: "a", URL: srv.URL},
		{Label: "b", URL: srv.URL},
		{Label: "c", URL: srv.URL},
		{Label: "d", URL: srv.URL},
	}

	ctx := context.Background()
	urlsCh := GenerateStage(ctx, entries)

	c := NewChecker()
	const workers = 3
	stageOuts := make([]<-chan CheckResult, workers)
	for i := 0; i < workers; i++ {
		stageOuts[i] = c.CheckStage(ctx, urlsCh, time.Second)
	}

	merged := MergeStage(ctx, stageOuts...)

	got := make(map[string]bool, len(entries))
	for r := range merged {
		if !r.OK() {
			t.Errorf("%s: unexpected error %v", r.Entry.Name(), r.Err)
		}
		got[r.Entry.Label] = true
	}

	for _, e := range entries {
		if !got[e.Label] {
			t.Errorf("missing result for %q", e.Label)
		}
	}
}

func TestPipeline_CancellationStopsEarly(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	entries := make([]URLEntry, 20)
	for i := range entries {
		entries[i] = URLEntry{URL: srv.URL}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
	defer cancel()

	urlsCh := GenerateStage(ctx, entries)
	c := NewChecker()
	stageOuts := []<-chan CheckResult{
		c.CheckStage(ctx, urlsCh, time.Second),
		c.CheckStage(ctx, urlsCh, time.Second),
	}
	merged := MergeStage(ctx, stageOuts...)

	count := 0
	for range merged {
		count++
	}

	if count >= len(entries) {
		t.Fatalf("got %d results, want fewer than %d (ctx cancellation should stop the pipeline early)", count, len(entries))
	}
}
