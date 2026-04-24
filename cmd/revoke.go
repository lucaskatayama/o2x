package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/lucaskatayama/oauth2-cli/internal/storage"
)

var revokeCmd = &cobra.Command{
	Use:   "revoke",
	Short: "Revoke stored token",
	RunE:  runRevoke,
}

func runRevoke(cmd *cobra.Command, args []string) error {
	store, err := storage.NewTokenStorage()
	if err != nil {
		return err
	}
	if err := store.Delete(); err != nil {
		return fmt.Errorf("revoke: %w", err)
	}
	fmt.Println("Token revoked")
	return nil
}
