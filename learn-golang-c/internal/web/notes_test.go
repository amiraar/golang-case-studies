package web

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"learn-golang-c/internal/store"
)

// renderer di test pakai folder views/ asli (../../views dari package ini)
// supaya test benar-benar memverifikasi template yang dipakai server,
// bukan template tiruan - konsisten dengan semangat B.23 (test lewat
// httptest, tanpa server sungguhan, tapi tetap pakai kode produksi apa adanya).
func testRenderer() *Renderer {
	return NewRenderer("../../views")
}

// newTestNoteStore: NoteStore (A.56, SQLite) di atas DB in-memory - test
// handler di package ini tidak butuh state lintas-test, jadi tiap test
// cukup dapat DB baru yang otomatis lenyap begitu proses selesai.
func newTestNoteStore(t *testing.T) *store.NoteStore {
	t.Helper()
	db, err := store.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	s, err := store.NewNoteStore(db)
	if err != nil {
		t.Fatalf("NewNoteStore: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func multipartNoteBody(t *testing.T, title, body string) (*bytes.Buffer, string) {
	t.Helper()
	buf := &bytes.Buffer{}
	w := multipart.NewWriter(buf)
	w.WriteField("title", title)
	w.WriteField("body", body)
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf, w.FormDataContentType()
}

func TestNotesHandler_ListRenders(t *testing.T) {
	s := newTestNoteStore(t)
	if _, err := s.Create("Judul A", "isi A", nil); err != nil {
		t.Fatalf("Create: %v", err)
	}
	h := NewNotesHandler(s, testRenderer(), t.TempDir(), 1<<20)

	req := httptest.NewRequest(http.MethodGet, "/notes", nil)
	w := httptest.NewRecorder()
	h.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("Judul A")) {
		t.Error("expected response body to contain note title")
	}
}

func TestNotesHandler_Create_RedirectsAndSetsFlash(t *testing.T) {
	s := newTestNoteStore(t)
	h := NewNotesHandler(s, testRenderer(), t.TempDir(), 1<<20)

	body, contentType := multipartNoteBody(t, "Note Baru", "isinya")
	req := httptest.NewRequest(http.MethodPost, "/notes", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()
	h.Create(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303, body=%s", w.Code, w.Body.String())
	}
	if got := w.Header().Get("Location"); got != "/notes" {
		t.Errorf("Location = %q, want /notes", got)
	}
	notes, err := s.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(notes) != 1 {
		t.Fatalf("got %d notes in store, want 1", len(notes))
	}

	cookies := w.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == flashCookieName {
			found = true
		}
	}
	if !found {
		t.Error("expected flash cookie to be set after create")
	}
}

func TestNotesHandler_Create_MissingTitleRejected(t *testing.T) {
	s := newTestNoteStore(t)
	h := NewNotesHandler(s, testRenderer(), t.TempDir(), 1<<20)

	body, contentType := multipartNoteBody(t, "", "isinya")
	req := httptest.NewRequest(http.MethodPost, "/notes", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()
	h.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
	notes, err := s.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(notes) != 0 {
		t.Error("expected no note to be created when title is missing")
	}
}

func TestNotesHandler_Detail_NotFound(t *testing.T) {
	s := newTestNoteStore(t)
	h := NewNotesHandler(s, testRenderer(), t.TempDir(), 1<<20)

	req := httptest.NewRequest(http.MethodGet, "/notes/99", nil)
	req.SetPathValue("id", "99")
	w := httptest.NewRecorder()
	h.Detail(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestNotesHandler_Download(t *testing.T) {
	s := newTestNoteStore(t)
	n, err := s.Create("Judul Unduh", "isi untuk diunduh", nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	h := NewNotesHandler(s, testRenderer(), t.TempDir(), 1<<20)

	req := httptest.NewRequest(http.MethodGet, "/notes/1/download", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	h.Download(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if got := w.Header().Get("Content-Disposition"); got == "" {
		t.Error("expected Content-Disposition header for download")
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(n.Body)) {
		t.Error("expected downloaded content to contain note body")
	}
}

// multiFileBody: bikin body multipart dengan beberapa part field "files"
// (nama boleh sama - dipakai TestNotesHandler_UploadAttachments_DuplicateFilenames
// di bawah), plus satu form field biasa untuk membuktikan part non-file
// dilewati, bukan bikin handler error.
func multiFileBody(t *testing.T, files map[string]string) (*bytes.Buffer, string) {
	t.Helper()
	buf := &bytes.Buffer{}
	w := multipart.NewWriter(buf)
	if err := w.WriteField("note", "bukan file, harus dilewati"); err != nil {
		t.Fatal(err)
	}
	for name, content := range files {
		part, err := w.CreateFormFile("files", name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := part.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf, w.FormDataContentType()
}

func TestNotesHandler_UploadAttachments(t *testing.T) {
	s := newTestNoteStore(t)
	n, err := s.Create("Judul", "isi", nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	uploadDir := t.TempDir()
	h := NewNotesHandler(s, testRenderer(), uploadDir, 1<<20)

	body, contentType := multiFileBody(t, map[string]string{
		"a.txt": "isi a",
		"b.txt": "isi b",
	})
	req := httptest.NewRequest(http.MethodPost, "/notes/"+"1"+"/attachments", body)
	req.SetPathValue("id", "1")
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()
	h.UploadAttachments(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303, body=%s", w.Code, w.Body.String())
	}

	updated, _ := s.Get(n.ID)
	if len(updated.Attachments) != 2 {
		t.Fatalf("got %d attachments recorded, want 2", len(updated.Attachments))
	}
	for _, name := range updated.Attachments {
		if _, err := os.Stat(filepath.Join(uploadDir, name)); err != nil {
			t.Errorf("expected file %q to exist on disk: %v", name, err)
		}
	}
}

// TestNotesHandler_UploadAttachments_DuplicateFilenames (kebijakan yang
// dipilih untuk B.16: prefix unik "{id}-{index}-nama", bukan overwrite
// diam-diam) - dua file bernama sama di satu request tetap tersimpan
// sebagai dua file terpisah di disk.
func TestNotesHandler_UploadAttachments_DuplicateFilenames(t *testing.T) {
	s := newTestNoteStore(t)
	n, err := s.Create("Judul", "isi", nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	uploadDir := t.TempDir()
	h := NewNotesHandler(s, testRenderer(), uploadDir, 1<<20)

	buf := &bytes.Buffer{}
	mw := multipart.NewWriter(buf)
	for _, content := range []string{"isi pertama", "isi kedua"} {
		part, err := mw.CreateFormFile("files", "same.txt")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := part.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/notes/1/attachments", buf)
	req.SetPathValue("id", "1")
	req.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	h.UploadAttachments(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303, body=%s", w.Code, w.Body.String())
	}

	updated, _ := s.Get(n.ID)
	if len(updated.Attachments) != 2 {
		t.Fatalf("got %d attachments, want 2 distinct entries despite same original filename", len(updated.Attachments))
	}
	if updated.Attachments[0] == updated.Attachments[1] {
		t.Errorf("expected two distinct saved names, got same name %q twice", updated.Attachments[0])
	}
	for _, name := range updated.Attachments {
		if _, err := os.Stat(filepath.Join(uploadDir, name)); err != nil {
			t.Errorf("expected file %q to exist on disk: %v", name, err)
		}
	}
}

func TestNotesHandler_UploadAttachments_NoteNotFound(t *testing.T) {
	s := newTestNoteStore(t)
	h := NewNotesHandler(s, testRenderer(), t.TempDir(), 1<<20)

	body, contentType := multiFileBody(t, map[string]string{"a.txt": "isi"})
	req := httptest.NewRequest(http.MethodPost, "/notes/99/attachments", body)
	req.SetPathValue("id", "99")
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()
	h.UploadAttachments(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}
