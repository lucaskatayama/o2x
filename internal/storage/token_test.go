package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lucaskatayama/oauth2-cli/internal/flow"
)

func TestTokenStorage(t *testing.T) {
	tmp := t.TempDir()
	s := &TokenStorage{baseDir: tmp}

	tok := &flow.Token{AccessToken: "test-token", RefreshToken: "refresh", Expiry: 1234567890}
	if err := s.Save(tok); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := s.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.AccessToken != tok.AccessToken {
		t.Errorf("Load() AccessToken = %s, want %s", loaded.AccessToken, tok.AccessToken)
	}

	if err := s.Delete(); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}

func TestTokenStorageDefaultDir(t *testing.T) {
	tmp := t.TempDir()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	defer func() {
		_ = os.Chdir(previous)
	}()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}

	s, err := NewTokenStorage()
	if err != nil {
		t.Fatalf("NewTokenStorage() error = %v", err)
	}

	tok := &flow.Token{AccessToken: "cwd"}
	if err := s.Save(tok); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmp, "token.json")); err != nil {
		t.Fatalf("token.json not written to cwd: %v", err)
	}
}
