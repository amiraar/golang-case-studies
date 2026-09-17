package checker

import (
	"errors"
	"sync"
	"testing"
)

func TestStats_Add_ConcurrentSafe(t *testing.T) {
	stats := &Stats{}
	const goroutines = 50
	errBoom := errors.New("boom")

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(i int) {
			defer wg.Done()
			if i%2 == 0 {
				stats.Add(CheckResult{})
			} else {
				stats.Add(CheckResult{Err: errBoom})
			}
		}(i)
	}
	wg.Wait()

	total, ok, failed := stats.Snapshot()
	if total != goroutines {
		t.Errorf("total = %d, want %d", total, goroutines)
	}
	if ok != goroutines/2 || failed != goroutines/2 {
		t.Errorf("ok=%d failed=%d, want %d each", ok, failed, goroutines/2)
	}
}
