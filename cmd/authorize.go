package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/lucaskatayama/oauth2-cli/internal/flow"
	"github.com/lucaskatayama/oauth2-cli/internal/storage"
)

var authorizeCmd = &cobra.Command{
	Use:   "authorize",
	Short: "Start OAuth2 authorization flow",
	RunE:  runAuthorize,
}

var flowName string

func init() {
	authorizeCmd.Flags().String("auth-url", "", "OAuth2 authorization URL")
	authorizeCmd.Flags().String("token-url", "", "OAuth2 token URL")
	authorizeCmd.Flags().String("client-id", "", "OAuth2 client ID")
	authorizeCmd.Flags().String("client-secret", "", "OAuth2 client secret")
	authorizeCmd.Flags().String("scope", "openid profile email", "OAuth2 scopes")
	authorizeCmd.Flags().String("redirect-uri", "o2x://callback", "OAuth2 redirect URI")
	authorizeCmd.Flags().StringVarP(&flowName, "flow", "f", "authorization_code", "OAuth2 flow")
}

func runAuthorize(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	f, err := flow.Get(flowName)
	if err != nil {
		return err
	}

	cfg := &flow.Config{
		AuthURL:      mustGetString(cmd, "auth-url"),
		TokenURL:     mustGetString(cmd, "token-url"),
		ClientID:     mustGetString(cmd, "client-id"),
		ClientSecret: mustGetString(cmd, "client-secret"),
		Scope:        mustGetString(cmd, "scope"),
		RedirectURI:  mustGetString(cmd, "redirect-uri"),
	}

	tok, err := f.Authorize(ctx, cfg)
	if err != nil {
		return fmt.Errorf("authorize: %w", err)
	}

	store, err := storage.NewTokenStorage()
	if err != nil {
		return err
	}
	if err := store.Save(tok); err != nil {
		return fmt.Errorf("save token: %w", err)
	}

	fmt.Println("Authorization successful!")
	return nil
}

func mustGetString(cmd *cobra.Command, name string) string {
	val, _ := cmd.Flags().GetString(name)
	return val
}
