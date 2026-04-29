package cmd

import (
	"github.com/lucaskatayama/oauth2-cli/internal/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "o2x",
	Short: "OAuth2 CLI with pluggable flows",
	Long: `OAuth2 CLI for authentication with various OAuth2 flows.

Environment variables:
  OAUTH2_AUTH_URL         Authorization URL
  OAUTH2_TOKEN_URL      Token URL
  OAUTH2_CLIENT_ID      Client ID
  OAUTH2_CLIENT_SECRET Client Secret
  OAUTH2_SCOPE         Scopes (default: openid profile email)
  OAUTH2_JWKS_URI      JWKS URI for token verification

Examples:
  # Authorize with Auth0
  export OAUTH2_AUTH_URL=https://your-domain.auth0.com/authorize
  export OAUTH2_TOKEN_URL=https://your-domain.auth0.com/oauth/token
  export OAUTH2_CLIENT_ID=your-client-id
  export OAUTH2_CLIENT_SECRET=your-client-secret
  o2x authorize

  # Show access token (for piping)
  o2x token -n

  # Decode ID token
  o2x decode -t id-token`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Callback configuration flags
	rootCmd.PersistentFlags().String("callback-url", "", "Full callback URL (overrides host/port)")
	rootCmd.PersistentFlags().String("callback-host", "", "Hostname for callback listener (default: localhost)")
	rootCmd.PersistentFlags().String("callback-port", "", "Port for callback listener (default: 9999)")
	config.RegisterFlagSet(rootCmd.PersistentFlags())
	rootCmd.AddCommand(authorizeCmd, tokenCmd, idTokenCmd, refreshCmd, revokeCmd, verifyCmd, introspectCmd, userinfoCmd, decodeCmd)
}
