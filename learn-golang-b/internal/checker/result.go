package checker

import (
	"fmt"
	"time"
)

// CheckResult (A.24 struct): hasil satu pengecekan URL.
type CheckResult struct {
	Entry      URLEntry
	StatusCode int
	Duration   time.Duration // A.42 time duration: lama request, bukan cuma timestamp
	ServerDate time.Time     // A.40 parsing time: hasil parse header Date response, zero value kalau tak ada/gagal parse
	CheckedAt  time.Time     // A.40 format time: kapan request ini mulai dikirim
	Err        error
}

// OK: value receiver (cuma baca) karena tak perlu ubah struct (A.25).
func (r CheckResult) OK() bool {
	return r.Err == nil
}

// String (A.38 format string): %-24s/%-8s dst = rata kiri lebar tetap
// biar kolom-kolomnya sejajar di terminal. fmt.Println(result) otomatis
// panggil ini lewat fmt.Stringer.
func (r CheckResult) String() string {
	status := "ERROR"
	if r.OK() {
		status = fmt.Sprintf("%d", r.StatusCode)
	}

	line := fmt.Sprintf("%-24s %-6s %8s  %s",
		r.Entry.Name(), status, r.Duration.Round(time.Millisecond), r.CheckedAt.Format("15:04:05.000"))
	if r.Err != nil {
		line += "  error: " + r.Err.Error()
	}
	return line
}
