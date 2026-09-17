package checker

import "sync"

// resultCache (A.61 sync.Mutex): cache hasil per-URL supaya kalau urls.txt
// punya baris URL yang sama lebih dari sekali, worker cukup request sekali
// dan sisanya pakai hasil yang sudah ada. map BUKAN tipe yang aman diakses
// bersamaan dari banyak goroutine tanpa proteksi - tanpa mutex ini,
// "concurrent map read and map write" bisa bikin program crash, bukan
// cuma salah hasil seperti Stats di atas.
type resultCache struct {
	mu    sync.Mutex
	cache map[string]CheckResult
}

func newResultCache() *resultCache {
	return &resultCache{cache: make(map[string]CheckResult)}
}

func (c *resultCache) get(url string) (CheckResult, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	r, ok := c.cache[url]
	return r, ok
}

func (c *resultCache) set(url string, r CheckResult) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[url] = r
}
