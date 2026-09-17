package task

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// fileFormat: bentuk file JSON di disk, terpisah dari Store karena
// Store.tasks/nextID unexported (A.26, tak terbaca encoding/json).
// nextID ikut disimpan supaya ID baru tak bentrok setelah reload.
type fileFormat struct {
	NextID int    `json:"next_id"`
	Tasks  []Task `json:"tasks"`
}

// LoadFromFile (A.50 file + A.53 JSON). Dipanggil dari main.go di awal
// tiap cmdXxx, dan dari parallel.go untuk tiap file yang di-scan.
func LoadFromFile(path string) (*Store, error) {
	f, err := os.Open(path)
	// File belum ada dianggap wajar (mis. run pertama), bukan error fatal.
	if errors.Is(err, os.ErrNotExist) {
		return NewStore(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	// A.36 defer: Close jalan di semua jalur keluar, mencegah file handle bocor.
	defer f.Close()

	var data fileFormat
	if err := json.NewDecoder(f).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}

	s := NewStore()
	for _, t := range data.Tasks {
		s.tasks[t.ID] = t
	}
	s.nextID = data.NextID
	if s.nextID <= 0 {
		s.nextID = 1
	}
	return s, nil
}

// SaveToFile: kebalikan LoadFromFile. Dipanggil main.go di akhir
// cmdAdd/cmdUpdate/cmdDelete setelah state di memory berubah.
// Named return (err error) dipakai supaya defer bisa mengisi err.
func (s *Store) SaveToFile(path string) (err error) {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	// f.Close() sendiri bisa gagal (mis. disk penuh). Error itu hanya
	// dipakai kalau belum ada error lain, biar tak menutupi error asli.
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("close %s: %w", path, cerr)
		}
	}()

	data := fileFormat{NextID: s.nextID, Tasks: s.List()}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ") // biar tasks.json enak dibaca manual
	if encErr := enc.Encode(data); encErr != nil {
		return fmt.Errorf("encode %s: %w", path, encErr)
	}
	return nil
}
