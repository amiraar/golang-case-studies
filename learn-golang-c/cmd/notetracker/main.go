// Package main: entry point web service "Habit/Note Tracker" (Project 3,
// menutup B penuh). Tugasnya cuma wiring - baca config, bikin store,
// daftarkan route+middleware, jalankan server dengan graceful shutdown.
// Logika sesungguhnya ada di internal/store dan internal/web.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"learn-golang-c/internal/config"
	"learn-golang-c/internal/store"
	"learn-golang-c/internal/web"
)

func main() {
	configPath := flag.String("config", "config.json", "path to config.json")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	if err := os.MkdirAll(cfg.Upload.Dir, 0755); err != nil {
		log.Fatal(err)
	}

	db, err := store.OpenDB(cfg.Database.Path) // A.56: NoteStore satu-satunya entity yang dipindah ke SQL
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	noteStore, err := store.NewNoteStore(db)
	if err != nil {
		log.Fatal(err)
	}
	defer noteStore.Close()

	habitStore := store.NewHabitStore()
	renderer := web.NewRenderer(cfg.ViewsDir)

	notesHandler := web.NewNotesHandler(noteStore, renderer, cfg.Upload.Dir, cfg.Upload.MaxBytes)
	habitsHandler := web.NewHabitsHandler(habitStore, renderer)
	apiHandler := web.NewAPIHandler(habitStore)
	statsHandler := web.NewHabitsStatsHandler(habitStore, renderer, 300*time.Millisecond)

	auth := web.BasicAuthMiddleware(cfg.Auth.Username, cfg.Auth.Password) // B.18

	mux := web.NewCustomMux() // B.20
	// B.19: Recover di-Use PALING TERAKHIR supaya jadi lapisan TERLUAR
	// (lihat mux.go ServeHTTP) - panic di LoggingMiddleware sendiri pun
	// masih tertangkap.
	mux.Use(web.LoggingMiddleware)
	mux.Use(web.RecoverMiddleware)

	// B.3: static assets. /uploads/ disajikan dengan cara yang sama supaya
	// lampiran note (B.13) bisa dibuka langsung dari note_detail.html.
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir(cfg.StaticDir))))
	mux.Handle("GET /uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(cfg.Upload.Dir))))

	// B.1/B.2: "/" tanpa method = subtree pattern, menangkap SEMUA path
	// yang tak ke-match pattern lain (persis idiom tutorial B.26) -
	// Dashboard() sendiri yang membedakan path root asli vs 404 lewat
	// pengecekan r.URL.Path.
	mux.HandleFunc("/", habitsHandler.Dashboard(noteStore))

	// Notes: baca publik, tulis (new form/create/delete) diproteksi B.18.
	mux.HandleFunc("GET /notes", notesHandler.List)
	mux.Handle("GET /notes/new", auth(http.HandlerFunc(notesHandler.NewForm)))
	mux.Handle("POST /notes", auth(http.HandlerFunc(notesHandler.Create)))
	mux.HandleFunc("GET /notes/{id}", notesHandler.Detail)
	mux.Handle("POST /notes/{id}/delete", auth(http.HandlerFunc(notesHandler.Delete)))
	mux.HandleFunc("GET /notes/{id}/download", notesHandler.Download)
	mux.Handle("POST /notes/{id}/attachments", auth(http.HandlerFunc(notesHandler.UploadAttachments)))

	// Habits: pola sama seperti notes.
	mux.HandleFunc("GET /habits", habitsHandler.List)
	mux.Handle("GET /habits/new", auth(http.HandlerFunc(habitsHandler.NewForm)))
	mux.Handle("POST /habits", auth(http.HandlerFunc(habitsHandler.Create)))
	mux.Handle("POST /habits/{id}/delete", auth(http.HandlerFunc(habitsHandler.Delete)))
	mux.Handle("POST /habits/{id}/checkin", auth(http.HandlerFunc(habitsHandler.Checkin)))

	// B.28: http.TimeoutHandler membungkus PALING LUAR (dicek duluan
	// dibanding auth) - kalau computasinya lewat 3 detik, client dapat 503
	// walau sebenarnya sudah lolos auth.
	mux.Handle("GET /habits/stats",
		http.TimeoutHandler(auth(http.HandlerFunc(statsHandler.Stats)), 3*time.Second, "stats timeout"))

	// B.14/B.15: API JSON murni, terpisah dari handler form di atas.
	mux.HandleFunc("GET /api/habits", apiHandler.ListHabits)
	mux.Handle("POST /api/habits", auth(http.HandlerFunc(apiHandler.CreateHabit)))

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      mux,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	// B.24 graceful shutdown: server jalan di goroutine terpisah supaya
	// goroutine main bebas menunggu sinyal OS (Ctrl+C / SIGTERM).
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("notetracker listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	log.Println("shutdown signal received, draining in-flight requests...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal(err)
	}
	log.Println("server stopped")
}
