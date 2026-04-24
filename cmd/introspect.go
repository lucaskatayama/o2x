package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/lucaskatayama/oauth2-cli/internal/jwt"
	"github.com/lucaskatayama/oauth2-cli/internal/storage"
)

var introspectCmd = &cobra.Command{
	Use:   "introspect",
	Short: "Decode and print JWT as JSON",
	RunE:  runIntrospect,
}

func runIntrospect(cmd *cobra.Command, args []string) error {
	store, err := storage.NewTokenStorage()
	if err != nil {
		return err
	}
	tok, err := store.Load()
	if err != nil {
		return fmt.Errorf("no token: %w", err)
	}

	v := jwt.NewValidator("")

	var tokenString string
	if tok.IdToken != "" {
		tokenString = tok.IdToken
	} else {
		tokenString = tok.AccessToken
	}

	claims, err := v.Decode(tokenString)
	if err != nil {
		return fmt.Errorf("decode: %w", err)
	}

	out, _ := json.MarshalIndent(claims, "", "  ")
	fmt.Println(string(out))
	return nil
}
