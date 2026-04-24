package flows

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/lucaskatayama/oauth2-cli/internal/flow"
)

type ClientCredentialsFlow struct{}

func (f *ClientCredentialsFlow) Name() string { return "client_credentials" }

func (f *ClientCredentialsFlow) Authorize(ctx context.Context, cfg *flow.Config) (*flow.Token, error) {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", cfg.ClientID)
	data.Set("client_secret", cfg.ClientSecret)
	if cfg.Scope != "" {
		data.Set("scope", cfg.Scope)
	}

	resp, err := http.PostForm(cfg.TokenURL, data)
	if err != nil {
		return nil, fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
		Scope        string `json:"scope"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &flow.Token{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		TokenType:    result.TokenType,
		Expiry:       time.Now().Unix() + int64(result.ExpiresIn),
		Scope:        result.Scope,
	}, nil
}

func (f *ClientCredentialsFlow) Refresh(ctx context.Context, cfg *flow.Config, refreshToken string) (*flow.Token, error) {
	return nil, fmt.Errorf("client_credentials does not support refresh")
}

func (f *ClientCredentialsFlow) Revoke(ctx context.Context, cfg *flow.Config, token string) error {
	return nil
}
