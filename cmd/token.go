package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/lucaskatayama/oauth2-cli/internal/storage"
)

var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Show stored access token",
	RunE:  runToken,
}

var idTokenCmd = &cobra.Command{
	Use:   "id-token",
	Short: "Show stored ID token",
	RunE:  runIdToken,
}

func init() {
	tokenCmd.Flags().BoolP("no-newline", "n", false, "Do not print trailing newline")
	idTokenCmd.Flags().BoolP("no-newline", "n", false, "Do not print trailing newline")
}

func init() {
}

func runToken(cmd *cobra.Command, args []string) error {
	store, err := storage.NewTokenStorage()
	if err != nil {
		return err
	}
	tok, err := store.Load()
	if err != nil {
		return fmt.Errorf("no token found: %w", err)
	}
	fmt.Println(tok.AccessToken)
	return nil
}

func runIdToken(cmd *cobra.Command, args []string) error {
	store, err := storage.NewTokenStorage()
	if err != nil {
		return err
	}
	tok, err := store.Load()
	if err != nil {
		return fmt.Errorf("no token found: %w", err)
	}
	if tok.IdToken == "" {
		return fmt.Errorf("no ID token found")
	}
	fmt.Println(tok.IdToken)
	return nil
}
