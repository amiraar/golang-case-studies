package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Note: domain "habit/note tracker" - catatan bebas, opsional ada lampiran
// file (diisi lewat B.13 file upload dan B.16 multi-upload di
// internal/web/notes.go).
type Note struct {
	ID          int
	Title       string
	Body        string
	Attachments []string // nama file di bawah Upload.Dir, nil/kosong = tanpa lampiran
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ErrNotFound: dikembalikan Get/AddAttachments ketika id tidak ada baris -
// membungkus sql.ErrNoRows supaya caller di package web tidak perlu tahu
// detail driver SQL, cukup errors.Is(err, store.ErrNotFound).
var ErrNotFound = errors.New("note not found")

// NoteStore: A.56 SQL - satu-satunya entity yang dipindah dari in-memory
// map (pola HabitStore) ke database/sql+SQLite. Attachments tetap kolom
// tunggal (JSON-encoded TEXT), bukan tabel terpisah - keputusan yang sama
// dengan waktu masih in-memory (B.16): normalisasi ke tabel relasi baru
// masuk akal kalau ada query yang butuh JOIN per-attachment, yang sampai
// sekarang tidak ada.
type NoteStore struct {
	db      *sql.DB
	getStmt *sql.Stmt // contoh idiom prepared statement A.56 - query paling sering dipanggil (Get per-request)
}

func NewNoteStore(db *sql.DB) (*NoteStore, error) {
	stmt, err := db.Prepare(`SELECT id, title, body, attachments, created_at, updated_at FROM notes WHERE id = ?`)
	if err != nil {
		return nil, fmt.Errorf("prepare get statement: %w", err)
	}
	return &NoteStore{db: db, getStmt: stmt}, nil
}

// Close: tutup prepared statement. Koneksi *sql.DB sendiri dipegang dan
// ditutup oleh main.go (dipakai bareng store lain andai suatu saat perlu).
func (s *NoteStore) Close() error {
	return s.getStmt.Close()
}

func encodeAttachments(a []string) (string, error) {
	if a == nil {
		a = []string{}
	}
	b, err := json.Marshal(a)
	return string(b), err
}

func decodeAttachments(s string) ([]string, error) {
	var a []string
	if err := json.Unmarshal([]byte(s), &a); err != nil {
		return nil, err
	}
	return a, nil
}

// Create: dipanggil handler POST /notes (notes.go) setelah form di-parse.
func (s *NoteStore) Create(title, body string, attachments []string) (Note, error) {
	attJSON, err := encodeAttachments(attachments)
	if err != nil {
		return Note{}, err
	}

	now := time.Now()
	res, err := s.db.Exec(
		`INSERT INTO notes (title, body, attachments, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		title, body, attJSON, now, now,
	)
	if err != nil {
		return Note{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Note{}, err
	}

	return Note{
		ID:          int(id),
		Title:       title,
		Body:        body,
		Attachments: attachments,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// AddAttachments: dipanggil handler POST /notes/{id}/attachments (B.16)
// setelah tiap part tersimpan ke disk. Baca-ubah-tulis dibungkus satu
// transaksi (tx.Begin/Commit) - dulu waktu in-memory, atomicity-nya dijamin
// sync.Mutex; sekarang mutex itu hilang bareng map-nya, jadi transaksi SQL
// yang menggantikan perannya supaya dua upload bersamaan ke note yang sama
// tidak saling menimpa (lost update).
func (s *NoteStore) AddAttachments(id int, names []string) (Note, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return Note{}, err
	}
	defer tx.Rollback()

	var attJSON string
	err = tx.QueryRow(`SELECT attachments FROM notes WHERE id = ?`, id).Scan(&attJSON)
	if errors.Is(err, sql.ErrNoRows) {
		return Note{}, ErrNotFound
	}
	if err != nil {
		return Note{}, err
	}

	existing, err := decodeAttachments(attJSON)
	if err != nil {
		return Note{}, err
	}
	existing = append(existing, names...)
	newJSON, err := encodeAttachments(existing)
	if err != nil {
		return Note{}, err
	}

	if _, err := tx.Exec(`UPDATE notes SET attachments = ?, updated_at = ? WHERE id = ?`, newJSON, time.Now(), id); err != nil {
		return Note{}, err
	}
	if err := tx.Commit(); err != nil {
		return Note{}, err
	}

	return s.Get(id)
}

// List: terbaru dulu (CreatedAt desc, ID desc sebagai tie-break kalau dua
// baris kebetulan punya created_at identik) - urutan ini sekarang jadi
// tugas SQL (ORDER BY), bukan sort.Slice manual seperti versi in-memory.
func (s *NoteStore) List() ([]Note, error) {
	rows, err := s.db.Query(`SELECT id, title, body, attachments, created_at, updated_at FROM notes ORDER BY created_at DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Note
	for rows.Next() {
		var n Note
		var attJSON string
		if err := rows.Scan(&n.ID, &n.Title, &n.Body, &attJSON, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		attachments, err := decodeAttachments(attJSON)
		if err != nil {
			return nil, err
		}
		n.Attachments = attachments
		out = append(out, n)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// Get: pakai prepared statement (getStmt) - contoh idiom A.56 "Prepare
// sekali, QueryRow berkali-kali", relevan di sini karena Get dipanggil di
// hampir setiap request (Detail, Download, UploadAttachments).
func (s *NoteStore) Get(id int) (Note, error) {
	var n Note
	var attJSON string
	err := s.getStmt.QueryRow(id).Scan(&n.ID, &n.Title, &n.Body, &attJSON, &n.CreatedAt, &n.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Note{}, ErrNotFound
	}
	if err != nil {
		return Note{}, err
	}

	attachments, err := decodeAttachments(attJSON)
	if err != nil {
		return Note{}, err
	}
	n.Attachments = attachments
	return n, nil
}

func (s *NoteStore) Delete(id int) (bool, error) {
	res, err := s.db.Exec(`DELETE FROM notes WHERE id = ?`, id)
	if err != nil {
		return false, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}
