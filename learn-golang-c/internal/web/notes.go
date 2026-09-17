package web

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"learn-golang-c/internal/store"
)

// NotesHandler: mengelompokkan semua handler /notes/* jadi satu struct
// supaya dependency (store, renderer, folder upload) dioper sekali lewat
// constructor, bukan parameter global.
type NotesHandler struct {
	store          *store.NoteStore
	render         *Renderer
	uploadDir      string
	maxUploadBytes int64
}

func NewNotesHandler(s *store.NoteStore, rd *Renderer, uploadDir string, maxUploadBytes int64) *NotesHandler {
	return &NotesHandler{store: s, render: rd, uploadDir: uploadDir, maxUploadBytes: maxUploadBytes}
}

// List (B.4/B.9 render): GET /notes. Flash (B.21) dibaca di sini karena
// ini tujuan redirect Create/Delete di bawah.
func (h *NotesHandler) List(w http.ResponseWriter, r *http.Request) {
	data := struct{ Notes []store.Note }{Notes: h.store.List()}
	h.render.Page(w, r, "notes_list.html", data, "note_card.html")
}

// NewForm: GET /notes/new, form kosong (B.4).
func (h *NotesHandler) NewForm(w http.ResponseWriter, r *http.Request) {
	h.render.Page(w, r, "note_form.html", nil)
}

// Create (B.11 POST, B.12 form value, B.13 file upload opsional,
// B.27 body size limit, B.25 redirect PRG): POST /notes.
func (h *NotesHandler) Create(w http.ResponseWriter, r *http.Request) {
	// B.27: batasi ukuran body SEBELUM diparse - urutannya penting, kalau
	// dipasang setelah ParseMultipartForm sudah kebaca duluan tak ada gunanya.
	r.Body = http.MaxBytesReader(w, r.Body, h.maxUploadBytes)

	if err := r.ParseMultipartForm(h.maxUploadBytes); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			http.Error(w, "file terlalu besar", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	title := r.FormValue("title") // B.12
	body := r.FormValue("body")
	if title == "" {
		http.Error(w, "title wajib diisi", http.StatusBadRequest)
		return
	}

	// B.13: lampiran opsional - ErrMissingFile bukan error sungguhan di
	// sini, cuma tanda user tidak pilih file.
	var attachments []string
	file, header, err := r.FormFile("attachment")
	switch {
	case err == nil:
		defer file.Close()
		dst, err := os.Create(filepath.Join(h.uploadDir, header.Filename))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer dst.Close()
		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		attachments = []string{header.Filename}
	case errors.Is(err, http.ErrMissingFile):
		// tidak apa-apa, lanjut tanpa lampiran
	default:
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	n := h.store.Create(title, body, attachments)
	SetFlash(w, fmt.Sprintf("Note %q berhasil dibuat", n.Title))
	// B.25: 303 See Other, bukan 301/302 - browser WAJIB ganti method jadi
	// GET saat redirect, mencegah resubmit form kalau user refresh halaman
	// hasil redirect (pola Post/Redirect/Get).
	http.Redirect(w, r, "/notes", http.StatusSeeOther)
}

// Detail: GET /notes/{id}.
func (h *NotesHandler) Detail(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	n, ok := h.store.Get(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	h.render.Page(w, r, "note_detail.html", n)
}

// Delete: POST /notes/{id}/delete (form biasa tak bisa kirim method
// DELETE tanpa JS, jadi dipakai POST + suffix path, idiom umum server-
// rendered app).
func (h *NotesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if h.store.Delete(id) {
		SetFlash(w, "Note dihapus")
	}
	http.Redirect(w, r, "/notes", http.StatusSeeOther)
}

// Download (B.10 render ke string + B.17 download file): GET
// /notes/{id}/download. Isi file dibangun lewat template text/template
// (render.go), bukan cuma fmt.Sprintf, supaya format-nya konsisten dan
// gampang diubah tanpa menyentuh kode Go.
func (h *NotesHandler) Download(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	n, ok := h.store.Get(id)
	if !ok {
		http.NotFound(w, r)
		return
	}

	content, err := RenderNoteExportText(n)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("note-%d.txt", n.ID)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Write([]byte(content))
}

// UploadAttachments (B.16 Multiple File Upload: MultipartReader): POST
// /notes/{id}/attachments. Sengaja endpoint terpisah dari Create (yang
// pakai ParseMultipartForm+FormFile, B.13) supaya kontras pola single-
// upload-lewat-buffer vs multi-upload-lewat-stream kelihatan jelas di
// struktur kode, bukan dicampur satu handler.
func (h *NotesHandler) UploadAttachments(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if _, ok := h.store.Get(id); !ok {
		http.NotFound(w, r)
		return
	}

	// B.27: sama seperti Create, batas ukuran dipasang SEBELUM body dibaca
	// sama sekali - endpoint ini juga multipart, jadi harus konsisten
	// dengan Create supaya bukan celah bypass limit ukuran.
	r.Body = http.MaxBytesReader(w, r.Body, h.maxUploadBytes)

	reader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var saved []string
	for i := 0; ; i++ {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			var maxErr *http.MaxBytesError
			if errors.As(err, &maxErr) {
				http.Error(w, "file terlalu besar", http.StatusRequestEntityTooLarge)
				return
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Part tanpa FileName() adalah form field biasa (mis. field teks
		// yang nyasar ikut terkirim), bukan file - dilewati, bukan dicoba
		// os.Create("") yang pasti gagal.
		if part.FileName() == "" {
			continue
		}

		// Prefix {id}-{index}- membuat nama file unik per note+urutan
		// upload, walau dua file di request yang sama punya nama identik -
		// os.Create tanpa prefix ini akan overwrite diam-diam.
		saveName := fmt.Sprintf("%d-%d-%s", id, i, part.FileName())
		dst, err := os.Create(filepath.Join(h.uploadDir, saveName))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := io.Copy(dst, part); err != nil {
			dst.Close()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Close eksplisit di sini, bukan defer - defer di dalam loop baru
		// jalan setelah handler selesai, menahan file descriptor tetap
		// terbuka selama seluruh upload berlangsung.
		dst.Close()
		saved = append(saved, saveName)
	}

	h.store.AddAttachments(id, saved)
	SetFlash(w, fmt.Sprintf("%d lampiran berhasil diunggah", len(saved)))
	http.Redirect(w, r, fmt.Sprintf("/notes/%d", id), http.StatusSeeOther)
}
