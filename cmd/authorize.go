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

	viper.BindPFlag("auth-url", authorizeCmd.Flags().Lookup("auth-url"))
	viper.BindPFlag("token-url", authorizeCmd.Flags().Lookup("token-url"))
	viper.BindPFlag("client-id", authorizeCmd.Flags().Lookup("client-id"))
	viper.BindPFlag("client-secret", authorizeCmd.Flags().Lookup("client-secret"))
	viper.BindPFlag("scope", authorizeCmd.Flags().Lookup("scope"))
	viper.BindPFlag("redirect-uri", authorizeCmd.Flags().Lookup("redirect-uri"))
}

func runAuthorize(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	f, err := flow.Get(flowName)
	if err != nil {
		return err
	}

	cfg := &flow.Config{
		AuthURL:      viper.GetString("auth-url"),
		TokenURL:     viper.GetString("token-url"),
		ClientID:     viper.GetString("client-id"),
		ClientSecret: viper.GetString("client-secret"),
		Scope:        viper.GetString("scope"),
		RedirectURI:  viper.GetString("redirect-uri"),
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
