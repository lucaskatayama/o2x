package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/lucaskatayama/oauth2-cli/internal/jwt"
	"github.com/lucaskatayama/oauth2-cli/internal/storage"
)

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify and decode JWT token",
	RunE:  runVerify,
}

func init() {
	verifyCmd.Flags().String("jwks-uri", "", "JWKS URI for key validation")
}

func runVerify(cmd *cobra.Command, args []string) error {
	store, err := storage.NewTokenStorage()
	if err != nil {
		return err
	}
	tok, err := store.Load()
	if err != nil {
		return fmt.Errorf("no token: %w", err)
	}

	jwksURI := os.Getenv("OAUTH2_JWKS_URI")
	v := jwt.NewValidator(jwksURI)

	var tokenString string
	if tok.IdToken != "" {
		tokenString = tok.IdToken
	} else {
		tokenString = tok.AccessToken
	}

	claims, err := v.Validate(tokenString)
	if err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	out, _ := json.MarshalIndent(struct {
		Subject string `json:"sub"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Valid   bool   `json:"valid"`
		Expiry  int64  `json:"exp"`
	}{
		Subject: claims.Subject,
		Email:   claims.Email,
		Name:    claims.Name,
		Valid:   true,
		Expiry:  claims.ExpiresAt.Unix(),
	}, "", "  ")
	fmt.Println(string(out))
	return nil
}
