package config

import (
    "fmt"
    "net"
    "net/url"
    "os"
    "strconv"
    "strings"

    "github.com/spf13/pflag"
)

type CallbackConfig struct {
    URL  string // full URL used for OAuth redirect_uri
    Host string // host for listener
    Port string // port for listener
}

// Resolve reads flag values (if set), falls back to environment variables, and finally to defaults.
// Precedence: flag > env var > default.
func Resolve(flagSet *pflag.FlagSet) (*CallbackConfig, error) {
    // Helpers to fetch flag if set
    getFlag := func(name string) (string, bool) {
        if flagSet != nil {
            if f := flagSet.Lookup(name); f != nil && f.Changed {
                return f.Value.String(), true
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
        // Build URL – keep same path used previously (/callback)
        u := fmt.Sprintf("http://%s:%s/callback", host, port)
        return &CallbackConfig{URL: u, Host: host, Port: port}, nil
    }

    // 3. Fallback defaults (same as original implementation)
    return &CallbackConfig{URL: "http://localhost:9999/callback", Host: "localhost", Port: "9999"}, nil
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
    host, port, err := net.SplitHostPort(u.Host)
    if err != nil {
        // Host may not include port – try to use default 9999
        host = u.Host
        port = "9999"
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
    // Ensure we have a path ending with /callback
    path := strings.TrimSuffix(u.Path, "/")
    if !strings.HasSuffix(path, "/callback") {
        // Preserve whatever path user supplied – we only need the URL as‑is for OAuth redirect.
        // No extra validation needed.
    }
    return &CallbackConfig{URL: u.String(), Host: host, Port: port}, nil
}
