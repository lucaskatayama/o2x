package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/lucaskatayama/oauth2-cli/internal/storage"
)

var userinfoCmd = &cobra.Command{
	Use:   "userinfo",
	Short: "Get user info from /userinfo endpoint",
	RunE:  runUserInfo,
}

func init() {
	userinfoCmd.Flags().String("auth-url", "", "OAuth2 authorization URL (used to construct userinfo URL)")
	viper.BindPFlag("auth-url", userinfoCmd.Flags().Lookup("auth-url"))
}

func runUserInfo(cmd *cobra.Command, args []string) error {
	store, err := storage.NewTokenStorage()
	if err != nil {
		return err
	}
	tok, err := store.Load()
	if err != nil {
		return fmt.Errorf("no token: %w", err)
	}

	userinfoURL := viper.GetString("auth-url")
	if userinfoURL != "" {
		u, err := url.Parse(userinfoURL)
		if err == nil {
			u.Path = "/userinfo"
			userinfoURL = u.String()
		}
	}

	if userinfoURL == "" {
		return fmt.Errorf("--auth-url required for userinfo")
	}

	req, err := http.NewRequest("GET", userinfoURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("userinfo request: %w", err)
	}
	defer resp.Body.Close()

	var userinfo interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userinfo); err != nil {
		return fmt.Errorf("decode userinfo: %w", err)
	}

	out, _ := json.MarshalIndent(userinfo, "", "  ")
	fmt.Println(string(out))
	return nil
}
