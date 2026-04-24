package flow

import (
	"context"
	"testing"
)

type mockFlow struct{}

func (m *mockFlow) Name() string { return "mock" }
func (m *mockFlow) Authorize(ctx context.Context, cfg *Config) (*Token, error) {
	return &Token{AccessToken: "test"}, nil
}
func (m *mockFlow) Refresh(ctx context.Context, cfg *Config, refreshToken string) (*Token, error) {
	return nil, nil
}
func (m *mockFlow) Revoke(ctx context.Context, cfg *Config, token string) error {
	return nil
}

func TestRegistry(t *testing.T) {
	f := &mockFlow{}
	Register(f)

	got, err := Get("mock")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Name() != "mock" {
		t.Errorf("Get() name = %s, want mock", got.Name())
	}

	_, err = Get("nonexistent")
	if err == nil {
		t.Error("Get(nonexistent) expected error")
	}
}
