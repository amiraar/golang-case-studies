// Package main: entry point CLI pipeline. Versi lain dari urlchecker,
// dibangun dari stage eksplisit (internal/checker: GenerateStage ->
// CheckStage (fan-out) -> MergeStage (fan-in)) - A.63-A.65, dipisah dari
// worker pool CheckAll di urlchecker supaya polanya kelihatan jelas.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"learn-golang-b/internal/checker"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) (exitCode int) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "fatal: %v\n", r)
			exitCode = 1
		}
	}()

	fs := flag.NewFlagSet("pipeline", flag.ExitOnError)
	file := fs.String("file", "urls.txt", "path to file with one URL per line")
	workers := fs.Int("workers", 4, "number of fan-out check workers")
	timeout := fs.Duration("timeout", 3*time.Second, "per-request timeout")
	overallTimeout := fs.Duration("overall-timeout", 0, "whole-pipeline deadline, 0 = no deadline")
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

	// ctx tunggal dioper ke ketiga stage (A.65 context cancellation
	// pipeline): begitu di-cancel (overall timeout ATAU proses selesai
	// normal lewat defer cancel()), sinyalnya nyampai ke generator,
	// semua worker fan-out, dan merger sekaligus - tak perlu channel
	// "done" terpisah per stage.
	ctx := context.Background()
	if *overallTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *overallTimeout)
		defer cancel()
	}

	start := time.Now()

	// Stage 1: generate.
	urlsCh := checker.GenerateStage(ctx, entries)

	// Stage 2: fan-out - N goroutine checker rebutan baca urlsCh yang sama.
	c := checker.NewChecker()
	stageOuts := make([]<-chan checker.CheckResult, *workers)
	for i := 0; i < *workers; i++ {
		stageOuts[i] = c.CheckStage(ctx, urlsCh, *timeout)
	}

	// Stage 3: fan-in - gabung semua output worker jadi satu channel.
	merged := checker.MergeStage(ctx, stageOuts...)

	ok, failed := 0, 0
	for r := range merged { // A.34 range: berhenti otomatis begitu MergeStage close(out)-nya
		fmt.Println(r)
		if r.OK() {
			ok++
		} else {
			failed++
		}
	}
	fmt.Printf("\npipeline summary: %d ok, %d failed, %d total, took %s\n",
		ok, failed, ok+failed, time.Since(start).Round(time.Millisecond))

	if ctx.Err() != nil {
		fmt.Fprintln(os.Stderr, "pipeline cancelled:", ctx.Err())
		return 1
	}
	if failed > 0 {
		return 1
	}
	return 0
}
