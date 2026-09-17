package web

import (
	"bytes"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
	texttemplate "text/template"
	"time"
)

// templateFuncs (B.8 Template: Custom Functions): dipetakan lewat
// template.FuncMap, dipanggil di dalam file .html pakai {{formatDate .X}}
// dst (B.7 built-in functions seperti len/eq dipakai langsung di view,
// tidak perlu didaftarkan manual).
var templateFuncs = template.FuncMap{
	"formatDate": func(t time.Time) string { return t.Format("02 Jan 2006 15:04") },
	"truncate": func(s string, n int) string {
		if len(s) <= n {
			return s
		}
		return s[:n] + "..."
	},
}

// Renderer: menyimpan path folder views/ supaya handler tidak hardcode
// "views/..." berulang-ulang. Dibuat sekali di main.go, dioper ke tiap
// handler struct (notes.go/habits.go).
type Renderer struct {
	dir string
}

func NewRenderer(dir string) *Renderer {
	return &Renderer{dir: dir}
}

func (rd *Renderer) path(name string) string {
	return filepath.Join(rd.dir, filepath.FromSlash(name))
}

// pageData: layout.html cuma butuh dua hal - Flash (B.21, boleh kosong di
// halaman mana pun) dan Body (data asli tiap page, bebas bentuknya).
// Dibungkus di sini (bukan tiap handler nambahin field Flash sendiri-
// sendiri) supaya {{if .Flash}} di layout.html aman dieksekusi untuk
// SEMUA page - html/template error keras kalau field itu tak ada sama
// sekali di tipe data yang dioper (beda dari map yang field hilang = zero
// value, struct field hilang = execution error).
type pageData struct {
	Flash string
	Body  any
}

// Page (B.4 Render HTML + B.9 Render Specific HTML Template): tiap
// pemanggilan HANYA mem-parse layout.html + file halaman yang relevan +
// partial yang dibutuhkan (bukan seluruh folder views/ sekaligus) - kalau
// semua page diparse bareng, tiap file yang punya {{define "content"}}
// akan tabrakan nama (B.5 partial butuh nama unik per fragment, tapi
// "content" sengaja dipakai ulang di tiap page.html karena hanya satu yang
// diparse per request). layout.html jadi root template (dieksekusi lewat
// Execute, bukan ExecuteTemplate) supaya {{template "content" .Body}} di
// dalamnya otomatis mengambil definisi dari file page yang baru diparse.
func (rd *Renderer) Page(w http.ResponseWriter, r *http.Request, page string, body any, partials ...string) {
	files := []string{rd.path("layout.html"), rd.path(page)}
	for _, p := range partials {
		files = append(files, rd.path(filepath.Join("partials", p)))
	}

	tmpl, err := template.New("layout.html").Funcs(templateFuncs).ParseFiles(files...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// B.21: flash dibaca (dan dihapus) DI SINI, sekali, untuk semua page -
	// handler individual tak perlu ingat memanggil ReadAndClearFlash sendiri.
	data := pageData{Flash: ReadAndClearFlash(w, r), Body: body}
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// NoteExportText (B.10 Render HTML String): merender ke string/buffer,
// BUKAN langsung ke ResponseWriter, karena hasilnya dipakai notes.go
// sebagai isi file download (B.17), bukan halaman. Pakai text/template
// (bukan html/template seperti Page di atas) karena keluarannya .txt polos
// - html/template akan meng-escape karakter seperti & < > yang justru
// tidak diinginkan di file teks biasa.
var noteExportTmpl = texttemplate.Must(texttemplate.New("note_export").Funcs(texttemplate.FuncMap{
	"join": strings.Join,
}).Parse(
	strings.TrimSpace(`
Judul   : {{.Title}}
Dibuat  : {{.CreatedAt.Format "02 Jan 2006 15:04"}}
Diubah  : {{.UpdatedAt.Format "02 Jan 2006 15:04"}}
{{if .Attachments}}Lampiran: {{join .Attachments ", "}}
{{end}}
{{.Body}}
`) + "\n"))

func RenderNoteExportText(data any) (string, error) {
	var buf bytes.Buffer
	if err := noteExportTmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
