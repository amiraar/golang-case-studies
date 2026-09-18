// Package main: entry point Project 4 - rebuild Project 3 (Habit/Note
// Tracker) di atas Echo (C block). Tugasnya cuma wiring, sama seperti
// cmd/notetracker/main.go di learn-golang-c - logika ada di internal/.
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"learn-golang-d/internal/config"
	"learn-golang-d/internal/store"
	"learn-golang-d/internal/web"
)

func main() {
	configDir := flag.String("config-dir", "configs", "folder berisi config.json")
	flag.Parse()

	cfg, err := config.Load(*configDir, "config")
	if err != nil {
		log.Fatal(err)
	}

	noteStore := store.NewNoteStore()
	habitStore := store.NewHabitStore()

	notesHandler := web.NewNotesHandler(noteStore)
	habitsHandler := web.NewHabitsHandler(habitStore)
	authHandler := web.NewAuthHandler(cfg.Auth.Username, cfg.Auth.Password, cfg.JWT.SigningKey, cfg.JWT.ExpiryMinutes)

	e := echo.New()
	e.Validator = web.NewCustomValidator() // C.5
	e.HTTPErrorHandler = web.ErrorHandler  // C.6
	e.Use(middleware.Recover())            // panic recovery bawaan Echo
	e.Use(web.LoggingMiddleware)           // C.8, Logrus custom middleware

	web.RegisterRoutes(e, notesHandler, habitsHandler, authHandler, cfg.JWT.SigningKey)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("%s listening on %s", cfg.AppName, addr)
	e.Logger.Fatal(e.Start(addr))
}
