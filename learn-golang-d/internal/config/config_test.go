package config

import (
	"os"
	"testing"
)

func TestLoad_Success(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "test-signing-key")

	cfg, err := Load("../../configs", "config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Port == 0 {
		t.Error("expected server.port to be read from config file")
	}
	if cfg.Auth.Username == "" {
		t.Error("expected auth.username to be read from config file")
	}
	if string(cfg.JWT.SigningKey) != "test-signing-key" {
		t.Errorf("SigningKey = %q, want value from env var", cfg.JWT.SigningKey)
	}
}

func TestLoad_MissingSigningKeyEnvVar(t *testing.T) {
	os.Unsetenv("JWT_SIGNING_KEY")

	if _, err := Load("../../configs", "config"); err == nil {
		t.Fatal("expected error when JWT_SIGNING_KEY env var is not set")
	}
}
