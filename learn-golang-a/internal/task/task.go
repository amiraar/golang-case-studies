package task

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Status: tipe string (A.10) karena disimpan apa adanya ke JSON dan flag CLI,
// tidak butuh tabel konversi angka<->nama seperti Priority di bawah.
type Status string

// A.11 konstanta: enum ala Go.
const (
	StatusPending    Status = "pending"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

// Value receiver karena cuma baca, tidak ubah data (A.25).
// switch tanpa kondisi (A.13) lebih ringkas dari if/else berantai.
func (s Status) Valid() bool {
	switch s {
	case StatusPending, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}

// Priority: tipe int+iota (beda gaya dari Status) karena punya urutan
// alami (low < medium < high) yang berguna untuk dibandingkan/diurutkan.
type Priority int

const (
	PriorityLow    Priority = iota // 0
	PriorityMedium                 // 1
	PriorityHigh                   // 2
)

// A.15 array, bukan map: index array = nilai Priority itu sendiri (0,1,2),
// jadi lookup langsung tanpa hashing, dan ukurannya memang tetap 3.
var priorityNames = [3]string{"low", "medium", "high"}

// String() bikin Priority otomatis kompatibel fmt.Stringer (A.27, implisit).
// Efeknya: fmt.Println/%s manggil ini otomatis, tanpa dipanggil manual.
func (p Priority) String() string {
	if p < 0 || int(p) >= len(priorityNames) {
		return "unknown"
	}
	return priorityNames[p]
}

// ParsePriority: kebalikan String(), dipanggil main.go tiap flag -priority
// diproses. Coba nama dulu (switch), lalu fallback ke angka via
// strconv.Atoi + cast eksplisit int->Priority (A.43, Go tak punya konversi implisit).
func ParsePriority(s string) (Priority, error) {
	switch strings.ToLower(s) {
	case "low":
		return PriorityLow, nil
	case "medium":
		return PriorityMedium, nil
	case "high":
		return PriorityHigh, nil
	}

	if n, err := strconv.Atoi(s); err == nil {
		p := Priority(n) // A.43 konversi eksplisit
		if p < PriorityLow || p > PriorityHigh {
			return 0, fmt.Errorf("%w: %q", ErrInvalidPriority, s)
		}
		return p, nil
	}

	// %w membungkus error asli supaya errors.Is masih bisa dipakai caller.
	return 0, fmt.Errorf("%w: %q", ErrInvalidPriority, s)
}

// Task (A.24 struct). Tag `json:"..."` dibaca encoding/json (A.53) di
// storage.go untuk kontrol nama key di file tasks.json.
type Task struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Status    Status    `json:"status"`
	Priority  Priority  `json:"priority"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// A.38 format string: %-3d/%-11s/%-6s = rata kiri lebar tetap, biar rapi
// sejajar di terminal. Dipanggil otomatis oleh fmt.Println di main.go.
func (t Task) String() string {
	return fmt.Sprintf("#%-3d [%-11s] (%-6s) %s", t.ID, t.Status, t.Priority, t.Title)
}

// touch/markStatus unexported (A.26): cuma dipakai dalam package ini,
// lewat store.go WithStatus, supaya UpdatedAt selalu ikut ter-update.
// Pointer receiver (beda dari String/Valid di atas) karena method ini
// MENGUBAH field struct - value receiver cuma akan ubah salinan lokal.
func (t *Task) touch() {
	t.UpdatedAt = time.Now()
}

func (t *Task) markStatus(s Status) {
	t.Status = s
	t.touch()
}

// Dump pakai `any` (A.28) karena harus terima beberapa tipe berbeda
// (Task, []Task, dst) tanpa bikin fungsi terpisah per tipe.
// Type switch (A.13+A.28) mengecek tipe asli di balik interface{}.
// Dipanggil dari main.go cmdShow.
func Dump(v any) string {
	switch val := v.(type) {
	case Task:
		return fmt.Sprintf("Task{ID:%d, Title:%q, Status:%s, Priority:%s, CreatedAt:%s}",
			val.ID, val.Title, val.Status, val.Priority, val.CreatedAt.Format(time.RFC3339))
	case []Task:
		var b strings.Builder
		for _, t := range val {
			b.WriteString(t.String())
			b.WriteByte('\n')
		}
		return b.String()
	default:
		return fmt.Sprintf("%v", val)
	}
}
