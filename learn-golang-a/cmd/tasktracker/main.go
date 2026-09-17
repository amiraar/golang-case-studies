// Package main: entry point CLI. Cuma baca argumen -> panggil package task
// (logika sebenarnya) -> cetak hasil/error. Logika inti ada di internal/task
// supaya bisa dites langsung tanpa lewat CLI.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"learn-golang-a/internal/task"
)

const defaultFile = "tasks.json"

// main() cuma manggil run() lalu os.Exit di paling luar. Kalau os.Exit
// dipanggil di tengah (bukan di sini), defer yang belum sempat jalan
// (mis. f.Close() di storage.go) bisa terlewat.
func main() {
	os.Exit(run(os.Args[1:]))
}

// exitCode = named return, dipakai supaya defer bisa mengubah nilainya.
// defer+recover (A.37) = jaring pengaman terakhir untuk panic tak terduga,
// bukan pengganti error handling biasa di cmdXxx bawah.
func run(args []string) (exitCode int) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "fatal: %v\n", r)
			exitCode = 1
		}
	}()

	if len(args) == 0 {
		printUsage()
		return 1
	}

	// args[0] = nama command, rest = argumen sisanya untuk flag.NewFlagSet
	// masing-masing cmdXxx (A.13 switch-case string).
	cmd, rest := args[0], args[1:]

	var err error
	switch cmd {
	case "add":
		err = cmdAdd(rest)
	case "list":
		err = cmdList(rest)
	case "update":
		err = cmdUpdate(rest)
	case "delete":
		err = cmdDelete(rest)
	case "show":
		err = cmdShow(rest)
	case "scan":
		err = cmdScan(rest)
	default:
		printUsage()
		return 1
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	return 0
}

func printUsage() {
	fmt.Println(`usage: tasktracker <command> [flags]

commands:
  add     -title T [-priority low|medium|high] [-file path]
  list    [-status pending|in_progress|done] [-priority low|medium|high] [-file path]
  update  -id N [-title T] [-status S] [-priority P] [-file path]
  delete  -id N [-file path]
  show    -id N [-file path]
  scan    -files a.json,b.json,c.json [-timeout-ms N]`)
}

// Alur tiap cmdXxx di bawah: parse flag -> LoadFromFile (baca state dari
// disk) -> panggil method Store (logika sebenarnya) -> SaveToFile kalau
// ada perubahan -> cetak hasil. Disk selalu jadi sumber kebenaran,
// dibaca ulang tiap command dijalankan - main.go tak menyimpan state sendiri.

func cmdAdd(args []string) error {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	file := fs.String("file", defaultFile, "task file path")
	title := fs.String("title", "", "task title")
	priorityStr := fs.String("priority", "medium", "low|medium|high")
	fs.Parse(args)

	// String flag dikonversi ke task.Priority lewat ParsePriority (A.43).
	priority, err := task.ParsePriority(*priorityStr)
	if err != nil {
		return err
	}

	s, err := task.LoadFromFile(*file)
	if err != nil {
		return err
	}

	t, err := s.Add(*title, priority)
	if err != nil {
		return err
	}

	if err := s.SaveToFile(*file); err != nil {
		return err
	}

	fmt.Println("added:", t) // dicetak lewat Task.String() otomatis
	return nil
}

func cmdList(args []string) error {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	file := fs.String("file", defaultFile, "task file path")
	statusStr := fs.String("status", "", "filter: pending|in_progress|done")
	priorityStr := fs.String("priority", "", "filter: low|medium|high")
	fs.Parse(args)

	s, err := task.LoadFromFile(*file)
	if err != nil {
		return err
	}

	// Tiap flag yang diisi menambah satu closure filter (ByStatus/ByPriority
	// dari store.go). Kalau tak ada flag, filters kosong = tanpa filter.
	var filters []task.FilterFunc
	if *statusStr != "" {
		filters = append(filters, task.ByStatus(task.Status(*statusStr)))
	}
	if *priorityStr != "" {
		p, err := task.ParsePriority(*priorityStr)
		if err != nil {
			return err
		}
		filters = append(filters, task.ByPriority(p))
	}

	results := s.List(filters...)
	if len(results) == 0 {
		fmt.Println("no tasks")
		return nil
	}
	for _, t := range results {
		fmt.Println(t)
	}
	return nil
}

