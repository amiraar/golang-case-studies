package task

import "errors"

// Sentinel error (A.37): error = nilai biasa, bukan exception.
// Pakai errors.Is(err, ErrX) untuk cek jenis error, bukan cocok teks pesan.
// Dipakai di: store.go (dibuat), main.go (ditampilkan ke user).
var (
	ErrTaskNotFound    = errors.New("task not found")
	ErrEmptyTitle      = errors.New("title must not be empty")
	ErrInvalidPriority = errors.New("invalid priority")
)
