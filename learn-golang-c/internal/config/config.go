// Package config: B.22 Simple Configuration. Nilai server (port, timeout),
// kredensial basic auth (B.18), dan lokasi upload (B.13/B.16) disatukan di
// satu file JSON supaya tidak hardcode di main.go - konsisten dengan
// tutorial (struct bertingkat + json tag, dibaca sekali via Load).
package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Server struct {
		Port         int `json:"port"`
		ReadTimeout  int `json:"read_timeout_seconds"`
		WriteTimeout int `json:"write_timeout_seconds"`
	} `json:"server"`
	Auth struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"auth"`
	Upload struct {
		Dir      string `json:"dir"`
		MaxBytes int64  `json:"max_bytes"`
	} `json:"upload"`
	ViewsDir  string `json:"views_dir"`
	StaticDir string `json:"static_dir"`
}

// Load: dipanggil main.go sekali saat startup (bukan init(), beda dari
// contoh tutorial, supaya path config.json bisa dioper lewat flag -config
// dan gampang diuji tanpa bergantung ke file di disk kerja package).
func Load(path string) (Config, error) {
	bts, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("baca config %s: %w", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(bts, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config %s: %w", path, err)
	}
	return cfg, nil
}