func cmdUpdate(args []string) error {
	fs := flag.NewFlagSet("update", flag.ExitOnError)
	file := fs.String("file", defaultFile, "task file path")
	id := fs.Int("id", 0, "task id")
	statusStr := fs.String("status", "", "pending|in_progress|done")
	priorityStr := fs.String("priority", "", "low|medium|high")
	title := fs.String("title", "", "new title")
	fs.Parse(args)

	s, err := task.LoadFromFile(*file)
	if err != nil {
		return err
	}

	// Sama pola dengan filters di cmdList, tapi ini UpdateOption (WithXxx
	// dari store.go): cuma field yang diisi flag yang ikut diubah.
	var opts []task.UpdateOption
	if *statusStr != "" {
		opts = append(opts, task.WithStatus(task.Status(*statusStr)))
	}
	if *priorityStr != "" {
		p, err := task.ParsePriority(*priorityStr)
		if err != nil {
			return err
		}
		opts = append(opts, task.WithPriority(p))
	}
	if *title != "" {
		opts = append(opts, task.WithTitle(*title))
	}

	t, err := s.Update(*id, opts...)
	if err != nil {
		return err
	}

	if err := s.SaveToFile(*file); err != nil {
		return err
	}

	fmt.Println("updated:", t)
	return nil
}

func cmdDelete(args []string) error {
	fs := flag.NewFlagSet("delete", flag.ExitOnError)
	file := fs.String("file", defaultFile, "task file path")
	id := fs.Int("id", 0, "task id")
	fs.Parse(args)

	s, err := task.LoadFromFile(*file)
	if err != nil {
		return err
	}

	if err := s.Delete(*id); err != nil {
		return err
	}

	if err := s.SaveToFile(*file); err != nil {
		return err
	}

	fmt.Println("deleted:", *id)
	return nil
}

func cmdShow(args []string) error {
	fs := flag.NewFlagSet("show", flag.ExitOnError)
	file := fs.String("file", defaultFile, "task file path")
	id := fs.Int("id", 0, "task id")
	fs.Parse(args)

	s, err := task.LoadFromFile(*file)
	if err != nil {
		return err
	}

	t, err := s.Get(*id)
	if err != nil {
		return err
	}

	// task.Dump dipakai (bukan fmt.Println biasa) untuk contoh nyata
	// fungsi ber-parameter `any` (A.28), hasilnya lebih detail dari Task.String().
	fmt.Println(task.Dump(t))
	return nil
}

func cmdScan(args []string) error {
	fs := flag.NewFlagSet("scan", flag.ExitOnError)
	filesStr := fs.String("files", "", "comma-separated file paths")
	timeoutMs := fs.Int("timeout-ms", 0, "optional timeout in milliseconds (0 = no timeout)")
	fs.Parse(args)

	if *filesStr == "" {
		return fmt.Errorf("scan: -files is required")
	}
	paths := strings.Split(*filesStr, ",")

	// Pilih versi timeout atau tidak (keduanya di parallel.go, sama-sama paralel).
	var results []task.ScanResult
	if *timeoutMs > 0 {
		r, err := task.ScanFilesWithTimeout(paths, time.Duration(*timeoutMs)*time.Millisecond)
		results = r
		if err != nil {
			// Timeout bukan fatal: hasil yang sempat masuk tetap ditampilkan.
			fmt.Fprintln(os.Stderr, "warning:", err)
		}
	} else {
		results = task.ScanFiles(paths)
	}

	total := 0
	for _, r := range results {
		if r.Err != nil {
			fmt.Printf("%s: error: %v\n", r.Path, r.Err)
			continue
		}
		fmt.Printf("%s: %d task(s)\n", r.Path, len(r.Tasks))
		for _, t := range r.Tasks {
			fmt.Println(" ", t)
		}
		total += len(r.Tasks)
	}
	fmt.Println("total:", total)

	if dups := task.DuplicateTitles(results); len(dups) > 0 {
		fmt.Println("duplicate titles across files:", strings.Join(dups, ", "))
	}
	return nil
}
