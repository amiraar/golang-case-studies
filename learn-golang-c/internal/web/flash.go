package web

import (
	"net/http"
	"net/url"
	"time"
)

// Flash message lewat cookie (B.21 HTTP Cookie): pola PRG (B.25 redirect
// setelah POST) butuh cara mengirim pesan sukses/gagal ke halaman TUJUAN
// redirect - request baru itu tidak lagi bawa data POST asli, jadi
// pesannya dititip di cookie berumur pendek, dibaca sekali lalu dihapus.
const flashCookieName = "flash"

// SetFlash: dipanggil notes.go/habits.go tepat sebelum http.Redirect.
func SetFlash(w http.ResponseWriter, message string) {
	http.SetCookie(w, &http.Cookie{
		Name:     flashCookieName,
		Value:    url.QueryEscape(message), // escape: cookie value tak boleh berisi spasi/karakter tertentu apa adanya
		Path:     "/",
		MaxAge:   10, // cukup untuk satu kali render setelah redirect
		HttpOnly: true,
	})
}

// ReadAndClearFlash: dipanggil tiap handler GET yang bisa jadi TUJUAN
// redirect (ListNotes, ListHabits). "Clear" dengan cara set ulang cookie
// yang sama tapi MaxAge -1 + Expires di masa lalu - itu idiom Go untuk
// menghapus cookie (tidak ada http.DeleteCookie).
func ReadAndClearFlash(w http.ResponseWriter, r *http.Request) string {
	c, err := r.Cookie(flashCookieName)
	if err != nil {
		return ""
	}

	http.SetCookie(w, &http.Cookie{
		Name:    flashCookieName,
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),
		MaxAge:  -1,
	})

	msg, err := url.QueryUnescape(c.Value)
	if err != nil {
		return ""
	}
	return msg
}
