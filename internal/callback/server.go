package callback

import (
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

func NewServer(port int) *Server {
	return &Server{
		port:   port,
		result: make(chan Result, 1),
	}
}

func (s *Server) Start() (string, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", s.port))
	if err != nil {
		return "", fmt.Errorf("listen: %w", err)
	}
	s.listener = listener

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handle)

	s.server = &http.Server{Handler: mux}
	go s.server.Serve(s.listener)

	addr := listener.Addr().(*net.TCPAddr)
	return fmt.Sprintf("http://localhost:%d", addr.Port), nil
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
		w.Write([]byte("Authorization failed: " + errDesc))
		return
	}

	if code == "" {
		s.result <- Result{Error: fmt.Errorf("no code in callback")}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("No code received"))
		return
	}

	s.result <- Result{Code: code, State: state}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Authorization complete! You can close this tab."))
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
