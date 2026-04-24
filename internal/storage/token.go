package storage

import (
	"encoding/json"
	"os"
	"os/user"
	"path/filepath"

	"github.com/lucaskatayama/oauth2-cli/internal/flow"
)

type TokenStorage struct {
	baseDir string
}

func NewTokenStorage() (*TokenStorage, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}
	baseDir := filepath.Join(usr.HomeDir, ".config", "o2x")
	if err := os.MkdirAll(baseDir, 0700); err != nil {
		return nil, err
	}
	return &TokenStorage{baseDir: baseDir}, nil
}

func (s *TokenStorage) Save(tok *flow.Token) error {
	data, err := json.MarshalIndent(tok, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(s.baseDir, "token.json")
	return os.WriteFile(path, data, 0600)
}

func (s *TokenStorage) Load() (*flow.Token, error) {
	path := filepath.Join(s.baseDir, "token.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var tok flow.Token
	if err := json.Unmarshal(data, &tok); err != nil {
		return nil, err
	}
	return &tok, nil
}

func (s *TokenStorage) Delete() error {
	path := filepath.Join(s.baseDir, "token.json")
	return os.Remove(path)
}
