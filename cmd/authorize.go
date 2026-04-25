package cmd

import (
	"context"
	"fmt"
	"os"
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
	authorizeCmd.Flags().String("scope", "", "OAuth2 scopes")
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
		AuthURL:      getEnv("OAUTH2_AUTH_URL"),
		TokenURL:     getEnv("OAUTH2_TOKEN_URL"),
		ClientID:     getEnv("OAUTH2_CLIENT_ID"),
		ClientSecret: getEnv("OAUTH2_CLIENT_SECRET"),
		Scope:        getEnvOrDefault("OAUTH2_SCOPE", "openid profile email"),
		RedirectURI:  getEnv("OAUTH2_REDIRECT_URI"),
	}

	fmt.Printf("DEBUG cfg: %+v\n", cfg)

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

func getEnv(key string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return ""
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
