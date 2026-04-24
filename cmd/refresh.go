package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

	viper.BindPFlag("token-url", refreshCmd.Flags().Lookup("token-url"))
	viper.BindPFlag("client-id", refreshCmd.Flags().Lookup("client-id"))
	viper.BindPFlag("client-secret", refreshCmd.Flags().Lookup("client-secret"))
	viper.BindPFlag("scope", refreshCmd.Flags().Lookup("scope"))
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
		TokenURL:     viper.GetString("token-url"),
		ClientID:     viper.GetString("client-id"),
		ClientSecret: viper.GetString("client-secret"),
		Scope:        viper.GetString("scope"),
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
