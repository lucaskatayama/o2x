package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "o2x",
	Short: "OAuth2 CLI with pluggable flows",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	viper.SetEnvPrefix("OAUTH2")
	viper.AutomaticEnv()

	rootCmd.AddCommand(authorizeCmd, tokenCmd, refreshCmd, revokeCmd, verifyCmd, introspectCmd, userinfoCmd)
}
