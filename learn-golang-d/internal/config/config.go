// Package config: menggabungkan dua idiom config C block secara sengaja
// dikontraskan - C.10 Viper untuk nilai-nilai "biasa" (port, timeout, auth
// demo) yang wajar ditaruh di file, dan C.11 environment variable untuk
// satu nilai sensitif (JWT signing key) yang menurut tutorial C.11
// sebaiknya TIDAK ikut ter-commit di file config sama sekali.
package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	AppName string
	Server  ServerConfig
	JWT     JWTConfig
	Auth    AuthConfig
}

type ServerConfig struct {
	Port         int
	ReadTimeout  int
	WriteTimeout int
}

type JWTConfig struct {
	ExpiryMinutes int
	// SigningKey SENGAJA tidak dibaca lewat Viper (C.10) - diisi terpisah
	// lewat env var JWT_SIGNING_KEY (C.11), supaya secret tidak pernah
	// tertulis di configs/config.json yang ikut masuk git.
	SigningKey []byte
}

type AuthConfig struct {
	Username string
	Password string
}

// Load (C.10 Viper + C.11 env var): dipanggil sekali di cmd/api/main.go,
// bukan lewat init() - sama seperti learn-golang-c, supaya path config
// bisa diganti lewat flag saat testing tanpa menyentuh file asli.
func Load(dir, name string) (Config, error) {
	viper.SetConfigType("json")
	viper.AddConfigPath(dir)
	viper.SetConfigName(name)
	if err := viper.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("gagal baca config: %w", err)
	}

	cfg := Config{
		AppName: viper.GetString("app_name"),
		Server: ServerConfig{
			Port:         viper.GetInt("server.port"),
			ReadTimeout:  viper.GetInt("server.read_timeout_seconds"),
			WriteTimeout: viper.GetInt("server.write_timeout_seconds"),
		},
		JWT: JWTConfig{
			ExpiryMinutes: viper.GetInt("jwt.expiry_minutes"),
		},
		Auth: AuthConfig{
			Username: viper.GetString("auth.username"),
			Password: viper.GetString("auth.password"),
		},
	}

	// C.11: SERVER_READ_TIMEOUT_IN_MINUTE-style pattern di tutorial -
	// os.Getenv dicek manual, string kosong berarti "tidak di-override".
	// Di sini dipakai untuk satu nilai wajib (bukan opsional): signing key.
	signingKey := os.Getenv("JWT_SIGNING_KEY")
	if signingKey == "" {
		return Config{}, fmt.Errorf("env var JWT_SIGNING_KEY wajib diisi")
	}
	cfg.JWT.SigningKey = []byte(signingKey)

	return cfg, nil
}
