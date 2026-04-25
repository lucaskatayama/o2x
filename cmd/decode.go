package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/lucaskatayama/oauth2-cli/internal/jwt"
	"github.com/lucaskatayama/oauth2-cli/internal/storage"
)

var decodeCmd = &cobra.Command{
	Use:   "decode [token]",
	Short: "Decode a JWT token (access-token, id-token, or raw JWT)",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runDecode,
}

func init() {
	decodeCmd.Flags().StringP("token-type", "t", "access-token", "Token type: access-token, id-token, or raw (use raw JWT as argument)")
	rootCmd.AddCommand(decodeCmd)
}

func runDecode(cmd *cobra.Command, args []string) error {
	var tokenString string

	if len(args) == 1 {
		tokenString = args[0]
	} else {
		tokenType, _ := cmd.Flags().GetString("token-type")
		store, err := storage.NewTokenStorage()
		if err != nil {
			return err
		}
		tok, err := store.Load()
		if err != nil {
			return fmt.Errorf("no token found: %w", err)
		}
		switch tokenType {
		case "access-token":
			tokenString = tok.AccessToken
		case "id-token":
			tokenString = tok.IdToken
			if tokenString == "" {
				return fmt.Errorf("no ID token found")
			}
		default:
			return fmt.Errorf("unknown token type: %s", tokenType)
		}
	}

	parts, err := jwt.DecodeJWT(tokenString)
	if err != nil {
		return fmt.Errorf("decode JWT: %w", err)
	}

	headerJSON, _ := json.MarshalIndent(parts.Header, "", "  ")
	bodyJSON, _ := json.MarshalIndent(parts.Body, "", "  ")

	fmt.Println("=== HEADER ===")
	fmt.Println(string(headerJSON))
	fmt.Println("=== BODY ===")
	fmt.Println(string(bodyJSON))
	fmt.Println("=== SIGNATURE ===")
	fmt.Println(parts.Signature)

	return nil
}