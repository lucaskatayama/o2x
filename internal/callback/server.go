package callback

import (
	"strconv"
	"github.com/lucaskatayama/oauth2-cli/internal/config"

	"context"
	"fmt"
	"net"
	"net/http"
)

type Result struct {
	Code  string
	State string
	Error error
}

type Server struct {
	port     int
	result   chan Result
	listener net.Listener
	server   *http.Server
}

func NewServer(_ int) *Server {
	// Resolve callback configuration (host/port) using env vars or flags.
	cfg, err := config.Resolve(nil)
	if err != nil {
		// Fallback to defaults on error
		cfg = &config.CallbackConfig{URL: "http://localhost:9999/callback", Host: "localhost", Port: "9999"}
	}
	// Use resolved port; ignore passed argument.
	portNum, _ := strconv.Atoi(cfg.Port)
	return &Server{
		port:   portNum,
		result: make(chan Result, 1),
	}
}

func (s *Server) Start() (string, error) {
	// Resolve configuration on each start to capture latest flag values.
	cfg, err := config.Resolve(nil)
	if err != nil {
		return "", err
	}
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return "", fmt.Errorf("could not start callback listener on %s: %w", addr, err)
	}
	s.listener = ln
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", s.handle)
	s.server = &http.Server{Handler: mux}
	go s.server.Serve(s.listener)
	return cfg.URL, nil
}

func (s *Server) handle(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	code := query.Get("code")
	state := query.Get("state")
	errParam := query.Get("error")

	if errParam != "" {
		errDesc := query.Get("error_description")
		s.result <- Result{Error: fmt.Errorf("%s: %s", errParam, errDesc)}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Authorization failed. You can close this window."))
		return
	}

	if code == "" {
		s.result <- Result{Error: fmt.Errorf("no code in callback")}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("No code received. You can close this window."))
		return
	}

	s.result <- Result{Code: code, State: state}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("<html><body style='font-family: sans-serif; text-align: center; padding: 50px;'><h2>✅ Authorization complete!</h2><p>You can close this window and return to the CLI.</p></body></html>"))
}

func (s *Server) Wait(ctx context.Context) (string, string, error) {
	select {
	case result := <-s.result:
		if result.Error != nil {
			return "", "", result.Error
		}
		return result.Code, result.State, nil
	case <-ctx.Done():
		return "", "", ctx.Err()
	}
}

func (s *Server) Close() error {
	if s.server != nil {
		return s.server.Shutdown(context.Background())
	}
	return nil
}

func (s *Server) DeepLinkURL() string {
	return "o2x://callback"
}
