// Package main: entry point CLI urlchecker. Baca daftar URL dari file,
// cek semuanya paralel (internal/checker), cetak hasil + ringkasan.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"learn-golang-b/internal/checker"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

// exitCode = named return supaya defer di bawah bisa mengubahnya (sama
// pola dengan learn-golang-a: A.37 defer+recover jadi jaring pengaman
// terakhir, bukan pengganti error handling biasa).
func run(args []string) (exitCode int) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "fatal: %v\n", r)
			exitCode = 1
		}
	}()

	fs := flag.NewFlagSet("urlchecker", flag.ExitOnError)
	file := fs.String("file", "urls.txt", "path to file with one URL per line (\"url\" or \"label,url\", # for comments)")
	concurrency := fs.Int("concurrency", 5, "number of concurrent workers")
	perTimeout := fs.Duration("timeout", 3*time.Second, "per-request timeout, e.g. 2s, 500ms")
	overallTimeout := fs.Duration("overall-timeout", 0, "whole-scan deadline, 0 = no deadline")
	jitter := fs.Duration("jitter", 0, "max random delay added before each request, e.g. 200ms")
	progressEvery := fs.Duration("progress-every", 2*time.Second, "how often to print a progress line while scanning, 0 = disabled")
	fs.Parse(args)

	entries, err := checker.LoadURLsFromFile(*file)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	if len(entries) == 0 {
		fmt.Fprintln(os.Stderr, "no URLs found in", *file)
		return 1
	}

	opts := checker.Options{
		Concurrency:       *concurrency,
		PerRequestTimeout: *perTimeout,
		OverallTimeout:    *overallTimeout,
		MaxJitter:         *jitter,
		ProgressEvery:     *progressEvery,
	}

	start := time.Now()
	results, stats := checker.CheckAll(entries, opts, func(done, total int) {
		fmt.Fprintf(os.Stderr, "progress: %d/%d checked (%s elapsed)\n", done, total, time.Since(start).Round(time.Millisecond))
	})

	ok, failed := 0, 0
	for _, r := range results {
		fmt.Println(r)
		if r.OK() {
			ok++
		} else {
			failed++
		}
	}
	fmt.Printf("\nsummary: %d ok, %d failed, %d/%d total, took %s\n",
		ok, failed, len(results), len(entries), time.Since(start).Round(time.Millisecond))

	// stats dihitung terpisah dari "results" di atas (channel vs sync.Mutex,
	// lihat pool.go) - dicetak juga sebagai cross-check dua pendekatan
	// concurrency itu selalu berakhir di angka yang sama.
	statTotal, statOK, statFailed := stats.Snapshot()
	fmt.Printf("stats (sync.Mutex cross-check): %d ok, %d failed, %d total\n", statOK, statFailed, statTotal)

	if failed > 0 {
		return 1
	}
	return 0
}
