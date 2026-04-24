package flows

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"

	"github.com/google/uuid"
	"github.com/lucaskatayama/oauth2-cli/internal/browser"
	"github.com/lucaskatayama/oauth2-cli/internal/callback"
	"github.com/lucaskatayama/oauth2-cli/internal/flow"
	"golang.org/x/oauth2"
)

type AuthCodeFlow struct{}

func (f *AuthCodeFlow) Name() string { return "authorization_code" }

func (f *AuthCodeFlow) Authorize(ctx context.Context, cfg *flow.Config) (*flow.Token, error) {
	redirectURI := cfg.RedirectURI
	if redirectURI == "" {
		redirectURI = "o2x://callback"
	}

	cb := callback.NewServer(0)
	if _, err := cb.Start(); err != nil {
		return nil, fmt.Errorf("callback server: %w", err)
	}
	defer cb.Close()

	verifier := generateRandomString(32)
	h := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(h[:])

	scopes := cfg.Scope
	if scopes == "" {
		scopes = "openid profile email"
	}

	oauth2Cfg := &oauth2.Config{
		ClientID:    cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL: redirectURI,
		Endpoint: oauth2.Endpoint{
			AuthURL:  cfg.AuthURL,
			TokenURL: cfg.TokenURL,
		},
		Scopes: parseScopes(scopes),
	}

	state := uuid.New().String()
	authURL := oauth2Cfg.AuthCodeURL(state,
		oauth2.SetAuthURLParam("code_challenge", challenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"))

	if err := browser.Open(authURL); err != nil {
		fmt.Printf("Open this URL manually: %s\n", authURL)
	}

	code, returnedState, err := cb.Wait(ctx)
	if err != nil {
		return nil, fmt.Errorf("callback wait: %w", err)
	}

	if state != returnedState {
		return nil, fmt.Errorf("state mismatch")
	}

	tok, err := oauth2Cfg.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", verifier))
	if err != nil {
		return nil, fmt.Errorf("token exchange: %w", err)
	}

	idToken := ""
	if idTok, ok := tok.Extra("id_token").(string); ok {
		idToken = idTok
	}

	return &flow.Token{
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
		IdToken:      idToken,
		TokenType:    tok.TokenType,
		Expiry:       tok.Expiry.Unix(),
		Scope:        scopes,
	}, nil
}

func (f *AuthCodeFlow) Refresh(ctx context.Context, cfg *flow.Config, refreshToken string) (*flow.Token, error) {
	oauth2Cfg := &oauth2.Config{
		ClientID:    cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  cfg.AuthURL,
			TokenURL: cfg.TokenURL,
		},
	}

	tok, err := oauth2Cfg.TokenSource(ctx, &oauth2.Token{RefreshToken: refreshToken}).Token()
	if err != nil {
		return nil, err
	}

	idToken := ""
	if idTok, ok := tok.Extra("id_token").(string); ok {
		idToken = idTok
	}

	return &flow.Token{
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
		IdToken:      idToken,
		TokenType:    tok.TokenType,
		Expiry:       tok.Expiry.Unix(),
	}, nil
}

func (f *AuthCodeFlow) Revoke(ctx context.Context, cfg *flow.Config, token string) error {
	return nil
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~"
	result := make([]byte, length)
	for i := range result {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		result[i] = charset[n.Int64()]
	}
	return string(result)
}

func parseScopes(scope string) []string {
	if scope == "" {
		return nil
	}
	var scopes []string
	for _, s := range strings.Split(scope, ",") {
		s = strings.TrimSpace(s)
		if s != "" {
			scopes = append(scopes, s)
		}
	}
	return scopes
}
