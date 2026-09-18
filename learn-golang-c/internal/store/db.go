// Package store: db.go buka koneksi SQLite (A.56 SQL). Driver
// modernc.org/sqlite dipakai karena pure-Go (tanpa cgo, tanpa instalasi
// binary SQLite terpisah) - proyek belajar ini harus tetap `go build` di
// mana saja tanpa toolchain C.
package store

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

const noteSchema = `
CREATE TABLE IF NOT EXISTS notes (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT NOT NULL,
	body TEXT NOT NULL,
	attachments TEXT NOT NULL DEFAULT '[]',
	created_at DATETIME NOT NULL,
	updated_at DATETIME NOT NULL
);`

// OpenDB: satu koneksi dipakai untuk NoteStore (satu-satunya entity yang
// pindah dari in-memory map ke SQL - lihat catatan scope di note.go).
// HabitStore tetap in-memory, tidak disentuh.
func OpenDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("buka database %s: %w", path, err)
	}
	if _, err := db.Exec(noteSchema); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrasi schema notes: %w", err)
	}
	return db, nil
}
