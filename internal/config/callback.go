package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/pflag"
)

type CallbackConfig struct {
	URL    string // full URL used for OAuth redirect_uri
	Host   string // host for listener
	Port   string // port for listener
	Scheme string // scheme for listener (http or https)
	Path   string // path to register on listener (e.g., /callback or /v1/callback)
}

var registeredFlagSet *pflag.FlagSet

// RegisterFlagSet allows CLI packages to supply the parsed flag set used for callback configuration.
func RegisterFlagSet(fs *pflag.FlagSet) {
	registeredFlagSet = fs
}

// Resolve reads flag values (if set), falls back to environment variables, and finally to defaults.
// Precedence: flag > env var > default.
func Resolve(flagSet *pflag.FlagSet) (*CallbackConfig, error) {
	if flagSet == nil {
		flagSet = registeredFlagSet
	}

	// Helpers to fetch flag if set
	getFlag := func(name string) (string, bool) {
		if flagSet == nil {
			return "", false
		}
		if f := flagSet.Lookup(name); f != nil {
			if f.Changed {
				return f.Value.String(), true
			}
			if val := f.Value.String(); val != "" {
				return val, true
			}
		}
		return "", false
	}

	// 1. Full URL
	if val, ok := getFlag("callback-url"); ok && val != "" {
		return buildFromURL(val)
	}
	if env := os.Getenv("OAUTH2_CALLBACK_URL"); env != "" {
		return buildFromURL(env)
	}

	// 2. Host + Port (may be partially missing)
	host, hostSet := getFlag("callback-host")
	port, portSet := getFlag("callback-port")
	if !hostSet {
		host = os.Getenv("OAUTH2_CALLBACK_HOST")
	}
	if !portSet {
		port = os.Getenv("OAUTH2_CALLBACK_PORT")
	}
	// If either is present, we build a URL using defaults where missing.
	if host != "" || port != "" {
		// Apply defaults if missing
		if host == "" {
			host = "localhost"
		}
		if port == "" {
			port = "9999"
		}
		// Validate port is numeric
		if _, err := strconv.Atoi(port); err != nil {
			return nil, fmt.Errorf("invalid port %s", port)
		}
		// Determine scheme: default to http, but if port is 443 use https
		scheme := "http"
		if port == "443" {
			scheme = "https"
		}
		path := "/callback"
		// Build URL
		u := fmt.Sprintf("%s://%s:%s%s", scheme, host, port, path)
		return &CallbackConfig{URL: u, Host: host, Port: port, Scheme: scheme, Path: path}, nil
	}

	// 3. Fallback defaults (same as original implementation)
	return &CallbackConfig{URL: "http://localhost:9999/callback", Host: "localhost", Port: "9999", Scheme: "http", Path: "/callback"}, nil
}

// buildFromURL parses a full URL and extracts host/port information.
func buildFromURL(raw string) (*CallbackConfig, error) {
	if !strings.Contains(raw, "://") {
		raw = "http://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid callback URL %s: %w", raw, err)
	}
	if u.Scheme == "" {
		u.Scheme = "http"
	}
	host := u.Hostname()
	port := u.Port()
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		switch strings.ToLower(u.Scheme) {
		case "https":
			port = "443"
		case "http":
			port = "9999"
		default:
			port = "9999"
		}
	}
	if host == "" {
		host = "localhost"
	}
	// Validate port numeric if present
	if port != "" {
		if _, err := strconv.Atoi(port); err != nil {
			return nil, fmt.Errorf("invalid port in callback URL %s", raw)
		}
	}
	// Determine path
	path := u.Path
	if path == "" {
		path = "/callback"
	}
	return &CallbackConfig{URL: u.String(), Host: host, Port: port, Scheme: u.Scheme, Path: path}, nil
}
