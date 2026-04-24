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
