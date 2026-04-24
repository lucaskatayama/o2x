package flow

import "context"

type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Expiry       int64  `json:"expiry"`
	TokenType    string `json:"token_type"`
	IdToken      string `json:"id_token"`
	Scope        string `json:"scope"`
}

type Config struct {
	AuthURL      string
	TokenURL     string
	ClientID     string
	ClientSecret string
	Scope        string
	RedirectURI  string
	JWKSURI      string
}

type Flow interface {
	Name() string
	Authorize(ctx context.Context, cfg *Config) (*Token, error)
	Refresh(ctx context.Context, cfg *Config, refreshToken string) (*Token, error)
	Revoke(ctx context.Context, cfg *Config, token string) error
}
