package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/lucaskatayama/oauth2-cli/internal/flow"
	"github.com/lucaskatayama/oauth2-cli/internal/storage"
)

var refreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh access token",
	RunE:  runRefresh,
}

func init() {
	refreshCmd.Flags().String("token-url", "", "OAuth2 token URL")
	refreshCmd.Flags().String("client-id", "", "OAuth2 client ID")
	refreshCmd.Flags().String("client-secret", "", "OAuth2 client secret")
	refreshCmd.Flags().String("scope", "openid profile email", "OAuth2 scopes")
	refreshCmd.Flags().StringVarP(&flowName, "flow", "f", "authorization_code", "OAuth2 flow")
}

func runRefresh(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	store, err := storage.NewTokenStorage()
	if err != nil {
		return err
	}
	tok, err := store.Load()
	if err != nil {
		return fmt.Errorf("no token: %w", err)
	}
	if tok.RefreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	f, err := flow.Get(flowName)
	if err != nil {
		return err
	}

	cfg := &flow.Config{
		TokenURL:     getEnv("OAUTH2_TOKEN_URL"),
		ClientID:     getEnv("OAUTH2_CLIENT_ID"),
		ClientSecret: getEnv("OAUTH2_CLIENT_SECRET"),
		Scope:        getEnvOrDefault("OAUTH2_SCOPE", "openid profile email"),
	}

	newTok, err := f.Refresh(ctx, cfg, tok.RefreshToken)
	if err != nil {
		return err
	}

	if err := store.Save(newTok); err != nil {
		return err
	}
	fmt.Println("Token refreshed")
	return nil
}
